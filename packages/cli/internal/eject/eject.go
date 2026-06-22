package eject

import (
	"os"
	"path/filepath"
)

var metadataRelPaths = []string{
	filepath.Join(".ai", "lazyai.json"),
	filepath.Join(".ai", "lock.json"),
	filepath.Join(".ai", "migration-report.md"),
	".ai-setup.json",
	".ai-setup.db",
}

type Plan struct {
	Existing []string
}

type Result struct {
	Removed []string
}

func Inspect(root string) Plan {
	existing := make([]string, 0, len(metadataRelPaths))
	for _, rel := range metadataRelPaths {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return Plan{Existing: existing}
}

func Run(root string) (*Result, error) {
	plan := Inspect(root)
	removed := make([]string, 0, len(plan.Existing))
	for _, path := range plan.Existing {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		removed = append(removed, path)
	}
	return &Result{Removed: removed}, nil
}
