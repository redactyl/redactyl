package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/redactyl/redactyl/internal/report"
	"github.com/redactyl/redactyl/internal/types"
)

func withTempDir(t *testing.T, fn func(dir string)) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current wd: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore wd: %v", err)
		}
	}()

	fn(dir)
}

func TestIgnoreFile(t *testing.T) {
	withTempDir(t, func(dir string) {
		findings := []types.Finding{
			{Path: "secret.txt"},
		}
		m := NewModel(findings, nil)
		m.table.SetCursor(0)

		// Execute ignoreFile action
		cmd := m.ignoreFile()
		if cmd == nil {
			t.Fatal("Expected command from ignoreFile")
		}

		// Run the command (it returns a statusMsg)
		cmd()

		// Verify .redactylignore content
		content, err := os.ReadFile(".redactylignore")
		if err != nil {
			t.Fatalf("Failed to read .redactylignore: %v", err)
		}
		if string(content) != "secret.txt\n" {
			t.Errorf("Expected 'secret.txt\n', got %q", string(content))
		}

		// Test unignore
		cmd = m.unignoreFile()
		cmd()

		content, err = os.ReadFile(".redactylignore")
		if err != nil {
			t.Fatalf("Failed to read .redactylignore: %v", err)
		}
		if string(content) != "" {
			t.Errorf("Expected empty .redactylignore, got %q", string(content))
		}
	})
}

func TestAddToBaseline(t *testing.T) {
	withTempDir(t, func(dir string) {
		findings := []types.Finding{
			{Path: "file.go", Detector: "det", Match: "match"},
		}
		m := NewModel(findings, nil)
		m.table.SetCursor(0)

		// Execute addToBaseline
		cmd := m.addToBaseline()
		cmd()

		// Verify baseline file
		base, err := report.LoadBaseline("redactyl.baseline.json")
		if err != nil {
			t.Fatalf("Failed to load baseline: %v", err)
		}

		key := "file.go|det|match"
		if !base.Items[key] {
			t.Error("Finding not found in baseline")
		}

		// Test removeFromBaseline
		// We need to set up the model with knowledge that it is baselined
		m.baselinedSet = map[string]bool{key: true}

		cmd = m.removeFromBaseline()
		cmd()

		base, err = report.LoadBaseline("redactyl.baseline.json")
		if err != nil {
			t.Fatalf("Failed to load baseline: %v", err)
		}
		if base.Items[key] {
			t.Error("Finding should have been removed from baseline")
		}
	})
}

func TestBulkActions(t *testing.T) {
	withTempDir(t, func(dir string) {
		findings := []types.Finding{
			{Path: "f1.go", Detector: "d", Match: "m1"},
			{Path: "f2.go", Detector: "d", Match: "m2"},
		}
		m := NewModel(findings, nil)

		// Select both
		m.selectedFindings[0] = true
		m.selectedFindings[1] = true

		// Bulk Ignore
		cmd := m.bulkIgnore()
		cmd()

		content, err := os.ReadFile(".redactylignore")
		if err != nil {
			t.Fatal(err)
		}
		strContent := string(content)
		// Order is not guaranteed
		if len(strContent) == 0 {
			t.Error(".redactylignore is empty")
		}
		// Cleanup for next test
		os.Remove(".redactylignore")

		// Re-select for bulk baseline (bulkIgnore cleared selection)
		m.selectedFindings[0] = true
		m.selectedFindings[1] = true

		// Bulk Baseline
		cmd = m.bulkBaseline()
		msg := cmd()
		if s, ok := msg.(statusMsg); ok && len(s) > 0 && s != "Added 2 findings to baseline" {
			// If it's a status message but not success (exact text check might be brittle, but helpful for debug)
			// Actually, let's just check if it starts with "Error"
			if string(s)[0:5] == "Error" {
				t.Fatalf("bulkBaseline failed: %s", s)
			}
		}

		base, err := report.LoadBaseline("redactyl.baseline.json")
		if err != nil {
			t.Fatal(err)
		}
		if len(base.Items) != 2 {
			t.Errorf("Expected 2 baselined items, got %d", len(base.Items))
		}
	})
}

func TestExportFindings(t *testing.T) {
	withTempDir(t, func(dir string) {
		findings := []types.Finding{
			{Path: "file.go", Detector: "det", Match: "match", Severity: types.SevHigh},
		}
		m := NewModel(findings, nil)

		// Export JSON
		cmd := m.exportFindings("json")
		msg := cmd() // Run command
		status := msg.(statusMsg)

		if string(status) == "Export error" {
			t.Error("Export returned error")
		}

		// Check for file
		files, _ := filepath.Glob("redactyl-export-*.json")
		if len(files) != 1 {
			t.Error("JSON export file not created")
		} else {
			content, _ := os.ReadFile(files[0])
			var exported []types.Finding
			if err := json.Unmarshal(content, &exported); err != nil {
				t.Error("Failed to unmarshal exported JSON")
			}
			if len(exported) != 1 || exported[0].Path != "file.go" {
				t.Error("Exported content incorrect")
			}
		}

		// Export CSV
		cmd = m.exportFindings("csv")
		cmd()
		files, _ = filepath.Glob("redactyl-export-*.csv")
		if len(files) != 1 {
			t.Error("CSV export file not created")
		}

		// Export SARIF
		cmd = m.exportFindings("sarif")
		cmd()
		files, _ = filepath.Glob("redactyl-export-*.sarif")
		if len(files) != 1 {
			t.Error("SARIF export file not created")
		}
	})
}

func TestVirtualPathExtraction(t *testing.T) {
	// This tests the helpers logic without full integration
	// We are testing parsing logic primarily here as actual extraction
	// requires valid archives which is harder to mock without fixtures

	// Test parsing
	archive, internal := parseVirtualPath("foo.zip::bar.txt")
	if archive != "foo.zip" || internal != "bar.txt" {
		t.Errorf("Failed to parse virtual path: %s, %s", archive, internal)
	}

	// Test nested parsing
	archive, internal = parseVirtualPath("outer.zip::inner.tar::file.txt")
	if archive != "outer.zip" || internal != "inner.tar::file.txt" {
		t.Errorf("Failed to parse nested path: %s, %s", archive, internal)
	}
}

func TestOpenEditor(t *testing.T) {
	// We can't easily test opening an actual editor, but we can test the command generation logic
	// if we extract it. Since it's inside the method, we'll test that it returns a command
	// and doesn't crash.

	findings := []types.Finding{{Path: "file.go", Line: 10}}
	m := NewModel(findings, nil)
	m.table.SetCursor(0)

	cmd := m.openEditor()
	if cmd == nil {
		t.Error("openEditor should return a command")
	}

	// Test virtual file opening
	findings = []types.Finding{{Path: "archive.zip::file.txt"}}
	m = NewModel(findings, nil)
	m.table.SetCursor(0)

	cmd = m.openEditor()
	if cmd == nil {
		t.Error("openEditor (virtual) should return a command")
	}
}

func TestCopyClipboard(t *testing.T) {
	// Note: This might fail if no clipboard is available (e.g. CI/headless)
	// We just check it returns a command.
	findings := []types.Finding{{Path: "file.go"}}
	m := NewModel(findings, nil)
	m.table.SetCursor(0)

	cmd := m.copyPathToClipboard()
	if cmd == nil {
		t.Error("copyPathToClipboard should return a command")
	}

	cmd = m.copyFindingToClipboard()
	if cmd == nil {
		t.Error("copyFindingToClipboard should return a command")
	}
}
