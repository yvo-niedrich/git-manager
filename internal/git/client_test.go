package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yvo.niedrich/git-manager/internal/config"
	"github.com/yvo.niedrich/git-manager/internal/git"
	"github.com/yvo.niedrich/git-manager/internal/testutil"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func fixture(t *testing.T) *git.Client {
	t.Helper()
	return git.NewClient(testutil.FixtureRepo(t), config.NewStatic())
}

func findBranch(branches []git.Branch, name string, isRemote bool) *git.Branch {
	for i := range branches {
		if branches[i].Name == name && branches[i].IsRemote == isRemote {
			return &branches[i]
		}
	}
	return nil
}

// ── CurrentBranch ─────────────────────────────────────────────────────────────

func TestCurrentBranch(t *testing.T) {
	if got := fixture(t).CurrentBranch(); got != "main" {
		t.Fatalf("got %q, want %q", got, "main")
	}
}

// ── ListBranches ──────────────────────────────────────────────────────────────

func TestListBranches_LocalMain(t *testing.T) {
	branches, err := fixture(t).ListBranches()
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	b := findBranch(branches, "main", false)
	if b == nil {
		t.Fatal("local branch 'main' not found")
	}
	if !b.IsCurrent {
		t.Error("IsCurrent: got false, want true")
	}
	if b.Upstream != "origin/main" {
		t.Errorf("Upstream: got %q, want %q", b.Upstream, "origin/main")
	}
}

func TestListBranches_FeatureBranch(t *testing.T) {
	branches, err := fixture(t).ListBranches()
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	b := findBranch(branches, "feature/my-feature", false)
	if b == nil {
		t.Fatal("local branch 'feature/my-feature' not found")
	}
	if b.IsCurrent {
		t.Error("IsCurrent: got true, want false")
	}
	if b.Upstream != "" {
		t.Errorf("Upstream: got %q, want empty", b.Upstream)
	}
}

func TestListBranches_RemoteTracking(t *testing.T) {
	branches, err := fixture(t).ListBranches()
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	b := findBranch(branches, "main", true)
	if b == nil {
		t.Fatal("remote branch 'origin/main' not found")
	}
	if b.Remote != "origin" {
		t.Errorf("Remote: got %q, want %q", b.Remote, "origin")
	}
}

func TestListBranches_NoHEADEntry(t *testing.T) {
	branches, err := fixture(t).ListBranches()
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	for _, b := range branches {
		if b.Name == "HEAD" {
			t.Error("HEAD should be filtered out of branch list")
		}
	}
}

// ── ListCommits ───────────────────────────────────────────────────────────────

func TestListCommits_Count(t *testing.T) {
	commits, err := fixture(t).ListCommits("main", 10)
	if err != nil {
		t.Fatalf("ListCommits: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("got %d commits, want 3", len(commits))
	}
}

func TestListCommits_Order(t *testing.T) {
	commits, err := fixture(t).ListCommits("main", 10)
	if err != nil {
		t.Fatalf("ListCommits: %v", err)
	}
	want := []string{"third commit", "second commit", "initial commit"}
	for i, w := range want {
		if commits[i].Subject != w {
			t.Errorf("commits[%d].Subject = %q, want %q", i, commits[i].Subject, w)
		}
	}
}

func TestListCommits_HeadFlag(t *testing.T) {
	commits, err := fixture(t).ListCommits("main", 10)
	if err != nil {
		t.Fatalf("ListCommits: %v", err)
	}
	if !commits[0].IsHead {
		t.Error("commits[0].IsHead: got false, want true")
	}
	for _, c := range commits[1:] {
		if c.IsHead {
			t.Errorf("commit %q: IsHead should be false", c.Subject)
		}
	}
}

func TestListCommits_Tag(t *testing.T) {
	commits, err := fixture(t).ListCommits("main", 10)
	if err != nil {
		t.Fatalf("ListCommits: %v", err)
	}
	// v0.1.0 was tagged on "second commit"
	tagged := commits[1]
	if len(tagged.Tags) == 0 {
		t.Fatal("expected tag on second commit, got none")
	}
	if tagged.Tags[0] != "v0.1.0" {
		t.Errorf("Tags[0] = %q, want %q", tagged.Tags[0], "v0.1.0")
	}
}

// ── ShowCommit ────────────────────────────────────────────────────────────────

func TestShowCommit_Subject(t *testing.T) {
	c := fixture(t)
	commits, err := c.ListCommits("main", 1)
	if err != nil || len(commits) == 0 {
		t.Fatal("could not get HEAD commit")
	}
	detail, err := c.ShowCommit(commits[0].Hash)
	if err != nil {
		t.Fatalf("ShowCommit: %v", err)
	}
	if detail.Subject != "third commit" {
		t.Errorf("Subject = %q, want %q", detail.Subject, "third commit")
	}
	if detail.Author == "" {
		t.Error("Author should not be empty")
	}
	if len(detail.StatLines) == 0 {
		t.Error("StatLines should not be empty for a commit that modified a file")
	}
}

func TestShowCommit_Body(t *testing.T) {
	c := fixture(t)
	commits, err := c.ListCommits("main", 10)
	if err != nil || len(commits) < 2 {
		t.Fatal("could not list commits")
	}
	// "second commit" has an explicit body
	detail, err := c.ShowCommit(commits[1].Hash)
	if err != nil {
		t.Fatalf("ShowCommit: %v", err)
	}
	if detail.Body == "" {
		t.Error("Body should not be empty for second commit")
	}
}

// ── Status ────────────────────────────────────────────────────────────────────

func TestStatus_TrackedAndUntracked(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())

	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello\nworld\nmodified\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	status, err := c.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Tracked) != 1 || status.Tracked[0] != "hello.txt" {
		t.Errorf("Tracked = %v, want [hello.txt]", status.Tracked)
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "new.txt" {
		t.Errorf("Untracked = %v, want [new.txt]", status.Untracked)
	}
}

func TestStatus_OnlyUntracked(t *testing.T) {
	root := testutil.MutableRepo(t)
	c := git.NewClient(root, config.NewStatic())

	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	status, err := c.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Tracked) != 0 {
		t.Errorf("Tracked = %v, want none", status.Tracked)
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "new.txt" {
		t.Errorf("Untracked = %v, want [new.txt]", status.Untracked)
	}
}

func TestStatus_Clean(t *testing.T) {
	status, err := fixture(t).Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Tracked) != 0 || len(status.Untracked) != 0 {
		t.Errorf("Status = %+v, want empty on a clean repo", status)
	}
}
