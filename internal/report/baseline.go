package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/redactyl/redactyl/internal/types"
)

type Baseline struct {
	Items map[string]bool `json:"items"`
}

const FingerprintMetadataKey = "redactyl_fingerprint"

func LoadBaseline(path string) (Baseline, error) {
	b := Baseline{Items: map[string]bool{}}
	f, err := os.ReadFile(path)
	if err != nil {
		return b, err
	}
	if err := json.Unmarshal(f, &b); err != nil {
		return b, err
	}
	return b, nil
}

func SaveBaseline(path string, findings []types.Finding) error {
	b := Baseline{Items: map[string]bool{}}
	for _, f := range findings {
		b.Items[FindingKey(f)] = true
	}
	buf, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

func FilterNewFindings(findings []types.Finding, base Baseline) []types.Finding {
	var out []types.Finding
	for _, f := range findings {
		if !IsBaselined(f, base.Items) {
			out = append(out, f)
		}
	}
	return out
}

// FindingKey returns a stable fingerprint for a finding without storing raw match text.
// It prefers a precomputed fingerprint in metadata when present.
func FindingKey(f types.Finding) string {
	if f.Metadata != nil {
		if fp := f.Metadata[FingerprintMetadataKey]; fp != "" {
			return fp
		}
	}
	return fingerprintValue(f.Path, f.Detector, f.Match)
}

// LegacyFindingKey returns the pre-1.0.1 baseline key format.
func LegacyFindingKey(f types.Finding) string {
	return f.Path + "|" + f.Detector + "|" + f.Match
}

// IsBaselined checks both current and legacy baseline key formats.
func IsBaselined(f types.Finding, items map[string]bool) bool {
	if items == nil {
		return false
	}
	if items[FindingKey(f)] {
		return true
	}
	if items[LegacyFindingKey(f)] {
		return true
	}
	return false
}

func fingerprintValue(path, detector, match string) string {
	sum := sha256.Sum256([]byte(path + "|" + detector + "|" + match))
	return hex.EncodeToString(sum[:])
}

func ShouldFail(findings []types.Finding, failOn string) bool {
	level := map[string]int{"low": 1, "medium": 2, "high": 3}
	th := level[failOn]
	if th == 0 {
		th = 2
	}
	for _, f := range findings {
		if level[string(f.Severity)] >= th {
			return true
		}
	}
	return false
}
