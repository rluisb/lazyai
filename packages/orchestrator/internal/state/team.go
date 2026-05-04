package state

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// CreateTeamState builds initial TeamState from a team definition.
func CreateTeamState(def *types.TeamDefinition, plan *types.ExecutionPlan) *types.TeamState {
	now := time.Now().UTC().Format(time.RFC3339)
	tasks := make([]types.TeamTaskState, 0, len(def.Parallel)+1)

	for i, m := range def.Parallel {
		tasks = append(tasks, types.TeamTaskState{
			TaskID: fmt.Sprintf("%s-%d", m.Role, i),
			Kind:   "member",
			Role:   m.Role,
			Agent:  m.Agent,
			Skills: m.Skills,
			Focus:  m.Focus,
			State:  types.TaskPending,
			Order:  i,
			Usage:  types.StepUsage{},
		})
	}

	synthID := "synthesize"
	tasks = append(tasks, types.TeamTaskState{
		TaskID:    synthID,
		Kind:      "synthesize",
		Role:      "synthesizer",
		Agent:     def.Synthesize.Agent,
		State:     types.TaskBlocked,
		Order:     len(def.Parallel),
		DependsOn: memberTaskIDs(tasks),
		Usage:     types.StepUsage{},
	})

	readyIDs := make([]string, 0, len(def.Parallel))
	for _, t := range tasks {
		if t.Kind == "member" {
			readyIDs = append(readyIDs, t.TaskID)
		}
	}

	return &types.TeamState{
		TeamID:            uuid.NewString(),
		DefinitionName:    def.Name,
		DefinitionVersion: def.Version,
		State:             types.TeamRunning,
		Task:              plan.Task,
		Tasks:             tasks,
		ReadyTaskIDs:      readyIDs,
		SynthesisTaskID:   synthID,
		BudgetPolicy:      plan.BudgetPolicy,
		Budget:            createEmptyBudgetForTeam(plan),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// AssignTeamTask assigns or claims a task to an assignee.
func AssignTeamTask(state *types.TeamState, taskID, assignee string, claim bool) (*types.TeamState, error) {
	next := cloneTeamState(state)
	task := requireTeamTask(next, taskID)

	if task.State != types.TaskPending && task.State != types.TaskAssigned {
		return nil, fmt.Errorf("task %q is not available for assignment", taskID)
	}

	if claim && task.State == types.TaskPending {
		return nil, fmt.Errorf("cannot claim unassigned task %q", taskID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if claim {
		task.State = types.TaskClaimed
		task.ClaimedBy = assignee
		task.ClaimedAt = now
	} else {
		task.State = types.TaskAssigned
		task.Assignee = assignee
		task.AssignedAt = now
	}
	next.UpdatedAt = now
	return next, nil
}

// CompleteTeamTask completes a team task with success or failure.
func CompleteTeamTask(state *types.TeamState, taskID, outcome string, result map[string]any, usage *types.StepUsage, err *types.StructuredError) (*types.TeamState, error) {
	next := cloneTeamState(state)
	task := requireTeamTask(next, taskID)

	if task.State != types.TaskClaimed {
		return nil, fmt.Errorf("task %q must be claimed before completion", taskID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if outcome == "success" {
		task.State = types.TaskCompleted
		task.CompletedAt = now
		if result != nil {
			task.Result = result
		}
		if usage != nil {
			task.Usage = *usage
		}
	} else {
		task.State = types.TaskFailed
		task.Error = err
	}
	next.UpdatedAt = now

	// Check if all members are done; unlock synthesis
	if allMembersDone(next.Tasks) {
		for i := range next.Tasks {
			if next.Tasks[i].Kind == "synthesize" {
				next.Tasks[i].State = types.TaskPending
				next.ReadyTaskIDs = append(next.ReadyTaskIDs, next.Tasks[i].TaskID)
				next.State = types.TeamSynthesizing
				break
			}
		}
	}

	// Check for any failed member tasks
	for _, t := range next.Tasks {
		if t.State == types.TaskFailed {
			next.State = types.TeamFailed
			return next, nil
		}
	}

	// Check if synthesis completed
	if task.Kind == "synthesize" && outcome == "success" {
		next.State = types.TeamCompleted
		next.Summary = result
	}

	return next, nil
}

// RetryTeamTask resets a failed team task for retry.
func RetryTeamTask(state *types.TeamState, taskID string) (*types.TeamState, error) {
	next := cloneTeamState(state)
	task := requireTeamTask(next, taskID)
	if task.State != types.TaskFailed {
		return nil, fmt.Errorf("task %q is not eligible for retry", taskID)
	}
	task.State = types.TaskPending
	task.Error = nil
	task.CompletedAt = ""
	next.State = types.TeamRunning
	next.ReadyTaskIDs = append(next.ReadyTaskIDs, taskID)
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return next, nil
}

// EscalateTeamTask escalates a failed task to a target agent.
func EscalateTeamTask(state *types.TeamState, taskID, targetAgent string) (*types.TeamState, error) {
	next := cloneTeamState(state)
	task := requireTeamTask(next, taskID)
	if task.State != types.TaskFailed && task.State != types.TaskBlocked {
		return nil, fmt.Errorf("task %q is not eligible for escalation", taskID)
	}
	task.Agent = targetAgent
	task.State = types.TaskPending
	task.Error = nil
	task.CompletedAt = ""
	next.State = types.TeamRunning
	next.ReadyTaskIDs = append(next.ReadyTaskIDs, taskID)
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return next, nil
}

// ──────────────────────── helpers ────────────────────────────

func memberTaskIDs(tasks []types.TeamTaskState) []string {
	ids := make([]string, 0)
	for _, t := range tasks {
		if t.Kind == "member" {
			ids = append(ids, t.TaskID)
		}
	}
	return ids
}

func allMembersDone(tasks []types.TeamTaskState) bool {
	for _, t := range tasks {
		if t.Kind == "member" && t.State != types.TaskCompleted && t.State != types.TaskFailed {
			return false
		}
	}
	return true
}

func requireTeamTask(state *types.TeamState, taskID string) *types.TeamTaskState {
	for i := range state.Tasks {
		if state.Tasks[i].TaskID == taskID {
			return &state.Tasks[i]
		}
	}
	panic(fmt.Sprintf("unknown team task: %s", taskID))
}

func createEmptyBudgetForTeam(plan *types.ExecutionPlan) types.BudgetState {
	return types.BudgetState{
		PolicyID:      plan.BudgetPolicy.ID,
		Scope:         "team",
		Tokens:        buildBudgetDim(plan.BudgetPolicy.Tokens),
		CostUsd:       buildBudgetDim(plan.BudgetPolicy.CostUsd),
		WallClockMs:   buildBudgetDim(plan.BudgetPolicy.WallClockMs),
		Retries:       buildBudgetDim(plan.BudgetPolicy.Retries),
		ByStep:        map[string]types.StepUsage{},
		LastUpdatedAt: plan.CreatedAt,
	}
}

func buildBudgetDim(threshold *types.BudgetThreshold) types.BudgetDimensionState {
	d := types.BudgetDimensionState{}
	if threshold != nil {
		d.Limit = threshold.Limit
		d.Remaining = threshold.Limit
	}
	return d
}

func cloneTeamState(s *types.TeamState) *types.TeamState {
	b, _ := json.Marshal(s)
	var c types.TeamState
	json.Unmarshal(b, &c)
	return &c
}
