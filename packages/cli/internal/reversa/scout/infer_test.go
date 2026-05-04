package scout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferDatabase(t *testing.T) {
	tests := []struct {
		name  string
		hints []DBHint
		want  string
	}{
		{"empty hints", nil, ""},
		{"prisma_schema", []DBHint{{Path: "prisma/schema.prisma", Type: "prisma_schema"}}, "Prisma-managed database"},
		{"migrations_dir", []DBHint{{Path: "migrations", Type: "migrations_dir"}}, "Relational database"},
		{"postgres migrations", []DBHint{{Path: "postgres/migrations", Type: "migrations_dir"}}, "PostgreSQL"},
		{"alembic", []DBHint{{Path: "alembic", Type: "alembic"}}, "SQLAlchemy-managed database"},
		{"rails", []DBHint{{Path: "db/migrate", Type: "rails_migrations"}}, "Rails-managed database"},
		{"sql_schema", []DBHint{{Path: "schema.sql", Type: "sql_schema"}}, "SQL database"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferDatabase(tt.hints); got != tt.want {
				t.Errorf("InferDatabase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferORM(t *testing.T) {
	tests := []struct {
		name  string
		hints []DBHint
		want  string
	}{
		{"empty hints", nil, ""},
		{"prisma_schema", []DBHint{{Path: "prisma/schema.prisma", Type: "prisma_schema"}}, "Prisma"},
		{"alembic", []DBHint{{Path: "alembic", Type: "alembic"}}, "SQLAlchemy"},
		{"migrations only", []DBHint{{Path: "migrations", Type: "migrations_dir"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferORM(tt.hints); got != tt.want {
				t.Errorf("InferORM() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferMigrationsPath(t *testing.T) {
	tests := []struct {
		name  string
		hints []DBHint
		want  string
	}{
		{"empty hints", nil, ""},
		{"migrations_dir", []DBHint{{Path: "db/migrate", Type: "migrations_dir"}}, "db/migrate"},
		{"prisma", []DBHint{{Path: "prisma/schema.prisma", Type: "prisma_schema"}}, "prisma/migrations"},
		{"alembic", []DBHint{{Path: "alembic", Type: "alembic"}}, "alembic"},
		{"rails", []DBHint{{Path: "db/migrate", Type: "rails_migrations"}}, "db/migrate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferMigrationsPath(tt.hints); got != tt.want {
				t.Errorf("InferMigrationsPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferInstallCommand(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{"pnpm", "pnpm install"},
		{"yarn", "yarn install"},
		{"npm", "npm install"},
		{"go modules", "go mod tidy"},
		{"cargo", "cargo build"},
		{"bundler", "bundle install"},
		{"poetry", "poetry install"},
		{"pipenv", "pipenv install"},
		{"composer", "composer install"},
		{"", ""},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			if got := InferInstallCommand(tt.pkg); got != tt.want {
				t.Errorf("InferInstallCommand(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}

func TestInferLintCommand(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		lang string
		want string
	}{
		{"go language", "go modules", "Go", "go vet ./..."},
		{"python language", "poetry", "Python", "ruff check ."},
		{"rust language", "cargo", "Rust", "cargo clippy"},
		{"ruby language", "bundler", "Ruby", "bundle exec rubocop"},
		{"pnpm fallback", "pnpm", "TypeScript", "pnpm lint"},
		{"yarn fallback", "yarn", "JavaScript", "yarn lint"},
		{"npm fallback", "npm", "TypeScript", "npm run lint"},
		{"unknown", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferLintCommand(tt.pkg, tt.lang); got != tt.want {
				t.Errorf("InferLintCommand(%q, %q) = %q, want %q", tt.pkg, tt.lang, got, tt.want)
			}
		})
	}
}

func TestInferTestPath(t *testing.T) {
	tests := []struct {
		framework string
		lang      string
		want      string
	}{
		{"", "Go", "./..."},
		{"", "Python", "tests/"},
		{"", "Ruby", "spec/"},
		{"Playwright", "TypeScript", "e2e/"},
		{"Cypress", "JavaScript", "cypress/"},
		{"Jest", "TypeScript", "__tests__/"},
		{"", "TypeScript", "__tests__/"},
	}
	for _, tt := range tests {
		t.Run(tt.framework+" "+tt.lang, func(t *testing.T) {
			if got := InferTestPath(tt.framework, tt.lang); got != tt.want {
				t.Errorf("InferTestPath(%q, %q) = %q, want %q", tt.framework, tt.lang, got, tt.want)
			}
		})
	}
}

func TestInferTestCommand(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		lang string
		want string
	}{
		{"go language", "go modules", "Go", "go test ./..."},
		{"python language", "poetry", "Python", "pytest"},
		{"ruby language", "bundler", "Ruby", "bundle exec rspec"},
		{"rust language", "cargo", "Rust", "cargo test"},
		{"pnpm fallback", "pnpm", "TypeScript", "pnpm test"},
		{"yarn fallback", "yarn", "JavaScript", "yarn test"},
		{"npm fallback", "npm", "TypeScript", "npm test"},
		{"unknown", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferTestCommand(tt.pkg, tt.lang); got != tt.want {
				t.Errorf("InferTestCommand(%q, %q) = %q, want %q", tt.pkg, tt.lang, got, tt.want)
			}
		})
	}
}

func TestBuildCodebaseMapEntries(t *testing.T) {
	t.Run("with entry points and modules", func(t *testing.T) {
		modules := []string{"cmd", "internal", "pkg"}
		entryPoints := []EntryPoint{
			{Path: "cmd/ai-setup/main.go", Type: "go_cmd"},
			{Path: ".", Type: "project_root"},
		}
		entries := BuildCodebaseMapEntries(modules, entryPoints)

		// cmd should be covered by the entry point, so not duplicated.
		foundCmd := false
		for _, e := range entries {
			if e.Path == "cmd" {
				foundCmd = true
			}
		}
		if foundCmd {
			t.Error("expected cmd module to be covered by entry point, but it was added as a module")
		}

		// Should have entry point + internal + pkg + shared marker.
		if len(entries) < 3 {
			t.Errorf("expected at least 3 entries, got %d", len(entries))
		}

		// Entry point should have a responsibility.
		ep := entries[0]
		if ep.Path != "cmd/ai-setup/main.go" {
			t.Errorf("first entry path = %q, want cmd/ai-setup/main.go", ep.Path)
		}
		if ep.Responsibility == "<!-- fill-in: responsibility -->" {
			t.Error("expected auto-filled responsibility for go_cmd entry point")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		entries := BuildCodebaseMapEntries(nil, nil)
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for empty input, got %d", len(entries))
		}
	})

	t.Run("entry point responsibility mapping", func(t *testing.T) {
		tests := []struct {
			epType string
			want   string
		}{
			{"package_main", "Application entry point"},
			{"app_entry", "Application entry point"},
			{"server_entry", "HTTP server entry point"},
			{"go_cmd", "CLI command entry point"},
			{"nextjs_layout", "Next.js layout"},
			{"django_manage", "Django management command"},
			{"wsgi_entry", "WSGI application entry"},
			{"unknown_type", "<!-- fill-in: responsibility -->"},
		}
		for _, tt := range tests {
			t.Run(tt.epType, func(t *testing.T) {
				ep := EntryPoint{Path: "test", Type: tt.epType}
				got := entryPointResponsibility(ep)
				if got != tt.want {
					t.Errorf("entryPointResponsibility(%q) = %q, want %q", tt.epType, got, tt.want)
				}
			})
		}
	})
}

func TestInferDatabaseFromMigrationsPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"migrations", "Relational database"},
		{"postgres/migrations", "PostgreSQL"},
		{"pg/migrations", "PostgreSQL"},
		{"mysql/migrations", "MySQL"},
		{"sqlite/migrations", "SQLite"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := inferDatabaseFromMigrationsPath(tt.path); got != tt.want {
				t.Errorf("inferDatabaseFromMigrationsPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestInferInstallCommandFromSurface(t *testing.T) {
	t.Run("nil surface", func(t *testing.T) {
		if got := InferInstallCommandFromSurface(nil); got != "" {
			t.Errorf("expected empty string for nil surface, got %q", got)
		}
	})
	t.Run("pnpm surface", func(t *testing.T) {
		s := &SurfaceData{PackageManager: "pnpm"}
		if got := InferInstallCommandFromSurface(s); got != "pnpm install" {
			t.Errorf("expected 'pnpm install', got %q", got)
		}
	})
}

func TestInferLintCommandFromSurface(t *testing.T) {
	t.Run("nil surface", func(t *testing.T) {
		if got := InferLintCommandFromSurface(nil); got != "" {
			t.Errorf("expected empty string for nil surface, got %q", got)
		}
	})
	t.Run("go surface", func(t *testing.T) {
		s := &SurfaceData{PackageManager: "go modules", PrimaryLanguage: "Go"}
		if got := InferLintCommandFromSurface(s); got != "go vet ./..." {
			t.Errorf("expected 'go vet ./...', got %q", got)
		}
	})
}

func TestInferTestCommandFromSurface(t *testing.T) {
	t.Run("nil surface", func(t *testing.T) {
		if got := InferTestCommandFromSurface(nil); got != "" {
			t.Errorf("expected empty string for nil surface, got %q", got)
		}
	})
	t.Run("yarn TypeScript surface", func(t *testing.T) {
		s := &SurfaceData{PackageManager: "yarn", PrimaryLanguage: "TypeScript"}
		if got := InferTestCommandFromSurface(s); got != "yarn test" {
			t.Errorf("expected 'yarn test', got %q", got)
		}
	})
}

func TestInferTestPathFromSurface(t *testing.T) {
	t.Run("nil surface", func(t *testing.T) {
		if got := InferTestPathFromSurface(nil); got != "" {
			t.Errorf("expected empty string for nil surface, got %q", got)
		}
	})
	t.Run("go surface", func(t *testing.T) {
		s := &SurfaceData{TestFramework: "Go test", PrimaryLanguage: "Go"}
		if got := InferTestPathFromSurface(s); got != "./..." {
			t.Errorf("expected './...', got %q", got)
		}
	})
}

func TestInferStrictMode(t *testing.T) {
	t.Run("nonexistent tsconfig", func(t *testing.T) {
		dir := t.TempDir()
		if got := InferStrictMode(dir); got != "" {
			t.Errorf("expected empty string for nonexistent tsconfig, got %q", got)
		}
	})

	t.Run("tsconfig with strict true", func(t *testing.T) {
		dir := t.TempDir()
		content := `{"compilerOptions": {"strict": true}}`
		if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		if got := InferStrictMode(dir); got != "TypeScript strict" {
			t.Errorf("expected 'TypeScript strict', got %q", got)
		}
	})
}
