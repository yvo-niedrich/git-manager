package model

import (
	"testing"
)

func TestCommitsModel_InitialSelectionIsFirst(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	sel := m.Selected()
	if sel == nil || sel.Subject != "third commit" {
		t.Errorf("expected 'third commit', got %+v", sel)
	}
}

func TestCommitsModel_SelectedNilWhenEmpty(t *testing.T) {
	m := NewCommitsModel()
	if m.Selected() != nil {
		t.Error("Selected() should be nil with no commits")
	}
}

func TestCommitsModel_CursorMovesDown(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("j"))
	sel := m.Selected()
	if sel == nil || sel.Subject != "second commit" {
		t.Errorf("expected 'second commit' after j, got %+v", sel)
	}
}

func TestCommitsModel_CursorClampsAtBottom(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	for range makeCommits() {
		m, _ = m.Update(press("j"))
	}
	sel := m.Selected()
	if sel == nil || sel.Subject != "first commit" {
		t.Errorf("expected 'first commit' (last), got %+v", sel)
	}
}

func TestCommitsModel_CursorClampsAtTop(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("k"))
	if m.cursor != 0 {
		t.Errorf("cursor = %d after k at top, want 0", m.cursor)
	}
}

func TestCommitsModel_FilterBySubject(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "second" {
		m, _ = m.Update(press(string(ch)))
	}

	filtered := m.filteredCommits()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 match, got %d", len(filtered))
	}
	if filtered[0].Subject != "second commit" {
		t.Errorf("Subject = %q, want %q", filtered[0].Subject, "second commit")
	}
}

func TestCommitsModel_FilterByAuthor(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "bob" {
		m, _ = m.Update(press(string(ch)))
	}

	filtered := m.filteredCommits()
	if len(filtered) != 1 || filtered[0].Author != "Bob" {
		t.Errorf("expected one commit by Bob, got %+v", filtered)
	}
}

func TestCommitsModel_FilterByHashPrefix(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "cccc" {
		m, _ = m.Update(press(string(ch)))
	}

	filtered := m.filteredCommits()
	if len(filtered) != 1 || filtered[0].Hash != "cccc3333" {
		t.Errorf("expected cccc3333, got %+v", filtered)
	}
}

func TestCommitsModel_FilterRejectsNonMatch(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("/"))
	for _, ch := range "zzznomatch" {
		m, _ = m.Update(press(string(ch)))
	}

	if len(m.filteredCommits()) != 0 {
		t.Error("expected no matches for 'zzznomatch'")
	}
	if m.Selected() != nil {
		t.Error("Selected() should be nil when filter has no matches")
	}
}

func TestCommitsModel_MultiSelectToggle(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())

	if m.IsMultiSelect() {
		t.Error("should not be in multi-select initially")
	}
	m, _ = m.Update(press("s"))
	if !m.IsMultiSelect() {
		t.Error("should be in multi-select after s")
	}
	m, _ = m.Update(press("s"))
	if m.IsMultiSelect() {
		t.Error("should exit multi-select after second s")
	}
}

func TestCommitsModel_MultiSelectMark(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("s")) // enter multi-select
	m, _ = m.Update(press(" ")) // mark first commit
	if !m.selected[0] {
		t.Error("commit at index 0 should be marked after space")
	}

	m, _ = m.Update(press(" ")) // unmark
	if m.selected[0] {
		t.Error("commit at index 0 should be unmarked after second space")
	}
}

func TestCommitsModel_SelectedHashesReturnsMarked(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("s"))
	m, _ = m.Update(press(" ")) // mark index 0 (third commit)
	m, _ = m.Update(press("j"))
	m, _ = m.Update(press(" ")) // mark index 1 (second commit)

	hashes := m.SelectedHashes()
	if len(hashes) != 2 {
		t.Fatalf("expected 2 hashes, got %d: %v", len(hashes), hashes)
	}
	// log order: newest first, so index 0 before index 1
	if hashes[0] != "aaaa1111" || hashes[1] != "bbbb2222" {
		t.Errorf("hashes = %v, want [aaaa1111 bbbb2222]", hashes)
	}
}

func TestCommitsModel_SelectedHashesFallbackToCursor(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	m, _ = m.Update(press("j")) // cursor on second commit

	hashes := m.SelectedHashes()
	if len(hashes) != 1 || hashes[0] != "bbbb2222" {
		t.Errorf("SelectedHashes() = %v, want [bbbb2222]", hashes)
	}
}

func TestCommitsModel_MultiSelectClearedOnExit(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())

	m, _ = m.Update(press("s"))
	m, _ = m.Update(press(" "))
	m, _ = m.Update(press("s")) // exit multi-select

	if len(m.selected) != 0 {
		t.Error("selected set should be cleared when exiting multi-select")
	}
}

func TestCommitsModel_FilterClearsMultiSelect(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())

	m, _ = m.Update(press("s"))
	m, _ = m.Update(press(" "))
	m, _ = m.Update(press("/")) // open filter

	if m.IsMultiSelect() {
		t.Error("multi-select should be cleared when filter opens")
	}
	if len(m.selected) != 0 {
		t.Error("selected set should be cleared when filter opens")
	}
}

func TestCommitsModel_CommitSelectedMsgEmitted(t *testing.T) {
	m := NewCommitsModel()
	m.SetCommits(makeCommits())
	m.SetSize(80, 40)

	_, cmd := m.Update(press("j"))
	msg := cmdMsg(cmd)
	sel, ok := msg.(CommitSelectedMsg)
	if !ok {
		t.Fatalf("expected CommitSelectedMsg, got %T", msg)
	}
	if sel.Commit.Subject != "second commit" {
		t.Errorf("Commit.Subject = %q, want %q", sel.Commit.Subject, "second commit")
	}
}
