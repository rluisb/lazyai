package detect

import (
	"os"
	"path/filepath"
)

type ProjectMetadata struct {
	PrimaryLanguage string
	Framework       string
	PackageManager  string
	WorkspaceType   string
}

func DetectProjectMetadata(targetDir string) *ProjectMetadata {
	meta := &ProjectMetadata{}

	// 1. package.json -> JS/TS
	if filesExists(targetDir, "package.json") {
		meta.PrimaryLanguage = "TypeScript/JavaScript"
		meta.WorkspaceType = "monorepo" // default if package.json present

		// Attempt to detect package manager from package.json content or lockfiles
		if filesExists(targetDir, "pnpm-lock.yaml") {
			meta.PackageManager = "pnpm"
		} else if filesExists(targetDir, "yarn.lock") {
			meta.PackageManager = "yarn"
		} else if filesExists(targetDir, "package-lock.json") {
			meta.PackageManager = "npm"
		} else {
			meta.PackageManager = "npm" // default
		}
		return meta
	}

	// 2. go.mod -> Go
	if filesExists(targetDir, "go.mod") {
		meta.PrimaryLanguage = "Go"
		meta.WorkspaceType = "single-module"
		return meta
	}

	// 3. Cargo.toml -> Rust
	if filesExists(targetDir, "Cargo.toml") {
		meta.PrimaryLanguage = "Rust"
		return meta
	}

	// 4. pyproject.toml / requirements.txt -> Python
	if filesExists(targetDir, "pyproject.toml") || filesExists(targetDir, "requirements.txt") {
		meta.PrimaryLanguage = "Python"
		return meta
	}

	// 5. pom.xml -> Java/Maven
	if filesExists(targetDir, "pom.xml") {
		meta.PrimaryLanguage = "Java"
		meta.Framework = "Maven"
		return meta
	}

	// 6. Gemfile -> Ruby
	if filesExists(targetDir, "Gemfile") {
		meta.PrimaryLanguage = "Ruby"
		if filesExists(targetDir, "config/routes.rb") {
			meta.Framework = "Rails"
		}
		return meta
	}

	return nil
}

func filesExists(dir, file string) bool {
	_, err := os.Stat(filepath.Join(dir, file))
	return err == nil
}
