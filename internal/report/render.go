package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/accrava/redactyl/internal/engine"
)

func PrintTable(w io.Writer, findings []engine.Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Path < findings[j].Path
	})
	if len(findings) == 0 {
		fmt.Fprintln(w, "No secrets found ✅")
		return
	}
	fmt.Fprintf(w, "Findings: %d\n", len(findings))
	for _, f := range findings {
		mask := maskValue(f.Match)
		fmt.Fprintf(w, "%-6s %-8s %s:%d  %s\n", f.Severity, f.Detector, f.Path, f.Line, mask)
	}
}

func maskValue(s string) string {
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "…" + s[len(s)-4:]
}
