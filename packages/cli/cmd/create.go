package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/generator"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var createCmd = &cobra.Command{
	Use:   "create [type] [name]",
	Short: "Generate new artifacts",
	Long:  "Create new artifacts such as agents, skills, workflows, domains, modes, prompts, commands, or templates.",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().Bool("force", false, "Overwrite existing artifact")
	createCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	createCmd.Flags().String("description", "", "Description of the artifact")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")
	force, _ := cmd.Flags().GetBool("force")
	description, _ := cmd.Flags().GetString("description")

	var artifactType, artifactName string

	if len(args) >= 2 {
		artifactType = args[0]
		artifactName = args[1]
	} else if len(args) == 1 {
		artifactType = args[0]
	} else if !nonInteractive {
		// Interactive: ask for type and name.
		var typeValue string
		typeSelect := huh.NewSelect[string]().
			Title("What type of artifact to create?").
			Options(
				huh.NewOption("Agent", "agent"),
				huh.NewOption("Skill", "skill"),
				huh.NewOption("Workflow", "workflow"),
				huh.NewOption("Domain", "domain"),
				huh.NewOption("Mode", "mode"),
				huh.NewOption("Prompt", "prompt"),
				huh.NewOption("Command", "command"),
				huh.NewOption("Template", "template"),
			).
			Value(&typeValue)

		if err := theme.NewForm(huh.NewGroup(typeSelect)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		artifactType = typeValue

		var nameValue string
		nameInput := huh.NewInput().
			Title(fmt.Sprintf("Name for the %s:", artifactType)).
			Placeholder("my-new-artifact").
			Value(&nameValue).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("name is required")
				}
				return nil
			})

		if err := theme.NewForm(huh.NewGroup(nameInput)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		artifactName = nameValue
	} else {
		return fmt.Errorf("type and name are required in non-interactive mode")
	}

	if artifactType == "" {
		return fmt.Errorf("artifact type is required")
	}
	if artifactName == "" {
		return fmt.Errorf("artifact name is required")
	}

	// Validate artifact type.
	validTypes := map[string]bool{
		"agent": true, "skill": true, "workflow": true,
		"domain": true, "mode": true, "prompt": true,
		"command": true, "template": true,
	}
	if !validTypes[artifactType] {
		return fmt.Errorf("invalid artifact type: %s (valid: agent, skill, workflow, domain, mode, prompt, command, template)", artifactType)
	}

	// Get the target directory.
	targetDir, _ := os.Getwd()

	// Use the generator registry to create the artifact.
	registry := generator.NewRegistry()
	gen, err := registry.Get(types.ArtifactType(artifactType))
	if err != nil {
		return fmt.Errorf("generator not found: %w", err)
	}

	// Build generator config.
	config := generator.GeneratorConfig{
		Name:        artifactName,
		Description: description,
		TargetDir:   targetDir,
		Force:       force,
		Answers:     make(map[string]string),
	}

	// In interactive mode, prompt for required fields.
	if !nonInteractive {
		questions := gen.GetPromptQuestions()
		for _, q := range questions {
			if q.Required && config.Answers[q.Key] == "" {
				var answer string
				input := huh.NewInput().
					Title(q.Label).
					Placeholder(q.Default).
					Value(&answer)

				if err := theme.NewForm(huh.NewGroup(input)).Run(); err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}
				config.Answers[q.Key] = answer
			}
		}
	}

	// Generate the artifact.
	files, err := gen.Generate(config)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No files generated.")
		return nil
	}

	// Write the generated files.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("✨ Created %s: %s", artifactType, artifactName)))
	fmt.Println()

	for _, f := range files {
		absPath := filepath.Join(targetDir, f.Path)

		// Check if file exists and handle force flag.
		if !force {
			if _, err := os.Stat(absPath); err == nil {
				fmt.Printf("  %s %s (already exists, use --force to overwrite)\n",
					lipgloss.NewStyle().Foreground(theme.Warning).Render("⚠"),
					f.Path)
				continue
			}
		}

		// Ensure directory exists.
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", filepath.Dir(absPath), err)
		}

		if err := os.WriteFile(absPath, []byte(f.Content), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", absPath, err)
		}

		fmt.Printf("  %s %s\n", greenStyle.Render("✓"), f.Path)
	}

	fmt.Println()
	fmt.Printf("  %s %d file(s) created\n", greenStyle.Render("→"), countCreated(files, force, targetDir))
	fmt.Println()

	return nil
}

// countCreated counts how many files were actually created.
func countCreated(files []generator.GeneratedFile, force bool, targetDir string) int {
	count := 0
	for _, f := range files {
		absPath := filepath.Join(targetDir, f.Path)
		if force {
			count++
		} else {
			if _, err := os.Stat(absPath); err != nil {
				count++
			}
		}
	}
	// If all would be new, return total.
	if count == 0 {
		return len(files)
	}
	return count
}

// slugify converts a name to a filesystem-safe slug.
func slugify(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", "-"), "_", "-"))
}
