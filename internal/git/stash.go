package git

import (
	"fmt"
	"strings"
)

type StashResult struct {
	Ref       string
	WasNeeded bool
}

func (c *Client) AutoStash(label string) (StashResult, error) {
	if !c.HasUncommittedChanges() {
		return StashResult{WasNeeded: false}, nil
	}
	msg := fmt.Sprintf("gitmg: %s", label)
	_, err := c.run("stash", "push", "-u", "-m", msg)
	if err != nil {
		return StashResult{}, fmt.Errorf("stash failed: %w", err)
	}
	// get the ref of the stash we just created
	ref, err := c.run("rev-parse", "stash@{0}")
	if err != nil {
		ref = "stash@{0}"
	}
	return StashResult{Ref: ref, WasNeeded: true}, nil
}

func (c *Client) AutoUnstash(r StashResult) error {
	if !r.WasNeeded {
		return nil
	}
	// find the stash entry by its commit hash
	lines, err := c.runLines("stash", "list", "--format=%gd\t%H")
	if err != nil {
		return fmt.Errorf("stash list failed: %w", err)
	}
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 && parts[1] == r.Ref {
			if _, err = c.run("stash", "pop", parts[0]); err != nil {
				return fmt.Errorf("stash pop failed: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("stash entry %s not found — pop manually: git stash list", r.Ref)
}
