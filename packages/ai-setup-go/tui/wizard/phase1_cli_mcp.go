package wizard

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"charm.land/huh/v2"
)

var cliToolLookPath = exec.LookPath

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
	paths := []string{filepath.Join("library", "mcp", "catalog.json")}
	if _, file, _, ok := runtime.Caller(0); ok {
		// Walk up from the source file looking for library/mcp/catalog.json.
		// Layout-agnostic: works whether the source file lives at
		// `<repo>/tui/wizard/` (pre-monorepo) or `<repo>/packages/ai-setup-go/tui/wizard/` (monorepo).
		dir := filepath.Dir(file)
		for i := 0; i < 10; i++ {
			paths = append(paths, filepath.Join(dir, "library", "mcp", "catalog.json"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	var data []byte
	var err error
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read mcp catalog: %w", err)
	}
	var catalog McpCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse mcp catalog: %w", err)
	}
	return &catalog, nil
}

func NewCliToolsSelect(defaults []string, preSelected []string) *huh.MultiSelect[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return huh.NewMultiSelect[string]().Title("CLI Tools (Error loading catalog)")
	}

	selected := cloneStrings(preSelected)
	if len(defaults) > 0 {
		selected = cloneStrings(defaults)
	}

	return huh.NewMultiSelect[string]().
		Title("Which CLI tools would you like to enable?").
		Options(cliToolOptionsFromCatalog(catalog)...).
		Value(&selected)
}

func NewMcpServersSelect(defaults []string) *huh.MultiSelect[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return huh.NewMultiSelect[string]().Title("MCP Servers (Error loading catalog)")
	}

	selected := defaults
	return huh.NewMultiSelect[string]().
		Title("Which MCP servers would you like to enable?").
		Options(mcpServerOptionsFromCatalog(catalog)...).
		Value(&selected)
}

func detectInstalledCliToolsFromCatalog() []string {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return nil
	}
	return detectInstalledCliTools(catalog)
}

func detectInstalledCliTools(catalog *McpCatalog) (installed []string) {
	defer func() {
		if recover() != nil {
			installed = nil
		}
	}()

	if catalog == nil {
		return nil
	}

	for id := range catalog.CliTools {
		if _, err := cliToolLookPath(id); err == nil {
			installed = append(installed, id)
		}
	}
	sort.Strings(installed)
	return installed
}

func cliToolOptionsFromCatalog(catalog *McpCatalog) []huh.Option[string] {
	ids := make([]string, 0, len(catalog.CliTools))
	for id := range catalog.CliTools {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	options := make([]huh.Option[string], 0, len(ids))
	for _, id := range ids {
		options = append(options, huh.NewOption(id, id))
	}
	return options
}

func mcpServerOptionsFromCatalog(catalog *McpCatalog) []huh.Option[string] {
	ids := make([]string, 0, len(catalog.Servers))
	for id := range catalog.Servers {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	options := make([]huh.Option[string], 0, len(ids))
	for _, id := range ids {
		options = append(options, huh.NewOption(id, id))
	}
	return options
}
