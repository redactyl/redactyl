package redactyl

import (
	"testing"
	"time"

	"github.com/varalys/redactyl/internal/config"
)

func strptr(s string) *string { return &s }

func TestResolveBudgets_Precedence(t *testing.T) {
	// Case 1: Only flags set
	b, g := resolveBudgets(5*time.Second, config.FileConfig{}, config.FileConfig{}, 7*time.Second)
	if b != 5*time.Second || g != 7*time.Second {
		t.Fatalf("flags precedence failed: got (%v,%v)", b, g)
	}

	// Case 2: Local overrides global and flags default
	l := config.FileConfig{ScanTimeBudget: strptr("2s"), GlobalArtifactBudget: strptr("3s")}
	b, g = resolveBudgets(5*time.Second, l, config.FileConfig{}, 7*time.Second)
	if b != 2*time.Second || g != 3*time.Second {
		t.Fatalf("local override failed: got (%v,%v)", b, g)
	}

	// Case 3: Global applies when local absent
	gcfg := config.FileConfig{ScanTimeBudget: strptr("9s"), GlobalArtifactBudget: strptr("11s")}
	b, g = resolveBudgets(0, config.FileConfig{}, gcfg, 0)
	if b != 9*time.Second || g != 11*time.Second {
		t.Fatalf("global fallback failed: got (%v,%v)", b, g)
	}
}
