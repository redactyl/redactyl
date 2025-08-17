package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return p
}

func TestLoadFile_Basic(t *testing.T) {
	dir := t.TempDir()
	p := writeTemp(t, dir, "redactyl.yaml", "threads: 4\nmax_bytes: 123\narchives: true\nscan_time_budget: 5s\nglobal_artifact_budget: 7s\n")
	cfg, err := LoadFile(p)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if cfg.Threads == nil || *cfg.Threads != 4 {
		t.Fatalf("expected threads=4, got %#v", cfg.Threads)
	}
	if cfg.MaxBytes == nil || *cfg.MaxBytes != 123 {
		t.Fatalf("expected max_bytes=123, got %#v", cfg.MaxBytes)
	}
	if cfg.Archives == nil || *cfg.Archives != true {
		t.Fatalf("expected archives=true")
	}
	if cfg.ScanTimeBudget == nil || *cfg.ScanTimeBudget != "5s" {
		t.Fatalf("expected scan_time_budget=5s, got %#v", cfg.ScanTimeBudget)
	}
	if cfg.GlobalArtifactBudget == nil || *cfg.GlobalArtifactBudget != "7s" {
		t.Fatalf("expected global_artifact_budget=7s, got %#v", cfg.GlobalArtifactBudget)
	}
}

func TestGlobalArtifactBudget_Precedence(t *testing.T) {
	// Ensure CLI > local > global precedence for the new field is respected by parsing logic in scan
	// Here we only verify parsing at the config layer and leave CLI precedence to e2e.
	dir := t.TempDir()
	// Global-like file content
	g := "threads: 1\nglobal_artifact_budget: 1s\n"
	p := writeTemp(t, dir, "redactyl.yaml", g)
	cfg, err := LoadFile(p)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if cfg.GlobalArtifactBudget == nil || *cfg.GlobalArtifactBudget != "1s" {
		t.Fatalf("expected global_artifact_budget=1s, got %#v", cfg.GlobalArtifactBudget)
	}
}

func TestLoadLocal_PrefersDotfile(t *testing.T) {
	dir := t.TempDir()
	// place both, expect the dotfile to be picked first by search order
	writeTemp(t, dir, "redactyl.yaml", "threads: 1\n")
	writeTemp(t, dir, ".redactyl.yaml", "threads: 7\n")
	cfg, err := LoadLocal(dir)
	if err != nil {
		t.Fatalf("LoadLocal: %v", err)
	}
	if cfg.Threads == nil || *cfg.Threads != 7 {
		t.Fatalf("expected threads=7 from .redactyl.yaml, got %#v", cfg.Threads)
	}
}

func TestLoadLocal_NoConfig(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadLocal(dir); err == nil {
		t.Fatal("expected error when no local config exists")
	}
}

func TestLoadGlobal_XDG_Config(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "redactyl")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	p := filepath.Join(cfgDir, "config.yml")
	if err := os.WriteFile(p, []byte("threads: 9\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal: %v", err)
	}
	if cfg.Threads == nil || *cfg.Threads != 9 {
		t.Fatalf("expected threads=9 from global config, got %#v", cfg.Threads)
	}
}

func TestLoadGlobal_NoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	// Simulate no HOME as well by clearing HOME; LoadGlobal should error
	t.Setenv("HOME", "")
	if _, err := LoadGlobal(); err == nil {
		t.Fatal("expected error when no global config dir exists")
	}
}
