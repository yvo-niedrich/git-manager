package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HintSet []struct{ Key, Desc string }

func BranchHints(isRemote, isCurrent, hasUpstream bool) HintSet {
	base := HintSet{
		{"tab", "next panel"},
		{"j/k", "navigate"},
		{"Enter/x", "menu"},
	}
	switch {
	case isRemote:
		base = append(base, HintSet{{"c", "checkout"}, {"f", "fetch"}}...)
	case isCurrent:
		acts := HintSet{{"m", "merge"}, {"r", "rebase"}, {"p", "push"}}
		if hasUpstream {
			acts = append(acts, struct{ Key, Desc string }{"l", "pull"})
		}
		base = append(base, acts...)
	default:
		acts := HintSet{{"c", "checkout"}, {"m", "merge"}, {"r", "rebase"}, {"p", "push"}}
		if hasUpstream {
			acts = append(acts, struct{ Key, Desc string }{"l", "pull"})
		}
		acts = append(acts, struct{ Key, Desc string }{"D", "delete"})
		base = append(base, acts...)
	}
	return append(base, HintSet{{"/", "filter"}, {"q", "quit"}}...)
}

func CommitHints(isHead, multiSelect bool) HintSet {
	if multiSelect {
		return HintSet{
			{"j/k", "navigate"},
			{"space", "toggle"},
			{"Enter", "squash"},
			{"s", "exit multi-select"},
			{"q", "quit"},
		}
	}
	base := HintSet{
		{"tab", "next panel"},
		{"j/k", "navigate"},
		{"Enter/x", "menu"},
		{"p", "cherry-pick"},
		{"R", "revert"},
	}
	if isHead {
		base = append(base, struct{ Key, Desc string }{"a", "amend"})
	} else {
		base = append(base, struct{ Key, Desc string }{"d", "drop"})
	}
	return append(base, HintSet{{"s", "multi-select"}, {"/", "filter"}, {"q", "quit"}}...)
}

var FilterHints = HintSet{
	{"type", "filter"},
	{"Enter", "apply"},
	{"Esc", "clear"},
}

var DetailHints = HintSet{
	{"tab", "next panel"},
	{"j/k", "scroll"},
	{"Esc", "cancel"},
	{"q", "quit"},
}

func RenderStatusBar(width int, msg string, isErr bool, hints HintSet) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, KeyHintStyle.Render("["+h.Key+"]")+" "+DescHintStyle.Render(h.Desc))
	}
	hintLine := strings.Join(parts, "  ")

	msgStyled := ""
	if msg != "" {
		if isErr {
			msgStyled = StatusErrStyle.Render("✗ " + msg)
		} else {
			msgStyled = StatusOKStyle.Render("✓ " + msg)
		}
	}

	line := hintLine
	if msgStyled != "" {
		gap := width - lipgloss.Width(hintLine) - lipgloss.Width(msgStyled) - 2
		if gap > 0 {
			line = hintLine + strings.Repeat(" ", gap) + msgStyled
		} else {
			line = msgStyled
		}
	}

	return StatusBarStyle.Width(width).Render(fmt.Sprintf(" %s ", line))
}
