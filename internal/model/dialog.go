package model

import tea "github.com/charmbracelet/bubbletea"

// DialogContent is implemented by each overlay dialog.
// DialogUpdate returns (nil, cmd) to signal the dialog should close;
// cmd may carry a follow-up message (e.g. MenuSelectedMsg, AmendSubmitMsg).
type DialogContent interface {
	Priority() int
	DialogUpdate(tea.Msg) (DialogContent, tea.Cmd)
	View() string
}

// dialogStack holds 0-N open dialogs. The one with the highest Priority()
// captures input and is rendered as the overlay.
type dialogStack struct {
	items []DialogContent
}

func (ds *dialogStack) Push(d DialogContent) {
	ds.items = append(ds.items, d)
}

func (ds *dialogStack) activeIndex() int {
	best := -1
	for i, d := range ds.items {
		if best < 0 || d.Priority() > ds.items[best].Priority() {
			best = i
		}
	}
	return best
}

func (ds *dialogStack) Active() DialogContent {
	if i := ds.activeIndex(); i >= 0 {
		return ds.items[i]
	}
	return nil
}

func (ds *dialogStack) Update(msg tea.Msg) tea.Cmd {
	i := ds.activeIndex()
	if i < 0 {
		return nil
	}
	next, cmd := ds.items[i].DialogUpdate(msg)
	if next == nil {
		ds.items = append(ds.items[:i], ds.items[i+1:]...)
	} else {
		ds.items[i] = next
	}
	return cmd
}

func (ds *dialogStack) IsOpen() bool {
	return len(ds.items) > 0
}
