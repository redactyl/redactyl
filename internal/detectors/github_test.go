package detectors

import "testing"

func TestGitHubToken(t *testing.T) {
	data := []byte("token=ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	fs := GitHubToken("x.txt", data)
	if len(fs) == 0 {
		t.Fatalf("expected github token finding")
	}
}

func TestGitHubToken_Negative_EmbeddedSubstring(t *testing.T) {
	data := []byte("prefix-XYZghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789XYZ-suffix")
	fs := GitHubToken("x.txt", data)
	if len(fs) != 0 {
		t.Fatalf("unexpected finding for embedded substring")
	}
}
