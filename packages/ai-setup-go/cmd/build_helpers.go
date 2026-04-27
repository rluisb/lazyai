package cmd

import (
	"fmt"
	"os"
)

// preflightOutDir validates that outDir is safe to write into. If it does
// not exist, nothing to do. If it exists and is empty, nothing to do. If
// it exists and is non-empty, require force (and wipe it). Shared between
// `build-plugin` (spec 016) and `build-gemini-extension` (spec 017).
func preflightOutDir(outDir string, force bool) error {
	info, err := os.Stat(outDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat %s: %w", outDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("--out %s exists and is not a directory", outDir)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("read %s: %w", outDir, err)
	}
	if len(entries) == 0 {
		return nil
	}
	if !force {
		return fmt.Errorf("--out %s is not empty; pass --force to overwrite", outDir)
	}
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("remove %s: %w", outDir, err)
	}
	return nil
}
