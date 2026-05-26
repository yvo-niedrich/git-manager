package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/gitmg/internal/git"
	"github.com/user/gitmg/internal/ui"
)

type panel int

const (
	panelBranches panel = iota
	panelCommits
	panelDetail
	panelCount
)

type WorkflowResultMsg struct {
	Result git.WorkflowResult
}

type clearStatusMsg struct{ gen int }

type loadCommitsMsg struct {
	ref     string
	commits []git.Commit
	err     error
}

type loadDetailMsg struct {
	detail git.CommitDetail
	err    error
}

type refreshBranchesMsg struct {
	branches []git.Branch
	err      error
}

type App struct {
	branches    BranchesModel
	commits     CommitsModel
	detail      DetailModel
	dialogs     dialogStack
	focus       panel
	wf          *git.Workflows
	client      *git.Client
	repoRoot    string
	termW       int
	termH       int
	statusMsg   string
	statusIsErr bool
	statusGen   int
}

func NewApp(repoRoot string) (*App, error) {
	client := git.NewClient(repoRoot)
	wf := git.NewWorkflows(client)

	branches, err := client.ListBranches()
	if err != nil {
		return nil, fmt.Errorf("could not list branches: %w", err)
	}

	bm := NewBranchesModel()
	bm.SetBranches(branches)
	cm := NewCommitsModel()
	dm := NewDetailModel()

	a := &App{
		branches: bm,
		commits:  cm,
		detail:   dm,
		focus:    panelBranches,
		wf:       wf,
		client:   client,
		repoRoot: repoRoot,
	}
	a.branches.focused = true

	if sel := bm.Selected(); sel != nil {
		commits, _ := client.ListCommits(sel.FullRef(), 200)
		cm.SetCommits(commits)
		cm.branchRef = sel.FullRef()
		a.commits = cm
		if sel2 := cm.Selected(); sel2 != nil {
			detail, _ := client.ShowCommit(sel2.Hash)
			dm.SetCommit(detail)
			a.detail = dm
		}
	}

	return a, nil
}

func (a *App) Init() tea.Cmd {
	title := filepath.Base(os.Args[0]) + ": " + filepath.Base(a.repoRoot)
	return tea.SetWindowTitle(title)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.termW = msg.Width
		a.termH = msg.Height
		a.relayout()
		return a, nil

	case tea.KeyMsg:
		if a.dialogs.IsOpen() {
			return a, a.dialogs.Update(msg)
		}
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		if a.isFiltering() {
			return a, a.handlePanelKey(msg)
		}
		switch msg.String() {
		case "q":
			return a, tea.Quit
		case "tab", "right":
			a.focus = (a.focus + 1) % panelCount
			a.syncFocus()
			a.relayout()
			return a, nil
		case "shift+tab", "left":
			a.focus = panel((int(a.focus) - 1 + int(panelCount)) % int(panelCount))
			a.syncFocus()
			a.relayout()
			return a, nil
		case "enter", "x":
			return a, a.openContextMenu()
		}
		if cmd := a.tryDirectShortcut(msg.String()); cmd != nil {
			return a, cmd
		}
		return a, a.handlePanelKey(msg)

	case MenuSelectedMsg:
		return a, a.handleMenuAction(msg.Action)

	case BranchSelectedMsg:
		return a, a.loadCommitsCmd(msg.Branch.FullRef())

	case CommitSelectedMsg:
		return a, a.loadDetailCmd(msg.Commit.Hash)

	case clearStatusMsg:
		if msg.gen == a.statusGen {
			a.statusMsg = ""
			a.statusIsErr = false
		}
		return a, nil

	case loadCommitsMsg:
		if msg.err != nil {
			return a, a.setStatus("load commits: "+msg.err.Error(), true)
		}
		a.commits.SetCommits(msg.commits)
		a.commits.branchRef = msg.ref
		if sel := a.commits.Selected(); sel != nil {
			return a, a.loadDetailCmd(sel.Hash)
		}
		return a, nil

	case loadDetailMsg:
		if msg.err != nil {
			return a, a.setStatus("load detail: "+msg.err.Error(), true)
		}
		a.detail.SetCommit(msg.detail)
		return a, nil

	case refreshBranchesMsg:
		if msg.err == nil {
			a.branches.SetBranches(msg.branches)
		}
		return a, nil

	case WorkflowResultMsg:
		r := msg.Result
		var statusCmd tea.Cmd
		if r.Err != nil {
			statusCmd = a.setStatus(r.Err.Error(), true)
		} else {
			statusCmd = a.setStatus(r.Message, false)
		}
		return a, tea.Batch(statusCmd, a.refreshAll())

	case AmendSubmitMsg:
		return a, a.runWorkflow(func() git.WorkflowResult {
			return a.wf.AmendLast(msg.NewMessage)
		})
	}

	return a, nil
}

func (a *App) View() string {
	h := a.termH - 2
	if h < 1 {
		h = 24
	}

	bg := a.renderPanels() + a.renderStatusBar()

	if a.dialogs.IsOpen() {
		const mX, mY = 4, 3    // outer margin from terminal edges
		const padX, padY = 2, 1 // black gutter around the box

		box := a.dialogs.Active().View()
		boxW := lipgloss.Width(box)
		boxH := lipgloss.Height(box)

		compW := boxW + 2*padX
		compH := boxH + 2*padY
		blackLine := lipgloss.NewStyle().
			Background(ui.ColorBackdrop).
			Render(strings.Repeat(" ", compW))
		backdropRows := make([]string, compH)
		for i := range backdropRows {
			backdropRows[i] = blackLine
		}
		backdrop := strings.Join(backdropRows, "\n")

		composite := ui.PlaceOverlay(padX, padY, box, backdrop)
		outerX := mX + max(0, (a.termW-2*mX-compW)/2)
		outerY := mY + max(0, (h-2*mY-compH)/2)
		return ui.PlaceOverlay(outerX, outerY, composite, bg)
	}
	return bg
}

func (a *App) renderPanels() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		a.branches.View(),
		a.commits.View(),
		a.detail.View(),
	)
}

func (a *App) renderStatusBar() string {
	if a.isFiltering() {
		return "\n" + ui.RenderStatusBar(a.termW, a.statusMsg, a.statusIsErr, ui.FilterHints)
	}
	var hints ui.HintSet
	switch a.focus {
	case panelBranches:
		isRemote, isCurrent, hasUpstream := false, false, false
		if sel := a.branches.Selected(); sel != nil {
			isRemote, isCurrent = sel.IsRemote, sel.IsCurrent
			hasUpstream = !sel.IsRemote && sel.Upstream != ""
		}
		hints = ui.BranchHints(isRemote, isCurrent, hasUpstream)
	case panelCommits:
		isHead := false
		if sel := a.commits.Selected(); sel != nil {
			isHead = sel.IsHead
		}
		hints = ui.CommitHints(isHead, a.commits.IsMultiSelect())
	case panelDetail:
		hints = ui.DetailHints
	}
	return "\n" + ui.RenderStatusBar(a.termW, a.statusMsg, a.statusIsErr, hints)
}

func (a *App) isFiltering() bool {
	switch a.focus {
	case panelBranches:
		return a.branches.IsFiltering()
	case panelCommits:
		return a.commits.IsFiltering()
	}
	return false
}

func (a *App) syncFocus() {
	a.branches.focused = a.focus == panelBranches
	a.commits.focused = a.focus == panelCommits
	a.detail.focused = a.focus == panelDetail
}

func (a *App) relayout() {
	ws := ui.Widths(a.termW, int(a.focus))
	h := a.termH - 2
	if h < 4 {
		h = 4
	}
	a.branches.SetSize(ws[0], h)
	a.commits.SetSize(ws[1], h)
	a.detail.SetSize(ws[2], h)
}

func (a *App) setStatus(msg string, isErr bool) tea.Cmd {
	a.statusMsg = msg
	a.statusIsErr = isErr
	if msg == "" || isErr {
		return nil
	}
	a.statusGen++
	gen := a.statusGen
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return clearStatusMsg{gen: gen}
	}
}

func (a *App) openContextMenu() tea.Cmd {
	var items []MenuItem
	switch a.focus {
	case panelBranches:
		sel := a.branches.Selected()
		if sel == nil {
			return nil
		}
		items = BranchMenuItems(sel.IsRemote, sel.IsCurrent, sel.Upstream)
	case panelCommits, panelDetail:
		sel := a.commits.Selected()
		if sel == nil {
			return nil
		}
		items = CommitMenuItems(sel.IsHead)
	}
	if len(items) == 0 {
		return nil
	}
	a.dialogs.Push(NewContextMenu(items))
	return nil
}

// panelOwnedKeys lists keys each panel model handles internally; direct shortcuts must not shadow them.
var branchOwnedKeys = map[string]bool{"/": true, "esc": true, "j": true, "k": true, "down": true, "up": true}
var commitOwnedKeys = map[string]bool{"/": true, "esc": true, "j": true, "k": true, "down": true, "up": true, "s": true, " ": true}

func (a *App) tryDirectShortcut(keyStr string) tea.Cmd {
	var reserved map[string]bool
	var items []MenuItem
	switch a.focus {
	case panelBranches:
		reserved = branchOwnedKeys
		sel := a.branches.Selected()
		if sel == nil {
			return nil
		}
		items = BranchMenuItems(sel.IsRemote, sel.IsCurrent, sel.Upstream)
	case panelCommits:
		reserved = commitOwnedKeys
		sel := a.commits.Selected()
		if sel == nil {
			return nil
		}
		items = CommitMenuItems(sel.IsHead)
	default:
		return nil
	}
	if reserved[keyStr] {
		return nil
	}
	for _, item := range items {
		if item.Key != "" && !item.MenuOnly && keyStr == item.Key {
			return a.handleMenuAction(item.Action)
		}
	}
	return nil
}

func (a *App) handleMenuAction(action MenuAction) tea.Cmd {
	switch action {
	case ActionCheckout:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.SwitchBranch(branch) })
		}
	case ActionCheckoutRemote:
		if sel := a.branches.Selected(); sel != nil {
			ref := sel.FullRef()
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.CheckoutRemote(ref) })
		}
	case ActionMerge:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.MergeInto(branch) })
		}
	case ActionRebase:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.RebaseOnto(branch) })
		}
	case ActionDeleteBranch:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			a.dialogs.Push(NewConfirmDialog(
				fmt.Sprintf("Delete branch %q?", branch),
				func() tea.Cmd {
					return a.runWorkflow(func() git.WorkflowResult { return a.wf.DeleteBranch(branch, false) })
				},
			))
		}
	case ActionPush:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.Push("origin", branch) })
		}
	case ActionForcePush:
		if sel := a.branches.Selected(); sel != nil {
			branch := sel.Name
			a.dialogs.Push(NewConfirmDialog(
				fmt.Sprintf("Force-push %q to origin? This overwrites remote history.", branch),
				func() tea.Cmd {
					return a.runWorkflow(func() git.WorkflowResult { return a.wf.ForcePush("origin", branch) })
				},
			))
		}
	case ActionPull:
		if sel := a.branches.Selected(); sel != nil {
			branch, upstream := sel.Name, sel.Upstream
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.Pull(branch, upstream) })
		}
	case ActionFetch:
		return a.runWorkflow(func() git.WorkflowResult { return a.wf.Fetch("origin") })
	case ActionCherryPick:
		if sel := a.commits.Selected(); sel != nil {
			hash := sel.Hash
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.CherryPick(hash) })
		}
	case ActionRevert:
		if sel := a.commits.Selected(); sel != nil {
			hash := sel.Hash
			return a.runWorkflow(func() git.WorkflowResult { return a.wf.RevertCommit(hash) })
		}
	case ActionDrop:
		if sel := a.commits.Selected(); sel != nil {
			hash := sel.Hash
			short := sel.ShortHash
			a.dialogs.Push(NewConfirmDialog(
				fmt.Sprintf("Drop commit %s from history?", short),
				func() tea.Cmd {
					return a.runWorkflow(func() git.WorkflowResult { return a.wf.DropCommit(hash) })
				},
			))
		}
	case ActionAmend:
		if sel := a.commits.Selected(); sel != nil {
			a.dialogs.Push(NewAmendDialog(sel.Subject))
		}
	case ActionSquash:
		hashes := a.commits.SelectedHashes()
		if len(hashes) < 2 {
			return a.setStatus("select at least 2 commits with [s] then [space]", true)
		}
		return a.runWorkflow(func() git.WorkflowResult { return a.wf.SquashCommits(hashes) })
	case ActionCopyHash:
		if sel := a.commits.Selected(); sel != nil {
			return a.setStatus("hash: "+sel.Hash, false)
		}
	}
	return nil
}

func (a *App) handlePanelKey(msg tea.KeyMsg) tea.Cmd {
	switch a.focus {
	case panelBranches:
		var cmd tea.Cmd
		a.branches, cmd = a.branches.Update(msg)
		return cmd
	case panelCommits:
		var cmd tea.Cmd
		a.commits, cmd = a.commits.Update(msg)
		return cmd
	case panelDetail:
		var cmd tea.Cmd
		a.detail, cmd = a.detail.Update(msg)
		return cmd
	}
	return nil
}

func (a *App) loadCommitsCmd(ref string) tea.Cmd {
	client := a.client
	return func() tea.Msg {
		commits, err := client.ListCommits(ref, 200)
		return loadCommitsMsg{ref: ref, commits: commits, err: err}
	}
}

func (a *App) loadDetailCmd(hash string) tea.Cmd {
	client := a.client
	return func() tea.Msg {
		detail, err := client.ShowCommit(hash)
		return loadDetailMsg{detail: detail, err: err}
	}
}

func (a *App) runWorkflow(fn func() git.WorkflowResult) tea.Cmd {
	return func() tea.Msg {
		return WorkflowResultMsg{Result: fn()}
	}
}

func (a *App) refreshAll() tea.Cmd {
	client := a.client
	current := a.client.CurrentBranch()
	return tea.Batch(
		func() tea.Msg {
			branches, err := client.ListBranches()
			return refreshBranchesMsg{branches: branches, err: err}
		},
		func() tea.Msg {
			commits, err := client.ListCommits(current, 200)
			return loadCommitsMsg{ref: current, commits: commits, err: err}
		},
	)
}
