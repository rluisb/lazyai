package db

import (
	"strings"
	"testing"
)

func TestMigration007AddsConstitutionConfigColumns(t *testing.T) {
	database := openTestDB(t)
	applyMigrationsThrough(t, database, 6)

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	expectedColumns := map[string]string{
		"projectOverview":   "TEXT",
		"namingConventions": "TEXT",
		"errorHandling":     "TEXT",
		"apiConventions":    "TEXT",
		"importOrder":       "TEXT",
		"protectedBranch":   "TEXT",
		"testCommand":       "TEXT",
		"lintCommand":       "TEXT",
		"buildCommand":      "TEXT",
		"coverageThreshold": "INTEGER",
	}

	columns := configColumns(t, database)
	for name, wantType := range expectedColumns {
		gotType, ok := columns[name]
		if !ok {
			t.Fatalf("config column %q missing after migration 007", name)
		}
		if gotType != wantType {
			t.Fatalf("config column %q type = %q, want %q", name, gotType, wantType)
		}
	}

	_, err := database.Exec(`
		INSERT INTO config (
			id, scope, tools, cli_tools, enable_servers, project_name, workspace_name,
			target_dir, planning_dir, planning_repo_path, repos, global_ref,
			projectOverview, namingConventions, errorHandling, apiConventions,
			importOrder, protectedBranch, testCommand, lintCommand, buildCommand,
			coverageThreshold
		) VALUES (
			1, 'project', '[]', '[]', '[]', 'migration-test', '',
			'/tmp/migration-test', 'specs', '', '[]', '',
			'Overview', 'Use camelCase', 'Return errors', 'JSON responses',
			'Standard imports', 'main', 'go test ./...', 'go vet ./...', 'go build ./...',
			95
		)`)
	if err != nil {
		t.Fatalf("insert config with migration 007 columns: %v", err)
	}

	var projectOverview, namingConventions, errorHandling, apiConventions string
	var importOrder, protectedBranch, testCommand, lintCommand, buildCommand string
	var coverageThreshold int
	err = database.QueryRow(`
		SELECT projectOverview, namingConventions, errorHandling, apiConventions,
		       importOrder, protectedBranch, testCommand, lintCommand, buildCommand,
		       coverageThreshold
		FROM config WHERE id = 1
	`).Scan(
		&projectOverview, &namingConventions, &errorHandling, &apiConventions,
		&importOrder, &protectedBranch, &testCommand, &lintCommand, &buildCommand,
		&coverageThreshold,
	)
	if err != nil {
		t.Fatalf("query config with migration 007 columns: %v", err)
	}

	if projectOverview != "Overview" || namingConventions != "Use camelCase" || errorHandling != "Return errors" || apiConventions != "JSON responses" {
		t.Fatalf("profile columns were not preserved: %q %q %q %q", projectOverview, namingConventions, errorHandling, apiConventions)
	}
	if importOrder != "Standard imports" || protectedBranch != "main" || testCommand != "go test ./..." || lintCommand != "go vet ./..." || buildCommand != "go build ./..." {
		t.Fatalf("command/config columns were not preserved: %q %q %q %q %q", importOrder, protectedBranch, testCommand, lintCommand, buildCommand)
	}
	if coverageThreshold != 95 {
		t.Fatalf("coverageThreshold = %d, want 95", coverageThreshold)
	}
}

func TestMigration007CoverageThresholdConstraint(t *testing.T) {
	database := openTestDB(t)
	applyMigrationsThrough(t, database, 6)

	if err := RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	_, err := database.Exec(`
		INSERT INTO config (id, scope, tools, cli_tools, enable_servers, project_name, target_dir, repos, coverageThreshold)
		VALUES (1, 'project', '[]', '[]', '[]', 'migration-test', '/tmp/migration-test', '[]', 0)
	`)
	if err == nil {
		t.Fatal("coverageThreshold=0 insert succeeded, want CHECK constraint failure")
	}
	if !strings.Contains(err.Error(), "CHECK constraint failed") {
		t.Fatalf("coverageThreshold=0 error = %v, want CHECK constraint failure", err)
	}
}

func applyMigrationsThrough(t *testing.T, database *DB, maxVersion uint) {
	t.Helper()

	if _, err := database.Exec(createMigrationsTableSQL); err != nil {
		t.Fatalf("create schema_migrations table: %v", err)
	}

	for _, migration := range migrations {
		if migration.Version > maxVersion {
			continue
		}
		for _, stmt := range splitStatements(migration.Up) {
			if _, err := database.Exec(stmt); err != nil {
				t.Fatalf("apply migration %d statement %q: %v", migration.Version, truncate(stmt, 80), err)
			}
		}
		if _, err := database.Exec(setVersionSQL, migration.Version); err != nil {
			t.Fatalf("record migration version %d: %v", migration.Version, err)
		}
	}
}

func configColumns(t *testing.T, database *DB) map[string]string {
	t.Helper()

	rows, err := database.Query("PRAGMA table_info(config)")
	if err != nil {
		t.Fatalf("PRAGMA table_info(config): %v", err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan config column: %v", err)
		}
		columns[name] = columnType
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate config columns: %v", err)
	}

	return columns
}
