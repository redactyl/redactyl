package detectors

import "testing"

func TestOpenAIAPIKey(t *testing.T) {
	// Valid 48-character key (after sk- prefix) - total 51 chars including sk-
	data := []byte("OPENAI_API_KEY=sk-abcdefghijklmnopqrstuvwxyz0123456789ABCDEF")
	fs := OpenAIAPIKey("x.txt", data)
	if len(fs) == 0 {
		t.Fatalf("expected openai api key finding")
	}

	// Test negative cases
	negativeTests := []string{
		"sk-tooshort", // Too short
		"sk-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHtoolong", // Too long
		"not-sk-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGH",    // Wrong prefix
		"just some random text with sk- prefix but no context",
	}

	for _, test := range negativeTests {
		fs := OpenAIAPIKey("x.txt", []byte(test))
		if len(fs) > 0 {
			t.Fatalf("unexpected finding for negative test: %s", test)
		}
	}
}
