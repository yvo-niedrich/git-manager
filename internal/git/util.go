package git

import (
	"os"
	"path/filepath"
)

func writeFile(path, content string, mode os.FileMode) error {
	return os.WriteFile(path, []byte(content), mode)
}

func baseEnv() []string {
	return os.Environ()
}

// FindRepoRoot walks up from dir until it finds a .git directory.
func FindRepoRoot(dir string) (string, error) {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", os.ErrNotExist
		}
		current = parent
	}
}
