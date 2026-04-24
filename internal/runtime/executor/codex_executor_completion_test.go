package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
	"github.com/tidwall/gjson"
)

func TestNormalizeCodexCompletionEventResponseDone(t *testing.T) {
	input := []byte(`{"type":"response.done","response":{"id":"resp_1"}}`)
	got := normalizeCodexCompletionEvent(input)
	if gotType := gjson.GetBytes(got, "type").String(); gotType != "response.completed" {
		t.Fatalf("type = %q, want response.completed", gotType)
	}
}

func TestPatchCodexCompletedOutputReconstructsOutput(t *testing.T) {
	completed := []byte(`{"type":"response.completed","response":{"id":"resp_1","output":[]}}`)
	item0 := []byte(`{"type":"message","id":"msg_1","role":"assistant","content":[{"type":"output_text","text":"hello"}]}`)
	item1 := []byte(`{"type":"message","id":"msg_2","role":"assistant","content":[{"type":"output_text","text":"world"}]}`)

	got := patchCodexCompletedOutput(completed, map[int64][]byte{
		1: item1,
		0: item0,
	}, nil)

	output := gjson.GetBytes(got, "response.output").Array()
	if len(output) != 2 {
		t.Fatalf("output length = %d, want 2", len(output))
	}
	if text := output[0].Get("content.0.text").String(); text != "hello" {
		t.Fatalf("first output text = %q, want hello", text)
	}
	if text := output[1].Get("content.0.text").String(); text != "world" {
		t.Fatalf("second output text = %q, want world", text)
	}
}

func TestCodexExecutorExecuteAcceptsResponseDone(t *testing.T) {
	var upstreamModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamModel = gjson.GetBytes(mustReadAll(t, r.Body), "model").String()
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "data: {\"type\":\"response.output_item.done\",\"output_index\":0,\"item\":{\"type\":\"message\",\"id\":\"msg_1\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"hello from codex\"}]}}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"type\":\"response.done\",\"response\":{\"id\":\"resp_1\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":3}}}\n\n")
	}))
	defer server.Close()

	exec := NewCodexExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Attributes: map[string]string{
			"api_key":  "test-key",
			"base_url": server.URL,
		},
	}
	req := cliproxyexecutor.Request{
		Model: "gpt-5.5",
		Payload: []byte(`{
			"model":"gpt-5.5",
			"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}]
		}`),
	}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("codex")}

	resp, err := exec.Execute(context.Background(), auth, req, opts)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if text := gjson.GetBytes(resp.Payload, "response.output.0.content.0.text").String(); text != "hello from codex" {
		t.Fatalf("response output text = %q, want hello from codex", text)
	}
	if gotType := gjson.GetBytes(resp.Payload, "type").String(); gotType != "response.completed" {
		t.Fatalf("response type = %q, want response.completed", gotType)
	}
	if upstreamModel != "gpt-5.5" {
		t.Fatalf("upstream model = %q, want gpt-5.5", upstreamModel)
	}
}

func TestCodexExecutorExecutePreservesChatGPTAccountModel(t *testing.T) {
	var upstreamModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamModel = gjson.GetBytes(mustReadAll(t, r.Body), "model").String()
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "data: {\"type\":\"response.done\",\"response\":{\"id\":\"resp_1\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":3}}}\n\n")
	}))
	defer server.Close()

	exec := NewCodexExecutor(&config.Config{})
	auth := &cliproxyauth.Auth{
		Attributes: map[string]string{
			"base_url": server.URL,
		},
		Metadata: map[string]any{
			"access_token": "test-access-token",
		},
	}
	req := cliproxyexecutor.Request{
		Model: "gpt-5.5",
		Payload: []byte(`{
			"model":"gpt-5.5",
			"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}]
		}`),
	}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("codex")}

	if _, err := exec.Execute(context.Background(), auth, req, opts); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if upstreamModel != "gpt-5.5" {
		t.Fatalf("upstream model = %q, want gpt-5.5", upstreamModel)
	}
}

func mustReadAll(t *testing.T, body io.ReadCloser) []byte {
	t.Helper()
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return data
}
