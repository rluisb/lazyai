package scaffold

import (
	"strings"
	"testing"
)

// TestFillClaudeMdPlaceholders covers the hybrid auto-infer + fill-in marker
// substitution helper.
func TestFillClaudeMdPlaceholders(t *testing.T) {
	tmpl := `Project: [YOUR_PROJECT_NAME]
Description: [YOUR_PROJECT_DESCRIPTION]
Overview: [YOUR_PROJECT_OVERVIEW]
Stack: [YOUR_TECH_STACK]
Language: [YOUR_LANGUAGE]
Framework: [YOUR_FRAMEWORK]
Database: [YOUR_DATABASE]
ORM: [YOUR_ORM]
Testing: [YOUR_TEST_FRAMEWORK]
Package manager: [YOUR_PACKAGE_MANAGER]
Org: [YOUR_ORG]
Team: [YOUR_TEAM]
Install: [YOUR_INSTALL_COMMAND]
Dev: [YOUR_DEV_COMMAND]
Test: [YOUR_TEST_COMMAND]
Lint: [YOUR_LINT_COMMAND]
Build: [YOUR_BUILD_COMMAND]
Coverage: [YOUR_COVERAGE_THRESHOLD]
Naming: [YOUR_NAMING_CONVENTION]
Names: [YOUR_NAMING_CONVENTIONS]
Errors: [YOUR_ERROR_PATTERN]
API: [YOUR_API_CONVENTION]
Imports: [YOUR_IMPORT_ORDER]
Branch: [YOUR_PROTECTED_BRANCH]
Migrations: [YOUR_MIGRATIONS_PATH]
Strict: [YOUR_STRICT_MODE]
Shared: [YOUR_SHARED_PATH]
Test path: [YOUR_TEST_PATH]
Rule1: [YOUR_RULE_1]
Arch: [YOUR_ARCHITECTURE_NOTES]
Session: [YOUR_SESSION_CHECK]`

	t.Run("with known language and org/team filled", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName:        "demo-app",
			ProjectDescription: "demo description",
			ProjectOverview:    "demo overview",
			PrimaryLanguage:    "Go",
			Framework:          "Cobra",
			Database:           "PostgreSQL",
			ORM:                "sqlc",
			TestFramework:      "Go test",
			PackageManager:     "go modules",
			Organization:       "Acme",
			Team:               "Platform",
			MigrationsPath:     "db/migrate",
			TestPath:           "./...",
			StrictMode:         "TypeScript strict",
			InstallCommand:     "make install",
			ProtectedBranch:    "trunk",
			NamingConventions:  "camelCase values",
			ErrorHandling:      "return errors",
			APIConventions:     "JSON envelopes",
			ImportOrder:        "stdlib, third-party, local",
			CoverageThreshold:  90,
		})

		mustContain(t, out, "Project: demo-app")
		mustContain(t, out, "Description: demo description")
		mustContain(t, out, "Overview: demo overview")
		mustContain(t, out, "Stack: Go · Cobra")
		mustContain(t, out, "Language: Go")
		mustContain(t, out, "Framework: Cobra")
		mustContain(t, out, "Database: PostgreSQL")
		mustContain(t, out, "ORM: sqlc")
		mustContain(t, out, "Testing: Go test")
		mustContain(t, out, "Package manager: go modules")
		mustContain(t, out, "Org: Acme")
		mustContain(t, out, "Team: Platform")
		mustContain(t, out, "Install: make install")
		mustContain(t, out, "Dev: go run .")
		mustContain(t, out, "Test: go test ./...")
		mustContain(t, out, "Lint: go vet ./...")
		mustContain(t, out, "Build: go build ./...")
		mustContain(t, out, "Coverage: 90")
		mustContain(t, out, "Naming: camelCase values")
		mustContain(t, out, "Names: camelCase values")
		mustContain(t, out, "Errors: return errors")
		mustContain(t, out, "API: JSON envelopes")
		mustContain(t, out, "Imports: stdlib, third-party, local")
		mustContain(t, out, "Branch: trunk")
		mustContain(t, out, "Migrations: db/migrate")
		mustContain(t, out, "Strict: TypeScript strict")
		mustContain(t, out, "Shared: <!-- fill-in: shared path -->")
		mustContain(t, out, "Test path: ./...")
		mustContain(t, out, "Rule1: <!-- fill-in: rule 1 -->")
		mustContain(t, out, "Arch: <!-- fill-in: architecture and key patterns -->")
		mustContain(t, out, "Session: <!-- fill-in: team-specific session check -->")
	})

	t.Run("no language and no org/team uses fill-in markers", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName: "demo-app",
		})

		mustContain(t, out, "Project: demo-app")
		mustContain(t, out, "Stack: <!-- fill-in: tech stack -->")
		mustContain(t, out, "Org: <!-- fill-in: your org -->")
		mustContain(t, out, "Team: <!-- fill-in: your team -->")
		mustContain(t, out, "Dev: <!-- fill-in: dev command -->")
		mustContain(t, out, "Description: <!-- fill-in: project description -->")
		mustContain(t, out, "Language: <!-- fill-in: language -->")
		mustContain(t, out, "Install: <!-- fill-in: install command -->")
		mustContain(t, out, "Branch: <!-- fill-in: protected branch -->")
	})

	t.Run("no YOUR_ placeholder remains for known keys", func(t *testing.T) {
		out := fillClaudeMdPlaceholders(tmpl, ScaffoldCompiledRootOptions{
			ProjectName:     "demo",
			PrimaryLanguage: "Go",
		})
		if strings.Contains(out, "[YOUR_") {
			t.Errorf("output still contains raw [YOUR_*] placeholder:\n%s", out)
		}
	})
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("output missing %q\n---\n%s", needle, haystack)
	}
}
