package model

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yvo.niedrich/git-manager/internal/git"
	"github.com/yvo.niedrich/git-manager/internal/ui"
)

type DetailModel struct {
	commit  *git.CommitDetail
	vp      viewport.Model
	width   int
	height  int
	focused bool
}

func NewDetailModel() DetailModel {
	return DetailModel{vp: viewport.New(40, 20)}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.vp.Width = ui.InnerWidth(w)
	m.vp.Height = h - 6
}

func (m *DetailModel) SetCommit(c git.CommitDetail) {
	m.commit = &c
	m.vp.SetContent(m.renderCommit(c))
	m.vp.GotoTop()
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, key.NewBinding(key.WithKeys("j", "down"))):
			m.vp.LineDown(1)
		case key.Matches(km, key.NewBinding(key.WithKeys("k", "up"))):
			m.vp.LineUp(1)
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	title := ui.TitleStyle(m.focused).Render(ui.TitleDetail)
	inner := title + "\n" + m.vp.View()
	return ui.PanelStyle(m.focused).Width(m.width).Height(m.height).Render(inner)
}

func (m DetailModel) renderCommit(c git.CommitDetail) string {
	var sb strings.Builder
	sb.WriteString(ui.HeadStyle.Render(ui.DetailLabelCommit+c.Hash[:min(12, len(c.Hash))]) + "\n")
	sb.WriteString(ui.DimItemStyle.Render(ui.DetailLabelAuthor+c.Author) + "\n")
	if !c.Date.IsZero() {
		sb.WriteString(ui.DimItemStyle.Render(ui.DetailLabelDate+c.Date.Format(time.RFC1123)) + "\n")
	}
	if len(c.Tags) > 0 {
		sb.WriteString(ui.TagStyle.Render(ui.DetailLabelTags+strings.Join(c.Tags, ", ")) + "\n")
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
		sb.WriteString("\n" + ui.SectionStyle.Render(ui.DetailLabelFiles) + "\n")
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
