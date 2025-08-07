package detectors

import "github.com/accrava/redactyl/internal/types"

type Detector func(path string, data []byte) []types.Finding

var all = []Detector{
	AWSKeys, GitHubToken, SlackToken, JWTToken, PrivateKeyBlock, EntropyNearbySecrets, StripeSecret,
}

func RunAll(path string, data []byte) []types.Finding {
	var out []types.Finding
	for _, d := range all {
		out = append(out, d(path, data)...)
	}
	return dedupe(out)
}

func dedupe(findings []types.Finding) []types.Finding {
	seen := make(map[string]bool)
	var result []types.Finding

	for _, f := range findings {
		key := f.Path + "|" + f.Detector + "|" + f.Match
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}
