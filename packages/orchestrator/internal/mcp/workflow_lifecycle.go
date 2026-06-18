package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/state"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

// ──────────────────────── Compile workflow plan ───────────────────────────

func (o *Orchestrator) compileWorkflowPlan(input types.StartWorkflowInput) (*types.ExecutionPlan, error) {
	version, err := o.getActiveOrLatestVersion(string(types.KindWorkflow), input.Workflow)
	if err != nil {
		return nil, fmt.Errorf("unknown workflow definition %q: %w", input.Workflow, err)
	}

	definition, err := decodeWorkflowDefinition(version)
	if err != nil {
		return nil, err
	}
	if err := validateWorkflowDefinition(definition); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	budgetPolicy := buildWorkflowBudgetPolicy(input.Budget)

	plan := &types.ExecutionPlan{
		ID:   uuid.NewString(),
		Kind: string(types.RunKindWorkflow),
		Definition: types.DefinitionRef{
			Kind:    string(types.KindWorkflow),
			Name:    definition.Name,
			Version: definition.Version,
			Source:  definition.Source,
			Path:    definition.Path,
		},
		Cli:           cliContextFromWorkflow(input),
		Project:       types.ProjectStackContext{RootPath: o.projectRoot()},
		BudgetPolicy:  budgetPolicy,
		Entrypoint:    definition.Entry,
		CompiledSteps: []types.CompiledStepPlan{},
		CreatedAt:     now,
		Task:          input.Task,
	}
	if input.Context != nil {
		plan.RootContext = &input.Context.RootContext
	}
	return plan, nil
}

func decodeWorkflowDefinition(version *domain.CatalogVersion) (*types.WorkflowDefinition, error) {
	var definition types.WorkflowDefinition
	if err := json.Unmarshal([]byte(version.Body), &definition); err != nil {
		return nil, fmt.Errorf("active workflow %s/%s version %d body must be a JSON workflow definition: %w", version.Kind, version.Name, version.Version, err)
	}
	if definition.Name == "" {
		definition.Name = version.Name
	}
	if definition.Kind == "" {
		definition.Kind = string(types.KindWorkflow)
	}
	if definition.Version == "" {
		definition.Version = fmt.Sprintf("%d", version.Version)
	}
	if definition.Source == "" {
		definition.Source = types.SourceDB
	}
	if definition.Path == "" {
		definition.Path = fmt.Sprintf("catalog://%s/%s/%d", version.Kind, version.Name, version.Version)
	}
	return &definition, nil
}

func validateWorkflowDefinition(definition *types.WorkflowDefinition) error {
	if definition.Entry == "" {
		return fmt.Errorf("workflow %q must define an entry phase", definition.Name)
	}
	if len(definition.Phases) == 0 {
		return fmt.Errorf("workflow %q must define at least one phase", definition.Name)
	}
	phaseIDs := map[string]bool{}
	for _, phase := range definition.Phases {
		if phase.ID == "" {
			return fmt.Errorf("workflow %q contains a phase without an id", definition.Name)
		}
		phaseIDs[phase.ID] = true
	}
	if !phaseIDs[definition.Entry] {
		return fmt.Errorf("workflow entry phase %q does not exist", definition.Entry)
	}
	return nil
}

func buildWorkflowBudgetPolicy(overrides *types.BudgetPolicy) types.BudgetPolicy {
	policy := types.BudgetPolicy{
		ID:                   uuid.NewString(),
		Scope:                string(types.RunKindWorkflow),
		DefaultActionOnLimit: "pause",
	}
	if overrides == nil {
		return policy
	}
	if overrides.ID != "" {
		policy.ID = overrides.ID
	}
	if overrides.Scope != "" {
		policy.Scope = overrides.Scope
	}
	if overrides.DefaultActionOnLimit != "" {
		policy.DefaultActionOnLimit = overrides.DefaultActionOnLimit
	}
	policy.Tokens = overrides.Tokens
	policy.CostUsd = overrides.CostUsd
	policy.WallClockMs = overrides.WallClockMs
	policy.Retries = overrides.Retries
	return policy
}

func cliContextFromWorkflow(input types.StartWorkflowInput) types.CliContext {
	host := types.HostOpenCode
	if input.Context != nil && input.Context.CliTool != "" {
		host = input.Context.CliTool
	}
	ctx := types.CliContext{Host: host, MCPServerName: defaultMCPServerName}
	switch host {
	case types.HostClaudeCode:
		ctx.DispatchMode = types.DispatchTaskTool
		ctx.SupportsSubagents = true
		ctx.SupportsParallelTeams = true
		ctx.SupportsStructuredOutput = true
	case types.HostCopilot:
		ctx.DispatchMode = types.DispatchInstructionOnly
		ctx.SupportsSubagents = false
		ctx.SupportsParallelTeams = false
		ctx.SupportsStructuredOutput = false
	default:
		ctx.DispatchMode = types.DispatchTaskTool
		ctx.SupportsSubagents = true
		ctx.SupportsParallelTeams = false
		ctx.SupportsStructuredOutput = true
	}
	return ctx
}

// ──────────────────────── Workflow handoff ────────────────────────────────

func createAndSaveWorkflowHandoff(store ports.HandoffStore, workflowState *types.WorkflowState, plan *types.ExecutionPlan, input handoffInput) (*types.HandoffDocument, string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	workflowState.State = types.WorkflowHandoff
	workflowState.UpdatedAt = now

	summary := input.Summary
	if summary == "" {
		summary = fmt.Sprintf("Handoff for workflow %s", workflowState.WorkflowID)
	}
	doc := &types.HandoffDocument{
		ID:        uuid.NewString(),
		RunID:     workflowState.WorkflowID,
		Kind:      types.RunKindWorkflow,
		Summary:   summary,
		Recipient: input.Recipient,
		CreatedAt: now,
		Resumable: true,
		Status:    workflowState,
		Plan:      plan,
	}
	if err := store.SaveHandoffDocument(doc); err != nil {
		return nil, "", err
	}
	path := handoffPathURIPrefix + doc.ID
	return doc, path, nil
}

// ──────────────────────── Workflow phase status ───────────────────────────

func currentWorkflowPhaseStatus(plan *types.ExecutionPlan, workflowState *types.WorkflowState) map[string]any {
	phases := make([]map[string]any, 0, len(workflowState.Phases))
	completedCount := 0
	runningCount := 0
	waitingCount := 0
	failedCount := 0

	for _, phase := range workflowState.Phases {
		phaseMap := map[string]any{
			"phaseId": phase.PhaseID,
			"kind":    phase.Kind,
			"state":   phase.State,
			"ref":     phase.Ref,
		}
		if phase.StartedAt != "" {
			phaseMap["startedAt"] = phase.StartedAt
		}
		if phase.CompletedAt != "" {
			phaseMap["completedAt"] = phase.CompletedAt
		}
		if phase.LastOutcome != "" {
			phaseMap["lastOutcome"] = phase.LastOutcome
		}
		if phase.ChildRun != nil {
			phaseMap["childRun"] = phase.ChildRun
		}
		phases = append(phases, phaseMap)

		switch phase.State {
		case types.PhaseCompleted:
			completedCount++
		case types.PhaseRunning:
			runningCount++
		case types.PhaseWaitingOnChild:
			waitingCount++
		case types.PhaseFailed:
			failedCount++
		}
	}

	return map[string]any{
		"workflowId":      workflowState.WorkflowID,
		"definitionName":  workflowState.DefinitionName,
		"state":           workflowState.State,
		"task":            workflowState.Task,
		"entryPhaseId":    workflowState.EntryPhaseID,
		"currentPhaseId":  workflowState.CurrentPhaseID,
		"phases":          phases,
		"completedPhases": completedCount,
		"runningPhases":   runningCount,
		"waitingPhases":   waitingCount,
		"failedPhases":    failedCount,
		"totalPhases":     len(workflowState.Phases),
		"budget":          workflowState.Budget,
		"childRuns":       workflowState.ChildRuns,
	}
}

// ──────────────────────── Workflow lifecycle helpers ───────────────────────

// startWorkflow compiles a workflow plan, creates state, persists it, and publishes an event.
func startWorkflow(o *Orchestrator, input types.StartWorkflowInput) (*types.WorkflowState, *types.ExecutionPlan, error) {
	plan, err := o.compileWorkflowPlan(input)
	if err != nil {
		return nil, nil, fmt.Errorf("compile workflow plan: %w", err)
	}

	definition, err := decodeWorkflowDefinitionFromPlan(o, plan)
	if err != nil {
		return nil, nil, fmt.Errorf("decode workflow definition: %w", err)
	}

	workflowState := state.CreateWorkflowState(definition, plan)

	if err := o.ExecutionPlans.SaveExecutionPlan(plan); err != nil {
		return nil, nil, fmt.Errorf("save workflow execution plan: %w", err)
	}
	if err := o.WorkflowStates.SaveWorkflowState(o.projectRoot(), workflowState); err != nil {
		return nil, nil, fmt.Errorf("save workflow state: %w", err)
	}

	o.Events.Publish(workflowState.WorkflowID, "workflow.started", map[string]any{
		"definitionName": workflowState.DefinitionName,
		"task":           workflowState.Task,
		"currentPhaseId": workflowState.CurrentPhaseID,
		"state":          workflowState.State,
	})

	return workflowState, plan, nil
}

func decodeWorkflowDefinitionFromPlan(o *Orchestrator, plan *types.ExecutionPlan) (*types.WorkflowDefinition, error) {
	version, err := o.getActiveOrLatestVersion(plan.Definition.Kind, plan.Definition.Name)
	if err != nil {
		return nil, err
	}
	return decodeWorkflowDefinition(version)
}
