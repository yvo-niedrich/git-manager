package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/gitmg/internal/git"
	"github.com/user/gitmg/internal/model"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitwise: cannot determine working directory:", err)
		os.Exit(1)
	}

	repoRoot, err := git.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitwise: not inside a git repository")
		os.Exit(1)
	}

	app, err := model.NewApp(repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gitwise:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "gitwise:", err)
		os.Exit(1)
	}
}
