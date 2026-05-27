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

type BranchesModel struct {
	branches     []git.Branch
	cursor       int
	offset       int
	width        int
	height       int
	focused      bool
	filterInput  textinput.Model
	filterActive bool
}

type BranchesLoadedMsg struct{ Branches []git.Branch }
type BranchSelectedMsg struct{ Branch git.Branch }

func NewBranchesModel() BranchesModel {
	ti := textinput.New()
	ti.Placeholder = ui.PlaceholderBranches
	ti.CharLimit = 64
	ti.Prompt = "/ "
	ti.PromptStyle = ui.DimItemStyle
	ti.TextStyle = ui.NormalItemStyle
	ti.PlaceholderStyle = ui.DimItemStyle
	return BranchesModel{filterInput: ti}
}

func (m *BranchesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.filterInput.Width = ui.InnerWidth(w) - 4
}

func (m *BranchesModel) SetBranches(branches []git.Branch) {
	m.branches = branches
	if m.cursor > len(m.filteredBranches()) {
		m.cursor = 0
		m.offset = 0
	}
}

func (m BranchesModel) IsFiltering() bool {
	return m.filterActive
}

// IsNewBranchSelected reports whether the cursor is on the «new branch» button.
func (m BranchesModel) IsNewBranchSelected() bool {
	return m.cursor == m.localCount()
}

func (m BranchesModel) filteredBranches() []git.Branch {
	q := strings.ToLower(m.filterInput.Value())
	if q == "" {
		return m.branches
	}
	var out []git.Branch
	for _, b := range m.branches {
		if strings.Contains(strings.ToLower(b.Name), q) ||
			(b.IsRemote && strings.Contains(strings.ToLower(b.Remote), q)) {
			out = append(out, b)
		}
	}
	return out
}

// localCount returns the number of local branches in the filtered list.
// The virtual «new branch» button sits at cursor position localCount().
func (m BranchesModel) localCount() int {
	for i, b := range m.filteredBranches() {
		if b.IsRemote {
			return i
		}
	}
	return len(m.filteredBranches())
}

// LocalNames returns the names of all local branches currently loaded.
func (m BranchesModel) LocalNames() []string {
	var out []string
	for _, b := range m.branches {
		if !b.IsRemote {
			out = append(out, b.Name)
		}
	}
	return out
}

func (m BranchesModel) Selected() *git.Branch {
	if m.IsNewBranchSelected() {
		return nil
	}
	filtered := m.filteredBranches()
	lc := m.localCount()
	idx := m.cursor
	if m.cursor > lc {
		idx = m.cursor - 1
	}
	if idx < 0 || idx >= len(filtered) {
		return nil
	}
	b := filtered[idx]
	return &b
}

func (m BranchesModel) Update(msg tea.Msg) (BranchesModel, tea.Cmd) {
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
			filtered := m.filteredBranches()
			if m.cursor < len(filtered) {
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
			filtered := m.filteredBranches()
			m.cursor = min(m.cursor+5, len(filtered))
			m.clampOffset()
			return m, m.selectionCmd()
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			m.cursor = max(m.cursor-5, 0)
			m.clampOffset()
			return m, m.selectionCmd()
		}
	}
	return m, nil
}

func (m *BranchesModel) clampOffset() {
	innerH := m.height - 8
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

func (m BranchesModel) selectionCmd() tea.Cmd {
	if m.IsNewBranchSelected() {
		return nil
	}
	if sel := m.Selected(); sel != nil {
		b := *sel
		return func() tea.Msg { return BranchSelectedMsg{Branch: b} }
	}
	return nil
}

func (m BranchesModel) View() string {
	title := ui.TitleStyle(m.focused).Render(ui.TitleBranches)
	innerW := ui.InnerWidth(m.width)
	innerH := m.height - 8
	if innerH < 1 {
		innerH = 1
	}

	var filterLine string
	switch {
	case m.filterActive:
		filterLine = "  " + m.filterInput.View()
	case m.filterInput.Value() != "":
		filterLine = ui.KeyHintStyle.Render(ui.FilterActiveMark) + " " + ui.NormalItemStyle.Render(m.filterInput.Value())
	default:
		filterLine = ui.DimItemStyle.Render(ui.FilterPrompt)
	}
	separator := ui.DimItemStyle.Render(strings.Repeat("─", innerW))

	remoteSet := map[string]bool{}
	for _, b := range m.branches {
		if b.IsRemote && b.Remote != "" {
			remoteSet[b.Remote] = true
		}
	}
	multiRemote := len(remoteSet) > 1

	filtered := m.filteredBranches()
	lc := m.localCount()
	totalVirtual := len(filtered) + 1 // +1 for the «new branch» button at position lc

	var lines []string
	localHeaderDone := false
	remoteHeaderDone := false
	lastRemote := ""

	count := 0
	for vi := m.offset; vi < totalVirtual && count < innerH; vi++ {
		isButton := vi == lc

		var b *git.Branch
		if !isButton {
			idx := vi
			if vi > lc {
				idx = vi - 1
			}
			if idx < len(filtered) {
				bv := filtered[idx]
				b = &bv
			}
		}

		// LOCAL section header — before first local branch or before the button
		if !localHeaderDone && (isButton || (b != nil && !b.IsRemote)) {
			lines = append(lines, ui.SectionStyle.Render(ui.SectionLocal))
			count++
			localHeaderDone = true
			if count >= innerH {
				break
			}
		}

		// REMOTE section header
		if !isButton && b != nil && b.IsRemote && !remoteHeaderDone {
			if count < innerH {
				lines = append(lines, ui.SectionStyle.Render(ui.SectionRemote))
				count++
			}
			remoteHeaderDone = true
			if count >= innerH {
				break
			}
		}

		// Multi-remote sub-header
		if !isButton && b != nil && b.IsRemote && multiRemote && b.Remote != lastRemote {
			if count < innerH {
				lines = append(lines, ui.SectionStyle.Render("  "+b.Remote))
				count++
			}
			lastRemote = b.Remote
			if count >= innerH {
				break
			}
		}

		if isButton {
			var text string
			if m.IsNewBranchSelected() {
				text = ui.SelectedItemStyle.Width(innerW).Render("  " + ui.NewBranchButtonLabel)
			} else {
				text = ui.DimItemStyle.Render("  " + ui.NewBranchButtonLabel)
			}
			lines = append(lines, text)
			count++
			continue
		}

		if b == nil {
			continue
		}

		indent := "  "
		if b.IsRemote && multiRemote {
			indent = "    "
		}
		prefix := indent
		if b.IsCurrent {
			prefix = ui.HeadStyle.Render(ui.MarkerCurrentBranch)
		}

		label := b.Name
		if b.IsRemote {
			label = ui.RemoteStyle.Render(label)
		}
		text := fmt.Sprintf("%s%s", prefix, label)
		visLen := lipgloss.Width(text)
		if visLen > innerW {
			over := visLen - innerW
			truncated := b.Name
			if over < len(truncated) {
				truncated = truncated[:len(truncated)-over-1] + "…"
			}
			if b.IsRemote {
				label = ui.RemoteStyle.Render(truncated)
			} else {
				label = truncated
			}
			text = fmt.Sprintf("%s%s", prefix, label)
		}
		if vi == m.cursor {
			namePrefix := indent
			if b.IsCurrent {
				namePrefix = ui.MarkerCurrentBranch
			}
			if !b.IsRemote && b.Upstream != "" {
				arrow := ui.MarkerUpstreamArrow + b.Upstream
				namePart := ui.SelectedItemStyle.Render(namePrefix + b.Name)
				hintPart := ui.SelectedItemDimStyle.Render(arrow)
				combined := namePart + hintPart
				combinedW := lipgloss.Width(combined)
				if combinedW <= innerW {
					pad := ui.SelectedItemDimStyle.Render(strings.Repeat(" ", innerW-combinedW))
					text = combined + pad
				} else {
					plain := namePrefix + b.Name
					if len(plain) > innerW {
						plain = plain[:innerW]
					}
					text = ui.SelectedItemStyle.Width(innerW).Render(plain)
				}
			} else {
				plain := namePrefix + b.Name
				if len(plain) > innerW {
					plain = plain[:innerW]
				}
				text = ui.SelectedItemStyle.Width(innerW).Render(plain)
			}
		}
		lines = append(lines, text)
		count++
	}

	body := strings.Join(lines, "\n")
	inner := title + "\n" + filterLine + "\n" + separator + "\n" + body
	return ui.PanelStyle(m.focused).Width(m.width).Height(m.height).Render(inner)
}
