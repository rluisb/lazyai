package agentmemory

import (
	"fmt"
	"os"
	"path/filepath"
)

const dataDirGitignore = "*.sqlite\n*.sqlite-*\n"

// EnsureDataDir creates an agent memory data directory with private permissions.
func EnsureDataDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("chmod data dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(dataDirGitignore), 0o600); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}
