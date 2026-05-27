package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime/session"
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

// getDB opens the legacy ai-setup database for backward compatibility.
// New runtime commands should prefer openRuntimeDB().
func getDB() (*db.DB, error) {
	dir, _ := os.Getwd()
	dbPath := db.DefaultDBPath(dir)
	if !files.FileExists(dbPath) {
		return nil, fmt.Errorf("database not found at %s. Run 'lazyai-cli init' first", dbPath)
	}
	return db.Open(dbPath)
}

func runSessionList(cmd *cobra.Command, args []string) error {
	db, err := openRuntimeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	mgr := session.NewManager(db)
	sessions, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	fmt.Println("Sessions:")
	fmt.Println("---------")
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}
	for _, s := range sessions {
		statusEmoji := "🟢"
		if s.Status == session.SessionEnded {
			statusEmoji = "🔴"
		}
		startedAt := s.StartedAt.Format(time.RFC3339)
		fmt.Printf("%s %s | %s | %s\n", statusEmoji, s.ID, s.Goal, startedAt)
	}
	return nil
}

func runSessionStart(cmd *cobra.Command, args []string) error {
	db, err := openRuntimeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	goal := args[0]
	agentName := os.Getenv("LAZYAI_AGENT")
	if agentName == "" {
		agentName = "loop-driver"
	}

	mgr := session.NewManager(db)
	s, err := mgr.Start(goal, session.StartOptions{
		Agent: agentName,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Append to ledger (best-effort)
	_ = appendToLedger("session_start", map[string]string{
		"session_id": s.ID,
		"goal":       goal,
		"status":     "active",
	})

	fmt.Printf("✅ Session started: %s\n", s.ID)
	fmt.Printf("   Goal: %s\n", goal)
	fmt.Printf("   Started: %s\n", s.StartedAt.Format(time.RFC3339))
	return nil
}

func runSessionEnd(cmd *cobra.Command, args []string) error {
	db, err := openRuntimeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	sessionID := args[0]

	mgr := session.NewManager(db)
	if err := mgr.End(sessionID); err != nil {
		return err
	}

	// Append to ledger (best-effort)
	_ = appendToLedger("session_end", map[string]string{
		"session_id": sessionID,
		"status":     "ended",
	})

	// Record quality metric (best-effort)
	_ = recordQualityMetric(sessionID, "session_duration", "session", "")

	fmt.Printf("✅ Session ended: %s\n", sessionID)
	fmt.Printf("   Ended: %s\n", time.Now().UTC().Format(time.RFC3339))
	return nil
}

func runSessionShow(cmd *cobra.Command, args []string) error {
	db, err := openRuntimeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	sessionID := args[0]

	mgr := session.NewManager(db)
	s, err := mgr.Get(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	fmt.Printf("Session: %s\n", s.ID)
	fmt.Printf("Goal: %s\n", s.Goal)
	fmt.Printf("Status: %s\n", s.Status)
	fmt.Printf("Started: %s\n", s.StartedAt.Format(time.RFC3339))
	if s.EndedAt != nil {
		fmt.Printf("Ended: %s\n", s.EndedAt.Format(time.RFC3339))
	}

	// Show dispatches
	dispatches, err := mgr.ListDispatches(sessionID)
	if err != nil {
		return err
	}

	fmt.Println("\nDispatches:")
	if len(dispatches) == 0 {
		fmt.Println("  No dispatches")
		return nil
	}
	for _, d := range dispatches {
		status := "pending"
		if d.Result != "" {
			status = "completed"
		} else if d.ErrorMessage != "" {
			status = "failed"
		}
		startedAt := ""
		if d.StartedAt != nil {
			startedAt = d.StartedAt.Format(time.RFC3339)
		}
		fmt.Printf("  [%s] %s: %s\n", status, d.Agent, d.Task)
		if startedAt != "" {
			fmt.Printf("    Dispatched: %s\n", startedAt)
		}
	}

	return nil
}
