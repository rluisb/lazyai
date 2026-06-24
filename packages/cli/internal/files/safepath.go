package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafeJoin joins name under root and rejects paths that escape root via
// absolute paths, rooted paths, or ".." traversal. It rejects both Unix-style
// and Windows-style rooted names regardless of the host OS.
func SafeJoin(root, name string) (string, error) {
	if filepath.IsAbs(name) || strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", name)
	}
	dest := filepath.Join(root, name)
	rel, err := filepath.Rel(root, dest)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root: %s", name)
	}
	return dest, nil
}
