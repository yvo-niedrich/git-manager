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

// ── BranchMenuItems ───────────────────────────────────────────────────────────

func hasAction(items []MenuItem, action MenuAction) bool {
	for _, it := range items {
		if it.Action == action {
			return true
		}
	}
	return false
}

func TestBranchMenuItems_CurrentBranchExcludesCheckoutDelete(t *testing.T) {
	items := BranchMenuItems(false, true, "")
	// Checkout and Delete cannot target the branch you are on. Merge and Rebase
	// remain available: the branch picker chooses the counterpart branch, so
	// "merge current into <target>" / "rebase current onto <target>" are valid.
	for _, banned := range []MenuAction{ActionCheckout, ActionDeleteBranch} {
		if hasAction(items, banned) {
			t.Errorf("current-branch menu should not contain action %v", banned)
		}
	}
	for _, want := range []MenuAction{ActionMerge, ActionRebase} {
		if !hasAction(items, want) {
			t.Errorf("current-branch menu should contain action %v (picker selects the target)", want)
		}
	}
}

func TestBranchMenuItems_NonCurrentBranchIncludesMergeRebaseDelete(t *testing.T) {
	items := BranchMenuItems(false, false, "")
	for _, want := range []MenuAction{ActionCheckout, ActionMerge, ActionRebase, ActionDeleteBranch} {
		if !hasAction(items, want) {
			t.Errorf("non-current-branch menu should contain action %v", want)
		}
	}
}

func TestBranchMenuItems_CurrentBranchRetainsPushForcePush(t *testing.T) {
	items := BranchMenuItems(false, true, "")
	for _, want := range []MenuAction{ActionPush, ActionForcePush} {
		if !hasAction(items, want) {
			t.Errorf("current-branch menu should contain action %v", want)
		}
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

func TestConfirmModelWithSubject_YesCallsFn(t *testing.T) {
	called := false
	m := NewConfirmDialogWithSubject("Delete branch?", "feature/x", func() tea.Cmd {
		called = true
		return nil
	})
	m.DialogUpdate(press("y"))
	if !called {
		t.Error("confirmFn was not called on y")
	}
}

// ── NewBranchModel ────────────────────────────────────────────────────────────

func TestNewBranchModel_EscCloses(t *testing.T) {
	m := NewBranchDialog("main", nil)
	next, cmd := m.DialogUpdate(press("esc"))
	if next != nil {
		t.Error("esc: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("esc: expected nil cmd")
	}
}

func TestNewBranchModel_EmptyNameShowsError(t *testing.T) {
	m := NewBranchDialog("main", nil)
	next, cmd := m.DialogUpdate(press("enter"))
	if next == nil {
		t.Fatal("enter with empty name: dialog should stay open")
	}
	if cmd != nil {
		t.Error("enter with empty name: expected nil cmd")
	}
	nb, ok := next.(*NewBranchModel)
	if !ok {
		t.Fatalf("next = %T, want *NewBranchModel", next)
	}
	if nb.errMsg == "" {
		t.Error("errMsg should be set after empty-name submit")
	}
}

func TestNewBranchModel_DuplicateNameShowsError(t *testing.T) {
	var dc DialogContent = NewBranchDialog("main", []string{"main", "feature/a"})
	for _, ch := range "main" {
		dc, _ = dc.DialogUpdate(press(string(ch)))
	}
	dc, _ = dc.DialogUpdate(press("enter"))
	nb, ok := dc.(*NewBranchModel)
	if !ok {
		t.Fatalf("dialog should stay open on duplicate, got %T", dc)
	}
	if nb.errMsg == "" {
		t.Error("errMsg should be set for duplicate branch name")
	}
}

func TestNewBranchModel_ValidNameEmitsSubmit(t *testing.T) {
	m := NewBranchDialog("main", []string{"main"})
	var dc DialogContent = m
	for _, ch := range "new-feature" {
		dc, _ = dc.DialogUpdate(press(string(ch)))
	}
	next, cmd := dc.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter with valid name: dialog should close (nil content)")
	}
	msg := cmdMsg(cmd)
	submit, ok := msg.(NewBranchSubmitMsg)
	if !ok {
		t.Fatalf("expected NewBranchSubmitMsg, got %T", msg)
	}
	if submit.Name != "new-feature" {
		t.Errorf("Name = %q, want %q", submit.Name, "new-feature")
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

// ── BranchPickerModel ─────────────────────────────────────────────────────────

func newTestPicker(action MenuAction) *BranchPickerModel {
	localNames := []string{"main", "develop", "feature/a", "feature/b", "release/v1.0"}
	return NewBranchPickerDialog(action, "feature/a", "main", localNames, 40)
}

func TestBranchPicker_ExcludesSource(t *testing.T) {
	m := newTestPicker(ActionMerge)
	for _, name := range m.branches {
		if name == "feature/a" {
			t.Error("source branch 'feature/a' should be excluded")
		}
	}
}

func TestBranchPicker_DefaultsToCurrentBranch(t *testing.T) {
	// current="main" is in the list and should be pre-selected.
	m := newTestPicker(ActionMerge)
	if m.branches[m.cursor] != "main" {
		t.Errorf("cursor branch = %q, want %q", m.branches[m.cursor], "main")
	}
}

func TestBranchPicker_EscWithFilterClearsFilterOnly(t *testing.T) {
	m := newTestPicker(ActionMerge)
	// Type something into the filter.
	for _, ch := range "feat" {
		m.DialogUpdate(press(string(ch)))
	}
	// Esc should clear the filter but keep the dialog open.
	next, cmd := m.DialogUpdate(press("esc"))
	if next == nil {
		t.Fatal("esc with active filter should keep dialog open")
	}
	if cmd != nil {
		t.Error("expected nil cmd when clearing filter")
	}
	picker := next.(*BranchPickerModel)
	if picker.filter.Value() != "" {
		t.Errorf("filter value = %q after esc, want empty", picker.filter.Value())
	}
}

func TestBranchPicker_EscWithEmptyFilterCloses(t *testing.T) {
	m := newTestPicker(ActionMerge)
	next, cmd := m.DialogUpdate(press("esc"))
	if next != nil {
		t.Error("esc with empty filter should close dialog (nil content)")
	}
	if cmd != nil {
		t.Error("expected nil cmd on cancel")
	}
}

func TestBranchPicker_EnterEmitsSubmit(t *testing.T) {
	m := newTestPicker(ActionMerge)
	// cursor starts at "main" (current branch, pre-selected).
	next, cmd := m.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter should close the picker dialog")
	}
	msg := cmdMsg(cmd)
	sub, ok := msg.(BranchPickerSubmitMsg)
	if !ok {
		t.Fatalf("expected BranchPickerSubmitMsg, got %T", msg)
	}
	if sub.Action != ActionMerge {
		t.Errorf("Action = %v, want ActionMerge", sub.Action)
	}
	if sub.Source != "feature/a" {
		t.Errorf("Source = %q, want %q", sub.Source, "feature/a")
	}
	if sub.Target != "main" {
		t.Errorf("Target = %q, want %q", sub.Target, "main")
	}
}

func TestBranchPicker_UpDownNavigation(t *testing.T) {
	m := newTestPicker(ActionMerge)
	// cursor starts at "main" (current branch, pre-selected).
	startCursor := m.cursor

	next, _ := m.DialogUpdate(press("down"))
	m = next.(*BranchPickerModel)
	if m.cursor != startCursor+1 {
		t.Errorf("cursor after down = %d, want %d", m.cursor, startCursor+1)
	}

	next, _ = m.DialogUpdate(press("up"))
	m = next.(*BranchPickerModel)
	if m.cursor != startCursor {
		t.Errorf("cursor after up = %d, want %d", m.cursor, startCursor)
	}
}

func TestBranchPicker_UpDoesNotGoNegative(t *testing.T) {
	m := newTestPicker(ActionMerge)
	m.cursor = 0
	next, _ := m.DialogUpdate(press("up"))
	if next.(*BranchPickerModel).cursor != 0 {
		t.Error("cursor should not go below 0")
	}
}

func TestBranchPicker_FilterNarrowsList(t *testing.T) {
	m := newTestPicker(ActionMerge)
	// Type "feat" — should match "feature/b" only (feature/a is excluded).
	for _, ch := range "feat" {
		next, _ := m.DialogUpdate(press(string(ch)))
		m = next.(*BranchPickerModel)
	}
	filtered := m.filteredBranches()
	if len(filtered) != 1 || filtered[0] != "feature/b" {
		t.Errorf("filtered = %v, want [feature/b]", filtered)
	}
}

func TestBranchPicker_FilterResetsCursor(t *testing.T) {
	m := newTestPicker(ActionMerge)
	m.cursor = 2
	next, _ := m.DialogUpdate(press("f"))
	m = next.(*BranchPickerModel)
	if m.cursor != 0 {
		t.Errorf("cursor after filter change = %d, want 0", m.cursor)
	}
}

func TestBranchPicker_EnterOnEmptyFilteredListIsNoop(t *testing.T) {
	m := newTestPicker(ActionMerge)
	for _, ch := range "zzz" {
		next, _ := m.DialogUpdate(press(string(ch)))
		m = next.(*BranchPickerModel)
	}
	if len(m.filteredBranches()) != 0 {
		t.Skip("expected empty filtered list")
	}
	next, cmd := m.DialogUpdate(press("enter"))
	if next == nil {
		t.Error("enter on empty list should keep dialog open")
	}
	if cmd != nil {
		t.Error("expected nil cmd on enter with empty list")
	}
}
