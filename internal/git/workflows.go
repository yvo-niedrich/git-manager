package git

import (
	"fmt"
	"strings"
)

type WorkflowResult struct {
	Message string
	Err     error
}

func (r WorkflowResult) OK() bool { return r.Err == nil }

type Workflows struct {
	c *Client
}

func NewWorkflows(c *Client) *Workflows {
	return &Workflows{c: c}
}

func (w *Workflows) stashAround(label string, fn func() error) WorkflowResult {
	sr, err := w.c.AutoStash(label)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	opErr := fn()
	unstashErr := w.c.AutoUnstash(sr)
	if opErr != nil {
		return WorkflowResult{Err: opErr}
	}
	if unstashErr != nil {
		return WorkflowResult{
			Message: fmt.Sprintf("done, but stash pop failed: %v", unstashErr),
			Err:     unstashErr,
		}
	}
	msg := ""
	if sr.WasNeeded {
		msg = " (uncommitted changes stashed and restored)"
	}
	return WorkflowResult{Message: label + " complete" + msg}
}

func (w *Workflows) CreateBranch(name string) WorkflowResult {
	err := w.c.CreateBranch(name)
	if err != nil {
		return WorkflowResult{Err: fmt.Errorf("create branch %q: %w", name, err)}
	}
	return WorkflowResult{Message: "created branch " + name}
}

func (w *Workflows) SwitchBranch(branch string) WorkflowResult {
	return w.stashAround("switch to "+branch, func() error {
		return w.c.Checkout(branch)
	})
}

func (w *Workflows) CheckoutRemote(remoteBranch string) WorkflowResult {
	return w.stashAround("checkout "+remoteBranch, func() error {
		return w.c.CheckoutNewTracking(remoteBranch)
	})
}

func (w *Workflows) MergeInto(branch string) WorkflowResult {
	return w.stashAround("merge "+branch, func() error {
		return w.c.Merge(branch)
	})
}

func (w *Workflows) RebaseOnto(branch string) WorkflowResult {
	return w.stashAround("rebase onto "+branch, func() error {
		return w.c.Rebase(branch)
	})
}

func (w *Workflows) CherryPick(hash string) WorkflowResult {
	err := w.c.CherryPick(hash)
	if err != nil {
		return WorkflowResult{Err: fmt.Errorf("cherry-pick %s: %w", hash[:7], err)}
	}
	return WorkflowResult{Message: fmt.Sprintf("cherry-picked %s", hash[:7])}
}

func (w *Workflows) RevertCommit(hash string) WorkflowResult {
	err := w.c.RevertCommit(hash)
	if err != nil {
		return WorkflowResult{Err: fmt.Errorf("revert %s: %w", hash[:7], err)}
	}
	return WorkflowResult{Message: fmt.Sprintf("reverted %s", hash[:7])}
}

func (w *Workflows) DropCommit(hash string) WorkflowResult {
	err := w.c.DropCommit(hash)
	if err != nil {
		return WorkflowResult{Err: fmt.Errorf("drop %s: %w", hash[:7], err)}
	}
	return WorkflowResult{Message: fmt.Sprintf("dropped %s", hash[:7])}
}

func (w *Workflows) AmendLast(newMsg string) WorkflowResult {
	err := w.c.AmendCommit(newMsg)
	if err != nil {
		return WorkflowResult{Err: fmt.Errorf("amend: %w", err)}
	}
	return WorkflowResult{Message: "amended HEAD"}
}

func (w *Workflows) SquashCommits(hashes []string) WorkflowResult {
	err := w.c.SquashCommits(hashes)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	return WorkflowResult{Message: fmt.Sprintf("squashed %d commits", len(hashes))}
}

func (w *Workflows) DeleteBranch(name string, force bool) WorkflowResult {
	err := w.c.DeleteBranch(name, force)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	return WorkflowResult{Message: "deleted branch " + name}
}

func (w *Workflows) Push(remote, branch string) WorkflowResult {
	err := w.c.Push(remote, branch)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	return WorkflowResult{Message: fmt.Sprintf("pushed %s to %s", branch, remote)}
}

func (w *Workflows) ForcePush(remote, branch string) WorkflowResult {
	err := w.c.ForcePush(remote, branch)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	return WorkflowResult{Message: fmt.Sprintf("force-pushed %s to %s", branch, remote)}
}

func (w *Workflows) Fetch(remote string) WorkflowResult {
	err := w.c.Fetch(remote)
	if err != nil {
		return WorkflowResult{Err: err}
	}
	return WorkflowResult{Message: "fetched " + remote}
}

// Pull updates branch from its upstream (e.g. "origin/main").
// If branch is current: git pull (stash-aware). Otherwise: fast-forward fetch refspec.
func (w *Workflows) Pull(branch, upstream string) WorkflowResult {
	parts := strings.SplitN(upstream, "/", 2)
	if len(parts) != 2 {
		return WorkflowResult{Err: fmt.Errorf("cannot parse upstream %q", upstream)}
	}
	remote, remoteBranch := parts[0], parts[1]

	if w.c.CurrentBranch() == branch {
		r := w.stashAround("pull "+branch, func() error {
			return w.c.Pull()
		})
		if r.Err == nil {
			r.Message = strings.Replace(r.Message, "pull "+branch+" complete", "pulled "+branch+" from "+upstream, 1)
		}
		return r
	}

	if err := w.c.FetchBranchFromUpstream(remote, remoteBranch, branch); err != nil {
		return WorkflowResult{Err: fmt.Errorf("pull %s: %w", branch, err)}
	}
	return WorkflowResult{Message: fmt.Sprintf("pulled %s from %s", branch, upstream)}
}
