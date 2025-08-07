package detectors

import (
	"bufio"
	"bytes"
	"regexp"

	"github.com/accrava/redactyl/internal/engine"
)

var reJWT = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+?\.[A-Za-z0-9._-]+?\.[A-Za-z0-9._-]+`)

func JWTToken(path string, data []byte) []engine.Finding {
	var out []engine.Finding
	sc := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		if reJWT.FindStringIndex(sc.Text()) != nil {
			out = append(out, engine.Finding{
				Path: path, Line: line, Match: reJWT.FindString(sc.Text()),
				Detector: "jwt", Severity: engine.SevMed, Confidence: 0.7,
			})
		}
	}
	return out
}
