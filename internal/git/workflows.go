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

// stashAroundCheckout is like stashAround but explicitly manages a cross-branch
// checkout: it switches to target, runs fn, then returns to original before
// popping the stash. If the return checkout fails the stash is intentionally
// left in place to avoid applying changes to the wrong branch.
func (w *Workflows) stashAroundCheckout(label, original, target string, fn func() error) WorkflowResult {
	sr, err := w.c.AutoStash(label)
	if err != nil {
		return WorkflowResult{Err: err}
	}

	if original != target {
		if err := w.c.Checkout(target); err != nil {
			_ = w.c.AutoUnstash(sr) // still on original; safe to pop
			return WorkflowResult{Err: fmt.Errorf("checkout %s: %w", target, err)}
		}
	}

	opErr := fn()

	if original != target {
		if checkoutErr := w.c.Checkout(original); checkoutErr != nil {
			// Cannot safely pop the stash from the wrong branch. Leave it intact.
			if opErr != nil {
				return WorkflowResult{Err: fmt.Errorf("%v; could not return to %s (%w) — stash not restored, pop manually on %s", opErr, original, checkoutErr, original)}
			}
			return WorkflowResult{Err: fmt.Errorf("could not return to %s: %w — stash not restored, pop manually on %s", original, checkoutErr, original)}
		}
	}

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

// CreateBranch creates a new branch from `from` and checks it out.
// When `from` is the current branch, uncommitted changes carry forward naturally
// (git preserves them on the new branch). When `from` differs, stashAround
// ensures uncommitted changes are saved and restored on the new branch.
func (w *Workflows) CreateBranch(name, from string) WorkflowResult {
	current := w.c.CurrentBranch()
	label := fmt.Sprintf("create branch %s", name)
	if from != "" && from != current {
		// Switching to a different base: stash first to avoid dirty-tree conflicts.
		return w.stashAround(label, func() error {
			return w.c.CreateBranch(name, from)
		})
	}
	if err := w.c.CreateBranch(name, ""); err != nil {
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

// MergeInto merges source into target. The operation is always stash-aware:
// uncommitted changes are stashed, the operation runs, then the stash is
// restored on the original branch.
func (w *Workflows) MergeInto(source, target string) WorkflowResult {
	original := w.c.CurrentBranch()
	label := fmt.Sprintf("merge %s into %s", source, target)
	return w.stashAroundCheckout(label, original, target, func() error {
		mergeErr := w.c.Merge(source)
		if mergeErr != nil {
			w.c.run("merge", "--abort") //nolint — best-effort cleanup
		}
		return mergeErr
	})
}

// RebaseOnto rebases source onto onto. The operation is always stash-aware:
// uncommitted changes are stashed, the operation runs, then the stash is
// restored on the original branch.
func (w *Workflows) RebaseOnto(source, onto string) WorkflowResult {
	original := w.c.CurrentBranch()
	label := fmt.Sprintf("rebase %s onto %s", source, onto)
	return w.stashAroundCheckout(label, original, source, func() error {
		rebaseErr := w.c.Rebase(onto)
		if rebaseErr != nil {
			w.c.run("rebase", "--abort") //nolint — best-effort cleanup
		}
		return rebaseErr
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

// ErrNotFullyMerged is returned by DeleteBranch when git refuses the safe
// delete because the branch contains commits unreachable from any other branch.
type ErrNotFullyMerged struct{ Branch string }

func (e ErrNotFullyMerged) Error() string {
	return fmt.Sprintf("branch %q has unmerged commits", e.Branch)
}

func (w *Workflows) DeleteBranch(name string, force bool) WorkflowResult {
	err := w.c.DeleteBranch(name, force)
	if err != nil {
		if !force && strings.Contains(err.Error(), "not fully merged") {
			return WorkflowResult{Err: ErrNotFullyMerged{Branch: name}}
		}
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
