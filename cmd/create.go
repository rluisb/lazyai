package cmd

import (
	"fmt"

	"charm.land/huh/v2"

	"github.com/spf13/cobra"
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
	createCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	force, _ := cmd.Flags().GetBool("force")

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

		if err := huh.NewForm(huh.NewGroup(typeSelect)).Run(); err != nil {
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

		if err := huh.NewForm(huh.NewGroup(nameInput)).Run(); err != nil {
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

	// TODO: Implement actual creation logic once scaffold packages are ported.
	fmt.Printf("Would create %s/%s (force=%v)\n", artifactType, artifactName, force)
	return nil
}
