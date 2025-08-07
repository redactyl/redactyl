package types

type Severity string

const (
	SevLow  Severity = "low"
	SevMed  Severity = "medium"
	SevHigh Severity = "high"
)

type Finding struct {
	Path       string   `json:"path"`
	Line       int      `json:"line"`
	Match      string   `json:"match"`
	Detector   string   `json:"detector"`
	Severity   Severity `json:"severity"`
	Confidence float64  `json:"confidence"`
}

