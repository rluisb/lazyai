package scout

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type frameworkMatch struct {
	key  string
	name string
}

// DetectFrameworks detects frameworks and major libraries from known manifest files.
func DetectFrameworks(targetDir string) []FrameworkEntry {
	var entries []FrameworkEntry
	seen := map[string]struct{}{}
	add := func(name, version, source string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		entries = append(entries, FrameworkEntry{Name: name, Version: version, Source: source})
	}

	detectPackageJSONFrameworks(targetDir, add)
	detectGoModFrameworks(targetDir, add)
	detectPythonFrameworks(targetDir, add)
	detectMarkerFrameworks(targetDir, add)

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Source == entries[j].Source {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Source < entries[j].Source
	})
	return entries
}

func detectPackageJSONFrameworks(targetDir string, add func(name, version, source string)) {
	path := filepath.Join(targetDir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return
	}
	matches := []frameworkMatch{
		{key: "next", name: "Next.js"},
		{key: "react", name: "React"},
		{key: "vue", name: "Vue.js"},
		{key: "@nestjs/core", name: "NestJS"},
		{key: "express", name: "Express"},
		{key: "fastify", name: "Fastify"},
		{key: "prisma", name: "Prisma"},
		{key: "@prisma/client", name: "Prisma"},
		{key: "typeorm", name: "TypeORM"},
		{key: "drizzle-orm", name: "Drizzle"},
		{key: "knex", name: "Knex"},
		{key: "jest", name: "Jest"},
		{key: "vitest", name: "Vitest"},
		{key: "mocha", name: "Mocha"},
		{key: "playwright", name: "Playwright"},
		{key: "cypress", name: "Cypress"},
		{key: "tailwindcss", name: "Tailwind CSS"},
		{key: "sass", name: "Sass"},
		{key: "@angular/core", name: "Angular"},
		{key: "svelte", name: "Svelte"},
		{key: "@remix-run/react", name: "Remix"},
		{key: "nuxt", name: "Nuxt"},
		{key: "gatsby", name: "Gatsby"},
		{key: "electron", name: "Electron"},
		{key: "graphql", name: "GraphQL"},
		{key: "@apollo/client", name: "Apollo"},
		{key: "redux", name: "Redux"},
		{key: "zustand", name: "Zustand"},
	}
	for _, match := range matches {
		if version, ok := manifest.Dependencies[match.key]; ok {
			add(match.name, version, "package.json:dependencies."+match.key)
			continue
		}
		if version, ok := manifest.DevDependencies[match.key]; ok {
			add(match.name, version, "package.json:devDependencies."+match.key)
		}
	}
}

func detectGoModFrameworks(targetDir string, add func(name, version, source string)) {
	path := filepath.Join(targetDir, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(data)
	matches := []frameworkMatch{
		{key: "gin-gonic/gin", name: "Gin"},
		{key: "labstack/echo", name: "Echo"},
		{key: "gofiber/fiber", name: "Fiber"},
		{key: "go-chi/chi", name: "Chi"},
		{key: "gorilla/mux", name: "Gorilla Mux"},
		{key: "gorm.io/gorm", name: "GORM"},
		{key: "jmoiron/sqlx", name: "SQLx"},
		{key: "entgo.io/ent", name: "Ent"},
		{key: "kyleconroy/sqlc", name: "sqlc"},
	}
	for _, match := range matches {
		if !strings.Contains(content, match.key) {
			continue
		}
		add(match.name, versionFromGoMod(content, match.key), "go.mod:"+match.key)
	}
}

func detectPythonFrameworks(targetDir string, add func(name, version, source string)) {
	matches := []frameworkMatch{
		{key: "django", name: "Django"},
		{key: "fastapi", name: "FastAPI"},
		{key: "flask", name: "Flask"},
		{key: "sqlalchemy", name: "SQLAlchemy"},
		{key: "pytest", name: "Pytest"},
		{key: "celery", name: "Celery"},
	}
	for _, filename := range []string{"pyproject.toml", "requirements.txt"} {
		data, err := os.ReadFile(filepath.Join(targetDir, filename))
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))
		for _, match := range matches {
			if !strings.Contains(content, match.key) {
				continue
			}
			add(match.name, versionFromPythonContent(string(data), match.key), filename+":"+match.key)
		}
	}
}

func detectMarkerFrameworks(targetDir string, add func(name, version, source string)) {
	markers := []struct {
		file string
		name string
	}{
		{file: "Cargo.toml", name: "Rust"},
		{file: "pom.xml", name: "Maven"},
		{file: "Gemfile", name: "Bundler"},
		{file: "composer.json", name: "Composer"},
		{file: "mix.exs", name: "Mix"},
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(targetDir, marker.file)); err == nil {
			add(marker.name, "", marker.file)
		}
	}
}

func versionFromGoMod(content, module string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		for i, field := range fields {
			if strings.Contains(field, module) && i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}
	return ""
}

func versionFromPythonContent(content, pkg string) string {
	lowerPkg := strings.ToLower(pkg)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, lowerPkg) {
			continue
		}
		for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<"} {
			if idx := strings.Index(trimmed, sep); idx >= 0 {
				return strings.TrimSpace(trimmed[idx+len(sep):])
			}
		}
	}
	return ""
}
