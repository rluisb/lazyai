package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage AI agent sessions",
	Long:  "Start, list, and manage AI agent sessions with tracking and metrics.",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE:  runSessionList,
}

var sessionStartCmd = &cobra.Command{
	Use:   "start [goal]",
	Short: "Start a new session",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSessionStart,
}

var sessionEndCmd = &cobra.Command{
	Use:   "end [session-id]",
	Short: "End a session",
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionEnd,
}

var sessionShowCmd = &cobra.Command{
	Use:   "show [session-id]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionShow,
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionStartCmd)
	sessionCmd.AddCommand(sessionEndCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.GroupID = "runtime"
}

func getDB() (*db.DB, error) {
	dir, _ := os.Getwd()
	dbPath := db.DefaultDBPath(dir)
	if !files.FileExists(dbPath) {
		return nil, fmt.Errorf("database not found at %s. Run 'lazyai-cli init' first", dbPath)
	}
	return db.Open(dbPath)
}

func runSessionList(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		return err
	}

	rows, err := database.Query("SELECT id, goal, status, started_at, ended_at FROM sessions ORDER BY started_at DESC")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Sessions:")
	fmt.Println("---------")
	count := 0
	for rows.Next() {
		var id, goal, status, startedAt string
		var endedAt *string
		if err := rows.Scan(&id, &goal, &status, &startedAt, &endedAt); err != nil {
			continue
		}
		count++
		statusEmoji := "🟢"
		if status == "ended" {
			statusEmoji = "🔴"
		}
		fmt.Printf("%s %s | %s | %s\n", statusEmoji, id, goal, startedAt)
	}
	if count == 0 {
		fmt.Println("No sessions found")
	}
	return nil
}

func runSessionStart(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		return err
	}

	goal := args[0]
	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	startedAt := time.Now().UTC().Format(time.RFC3339)

	_, err = database.Exec(
		"INSERT INTO sessions (id, goal, status, started_at) VALUES (?, ?, ?, ?)",
		sessionID, goal, "active", startedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Append to ledger
	_ = appendToLedger("session_start", map[string]string{
		"session_id": sessionID,
		"goal":       goal,
		"status":     "active",
	})

	fmt.Printf("✅ Session started: %s\n", sessionID)
	fmt.Printf("   Goal: %s\n", goal)
	fmt.Printf("   Started: %s\n", startedAt)
	return nil
}

func runSessionEnd(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	sessionID := args[0]
	endedAt := time.Now().UTC().Format(time.RFC3339)

	result, err := database.Exec(
		"UPDATE sessions SET status = ?, ended_at = ? WHERE id = ?",
		"ended", endedAt, sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Append to ledger
	_ = appendToLedger("session_end", map[string]string{
		"session_id": sessionID,
		"status":     "ended",
	})

	// Record quality metric
	_ = recordQualityMetric(sessionID, "session_duration", "session", "")

	fmt.Printf("✅ Session ended: %s\n", sessionID)
	fmt.Printf("   Ended: %s\n", endedAt)
	return nil
}

func runSessionShow(cmd *cobra.Command, args []string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	sessionID := args[0]

	var goal, status, startedAt string
	var endedAt *string
	err = database.QueryRow(
		"SELECT goal, status, started_at, ended_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&goal, &status, &startedAt, &endedAt)
	if err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	fmt.Printf("Session: %s\n", sessionID)
	fmt.Printf("Goal: %s\n", goal)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Started: %s\n", startedAt)
	if endedAt != nil {
		fmt.Printf("Ended: %s\n", *endedAt)
	}

	// Show dispatches
	rows, err := database.Query(
		"SELECT agent, task, status, dispatched_at FROM agent_dispatches WHERE session_id = ? ORDER BY seq",
		sessionID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("\nDispatches:")
	dispatchCount := 0
	for rows.Next() {
		var agent, task, dispatchStatus, dispatchedAt string
		if err := rows.Scan(&agent, &task, &dispatchStatus, &dispatchedAt); err != nil {
			continue
		}
		dispatchCount++
		fmt.Printf("  [%s] %s: %s\n", dispatchStatus, agent, task)
	}
	if dispatchCount == 0 {
		fmt.Println("  No dispatches")
	}

	return nil
}
