package wizard

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

var cliToolLookPath = exec.LookPath

type McpPreset string

const (
	McpPresetMinimal     McpPreset = "minimal"
	McpPresetRecommended McpPreset = "recommended"
	McpPresetFull        McpPreset = "full"
)

type McpCatalog struct {
	Servers  map[string]McpServer `json:"servers"`
	CliTools map[string]CliTool   `json:"cliTools"`
}

type McpServer struct {
	Description     string `json:"description"`
	Enabled         bool   `json:"enabled"`
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
		paths = append(paths, filepath.Join(filepath.Dir(file), "..", "..", "library", "mcp", "catalog.json"))
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
		Description("Start from the preset selection, then toggle individual setup resources.").
		Options(mcpServerOptionsFromCatalog(catalog)...).
		Value(&selected)
}

func normalizeMcpPreset(preset McpPreset) McpPreset {
	switch preset {
	case McpPresetMinimal, McpPresetFull:
		return preset
	case McpPresetRecommended, "":
		return McpPresetRecommended
	default:
		return McpPresetRecommended
	}
}

func defaultMcpSelection(current []string, preset McpPreset) []string {
	if len(current) > 0 {
		return cloneStrings(current)
	}
	return defaultMcpServersForPreset(preset)
}

func defaultMcpServersForPreset(preset McpPreset) []string {
	catalog, err := loadMcpCatalog()
	if err != nil || catalog == nil {
		return nil
	}

	ids := sortedCatalogServerIDs(catalog)
	switch normalizeMcpPreset(preset) {
	case McpPresetMinimal:
		return filterCatalogServerIDs(ids, map[string]struct{}{
			"filesystem": {},
			"ripgrep":    {},
		})
	case McpPresetFull:
		return ids
	default:
		selected := make([]string, 0, len(ids))
		for _, id := range ids {
			if server, ok := catalog.Servers[id]; ok && server.Enabled {
				selected = append(selected, id)
			}
		}
		return selected
	}
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
	ids := sortedCatalogServerIDs(catalog)

	options := make([]huh.Option[string], 0, len(ids))
	for _, id := range ids {
		options = append(options, huh.NewOption(id, id))
	}
	return options
}

func sortedCatalogServerIDs(catalog *McpCatalog) []string {
	ids := make([]string, 0, len(catalog.Servers))
	for id := range catalog.Servers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func filterCatalogServerIDs(ids []string, allowed map[string]struct{}) []string {
	filtered := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func skillOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(types.ALL_SKILLS))
	for _, skill := range types.ALL_SKILLS {
		options = append(options, huh.NewOption(string(skill), string(skill)))
	}
	return options
}

func agentOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(types.ALL_AGENTS))
	for _, agent := range types.ALL_AGENTS {
		options = append(options, huh.NewOption(string(agent), string(agent)))
	}
	return options
}
