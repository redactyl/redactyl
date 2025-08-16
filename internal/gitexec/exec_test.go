package gitexec

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRun_DelegatesToRunner(t *testing.T) {
	called := false
	var gotName string
	var gotArgs []string
	old := runCommand
	t.Cleanup(func() { runCommand = old })
	runCommand = func(ctx context.Context, name string, args ...string) error {
		called = true
		gotName = name
		gotArgs = args
		return nil
	}
	if err := Run(context.Background(), "echo", "hello"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !called || gotName != "echo" || len(gotArgs) != 1 || gotArgs[0] != "hello" {
		t.Fatalf("delegation mismatch: called=%v name=%q args=%v", called, gotName, gotArgs)
	}
}

func TestGit_UsesGitBinary(t *testing.T) {
	called := false
	var gotName string
	var gotArgs []string
	old := runCommand
	t.Cleanup(func() { runCommand = old })
	runCommand = func(ctx context.Context, name string, args ...string) error {
		called = true
		gotName = name
		gotArgs = args
		return nil
	}
	if err := Git(context.Background(), "--version"); err != nil {
		t.Fatalf("Git returned error: %v", err)
	}
	if !called || gotName != "git" || len(gotArgs) != 1 || gotArgs[0] != "--version" {
		t.Fatalf("expected git binary and args; got name=%q args=%v", gotName, gotArgs)
	}
}

func TestDetectFilterRepo_MapsError(t *testing.T) {
	old := runCommand
	t.Cleanup(func() { runCommand = old })
	runCommand = func(ctx context.Context, name string, args ...string) error {
		return errors.New("not found")
	}
	if err := DetectFilterRepo(); err == nil {
		t.Fatal("expected error when runner fails")
	}
}

func TestWithTimeout_Cancels(t *testing.T) {
	ctx, cancel := WithTimeout(10 * time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		// ok, might already be done on slow machines
	case <-time.After(50 * time.Millisecond):
		if ctx.Err() == nil {
			t.Fatal("expected context to be canceled within timeout")
		}
	}
}
