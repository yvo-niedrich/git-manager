package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/gitmg/internal/ui"
)

// BranchPickerSubmitMsg is emitted when the user confirms a target branch.
type BranchPickerSubmitMsg struct {
	Action MenuAction
	Source string
	Target string
}

// BranchPickerModel is a dialog with a filterable branch list for choosing a
// merge or rebase target. It excludes the current branch and the source branch.
type BranchPickerModel struct {
	action      MenuAction
	source      string
	branches    []string // local branches, source already removed
	cursor      int
	offset      int
	visibleRows int
	listW       int // display width for list rows and separator
	filter      textinput.Model
}

// defaultTargetBranches is checked in order when pre-selecting the cursor.
var defaultTargetBranches = []string{"main", "master", "develop", "development"}

func NewBranchPickerDialog(action MenuAction, source, current string, localNames []string, termH int) *BranchPickerModel {
	var branches []string
	for _, name := range localNames {
		if name != source {
			branches = append(branches, name)
		}
	}

	// Pre-select: prefer the current branch, then fall back to well-known defaults.
	cursor := 0
	found := false
	if current != source {
		for i, name := range branches {
			if name == current {
				cursor = i
				found = true
				break
			}
		}
	}
	if !found {
	outer:
		for _, pref := range defaultTargetBranches {
			for i, name := range branches {
				if name == pref {
					cursor = i
					break outer
				}
			}
		}
	}

	// visibleRows: 4–10 scaled by terminal height.
	visibleRows := termH / 5
	if visibleRows < 4 {
		visibleRows = 4
	}
	if visibleRows > 10 {
		visibleRows = 10
	}

	// listW: wide enough for the longest name + "▶ " prefix, capped at 52.
	listW := 24
	for _, name := range branches {
		if w := lipgloss.Width(name) + 2; w > listW {
			listW = w
		}
	}
	if listW > 52 {
		listW = 52
	}

	ti := textinput.New()
	ti.Placeholder = ui.BranchPickerFilterPlaceholder
	ti.Prompt = "/ "
	ti.PromptStyle = ui.DimItemStyle
	ti.TextStyle = ui.NormalItemStyle
	ti.PlaceholderStyle = ui.DimItemStyle
	ti.Width = listW - 2 // subtract prompt width
	ti.Focus()

	return &BranchPickerModel{
		action:      action,
		source:      source,
		branches:    branches,
		cursor:      cursor,
		visibleRows: visibleRows,
		listW:       listW,
		filter:      ti,
	}
}

func (m *BranchPickerModel) Priority() int { return 20 }

func (m *BranchPickerModel) filteredBranches() []string {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		return m.branches
	}
	var out []string
	for _, name := range m.branches {
		if strings.Contains(strings.ToLower(name), q) {
			out = append(out, name)
		}
	}
	return out
}

func (m *BranchPickerModel) clampOffset() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.visibleRows {
		m.offset = m.cursor - m.visibleRows + 1
	}
}

func (m *BranchPickerModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}

	switch km.String() {
	case "esc":
		if m.filter.Value() != "" {
			m.filter.SetValue("")
			m.cursor = 0
			m.offset = 0
			return m, nil
		}
		return nil, nil

	case "enter":
		filtered := m.filteredBranches()
		if len(filtered) == 0 {
			return m, nil
		}
		target := filtered[m.cursor]
		return nil, func() tea.Msg {
			return BranchPickerSubmitMsg{Action: m.action, Source: m.source, Target: target}
		}

	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
			m.clampOffset()
		}
		return m, nil

	case "down", "ctrl+n":
		filtered := m.filteredBranches()
		if m.cursor < len(filtered)-1 {
			m.cursor++
			m.clampOffset()
		}
		return m, nil
	}

	prev := m.filter.Value()
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	if m.filter.Value() != prev {
		m.cursor = 0
		m.offset = 0
	}
	return m, cmd
}

func (m *BranchPickerModel) View() string {
	var actionLabel string
	switch m.action {
	case ActionMerge:
		actionLabel = fmt.Sprintf(ui.BranchPickerMergeFmt, m.source)
	case ActionRebase:
		actionLabel = fmt.Sprintf(ui.BranchPickerRebaseFmt, m.source)
	}

	filtered := m.filteredBranches()

	// Clamp cursor for rendering without mutating state.
	cursor := m.cursor
	if cursor >= len(filtered) {
		cursor = len(filtered) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	sep := ui.DimItemStyle.Render(strings.Repeat("─", m.listW))

	hints := ui.KeyHintStyle.Render("[↑↓]") + " " +
		ui.DescHintStyle.Render(ui.HintNavigate+"  ") +
		ui.KeyHintStyle.Render("[Enter]") + " " +
		ui.DescHintStyle.Render(ui.HintApply+"  ") +
		ui.KeyHintStyle.Render("[Esc]") + " " +
		ui.DescHintStyle.Render(ui.HintCancel)

	var sb strings.Builder
	sb.WriteString("\n  ")
	sb.WriteString(actionLabel)
	sb.WriteString("\n\n  ")
	sb.WriteString(m.filter.View())
	sb.WriteString("\n  ")
	sb.WriteString(sep)
	sb.WriteString("\n")

	if len(filtered) == 0 {
		sb.WriteString("  ")
		sb.WriteString(ui.DimItemStyle.Render(ui.BranchPickerEmpty))
		sb.WriteString("\n")
	} else {
		end := m.offset + m.visibleRows
		if end > len(filtered) {
			end = len(filtered)
		}
		for i := m.offset; i < end; i++ {
			name := filtered[i]
			if lipgloss.Width(name) > m.listW-2 {
				runes := []rune(name)
				for len(runes) > 0 && lipgloss.Width(string(runes)) > m.listW-3 {
					runes = runes[:len(runes)-1]
				}
				name = string(runes) + "…"
			}
			if i == cursor {
				sb.WriteString(ui.SelectedItemStyle.Width(m.listW).Render("▶ " + name))
			} else {
				sb.WriteString(ui.NormalItemStyle.Render("  " + name))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n  ")
	sb.WriteString(hints)

	return ui.MenuBorderStyle.Render(sb.String())
}
