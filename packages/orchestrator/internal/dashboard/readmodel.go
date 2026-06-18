package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/budget"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

// ReadModel provides bounded, read-only dashboard queries over existing SQLite state.
type ReadModel struct {
	runStore           ports.RunReadStore
	activityStore      ports.ActivityStore
	handoffStore       ports.HandoffQueryStore
	eventStore         ports.RunEventStore
	executionPlanStore ports.ExecutionPlanStore
	errorJournalStore  ports.ErrorJournalStore
}

// NewReadModel creates a dashboard read model backed by existing orchestrator tables.
func NewReadModel(runStore ports.RunReadStore, activityStore ports.ActivityStore, handoffStore ports.HandoffQueryStore, eventStore ports.RunEventStore, executionPlanStore ports.ExecutionPlanStore, errorJournalStore ports.ErrorJournalStore) *ReadModel {
	return &ReadModel{runStore: runStore, activityStore: activityStore, handoffStore: handoffStore, eventStore: eventStore, executionPlanStore: executionPlanStore, errorJournalStore: errorJournalStore}
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
	activeRuns, err := m.activityStore.ActiveRunCounts(ctx)
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
	page, err := m.runStore.ListRuns(ctx, domain.RunListFilter{
		Kind:      opts.Kind,
		State:     opts.State,
		Search:    opts.Search,
		Attention: opts.Attention,
		HasErrors: opts.HasErrors,
		Limit:     opts.Limit,
		Cursor:    opts.Cursor,
	})
	if err != nil {
		return RunListResponse{}, err
	}

	items, err := m.summariesFromRows(ctx, page.Items)
	if err != nil {
		return RunListResponse{}, err
	}
	return RunListResponse{Items: items, NextCursor: page.NextCursor}, nil
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
	if err := json.Unmarshal([]byte(row.StateJSON), &state); err != nil {
		detail.StateDecodeError = err.Error()
	} else {
		detail.State = state
		m.populateTypedState(kind, row.StateJSON, detail)
	}

	budgetView, err := m.budgetFromState(ctx, kind, row.StateJSON)
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
	detail.ExecutionPlan = m.executionPlanFromState(ctx, kind, row.StateJSON)
	return detail, nil
}

// GetBudget returns decoded budget state and evaluation for a run when available.
func (m *ReadModel) GetBudget(ctx context.Context, kind types.RunKind, id string) (*BudgetView, error) {
	row, err := m.getRunRow(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	return m.budgetFromState(ctx, kind, row.StateJSON)
}

// ListEvents returns persisted run events ordered by id ascending.
func (m *ReadModel) ListEvents(ctx context.Context, runID string, sinceID int, limit int) ([]DashboardEvent, error) {
	if m.eventStore == nil {
		return nil, nil
	}
	events, err := m.eventStore.ReplayRunEvents(runID, sinceID)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	return busEventsToDashboard(events), nil
}

// ListErrors returns bounded error journal entries, optionally for one run.
func (m *ReadModel) ListErrors(ctx context.Context, opts ErrorListOptions) ([]ErrorEntry, error) {
	if m.errorJournalStore == nil {
		return nil, nil
	}
	limit := NormalizeLimit(opts.Limit, DefaultErrorLimit, MaxErrorLimit)
	entries, err := m.errorJournalStore.ListErrorJournalEntries(ctx, opts.RunID, limit)
	if err != nil {
		return nil, err
	}
	return errorEntriesFromDomain(entries), nil
}

func (m *ReadModel) runCountsByState(ctx context.Context) (map[string]int, error) {
	return m.runStore.CountRunsByState(ctx)
}

func (m *ReadModel) summariesFromRows(ctx context.Context, rows []domain.RunRow) ([]RunSummary, error) {
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

func (m *ReadModel) summaryFromRow(row domain.RunRow) (RunSummary, error) {
	errorCount, err := m.errorCount(row.Kind, row.ID)
	if err != nil {
		return RunSummary{}, err
	}
	budgetHealth := ""
	if budgetView, err := m.budgetFromState(context.Background(), row.Kind, row.StateJSON); err == nil && budgetView != nil && budgetView.Evaluation != nil {
		budgetHealth = string(budgetView.Evaluation.Overall)
	}
	return runSummary(row, budgetHealth, errorCount), nil
}

type runKey struct {
	kind types.RunKind
	id   string
}

func summaryFromRowWithLookups(row domain.RunRow, errorCounts map[runKey]int, chainPolicies map[string]types.BudgetPolicy) RunSummary {
	return runSummary(row, budgetHealthFromState(row, chainPolicies), errorCounts[runKey{kind: row.Kind, id: row.ID}])
}

func runSummary(row domain.RunRow, budgetHealth string, errorCount int) RunSummary {
	return RunSummary{
		Kind:              row.Kind,
		ID:                row.ID,
		DefinitionName:    row.DefinitionName,
		DefinitionVersion: row.DefinitionVersion,
		State:             row.State,
		Current:           row.Current,
		ProjectRoot:       row.ProjectRoot,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		BudgetHealth:      budgetHealth,
		ErrorCount:        errorCount,
	}
}

func (m *ReadModel) errorCount(kind types.RunKind, id string) (int, error) {
	if m.errorJournalStore == nil {
		return 0, nil
	}
	return m.errorJournalStore.CountErrorJournalEntry(context.Background(), kind, id)
}

func (m *ReadModel) errorCountsByRun(ctx context.Context, rows []domain.RunRow) (map[runKey]int, error) {
	counts := map[runKey]int{}
	if len(rows) == 0 || m.errorJournalStore == nil {
		return counts, nil
	}
	refs := make([]domain.RunRef, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		key := string(row.Kind) + ":" + row.ID
		if !seen[key] {
			refs = append(refs, domain.RunRef{Kind: string(row.Kind), ID: row.ID})
			seen[key] = true
		}
	}
	domainCounts, err := m.errorJournalStore.CountErrorJournalEntriesByRun(ctx, refs)
	if err != nil {
		return nil, err
	}
	for ref, count := range domainCounts {
		counts[runKey{kind: types.RunKind(ref.Kind), id: ref.ID}] = count
	}
	return counts, nil
}

func (m *ReadModel) chainBudgetPoliciesForRows(ctx context.Context, rows []domain.RunRow) map[string]types.BudgetPolicy {
	planIDs := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Kind != types.RunKindChain {
			continue
		}
		var state types.ChainState
		if json.Unmarshal([]byte(row.StateJSON), &state) != nil || state.ExecutionPlanID == "" || seen[state.ExecutionPlanID] {
			continue
		}
		planIDs = append(planIDs, state.ExecutionPlanID)
		seen[state.ExecutionPlanID] = true
	}
	if len(planIDs) == 0 || m.executionPlanStore == nil {
		return map[string]types.BudgetPolicy{}
	}

	policies := map[string]types.BudgetPolicy{}
	for _, id := range planIDs {
		plan, err := m.executionPlanStore.LoadExecutionPlan(id)
		if err != nil || plan == nil {
			continue
		}
		policies[id] = plan.BudgetPolicy
	}
	return policies
}

func budgetHealthFromState(row domain.RunRow, chainPolicies map[string]types.BudgetPolicy) string {
	switch row.Kind {
	case types.RunKindChain:
		var state types.ChainState
		if json.Unmarshal([]byte(row.StateJSON), &state) != nil || state.ExecutionPlanID == "" {
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
		if json.Unmarshal([]byte(row.StateJSON), &state) != nil {
			return ""
		}
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		return string(evaluation.Overall)
	case types.RunKindWorkflow:
		var state types.WorkflowState
		if json.Unmarshal([]byte(row.StateJSON), &state) != nil {
			return ""
		}
		evaluation := budget.Evaluate(&state.Budget, &state.BudgetPolicy)
		return string(evaluation.Overall)
	default:
		return ""
	}
}

func (m *ReadModel) getRunRow(ctx context.Context, kind types.RunKind, id string) (domain.RunRow, error) {
	row, err := m.runStore.FindRunRow(ctx, kind, id)
	if err != nil {
		var notFound domain.RunReadNotFoundError
		if errors.As(err, &notFound) {
			return domain.RunRow{}, NewNotFoundError(notFound.Resource, notFound.ID)
		}
		return domain.RunRow{}, err
	}
	return row, nil
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
	if m.executionPlanStore == nil {
		return nil
	}
	plan, err := m.executionPlanStore.LoadExecutionPlan(id)
	if err != nil {
		return nil
	}
	return plan
}

func (m *ReadModel) loadExecutionPlanMap(ctx context.Context, id string) map[string]any {
	plan := m.loadExecutionPlan(ctx, id)
	if plan == nil {
		return nil
	}
	encoded, err := json.Marshal(plan)
	if err != nil {
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil
	}
	return decoded
}

func (m *ReadModel) listHandoffs(ctx context.Context, kind types.RunKind, id string) ([]map[string]any, error) {
	if m.handoffStore == nil {
		return nil, nil
	}
	docs, err := m.handoffStore.ListHandoffDocuments(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	handoffs := make([]map[string]any, 0, len(docs))
	for _, doc := range docs {
		encoded, err := json.Marshal(doc)
		if err != nil {
			continue
		}
		var handoff map[string]any
		if json.Unmarshal(encoded, &handoff) == nil {
			handoffs = append(handoffs, handoff)
		}
	}
	return handoffs, nil
}

func errorEntriesFromDomain(entries []domain.ErrorJournalEntry) []ErrorEntry {
	converted := make([]ErrorEntry, 0, len(entries))
	for _, entry := range entries {
		converted = append(converted, ErrorEntry{
			ID:             entry.ID,
			RunID:          entry.RunID,
			RunKind:        types.RunKind(entry.RunKind),
			DefinitionName: entry.DefinitionName,
			StepID:         entry.StepID,
			Category:       entry.Category,
			Code:           entry.Code,
			Message:        entry.Message,
			CreatedAt:      entry.CreatedAt,
		})
	}
	return converted
}
