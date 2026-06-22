package scaffold

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// PrintMCPEnvGuidance prints required environment variables from enabled MCP
// servers. LazyAI does not generate or manage .env files.
func PrintMCPEnvGuidance(targetDir string, _ *[]types.TrackedFile, _ types.ConflictStrategy, _ map[string]types.ConflictStrategy) error {
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

	sort.Slice(envVars, func(i, j int) bool {
		if envVars[i].name == envVars[j].name {
			return envVars[i].server < envVars[j].server
		}
		return envVars[i].name < envVars[j].name
	})

	fmt.Println("\n💡 MCP environment variables required")
	fmt.Println("   LazyAI does not create or manage .env files. Export these values in your shell,")
	fmt.Println("   secret manager, or local env file before running tools that need these MCP servers:")
	fmt.Println()

	seen := make(map[string]bool)
	for _, ev := range envVars {
		if seen[ev.name] {
			continue
		}
		seen[ev.name] = true
		fmt.Printf("   # Required by: %s\n", ev.server)
		fmt.Printf("   export %s=\"\"\n\n", ev.name)
	}

	return nil
}
