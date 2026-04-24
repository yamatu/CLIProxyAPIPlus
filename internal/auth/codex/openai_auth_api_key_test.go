package codex

import "testing"

func TestExtractCodexAPIKey(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "top level api_key",
			body: `{"api_key":"sk-test-1"}`,
			want: "sk-test-1",
		},
		{
			name: "nested api_key_data",
			body: `{"api_key_data":{"key":"sk-test-2"}}`,
			want: "sk-test-2",
		},
		{
			name: "missing",
			body: `{"access_token":"tok"}`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractCodexAPIKey([]byte(tt.body)); got != tt.want {
				t.Fatalf("extractCodexAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
