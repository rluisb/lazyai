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

// ──────────────────────── Compile team plan ───────────────────────────────

func (o *Orchestrator) compileTeamPlan(input types.BuildTeamInput) (*types.ExecutionPlan, error) {
	version, err := o.getActiveOrLatestVersion(string(types.KindTeam), input.Team)
	if err != nil {
		return nil, fmt.Errorf("unknown team definition %q: %w", input.Team, err)
	}

	definition, err := decodeTeamDefinition(version)
	if err != nil {
		return nil, err
	}
	if err := validateTeamDefinition(definition); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	budgetPolicy := buildTeamBudgetPolicy(input.Budget)

	plan := &types.ExecutionPlan{
		ID:   uuid.NewString(),
		Kind: string(types.RunKindTeam),
		Definition: types.DefinitionRef{
			Kind:    string(types.KindTeam),
			Name:    definition.Name,
			Version: definition.Version,
			Source:  definition.Source,
			Path:    definition.Path,
		},
		Cli:           cliContextFromTeam(),
		Project:       types.ProjectStackContext{RootPath: o.projectRoot()},
		BudgetPolicy:  budgetPolicy,
		Entrypoint:    "parallel",
		CompiledSteps: []types.CompiledStepPlan{},
		CreatedAt:     now,
		Task:          input.Task,
	}
	return plan, nil
}

func decodeTeamDefinition(version *domain.CatalogVersion) (*types.TeamDefinition, error) {
	var definition types.TeamDefinition
	if err := json.Unmarshal([]byte(version.Body), &definition); err != nil {
		return nil, fmt.Errorf("active team %s/%s version %d body must be a JSON team definition: %w", version.Kind, version.Name, version.Version, err)
	}
	if definition.Name == "" {
		definition.Name = version.Name
	}
	if definition.Kind == "" {
		definition.Kind = string(types.KindTeam)
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

func validateTeamDefinition(definition *types.TeamDefinition) error {
	if len(definition.Parallel) == 0 {
		return fmt.Errorf("team %q must define at least one parallel member", definition.Name)
	}
	if definition.Synthesize.Agent == "" {
		return fmt.Errorf("team %q must define a synthesize agent", definition.Name)
	}
	return nil
}

func buildTeamBudgetPolicy(overrides *types.BudgetPolicy) types.BudgetPolicy {
	policy := types.BudgetPolicy{
		ID:                   uuid.NewString(),
		Scope:                string(types.RunKindTeam),
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

func cliContextFromTeam() types.CliContext {
	return types.CliContext{
		Host:                     types.HostOpenCode,
		DispatchMode:             types.DispatchTaskTool,
		SupportsSubagents:        true,
		SupportsParallelTeams:    true,
		SupportsStructuredOutput: true,
		MCPServerName:            defaultMCPServerName,
	}
}

// ──────────────────────── Team handoff ─────────────────────────────────────

func createAndSaveTeamHandoff(store ports.HandoffStore, teamState *types.TeamState, plan *types.ExecutionPlan, input handoffInput) (*types.HandoffDocument, string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	teamState.State = types.TeamHandoff
	teamState.UpdatedAt = now

	summary := input.Summary
	if summary == "" {
		summary = fmt.Sprintf("Handoff for team %s", teamState.TeamID)
	}
	doc := &types.HandoffDocument{
		ID:        uuid.NewString(),
		RunID:     teamState.TeamID,
		Kind:      types.RunKindTeam,
		Summary:   summary,
		Recipient: input.Recipient,
		CreatedAt: now,
		Resumable: true,
		Status:    teamState,
		Plan:      plan,
	}
	if err := store.SaveHandoffDocument(doc); err != nil {
		return nil, "", err
	}
	path := handoffPathURIPrefix + doc.ID
	return doc, path, nil
}

// ──────────────────────── Team task status ─────────────────────────────────

func currentTeamTaskStatus(plan *types.ExecutionPlan, teamState *types.TeamState) map[string]any {
	members := make([]map[string]any, 0, len(teamState.Tasks))
	completedCount := 0
	runningCount := 0
	pendingCount := 0
	failedCount := 0

	for _, task := range teamState.Tasks {
		taskMap := map[string]any{
			"taskId": task.TaskID,
			"role":   task.Role,
			"agent":  task.Agent,
			"state":  task.State,
			"kind":   task.Kind,
			"order":  task.Order,
		}
		if task.Assignee != "" {
			taskMap["assignee"] = task.Assignee
		}
		if task.ClaimedBy != "" {
			taskMap["claimedBy"] = task.ClaimedBy
		}
		if task.CompletedAt != "" {
			taskMap["completedAt"] = task.CompletedAt
		}
		if task.Error != nil {
			taskMap["error"] = task.Error
		}
		members = append(members, taskMap)

		switch task.State {
		case types.TaskCompleted:
			completedCount++
		case types.TaskClaimed, types.TaskAssigned:
			runningCount++
		case types.TaskFailed:
			failedCount++
		case types.TaskPending, types.TaskBlocked:
			pendingCount++
		}
	}

	return map[string]any{
		"teamId":           teamState.TeamID,
		"definitionName":   teamState.DefinitionName,
		"state":            teamState.State,
		"task":             teamState.Task,
		"synthesisTaskId":  teamState.SynthesisTaskID,
		"readyTaskIds":     teamState.ReadyTaskIDs,
		"members":          members,
		"completedMembers": completedCount,
		"runningMembers":   runningCount,
		"pendingMembers":   pendingCount,
		"failedMembers":    failedCount,
		"totalMembers":     len(teamState.Tasks),
		"budget":           teamState.Budget,
	}
}

// ──────────────────────── Team lifecycle helpers ───────────────────────────

// startTeam compiles a team plan, creates state, persists it, and publishes an event.
func startTeam(o *Orchestrator, input types.BuildTeamInput) (*types.TeamState, *types.ExecutionPlan, error) {
	plan, err := o.compileTeamPlan(input)
	if err != nil {
		return nil, nil, fmt.Errorf("compile team plan: %w", err)
	}

	definition, err := decodeTeamDefinitionFromPlan(o, plan)
	if err != nil {
		return nil, nil, fmt.Errorf("decode team definition: %w", err)
	}

	teamState := state.CreateTeamState(definition, plan)

	if err := o.ExecutionPlans.SaveExecutionPlan(plan); err != nil {
		return nil, nil, fmt.Errorf("save team execution plan: %w", err)
	}
	if err := o.TeamStates.SaveTeamState(o.projectRoot(), teamState); err != nil {
		return nil, nil, fmt.Errorf("save team state: %w", err)
	}

	o.Events.Publish(teamState.TeamID, "team.started", map[string]any{
		"definitionName": teamState.DefinitionName,
		"task":           teamState.Task,
		"state":          teamState.State,
	})

	return teamState, plan, nil
}

func decodeTeamDefinitionFromPlan(o *Orchestrator, plan *types.ExecutionPlan) (*types.TeamDefinition, error) {
	version, err := o.getActiveOrLatestVersion(plan.Definition.Kind, plan.Definition.Name)
	if err != nil {
		return nil, err
	}
	return decodeTeamDefinition(version)
}
