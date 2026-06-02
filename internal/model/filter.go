package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// applyFilterKey handles the Esc / Enter / default key logic that is identical
// for every panel that has an active text filter (branches, commits).
//
// It mutates input and active in-place and returns:
//   - resetCursor: caller should zero cursor and offset
//   - emitSelection: caller should include a selectionCmd in the returned batch
//   - inputCmd: the textinput's own update command (nil for Esc/Enter)
func applyFilterKey(
	msg tea.KeyMsg,
	input *textinput.Model,
	active *bool,
) (resetCursor, emitSelection bool, inputCmd tea.Cmd) {
	switch msg.String() {
	case "esc":
		*active = false
		input.SetValue("")
		input.Blur()
		return true, true, nil
	case "enter":
		*active = false
		input.Blur()
		return false, false, nil
	default:
		prev := input.Value()
		*input, inputCmd = input.Update(msg)
		changed := input.Value() != prev
		return changed, true, inputCmd
	}
}
