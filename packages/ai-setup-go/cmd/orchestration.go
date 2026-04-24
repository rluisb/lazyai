package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/generator"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	internalorchestrator "github.com/ricardoborges-teachable/ai-setup/internal/orchestrator"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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

	orchestrationCmd.AddCommand(orchestrationListCmd)
	orchestrationCmd.AddCommand(orchestrationCreateCmd)
	orchestrationCmd.AddCommand(orchestrationStatusCmd)
	rootCmd.AddCommand(orchestrationCmd)
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
	case string(types.ArtifactTypeDomain):
		return types.ArtifactTypeDomain, nil
	case string(types.ArtifactTypeMode):
		return types.ArtifactTypeMode, nil
	case "chain", "team":
		return "", fmt.Errorf("orchestration create %s is not implemented", value)
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
