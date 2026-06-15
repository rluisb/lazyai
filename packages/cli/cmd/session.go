package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/handoff"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
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
		agentName = "primary-agent"
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

	handoffPath, err := writeSessionHandoff(db, mgr, sessionID)
	if err != nil {
		return fmt.Errorf("write session handoff: %w", err)
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
	fmt.Printf("   Handoff: %s\n", handoffPath)
	return nil
}

func writeSessionHandoff(db *runtime.DB, mgr *session.Manager, sessionID string) (string, error) {
	s, err := mgr.Get(sessionID)
	if err != nil {
		return "", fmt.Errorf("load session: %w", err)
	}

	dispatches, err := mgr.ListDispatches(sessionID)
	if err != nil {
		return "", fmt.Errorf("list dispatches: %w", err)
	}

	existingPath, err := existingHandoffPath(db, sessionID)
	if err != nil {
		return "", fmt.Errorf("load existing handoff metadata: %w", err)
	}

	path := existingPath
	if path == "" {
		path = defaultHandoffPath(s)
		if path == "" {
			path = filepath.Join("specs", "memory", "handoffs", fmt.Sprintf("%s-session-%s.md", time.Now().UTC().Format("2006-01-02"), strings.TrimPrefix(s.ID, "ses_")))
		}
		if files.FileExists(path) {
			path = filepath.Join("specs", "memory", "handoffs", fmt.Sprintf("%s-%s-%s.md", time.Now().UTC().Format("2006-01-02"), handoffTopicSlug(s.Goal), strings.TrimPrefix(s.ID, "ses_")))
		}
	}

	doc := buildSessionHandoffDocument(s, dispatches)
	if err := handoff.Write(path, doc); err != nil {
		return "", err
	}
	if err := upsertHandoffMetadata(db, sessionID, path, s.Goal, string(doc.Progress), runtime.Now()); err != nil {
		return "", err
	}

	return path, nil
}

func defaultHandoffPath(s *session.Session) string {
	datePart := time.Now().UTC().Format("2006-01-02")
	if s.EndedAt != nil {
		datePart = s.EndedAt.UTC().Format("2006-01-02")
	}

	topic := handoffTopicSlug(s.Goal)
	if topic == "" {
		return ""
	}

	return filepath.Join("specs", "memory", "handoffs", fmt.Sprintf("%s-%s.md", datePart, topic))
}

func buildSessionHandoffDocument(s *session.Session, dispatches []session.Dispatch) handoff.Document {
	items := buildProgressItems(s, dispatches)
	lastDispatch := lastDispatch(dispatches)

	doc := handoff.Document{
		Goal:            s.Goal,
		Constraints:     buildConstraints(s),
		Progress:        deriveHandoffProgress(items),
		Decisions:       buildDecisions(s, dispatches),
		CriticalContext: buildCriticalContext(s, lastDispatch),
		NextSteps:       buildNextSteps(s, items, lastDispatch),
		Risks:           buildRisks(items, lastDispatch),
		Owner:           s.Agent,
		SessionID:       s.ID,
		OpenQuestions:   buildOpenQuestions(s, dispatches),
		Items:           items,
	}

	return doc
}

func buildConstraints(s *session.Session) []string {
	var constraints []string
	if s.Repo != "" {
		constraints = append(constraints, "Repo: "+s.Repo)
	}
	if s.Worktree != "" {
		constraints = append(constraints, "Worktree: "+s.Worktree)
	}
	if s.Agent != "" {
		constraints = append(constraints, "Agent: "+s.Agent)
	}
	if s.Model != "" {
		constraints = append(constraints, "Model: "+s.Model)
	}
	if len(s.Tags) > 0 {
		constraints = append(constraints, "Tags: "+strings.Join(s.Tags, ", "))
	}
	return constraints
}

func buildProgressItems(s *session.Session, dispatches []session.Dispatch) handoff.ProgressItems {
	var items handoff.ProgressItems
	for _, dispatch := range dispatches {
		label := dispatchLabel(dispatch)
		switch {
		case dispatch.ErrorMessage != "":
			items.Pending = append(items.Pending, fmt.Sprintf("%s — failed: %s", label, trimForHandoff(dispatch.ErrorMessage)))
		case dispatch.Result != "" || dispatch.EndedAt != nil:
			items.Done = append(items.Done, fmt.Sprintf("%s — completed", label))
		default:
			items.InProgress = append(items.InProgress, fmt.Sprintf("%s — still open", label))
		}
	}

	if len(items.Done) == 0 && len(items.InProgress) == 0 && len(items.Pending) == 0 {
		if s.Status == session.SessionEnded {
			items.Pending = append(items.Pending, "No dispatches were recorded before the session closed.")
		} else {
			items.InProgress = append(items.InProgress, "Session work is still in progress.")
		}
	}

	return items
}

func buildDecisions(s *session.Session, dispatches []session.Dispatch) []string {
	seen := make(map[string]bool)
	var decisions []string
	appendDecision := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		decisions = append(decisions, value)
	}

	appendDecision(s.Summary)
	for _, dispatch := range dispatches {
		appendDecision(dispatch.Summary)
	}
	return decisions
}

func buildCriticalContext(s *session.Session, last *session.Dispatch) string {
	parts := []string{fmt.Sprintf("Session %s closed with status %s.", s.ID, s.Status)}
	if s.Summary != "" {
		parts = append(parts, "Summary: "+trimForHandoff(s.Summary))
	}
	if last != nil {
		switch {
		case last.ErrorMessage != "":
			parts = append(parts, fmt.Sprintf("Latest dispatch %s failed: %s", dispatchLabel(*last), trimForHandoff(last.ErrorMessage)))
		case last.Result != "":
			parts = append(parts, fmt.Sprintf("Latest dispatch %s result: %s", dispatchLabel(*last), trimForHandoff(last.Result)))
		default:
			parts = append(parts, fmt.Sprintf("Latest dispatch %s has no recorded result yet.", dispatchLabel(*last)))
		}
	}
	return strings.Join(parts, " ")
}

func buildNextSteps(s *session.Session, items handoff.ProgressItems, last *session.Dispatch) []string {
	var steps []string
	if len(items.Pending) > 0 {
		steps = append(steps, "Investigate the pending or failed dispatches recorded in this handoff.")
	}
	if last != nil && last.Result != "" {
		steps = append(steps, "Resume from the latest completed dispatch result before starting new work.")
	}
	if s.Summary == "" {
		steps = append(steps, "Add a concise session summary before making the next round of changes.")
	}
	if len(steps) == 0 {
		steps = append(steps, "Review the recorded session context and continue from the latest verified state.")
	}
	if len(steps) > 2 {
		steps = steps[:2]
	}
	return steps
}

func buildRisks(items handoff.ProgressItems, last *session.Dispatch) []string {
	var risks []string
	if len(items.Pending) > 0 {
		risks = append(risks, "Session closed with pending or failed dispatch work.")
	}
	if last != nil && last.ErrorMessage != "" {
		risks = append(risks, "Latest dispatch error: "+trimForHandoff(last.ErrorMessage))
	}
	return risks
}

func buildOpenQuestions(s *session.Session, dispatches []session.Dispatch) []string {
	if s.Summary == "" && len(dispatches) > 0 {
		return []string{"What concise decision summary should be recorded for this session before the next handoff?"}
	}
	return nil
}

func deriveHandoffProgress(items handoff.ProgressItems) handoff.ProgressStatus {
	switch {
	case len(items.Pending) > 0:
		return handoff.ProgressPending
	case len(items.InProgress) > 0:
		return handoff.ProgressInProgress
	default:
		return handoff.ProgressDone
	}
}

func lastDispatch(dispatches []session.Dispatch) *session.Dispatch {
	if len(dispatches) == 0 {
		return nil
	}
	last := dispatches[len(dispatches)-1]
	return &last
}

func dispatchLabel(dispatch session.Dispatch) string {
	task := strings.TrimSpace(dispatch.Task)
	if task == "" {
		task = fmt.Sprintf("dispatch %d", dispatch.Seq)
	}
	if dispatch.Agent == "" {
		return task
	}
	return fmt.Sprintf("%s (%s)", task, dispatch.Agent)
}

func trimForHandoff(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	if len(value) <= 180 {
		return value
	}
	return value[:177] + "..."
}

func handoffTopicSlug(value string) string {
	var b strings.Builder
	lastDash := true
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func existingHandoffPath(db *runtime.DB, sessionID string) (string, error) {
	var path string
	err := db.QueryRow("SELECT path FROM handoff WHERE session_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", sessionID).Scan(&path)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query handoff path: %w", err)
	}
	return path, nil
}

func upsertHandoffMetadata(db *runtime.DB, sessionID, path, goal, status, createdAt string) error {
	return db.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec("DELETE FROM handoff WHERE session_id = ?", sessionID); err != nil {
			return fmt.Errorf("delete existing handoff metadata: %w", err)
		}
		if _, err := tx.Exec(
			"INSERT INTO handoff (session_id, path, goal, status, created_at) VALUES (?, ?, ?, ?, ?)",
			sessionID, path, goal, status, createdAt,
		); err != nil {
			return fmt.Errorf("insert handoff metadata: %w", err)
		}
		return nil
	})
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
