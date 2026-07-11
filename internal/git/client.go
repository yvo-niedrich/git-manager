package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yvo.niedrich/git-manager/internal/config"
)

type Branch struct {
	Name      string
	IsRemote  bool
	IsCurrent bool
	Remote    string // remote name for remote tracking branches (e.g. "origin")
	Upstream  string // upstream tracking ref for local branches (e.g. "origin/main")
}

// FullRef returns the ref suitable for git commands: "remote/name" for remote
// tracking branches, just "name" for local branches.
func (b Branch) FullRef() string {
	if b.IsRemote && b.Remote != "" {
		return b.Remote + "/" + b.Name
	}
	return b.Name
}

type Commit struct {
	Hash      string
	ShortHash string
	Subject   string
	Author    string
	Date      time.Time
	IsHead    bool
	Tags      []string
}

type CommitDetail struct {
	Commit
	Body      string
	StatLines []string // e.g. ["src/foo.go | 8 ++------"]
	Diff      string
}

type Client struct {
	repoRoot string
	cfg      config.Config
}

func NewClient(repoRoot string, cfg config.Config) *Client {
	return &Client{repoRoot: repoRoot, cfg: cfg}
}

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.repoRoot
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (c *Client) runLines(args ...string) ([]string, error) {
	out, err := c.run(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func (c *Client) CurrentBranch() string {
	out, err := c.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "HEAD"
	}
	if out == "HEAD" {
		// Detached HEAD — capture the exact commit hash so Checkout can re-detach
		// at the pre-operation position rather than at a newly-created merge commit.
		if hash, err := c.run("rev-parse", "HEAD"); err == nil {
			return hash
		}
	}
	return out
}

func (c *Client) HasUncommittedChanges() bool {
	out, err := c.run("status", "--porcelain")
	return err == nil && strings.TrimSpace(out) != ""
}

// StatusResult splits uncommitted changes into paths git already tracks
// (modified, deleted, renamed) and paths it doesn't know about yet.
type StatusResult struct {
	Tracked   []string
	Untracked []string
}

// Status parses "git status --porcelain=v1 -z" (NUL-separated so paths with
// spaces or special characters are not mangled by shell-style quoting).
func (c *Client) Status() (StatusResult, error) {
	out, err := c.run("status", "--porcelain=v1", "-z")
	if err != nil {
		return StatusResult{}, err
	}
	var result StatusResult
	fields := strings.Split(out, "\x00")
	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if len(entry) < 4 {
			continue
		}
		x, y := entry[0], entry[1]
		path := entry[3:]
		if x == '?' && y == '?' {
			result.Untracked = append(result.Untracked, path)
			continue
		}
		result.Tracked = append(result.Tracked, path)
		if x == 'R' || x == 'C' {
			i++ // renames/copies carry the old path as a second NUL-terminated field
		}
	}
	return result, nil
}

// StageTrackedChanges stages modifications to files git already tracks
// (equivalent to "git add -u"), leaving untracked files alone.
func (c *Client) StageTrackedChanges() error {
	_, err := c.run("add", "-u")
	return err
}

// StagePaths stages the given paths explicitly (used for untracked files).
func (c *Client) StagePaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	_, err := c.run(args...)
	return err
}

func (c *Client) Commit(message string) error {
	_, err := c.run("commit", "-m", message)
	return err
}

func (c *Client) ListBranches() ([]Branch, error) {
	lines, err := c.runLines("branch", "-a", "--format=%(refname)\t%(HEAD)\t%(upstream:short)")
	if err != nil {
		return nil, err
	}
	current := c.CurrentBranch()
	var branches []Branch
	seen := map[string]bool{}
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		ref := parts[0]
		isRemote := strings.HasPrefix(ref, "refs/remotes/")
		var name, remote string
		switch {
		case isRemote:
			withoutPrefix := strings.TrimPrefix(ref, "refs/remotes/")
			if idx := strings.Index(withoutPrefix, "/"); idx >= 0 {
				remote = withoutPrefix[:idx]
				name = withoutPrefix[idx+1:]
			} else {
				name = withoutPrefix
			}
		case strings.HasPrefix(ref, "refs/heads/"):
			name = strings.TrimPrefix(ref, "refs/heads/")
		default:
			name = ref
		}
		if name == "HEAD" {
			continue
		}
		seenKey := name
		if isRemote {
			seenKey = remote + "/" + name
		}
		if seen[seenKey] {
			continue
		}
		seen[seenKey] = true
		upstream := ""
		if !isRemote && len(parts) == 3 {
			upstream = parts[2]
		}
		branches = append(branches, Branch{
			Name:      name,
			IsRemote:  isRemote,
			IsCurrent: !isRemote && name == current,
			Remote:    remote,
			Upstream:  upstream,
		})
	}
	return branches, nil
}

func (c *Client) ListCommits(ref string, n int) ([]Commit, error) {
	if ref == "" {
		ref = "HEAD"
	}
	format := "%H\t%h\t%s\t%an\t%aI"
	lines, err := c.runLines("log", ref, fmt.Sprintf("-n%d", n), "--pretty=format:"+format)
	if err != nil {
		return nil, err
	}
	headHash, _ := c.run("rev-parse", "HEAD")
	tagMap := c.fetchTagMap()
	var commits []Commit
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		t, _ := time.Parse(time.RFC3339, parts[4])
		commits = append(commits, Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Subject:   parts[2],
			Author:    parts[3],
			Date:      t,
			IsHead:    parts[0] == headHash,
			Tags:      tagMap[parts[0]],
		})
	}
	return commits, nil
}

func (c *Client) ShowCommit(hash string) (CommitDetail, error) {
	raw, err := c.run("show", "--stat", "--format=fuller", hash)
	if err != nil {
		return CommitDetail{}, err
	}
	// split header from stat
	lines := strings.Split(raw, "\n")
	var detail CommitDetail
	detail.Hash = hash

	inStat := false
	var headerLines, statLines []string
	for _, l := range lines {
		if !inStat && strings.Contains(l, "|") {
			inStat = true
		}
		if inStat {
			statLines = append(statLines, l)
		} else {
			headerLines = append(headerLines, l)
		}
	}
	detail.StatLines = statLines

	for _, l := range headerLines {
		switch {
		case strings.HasPrefix(l, "Author:"):
			detail.Author = strings.TrimSpace(strings.TrimPrefix(l, "Author:"))
		case strings.HasPrefix(l, "AuthorDate:"):
			t, _ := time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.TrimSpace(strings.TrimPrefix(l, "AuthorDate:")))
			detail.Date = t
		case strings.HasPrefix(l, "    ") && detail.Subject == "":
			detail.Subject = strings.TrimSpace(l)
		case strings.HasPrefix(l, "    ") && detail.Subject != "":
			detail.Body += strings.TrimSpace(l) + "\n"
		}
	}
	return detail, nil
}

// fetchTagMap returns a map of full commit hash → tag names.
// Annotated tags are dereferenced to their target commit.
func (c *Client) fetchTagMap() map[string][]string {
	lines, _ := c.runLines("tag", "--format=%(refname:short)\t%(objectname)\t%(*objectname)")
	m := map[string][]string{}
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		hash := parts[1]
		if len(parts) == 3 && parts[2] != "" {
			hash = parts[2] // annotated tag: use dereferenced commit hash
		}
		if hash != "" {
			m[hash] = append(m[hash], name)
		}
	}
	return m
}

func (c *Client) Checkout(ref string) error {
	_, err := c.run("checkout", ref)
	return err
}

// CreateBranch creates a new branch and checks it out.
// If from is empty the branch is created from HEAD; otherwise from that ref.
func (c *Client) CreateBranch(name, from string) error {
	if from == "" {
		_, err := c.run("checkout", "-b", name)
		return err
	}
	_, err := c.run("checkout", "-b", name, from)
	return err
}

func (c *Client) CheckoutNewTracking(remoteBranch string) error {
	// e.g. origin/feature/x -> feature/x
	parts := strings.SplitN(remoteBranch, "/", 2)
	local := remoteBranch
	if len(parts) == 2 {
		local = parts[1]
	}
	_, err := c.run("checkout", "-b", local, "--track", remoteBranch)
	return err
}

func (c *Client) Merge(branch string) error {
	_, err := c.run("merge", branch)
	return err
}

func (c *Client) Rebase(onto string) error {
	_, err := c.run("rebase", onto)
	return err
}

func (c *Client) CherryPick(hash string) error {
	_, err := c.run("cherry-pick", hash)
	return err
}

func (c *Client) RevertCommit(hash string) error {
	_, err := c.run("revert", "--no-edit", hash)
	return err
}

// DropCommit removes a specific commit from history via rebase --onto.
func (c *Client) DropCommit(hash string) error {
	_, err := c.run("rebase", "--onto", hash+"^", hash, "HEAD")
	return err
}

func (c *Client) AmendCommit(newMessage string) error {
	_, err := c.run("commit", "--amend", "-m", newMessage)
	return err
}

// UncommitLast undoes the most recent commit via a mixed reset: HEAD moves
// back one commit and the index is unstaged, but the working tree is left
// untouched — so the commit's changes reappear as ordinary uncommitted changes.
func (c *Client) UncommitLast() error {
	_, err := c.run("reset", "HEAD^")
	return err
}

// SquashCommits squashes the listed hashes (must be consecutive, most-recent first)
// into the first (oldest) one via a non-interactive rebase sequence script.
func (c *Client) SquashCommits(hashes []string) error {
	if len(hashes) < 2 {
		return fmt.Errorf("need at least 2 commits to squash")
	}
	// oldest is last in hashes slice (log order = newest first)
	oldest := hashes[len(hashes)-1]
	base, err := c.run("rev-parse", oldest+"^")
	if err != nil {
		return fmt.Errorf("cannot find base for squash: %w", err)
	}

	// Build rebase-todo: pick oldest, fixup the rest
	var todo strings.Builder
	for i := len(hashes) - 1; i >= 0; i-- {
		verb := "fixup"
		if i == len(hashes)-1 {
			verb = "pick"
		}
		fmt.Fprintf(&todo, "%s %s\n", verb, hashes[i][:7])
	}

	// Use GIT_SEQUENCE_EDITOR to inject the todo
	script := fmt.Sprintf("#!/bin/sh\necho %q > \"$1\"\n", todo.String())
	tmpScript := "/tmp/git-manager-squash-editor.sh"
	if err := writeFile(tmpScript, script, 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "rebase", "-i", base)
	cmd.Dir = c.repoRoot
	cmd.Env = append(baseEnv(), "GIT_SEQUENCE_EDITOR="+tmpScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("squash failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (c *Client) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := c.run("branch", flag, name)
	return err
}

func (c *Client) Push(remote, branch string) error {
	_, err := c.run("push", remote, branch)
	return err
}

func (c *Client) ForcePush(remote, branch string) error {
	_, err := c.run("push", "--force-with-lease", remote, branch)
	return err
}

func (c *Client) Fetch(remote string) error {
	_, err := c.run("fetch", remote)
	return err
}

// Pull runs "git pull" on the current branch using its configured upstream.
// The reconciliation strategy is taken from config and passed explicitly, so
// the pull does not depend on the user's global git settings being set.
func (c *Client) Pull() error {
	_, err := c.run("pull", c.cfg.PullStrategy().String())
	return err
}

// FetchBranchFromUpstream updates a local branch from its upstream without
// checking it out (fast-forward only): git fetch <remote> <src>:<dst>.
func (c *Client) FetchBranchFromUpstream(remote, remoteBranch, localBranch string) error {
	_, err := c.run("fetch", remote, remoteBranch+":"+localBranch)
	return err
}
