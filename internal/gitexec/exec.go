package gitexec

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

func Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

func Git(ctx context.Context, args ...string) error {
	return Run(ctx, "git", args...)
}

func WithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// DetectFilterRepo returns nil if `git filter-repo` is available in PATH.
func DetectFilterRepo() error {
	ctx, cancel := WithTimeout(2 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "filter-repo", "--help")
	if err := cmd.Run(); err != nil {
		return errors.New("git filter-repo not found. Install from https://github.com/newren/git-filter-repo")
	}
	return nil
}
