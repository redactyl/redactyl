package detectors

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/accrava/redactyl/internal/engine"
)

func PrivateKeyBlock(path string, data []byte) []engine.Finding {
	var out []engine.Finding
	sc := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		t := sc.Text()
		if strings.Contains(t, "-----BEGIN ") && strings.Contains(t, " PRIVATE KEY-----") {
			out = append(out, engine.Finding{
				Path: path, Line: line, Match: "BEGIN PRIVATE KEY",
				Detector: "private_key_block", Severity: engine.SevHigh, Confidence: 0.99,
			})
		}
	}
	return out
}
