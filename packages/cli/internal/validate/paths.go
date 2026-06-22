package validate

import (
	"os"
	"path/filepath"
	"strings"
)

// checkPaths verifies that no asset under .ai/ escapes the repository root via
// a symlink or a traversal target (FR; SC-003). Symlinks resolving outside the
// root are errors; symlinks within the root are reported as warnings so the
// operator is aware they exist.
func checkPaths(root, aiDir string, r *Report) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		rootAbs = root
	}
	rootAbs = filepath.Clean(rootAbs)

	_ = filepath.WalkDir(aiDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		rel := relForReport(aiDir, path)
		if d.Type()&os.ModeSymlink == 0 {
			return nil
		}

		target, readErr := os.Readlink(path)
		if readErr != nil {
			r.add(rel, "path", SeverityError, "unreadable symlink: %v", readErr)
			return nil
		}

		// Resolve the link target relative to its own directory.
		resolved := target
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(filepath.Dir(path), target)
		}
		resolved = filepath.Clean(resolved)
		if resolvedAbs, absErr := filepath.Abs(resolved); absErr == nil {
			resolved = resolvedAbs
		}

		if !withinRoot(rootAbs, resolved) {
			r.add(rel, "path", SeverityError, "symlink escapes repository root (-> %s)", target)
			return nil
		}
		r.add(rel, "path", SeverityWarning, "symlink present (-> %s)", target)
		return nil
	})
}

// withinRoot reports whether candidate is root itself or nested under it.
func withinRoot(root, candidate string) bool {
	if candidate == root {
		return true
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
