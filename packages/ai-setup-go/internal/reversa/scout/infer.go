package scout

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/compiler"
)

// InferDatabase returns a human-readable database name from DatabaseHints.
// It uses migration directories and Prisma schemas to infer the RDBMS.
func InferDatabase(hints []DBHint) string {
	for _, h := range hints {
		switch h.Type {
		case "prisma_schema":
			return "Prisma-managed database"
		case "migrations_dir":
			return inferDatabaseFromMigrationsPath(h.Path)
		case "alembic":
			return "SQLAlchemy-managed database"
		case "rails_migrations":
			return "Rails-managed database"
		case "sql_schema":
			return "SQL database"
		}
	}
	return ""
}

// InferORM returns a human-readable ORM name from DatabaseHints.
func InferORM(hints []DBHint) string {
	for _, h := range hints {
		switch h.Type {
		case "prisma_schema":
			return "Prisma"
		case "alembic":
			return "SQLAlchemy"
		}
	}
	return ""
}

// InferMigrationsPath returns the best migration directory path from hints,
// or empty string if none found.
func InferMigrationsPath(hints []DBHint) string {
	for _, h := range hints {
		switch h.Type {
		case "migrations_dir", "rails_migrations":
			return h.Path
		case "prisma_schema":
			// Prisma migrations are under prisma/migrations, but the schema
			// file itself is the authoritative source.
			return "prisma/migrations"
		case "alembic":
			return h.Path
		}
	}
	return ""
}

// InferProtectedBranch detects the default branch name from .git/HEAD.
// Returns "main" or "master" if detectable, empty string otherwise.
func InferProtectedBranch(targetDir string) string {
	gitHead := filepath.Join(targetDir, ".git", "HEAD")
	data, err := os.ReadFile(gitHead)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	// .git/HEAD format: "ref: refs/heads/main" or a detached commit hash
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	return ""
}

// InferInstallCommand returns the install command for a given package manager.
func InferInstallCommand(pkgManager string) string {
	switch strings.ToLower(pkgManager) {
	case "pnpm":
		return "pnpm install"
	case "yarn":
		return "yarn install"
	case "npm":
		return "npm install"
	case "go modules":
		return "go mod tidy"
	case "cargo":
		return "cargo build"
	case "bundler":
		return "bundle install"
	case "poetry":
		return "poetry install"
	case "pipenv":
		return "pipenv install"
	case "composer":
		return "composer install"
	default:
		return ""
	}
}

// InferLintCommand returns the lint command based on package manager and language.
func InferLintCommand(pkgManager, lang string) string {
	// Try language-based inference first.
	switch strings.ToLower(lang) {
	case "go":
		return "go vet ./..."
	case "python":
		return "ruff check ."
	case "rust":
		return "cargo clippy"
	case "ruby":
		return "bundle exec rubocop"
	}
	// Fall back to package-manager-based inference.
	switch strings.ToLower(pkgManager) {
	case "pnpm":
		return "pnpm lint"
	case "yarn":
		return "yarn lint"
	case "npm":
		return "npm run lint"
	}
	return ""
}

// InferTestCommand returns the conventional test command based on package
// manager and primary language.
func InferTestCommand(pkgManager, lang string) string {
	switch strings.ToLower(lang) {
	case "go":
		return "go test ./..."
	case "python":
		return "pytest"
	case "rust":
		return "cargo test"
	case "ruby":
		return "bundle exec rspec"
	case "java":
		return "./gradlew test"
	}
	switch strings.ToLower(pkgManager) {
	case "pnpm":
		return "pnpm test"
	case "yarn":
		return "yarn test"
	case "npm":
		return "npm test"
	}
	return ""
}

// InferTestPath returns a conventional test directory path based on
// the test framework and primary language.
func InferTestPath(testFramework, lang string) string {
	switch strings.ToLower(lang) {
	case "go":
		return "./..."
	case "python":
		return "tests/"
	case "ruby":
		return "spec/"
	default:
		// JS/TS conventions
		switch {
		case strings.EqualFold(testFramework, "playwright"):
			return "e2e/"
		case strings.EqualFold(testFramework, "cypress"):
			return "cypress/"
		default:
			return "__tests__/"
		}
	}
}

// InferStrictMode detects TypeScript strict mode.
func InferStrictMode(targetDir string) string {
	tsconfig := filepath.Join(targetDir, "tsconfig.json")
	data, err := os.ReadFile(tsconfig)
	if err != nil {
		return ""
	}
	content := strings.ToLower(string(data))
	// Check for "strict": true (with or without quotes around true)
	if strings.Contains(content, `"strict": true`) ||
		strings.Contains(content, `"strict":true`) {
		return "TypeScript strict"
	}
	return ""
}

// InferInstallCommandFromSurface derives the install command from SurfaceData.
func InferInstallCommandFromSurface(surface *SurfaceData) string {
	if surface == nil || surface.PackageManager == "" {
		return ""
	}
	return InferInstallCommand(surface.PackageManager)
}

// InferLintCommandFromSurface derives the lint command from SurfaceData.
func InferLintCommandFromSurface(surface *SurfaceData) string {
	if surface == nil {
		return ""
	}
	return InferLintCommand(surface.PackageManager, surface.PrimaryLanguage)
}

// InferTestCommandFromSurface derives the test command from SurfaceData.
func InferTestCommandFromSurface(surface *SurfaceData) string {
	if surface == nil {
		return ""
	}
	return InferTestCommand(surface.PackageManager, surface.PrimaryLanguage)
}

// InferTestPathFromSurface derives the test path from SurfaceData.
func InferTestPathFromSurface(surface *SurfaceData) string {
	if surface == nil {
		return ""
	}
	return InferTestPath(surface.TestFramework, surface.PrimaryLanguage)
}

// BuildCodebaseMapEntries converts Scout-detected modules and entry points
// into CodebaseMapEntry slices suitable for the AGENTS.md codebase map.
func BuildCodebaseMapEntries(modules []string, entryPoints []EntryPoint) []compiler.CodebaseMapEntry {
	entries := make([]compiler.CodebaseMapEntry, 0, len(modules)+len(entryPoints)+2)

	// Add entry points first.
	for _, ep := range entryPoints {
		if ep.Type == "project_root" {
			continue
		}
		responsibility := entryPointResponsibility(ep)
		entries = append(entries, compiler.CodebaseMapEntry{
			Path:           ep.Path,
			Responsibility: responsibility,
		})
	}

	// Add modules as directories.
	for _, mod := range modules {
		// Skip if already covered by an entry point path.
		if isModuleCoveredByEntry(mod, entryPoints) {
			continue
		}
		entries = append(entries, compiler.CodebaseMapEntry{
			Path:           mod,
			Responsibility: "<!-- fill-in: responsibility -->",
		})
	}

	// Append shared path and infra path markers.
	if len(entries) > 0 {
		entries = append(entries,
			compiler.CodebaseMapEntry{
				Path:           "<!-- fill-in: shared path -->",
				Responsibility: "Shared utilities — check all importers before editing",
			},
		)
	}

	return entries
}

func entryPointResponsibility(ep EntryPoint) string {
	switch ep.Type {
	case "package_main":
		return "Application entry point"
	case "app_entry":
		return "Application entry point"
	case "server_entry":
		return "HTTP server entry point"
	case "go_cmd":
		return "CLI command entry point"
	case "nextjs_layout":
		return "Next.js layout"
	case "nextjs_page":
		return "Next.js page"
	case "nextjs_pages":
		return "Next.js pages router"
	case "nextjs_app":
		return "Next.js app router"
	case "django_manage":
		return "Django management command"
	case "wsgi_entry":
		return "WSGI application entry"
	default:
		return "<!-- fill-in: responsibility -->"
	}
}

func isModuleCoveredByEntry(mod string, entryPoints []EntryPoint) bool {
	for _, ep := range entryPoints {
		// If the entry point is inside this module, don't add both.
		prefix := mod + "/"
		if strings.HasPrefix(ep.Path, prefix) {
			return true
		}
		if ep.Path == mod {
			return true
		}
	}
	return false
}

func inferDatabaseFromMigrationsPath(path string) string {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "postgres") || strings.Contains(lower, "pg") {
		return "PostgreSQL"
	}
	if strings.Contains(lower, "mysql") {
		return "MySQL"
	}
	if strings.Contains(lower, "sqlite") {
		return "SQLite"
	}
	// Default: most migration dirs are for relational databases.
	return "Relational database"
}
