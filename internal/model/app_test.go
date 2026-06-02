package model

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/git"
)

// ── focus ─────────────────────────────────────────────────────────────────────

func TestApp_InitialFocusIsBranches(t *testing.T) {
	a := fixtureApp(t)
	if a.focus != panelBranches {
		t.Errorf("focus = %v, want panelBranches", a.focus)
	}
}

func TestApp_TabCyclesFocus(t *testing.T) {
	a := fixtureApp(t)

	a.Update(press("tab"))
	if a.focus != panelCommits {
		t.Errorf("after tab: focus = %v, want panelCommits", a.focus)
	}

	a.Update(press("tab"))
	if a.focus != panelDetail {
		t.Errorf("after tab×2: focus = %v, want panelDetail", a.focus)
	}

	a.Update(press("tab"))
	if a.focus != panelBranches {
		t.Errorf("after tab×3: focus = %v, want panelBranches (wrapped)", a.focus)
	}
}

func TestApp_ShiftTabCyclesFocusBack(t *testing.T) {
	a := fixtureApp(t)

	a.Update(press("shift+tab"))
	if a.focus != panelDetail {
		t.Errorf("after shift+tab: focus = %v, want panelDetail", a.focus)
	}
}

// ── dialog routing ────────────────────────────────────────────────────────────

func TestApp_KeyRoutesToDialogNotPanel(t *testing.T) {
	a := fixtureApp(t)
	initialCursor := a.branches.cursor

	// Push a menu with multiple items so 'j' moves the dialog cursor, not the panel.
	a.dialogs.Push(NewContextMenu(BranchMenuItems(false, true, "")))

	a.Update(press("j"))

	// Branch panel cursor must not have moved.
	if a.branches.cursor != initialCursor {
		t.Error("branch cursor moved despite dialog being open — key was not routed to dialog")
	}
	// Dialog should still be open ('j' navigates, doesn't close).
	if !a.dialogs.IsOpen() {
		t.Error("dialog closed unexpectedly on 'j'")
	}
	// Dialog cursor should have advanced.
	if menu, ok := a.dialogs.Active().(ContextMenuModel); ok {
		if menu.cursor != 1 {
			t.Errorf("dialog cursor = %d, want 1", menu.cursor)
		}
	}
}

func TestApp_DialogClosedKeyReachesPanel(t *testing.T) {
	a := fixtureApp(t)
	initialCursor := a.branches.cursor

	// No dialog open — 'j' should reach the branch panel.
	a.Update(press("j"))

	if a.branches.cursor == initialCursor {
		t.Error("branch cursor did not move — 'j' should reach panel when no dialog is open")
	}
}

// ── message routing ───────────────────────────────────────────────────────────

func TestApp_BranchSelectedMsgTriggersLoad(t *testing.T) {
	a := fixtureApp(t)
	branches, _ := a.client.ListBranches()
	if len(branches) == 0 {
		t.Skip("no branches in fixture")
	}

	_, cmd := a.Update(BranchSelectedMsg{Branch: branches[0]})
	if cmd == nil {
		t.Error("expected loadCommitsCmd, got nil")
	}
}

func TestApp_CommitSelectedMsgTriggersLoad(t *testing.T) {
	a := fixtureApp(t)
	commits, _ := a.client.ListCommits("main", 1)
	if len(commits) == 0 {
		t.Skip("no commits in fixture")
	}

	_, cmd := a.Update(CommitSelectedMsg{Commit: commits[0]})
	if cmd == nil {
		t.Error("expected loadDetailCmd, got nil")
	}
}

func TestApp_WorkflowResultSuccess(t *testing.T) {
	a := fixtureApp(t)
	a.Update(WorkflowResultMsg{Result: git.WorkflowResult{Message: "branch deleted"}})

	if a.statusMsg != "branch deleted" {
		t.Errorf("statusMsg = %q, want %q", a.statusMsg, "branch deleted")
	}
	if a.statusIsErr {
		t.Error("statusIsErr should be false on success")
	}
}

func TestApp_WorkflowResultError(t *testing.T) {
	a := fixtureApp(t)
	a.Update(WorkflowResultMsg{Result: git.WorkflowResult{Err: errors.New("push rejected")}})

	if a.statusMsg != "push rejected" {
		t.Errorf("statusMsg = %q, want %q", a.statusMsg, "push rejected")
	}
	if !a.statusIsErr {
		t.Error("statusIsErr should be true on error")
	}
}

func TestApp_EnterOpensContextMenu(t *testing.T) {
	a := fixtureApp(t)
	// focus is panelBranches with a selection — enter should open the context menu
	a.Update(press("enter"))

	if !a.dialogs.IsOpen() {
		t.Fatal("expected context menu dialog after pressing enter")
	}
	if _, ok := a.dialogs.Active().(ContextMenuModel); !ok {
		t.Errorf("Active() = %T, want ContextMenuModel", a.dialogs.Active())
	}
}

func TestApp_MenuSelectedDispatchesAction(t *testing.T) {
	a := fixtureApp(t)
	// ActionFetch always produces a runWorkflow cmd regardless of panel state.
	_, cmd := a.Update(MenuSelectedMsg{Action: ActionFetch})
	if cmd == nil {
		t.Error("expected non-nil cmd for ActionFetch")
	}
}

func TestApp_MergeActionOpensBranchPicker(t *testing.T) {
	a := fixtureApp(t)
	// Select a non-current branch first.
	branches, _ := a.client.ListBranches()
	var nonCurrent *git.Branch
	for i := range branches {
		if !branches[i].IsRemote && !branches[i].IsCurrent {
			nonCurrent = &branches[i]
			break
		}
	}
	if nonCurrent == nil {
		t.Skip("no non-current local branch in fixture")
	}
	a.branches.SetBranches(branches)
	// Move cursor to the non-current branch.
	for {
		if sel := a.branches.Selected(); sel != nil && sel.Name == nonCurrent.Name {
			break
		}
		a.branches, _ = a.branches.Update(press("j"))
	}

	a.Update(MenuSelectedMsg{Action: ActionMerge})

	if !a.dialogs.IsOpen() {
		t.Fatal("expected branch picker dialog to open for ActionMerge")
	}
	if _, ok := a.dialogs.Active().(*BranchPickerModel); !ok {
		t.Errorf("Active() = %T, want *BranchPickerModel", a.dialogs.Active())
	}
}

func TestApp_MenuActionConfirmPushesDialog(t *testing.T) {
	a := fixtureApp(t)
	// ActionForcePush for a branch pushes a ConfirmModel onto the dialog stack.
	a.Update(MenuSelectedMsg{Action: ActionForcePush})

	if !a.dialogs.IsOpen() {
		t.Error("expected confirm dialog to be pushed for ActionForcePush")
	}
	if _, ok := a.dialogs.Active().(*ConfirmModel); !ok {
		t.Errorf("Active() = %T, want *ConfirmModel", a.dialogs.Active())
	}
}

func TestApp_MenuActionAmendPushesDialog(t *testing.T) {
	a := fixtureApp(t)
	// Select the HEAD commit (index 0) so ActionAmend has something to work with.
	commits, _ := a.client.ListCommits("main", 1)
	if len(commits) == 0 {
		t.Skip("no commits in fixture")
	}
	a.commits.SetCommits(commits)

	a.Update(MenuSelectedMsg{Action: ActionAmend})

	if !a.dialogs.IsOpen() {
		t.Error("expected amend dialog to be pushed for ActionAmend")
	}
	if _, ok := a.dialogs.Active().(*AmendModel); !ok {
		t.Errorf("Active() = %T, want *AmendModel", a.dialogs.Active())
	}
}

func TestApp_SquashWithoutSelectionSetsError(t *testing.T) {
	a := fixtureApp(t)
	// No multi-select active — SelectedHashes returns single cursor hash.
	// ActionSquash requires ≥2 commits.
	a.Update(MenuSelectedMsg{Action: ActionSquash})

	if !a.statusIsErr {
		t.Error("expected error status when squash has fewer than 2 commits selected")
	}
}

// ── window sizing ─────────────────────────────────────────────────────────────

func TestApp_WindowSizeMsgRelayouts(t *testing.T) {
	a := fixtureApp(t)
	a.Update(tea.WindowSizeMsg{Width: 200, Height: 50})

	if a.termW != 200 || a.termH != 50 {
		t.Errorf("term size = %d×%d, want 200×50", a.termW, a.termH)
	}
}
