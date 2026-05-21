package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/git"
	"github.com/user/gitmg/internal/ui"
)

type DetailMode int

const (
	DetailModeView DetailMode = iota
	DetailModeConfirm
	DetailModeAmend
)

type DetailModel struct {
	commit    *git.CommitDetail
	vp        viewport.Model
	width     int
	height    int
	focused   bool
	mode      DetailMode
	confirmFn func() tea.Cmd
	confirmMsg string
	input     textinput.Model
}

type DetailConfirmMsg struct{ Confirmed bool }
type AmendSubmitMsg struct{ NewMessage string }

func NewDetailModel() DetailModel {
	ti := textinput.New()
	ti.Placeholder = "Commit message..."
	ti.CharLimit = 256
	return DetailModel{
		vp:    viewport.New(40, 20),
		input: ti,
	}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.vp.Width = ui.InnerWidth(w)
	m.vp.Height = h - 6
}

func (m *DetailModel) SetCommit(c git.CommitDetail) {
	m.commit = &c
	m.mode = DetailModeView
	m.vp.SetContent(m.renderCommit(c))
	m.vp.GotoTop()
}

func (m *DetailModel) AskConfirm(msg string, fn func() tea.Cmd) {
	m.mode = DetailModeConfirm
	m.confirmMsg = msg
	m.confirmFn = fn
}

func (m *DetailModel) StartAmend(currentMsg string) {
	m.mode = DetailModeAmend
	m.input.SetValue(currentMsg)
	m.input.Focus()
	m.input.CursorEnd()
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case DetailModeView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
				m.vp.LineDown(1)
			case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
				m.vp.LineUp(1)
			}
		case DetailModeConfirm:
			switch msg.String() {
			case "y", "Y", "enter":
				m.mode = DetailModeView
				return m, m.confirmFn()
			case "n", "N", "esc":
				m.mode = DetailModeView
			}
		case DetailModeAmend:
			switch msg.String() {
			case "enter":
				val := m.input.Value()
				m.mode = DetailModeView
				m.input.Blur()
				return m, func() tea.Msg { return AmendSubmitMsg{NewMessage: val} }
			case "esc":
				m.mode = DetailModeView
				m.input.Blur()
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	title := ui.TitleStyle(m.focused).Render("Detail")
	var body string

	switch m.mode {
	case DetailModeConfirm:
		body = fmt.Sprintf("\n\n  %s\n\n  %s %s",
			ui.StatusErrStyle.Render(m.confirmMsg),
			ui.KeyHintStyle.Render("[y]"),
			ui.DescHintStyle.Render("confirm  ")+ui.KeyHintStyle.Render("[n/Esc]")+" "+ui.DescHintStyle.Render("cancel"),
		)
	case DetailModeAmend:
		body = fmt.Sprintf("\n  %s\n\n  %s\n\n  %s %s",
			ui.TitleStyle(true).Render("Amend commit message:"),
			m.input.View(),
			ui.KeyHintStyle.Render("[Enter]"),
			ui.DescHintStyle.Render("save  ")+ui.KeyHintStyle.Render("[Esc]")+" "+ui.DescHintStyle.Render("cancel"),
		)
	default:
		body = m.vp.View()
	}

	inner := title + "\n" + body
	return ui.PanelStyle(m.focused).Width(m.width).Height(m.height).Render(inner)
}

func (m DetailModel) renderCommit(c git.CommitDetail) string {
	var sb strings.Builder
	sb.WriteString(ui.HeadStyle.Render("commit " + c.Hash[:min(12, len(c.Hash))]) + "\n")
	sb.WriteString(ui.DimItemStyle.Render("Author: "+c.Author) + "\n")
	if !c.Date.IsZero() {
		sb.WriteString(ui.DimItemStyle.Render("Date:   "+c.Date.Format(time.RFC1123)) + "\n")
	}
	if len(c.Tags) > 0 {
		sb.WriteString(ui.TagStyle.Render("Tags:   "+strings.Join(c.Tags, ", ")) + "\n")
	}
	sb.WriteString("\n")
	if c.Subject != "" {
		sb.WriteString(ui.NormalItemStyle.Render("  "+c.Subject) + "\n")
	}
	if c.Body != "" {
		sb.WriteString("\n")
		for _, l := range strings.Split(strings.TrimSpace(c.Body), "\n") {
			sb.WriteString(ui.DimItemStyle.Render("  "+l) + "\n")
		}
	}
	if len(c.StatLines) > 0 {
		sb.WriteString("\n" + ui.SectionStyle.Render("Changed files:") + "\n")
		for _, l := range c.StatLines {
			sb.WriteString(ui.DimItemStyle.Render("  "+l) + "\n")
		}
	}
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
