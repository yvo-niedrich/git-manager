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

func BranchMenuItems(isRemote, isCurrent bool, upstream string) []MenuItem {
	if isRemote {
		return []MenuItem{
			{Label: ui.MenuCheckoutRemote, Action: ActionCheckoutRemote, Key: "c"},
			{Label: ui.MenuFetchRemote, Action: ActionFetch, Key: "f"},
		}
	}
	var items []MenuItem
	if !isCurrent {
		items = append(items, MenuItem{Label: ui.MenuCheckout, Action: ActionCheckout, Key: "c"})
	}
	items = append(items,
		MenuItem{Label: ui.MenuMerge, Action: ActionMerge, Key: "m"},
		MenuItem{Label: ui.MenuRebase, Action: ActionRebase, Key: "r"},
		MenuItem{Label: ui.MenuPush, Action: ActionPush, Key: "p"},
		MenuItem{Label: ui.MenuForcePush, Action: ActionForcePush, Key: "F", MenuOnly: true},
	)
	if upstream != "" {
		items = append(items, MenuItem{Label: fmt.Sprintf(ui.MenuPullFromFmt, upstream), Action: ActionPull, Key: "l"})
	}
	items = append(items, MenuItem{Label: ui.MenuNewBranch, Action: ActionNewBranch, Key: "n"})
	if !isCurrent {
		label := ui.MenuDeleteBranch
		if isRemote {
			label = ui.MenuDeleteRemoteBranch
		}
		items = append(items, MenuItem{Label: label, Action: ActionDeleteBranch, Key: "D"})
	}
	return items
}

func CommitMenuItems(isHead bool) []MenuItem {
	items := []MenuItem{
		{Label: ui.MenuCherryPick, Action: ActionCherryPick, Key: "p"},
		{Label: ui.MenuRevert, Action: ActionRevert, Key: "R"},
		{Label: ui.MenuCopyHash, Action: ActionCopyHash, Key: "y"},
	}
	if !isHead {
		items = append(items, MenuItem{Label: ui.MenuDropCommit, Action: ActionDrop, Key: "d"})
	} else {
		items = append(items, MenuItem{Label: ui.MenuAmend, Action: ActionAmend, Key: "a"})
	}
	items = append(items, MenuItem{Label: ui.MenuSquash, Action: ActionSquash, Key: "s"})
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
