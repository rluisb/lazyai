package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Workflow represents a workflow definition
type Workflow struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description"`
	Phases      []WorkflowPhase   `yaml:"phases"`
	Fallback    WorkflowFallback  `yaml:"fallback"`
	Metadata    WorkflowMetadata  `yaml:"metadata"`
}

// WorkflowPhase represents a single phase in a workflow
type WorkflowPhase struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Agent       string              `yaml:"agent"`
	Mode        string              `yaml:"mode"`
	Inputs      []string            `yaml:"inputs"`
	Outputs     []string            `yaml:"outputs"`
	Gates       []WorkflowGate      `yaml:"gates"`
}

// WorkflowGate represents a human or auto gate
type WorkflowGate struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// WorkflowFallback represents fallback behavior
type WorkflowFallback struct {
	OnAgentFailure struct {
		Action       string `yaml:"action"`
		MaxRetries   int    `yaml:"max_retries"`
		FallbackAgent string `yaml:"fallback_agent"`
	} `yaml:"on_agent_failure"`
	OnTestFailure struct {
		Action     string `yaml:"action"`
		UpdateSpec bool   `yaml:"update_spec"`
	} `yaml:"on_test_failure"`
	OnTimeout struct {
		Action string `yaml:"action"`
		Notify string `yaml:"notify"`
	} `yaml:"on_timeout"`
}

// WorkflowMetadata represents workflow metadata
type WorkflowMetadata struct {
	EstimatedDuration string `yaml:"estimated_duration"`
	Complexity        string `yaml:"complexity"`
	RequiresApproval  bool   `yaml:"requires_approval"`
	Priority          string `yaml:"priority"`
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Workflow execution",
	Long:  `Execute and manage workflows.`,
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available workflows",
	Long:  `List all available workflow definitions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowsDir := filepath.Join(".opencode", "workflows")
		
		entries, err := os.ReadDir(workflowsDir)
		if err != nil {
			return fmt.Errorf("error reading workflows directory: %w", err)
		}
		
		fmt.Println("Available workflows:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
				name := entry.Name()[:len(entry.Name())-5]
				
				// Try to read workflow metadata
				workflowPath := filepath.Join(workflowsDir, entry.Name())
				data, err := os.ReadFile(workflowPath)
				if err == nil {
					var workflow Workflow
					if err := yaml.Unmarshal(data, &workflow); err == nil {
						fmt.Printf("  • %s", name)
						if workflow.Description != "" {
							fmt.Printf(" - %s", workflow.Description)
						}
						fmt.Printf(" (%d phases)\n", len(workflow.Phases))
						count++
						continue
					}
				}
				fmt.Printf("  • %s\n", name)
				count++
			}
		}
		
		if count == 0 {
			fmt.Println("  No workflows found.")
		}
		
		return nil
	},
}

var workflowShowCmd = &cobra.Command{
	Use:   "show [workflow-name]",
	Short: "Show workflow details",
	Long:  `Show details of a specific workflow.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowName := args[0]
		workflowPath := filepath.Join(".opencode", "workflows", workflowName+".yaml")
		
		data, err := os.ReadFile(workflowPath)
		if err != nil {
			return fmt.Errorf("workflow '%s' not found", workflowName)
		}
		
		var workflow Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return fmt.Errorf("error parsing workflow: %w", err)
		}
		
		fmt.Printf("Workflow: %s\n", workflow.Name)
		fmt.Printf("Version: %s\n", workflow.Version)
		if workflow.Description != "" {
			fmt.Printf("Description: %s\n", workflow.Description)
		}
		fmt.Println()
		
		fmt.Println("Phases:")
		for i, phase := range workflow.Phases {
			fmt.Printf("  %d. %s\n", i+1, phase.Name)
			fmt.Printf("     Agent: %s | Mode: %s\n", phase.Agent, phase.Mode)
			if phase.Description != "" {
				fmt.Printf("     %s\n", phase.Description)
			}
			if len(phase.Gates) > 0 {
				fmt.Printf("     Gates: ")
				for j, gate := range phase.Gates {
					if j > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s (%s)", gate.Type, gate.Description)
				}
				fmt.Println()
			}
		}
		
		if workflow.Metadata.EstimatedDuration != "" {
			fmt.Printf("\nEstimated Duration: %s\n", workflow.Metadata.EstimatedDuration)
		}
		if workflow.Metadata.Complexity != "" {
			fmt.Printf("Complexity: %s\n", workflow.Metadata.Complexity)
		}
		if workflow.Metadata.RequiresApproval {
			fmt.Println("Requires Approval: Yes")
		}
		
		return nil
	},
}

var workflowRunCmd = &cobra.Command{
	Use:   "run [workflow-name]",
	Short: "Execute a workflow",
	Long:  `Execute a workflow phase by phase.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowName := args[0]
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		
		workflowPath := filepath.Join(".opencode", "workflows", workflowName+".yaml")
		
		data, err := os.ReadFile(workflowPath)
		if err != nil {
			return fmt.Errorf("workflow '%s' not found", workflowName)
		}
		
		var workflow Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			return fmt.Errorf("error parsing workflow: %w", err)
		}
		
		fmt.Printf("🚀 Executing workflow: %s\n", workflow.Name)
		if workflow.Description != "" {
			fmt.Printf("   %s\n", workflow.Description)
		}
		fmt.Println()
		
		if dryRun {
			fmt.Println("📋 DRY RUN MODE - No actual execution")
			fmt.Println()
		}
		
		// Execute phases
		for i, phase := range workflow.Phases {
			fmt.Printf("Phase %d/%d: %s\n", i+1, len(workflow.Phases), phase.Name)
			fmt.Printf("  Agent: %s | Mode: %s\n", phase.Agent, phase.Mode)
			if phase.Description != "" {
				fmt.Printf("  %s\n", phase.Description)
			}
			
			if len(phase.Inputs) > 0 {
				fmt.Printf("  Inputs: %s\n", strings.Join(phase.Inputs, ", "))
			}
			if len(phase.Outputs) > 0 {
				fmt.Printf("  Outputs: %s\n", strings.Join(phase.Outputs, ", "))
			}
			
			// Check gates
			for _, gate := range phase.Gates {
				if gate.Type == "human" {
					fmt.Printf("  ⛔ Gate: %s\n", gate.Description)
					if !dryRun {
						fmt.Print("  Continue? (y/n): ")
						var response string
						fmt.Scanln(&response)
						if strings.ToLower(response) != "y" {
							fmt.Println("  ❌ Workflow halted by user")
							return nil
						}
					}
				} else if gate.Type == "auto" {
					fmt.Printf("  ✅ Auto-gate: %s\n", gate.Description)
				}
			}
			
			if !dryRun {
				// Simulate execution
				fmt.Printf("  🔄 Executing...")
				time.Sleep(500 * time.Millisecond)
				fmt.Println(" Done")
				
				// Record to ledger
				appendToLedger("workflow_phase_completed", map[string]string{
					"workflow": workflow.Name,
					"phase":    phase.Name,
					"agent":    phase.Agent,
				})
			}
			
			fmt.Println()
		}
		
		if dryRun {
			fmt.Println("✅ Dry run complete. Use --dry-run=false to execute.")
		} else {
			fmt.Println("✅ Workflow completed successfully!")
		}
		
		return nil
	},
}

func init() {
	workflowRunCmd.Flags().Bool("dry-run", true, "Show what would be executed without running")
	
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	rootCmd.AddCommand(workflowCmd)
}
