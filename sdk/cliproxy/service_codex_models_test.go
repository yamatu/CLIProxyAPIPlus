package cliproxy

import (
	"testing"

	internalconfig "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/config"
)

func TestRegisterModelsForAuth_CodexConfigModelsKeepStaticCatalog(t *testing.T) {
	service := &Service{
		cfg: &config.Config{
			CodexKey: []config.CodexKey{
				{
					APIKey: "test-codex-key",
					Models: []internalconfig.CodexModel{
						{Name: "gpt-5", Alias: "team/gpt-5"},
					},
				},
			},
		},
	}
	auth := &coreauth.Auth{
		ID:       "codex-auth-model-merge",
		Provider: "codex",
		Status:   coreauth.StatusActive,
		Attributes: map[string]string{
			"auth_kind": "apikey",
			"api_key":   "test-codex-key",
		},
	}

	reg := registry.GetGlobalRegistry()
	reg.UnregisterClient(auth.ID)
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	service.registerModelsForAuth(auth)

	models := reg.GetModelsForClient(auth.ID)
	if len(models) == 0 {
		t.Fatal("expected codex models to be registered")
	}

	seenAlias := false
	seenGPT55 := false
	for _, model := range models {
		if model == nil {
			continue
		}
		switch model.ID {
		case "team/gpt-5":
			seenAlias = true
		case "gpt-5.5":
			seenGPT55 = true
		}
	}

	if !seenAlias {
		t.Fatal("expected configured codex alias to stay registered")
	}
	if !seenGPT55 {
		t.Fatal("expected static codex catalog to still include gpt-5.5")
	}
}

func TestRegisterModelsForAuth_CodexConfigExcludedModelsStillApply(t *testing.T) {
	service := &Service{
		cfg: &config.Config{
			CodexKey: []config.CodexKey{
				{
					APIKey:         "test-codex-key-excluded",
					ExcludedModels: []string{"gpt-5.5"},
					Models: []internalconfig.CodexModel{
						{Name: "gpt-5", Alias: "team/gpt-5"},
					},
				},
			},
		},
	}
	auth := &coreauth.Auth{
		ID:       "codex-auth-model-excluded",
		Provider: "codex",
		Status:   coreauth.StatusActive,
		Attributes: map[string]string{
			"auth_kind": "apikey",
			"api_key":   "test-codex-key-excluded",
		},
	}

	reg := registry.GetGlobalRegistry()
	reg.UnregisterClient(auth.ID)
	t.Cleanup(func() { reg.UnregisterClient(auth.ID) })

	service.registerModelsForAuth(auth)

	models := reg.GetModelsForClient(auth.ID)
	for _, model := range models {
		if model == nil {
			continue
		}
		if model.ID == "gpt-5.5" {
			t.Fatal("expected excluded codex model gpt-5.5 to be filtered out")
		}
	}
}
