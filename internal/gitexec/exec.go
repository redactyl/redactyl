package gitexec

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

// runCommand is a small indirection around exec.CommandContext to enable
// deterministic unit testing without requiring external binaries.
var runCommand = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

func Run(ctx context.Context, name string, args ...string) error {
	return runCommand(ctx, name, args...)
}

func Git(ctx context.Context, args ...string) error {
	return runCommand(ctx, "git", args...)
}

func WithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// DetectFilterRepo returns nil if `git filter-repo` is available in PATH.
func DetectFilterRepo() error {
	ctx, cancel := WithTimeout(2 * time.Second)
	defer cancel()
	if err := runCommand(ctx, "git", "filter-repo", "--help"); err != nil {
		return errors.New("git filter-repo not found. Install from https://github.com/newren/git-filter-repo")
	}
	return nil
}
