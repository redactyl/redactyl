package detectors

import "github.com/accrava/redactyl/internal/engine"

type Detector func(path string, data []byte) []engine.Finding

var all = []Detector{
	AWSKeys, GitHubToken, SlackToken, JWTToken, PrivateKeyBlock, EntropyNearbySecrets,
}

func RunAll(path string, data []byte) []engine.Finding {
	var out []engine.Finding
	for _, d := range all {
		out = append(out, d(path, data)...)
	}
	return dedupe(out)
}
