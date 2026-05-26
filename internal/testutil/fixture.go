package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	once    sync.Once
	repoDir string
	repoErr error
)

// FixtureRepo runs testdata/fixtures/setup.sh once per test binary and returns
// the path to the resulting read-only fixture repository.
func FixtureRepo(tb testing.TB) string {
	tb.Helper()
	once.Do(func() {
		root, err := moduleRoot()
		if err != nil {
			repoErr = err
			return
		}
		script := filepath.Join(root, "testdata", "fixtures", "setup.sh")
		cmd := exec.Command("bash", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			repoErr = fmt.Errorf("setup.sh: %w", err)
			return
		}
		repoDir = filepath.Join(root, "testdata", "fixtures", "repo")
	})
	if repoErr != nil {
		tb.Fatalf("fixture: %v", repoErr)
	}
	return repoDir
}

// MutableRepo copies the fixture into a fresh t.TempDir so the test can mutate
// git state without affecting the shared read-only fixture.
func MutableRepo(tb testing.TB) string {
	tb.Helper()
	src := FixtureRepo(tb)
	dst := tb.TempDir()
	if err := exec.Command("cp", "-r", src+"/.", dst).Run(); err != nil {
		tb.Fatalf("MutableRepo: copy fixture: %v", err)
	}
	return dst
}

// moduleRoot walks up from the test's working directory until it finds go.mod.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
