package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/generator"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	internalorchestrator "github.com/rluisb/lazyai/packages/cli/internal/orchestrator"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var orchestrationCmd = &cobra.Command{
	Use:   "orchestration",
	Short: "Manage orchestration (chains, teams, workflows)",
	Long:  "Create, list, and inspect orchestration configurations including chains, teams, and workflows.",
}

var orchestrationListCmd = &cobra.Command{
	Use:   "list [kind]",
	Short: "List orchestration configurations",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runOrchestrationList,
}

var orchestrationCreateCmd = &cobra.Command{
	Use:   "create <type> <name>",
	Short: "Create a new orchestration configuration",
	Args:  cobra.ExactArgs(2),
	RunE:  runOrchestrationCreate,
}

var orchestrationStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show orchestration status",
	RunE:  runOrchestrationStatus,
}

func init() {
	orchestrationListCmd.Flags().String("dir", "", "Target directory (default: current directory)")
	orchestrationListCmd.Flags().Bool("json", false, "Output as JSON")

	orchestrationCreateCmd.Flags().String("description", "", "Artifact description")
	orchestrationCreateCmd.Flags().Bool("force", false, "Overwrite files if they already exist")
	orchestrationCreateCmd.Flags().Bool("no-interactive", false, "Disable interactive prompts")
	orchestrationCreateCmd.Flags().String("chain", "", "Primary chain reference for workflow creation")
	orchestrationCreateCmd.Flags().String("team", "", "Optional review/synthesis team reference for workflow creation")

	orchestrationStatusCmd.Flags().String("dir", "", "Target directory (default: current directory)")
	orchestrationStatusCmd.Flags().Bool("json", false, "Output as JSON")

	// Catalog commands
	catalogCmd.Flags().String("db", "", "Orchestrator DB path (default: ~/.local/share/lazyai-orchestrator/orchestrator.db)")
	catalogCmd.Flags().Bool("json", false, "Output as JSON")

	catalogVersionsCmd.Flags().String("db", "", "Orchestrator DB path")
	catalogVersionsCmd.Flags().Bool("json", false, "Output as JSON")
	catalogVersionsCmd.Flags().String("kind", "", "Catalog kind (chain, team, workflow, skill, mode, agent, command)")

	catalogCreateCmd.Flags().String("db", "", "Orchestrator DB path")
	catalogCreateCmd.Flags().String("kind", "", "Catalog kind")
	catalogCreateCmd.Flags().String("file", "", "Path to definition JSON file (required)")

	catalogSetActiveCmd.Flags().String("db", "", "Orchestrator DB path")
	catalogSetActiveCmd.Flags().String("kind", "", "Catalog kind")

	catalogDiffCmd.Flags().String("db", "", "Orchestrator DB path")
	catalogDiffCmd.Flags().String("kind", "", "Catalog kind")

	catalogCmd.AddCommand(catalogVersionsCmd)
	catalogCmd.AddCommand(catalogCreateCmd)
	catalogCmd.AddCommand(catalogSetActiveCmd)
	catalogCmd.AddCommand(catalogDiffCmd)

	orchestrationCmd.AddCommand(orchestrationListCmd)
	orchestrationCmd.AddCommand(orchestrationCreateCmd)
	orchestrationCmd.AddCommand(orchestrationStatusCmd)
	orchestrationCmd.AddCommand(catalogCmd)
	rootCmd.AddCommand(orchestrationCmd)
}

// ---------------------------------------------------------------------------
// Catalog subcommands
// ---------------------------------------------------------------------------

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Manage catalog versioning (definitions, versions, active pointers)",
	Long: `Manage the internal versioned catalog.

Subcommands:
  catalog                        List all catalog definitions with their active version
  catalog versions <name>        List all versions of a definition
  catalog create <name>          Create a new version from a file
  catalog set-active <name> <v>  Set the active version for a definition
  catalog diff <name> <v1> <v2>  Compare two versions of a definition`,
	Args: cobra.NoArgs,
	RunE: runCatalogList,
}

var catalogVersionsCmd = &cobra.Command{
	Use:   "versions <name>",
	Short: "List all versions of a catalog definition",
	Args:  cobra.ExactArgs(1),
	RunE:  runCatalogVersions,
}

var catalogCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new version of a catalog definition from a file",
	Args:  cobra.ExactArgs(1),
	RunE:  runCatalogCreate,
}

var catalogSetActiveCmd = &cobra.Command{
	Use:   "set-active <name> <version>",
	Short: "Set the active version for a catalog definition",
	Args:  cobra.ExactArgs(2),
	RunE:  runCatalogSetActive,
}

var catalogDiffCmd = &cobra.Command{
	Use:   "diff <name> <v1> <v2>",
	Short: "Show diff between two versions of a catalog definition",
	Args:  cobra.ExactArgs(3),
	RunE:  runCatalogDiff,
}

// openCatalogDB resolves and opens the orchestrator catalog DB.
func openCatalogDB(cmd *cobra.Command) (*internalorchestrator.CatalogDB, error) {
	dbPath, _ := cmd.Flags().GetString("db")
	if dbPath == "" {
		var err error
		dbPath, err = internalorchestrator.DefaultCatalogDBPath()
		if err != nil {
			return nil, fmt.Errorf("resolve default db path: %w", err)
		}
	}
	return internalorchestrator.OpenCatalogDB(dbPath)
}

// resolveKind determines the catalog kind from --kind flag or by inferring from name.
func resolveKind(cmd *cobra.Command, name string) string {
	if kind, _ := cmd.Flags().GetString("kind"); kind != "" {
		return kind
	}
	// Try to infer kind from plural name forms.
	switch name {
	case "chains", "chain":
		return "chain"
	case "teams", "team":
		return "team"
	case "workflows", "workflow":
		return "workflow"
	case "domains", "domain", "skills":
		return "skill"
	case "modes", "mode":
		return "mode"
	case "agents", "agent":
		return "agent"
	case "commands", "command":
		return "command"
	}
	return ""
}

func runCatalogList(cmd *cobra.Command, args []string) error {
	cdb, err := openCatalogDB(cmd)
	if err != nil {
		return err
	}
	defer cdb.Close()

	outputJSON, _ := cmd.Flags().GetBool("json")
	defs, err := cdb.ListDefinitions("")
	if err != nil {
		return fmt.Errorf("list definitions: %w", err)
	}

	if outputJSON {
		return writeJSON(defs)
	}

	fmt.Print(internalorchestrator.FormatCatalogList(defs))
	return nil
}

func runCatalogVersions(cmd *cobra.Command, args []string) error {
	cdb, err := openCatalogDB(cmd)
	if err != nil {
		return err
	}
	defer cdb.Close()

	name := args[0]
	kind := resolveKind(cmd, name)
	if kind == "" {
		return fmt.Errorf("could not determine catalog kind for %q; use --kind flag", name)
	}

	outputJSON, _ := cmd.Flags().GetBool("json")
	versions, err := cdb.ListVersions(kind, name)
	if err != nil {
		return fmt.Errorf("list versions: %w", err)
	}

	if outputJSON {
		return writeJSON(versions)
	}

	fmt.Printf("Versions of %s/%s:\n", kind, name)
	fmt.Print(internalorchestrator.FormatVersionList(versions))
	return nil
}

func runCatalogCreate(cmd *cobra.Command, args []string) error {
	cdb, err := openCatalogDB(cmd)
	if err != nil {
		return err
	}
	defer cdb.Close()

	name := args[0]
	kind := resolveKind(cmd, name)
	if kind == "" {
		return fmt.Errorf("could not determine catalog kind for %q; use --kind flag", name)
	}

	filePath, _ := cmd.Flags().GetString("file")
	if filePath == "" {
		return fmt.Errorf("--file is required")
	}

	result, err := cdb.CreateVersion(kind, name, filePath, "", true)
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}

	if result.AlreadyExists {
		fmt.Printf("Version v%d already exists (unchanged, checksum %s)\n", result.Version, result.Checksum)
	} else {
		fmt.Printf("Created v%d of %s/%s (checksum %s)\n", result.Version, kind, name, result.Checksum)
	}
	return nil
}

func runCatalogSetActive(cmd *cobra.Command, args []string) error {
	cdb, err := openCatalogDB(cmd)
	if err != nil {
		return err
	}
	defer cdb.Close()

	name := args[0]
	kind := resolveKind(cmd, name)
	if kind == "" {
		return fmt.Errorf("could not determine catalog kind for %q; use --kind flag", name)
	}

	versionStr := args[1]
	version := 0
	if _, err := fmt.Sscanf(versionStr, "v%d", &version); err != nil {
		// Try plain integer
		if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
			return fmt.Errorf("invalid version: %q (expected vN or N)", versionStr)
		}
	}

	if err := cdb.SetActive(kind, name, version); err != nil {
		return fmt.Errorf("set active: %w", err)
	}

	fmt.Printf("Set %s/%s active version to v%d\n", kind, name, version)
	return nil
}

func runCatalogDiff(cmd *cobra.Command, args []string) error {
	cdb, err := openCatalogDB(cmd)
	if err != nil {
		return err
	}
	defer cdb.Close()

	name := args[0]
	kind := resolveKind(cmd, name)
	if kind == "" {
		return fmt.Errorf("could not determine catalog kind for %q; use --kind flag", name)
	}

	var fromV, toV int
	if _, err := fmt.Sscanf(args[1], "v%d", &fromV); err != nil {
		if _, err := fmt.Sscanf(args[1], "%d", &fromV); err != nil {
			return fmt.Errorf("invalid from-version: %q", args[1])
		}
	}
	if _, err := fmt.Sscanf(args[2], "v%d", &toV); err != nil {
		if _, err := fmt.Sscanf(args[2], "%d", &toV); err != nil {
			return fmt.Errorf("invalid to-version: %q", args[2])
		}
	}

	diff, err := cdb.DiffVersions(kind, name, fromV, toV)
	if err != nil {
		return fmt.Errorf("diff versions: %w", err)
	}

	outputJSON, _ := cmd.Flags().GetBool("json")
	if outputJSON {
		return writeJSON(diff)
	}

	if diff.From == nil {
		fmt.Printf("(v%d not found)\n", fromV)
	} else {
		fmt.Printf("--- v%d  %s  (%d bytes)\n", diff.From.Version, internalorchestrator.FormatCatalogDate(diff.From.CreatedAt), len(diff.From.Body))
	}
	if diff.To == nil {
		fmt.Printf("(v%d not found)\n", toV)
	} else {
		fmt.Printf("+++ v%d  %s  (%d bytes)\n", diff.To.Version, internalorchestrator.FormatCatalogDate(diff.To.CreatedAt), len(diff.To.Body))
	}

	if diff.From != nil && diff.To != nil && diff.From.Checksum == diff.To.Checksum {
		fmt.Println("(no changes)")
	}
	return nil
}

var orchestrationCategories = []internalorchestrator.ListCategory{
	internalorchestrator.CategoryWorkflows,
	internalorchestrator.CategoryChains,
	internalorchestrator.CategoryTeams,
	internalorchestrator.CategoryDomains,
	internalorchestrator.CategoryModes,
}

func runOrchestrationList(cmd *cobra.Command, args []string) error {
	projectDir, err := orchestrationProjectDir(cmd)
	if err != nil {
		return err
	}
	libraryDir, err := orchestrationLibraryDir()
	if err != nil {
		return err
	}
	outputJSON, _ := cmd.Flags().GetBool("json")

	var categories []internalorchestrator.ListCategory
	if len(args) == 1 {
		category, err := parseOrchestrationCategory(args[0])
		if err != nil {
			return err
		}
		categories = []internalorchestrator.ListCategory{category}
	}

	result, err := internalorchestrator.ListCatalog(projectDir, libraryDir, categories)
	if err != nil {
		return err
	}

	if outputJSON {
		return writeJSON(result)
	}

	for _, category := range orchestrationListOrder(categories) {
		items := result[category]
		fmt.Printf("%s (%d)\n", category, len(items))
		for _, item := range items {
			fmt.Printf("- %s [%s]", item.Name, item.Source)
			if item.Description != "" {
				fmt.Printf(" - %s", item.Description)
			}
			fmt.Println()
		}
		if len(items) == 0 {
			fmt.Println("- none")
		}
	}

	return nil
}

func runOrchestrationCreate(cmd *cobra.Command, args []string) error {
	artifactType, err := parseOrchestrationCreateType(args[0])
	if err != nil {
		return err
	}

	targetDir, err := os.Getwd()
	if err != nil {
		return err
	}
	description, _ := cmd.Flags().GetString("description")
	force, _ := cmd.Flags().GetBool("force")
	chainRef, _ := cmd.Flags().GetString("chain")
	teamRef, _ := cmd.Flags().GetString("team")

	registry := generator.NewRegistry()
	gen, err := registry.Get(artifactType)
	if err != nil {
		return fmt.Errorf("generator not found: %w", err)
	}

	config := generator.GeneratorConfig{
		Name:        args[1],
		Description: description,
		TargetDir:   targetDir,
		Force:       force,
		Answers:     map[string]string{},
	}
	if artifactType == types.ArtifactTypeWorkflow {
		if chainRef != "" {
			config.Answers["chain"] = chainRef
		}
		if teamRef != "" {
			config.Answers["team"] = teamRef
		}
	}

	files, err := gen.Generate(config)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	for _, file := range files {
		absPath := filepath.Join(targetDir, file.Path)
		if !force {
			if _, statErr := os.Stat(absPath); statErr == nil {
				return fmt.Errorf("file already exists: %s (use --force to overwrite)", file.Path)
			}
		}
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", filepath.Dir(absPath), err)
		}
		if err := os.WriteFile(absPath, []byte(file.Content), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", absPath, err)
		}
		fmt.Printf("created %s\n", file.Path)
	}

	return nil
}

func runOrchestrationStatus(cmd *cobra.Command, args []string) error {
	projectDir, err := orchestrationProjectDir(cmd)
	if err != nil {
		return err
	}
	libraryDir, err := orchestrationLibraryDir()
	if err != nil {
		return err
	}
	outputJSON, _ := cmd.Flags().GetBool("json")

	counts := internalorchestrator.GetCounts(projectDir, libraryDir)
	if outputJSON {
		return writeJSON(counts)
	}

	summary := "not scaffolded"
	if counts.Scaffolded {
		summary = "scaffolded"
	}
	fmt.Printf("orchestration: %s\n", summary)
	for _, category := range orchestrationCategories {
		fmt.Printf("- %s: project=%d library=%d\n", category, counts.Project[category], counts.Library[category])
	}

	return nil
}

func orchestrationProjectDir(cmd *cobra.Command) (string, error) {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		return os.Getwd()
	}
	return filepath.Abs(dir)
}

func orchestrationLibraryDir() (string, error) {
	dir, err := library.FindLibraryDir()
	if err != nil || dir == "" {
		return "", fmt.Errorf("could not resolve library directory")
	}
	return dir, nil
}

func parseOrchestrationCategory(value string) (internalorchestrator.ListCategory, error) {
	switch value {
	case string(internalorchestrator.CategoryWorkflows):
		return internalorchestrator.CategoryWorkflows, nil
	case string(internalorchestrator.CategoryChains):
		return internalorchestrator.CategoryChains, nil
	case string(internalorchestrator.CategoryTeams):
		return internalorchestrator.CategoryTeams, nil
	case string(internalorchestrator.CategoryDomains):
		return internalorchestrator.CategoryDomains, nil
	case string(internalorchestrator.CategoryModes):
		return internalorchestrator.CategoryModes, nil
	default:
		return "", fmt.Errorf("unsupported orchestration list kind: %s", value)
	}
}

func parseOrchestrationCreateType(value string) (types.ArtifactType, error) {
	switch value {
	case string(types.ArtifactTypeWorkflow):
		return types.ArtifactTypeWorkflow, nil
	case string(types.ArtifactTypeChain):
		return types.ArtifactTypeChain, nil
	case string(types.ArtifactTypeTeam):
		return types.ArtifactTypeTeam, nil
	case string(types.ArtifactTypeDomain):
		return types.ArtifactTypeDomain, nil
	case string(types.ArtifactTypeMode):
		return types.ArtifactTypeMode, nil
	default:
		return "", fmt.Errorf("unsupported orchestration create type: %s", value)
	}
}

func orchestrationListOrder(categories []internalorchestrator.ListCategory) []internalorchestrator.ListCategory {
	if len(categories) > 0 {
		return categories
	}
	return orchestrationCategories
}

func writeJSON(value any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
