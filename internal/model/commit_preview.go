package model

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yvo.niedrich/git-manager/internal/ui"
)

// CommitPreviewSubmitMsg is emitted when the user proceeds past the file-tree
// preview. It carries no data: the actual set of files to stage is
// recomputed from git status at commit time.
type CommitPreviewSubmitMsg struct{}

// CommitPreviewModel shows the files about to be committed as a tree, in a
// scrollable view, followed by a proceed/cancel choice.
type CommitPreviewModel struct {
	title       string
	lines       []string
	offset      int
	visibleRows int
	proceed     bool
}

func NewCommitPreviewDialog(title string, paths []string, termH int) *CommitPreviewModel {
	visibleRows := termH / 3
	if visibleRows < 5 {
		visibleRows = 5
	}
	if visibleRows > 16 {
		visibleRows = 16
	}
	return &CommitPreviewModel{
		title:       title,
		lines:       buildFileTree(paths),
		visibleRows: visibleRows,
		proceed:     true,
	}
}

func (m *CommitPreviewModel) Priority() int { return 20 }

func (m *CommitPreviewModel) maxOffset() int {
	if len(m.lines) <= m.visibleRows {
		return 0
	}
	return len(m.lines) - m.visibleRows
}

func (m *CommitPreviewModel) DialogUpdate(msg tea.Msg) (DialogContent, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "esc", "q":
		return nil, nil
	case "up", "k":
		if m.offset > 0 {
			m.offset--
		}
	case "down", "j":
		if m.offset < m.maxOffset() {
			m.offset++
		}
	case "left":
		m.proceed = false
	case "right":
		m.proceed = true
	case "enter":
		if m.proceed {
			return nil, func() tea.Msg { return CommitPreviewSubmitMsg{} }
		}
		return nil, nil
	}
	return m, nil
}

func (m *CommitPreviewModel) View() string {
	title := ui.TitleStyle(true).Render(m.title)

	end := m.offset + m.visibleRows
	if end > len(m.lines) {
		end = len(m.lines)
	}
	var tree strings.Builder
	treeLines := make([]string, 0, end-m.offset)
	for i := m.offset; i < end; i++ {
		line := ui.NormalItemStyle.Render(m.lines[i])
		treeLines = append(treeLines, line)
		tree.WriteString(line)
		tree.WriteString("\n")
	}

	hints := ui.KeyHintStyle.Render("[↑↓]") + " " + ui.DescHintStyle.Render(ui.HintScroll+"  ") +
		ui.KeyHintStyle.Render("[←→]") + " " + ui.DescHintStyle.Render(ui.HintSelect+"  ") +
		ui.KeyHintStyle.Render("[Enter]") + " " + ui.DescHintStyle.Render(ui.HintConfirm+"  ") +
		ui.KeyHintStyle.Render("[Esc]") + " " + ui.DescHintStyle.Render(ui.HintCancel)

	// Content width the box will actually render at: the widest of the other lines.
	width := lipgloss.Width(title)
	for _, l := range treeLines {
		if w := lipgloss.Width(l); w > width {
			width = w
		}
	}
	if w := lipgloss.Width(hints); w > width {
		width = w
	}
	buttonsRow := m.renderButtons(width)

	body := fmt.Sprintf("\n  %s\n\n%s\n  %s\n\n  %s",
		title, tree.String(), buttonsRow, hints)
	return ui.MenuBorderStyle.Render(body)
}

// renderButtons centers the Cancel/Proceed choice within width. Every
// segment — the gap between buttons and the centering padding — explicitly
// carries the menu background: embedding an already-rendered ANSI span (the
// selected pill) inside another Render() call resets the background after
// that span, so anything left unstyled around it would show the terminal's
// default background instead of blending with the dialog (see
// internal/ui/CLAUDE.md, "Inline ANSI styling breaks backgrounds").
func (m *CommitPreviewModel) renderButtons(width int) string {
	bg := ui.DimItemStyle.Background(ui.ColorMenuBg)

	cancelLabel, proceedLabel := " Cancel ", " Proceed "
	if m.proceed {
		proceedLabel = ui.SelectedItemStyle.Render(proceedLabel)
		cancelLabel = bg.Render(cancelLabel)
	} else {
		proceedLabel = bg.Render(proceedLabel)
		cancelLabel = ui.SelectedItemStyle.Render(cancelLabel)
	}
	buttons := cancelLabel + bg.Render("   ") + proceedLabel

	pad := width - lipgloss.Width(buttons)
	if pad < 0 {
		pad = 0
	}
	left, right := pad/2, pad-pad/2
	return bg.Render(strings.Repeat(" ", left)) + buttons + bg.Render(strings.Repeat(" ", right))
}

// fileTreeNode is an intermediate structure used only to render a flat list
// of changed paths as an indented tree.
type fileTreeNode struct {
	children map[string]*fileTreeNode
	order    []string
	isFile   bool
}

func newFileTreeNode() *fileTreeNode {
	return &fileTreeNode{children: map[string]*fileTreeNode{}}
}

// buildFileTree turns flat repo-relative paths into rendered tree lines
// (directories before files, alphabetical within each group).
func buildFileTree(paths []string) []string {
	root := newFileTreeNode()
	for _, p := range paths {
		node := root
		for _, part := range strings.Split(p, "/") {
			child, ok := node.children[part]
			if !ok {
				child = newFileTreeNode()
				node.children[part] = child
				node.order = append(node.order, part)
			}
			node = child
		}
		node.isFile = true
	}
	var lines []string
	appendTreeLines(root, "", &lines)
	return lines
}

func appendTreeLines(node *fileTreeNode, prefix string, lines *[]string) {
	names := append([]string(nil), node.order...)
	sort.Slice(names, func(i, j int) bool {
		ci, cj := node.children[names[i]], node.children[names[j]]
		if ci.isFile != cj.isFile {
			return !ci.isFile
		}
		return names[i] < names[j]
	})
	for i, name := range names {
		child := node.children[name]
		connector, childPrefix := "├── ", prefix+"│   "
		if i == len(names)-1 {
			connector, childPrefix = "└── ", prefix+"    "
		}
		label := name
		if !child.isFile {
			label += "/"
		}
		*lines = append(*lines, prefix+connector+label)
		if !child.isFile {
			appendTreeLines(child, childPrefix, lines)
		}
	}
}
