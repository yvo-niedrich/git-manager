package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yvo.niedrich/git-manager/internal/git"
	"github.com/yvo.niedrich/git-manager/internal/model"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "git-manager: cannot determine working directory:", err)
		os.Exit(1)
	}

	repoRoot, err := git.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "git-manager: not inside a git repository")
		os.Exit(1)
	}

	app, err := model.NewApp(repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "git-manager:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "git-manager:", err)
		os.Exit(1)
	}
}
