package adkflow

import "testing"

func TestChatCompletionsURL(t *testing.T) {
	tests := map[string]string{
		"https://example.com/v1":                  "https://example.com/v1/chat/completions",
		"https://example.com/v1/":                 "https://example.com/v1/chat/completions",
		"https://example.com/v1/chat/completions": "https://example.com/v1/chat/completions",
	}

	for input, want := range tests {
		if got := chatCompletionsURL(input); got != want {
			t.Fatalf("chatCompletionsURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtractJSONObject(t *testing.T) {
	raw := "```json\n{\"goal\":\"test\"}\n```"
	got := extractJSONObject(raw)
	if got != "{\"goal\":\"test\"}" {
		t.Fatalf("extractJSONObject(%q) = %q", raw, got)
	}
}
