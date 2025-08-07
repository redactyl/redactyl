package engine

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/you/redactyl/internal/ignore"
)

type fileJob struct {
	path string
	data []byte
}

type workerPool struct {
	ch   chan fileJob
	done chan struct{}
}

func newWorkerPool(n int) *workerPool {
	wp := &workerPool{ch: make(chan fileJob, n*2), done: make(chan struct{})}
	for i := 0; i < n; i++ {
		go func() {
			for range wp.ch {
			}
		}()
	}
	return wp
}
func (w *workerPool) Submit(j fileJob) { w.ch <- j }
func (w *workerPool) Wait()            { close(w.ch); <-w.done }

func Walk(ctx context.Context, cfg Config, ign ignore.Matcher, wp *workerPool, handle func(path string, data []byte)) error {
	defer close(wp.done)
	return filepath.WalkDir(cfg.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".git") || name == "node_modules" || name == "target" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(cfg.Root, p)
		if ign.Match(rel) {
			return nil
		}
		info, _ := d.Info()
		if info != nil && info.Size() > cfg.MaxBytes {
			return nil
		}
		// crude binary skip
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if looksBinary(b) {
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
