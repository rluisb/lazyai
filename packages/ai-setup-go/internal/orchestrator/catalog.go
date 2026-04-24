// Package orchestrator provides catalog listing and management for orchestration
// artifacts (chains, teams, workflows, domains, modes).
// Ported from src/orchestration/catalog.ts.
package orchestrator

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// ArtifactType identifies the kind of orchestration artifact.
type ArtifactType string

const (
	ArtifactTypeWorkflow ArtifactType = "workflow"
	ArtifactTypeChain    ArtifactType = "chain"
	ArtifactTypeTeam     ArtifactType = "team"
	ArtifactTypeDomain   ArtifactType = "domain"
	ArtifactTypeMode     ArtifactType = "mode"
)

// ListCategory identifies a category for listing.
type ListCategory string

const (
	CategoryWorkflows ListCategory = "workflows"
	CategoryChains    ListCategory = "chains"
	CategoryTeams     ListCategory = "teams"
	CategoryDomains   ListCategory = "domains"
	CategoryModes     ListCategory = "modes"
)

// Source indicates where an artifact comes from.
type Source string

const (
	SourceProject Source = "project"
	SourceLibrary Source = "library"
)

// CatalogItem describes a single orchestration artifact.
type CatalogItem struct {
	Type        ArtifactType      `json:"type"`
	Category    ListCategory      `json:"category"`
	Name        string            `json:"name"`
	Source      Source            `json:"source"`
	Path        string            `json:"path"`
	Description string            `json:"description,omitempty"`
	Content     string            `json:"-"`
	Data        map[string]any    `json:"data,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Body        string            `json:"body,omitempty"`
}

// CatalogSummary is a simplified view of a catalog item.
type CatalogSummary struct {
	Type        ArtifactType `json:"type"`
	Category    ListCategory `json:"category"`
	Name        string       `json:"name"`
	Source      Source       `json:"source"`
	Path        string       `json:"path"`
	Description string       `json:"description,omitempty"`
}

// CatalogCounts holds counts of artifacts in project and library.
type CatalogCounts struct {
	Scaffolded bool                  `json:"scaffolded"`
	Project    map[ListCategory]int  `json:"project"`
	Library    map[ListCategory]int  `json:"library"`
}

// ---------------------------------------------------------------------------
// Directory configuration
// ---------------------------------------------------------------------------

type dirConfig struct {
	Category    ListCategory
	Type        ArtifactType
	RelativeDir string
	Extension   string // ".json" or ".md"
}

var orchestrationDirs = []dirConfig{
	{Category: CategoryWorkflows, Type: ArtifactTypeWorkflow, RelativeDir: "workflows", Extension: ".json"},
	{Category: CategoryChains, Type: ArtifactTypeChain, RelativeDir: "chains", Extension: ".json"},
	{Category: CategoryTeams, Type: ArtifactTypeTeam, RelativeDir: "teams", Extension: ".json"},
	{Category: CategoryDomains, Type: ArtifactTypeDomain, RelativeDir: "skills/domains", Extension: ".md"},
	{Category: CategoryModes, Type: ArtifactTypeMode, RelativeDir: "skills/modes", Extension: ".md"},
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// ListCatalog lists orchestration artifacts by category.
// If kinds is empty, returns all categories.
func ListCatalog(projectDir, libraryDir string, kinds []ListCategory) (map[ListCategory][]CatalogSummary, error) {
	categories := kinds
	if len(categories) == 0 {
		categories = []ListCategory{CategoryWorkflows, CategoryChains, CategoryTeams, CategoryDomains, CategoryModes}
	}

	result := make(map[ListCategory][]CatalogSummary, len(categories))
	for _, cat := range categories {
		result[cat] = ListItems(projectDir, libraryDir, cat)
	}
	return result, nil
}

// ListItems lists all items for a given category, merging project and library
// sources (project overrides take precedence).
func ListItems(projectDir, libraryDir string, category ListCategory) []CatalogSummary {
	config, ok := getDirConfig(category)
	if !ok {
		return nil
	}

	projectRoot := getProjectRoot(projectDir)
	libraryRoot := getLibraryRoot(libraryDir)

	projectItems := readCatalogDirectory(projectRoot, config, SourceProject)
	libraryItems := readCatalogDirectory(libraryRoot, config, SourceLibrary)

	merged := mergeCatalogItems(projectItems, libraryItems)

	summaries := make([]CatalogSummary, 0, len(merged))
	for _, item := range merged {
		s := CatalogSummary{
			Type:     item.Type,
			Category: item.Category,
			Name:     item.Name,
			Source:   item.Source,
			Path:     item.Path,
		}
		if item.Description != "" {
			s.Description = item.Description
		}
		summaries = append(summaries, s)
	}
	return summaries
}

// FindItem searches all categories for an item matching query by name.
func FindItem(projectDir, libraryDir, query string) *CatalogItem {
	normalizedQuery := normalizeName(query)

	for _, config := range orchestrationDirs {
		items := ListItems(projectDir, libraryDir, config.Category)
		for _, item := range items {
			if normalizeName(item.Name) == normalizedQuery {
				rootDir := getProjectRoot(projectDir)
				source := SourceProject
				if item.Source == SourceLibrary {
					rootDir = getLibraryRoot(libraryDir)
					source = SourceLibrary
				}
				fullItems := readCatalogDirectory(rootDir, config, source)
				for _, fi := range fullItems {
					if normalizeName(fi.Name) == normalizedQuery {
						return &fi
					}
				}
			}
		}
	}
	return nil
}

// GetCounts returns the number of orchestration artifacts in project and library.
func GetCounts(projectDir, libraryDir string) CatalogCounts {
	projectRoot := getProjectRoot(projectDir)
	libraryRoot := getLibraryRoot(libraryDir)

	project := make(map[ListCategory]int)
	library := make(map[ListCategory]int)

	for _, config := range orchestrationDirs {
		project[config.Category] = countCategory(projectRoot, config)
		library[config.Category] = countCategory(libraryRoot, config)
	}

	return CatalogCounts{
		Scaffolded: files.DirExists(projectRoot),
		Project:    project,
		Library:    library,
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func getProjectRoot(projectDir string) string {
	return filepath.Join(projectDir, ".ai", "orchestration")
}

func getLibraryRoot(libraryDir string) string {
	return filepath.Join(libraryDir, "orchestration")
}

func getDirConfig(category ListCategory) (dirConfig, bool) {
	for _, c := range orchestrationDirs {
		if c.Category == category {
			return c, true
		}
	}
	return dirConfig{}, false
}

func readCatalogDirectory(rootDir string, config dirConfig, source Source) []CatalogItem {
	dir := filepath.Join(rootDir, config.RelativeDir)
	if !files.DirExists(dir) {
		return nil
	}

	entries := files.ListDir(dir)
	var items []CatalogItem

	for _, entry := range entries {
		if !strings.HasSuffix(entry, config.Extension) {
			continue
		}

		itemPath := filepath.Join(dir, entry)
		name := strings.TrimSuffix(entry, config.Extension)

		data, err := files.ReadFile(itemPath)
		if err != nil {
			continue
		}
		content := string(data)

		item := CatalogItem{
			Type:     config.Type,
			Category: config.Category,
			Name:     name,
			Source:   source,
			Path:     itemPath,
			Content:  content,
		}

		if config.Extension == ".json" {
			var jsonData map[string]any
			if err := json.Unmarshal(data, &jsonData); err == nil {
				item.Data = jsonData
				if desc, ok := jsonData["description"].(string); ok {
					item.Description = desc
				}
			}
		} else {
			fm, body, err := frontmatter.ParseYamlFrontmatter(content)
			if err == nil && fm != nil {
				item.Metadata = fm
				item.Body = body
				if desc, ok := fm["description"].(string); ok {
					item.Description = desc
				}
			}
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items
}

func mergeCatalogItems(projectItems, libraryItems []CatalogItem) []CatalogItem {
	merged := make(map[string]CatalogItem)

	for _, item := range projectItems {
		merged[normalizeName(item.Name)] = item
	}
	for _, item := range libraryItems {
		key := normalizeName(item.Name)
		if _, exists := merged[key]; !exists {
			merged[key] = item
		}
	}

	var result []CatalogItem
	for _, item := range merged {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func countCategory(rootDir string, config dirConfig) int {
	dir := filepath.Join(rootDir, config.RelativeDir)
	if !files.DirExists(dir) {
		return 0
	}
	count := 0
	for _, entry := range files.ListDir(dir) {
		if strings.HasSuffix(entry, config.Extension) {
			count++
		}
	}
	return count
}

func normalizeName(value string) string {
	s := strings.TrimSpace(value)
	s = strings.ToLower(s)
	// Remove extension.
	if idx := strings.LastIndex(s, "."); idx > 0 {
		s = s[:idx]
	}
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
