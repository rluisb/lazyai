package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime/dispatch"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime/workflow"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Workflow represents a workflow definition (legacy CLI format)
type Workflow struct {
	Name        string           `yaml:"name"`
	Version     string           `yaml:"version"`
	Description string           `yaml:"description"`
	Phases      []WorkflowPhase  `yaml:"phases"`
	Fallback    WorkflowFallback `yaml:"fallback"`
	Metadata    WorkflowMetadata `yaml:"metadata"`
}

// WorkflowPhase represents a single phase in a workflow
type WorkflowPhase struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Agent       string         `yaml:"agent"`
	Mode        string         `yaml:"mode"`
	Inputs      []string       `yaml:"inputs"`
	Outputs     []string       `yaml:"outputs"`
	Gates       []WorkflowGate `yaml:"gates"`
}

// WorkflowGate represents a human or auto gate
type WorkflowGate struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// WorkflowFallback represents fallback behavior
type WorkflowFallback struct {
	OnAgentFailure struct {
		Action        string `yaml:"action"`
		MaxRetries    int    `yaml:"max_retries"`
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
					var wf Workflow
					if err := yaml.Unmarshal(data, &wf); err == nil {
						fmt.Printf("  • %s", name)
						if wf.Description != "" {
							fmt.Printf(" - %s", wf.Description)
						}
						fmt.Printf(" (%d phases)\n", len(wf.Phases))
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

		var wf Workflow
		if err := yaml.Unmarshal(data, &wf); err != nil {
			return fmt.Errorf("error parsing workflow: %w", err)
		}

		fmt.Printf("Workflow: %s\n", wf.Name)
		fmt.Printf("Version: %s\n", wf.Version)
		if wf.Description != "" {
			fmt.Printf("Description: %s\n", wf.Description)
		}
		fmt.Println()

		fmt.Println("Phases:")
		for i, phase := range wf.Phases {
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

		if wf.Metadata.EstimatedDuration != "" {
			fmt.Printf("\nEstimated Duration: %s\n", wf.Metadata.EstimatedDuration)
		}
		if wf.Metadata.Complexity != "" {
			fmt.Printf("Complexity: %s\n", wf.Metadata.Complexity)
		}
		if wf.Metadata.RequiresApproval {
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

		var wf Workflow
		if err := yaml.Unmarshal(data, &wf); err != nil {
			return fmt.Errorf("error parsing workflow: %w", err)
		}

		// Open runtime DB for workflow tracking
		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Create a minimal session for this workflow run
		sessionID := os.Getenv("LAZYAI_SESSION")
		if sessionID == "" {
			sessionID = fmt.Sprintf("wf_%d", time.Now().Unix())
		}

		if err := ensureSession(db, sessionID); err != nil {
			return fmt.Errorf("error ensuring session: %w", err)
		}

		// Create workflow manager with a mock dispatcher for CLI simulation
		disp := dispatch.NewMockDispatcher()
		_ = workflow.NewManager(db, nil, disp)

		// Create workflow run record in DB
		result := &workflow.RunResult{
			WorkflowName: wf.Name,
			SessionID:    sessionID,
			Status:       "running",
			CurrentStep:  0,
			TotalSteps:   len(wf.Phases),
			StartedAt:    time.Now(),
		}
		if err := createWorkflowRun(db, result); err != nil {
			return fmt.Errorf("create workflow run record: %w", err)
		}

		fmt.Printf("🚀 Executing workflow: %s\n", wf.Name)
		if wf.Description != "" {
			fmt.Printf("   %s\n", wf.Description)
		}
		fmt.Println()

		if dryRun {
			fmt.Println("📋 DRY RUN MODE - No actual execution")
			fmt.Println()
		}

		// Execute phases
		for i, phase := range wf.Phases {
			result.CurrentStep = i + 1
			fmt.Printf("Phase %d/%d: %s\n", i+1, len(wf.Phases), phase.Name)
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
							result.Status = "halted"
							_ = updateWorkflowRun(db, result)
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
					"workflow": wf.Name,
					"phase":    phase.Name,
					"agent":    phase.Agent,
				})
			}

			fmt.Println()
		}

		// Mark complete
		if dryRun {
			result.Status = "dry_run"
			fmt.Println("✅ Dry run complete. Use --dry-run=false to execute.")
		} else {
			result.Status = "completed"
			fmt.Println("✅ Workflow completed successfully!")
		}
		now := time.Now()
		result.CompletedAt = &now
		_ = updateWorkflowRun(db, result)

		return nil
	},
}

var workflowSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync workflow definitions to runtime database",
	Long:  `Read workflow YAML files and sync them to the runtime database for tracking.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		workflowsDir := filepath.Join(".opencode", "workflows")
		wfMgr := workflow.NewManager(db, nil, nil)

		if err := wfMgr.Sync(workflowsDir); err != nil {
			return fmt.Errorf("sync workflows: %w", err)
		}

		fmt.Println("✅ Workflows synced to runtime database")
		return nil
	},
}

func init() {
	workflowRunCmd.Flags().Bool("dry-run", true, "Show what would be executed without running")

	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowCmd.AddCommand(workflowSyncCmd)
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.GroupID = "runtime"
}

// createWorkflowRun inserts a workflow run record into the database.
func createWorkflowRun(db *runtime.DB, result *workflow.RunResult) error {
	res, err := db.Exec(
		"INSERT INTO workflow_instances (workflow_name, session_id, status, current_step, total_steps, started_at) VALUES (?, ?, ?, ?, ?, ?)",
		result.WorkflowName, result.SessionID, result.Status, result.CurrentStep, result.TotalSteps, result.StartedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert workflow instance: %w", err)
	}
	id, _ := res.LastInsertId()
	result.InstanceID = int(id)
	return nil
}

// updateWorkflowRun updates a workflow run record in the database.
func updateWorkflowRun(db *runtime.DB, result *workflow.RunResult) error {
	var completedAt interface{}
	if result.CompletedAt != nil {
		completedAt = result.CompletedAt.Format(time.RFC3339)
	}
	_, err := db.Exec(
		"UPDATE workflow_instances SET status = ?, current_step = ?, result = ?, error_message = ?, completed_at = ? WHERE id = ?",
		result.Status, result.CurrentStep, result.Result, result.ErrorMessage, completedAt, result.InstanceID,
	)
	return err
}
