package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/budget"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// ReadModel provides bounded, read-only dashboard queries over existing SQLite state.
type ReadModel struct {
	database *db.DB
}

// NewReadModel creates a dashboard read model backed by existing orchestrator tables.
func NewReadModel(database *db.DB) *ReadModel {
	return &ReadModel{database: database}
}

// Attention filter values for the Runs screen. "budget" is intentionally
// not handled server-side — it is filtered client-side using the budgetHealth
// summary field already returned by ListRuns.
const (
	AttentionRunning = "running"
	AttentionFailed  = "failed"
	AttentionGated   = "gated"
	AttentionRecent  = "recent"
)

// RunListOptions controls run list filtering and pagination.
type RunListOptions struct {
	Kind      types.RunKind
	State     string
	Search    string // case-insensitive substring against run id and definition_name
	Attention string // running | failed | gated | recent (see Attention* constants)
	HasErrors bool   // when true, only runs with at least one error_journal entry
	Limit     int
	Cursor    string
}

// ErrorListOptions controls error journal filtering and bounds.
type ErrorListOptions struct {
	RunID string
	Limit int
}

// Overview returns health, active counts, recent runs/errors, and catalog counts.
func (m *ReadModel) Overview(ctx context.Context, health HealthView, catalogCounts CatalogCounts) (DashboardOverview, error) {
	activeRuns, err := m.database.ActiveRunCounts()
	if err != nil {
		return DashboardOverview{}, err
	}
	health.ActiveRuns = activeRuns

	runCounts, err := m.runCountsByState(ctx)
	if err != nil {
		return DashboardOverview{}, err
	}
	recentRuns, err := m.ListRuns(ctx, RunListOptions{Limit: DefaultRunLimit})
	if err != nil {
		return DashboardOverview{}, err
	}
	recentErrors, err := m.ListErrors(ctx, ErrorListOptions{Limit: DefaultErrorLimit})
	if err != nil {
		return DashboardOverview{}, err
	}

	return DashboardOverview{
		Health:           health,
		ActiveRuns:       activeRuns,
		RunCountsByState: runCounts,
		RecentRuns:       recentRuns.Items,
		RecentErrors:     recentErrors,
		CatalogCounts:    catalogCounts,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ListRuns returns a bounded page of chain/team/workflow run summaries.
func (m *ReadModel) ListRuns(ctx context.Context, opts RunListOptions) (RunListResponse, error) {
	limit := NormalizeLimit(opts.Limit, DefaultRunLimit, MaxRunLimit)
	offset := parseCursor(opts.Cursor)

	clauses := []string{}
	args := []any{}
	if opts.Kind != "" {
		clauses = append(clauses, "kind = ?")
		args = append(args, string(opts.Kind))
	}
	if opts.State != "" {
		clauses = append(clauses, "state = ?")
		args = append(args, opts.State)
	}
	if opts.Search != "" {
		pattern := "%" + likeEscape(opts.Search) + "%"
		clauses = append(clauses, `(LOWER(id) LIKE LOWER(?) ESCAPE '\' OR LOWER(definition_name) LIKE LOWER(?) ESCAPE '\')`)
		args = append(args, pattern, pattern)
	}
	switch opts.Attention {
	case AttentionRunning:
		clauses = append(clauses, "state = 'running'")
	case AttentionFailed:
		clauses = append(clauses, "state = 'failed'")
	case AttentionGated:
		clauses = append(clauses, "state IN ('gated','paused','awaiting_recovery','waiting_on_child')")
	case AttentionRecent:
		clauses = append(clauses, "updated_at >= datetime('now', '-1 hour')")
	case "":
		// no attention filter
	default:
		return RunListResponse{}, fmt.Errorf("unknown attention filter: %q", opts.Attention)
	}
	if opts.HasErrors {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM error_journal e WHERE e.run_id = runs.id AND e.run_kind = runs.kind)")
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	query := `
		SELECT kind, id, definition_name, definition_version, state, current, project_root, state_json, created_at, updated_at
		FROM (` + runUnionQuery() + `) AS runs
		` + where + `
		ORDER BY updated_at DESC, id DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)
	rows, err := m.database.QueryContext(ctx, query, args...)
	if err != nil {
		return RunListResponse{}, err
	}
	runRows, hasMore, err := scanRunRows(rows, limit)
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return RunListResponse{}, err
	}

	items, err := m.summariesFromRows(ctx, runRows)
	if err != nil {
		return RunListResponse{}, err
	}
	response := RunListResponse{Items: items}
	if hasMore {
		response.NextCursor = strconv.Itoa(offset + limit)
	}
	return response, nil
}

// likeEscape escapes SQL LIKE wildcards so user search terms cannot inject pattern metacharacters.
func likeEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// GetRunDetail returns a read-only detail view for one run.
func (m *ReadModel) GetRunDetail(ctx context.Context, kind types.RunKind, id string) (*RunDetail, error) {
	row, err := m.getRunRow(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	summary, err := m.summaryFromRow(row)
	if err != nil {
		return nil, err
	}

	detail := &RunDetail{Summary: summary}
	var state map[string]any
	if err := json.Unmarshal([]byte(row.stateJSON), &state); err != nil {
		detail.StateDecodeError = err.Error()
	} else {
		detail.State = state
		m.populateTypedState(kind, row.stateJSON, detail)
	}

	budgetView, err := m.budgetFromState(ctx, kind, row.stateJSON)
	if err != nil {
		return nil, err
	}
	detail.Budget = budgetView

	events, err := m.ListEvents(ctx, id, 0, 0)
	if err != nil {
		return nil, err
	}
	detail.Events = events
	errors, err := m.ListErrors(ctx, ErrorListOptions{RunID: id, Limit: MaxErrorLimit})
	if err != nil {
		return nil, err
	}
	detail.Errors = errors
	handoffs, err := m.listHandoffs(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	detail.Handoffs = handoffs
	detail.ExecutionPlan = m.executionPlanFromState(ctx, kind, row.stateJSON)
	return detail, nil
}

// GetBudget returns decoded budget state and evaluation for a run when available.
func (m *ReadModel) GetBudget(ctx context.Context, kind types.RunKind, id string) (*BudgetView, error) {
	row, err := m.getRunRow(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	return m.budgetFromState(ctx, kind, row.stateJSON)
}

// ListEvents returns persisted run events ordered by id ascending.
func (m *ReadModel) ListEvents(ctx context.Context, runID string, sinceID int, limit int) ([]DashboardEvent, error) {
	args := []any{runID}
	where := `run_id = ?`
	if sinceID > 0 {
		where += ` AND id > ?`
		args = append(args, sinceID)
	}
	query := `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE ` + where + ` ORDER BY id ASC`
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := m.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []DashboardEvent
	for rows.Next() {
		var event DashboardEvent
		var dataJSON string
		if err := rows.Scan(&event.ID, &event.RunID, &event.EventType, &dataJSON, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.Data = map[string]any{}
		_ = json.Unmarshal([]byte(dataJSON), &event.Data)
		events = append(events, event)
	}
	return events, rows.Err()
}

// ListErrors returns bounded error journal entries, optionally for one run.
func (m *ReadModel) ListErrors(ctx context.Context, opts ErrorListOptions) ([]ErrorEntry, error) {
	limit := NormalizeLimit(opts.Limit, DefaultErrorLimit, MaxErrorLimit)
	args := []any{}
	where := ""
	if opts.RunID != "" {
		where = `WHERE run_id = ?`
		args = append(args, opts.RunID)
	}
	args = append(args, limit)
	rows, err := m.database.QueryContext(ctx, `
		SELECT id, run_id, run_kind, definition_name, step_id, category, code, message, created_at
		FROM error_journal `+where+`
		ORDER BY created_at DESC, id DESC
		LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ErrorEntry
	for rows.Next() {
		entry, err := scanErrorEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (m *ReadModel) runCountsByState(ctx context.Context) (map[string]int, error) {
	rows, err := m.database.QueryContext(ctx, `
		SELECT state, COUNT(*)
		FROM (`+runUnionQuery()+`)
		GROUP BY state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, err
		}
		counts[state] = count
	}
	return counts, rows.Err()
}

func runUnionQuery() string {
	return `
		SELECT 'chain' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, COALESCE(current_step_id, '') AS current, project_root, state_json, created_at, updated_at FROM chain_runs
		UNION ALL
		SELECT 'team' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, '' AS current, project_root, state_json, created_at, updated_at FROM team_runs
		UNION ALL
		SELECT 'workflow' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, COALESCE(current_phase_id, '') AS current, project_root, state_json, created_at, updated_at FROM workflow_runs
	`
}

type runRow struct {
	kind              types.RunKind
	id                string
	definitionName    string
	definitionVersion string
	state             string
	current           string
	projectRoot       string
	stateJSON         string
	createdAt         string
	updatedAt         string
}

func scanRunRows(rows *sql.Rows, limit int) ([]runRow, bool, error) {
	items := make([]runRow, 0, limit)
	hasMore := false
	for rows.Next() {
		row, err := scanRunRow(rows)
		if err != nil {
			return nil, false, err
		}
		if len(items) < limit {
			items = append(items, row)
		} else {
			hasMore = true
		}
	}
	return items, hasMore, rows.Err()
}

func (m *ReadModel) summariesFromRows(ctx context.Context, rows []runRow) ([]RunSummary, error) {
	errorCounts, err := m.errorCountsByRun(ctx, rows)
	if err != nil {
		return nil, err
	}
	chainPolicies := m.chainBudgetPoliciesForRows(ctx, rows)

	items := make([]RunSummary, 0, len(rows))
	for _, row := range rows {
		summary := summaryFromRowWithLookups(row, errorCounts, chainPolicies)
		items = append(items, summary)
	}
	return items, nil
}

func scanRunRow(scanner interface{ Scan(dest ...any) error }) (runRow, error) {
	var row runRow
	if err := scanner.Scan(&row.kind, &row.id, &row.definitionName, &row.definitionVersion, &row.state, &row.current, &row.projectRoot, &row.stateJSON, &row.createdAt, &row.updatedAt); err != nil {
		return runRow{}, err
	}
	return row, nil
}

func (m *ReadModel) summaryFromRow(row runRow) (RunSummary, error) {
	errorCount, err := m.errorCount(row.kind, row.id)
	if err != nil {
		return RunSummary{}, err
	}
	budgetHealth := ""
	if budgetView, err := m.budgetFromState(context.Background(), row.kind, row.stateJSON); err == nil && budgetView != nil && budgetView.Evaluation != nil {
		budgetHealth = string(budgetView.Evaluation.Overall)
	}
	return runSummary(row, budgetHealth, errorCount), nil
}

type runKey struct {
	kind types.RunKind
	id   string
}

func summaryFromRowWithLookups(row runRow, errorCounts map[runKey]int, chainPolicies map[string]types.BudgetPolicy) RunSummary {
	return runSummary(row, budgetHealthFromState(row, chainPolicies), errorCounts[runKey{kind: row.kind, id: row.id}])
}

func runSummary(row runRow, budgetHealth string, errorCount int) RunSummary {
	return RunSummary{
		Kind:              row.kind,
		ID:                row.id,
		DefinitionName:    row.definitionName,
		DefinitionVersion: row.definitionVersion,
		State:             row.state,
		Current:           row.current,
		ProjectRoot:       row.projectRoot,
		CreatedAt:         row.createdAt,
		UpdatedAt:         row.updatedAt,
		BudgetHealth:      budgetHealth,
		ErrorCount:        errorCount,
	}
}

func (m *ReadModel) errorCount(kind types.RunKind, id string) (int, error) {
	var count int
	err := m.database.QueryRow(`SELECT COUNT(*) FROM error_journal WHERE run_id = ? AND run_kind = ?`, id, string(kind)).Scan(&count)
	return count, err
}

func (m *ReadModel) errorCountsByRun(ctx context.Context, rows []runRow) (map[runKey]int, error) {
	counts := map[runKey]int{}
	if len(rows) == 0 {
		return counts, nil
	}
	runIDs := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		if !seen[row.id] {
			runIDs = append(runIDs, row.id)
			seen[row.id] = true
		}
	}

	query := `SELECT run_id, run_kind, COUNT(*) FROM error_journal WHERE run_id IN (` + placeholders(len(runIDs)) + `) AND run_kind IS NOT NULL GROUP BY run_id, run_kind`
	args := make([]any, 0, len(runIDs))
	for _, id := range runIDs {
		args = append(args, id)
	}
	queryRows, err := m.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var id, kind string
		var count int
		if err := queryRows.Scan(&id, &kind, &count); err != nil {
			return nil, err
		}
		counts[runKey{kind: types.RunKind(kind), id: id}] = count
	}
	return counts, queryRows.Err()
}

func (m *ReadModel) chainBudgetPoliciesForRows(ctx context.Context, rows []runRow) map[string]types.BudgetPolicy {
	planIDs := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		if row.kind != types.RunKindChain {
			continue
		}
		var state types.ChainState
		if json.Unmarshal([]byte(row.stateJSON), &state) != nil || state.ExecutionPlanID == "" || seen[state.ExecutionPlanID] {
			continue
		}
		planIDs = append(planIDs, state.ExecutionPlanID)
		seen[state.ExecutionPlanID] = true
	}
	if len(planIDs) == 0 {
		return map[string]types.BudgetPolicy{}
	}

	query := `SELECT id, plan_json FROM execution_plans WHERE id IN (` + placeholders(len(planIDs)) + `)`
	args := make([]any, 0, len(planIDs))
	for _, id := range planIDs {
		args = append(args, id)
	}
	queryRows, err := m.database.QueryContext(ctx, query, args...)
	if err != nil {
		return map[string]types.BudgetPolicy{}
	}
	defer queryRows.Close()

	policies := map[string]types.BudgetPolicy{}
	for queryRows.Next() {
		var id, planJSON string
		if queryRows.Scan(&id, &planJSON) != nil {
			continue
		}
		var plan types.ExecutionPlan
		if json.Unmarshal([]byte(planJSON), &plan) == nil {
			policies[id] = plan.BudgetPolicy
		}
	}
	return policies
}

func budgetHealthFromState(row runRow, chainPolicies map[string]types.BudgetPolicy) string {
	switch row.kind {
	case types.RunKindChain:
		var state types.ChainState
		if json.Unmarshal([]byte(row.stateJSON), &state) != nil || state.ExecutionPlanID == "" {
			return ""
		}
		policy, ok := chainPolicies[state.ExecutionPlanID]
		if !ok {
			return ""
		}
		evaluation := budget.Evaluate(&state.Budget, &policy)
		return string(evaluation.Overall)
	case types.RunKindTeam:
		var state types.TeamState
		if json.Unmarshal([]byte(row.stateJSON), &state) != nil {
			return ""
		}
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		return string(evaluation.Overall)
	case types.RunKindWorkflow:
		var state types.WorkflowState
		if json.Unmarshal([]byte(row.stateJSON), &state) != nil {
			return ""
		}
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		return string(evaluation.Overall)
	default:
		return ""
	}
}

func placeholders(count int) string {
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func (m *ReadModel) getRunRow(ctx context.Context, kind types.RunKind, id string) (runRow, error) {
	table, currentColumn, err := runTable(kind)
	if err != nil {
		return runRow{}, err
	}
	query := fmt.Sprintf(`SELECT '%s' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, %s AS current, project_root, state_json, created_at, updated_at FROM %s WHERE id = ?`, kind, currentColumn, table)
	row, err := scanRunRow(m.database.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return runRow{}, NewNotFoundError("run", string(kind)+"/"+id)
		}
		return runRow{}, err
	}
	return row, nil
}

func runTable(kind types.RunKind) (table string, currentColumn string, err error) {
	switch kind {
	case types.RunKindChain:
		return "chain_runs", "COALESCE(current_step_id, '')", nil
	case types.RunKindTeam:
		return "team_runs", "''", nil
	case types.RunKindWorkflow:
		return "workflow_runs", "COALESCE(current_phase_id, '')", nil
	default:
		return "", "", NewNotFoundError("run kind", string(kind))
	}
}

func (m *ReadModel) populateTypedState(kind types.RunKind, stateJSON string, detail *RunDetail) {
	switch kind {
	case types.RunKindChain:
		var state types.ChainState
		if json.Unmarshal([]byte(stateJSON), &state) == nil {
			detail.Steps = state.Steps
		}
	case types.RunKindTeam:
		var state types.TeamState
		if json.Unmarshal([]byte(stateJSON), &state) == nil {
			detail.Tasks = state.Tasks
		}
	case types.RunKindWorkflow:
		var state types.WorkflowState
		if json.Unmarshal([]byte(stateJSON), &state) == nil {
			detail.Phases = state.Phases
		}
	}
}

func (m *ReadModel) budgetFromState(ctx context.Context, kind types.RunKind, stateJSON string) (*BudgetView, error) {
	switch kind {
	case types.RunKindChain:
		var state types.ChainState
		if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
			return &BudgetView{DecodeError: err.Error()}, nil
		}
		view := budgetViewFromState(&state.Budget)
		if policy := m.chainBudgetPolicy(ctx, state.ExecutionPlanID); policy != nil {
			evaluation := budget.Evaluate(&state.Budget, policy)
			view.Evaluation = &evaluation
		}
		return view, nil
	case types.RunKindTeam:
		var state types.TeamState
		if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
			return &BudgetView{DecodeError: err.Error()}, nil
		}
		view := budgetViewFromState(&state.Budget)
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		view.Evaluation = &evaluation
		return view, nil
	case types.RunKindWorkflow:
		var state types.WorkflowState
		if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
			return &BudgetView{DecodeError: err.Error()}, nil
		}
		view := budgetViewFromState(&state.Budget)
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		view.Evaluation = &evaluation
		return view, nil
	default:
		return nil, NewNotFoundError("run kind", string(kind))
	}
}

func budgetViewFromState(state *types.BudgetState) *BudgetView {
	return &BudgetView{
		State:         state,
		ByStep:        state.ByStep,
		LastUpdatedAt: state.LastUpdatedAt,
	}
}

func (m *ReadModel) chainBudgetPolicy(ctx context.Context, executionPlanID string) *types.BudgetPolicy {
	if executionPlanID == "" {
		return nil
	}
	plan := m.loadExecutionPlan(ctx, executionPlanID)
	if plan == nil {
		return nil
	}
	policy := plan.BudgetPolicy
	return &policy
}

func (m *ReadModel) executionPlanFromState(ctx context.Context, kind types.RunKind, stateJSON string) map[string]any {
	if kind != types.RunKindChain {
		return nil
	}
	var state types.ChainState
	if json.Unmarshal([]byte(stateJSON), &state) != nil || state.ExecutionPlanID == "" {
		return nil
	}
	plan := m.loadExecutionPlanMap(ctx, state.ExecutionPlanID)
	return plan
}

func (m *ReadModel) loadExecutionPlan(ctx context.Context, id string) *types.ExecutionPlan {
	var planJSON string
	if err := m.database.QueryRowContext(ctx, `SELECT plan_json FROM execution_plans WHERE id = ?`, id).Scan(&planJSON); err != nil {
		return nil
	}
	var plan types.ExecutionPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil
	}
	return &plan
}

func (m *ReadModel) loadExecutionPlanMap(ctx context.Context, id string) map[string]any {
	var planJSON string
	if err := m.database.QueryRowContext(ctx, `SELECT plan_json FROM execution_plans WHERE id = ?`, id).Scan(&planJSON); err != nil {
		return nil
	}
	var plan map[string]any
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil
	}
	return plan
}

func (m *ReadModel) listHandoffs(ctx context.Context, kind types.RunKind, id string) ([]map[string]any, error) {
	rows, err := m.database.QueryContext(ctx, `SELECT doc_json FROM handoffs WHERE run_id = ? AND run_kind = ? ORDER BY created_at DESC`, id, string(kind))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []map[string]any
	for rows.Next() {
		var docJSON string
		if err := rows.Scan(&docJSON); err != nil {
			return nil, err
		}
		var doc map[string]any
		if json.Unmarshal([]byte(docJSON), &doc) == nil {
			docs = append(docs, doc)
		}
	}
	return docs, rows.Err()
}

func scanErrorEntry(scanner interface{ Scan(dest ...any) error }) (ErrorEntry, error) {
	var entry ErrorEntry
	var runID, runKind, stepID sql.NullString
	if err := scanner.Scan(&entry.ID, &runID, &runKind, &entry.DefinitionName, &stepID, &entry.Category, &entry.Code, &entry.Message, &entry.CreatedAt); err != nil {
		return ErrorEntry{}, err
	}
	if runID.Valid {
		entry.RunID = runID.String
	}
	if runKind.Valid {
		entry.RunKind = types.RunKind(runKind.String)
	}
	if stepID.Valid {
		entry.StepID = stepID.String
	}
	return entry, nil
}

func parseCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	offset, err := strconv.Atoi(cursor)
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}
