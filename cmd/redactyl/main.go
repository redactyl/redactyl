package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/accrava/redactyl/internal/engine"
	"github.com/accrava/redactyl/internal/report"
)

func main() {
	var (
		path           string
		scanStaged     bool
		scanHistory    int
		baseBranch     string
		includeGlobs   string
		excludeGlobs   string
		maxBytes       int64
		jsonOut        bool
		sarifOut       bool
		updateBaseline bool
		failOn         string
		threads        int
	)
	flag.StringVar(&path, "path", ".", "path to scan")
	flag.BoolVar(&scanStaged, "staged", false, "scan staged changes")
	flag.IntVar(&scanHistory, "history", 0, "scan last N commits (0=off)")
	flag.StringVar(&baseBranch, "base", "", "scan diff vs base branch (e.g. main)")
	flag.StringVar(&includeGlobs, "include", "", "comma-separated include globs")
	flag.StringVar(&excludeGlobs, "exclude", "", "comma-separated exclude globs")
	flag.Int64Var(&maxBytes, "max-bytes", 1<<20, "skip files larger than this")
	flag.BoolVar(&jsonOut, "json", false, "emit JSON")
	flag.BoolVar(&sarifOut, "sarif", false, "emit SARIF")
	flag.BoolVar(&updateBaseline, "update-baseline", false, "write baseline file")
	flag.StringVar(&failOn, "fail-on", "medium", "lowest severity that fails (low|medium|high)")
	flag.IntVar(&threads, "threads", 0, "worker count (0 = GOMAXPROCS)")
	flag.Parse()

	abs, _ := filepath.Abs(path)

	cfg := engine.Config{
		Root:           abs,
		IncludeGlobs:   includeGlobs,
		ExcludeGlobs:   excludeGlobs,
		MaxBytes:       maxBytes,
		ScanStaged:     scanStaged,
		HistoryCommits: scanHistory,
		BaseBranch:     baseBranch,
		Threads:        threads,
	}

	results, err := engine.Scan(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan error:", err)
		os.Exit(2)
	}

	// Baseline load/diff
	baseline, _ := report.LoadBaseline("redactyl.baseline.json")
	newFindings := report.FilterNewFindings(results, baseline)

	// Output
	switch {
	case sarifOut:
		fmt.Fprintln(os.Stderr, "SARIF output not implemented yet")
		os.Exit(1)
	case jsonOut:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(newFindings)
	default:
		report.PrintTable(os.Stdout, newFindings)
	}

	if updateBaseline {
		if err := report.SaveBaseline("redactyl.baseline.json", results); err != nil {
			fmt.Fprintln(os.Stderr, "baseline write error:", err)
			os.Exit(2)
		}
	}

	// exit codes: 0=ok, 1=findings, 2=error
	if report.ShouldFail(newFindings, failOn) {
		os.Exit(1)
	}
}
