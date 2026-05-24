package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

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
		
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)
		
		taskID := fmt.Sprintf("task_%d", time.Now().Unix())
		
		_, err = database.Exec(
			"INSERT INTO tasks (task_id, description, status, created_at) VALUES (?, ?, 'pending', ?)",
			taskID, description, time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("error creating task: %w", err)
		}
		
		fmt.Printf("✅ Task created: %s\n", taskID)
		fmt.Printf("   Description: %s\n", description)
		
		// Record to ledger
		appendToLedger("task_created", map[string]string{
			"task_id": taskID,
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
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)
		
		rows, err := database.Query(
			"SELECT task_id, description, status, agent, created_at, claimed_by FROM tasks ORDER BY created_at DESC",
		)
		if err != nil {
			return err
		}
		defer rows.Close()
		
		fmt.Println("Tasks:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		
		count := 0
		for rows.Next() {
			var taskID, description, status, createdAt string
			var agent, claimedBy sql.NullString
			
			if err := rows.Scan(&taskID, &description, &status, &agent, &createdAt, &claimedBy); err != nil {
				continue
			}
			
			count++
			fmt.Printf("  %s [%s] %s\n", taskID, status, description)
			if agent.Valid {
				fmt.Printf("    Agent: %s", agent.String)
			}
			if claimedBy.Valid {
				fmt.Printf(" | Claimed by: %s", claimedBy.String)
			}
			fmt.Println()
		}
		
		if count == 0 {
			fmt.Println("  No tasks found.")
		}
		
		return nil
	},
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim [task-id]",
	Short: "Claim a task",
	Long:  `Claim a task for processing.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		
		if err := ValidateTaskID(taskID); err != nil {
			return err
		}
		
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)
		
		// Get agent name from env or default
		agentName := os.Getenv("LAZYAI_AGENT")
		if agentName == "" {
			agentName = "unknown"
		}
		
		result, err := database.Exec(
			"UPDATE tasks SET status = 'claimed', claimed_by = ?, claimed_at = ? WHERE task_id = ? AND status = 'pending'",
			agentName, time.Now().UTC().Format(time.RFC3339), taskID,
		)
		if err != nil {
			return fmt.Errorf("error claiming task: %w", err)
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("task %s not found or already claimed", taskID)
		}
		
		fmt.Printf("✅ Task claimed: %s\n", taskID)
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
		taskID := args[0]
		
		if err := ValidateTaskID(taskID); err != nil {
			return err
		}
		
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)
		
		_, err = database.Exec(
			"UPDATE tasks SET status = 'completed', completed_at = ? WHERE task_id = ?",
			time.Now().UTC().Format(time.RFC3339), taskID,
		)
		if err != nil {
			return fmt.Errorf("error completing task: %w", err)
		}
		
		fmt.Printf("✅ Task completed: %s\n", taskID)
		
		return nil
	},
}

func init() {
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskClaimCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	rootCmd.AddCommand(taskCmd)
}
