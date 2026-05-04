package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var listCmd = &cobra.Command{
	Use:   "list [category]",
	Short: "List installed artifacts",
	Long:  "List all installed agents, skills, workflows, and other artifacts in the current setup.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	listCmd.Flags().String("type", "", "Filter by artifact type (agents, skills, templates, rules, etc.)")
	listCmd.Flags().Bool("verbose", false, "Show detailed artifact information including file paths")
	listCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(listCmd)
}

// artifactGroup groups artifacts by their type.
type artifactGroup struct {
	Type  string
	Items []artifactItem
}

// artifactItem represents a single tracked artifact.
type artifactItem struct {
	Name string
	Path string
}

func runList(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	typeFilter, _ := cmd.Flags().GetString("type")
	verbose, _ := cmd.Flags().GetBool("verbose")
	outputJSON, _ := cmd.Flags().GetBool("json")

	// If category is given as positional arg, use it as the type filter
	if len(args) > 0 && typeFilter == "" {
		typeFilter = args[0]
	}

	// Read store data
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	// Classify artifacts by type from tracked files
	groups := classifyArtifacts(storeData.Files, typeFilter)

	if outputJSON {
		return outputListJSON(groups)
	}

	return outputListStyled(groups, verbose)
}

// classifyArtifacts groups tracked files by their artifact type based on path prefixes.
func classifyArtifacts(files []types.TrackedFile, typeFilter string) []artifactGroup {
	groupMap := make(map[string][]artifactItem)

	for _, f := range files {
		artifactType := classifyPath(f.Path)
		if typeFilter != "" && artifactType != typeFilter {
			continue
		}

		name := filepath.Base(f.Path)
		// Strip extension for display
		if ext := filepath.Ext(name); ext != "" {
			name = strings.TrimSuffix(name, ext)
		}

		groupMap[artifactType] = append(groupMap[artifactType], artifactItem{
			Name: name,
			Path: f.Path,
		})
	}

	// Sort groups by type name
	var groups []artifactGroup
	for typ, items := range groupMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
		groups = append(groups, artifactGroup{Type: typ, Items: items})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Type < groups[j].Type
	})

	return groups
}

// classifyPath returns the artifact type based on the file path prefix.
func classifyPath(path string) string {
	// Normalize to forward slashes
	path = filepath.ToSlash(path)

	// Check known directory prefixes (ordered longest first for best match)
	prefixes := []struct {
		prefix   string
		typeName string
	}{
		{".opencode/agents/", "agents"},
		{".opencode/skills/", "skills"},
		{".opencode/commands/", "commands"},
		{"spec/templates/", "templates"},
		{"spec/rules/", "rules"},
		{"spec/standards/", "standards"},
		{"spec/adrs/", "adrs"},
		{"spec/memory/", "memory"},
		{"spec/prompts/", "prompts"},
		{".ai/orchestration/", "orchestration"},
		{".ai/", "config"},
		{".github/", "github"},
		{".husky/", "husky"},
		{".pre-commit-config", "pre-commit"},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(path, p.prefix) {
			return p.typeName
		}
	}

	return "other"
}

func outputListJSON(groups []artifactGroup) error {
	result := make(map[string]any)
	for _, g := range groups {
		if len(g.Items) == 0 {
			continue
		}
		items := make([]map[string]string, len(g.Items))
		for i, item := range g.Items {
			items[i] = map[string]string{
				"name": item.Name,
				"path": item.Path,
			}
		}
		result[g.Type] = items
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputListStyled(groups []artifactGroup, verbose bool) error {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	typeEmoji := map[string]string{
		"agents":        "🤖",
		"skills":        "⚡",
		"templates":     "📄",
		"rules":         "📏",
		"commands":      "⌨️",
		"standards":     "📐",
		"adrs":          "📋",
		"memory":        "🧠",
		"prompts":       "💬",
		"orchestration": "🎛️",
		"config":        "⚙️",
		"github":        "🐙",
		"husky":         "🐕",
		"pre-commit":    "🪝",
		"other":         "📁",
	}

	emoji := func(t string) string {
		if e, ok := typeEmoji[t]; ok {
			return e
		}
		return "📁"
	}

	totalItems := 0
	fmt.Println()
	for _, g := range groups {
		if len(g.Items) == 0 {
			continue
		}
		totalItems += len(g.Items)

		fmt.Println(headerStyle.Render(fmt.Sprintf("%s %s", emoji(g.Type), capitalize(g.Type))))
		printKV("  Count", countStyle.Render(fmt.Sprintf("%d", len(g.Items))), labelStyle, lipgloss.NewStyle())
		fmt.Println()

		for _, item := range g.Items {
			if verbose {
				fmt.Printf("    %s %s\n", bulletStyle.Render("•"), item.Name)
				fmt.Printf("    %s %s\n", labelStyle.Render("path:"), pathStyle.Render(item.Path))
			} else {
				fmt.Printf("    %s %s\n", bulletStyle.Render("•"), item.Name)
			}
		}
		fmt.Println()
	}

	fmt.Printf("  %s items available\n", countStyle.Render(fmt.Sprintf("%d", totalItems)))
	return nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
