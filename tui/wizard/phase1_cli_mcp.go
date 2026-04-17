package wizard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
)

type McpCatalog struct {
	Servers  map[string]McpServer `json:"servers"`
	CliTools map[string]CliTool   `json:"cliTools"`
}

type McpServer struct {
	Description     string `json:"description"`
	RequiresInstall bool   `json:"requiresInstall"`
	InstallHint     string `json:"installHint"`
}

type CliTool struct {
	Description string `json:"description"`
	InstallHint string `json:"installHint"`
}

func loadMcpCatalog() (*McpCatalog, error) {
	// We assume the catalog is in the library directory relative to project root.
	// In a real setup, we'd use a consistent way to find the library path.
	// For now, we use a path relative to the current working directory.
	path := filepath.Join("library", "mcp", "catalog.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp catalog: %w", err)
	}
	var catalog McpCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse mcp catalog: %w", err)
	}
	return &catalog, nil
}

func NewCliToolsSelect(defaults []string) *huh.MultiSelect[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return huh.NewMultiSelect[string]().Title("CLI Tools (Error loading catalog)")
	}

	var options []huh.Option[string]
	for id := range catalog.CliTools {
		options = append(options, huh.NewOption(id, id)) // In a real version, we'd use tool.Description
	}

	selected := defaults
	return huh.NewMultiSelect[string]().
		Title("Which CLI tools would you like to enable?").
		Options(options...).
		Value(&selected)
}

func NewMcpServersSelect(defaults []string) *huh.MultiSelect[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return huh.NewMultiSelect[string]().Title("MCP Servers (Error loading catalog)")
	}

	var options []huh.Option[string]
	for id := range catalog.Servers {
		options = append(options, huh.NewOption(id, id))
	}

	selected := defaults
	return huh.NewMultiSelect[string]().
		Title("Which MCP servers would you like to enable?").
		Options(options...).
		Value(&selected)
}
