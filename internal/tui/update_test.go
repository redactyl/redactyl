package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/redactyl/redactyl/internal/types"
)

// TestUpdate_Navigation tests basic navigation keys (j, k, down, up)
func TestUpdate_Navigation(t *testing.T) {
	findings := []types.Finding{
		{Path: "file1.go"},
		{Path: "file2.go"},
		{Path: "file3.go"},
	}

	m := NewModel(findings, nil)
	m.ready = true
	m.height = 20
	m.width = 80

	// Helper to send key press
	sendKey := func(key string) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		if len(key) > 1 {
			switch key {
			case "down":
				msg = tea.KeyMsg{Type: tea.KeyDown}
			case "up":
				msg = tea.KeyMsg{Type: tea.KeyUp}
			}
		}
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(Model)
	}

	// Initial cursor should be 0
	if m.table.Cursor() != 0 {
		t.Errorf("Initial cursor should be 0, got %d", m.table.Cursor())
	}

	// Move down (j)
	sendKey("j")
	if m.table.Cursor() != 1 {
		t.Errorf("After 'j', cursor should be 1, got %d", m.table.Cursor())
	}

	// Move down (down arrow)
	sendKey("down")
	if m.table.Cursor() != 2 {
		t.Errorf("After 'down', cursor should be 2, got %d", m.table.Cursor())
	}

	// Move up (k)
	sendKey("k")
	if m.table.Cursor() != 1 {
		t.Errorf("After 'k', cursor should be 1, got %d", m.table.Cursor())
	}

	// Move up (up arrow)
	sendKey("up")
	if m.table.Cursor() != 0 {
		t.Errorf("After 'up', cursor should be 0, got %d", m.table.Cursor())
	}
}

// TestUpdate_Help tests toggling the help menu
func TestUpdate_Help(t *testing.T) {
	m := NewModel([]types.Finding{{Path: "foo"}}, nil)

	if m.showHelp {
		t.Error("Help should initially be hidden")
	}

	// Press '?' to show help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if !m.showHelp {
		t.Error("Help should be shown after pressing '?'")
	}

	// Press '?' again to hide help
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.showHelp {
		t.Error("Help should be hidden after pressing '?' again")
	}
}

// TestUpdate_Quit tests the quit command
func TestUpdate_Quit(t *testing.T) {
	m := NewModel([]types.Finding{{Path: "foo"}}, nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command, got nil")
	}
	// Note: We can't easily verify it's tea.Quit without comparing function pointers or running it
}

// TestUpdate_Search tests entering search mode
func TestUpdate_Search(t *testing.T) {
	m := NewModel([]types.Finding{{Path: "foo"}}, nil)

	if m.searchMode {
		t.Error("Search mode should initially be false")
	}

	// Press '/' to enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if !m.searchMode {
		t.Error("Search mode should be true after pressing '/'")
	}

	// Type "abc"
	for _, char := range "abc" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(Model)
	}

	if m.searchInput.Value() != "abc" {
		t.Errorf("Expected search input 'abc', got '%s'", m.searchInput.Value())
	}

	// Press Enter to confirm search
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.searchMode {
		t.Error("Search mode should be false after Enter")
	}
	if m.searchQuery != "abc" {
		t.Errorf("Expected search query 'abc', got '%s'", m.searchQuery)
	}
}

// TestUpdate_FilterSeverity tests filtering by severity keys
func TestUpdate_FilterSeverity(t *testing.T) {
	findings := []types.Finding{
		{Path: "high.go", Severity: types.SevHigh},
		{Path: "med.go", Severity: types.SevMed},
	}
	m := NewModel(findings, nil)

	// Press '1' for High severity
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.severityFilter != types.SevHigh {
		t.Errorf("Expected severity filter %s, got %s", types.SevHigh, m.severityFilter)
	}
	if len(m.getDisplayFindings()) != 1 {
		t.Errorf("Expected 1 finding, got %d", len(m.getDisplayFindings()))
	}

	// Press 'Esc' to clear filter
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.severityFilter != "" {
		t.Error("Severity filter should be cleared")
	}
	if len(m.getDisplayFindings()) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(m.getDisplayFindings()))
	}
}

// TestUpdate_Sort tests sorting keys
func TestUpdate_Sort(t *testing.T) {
	findings := []types.Finding{
		{Path: "b.go", Severity: types.SevHigh},
		{Path: "a.go", Severity: types.SevHigh},
	}
	m := NewModel(findings, nil)

	// Press 's' to cycle sort
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	
	// Default -> Severity
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	if m.sortColumn != SortSeverity {
		t.Errorf("Expected sort column %s, got %s", SortSeverity, m.sortColumn)
	}

	// Severity -> Path
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	if m.sortColumn != SortPath {
		t.Errorf("Expected sort column %s, got %s", SortPath, m.sortColumn)
	}
	
	// Verify it's sorted by path
	display := m.getDisplayFindings()
	if display[0].Path != "a.go" {
		t.Error("Expected findings to be sorted by path")
	}

	// Press 'S' to reverse sort
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	if !m.sortReverse {
		t.Error("Expected sort reverse to be true")
	}
	display = m.getDisplayFindings()
	if display[0].Path != "b.go" {
		t.Error("Expected findings to be reverse sorted by path")
	}
}

// TestUpdate_Selection tests selection keys
func TestUpdate_Selection(t *testing.T) {
	findings := []types.Finding{
		{Path: "a.go"},
		{Path: "b.go"},
	}
	m := NewModel(findings, nil)
	m.ready = true // needed for cursor

	// Press 'v' to select current
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.getSelectedCount() != 1 {
		t.Errorf("Expected 1 selected finding, got %d", m.getSelectedCount())
	}

	// Press 'V' to select all
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("V")}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	
	// Since one was already selected, 'V' might toggle. 
	// The logic is: if all are selected -> deselect all, else select all.
	// Here only 1/2 selected, so should select all (2).
	if m.getSelectedCount() != 2 {
		t.Errorf("Expected 2 selected findings, got %d", m.getSelectedCount())
	}
}

// TestUpdate_Context tests context expansion keys
func TestUpdate_Context(t *testing.T) {
	m := NewModel([]types.Finding{{Path: "foo"}}, nil)
	initialContext := m.contextLines

	// Press '+' to expand context
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.contextLines <= initialContext {
		t.Error("Context lines should have increased")
	}

	// Press '-' to contract context
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.contextLines != initialContext {
		t.Error("Context lines should have returned to initial")
	}
}

// TestUpdate_Grouping tests grouping keys
func TestUpdate_Grouping(t *testing.T) {
	findings := []types.Finding{
		{Path: "file.go", Detector: "det1"},
	}
	m := NewModel(findings, nil)

	// Press 'g' then 'f' (group by file)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.pendingKey != "g" {
		t.Error("Expected pending key 'g'")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.groupMode != GroupByFile {
		t.Errorf("Expected group mode %s, got %s", GroupByFile, m.groupMode)
	}
}

// TestUpdate_Rescan tests rescan key
func TestUpdate_Rescan(t *testing.T) {
	rescanCalled := false
	rescanFunc := func() ([]types.Finding, error) {
		rescanCalled = true
		return []types.Finding{{Path: "new.go"}}, nil
	}

	m := NewModel([]types.Finding{}, rescanFunc)

	// Press 'r' to rescan
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updatedModel, cmd := m.Update(msg)
	m = updatedModel.(Model)

	if !m.scanning {
		t.Error("Expected scanning to be true")
	}
	if cmd == nil {
		t.Error("Expected rescan command")
	}

	// Execute the command (simulated)
	resultMsg := cmd()
	if _, ok := resultMsg.(findingsMsg); !ok && !rescanCalled {
		// The anonymous function in rescan returns findingsMsg on success,
		// which calls rescanFunc.
		// We can't easily check rescanCalled because it happens inside the cmd execution
		// which is a closure.
		// However, if we execute the command returned:
	}
	
	if !rescanCalled {
		t.Error("rescanFunc was not called")
	}
}

// TestUpdate_ExportMenu tests export menu toggle
func TestUpdate_ExportMenu(t *testing.T) {
	m := NewModel([]types.Finding{{Path: "foo"}}, nil)

	// Press 'e' to show export menu
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if !m.showExportMenu {
		t.Error("Export menu should be shown")
	}

	// Press 'esc' to hide
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	if m.showExportMenu {
		t.Error("Export menu should be hidden")
	}
}
