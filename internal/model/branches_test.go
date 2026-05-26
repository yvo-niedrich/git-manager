package model

import (
	"testing"
)

func TestBranchesModel_InitialSelectionIsFirst(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	sel := m.Selected()
	if sel == nil {
		t.Fatal("Selected() returned nil")
	}
	if sel.Name != "main" || sel.IsRemote {
		t.Errorf("expected local 'main', got %+v", sel)
	}
}

func TestBranchesModel_SelectedNilWhenEmpty(t *testing.T) {
	m := NewBranchesModel()
	if m.Selected() != nil {
		t.Error("Selected() should be nil with no branches")
	}
}

func TestBranchesModel_CursorMovesDown(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("j"))
	sel := m.Selected()
	if sel == nil || sel.Name != "feature/a" {
		t.Errorf("expected 'feature/a' after j, got %+v", sel)
	}
}

func TestBranchesModel_CursorClampsAtBottom(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	for range makeBranches() {
		m, _ = m.Update(press("j"))
	}
	sel := m.Selected()
	last := makeBranches()[len(makeBranches())-1]
	if sel == nil || sel.Name != last.Name {
		t.Errorf("cursor should clamp at last item, got %+v", sel)
	}
}

func TestBranchesModel_CursorMovesUp(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("j"))
	m, _ = m.Update(press("k"))
	sel := m.Selected()
	if sel == nil || sel.Name != "main" || sel.IsRemote {
		t.Errorf("expected local 'main' after j then k, got %+v", sel)
	}
}

func TestBranchesModel_CursorClampsAtTop(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("k"))
	if m.cursor != 0 {
		t.Errorf("cursor = %d after k at top, want 0", m.cursor)
	}
}

func TestBranchesModel_FilterNarrows(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "feat" {
		m, _ = m.Update(press(string(ch)))
	}

	sel := m.Selected()
	if sel == nil {
		t.Fatal("Selected() returned nil after filtering for 'feat'")
	}
	if sel.Name != "feature/a" {
		t.Errorf("Selected().Name = %q, want %q", sel.Name, "feature/a")
	}
}

func TestBranchesModel_FilterResetsCursorOnChange(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	// Move cursor down then open filter — cursor should reset
	m, _ = m.Update(press("j"))
	m, _ = m.Update(press("/"))
	if m.cursor != 0 {
		t.Errorf("cursor = %d after opening filter, want 0", m.cursor)
	}

	// Type a character — cursor should stay at 0
	m, _ = m.Update(press("m"))
	if m.cursor != 0 {
		t.Errorf("cursor = %d after first filter char, want 0", m.cursor)
	}
}

func TestBranchesModel_FilterEscRestores(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "feat" {
		m, _ = m.Update(press(string(ch)))
	}
	m, _ = m.Update(press("esc"))

	if m.IsFiltering() {
		t.Error("IsFiltering() should be false after esc")
	}
	if m.filterInput.Value() != "" {
		t.Errorf("filter value = %q after esc, want empty", m.filterInput.Value())
	}
	// All branches visible again
	if len(m.filteredBranches()) != len(makeBranches()) {
		t.Errorf("filteredBranches len = %d, want %d", len(m.filteredBranches()), len(makeBranches()))
	}
}

func TestBranchesModel_FilterEnterCommits(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "feat" {
		m, _ = m.Update(press(string(ch)))
	}
	m, _ = m.Update(press("enter"))

	if m.IsFiltering() {
		t.Error("IsFiltering() should be false after enter")
	}
	// Filter value is preserved
	if m.filterInput.Value() != "feat" {
		t.Errorf("filter value = %q after enter, want %q", m.filterInput.Value(), "feat")
	}
	// Only matching branches visible
	if len(m.filteredBranches()) != 1 {
		t.Errorf("filteredBranches len = %d, want 1 (only feature/a)", len(m.filteredBranches()))
	}
}

func TestBranchesModel_BranchSelectedMsgEmitted(t *testing.T) {
	m := NewBranchesModel()
	m.SetBranches(makeBranches())
	m.SetSize(80, 40)

	_, cmd := m.Update(press("j"))
	msg := cmdMsg(cmd)
	sel, ok := msg.(BranchSelectedMsg)
	if !ok {
		t.Fatalf("expected BranchSelectedMsg, got %T", msg)
	}
	if sel.Branch.Name != "feature/a" {
		t.Errorf("Branch.Name = %q, want %q", sel.Branch.Name, "feature/a")
	}
}
