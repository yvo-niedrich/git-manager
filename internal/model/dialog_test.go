package model

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ── dialogStack ───────────────────────────────────────────────────────────────

func TestDialogStack_EmptyNotOpen(t *testing.T) {
	var ds dialogStack
	if ds.IsOpen() {
		t.Error("empty stack should not be open")
	}
	if ds.Active() != nil {
		t.Error("Active() should return nil on empty stack")
	}
}

func TestDialogStack_OpenAfterPush(t *testing.T) {
	var ds dialogStack
	ds.Push(NewContextMenu(BranchMenuItems(false, false, "")))
	if !ds.IsOpen() {
		t.Error("stack should be open after Push")
	}
}

func TestDialogStack_ActiveIsHighestPriority(t *testing.T) {
	var ds dialogStack
	ds.Push(NewContextMenu(BranchMenuItems(false, false, ""))) // priority 10
	ds.Push(NewConfirmDialog("sure?", func() tea.Cmd { return nil })) // priority 20

	if _, ok := ds.Active().(*ConfirmModel); !ok {
		t.Errorf("Active() = %T, want *ConfirmModel (priority 20)", ds.Active())
	}
}

func TestDialogStack_UpdateRoutesToActive(t *testing.T) {
	var ds dialogStack
	ds.Push(NewContextMenu(BranchMenuItems(false, false, ""))) // priority 10
	confirm := NewConfirmDialog("sure?", func() tea.Cmd { return nil })
	ds.Push(confirm) // priority 20

	// pressing 'n' cancels the confirm — the context menu should be untouched
	ds.Update(press("n"))

	if !ds.IsOpen() {
		t.Error("context menu should remain after confirm closes")
	}
	if _, ok := ds.Active().(ContextMenuModel); !ok {
		t.Errorf("Active() = %T, want ContextMenuModel after confirm closes", ds.Active())
	}
}

func TestDialogStack_CloseRemovesFromStack(t *testing.T) {
	var ds dialogStack
	ds.Push(NewContextMenu(BranchMenuItems(false, false, "")))

	ds.Update(press("esc"))

	if ds.IsOpen() {
		t.Error("stack should be empty after dialog closes via esc")
	}
}

func TestDialogStack_LowerPriorityRemainsAfterHighCloses(t *testing.T) {
	var ds dialogStack
	ds.Push(NewContextMenu(BranchMenuItems(false, false, ""))) // priority 10
	ds.Push(NewConfirmDialog("sure?", func() tea.Cmd { return nil })) // priority 20

	ds.Update(press("n")) // close confirm

	if !ds.IsOpen() {
		t.Error("context menu (priority 10) should remain open")
	}
}

// ── ContextMenuModel ──────────────────────────────────────────────────────────

func TestContextMenu_EscCloses(t *testing.T) {
	m := NewContextMenu(BranchMenuItems(false, false, ""))
	next, cmd := m.DialogUpdate(press("esc"))
	if next != nil {
		t.Error("esc: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("esc: expected nil cmd")
	}
}

func TestContextMenu_QCloses(t *testing.T) {
	m := NewContextMenu(BranchMenuItems(false, false, ""))
	next, _ := m.DialogUpdate(press("q"))
	if next != nil {
		t.Error("q: expected nil content (close signal)")
	}
}

func TestContextMenu_EnterSelectsCursorItem(t *testing.T) {
	m := NewContextMenu(BranchMenuItems(false, false, "")) // first item: Checkout
	next, cmd := m.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter: expected nil content (close signal)")
	}
	msg := cmdMsg(cmd)
	sel, ok := msg.(MenuSelectedMsg)
	if !ok {
		t.Fatalf("expected MenuSelectedMsg, got %T", msg)
	}
	if sel.Action != ActionCheckout {
		t.Errorf("Action = %v, want ActionCheckout", sel.Action)
	}
}

func TestContextMenu_ShortcutKey(t *testing.T) {
	m := NewContextMenu(BranchMenuItems(false, false, ""))
	next, cmd := m.DialogUpdate(press("m")) // 'm' → ActionMerge
	if next != nil {
		t.Error("shortcut: expected nil content (close signal)")
	}
	msg := cmdMsg(cmd)
	sel, ok := msg.(MenuSelectedMsg)
	if !ok {
		t.Fatalf("expected MenuSelectedMsg, got %T", msg)
	}
	if sel.Action != ActionMerge {
		t.Errorf("Action = %v, want ActionMerge", sel.Action)
	}
}

func TestContextMenu_CursorNavigation(t *testing.T) {
	m := NewContextMenu(BranchMenuItems(false, false, ""))
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}

	next, _ := m.DialogUpdate(press("j"))
	m = next.(ContextMenuModel)
	if m.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m.cursor)
	}

	next, _ = m.DialogUpdate(press("k"))
	m = next.(ContextMenuModel)
	if m.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m.cursor)
	}
}

// ── ConfirmModel ──────────────────────────────────────────────────────────────

func TestConfirmModel_YesCallsFn(t *testing.T) {
	called := false
	m := NewConfirmDialog("sure?", func() tea.Cmd {
		called = true
		return nil
	})

	next, _ := m.DialogUpdate(press("y"))

	if next != nil {
		t.Error("y: expected nil content (close signal)")
	}
	if !called {
		t.Error("confirmFn was not called on y")
	}
}

func TestConfirmModel_EnterCallsFn(t *testing.T) {
	called := false
	m := NewConfirmDialog("sure?", func() tea.Cmd {
		called = true
		return nil
	})
	m.DialogUpdate(press("enter"))
	if !called {
		t.Error("confirmFn was not called on enter")
	}
}

func TestConfirmModel_NoDoesNotCallFn(t *testing.T) {
	called := false
	m := NewConfirmDialog("sure?", func() tea.Cmd {
		called = true
		return nil
	})

	next, cmd := m.DialogUpdate(press("n"))

	if next != nil {
		t.Error("n: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("n: expected nil cmd")
	}
	if called {
		t.Error("confirmFn should not be called on n")
	}
}

func TestConfirmModel_EscDoesNotCallFn(t *testing.T) {
	called := false
	m := NewConfirmDialog("sure?", func() tea.Cmd {
		called = true
		return nil
	})
	m.DialogUpdate(press("esc"))
	if called {
		t.Error("confirmFn should not be called on esc")
	}
}

// ── AmendModel ────────────────────────────────────────────────────────────────

func TestAmendModel_Priority(t *testing.T) {
	m := NewAmendDialog("msg")
	if got := m.Priority(); got != 20 {
		t.Errorf("Priority() = %d, want 20", got)
	}
}

func TestAmendModel_EnterEmitsSubmit(t *testing.T) {
	m := NewAmendDialog("original message")
	next, cmd := m.DialogUpdate(press("enter"))

	if next != nil {
		t.Error("enter: expected nil content (close signal)")
	}
	msg := cmdMsg(cmd)
	submit, ok := msg.(AmendSubmitMsg)
	if !ok {
		t.Fatalf("expected AmendSubmitMsg, got %T", msg)
	}
	if submit.NewMessage != "original message" {
		t.Errorf("NewMessage = %q, want %q", submit.NewMessage, "original message")
	}
}

func TestAmendModel_EscCloses(t *testing.T) {
	m := NewAmendDialog("original message")
	next, cmd := m.DialogUpdate(press("esc"))

	if next != nil {
		t.Error("esc: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("esc: expected nil cmd (no message emitted)")
	}
}

func TestAmendModel_IgnoresUnknownKeys(t *testing.T) {
	m := NewAmendDialog("hello")
	_ = errors.New("unused") // ensure errors import doesn't get flagged as unused
	next, _ := m.DialogUpdate(press("tab"))
	if next == nil {
		t.Error("tab should not close the amend dialog")
	}
}
