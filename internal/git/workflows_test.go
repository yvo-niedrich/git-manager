package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yvo.niedrich/git-manager/internal/config"
	"github.com/yvo.niedrich/git-manager/internal/git"
	"github.com/yvo.niedrich/git-manager/internal/testutil"
)

func TestWorkflowsCommit_StagesTrackedChangesOnly(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())
	wf := git.NewWorkflows(c)

	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello\nworld\nmodified\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	res := wf.Commit("commit tracked only")
	if !res.OK() {
		t.Fatalf("Commit: %v", res.Err)
	}

	status, err := c.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Tracked) != 0 {
		t.Errorf("tracked changes remain after commit: %v", status.Tracked)
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "new.txt" {
		t.Errorf("Untracked = %v, want [new.txt] left alone", status.Untracked)
	}
}

func TestWorkflowsCommit_StagesUntrackedWhenNoTrackedChanges(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())
	wf := git.NewWorkflows(c)

	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	res := wf.Commit("commit untracked")
	if !res.OK() {
		t.Fatalf("Commit: %v", res.Err)
	}
	if c.HasUncommittedChanges() {
		t.Error("expected clean working tree after commit")
	}

	commits, err := c.ListCommits("HEAD", 1)
	if err != nil || len(commits) == 0 {
		t.Fatalf("ListCommits: %v", err)
	}
	if commits[0].Subject != "commit untracked" {
		t.Errorf("HEAD subject = %q, want %q", commits[0].Subject, "commit untracked")
	}
}

func TestWorkflowsCommit_NothingToCommit(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())
	wf := git.NewWorkflows(c)

	res := wf.Commit("should fail")
	if res.OK() {
		t.Fatal("expected error when nothing to commit")
	}
}

func TestWorkflowsUncommitLast_RestoresChangesUnstaged(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())
	wf := git.NewWorkflows(c)

	before, err := c.ListCommits("HEAD", 2)
	if err != nil || len(before) < 2 {
		t.Fatalf("ListCommits: %v", err)
	}
	previousHash := before[1].Hash

	res := wf.UncommitLast()
	if !res.OK() {
		t.Fatalf("UncommitLast: %v", res.Err)
	}

	if got := c.CurrentBranch(); got != "main" {
		t.Fatalf("CurrentBranch = %q, want main", got)
	}
	head, err := c.ListCommits("HEAD", 1)
	if err != nil || len(head) == 0 {
		t.Fatalf("ListCommits: %v", err)
	}
	if head[0].Hash != previousHash {
		t.Errorf("HEAD = %s, want %s (previous commit)", head[0].Hash, previousHash)
	}

	status, err := c.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Tracked) == 0 {
		t.Error("expected the uncommitted commit's changes to reappear as tracked changes")
	}
}

func TestWorkflowsUncommitLast_FailsOnRootCommit(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())
	wf := git.NewWorkflows(c)

	commits, err := c.ListCommits("HEAD", 1000)
	if err != nil || len(commits) == 0 {
		t.Fatalf("ListCommits: %v", err)
	}
	rootHash := commits[len(commits)-1].Hash
	if err := exec.Command("git", "-C", root, "checkout", rootHash).Run(); err != nil {
		t.Fatalf("checkout root commit: %v", err)
	}

	res := wf.UncommitLast()
	if res.OK() {
		t.Fatal("expected error when uncommitting the root commit")
	}
}
