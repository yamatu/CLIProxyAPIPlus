package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestFileTokenStoreReadAuthFilePromotesAPIKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "codex-user.json")
	if err := os.WriteFile(path, []byte(`{"type":"codex","email":"user@example.com","api_key":"sk-test"}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewFileTokenStore()
	store.SetBaseDir(dir)

	auth, err := store.readAuthFile(path, dir)
	if err != nil {
		t.Fatalf("readAuthFile() error = %v", err)
	}
	if auth == nil {
		t.Fatal("readAuthFile() returned nil auth")
	}
	if got := auth.Attributes["api_key"]; got != "sk-test" {
		t.Fatalf("api_key attribute = %q, want sk-test", got)
	}
}

func TestFileTokenStoreSavePersistsMetadataAPIKey(t *testing.T) {
	dir := t.TempDir()
	store := NewFileTokenStore()
	store.SetBaseDir(dir)

	auth := &cliproxyauth.Auth{
		ID:       "codex-user.json",
		FileName: "codex-user.json",
		Provider: "codex",
		Metadata: map[string]any{
			"type":    "codex",
			"email":   "user@example.com",
			"api_key": "sk-test",
		},
	}

	if _, err := store.Save(context.Background(), auth); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	saved, err := os.ReadFile(filepath.Join(dir, "codex-user.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(saved) == "" || !strings.Contains(string(saved), `"api_key":"sk-test"`) {
		t.Fatalf("saved file missing api_key: %s", string(saved))
	}
}
