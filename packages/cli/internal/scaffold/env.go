package scaffold

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldEnvExample generates a .env.example file listing required environment
// variables from enabled MCP servers. Ported from src/scaffold/env-example.ts.
func ScaffoldEnvExample(targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	mcpPath := filepath.Join(targetDir, ".ai", "mcp.json")
	if !files.FileExists(mcpPath) {
		return nil
	}

	data, err := files.ReadFile(mcpPath)
	if err != nil {
		return nil // silently skip if unreadable
	}

	var catalog mcpCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil // silently skip if invalid JSON
	}

	// Collect env vars from enabled servers.
	type envVar struct {
		name   string
		server string
	}
	var envVars []envVar

	for serverName, server := range catalog.Servers {
		if server.Enabled != nil && !*server.Enabled {
			continue
		}
		if server.Env == nil {
			continue
		}
		for key := range server.Env {
			envVars = append(envVars, envVar{name: key, server: serverName})
		}
	}

	if len(envVars) == 0 {
		return nil
	}

	dest := filepath.Join(targetDir, ".env.example")
	relPath, err := filepath.Rel(targetDir, dest)
	if err != nil {
		relPath = dest
	}

	action, err := conflict.ApplyStrategy(dest, strategy, perFileOverrides, targetDir)
	if err != nil {
		return err
	}
	if action == "skip" {
		log.Printf("Skipping existing file: %s", relPath)
		return nil
	}

	// Group by server for readability.
	var lines []string
	lines = append(lines,
		"# Environment variables required by enabled MCP servers",
		"# Copy this file to .env and fill in the values",
		"# NEVER commit .env to version control",
		"",
	)

	seen := make(map[string]bool)
	for _, ev := range envVars {
		if seen[ev.name] {
			continue
		}
		seen[ev.name] = true
		lines = append(lines,
			"# Required by: "+ev.server,
			ev.name+"=",
			"",
		)
	}

	if err := files.WriteFile(dest, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: "generated:env-example",
		Owner:  types.FileOwnerLibrary,
	})

	return nil
}
