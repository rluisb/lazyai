package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime/taskqueue"
	"github.com/spf13/cobra"
)

// Explicit runtime package use to satisfy compiler tracking
var _ *runtime.DB

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task queue management",
	Long:  `Create, claim, and manage tasks in the queue.`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [description]",
	Short: "Create a new task",
	Long:  `Create a new task in the queue.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := args[0]

		if err := ValidateNotEmpty(description, "description"); err != nil {
			return err
		}

		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Use a default session for CLI-created tasks
		sessionID := os.Getenv("LAZYAI_SESSION")
		if sessionID == "" {
			sessionID = "cli"
		}

		if err := ensureSession(db, sessionID); err != nil {
			return fmt.Errorf("error ensuring session: %w", err)
		}

		mgr := taskqueue.NewManager(db)
		task, err := mgr.Enqueue(sessionID, "cli", description, taskqueue.EnqueueOptions{})
		if err != nil {
			return fmt.Errorf("error creating task: %w", err)
		}

		// Build a CLI-friendly task ID using the runtime DB integer ID
		taskID := fmt.Sprintf("task_%d", task.ID)

		fmt.Printf("✅ Task created: %s\n", taskID)
		fmt.Printf("   Description: %s\n", description)

		// Record to ledger (best-effort)
		appendToLedger("task_created", map[string]string{
			"task_id":     taskID,
			"description": description,
		})

		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `List all tasks in the queue.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		sessionID := os.Getenv("LAZYAI_SESSION")
		if sessionID == "" {
			sessionID = "cli"
		}

		mgr := taskqueue.NewManager(db)
		tasks, err := mgr.List(sessionID, "")
		if err != nil {
			return err
		}

		fmt.Println("Tasks:")
		fmt.Println("───────────────────────────────────────────────────────────────")

		if len(tasks) == 0 {
			fmt.Println("  No tasks found.")
			return nil
		}

		for _, t := range tasks {
			taskID := fmt.Sprintf("task_%d", t.ID)
			fmt.Printf("  %s [%s] %s\n", taskID, t.Status, t.Task)
			if len(t.Claims) > 0 {
				agents := make([]string, len(t.Claims))
				for i, c := range t.Claims {
					agents[i] = c.Agent
				}
				fmt.Printf("    Claimed by: %s\n", strings.Join(agents, ", "))
			}
		}

		return nil
	},
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim [task-id]",
	Short: "Claim a task",
	Long:  `Claim a task for processing. If no task-id is provided, claims the next available task.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get agent name from env or default
		agentName := os.Getenv("LAZYAI_AGENT")
		if agentName == "" {
			agentName = "unknown"
		}

		sessionID := os.Getenv("LAZYAI_SESSION")
		if sessionID == "" {
			sessionID = "cli"
		}

		if err := ensureSession(db, sessionID); err != nil {
			return fmt.Errorf("error ensuring session: %w", err)
		}

		mgr := taskqueue.NewManager(db)

		var task *taskqueue.Task
		if len(args) > 0 {
			taskIDStr := args[0]
			if err := ValidateTaskID(taskIDStr); err != nil {
				return err
			}

			// Map CLI task ID (task_<runtime-id>) to DB integer ID
			tasks, err := mgr.List(sessionID, "")
			if err != nil {
				return fmt.Errorf("error listing tasks: %w", err)
			}
			var targetID int
			found := false
			for _, t := range tasks {
				if fmt.Sprintf("task_%d", t.ID) == taskIDStr {
					targetID = t.ID
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("task %s not found", taskIDStr)
			}

			task, err = mgr.ClaimByID(sessionID, "cli", agentName, targetID)
			if err != nil {
				return fmt.Errorf("error claiming task: %w", err)
			}
			if task == nil {
				return fmt.Errorf("task %s is not available to claim", taskIDStr)
			}
			fmt.Printf("✅ Task claimed: %s\n", taskIDStr)
		} else {
			// Claim next available task
			task, err = mgr.Claim(sessionID, "cli", agentName)
			if err != nil {
				return fmt.Errorf("error claiming task: %w", err)
			}
			if task == nil {
				return fmt.Errorf("no tasks available to claim")
			}
			claimedTaskID := fmt.Sprintf("task_%d", task.ID)
			fmt.Printf("✅ Task claimed: %s\n", claimedTaskID)
		}
		fmt.Printf("   Claimed by: %s\n", agentName)

		return nil
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark a task as complete",
	Long:  `Mark a task as completed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskIDStr := args[0]

		if err := ValidateTaskID(taskIDStr); err != nil {
			return err
		}

		db, err := openRuntimeDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// The CLI uses "task_<runtime-id>" but the runtime uses integer IDs.
		// For complete, we parse the integer ID from the string ID.
		targetID, err := parseTaskID(taskIDStr)
		if err != nil {
			return err
		}

		sessionID := os.Getenv("LAZYAI_SESSION")
		if sessionID == "" {
			sessionID = "cli"
		}

		mgr := taskqueue.NewManager(db)

		agentName := os.Getenv("LAZYAI_AGENT")
		if agentName == "" {
			agentName = "unknown"
		}

		if err := mgr.Complete(targetID, agentName); err != nil {
			return fmt.Errorf("error completing task: %w", err)
		}

		fmt.Printf("✅ Task completed: %s\n", taskIDStr)

		return nil
	},
}

// formatTaskID returns the CLI-visible task ID for a runtime DB integer ID.
func formatTaskID(id int) string {
	return fmt.Sprintf("task_%d", id)
}

// parseTaskID extracts the integer runtime ID from a CLI task ID string.
func parseTaskID(id string) (int, error) {
	parts := strings.Split(id, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid task ID format: %s", id)
	}
	raw, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid task ID format: %s", id)
	}
	return raw, nil
}

func init() {
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskClaimCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	rootCmd.AddCommand(taskCmd)
	taskCmd.GroupID = "runtime"
}
