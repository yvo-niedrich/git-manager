package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/ui"
)

type AmendSubmitMsg struct{ NewMessage string }

// ConfirmModel is a yes/no confirmation dialog overlay.
type ConfirmModel struct {
	msg       string
	subject   string // optional: shown on its own line, highlighted (e.g. branch name)
	confirmFn func() tea.Cmd
}

func NewConfirmDialog(msg string, fn func() tea.Cmd) *ConfirmModel {
	return &ConfirmModel{msg: msg, confirmFn: fn}
}

// NewConfirmDialogWithSubject is like NewConfirmDialog but renders subject (e.g. a
// branch name) on its own highlighted line so it stands out from the question text.
func NewConfirmDialogWithSubject(msg, subject string, fn func() tea.Cmd) *ConfirmModel {
	return &ConfirmModel{msg: msg, subject: subject, confirmFn: fn}
}

func (m *ConfirmModel) Priority() int { return 20 }

func (m *ConfirmModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "y", "Y", "enter":
			return nil, m.confirmFn()
		case "n", "N", "esc":
			return nil, nil
		}
	}
	return m, nil
}

func (m *ConfirmModel) View() string {
	hints := ui.KeyHintStyle.Render("[y]") + " " +
		ui.DescHintStyle.Render(ui.HintConfirm+"  ") +
		ui.KeyHintStyle.Render("[n/Esc]") + " " +
		ui.DescHintStyle.Render(ui.HintCancel)

	subjectLine := ""
	if m.subject != "" {
		subjectLine = "\n\n  " + ui.NormalItemStyle.Bold(true).Render(m.subject)
	}

	body := fmt.Sprintf("\n  %s%s\n\n  %s",
		ui.StatusErrStyle.Render(m.msg),
		subjectLine,
		hints,
	)
	return ui.MenuBorderStyle.Render(body)
}

// AmendModel is a text-edit dialog overlay for rewriting commit messages.
type AmendModel struct {
	input textinput.Model
}

func NewAmendDialog(currentMsg string) *AmendModel {
	ti := textinput.New()
	ti.Placeholder = ui.AmendPlaceholder
	ti.CharLimit = 256
	ti.SetValue(currentMsg)
	ti.Focus()
	ti.CursorEnd()
	return &AmendModel{input: ti}
}

func (m *AmendModel) Priority() int { return 20 }

func (m *AmendModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "enter":
			val := m.input.Value()
			return nil, func() tea.Msg { return AmendSubmitMsg{NewMessage: val} }
		case "esc":
			return nil, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *AmendModel) View() string {
	body := fmt.Sprintf("\n  %s\n\n  %s\n\n  %s %s",
		ui.TitleStyle(true).Render(ui.AmendTitle),
		m.input.View(),
		ui.KeyHintStyle.Render("[Enter]"),
		ui.DescHintStyle.Render(ui.HintSave+"  ")+ui.KeyHintStyle.Render("[Esc]")+" "+ui.DescHintStyle.Render(ui.HintCancel),
	)
	return ui.MenuBorderStyle.Render(body)
}

// NewBranchSubmitMsg is emitted when the user confirms a new branch name.
type NewBranchSubmitMsg struct {
	Name string
	From string // source branch to create from
}

// NewBranchModel is a text-input dialog for creating a new branch.
type NewBranchModel struct {
	fromBranch string
	input      textinput.Model
	existing   map[string]bool
	errMsg     string
}

func NewBranchDialog(fromBranch string, localNames []string) *NewBranchModel {
	existing := make(map[string]bool, len(localNames))
	for _, n := range localNames {
		existing[n] = true
	}
	ti := textinput.New()
	ti.Placeholder = ui.NewBranchPlaceholder
	ti.CharLimit = 128
	ti.Width = 40
	ti.PromptStyle = ui.DimItemStyle
	ti.TextStyle = ui.NormalItemStyle
	ti.PlaceholderStyle = ui.DimItemStyle
	ti.Focus()
	return &NewBranchModel{fromBranch: fromBranch, input: ti, existing: existing}
}

func (m *NewBranchModel) Priority() int { return 20 }

func (m *NewBranchModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "enter":
			name := strings.TrimSpace(m.input.Value())
			if name == "" {
				m.errMsg = ui.NewBranchErrEmpty
				return m, nil
			}
			if m.existing[name] {
				m.errMsg = fmt.Sprintf(ui.NewBranchErrExistsFmt, name)
				return m, nil
			}
			return nil, func() tea.Msg { return NewBranchSubmitMsg{Name: name, From: m.fromBranch} }
		case "esc":
			return nil, nil
		default:
			m.errMsg = ""
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *NewBranchModel) View() string {
	title := ui.TitleStyle(true).Render(fmt.Sprintf(ui.NewBranchTitleFmt, m.fromBranch))
	hints := ui.KeyHintStyle.Render("[Enter]") + " " +
		ui.DescHintStyle.Render(ui.HintCreate+"  ") +
		ui.KeyHintStyle.Render("[Esc]") + " " +
		ui.DescHintStyle.Render(ui.HintCancel)

	// Reserve a fixed line for the error so the box doesn't resize on validation.
	errLine := "  "
	if m.errMsg != "" {
		errLine = ui.StatusErrStyle.Render("  " + m.errMsg)
	}

	body := fmt.Sprintf("\n  %s\n\n  %s\n%s\n\n  %s",
		title, m.input.View(), errLine, hints)
	return ui.MenuBorderStyle.Render(body)
}
