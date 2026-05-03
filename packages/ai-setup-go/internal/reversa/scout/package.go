package scout

import (
	"os"
	"path/filepath"
)

// DetectPackageManager detects a package manager from lockfiles and manifests.
func DetectPackageManager(targetDir string) string {
	checks := []struct {
		file string
		name string
	}{
		{file: "pnpm-lock.yaml", name: "pnpm"},
		{file: "yarn.lock", name: "yarn"},
		{file: "package-lock.json", name: "npm"},
		{file: "go.sum", name: "go modules"},
		{file: "Cargo.lock", name: "cargo"},
		{file: "Gemfile.lock", name: "bundler"},
		{file: "poetry.lock", name: "poetry"},
		{file: "Pipfile.lock", name: "pipenv"},
		{file: "composer.lock", name: "composer"},
	}
	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(targetDir, check.file)); err == nil {
			return check.name
		}
	}
	if _, err := os.Stat(filepath.Join(targetDir, "package.json")); err == nil {
		return "npm"
	}
	return ""
}
