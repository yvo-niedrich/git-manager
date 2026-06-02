package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HintSet []struct{ Key, Desc string }

func BranchHints(isRemote, isCurrent, hasUpstream bool) HintSet {
	base := HintSet{
		{"tab", HintNextPanel},
		{"j/k", HintNavigate},
		{"Enter/x", HintMenu},
	}
	switch {
	case isRemote:
		base = append(base, HintSet{{"c", HintCheckout}, {"f", HintFetch}}...)
	case isCurrent:
		acts := HintSet{{"m", HintMerge}, {"r", HintRebase}, {"p", HintPush}}
		if hasUpstream {
			acts = append(acts, struct{ Key, Desc string }{"l", HintPull})
		}
		acts = append(acts, struct{ Key, Desc string }{"n", HintNewBranch})
		base = append(base, acts...)
	default:
		acts := HintSet{{"c", HintCheckout}, {"m", HintMerge}, {"r", HintRebase}, {"p", HintPush}}
		if hasUpstream {
			acts = append(acts, struct{ Key, Desc string }{"l", HintPull})
		}
		acts = append(acts, struct{ Key, Desc string }{"D", HintDelete})
		acts = append(acts, struct{ Key, Desc string }{"n", HintNewBranch})
		base = append(base, acts...)
	}
	return append(base, HintSet{{"/", HintFilter}, {"q", HintQuit}}...)
}

func CommitHints(isHead, multiSelect bool) HintSet {
	if multiSelect {
		return HintSet{
			{"j/k", HintNavigate},
			{"space", HintToggle},
			{"Enter", HintSquash},
			{"s", HintExitMultiSelect},
			{"q", HintQuit},
		}
	}
	base := HintSet{
		{"tab", HintNextPanel},
		{"j/k", HintNavigate},
		{"Enter/x", HintMenu},
		{"p", HintCherryPick},
		{"R", HintRevert},
	}
	if isHead {
		base = append(base, struct{ Key, Desc string }{"a", HintAmend})
	} else {
		base = append(base, struct{ Key, Desc string }{"d", HintDrop})
	}
	return append(base, HintSet{{"s", HintMultiSelect}, {"/", HintFilter}, {"q", HintQuit}}...)
}

var FilterHints = HintSet{
	{"type", HintFilter},
	{"Enter", HintApply},
	{"Esc", HintClear},
}

var DetailHints = HintSet{
	{"tab", HintNextPanel},
	{"j/k", HintScroll},
	{"Esc", HintCancel},
	{"q", HintQuit},
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
			msgStyled = StatusErrStyle.Render(StatusErrPrefix + msg)
		} else {
			msgStyled = StatusOKStyle.Render(StatusOKPrefix + msg)
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
