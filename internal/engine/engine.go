package engine

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/accrava/redactyl/internal/cache"
	"github.com/accrava/redactyl/internal/detectors"
	"github.com/accrava/redactyl/internal/git"
	"github.com/accrava/redactyl/internal/ignore"
	"github.com/accrava/redactyl/internal/types"
	xxhash "github.com/cespare/xxhash/v2"
)

type Config struct {
	Root             string
	IncludeGlobs     string
	ExcludeGlobs     string
	MaxBytes         int64
	ScanStaged       bool
	HistoryCommits   int
	BaseBranch       string
	Threads          int
	EnableDetectors  string
	DisableDetectors string
	MinConfidence    float64
	DryRun           bool
	NoColor          bool
	DefaultExcludes  bool
	NoCache          bool
	Progress         func()
}

var (
	EnableDetectors  string
	DisableDetectors string
)

func DetectorIDs() []string { return detectors.IDs() }

func Scan(cfg Config) ([]types.Finding, error) {
	res, err := ScanWithStats(cfg)
	if err != nil {
		return nil, err
	}
	return res.Findings, nil
}

type Result struct {
	Findings     []types.Finding
	FilesScanned int
	Duration     time.Duration
}

func ScanWithStats(cfg Config) (Result, error) {
	var result Result
	// Load incremental cache if available
	var db cache.DB
	if !cfg.NoCache {
		db, _ = cache.Load(cfg.Root)
	} else {
		db.Entries = map[string]string{}
	}
	// collect updated hashes to persist once at end
	updated := map[string]string{}
	threads := cfg.Threads
	if threads <= 0 {
		threads = runtime.GOMAXPROCS(0)
	}
	pool := newWorkerPool(threads)

	ign, _ := ignore.Load(filepath.Join(cfg.Root, ".redactylignore"))
	ctx := context.Background()

	var out []types.Finding
	started := time.Now()
	emit := func(fs []types.Finding) {
		out = append(out, fs...)
	}

	// working tree / staged
	if cfg.HistoryCommits == 0 && cfg.BaseBranch == "" {
		err := Walk(ctx, cfg, ign, pool, func(p string, data []byte) {
			// compute cheap content hash; small overhead but enables skipping next run
			h := fastHash(data)
			if !cfg.NoCache && db.Entries != nil && db.Entries[p] == h {
				return
			}
			result.FilesScanned++
			if cfg.Progress != nil {
				cfg.Progress()
			}
			if cfg.DryRun {
				return
			}
			fs := detectors.RunAll(p, data)
			fs = filterByConfidence(fs, cfg.MinConfidence)
			fs = filterByIDs(fs, cfg.EnableDetectors, cfg.DisableDetectors)
			emit(fs)
			if !cfg.NoCache {
				updated[p] = h
			}
		})
		if err != nil {
			return result, err
		}
	}

	// staged
	if cfg.ScanStaged {
		files, data, err := git.StagedDiff(cfg.Root)
		if err == nil {
			for i, p := range files {
				result.FilesScanned++
				fs := detectors.RunAll(p, data[i])
				fs = filterByConfidence(fs, cfg.MinConfidence)
				fs = filterByIDs(fs, cfg.EnableDetectors, cfg.DisableDetectors)
				emit(fs)
				if !cfg.NoCache {
					updated[p] = fastHash(data[i])
				}
				result.FilesScanned++
				if cfg.Progress != nil {
					cfg.Progress()
				}
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
					result.FilesScanned++
					if cfg.DryRun {
						continue
					}
					fs := detectors.RunAll(path, blob)
					fs = filterByConfidence(fs, cfg.MinConfidence)
					fs = filterByIDs(fs, cfg.EnableDetectors, cfg.DisableDetectors)
					emit(fs)
					if !cfg.NoCache {
						updated[path] = fastHash(blob)
					}
					result.FilesScanned++
					if cfg.Progress != nil {
						cfg.Progress()
					}
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
				result.FilesScanned++
				if cfg.DryRun {
					continue
				}
				fs := detectors.RunAll(p, bytes.TrimSpace(data[i]))
				fs = filterByConfidence(fs, cfg.MinConfidence)
				fs = filterByIDs(fs, cfg.EnableDetectors, cfg.DisableDetectors)
				emit(fs)
				if !cfg.NoCache {
					updated[p] = fastHash(data[i])
				}
				result.FilesScanned++
				if cfg.Progress != nil {
					cfg.Progress()
				}
			}
		}
	}

	pool.Wait()
	result.Findings = out
	result.Duration = time.Since(started)
	// Save cache best-effort
	if !cfg.NoCache && len(updated) > 0 {
		if db.Entries == nil {
			db.Entries = map[string]string{}
		}
		for k, v := range updated {
			db.Entries[k] = v
		}
		_ = cache.Save(cfg.Root, db)
	}
	return result, nil
}

// fastHash returns a short hex digest for quick change detection.
func fastHash(b []byte) string {
	if len(b) == 0 {
		return "0000000000000000"
	}
	sum := xxhash.Sum64(b)
	// fixed-width lower-hex for stable cache keys
	var buf [16]byte
	const hex = "0123456789abcdef"
	for i := 15; i >= 0; i-- {
		buf[i] = hex[sum&0xF]
		sum >>= 4
	}
	return string(buf[:])
}

func filterByConfidence(fs []types.Finding, min float64) []types.Finding {
	if min <= 0 {
		return fs
	}
	var out []types.Finding
	for _, f := range fs {
		if f.Confidence >= min {
			out = append(out, f)
		}
	}
	return out
}

func filterByIDs(fs []types.Finding, enable, disable string) []types.Finding {
	if enable == "" && disable == "" {
		return fs
	}
	allowed := map[string]bool{}
	if enable != "" {
		for _, id := range strings.Split(enable, ",") {
			allowed[strings.TrimSpace(id)] = true
		}
	}
	blocked := map[string]bool{}
	if disable != "" {
		for _, id := range strings.Split(disable, ",") {
			blocked[strings.TrimSpace(id)] = true
		}
	}
	var out []types.Finding
	for _, f := range fs {
		if enable != "" && !allowed[f.Detector] {
			continue
		}
		if disable != "" && blocked[f.Detector] {
			continue
		}
		out = append(out, f)
	}
	return out
}
