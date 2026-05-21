package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/gitmg/internal/git"
	"github.com/user/gitmg/internal/ui"
)

type CommitsModel struct {
	commits      []git.Commit
	cursor       int
	offset       int
	width        int
	height       int
	focused      bool
	multiSelect  bool
	selected     map[int]bool
	branchRef    string
	filterInput  textinput.Model
	filterActive bool
}

type CommitsLoadedMsg struct{ Commits []git.Commit }
type CommitSelectedMsg struct{ Commit git.Commit }

func NewCommitsModel() CommitsModel {
	ti := textinput.New()
	ti.Placeholder = "hash, message, author..."
	ti.CharLimit = 64
	ti.Prompt = "/ "
	ti.PromptStyle = ui.DimItemStyle
	ti.TextStyle = ui.NormalItemStyle
	ti.PlaceholderStyle = ui.DimItemStyle
	return CommitsModel{selected: map[int]bool{}, filterInput: ti}
}

func (m *CommitsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.filterInput.Width = ui.InnerWidth(w) - 4
}

func (m *CommitsModel) SetCommits(commits []git.Commit) {
	m.commits = commits
	m.cursor = 0
	m.offset = 0
	m.selected = map[int]bool{}
	m.multiSelect = false
	m.filterInput.SetValue("")
	m.filterActive = false
}

func (m CommitsModel) IsFiltering() bool {
	return m.filterActive
}

func (m CommitsModel) IsMultiSelect() bool {
	return m.multiSelect
}

func (m CommitsModel) filteredCommits() []git.Commit {
	q := strings.ToLower(m.filterInput.Value())
	if q == "" {
		return m.commits
	}
	var out []git.Commit
	for _, c := range m.commits {
		if strings.HasPrefix(c.Hash, q) ||
			strings.HasPrefix(c.ShortHash, q) ||
			strings.Contains(strings.ToLower(c.Subject), q) ||
			strings.Contains(strings.ToLower(c.Author), q) {
			out = append(out, c)
		}
	}
	return out
}

func (m CommitsModel) Selected() *git.Commit {
	filtered := m.filteredCommits()
	if len(filtered) == 0 || m.cursor >= len(filtered) {
		return nil
	}
	c := filtered[m.cursor]
	return &c
}

func (m CommitsModel) SelectedHashes() []string {
	filtered := m.filteredCommits()
	if len(m.selected) == 0 {
		if sel := m.Selected(); sel != nil {
			return []string{sel.Hash}
		}
		return nil
	}
	var hashes []string
	for i := range filtered {
		if m.selected[i] {
			hashes = append(hashes, filtered[i].Hash)
		}
	}
	return hashes
}

func (m CommitsModel) Update(msg tea.Msg) (CommitsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterActive {
			switch msg.String() {
			case "esc":
				m.filterActive = false
				m.filterInput.SetValue("")
				m.filterInput.Blur()
				m.cursor = 0
				m.offset = 0
				return m, m.selectionCmd()
			case "enter":
				m.filterActive = false
				m.filterInput.Blur()
				return m, nil
			default:
				prev := m.filterInput.Value()
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				if m.filterInput.Value() != prev {
					m.cursor = 0
					m.offset = 0
				}
				return m, tea.Batch(cmd, m.selectionCmd())
			}
		}
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			m.filterActive = true
			m.multiSelect = false
			m.selected = map[int]bool{}
			m.cursor = 0
			m.offset = 0
			return m, m.filterInput.Focus()
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			if m.filterInput.Value() != "" {
				m.filterInput.SetValue("")
				m.cursor = 0
				m.offset = 0
				return m, m.selectionCmd()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			filtered := m.filteredCommits()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				m.clampOffset()
			}
			return m, m.selectionCmd()
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
			return m, m.selectionCmd()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))):
			filtered := m.filteredCommits()
			m.cursor = min(m.cursor+5, len(filtered)-1)
			m.clampOffset()
			return m, m.selectionCmd()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			m.cursor = max(m.cursor-5, 0)
			m.clampOffset()
			return m, m.selectionCmd()
		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
			if m.multiSelect && m.filterInput.Value() == "" {
				if m.selected[m.cursor] {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = true
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			if m.filterInput.Value() == "" {
				m.multiSelect = !m.multiSelect
				if !m.multiSelect {
					m.selected = map[int]bool{}
				}
			}
		}
	}
	return m, nil
}

func (m *CommitsModel) clampOffset() {
	innerH := m.height - 6
	if innerH < 1 {
		innerH = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+innerH {
		m.offset = m.cursor - innerH + 1
	}
}

func (m CommitsModel) selectionCmd() tea.Cmd {
	if sel := m.Selected(); sel != nil {
		c := *sel
		return func() tea.Msg { return CommitSelectedMsg{Commit: c} }
	}
	return nil
}

func (m CommitsModel) View() string {
	title := ui.TitleStyle(m.focused).Render("Commits")
	if m.branchRef != "" {
		title += ui.DimItemStyle.Render("  " + m.branchRef)
	}
	if m.multiSelect {
		title += ui.HeadStyle.Render("  [multi-select: space=toggle, Enter=squash]")
	}

	innerW := ui.InnerWidth(m.width)
	innerH := m.height - 6
	if innerH < 1 {
		innerH = 1
	}

	var filterLine string
	switch {
	case m.filterActive:
		filterLine = "  " + m.filterInput.View()
	case m.filterInput.Value() != "":
		filterLine = ui.KeyHintStyle.Render("  ~") + " " + ui.NormalItemStyle.Render(m.filterInput.Value())
	default:
		filterLine = ui.DimItemStyle.Render("  / filter")
	}
	separator := ui.DimItemStyle.Render(strings.Repeat("─", innerW))

	filtered := m.filteredCommits()
	var lines []string
	count := 0
	for i, c := range filtered {
		if i < m.offset {
			continue
		}
		if count >= innerH {
			break
		}

		// Right-aligned tag indicator
		tagLabel := ""
		switch {
		case len(c.Tags) == 1:
			tagLabel = "[" + c.Tags[0] + "]"
		case len(c.Tags) > 1:
			tagLabel = fmt.Sprintf("[%d]", len(c.Tags))
		}
		tagW := 0
		if tagLabel != "" {
			tagW = lipgloss.Width(tagLabel) + 1 // +1 for gap between text and tag
		}
		availW := innerW - tagW

		plainBullet := "●"
		if c.IsHead {
			plainBullet = "◉"
		}
		bullet := plainBullet
		if c.IsHead {
			bullet = ui.HeadStyle.Render(plainBullet)
		}

		hash := ui.DimItemStyle.Render(c.ShortHash)
		subj := c.Subject
		text := fmt.Sprintf("%s %s %s", bullet, hash, subj)

		vis := lipgloss.Width(text)
		if vis > availW {
			over := vis - availW
			runes := []rune(subj)
			if over < len(runes) {
				subj = string(runes[:len(runes)-over-1]) + "…"
			}
			text = fmt.Sprintf("%s %s %s", bullet, hash, subj)
		}
		if tagLabel != "" {
			gap := availW - lipgloss.Width(text)
			if gap < 0 {
				gap = 0
			}
			text += strings.Repeat(" ", gap) + " " + ui.TagStyle.Render(tagLabel)
		}

		isSelected := m.selected[i]
		if i == m.cursor || isSelected {
			plain := fmt.Sprintf("%s %s %s", plainBullet, c.ShortHash, c.Subject)
			w := lipgloss.Width(plain)
			if w > availW {
				over := w - availW
				runes := []rune(c.Subject)
				if over < len(runes) {
					plain = fmt.Sprintf("%s %s %s…", plainBullet, c.ShortHash, string(runes[:len(runes)-over-1]))
				} else {
					plain = string([]rune(plain)[:availW])
				}
			}
			bgColor := ui.ColorSelected
			if isSelected && i != m.cursor {
				bgColor = ui.ColorMultiSelect
			}
			selStyle := lipgloss.NewStyle().Background(bgColor).Foreground(ui.ColorText)
			if tagLabel != "" {
				gap := availW - lipgloss.Width(plain)
				if gap < 0 {
					gap = 0
				}
				leftPart := selStyle.Render(plain + strings.Repeat(" ", gap))
				tagPart := lipgloss.NewStyle().Background(bgColor).Foreground(ui.ColorWarn).Bold(true).Render(" " + tagLabel)
				text = leftPart + tagPart
			} else {
				text = selStyle.Width(innerW).Render(plain)
			}
		}

		lines = append(lines, text)
		count++
	}

	body := strings.Join(lines, "\n")
	inner := title + "\n" + filterLine + "\n" + separator + "\n" + body
	return ui.PanelStyle(m.focused).Width(m.width).Height(m.height).Render(inner)
}
