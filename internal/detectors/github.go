package detectors

import (
	"bufio"
	"bytes"
	"regexp"

	"github.com/accrava/redactyl/internal/engine"
)

// PAT formats evolve; this covers classic ghp_ tokens.
var reGHP = regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`)

func GitHubToken(path string, data []byte) []engine.Finding {
	var out []engine.Finding
	sc := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		if reGHP.FindStringIndex(sc.Text()) != nil {
			out = append(out, engine.Finding{
				Path: path, Line: line, Match: reGHP.FindString(sc.Text()),
				Detector: "github_token", Severity: engine.SevHigh, Confidence: 0.9,
			})
		}
	}
	return out
}
