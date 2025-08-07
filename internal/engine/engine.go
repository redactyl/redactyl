package engine

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"

	"github.com/accrava/redactyl/internal/detectors"
	"github.com/accrava/redactyl/internal/git"
	"github.com/accrava/redactyl/internal/ignore"
)

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

type Config struct {
	Root           string
	IncludeGlobs   string
	ExcludeGlobs   string
	MaxBytes       int64
	ScanStaged     bool
	HistoryCommits int
	BaseBranch     string
	Threads        int
}

func Scan(cfg Config) ([]Finding, error) {
	threads := cfg.Threads
	if threads <= 0 {
		threads = runtime.GOMAXPROCS(0)
	}
	pool := newWorkerPool(threads)

	ign, _ := ignore.Load(filepath.Join(cfg.Root, ".redactylignore"))
	ctx := context.Background()

	var out []Finding
	emit := func(fs []Finding) {
		out = append(out, fs...)
	}

	// working tree / staged
	if cfg.HistoryCommits == 0 && cfg.BaseBranch == "" {
		err := Walk(ctx, cfg, ign, pool, func(p string, data []byte) {
			fs := detectors.RunAll(p, data)
			emit(fs)
		})
		if err != nil {
			return nil, err
		}
	}

	// staged
	if cfg.ScanStaged {
		files, data, err := git.StagedDiff(cfg.Root)
		if err == nil {
			for i, p := range files {
				fs := detectors.RunAll(p, data[i])
				emit(fs)
			}
		}
	}

	// history
	if cfg.HistoryCommits > 0 {
		entries, err := git.LastNCommits(cfg.Root, cfg.HistoryCommits)
		if err == nil {
			for _, e := range entries {
				for path, blob := range e.Files {
					if ign.Match(path) {
						continue
					}
					if int64(len(blob)) > cfg.MaxBytes {
						continue
					}
					fs := detectors.RunAll(path, blob)
					emit(fs)
				}
			}
		}
	}

	// diff vs base branch
	if cfg.BaseBranch != "" {
		files, data, err := git.DiffAgainst(cfg.Root, cfg.BaseBranch)
		if err == nil {
			for i, p := range files {
				if ign.Match(p) {
					continue
				}
				if int64(len(data[i])) > cfg.MaxBytes {
					continue
				}
				fs := detectors.RunAll(p, bytes.TrimSpace(data[i]))
				emit(fs)
			}
		}
	}

	pool.Wait()
	return out, nil
}
