package git

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// PR is an open GitHub pull request.
type PR struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	HeadRefName string `json:"headRefName"`
}

// GHInstalled reports whether the gh CLI is on PATH.
func GHInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// HasGitHubRemote reports whether any configured remote points to github.com.
func (c *Client) HasGitHubRemote() bool {
	out, err := c.run("remote", "-v")
	if err != nil {
		return false
	}
	return strings.Contains(out, "github.com")
}

// ListOpenPRs returns a map of local branch name → open PR.
// Returns nil if gh is not installed, no GitHub remote exists, or any error occurs.
func (c *Client) ListOpenPRs() map[string]PR {
	if !GHInstalled() || !c.HasGitHubRemote() {
		return nil
	}
	cmd := exec.Command("gh", "pr", "list",
		"--state", "open",
		"--json", "number,url,headRefName",
		"--limit", "100",
	)
	cmd.Dir = c.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var entries []PR
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil
	}
	m := make(map[string]PR, len(entries))
	for _, e := range entries {
		m[e.HeadRefName] = e
	}
	return m
}
