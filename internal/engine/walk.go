package engine

import (
	"context"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/redactyl/redactyl/internal/git"
	"github.com/redactyl/redactyl/internal/ignore"
)

// Walk traverses the working tree and invokes handle for each eligible file.
func Walk(ctx context.Context, cfg Config, ign ignore.Matcher, handle func(path string, data []byte)) error {
	return filepath.WalkDir(cfg.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if cfg.DefaultExcludes && isDefaultDirExcluded(name) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(cfg.Root, p)
		if !allowedByGlobs(rel, cfg) {
			return nil
		}
		if ign.Match(rel) {
			return nil
		}
		info, _ := d.Info()
		if info != nil && info.Size() > cfg.MaxBytes {
			return nil
		}
		lower := strings.ToLower(rel)
		if cfg.DefaultExcludes && isDefaultFileExcluded(lower) {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if strings.Contains(string(b), "redactyl:ignore-file") {
			return nil
		}
		if looksBinary(b) || looksNonTextMIME(rel, b) {
			return nil
		}
		handle(rel, b)
		return nil
	})
}

func looksBinary(b []byte) bool {
	const sniff = 800
	n := sniff
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			return true
		}
	}
	return false
}

func looksNonTextMIME(path string, b []byte) bool {
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		if strings.HasPrefix(ct, "image/") || strings.HasPrefix(ct, "video/") || strings.HasPrefix(ct, "audio/") {
			return true
		}
		if strings.Contains(ct, "zip") || strings.Contains(ct, "tar") || strings.Contains(ct, "gzip") {
			return true
		}
	}
	if len(b) >= 4 {
		if len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n" {
			return true
		}
		if b[0] == 'P' && b[1] == 'K' {
			return true
		}
	}
	return false
}

func CountTargets(cfg Config) (int, error) {
	ign, errIgnore := ignore.Load(filepath.Join(cfg.Root, ".redactylignore"))
	if errIgnore != nil {
		return 0, nil
	}
	if cfg.HistoryCommits > 0 {
		entries, err := git.LastNCommits(cfg.Root, cfg.HistoryCommits)
		if err != nil {
			return 0, nil
		}
		n := 0
		for _, e := range entries {
			for path, blob := range e.Files {
				if !allowedByGlobs(path, cfg) {
					continue
				}
				if ign.Match(path) {
					continue
				}
				if int64(len(blob)) > cfg.MaxBytes {
					continue
				}
				n++
			}
		}
		return n, nil
	}
	if cfg.BaseBranch != "" {
		files, data, err := git.DiffAgainst(cfg.Root, cfg.BaseBranch)
		if err != nil {
			return 0, nil
		}
		n := 0
		for i, p := range files {
			if len(data[i]) == 0 { // skip pure deletions/renames with no added lines
				continue
			}
			if !allowedByGlobs(p, cfg) {
				continue
			}
			if ign.Match(p) {
				continue
			}
			if int64(len(data[i])) > cfg.MaxBytes {
				continue
			}
			n++
		}
		return n, nil
	}
	if cfg.ScanStaged {
		files, data, err := git.StagedDiff(cfg.Root)
		if err != nil {
			return 0, nil
		}
		n := 0
		for i, p := range files {
			if !allowedByGlobs(p, cfg) {
				continue
			}
			if ign.Match(p) {
				continue
			}
			if int64(len(data[i])) > cfg.MaxBytes {
				continue
			}
			n++
		}
		return n, nil
	}
	count := 0
	_ = filepath.WalkDir(cfg.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if isDefaultDirExcluded(name) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(cfg.Root, p)
		if !allowedByGlobs(rel, cfg) {
			return nil
		}
		if ign.Match(rel) {
			return nil
		}
		info, _ := d.Info()
		if info != nil && info.Size() > cfg.MaxBytes {
			return nil
		}
		lower := strings.ToLower(rel)
		if isDefaultFileExcluded(lower) {
			return nil
		}
		count++
		return nil
	})
	return count, nil
}
