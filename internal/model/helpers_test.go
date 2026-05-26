package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/git"
	"github.com/user/gitmg/internal/testutil"
)

// press builds a tea.KeyMsg from a key name string, matching the same
// strings used in key.Matches bindings throughout the model package.
func press(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

// cmdMsg calls cmd and returns the resulting tea.Msg — useful for asserting
// what message a tea.Cmd will emit without running the Elm runtime.
func cmdMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func makeBranches() []git.Branch {
	return []git.Branch{
		{Name: "main", IsCurrent: true, Upstream: "origin/main"},
		{Name: "feature/a"},
		{Name: "main", IsRemote: true, Remote: "origin"},
	}
}

func makeCommits() []git.Commit {
	return []git.Commit{
		{Hash: "aaaa1111", ShortHash: "aaaa111", Subject: "third commit",  Author: "Alice", IsHead: true},
		{Hash: "bbbb2222", ShortHash: "bbbb222", Subject: "second commit", Author: "Bob",   Tags: []string{"v1.0"}},
		{Hash: "cccc3333", ShortHash: "cccc333", Subject: "first commit",  Author: "Alice"},
	}
}

func fixtureApp(t *testing.T) *App {
	t.Helper()
	a, err := NewApp(testutil.FixtureRepo(t))
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	return a
}
