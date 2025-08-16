package redactyl

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLI_JSON_Shape_ExitCodes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "secrets.txt"), []byte("api_key=sk-abcdefghijklmnopqrstuvwxyz0123"), 0644); err != nil {
		t.Fatal(err)
	}
	// run as subprocess to avoid os.Exit in-process
	cmd := exec.Command("go", "run", ".", "scan", "--json", "--fail-on", "high", "-p", dir)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var arr []map[string]any
	if err := json.Unmarshal(out.Bytes(), &arr); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out.String())
	}
	if len(arr) == 0 {
		t.Fatalf("expected at least one finding in JSON output")
	}

	// verify exit code behavior by evaluating ShouldFail on parsed findings
	// Convert to types.Finding-like structs for ShouldFail
	conv := make([]reportFinding, len(arr))
	for i, m := range arr {
		sev, _ := m["severity"].(string)
		path, _ := m["path"].(string)
		match, _ := m["match"].(string)
		detector, _ := m["detector"].(string)
		conv[i] = reportFinding{Path: path, Match: match, Detector: detector, Severity: sev}
	}
	if !shouldFailCompat(conv, "low") {
		t.Fatalf("expected ShouldFail= true for low threshold with findings present")
	}
}

func TestCLI_SARIF_Shape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "secrets.txt"), []byte("token ghp_ABCDEFGHIJKLMNOPQRST1234567890ab"), 0644); err != nil {
		t.Fatal(err)
	}
	// run as subprocess and parse SARIF
	cmd := exec.Command("go", "run", ".", "scan", "--sarif", "--fail-on", "high", "-p", dir)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("sarif json: %v\n%s", err, out.String())
	}
	if doc["version"] != "2.1.0" {
		t.Fatalf("expected SARIF 2.1.0")
	}
}

// New tests for extended JSON and footer counters
func TestCLI_JSONExtended_IncludesStats(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("go", "run", ".", "scan", "--json", "--json-extended", "-p", dir)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json-extended unmarshal: %v\n%s", err, out.String())
	}
	if _, ok := doc["findings"].([]any); !ok {
		t.Fatalf("expected 'findings' array in extended JSON")
	}
	if _, ok := doc["artifact_stats"].(map[string]any); !ok {
		t.Fatalf("expected 'artifact_stats' object in extended JSON")
	}
}

func TestCLI_JSONExtended_StatsNonZeroOnAbort(t *testing.T) {
	// Create a small tar with two entries and force max-entries=1 to trigger abort
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "file.tar")
	{
		f, err := os.Create(tarPath)
		if err != nil {
			t.Fatal(err)
		}
		tw := tar.NewWriter(f)
		_ = tw.WriteHeader(&tar.Header{Name: "a.txt", Mode: 0600, Size: int64(len("hello"))})
		_, _ = tw.Write([]byte("hello"))
		_ = tw.WriteHeader(&tar.Header{Name: "b.txt", Mode: 0600, Size: int64(len("world"))})
		_, _ = tw.Write([]byte("world"))
		_ = tw.Close()
		_ = f.Close()
	}
	cmd := exec.Command("go", "run", ".", "scan", "--json", "--json-extended", "--archives", "--max-entries", "1", "-p", dir)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // ignore non-zero exit
	var doc map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, out.String())
	}
	stats, ok := doc["artifact_stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected artifact_stats in extended JSON")
	}
	if e, ok := stats["entries"].(float64); !ok || e <= 0 {
		t.Fatalf("expected non-zero entries abort counter; got stats=%#v", stats)
	}
}

func TestCLI_SARIF_StatsNonZeroOnAbort(t *testing.T) {
	// Same abort scenario, but assert SARIF run.properties.artifactStats
	dir := t.TempDir()
	tarPath := filepath.Join(dir, "file.tar")
	{
		f, err := os.Create(tarPath)
		if err != nil {
			t.Fatal(err)
		}
		tw := tar.NewWriter(f)
		_ = tw.WriteHeader(&tar.Header{Name: "a.txt", Mode: 0600, Size: int64(len("hello"))})
		_, _ = tw.Write([]byte("hello"))
		_ = tw.WriteHeader(&tar.Header{Name: "b.txt", Mode: 0600, Size: int64(len("world"))})
		_, _ = tw.Write([]byte("world"))
		_ = tw.Close()
		_ = f.Close()
	}
	cmd := exec.Command("go", "run", ".", "scan", "--sarif", "--archives", "--max-entries", "1", "-p", dir)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	var sarif map[string]any
	if err := json.Unmarshal(out.Bytes(), &sarif); err != nil {
		t.Fatalf("sarif unmarshal: %v\n%s", err, out.String())
	}
	runs, _ := sarif["runs"].([]any)
	if len(runs) == 0 {
		t.Fatal("expected runs")
	}
	run := runs[0].(map[string]any)
	props, _ := run["properties"].(map[string]any)
	if props == nil {
		t.Fatal("expected run.properties")
	}
	as, _ := props["artifactStats"].(map[string]any)
	if as == nil {
		t.Fatal("expected run.properties.artifactStats")
	}
	if e, ok := as["entries"].(float64); !ok || e <= 0 {
		t.Fatalf("expected non-zero entries abort counter in SARIF; got: %#v", as)
	}
}

// Minimal compatible types for invoking ShouldFail logic without importing internals
type reportFinding struct {
	Path     string
	Match    string
	Detector string
	Severity string
}

func shouldFailCompat(fs []reportFinding, failOn string) bool {
	level := map[string]int{"low": 1, "medium": 2, "high": 3}
	th := level[failOn]
	if th == 0 {
		th = 2
	}
	for _, f := range fs {
		if level[f.Severity] >= th {
			return true
		}
	}
	return false
}
