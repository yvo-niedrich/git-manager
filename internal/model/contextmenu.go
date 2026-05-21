package model

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/ui"
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
)

type MenuItem struct {
	Label    string
	Action   MenuAction
	Key      string
	MenuOnly bool // if true, key is not registered as a panel-level direct shortcut
}

type ContextMenuModel struct {
	items  []MenuItem
	cursor int
}

type MenuSelectedMsg struct {
	Action MenuAction
}
type MenuClosedMsg struct{}

func BranchMenuItems(isRemote, isCurrent bool, upstream string) []MenuItem {
	if isRemote {
		return []MenuItem{
			{Label: "Checkout (create local tracking)", Action: ActionCheckoutRemote, Key: "c"},
			{Label: "Fetch remote", Action: ActionFetch, Key: "f"},
		}
	}
	var items []MenuItem
	if !isCurrent {
		items = append(items, MenuItem{Label: "Checkout", Action: ActionCheckout, Key: "c"})
	}
	items = append(items,
		MenuItem{Label: "Merge into current branch", Action: ActionMerge, Key: "m"},
		MenuItem{Label: "Rebase current onto this", Action: ActionRebase, Key: "r"},
		MenuItem{Label: "Push to remote", Action: ActionPush, Key: "p"},
		MenuItem{Label: "Force-push to remote", Action: ActionForcePush, Key: "F", MenuOnly: true},
	)
	if upstream != "" {
		items = append(items, MenuItem{Label: "Pull from " + upstream, Action: ActionPull, Key: "l"})
	}
	if !isCurrent {
		items = append(items, MenuItem{Label: "Delete branch", Action: ActionDeleteBranch, Key: "D"})
	}
	return items
}

func CommitMenuItems(isHead bool) []MenuItem {
	items := []MenuItem{
		{Label: "Cherry-pick onto current branch", Action: ActionCherryPick, Key: "p"},
		{Label: "Revert (create revert commit)", Action: ActionRevert, Key: "R"},
		{Label: "Copy commit hash", Action: ActionCopyHash, Key: "y"},
	}
	if !isHead {
		items = append(items, MenuItem{Label: "Drop commit from history", Action: ActionDrop, Key: "d"})
	} else {
		items = append(items, MenuItem{Label: "Amend commit message", Action: ActionAmend, Key: "a"})
	}
	items = append(items, MenuItem{Label: "Squash with next commit", Action: ActionSquash, Key: "s"})
	return items
}

func NewContextMenu(items []MenuItem) ContextMenuModel {
	return ContextMenuModel{items: items}
}

func (m ContextMenuModel) Update(msg tea.Msg) (ContextMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m, func() tea.Msg { return MenuClosedMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			m.cursor = (m.cursor + 1) % len(m.items)
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
			if len(m.items) > 0 {
				action := m.items[m.cursor].Action
				return m, func() tea.Msg { return MenuSelectedMsg{Action: action} }
			}
		default:
			// shortcut keys
			for _, item := range m.items {
				if item.Key != "" && msg.String() == item.Key {
					return m, func() tea.Msg { return MenuSelectedMsg{Action: item.Action} }
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
	sb.WriteString(ui.TitleStyle(true).Render("Actions") + "\n\n")
	for i, item := range m.items {
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
