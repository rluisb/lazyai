package state

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestCreateTeamState(t *testing.T) {
	def := &types.TeamDefinition{
		DefinitionMetadata: types.DefinitionMetadata{Name: "test-team", Version: "1.0"},
		Parallel: []types.TeamMemberDefinition{
			{Role: "researcher", Agent: "claude", Skills: []string{"research"}, Focus: "gather info"},
			{Role: "coder", Agent: "claude", Skills: []string{"implement"}, Focus: "write code"},
		},
		Synthesize: types.TeamSynthesizeDefinition{Agent: "claude-sonnet"},
	}
	plan := &types.ExecutionPlan{
		ID:        "plan-1",
		Task:      "build something",
		BudgetPolicy: types.BudgetPolicy{ID: "policy-1"},
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	state := CreateTeamState(def, plan)

	// Should have 3 tasks: 2 members + 1 synthesis
	if len(state.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(state.Tasks))
	}

	// First two should be member tasks, pending
	for i := 0; i < 2; i++ {
		if state.Tasks[i].Kind != "member" {
			t.Errorf("task %d: expected kind=member, got %s", i, state.Tasks[i].Kind)
		}
		if state.Tasks[i].State != types.TaskPending {
			t.Errorf("task %d: expected state=pending, got %s", i, state.Tasks[i].State)
		}
	}

	// Last should be synthesize task, blocked
	synthTask := state.Tasks[2]
	if synthTask.Kind != "synthesize" {
		t.Errorf("task 2: expected kind=synthesize, got %s", synthTask.Kind)
	}
	if synthTask.State != types.TaskBlocked {
		t.Errorf("task 2: expected state=blocked, got %s", synthTask.State)
	}

	// ReadyTaskIDs should contain the two member task IDs
	if len(state.ReadyTaskIDs) != 2 {
		t.Fatalf("expected 2 ready task IDs, got %d", len(state.ReadyTaskIDs))
	}

	// Team state should be running
	if state.State != types.TeamRunning {
		t.Errorf("expected team state=running, got %s", state.State)
	}
}

func TestAssignTeamTask(t *testing.T) {
	state := &types.TeamState{
		Tasks: []types.TeamTaskState{
			{TaskID: "task-1", Kind: "member", State: types.TaskPending, Role: "researcher"},
			{TaskID: "task-2", Kind: "member", State: types.TaskClaimed, Role: "coder"},
		},
	}

	// Assign pending task without claim
	next, err := AssignTeamTask(state, "task-1", "worker-a", false)
	if err != nil {
		t.Fatalf("assign task-1: %v", err)
	}
	if next.Tasks[0].State != types.TaskAssigned {
		t.Errorf("task-1: expected state=assigned, got %s", next.Tasks[0].State)
	}
	if next.Tasks[0].Assignee != "worker-a" {
		t.Errorf("task-1: expected assignee=worker-a, got %s", next.Tasks[0].Assignee)
	}

	// Claim an assigned task (must assign first)
	assigned, err := AssignTeamTask(state, "task-1", "worker-a", false)
	if err != nil {
		t.Fatalf("assign task-1: %v", err)
	}
	next, err = AssignTeamTask(assigned, "task-1", "worker-a", true)
	if err != nil {
		t.Fatalf("claim task-1: %v", err)
	}
	if err != nil {
		t.Fatalf("claim task-1: %v", err)
	}
	if next.Tasks[0].State != types.TaskClaimed {
		t.Errorf("task-1: expected state=claimed, got %s", next.Tasks[0].State)
	}
	if next.Tasks[0].ClaimedBy != "worker-a" {
		t.Errorf("task-1: expected claimedBy=worker-a, got %s", next.Tasks[0].ClaimedBy)
	}

	// Cannot claim a pending task directly
	_, err = AssignTeamTask(state, "task-1", "worker-b", true)
	if err == nil {
		t.Error("expected error when claiming pending task, got nil")
	}

	// Cannot assign a claimed task
	_, err = AssignTeamTask(state, "task-2", "worker-b", false)
	if err == nil {
		t.Error("expected error when assigning claimed task, got nil")
	}
}

func TestCompleteTeamTask(t *testing.T) {
	state := &types.TeamState{
		State: types.TeamRunning,
		Tasks: []types.TeamTaskState{
			{TaskID: "member-0", Kind: "member", State: types.TaskClaimed, Role: "researcher"},
			{TaskID: "member-1", Kind: "member", State: types.TaskClaimed, Role: "coder"},
			{TaskID: "synthesize", Kind: "synthesize", State: types.TaskBlocked, DependsOn: []string{"member-0", "member-1"}},
		},
		ReadyTaskIDs: []string{"member-0", "member-1"},
		SynthesisTaskID: "synthesize",
	}

	// Complete first member
	next, err := CompleteTeamTask(state, "member-0", "success", map[string]any{"output": "report"}, nil, nil)
	if err != nil {
		t.Fatalf("complete member-0: %v", err)
	}
	if next.Tasks[0].State != types.TaskCompleted {
		t.Errorf("member-0: expected state=completed, got %s", next.Tasks[0].State)
	}
	// Team still running since second member not done
	if next.State != types.TeamRunning {
		t.Errorf("expected team state=running after one member done, got %s", next.State)
	}

	// Complete second member — synthesis should unlock
	next, err = CompleteTeamTask(next, "member-1", "success", map[string]any{"output": "code"}, nil, nil)
	if err != nil {
		t.Fatalf("complete member-1: %v", err)
	}
	if next.State != types.TeamSynthesizing {
		t.Errorf("expected team state=synthesizing after all members done, got %s", next.State)
	}
	synthTask := next.Tasks[2]
	if synthTask.State != types.TaskPending {
		t.Errorf("synthesis task: expected state=pending, got %s", synthTask.State)
	}

	// Synthesis task is now pending — must assign and claim before completing
	synthAssigned, err := AssignTeamTask(next, "synthesize", "orchestrator", false)
	if err != nil {
		t.Fatalf("assign synthesize: %v", err)
	}
	synthClaimed, err := AssignTeamTask(synthAssigned, "synthesize", "orchestrator", true)
	if err != nil {
		t.Fatalf("claim synthesize: %v", err)
	}
	next = synthClaimed

	// Complete synthesis — team should complete
	next, err = CompleteTeamTask(next, "synthesize", "success", map[string]any{"summary": "final"}, nil, nil)
	if err != nil {
		t.Fatalf("complete synthesize: %v", err)
	}
	if next.State != types.TeamCompleted {
		t.Errorf("expected team state=completed, got %s", next.State)
	}
}

func TestCompleteTeamTaskFailure(t *testing.T) {
	state := &types.TeamState{
		State: types.TeamRunning,
		Tasks: []types.TeamTaskState{
			{TaskID: "member-0", Kind: "member", State: types.TaskClaimed, Role: "researcher"},
		},
		ReadyTaskIDs: []string{"member-0"},
	}

	// Fail a member task
	structErr := &types.StructuredError{
		Category: types.ErrorTransient,
		Code:     "timeout",
		Message:  "step timed out",
		StepID:   "member-0",
	}
	next, err := CompleteTeamTask(state, "member-0", "failure", nil, nil, structErr)
	if err != nil {
		t.Fatalf("fail member-0: %v", err)
	}
	if next.State != types.TeamFailed {
		t.Errorf("expected team state=failed after member failure, got %s", next.State)
	}
	if next.Tasks[0].State != types.TaskFailed {
		t.Errorf("member-0: expected state=failed, got %s", next.Tasks[0].State)
	}
	if next.Tasks[0].Error == nil {
		t.Error("member-0: expected error to be set")
	}
}

func TestRetryTeamTask(t *testing.T) {
	state := &types.TeamState{
		State: types.TeamFailed,
		Tasks: []types.TeamTaskState{
			{TaskID: "member-0", Kind: "member", State: types.TaskFailed, Role: "researcher", Error: &types.StructuredError{Code: "oops"}},
		},
	}

	// Retry failed task
	next, err := RetryTeamTask(state, "member-0")
	if err != nil {
		t.Fatalf("retry member-0: %v", err)
	}
	if next.Tasks[0].State != types.TaskPending {
		t.Errorf("member-0: expected state=pending after retry, got %s", next.Tasks[0].State)
	}
	if next.Tasks[0].Error != nil {
		t.Error("member-0: expected error to be cleared after retry")
	}
	if next.State != types.TeamRunning {
		t.Errorf("expected team state=running after retry, got %s", next.State)
	}
	if len(next.ReadyTaskIDs) == 0 || next.ReadyTaskIDs[0] != "member-0" {
		t.Errorf("member-0 should be in readyTaskIDs after retry")
	}

	// Cannot retry a non-failed task
	_, err = RetryTeamTask(next, "member-0")
	if err == nil {
		t.Error("expected error when retrying non-failed task")
	}
}

func TestEscalateTeamTask(t *testing.T) {
	state := &types.TeamState{
		State: types.TeamFailed,
		Tasks: []types.TeamTaskState{
			{TaskID: "member-0", Kind: "member", State: types.TaskFailed, Agent: "claude-haiku"},
		},
	}

	// Escalate failed task to a stronger agent
	next, err := EscalateTeamTask(state, "member-0", "claude-sonnet")
	if err != nil {
		t.Fatalf("escalate member-0: %v", err)
	}
	if next.Tasks[0].Agent != "claude-sonnet" {
		t.Errorf("member-0: expected agent=claude-sonnet, got %s", next.Tasks[0].Agent)
	}
	if next.Tasks[0].State != types.TaskPending {
		t.Errorf("member-0: expected state=pending after escalation, got %s", next.Tasks[0].State)
	}
	if next.State != types.TeamRunning {
		t.Errorf("expected team state=running after escalation, got %s", next.State)
	}

	// Cannot escalate a completed task
	_, err = EscalateTeamTask(next, "member-0", "claude-opus")
	if err == nil {
		t.Error("expected error when escalating completed task")
	}
}