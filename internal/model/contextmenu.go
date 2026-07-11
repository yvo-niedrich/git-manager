package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yvo.niedrich/git-manager/internal/ui"
)

type MenuAction int

const (
	ActionNone MenuAction = iota
	ActionCheckout
	ActionCheckoutRemote
	ActionMerge
	ActionRebase
	ActionDeleteBranch
	ActionPush
	ActionForcePush
	ActionPull
	ActionFetch
	ActionCherryPick
	ActionRevert
	ActionDrop
	ActionAmend
	ActionSquash
	ActionCopyHash
	ActionNewBranch
	ActionCommit
	ActionUncommit
)

type MenuItem struct {
	Label    string
	Action   MenuAction
	Key      string
	MenuOnly bool // if true, key is not registered as a panel-level direct shortcut
	NewGroup bool // if true, a blank line is rendered before this item
}

type ContextMenuModel struct {
	items  []MenuItem
	cursor int
}

type MenuSelectedMsg struct {
	Action MenuAction
}

func BranchMenuItems(isRemote, isCurrent, hasUncommittedChanges bool, upstream string) []MenuItem {
	if isRemote {
		return []MenuItem{
			{Label: ui.MenuCheckoutRemote, Action: ActionCheckoutRemote, Key: "c"},
			{Label: ui.MenuFetchRemote, Action: ActionFetch, Key: "f", NewGroup: true},
		}
	}

	var items []MenuItem
	switch {
	case !isCurrent:
		items = append(items, MenuItem{Label: ui.MenuCheckout, Action: ActionCheckout, Key: "c"})
	case hasUncommittedChanges:
		items = append(items, MenuItem{Label: ui.MenuCommit, Action: ActionCommit, Key: "c"})
	}

	// Sync with remote.
	syncStart := len(items)
	if upstream != "" {
		items = append(items, MenuItem{Label: fmt.Sprintf(ui.MenuPullFromFmt, upstream), Action: ActionPull, Key: "l"})
	}
	items = append(items,
		MenuItem{Label: ui.MenuPush, Action: ActionPush, Key: "p"},
		MenuItem{Label: ui.MenuForcePush, Action: ActionForcePush, Key: "F", MenuOnly: true},
	)
	items[syncStart].NewGroup = syncStart > 0

	// Integrate with another branch.
	integrateStart := len(items)
	items = append(items,
		MenuItem{Label: ui.MenuMerge, Action: ActionMerge, Key: "m"},
		MenuItem{Label: ui.MenuRebase, Action: ActionRebase, Key: "r"},
	)
	items[integrateStart].NewGroup = true

	// Create.
	items = append(items, MenuItem{Label: ui.MenuNewBranch, Action: ActionNewBranch, Key: "n", NewGroup: true})

	// Destructive.
	if !isCurrent {
		items = append(items, MenuItem{Label: ui.MenuDeleteBranch, Action: ActionDeleteBranch, Key: "D", NewGroup: true})
	}
	return items
}

func CommitMenuItems(isHead bool) []MenuItem {
	// Inspect.
	items := []MenuItem{
		{Label: ui.MenuCopyHash, Action: ActionCopyHash, Key: "y"},
	}

	// Apply elsewhere (non-destructive to history).
	items = append(items,
		MenuItem{Label: ui.MenuCherryPick, Action: ActionCherryPick, Key: "p", NewGroup: true},
		MenuItem{Label: ui.MenuRevert, Action: ActionRevert, Key: "R"},
	)

	// Rewrite history.
	if isHead {
		items = append(items,
			MenuItem{Label: ui.MenuAmend, Action: ActionAmend, Key: "a", NewGroup: true},
			MenuItem{Label: ui.MenuUncommit, Action: ActionUncommit, Key: "u"},
		)
		items = append(items, MenuItem{Label: ui.MenuSquash, Action: ActionSquash, Key: "s"})
	} else {
		items = append(items, MenuItem{Label: ui.MenuSquash, Action: ActionSquash, Key: "s", NewGroup: true})
		// Destructive.
		items = append(items, MenuItem{Label: ui.MenuDropCommit, Action: ActionDrop, Key: "d", NewGroup: true})
	}
	return items
}

func NewContextMenu(items []MenuItem) ContextMenuModel {
	return ContextMenuModel{items: items}
}

func (m ContextMenuModel) Priority() int { return 10 }

func (m ContextMenuModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return nil, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			m.cursor = (m.cursor + 1) % len(m.items)
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			if len(m.items) > 0 {
				action := m.items[m.cursor].Action
				return nil, func() tea.Msg { return MenuSelectedMsg{Action: action} }
			}
		default:
			for _, item := range m.items {
				if item.Key != "" && msg.String() == item.Key {
					return nil, func() tea.Msg { return MenuSelectedMsg{Action: item.Action} }
				}
			}
		}
	}
	return m, nil
}

func (m ContextMenuModel) View() string {
	if len(m.items) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(ui.TitleStyle(true).Render(ui.MenuTitle) + "\n\n")
	for i, item := range m.items {
		if item.NewGroup && i > 0 {
			sb.WriteString("\n")
		}
		keyPart := ui.KeyHintStyle.Render("[" + item.Key + "]")
		labelPart := ui.DescHintStyle.Render(" " + item.Label)
		line := keyPart + labelPart
		if i == m.cursor {
			line = ui.SelectedItemStyle.Render(line)
		}
		sb.WriteString(line + "\n")
	}
	return ui.MenuBorderStyle.Render(sb.String())
}
