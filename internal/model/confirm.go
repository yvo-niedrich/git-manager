package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/ui"
)

type AmendSubmitMsg struct{ NewMessage string }

// ConfirmModel is a yes/no confirmation dialog overlay.
type ConfirmModel struct {
	msg       string
	confirmFn func() tea.Cmd
}

func NewConfirmDialog(msg string, fn func() tea.Cmd) *ConfirmModel {
	return &ConfirmModel{msg: msg, confirmFn: fn}
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
	body := fmt.Sprintf("\n  %s\n\n  %s %s",
		ui.StatusErrStyle.Render(m.msg),
		ui.KeyHintStyle.Render("[y]"),
		ui.DescHintStyle.Render("confirm  ")+ui.KeyHintStyle.Render("[n/Esc]")+" "+ui.DescHintStyle.Render("cancel"),
	)
	return ui.MenuBorderStyle.Render(body)
}

// AmendModel is a text-edit dialog overlay for rewriting commit messages.
type AmendModel struct {
	input textinput.Model
}

func NewAmendDialog(currentMsg string) *AmendModel {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
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
		ui.TitleStyle(true).Render("Amend commit message:"),
		m.input.View(),
		ui.KeyHintStyle.Render("[Enter]"),
		ui.DescHintStyle.Render("save  ")+ui.KeyHintStyle.Render("[Esc]")+" "+ui.DescHintStyle.Render("cancel"),
	)
	return ui.MenuBorderStyle.Render(body)
}
