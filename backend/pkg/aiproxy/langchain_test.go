package aiproxy

import "testing"

func TestDefaultAgentSystemPrompt(t *testing.T) {
	prompt := DefaultAgentSystemPrompt()
	if prompt == "" {
		t.Fatal("expected non-empty default system prompt")
	}
	if !containsAll(prompt, "Smart Ledger", "rag-export") {
		t.Fatalf("prompt missing expected content: %q", prompt[:min(120, len(prompt))])
	}
}

func TestNormalizeLLMBaseURL(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		want string
	}{
		{"https://api.deepseek.com/v1", true, "https://api.deepseek.com/v1"},
		{"https://api.deepseek.com", true, "https://api.deepseek.com/v1"},
		{"http://127.0.0.1:11434", true, "http://127.0.0.1:11434/v1"},
		{"http://192.168.1.1:11434", false, ""},
	}
	for _, tc := range cases {
		got, err := normalizeLLMBaseURL(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
		if tc.ok && got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
