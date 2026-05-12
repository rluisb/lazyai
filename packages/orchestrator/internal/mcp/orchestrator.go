package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/budget"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/dispatch"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/events"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/queue"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/state"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

type Orchestrator struct {
	DB         *db.DB
	Catalog    *catalog.Store
	Events     *events.Bus
	Queue      *queue.Queue
	Scope      *ScopeContext
	Runs       map[string]*RunContext
	Runtime    RuntimeConfig
	Dispatcher dispatch.Dispatcher
}

type RunContext struct {
	Kind       types.RunKind
	State      any
	AdvanceCh  chan string
	CancelFunc context.CancelFunc
}

func NewOrchestrator(database *db.DB, scope *ScopeContext, options ...OrchestratorOption) *Orchestrator {
	runtime := DefaultRuntimeConfig()
	o := &Orchestrator{
		DB:         database,
		Catalog:    catalog.NewStore(database),
		Events:     events.NewBus(database),
		Queue:      queue.New(database),
		Scope:      scope,
		Runs:       make(map[string]*RunContext),
		Runtime:    runtime,
		Dispatcher: defaultDispatcherFor(runtime),
	}
	for _, option := range options {
		if option != nil {
			option(o)
		}
	}
	return o
}

func (o *Orchestrator) RegisterTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("list_catalog",
		mcp.WithDescription("List orchestration catalog definitions."),
		mcp.WithString("kinds", mcp.Description("Filter by kinds (e.g. chain, team, workflow)")),
	), o.ListCatalog)
	s.AddTool(mcp.NewTool("compose_agent",
		mcp.WithDescription("Compose a runtime agent prompt."),
		mcp.WithString("base", mcp.Required(), mcp.Description("Base agent name")),
		mcp.WithString("domainSkill", mcp.Description("Optional domain skill")),
		mcp.WithString("modeSkill", mcp.Description("Optional mode skill")),
	), o.ComposeAgent)
	s.AddTool(mcp.NewTool("start_chain",
		mcp.WithDescription("Compile and start a chain execution plan."),
		mcp.WithString("chain", mcp.Required(), mcp.Description("Chain definition name")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description for the chain")),
		mcp.WithString("domainSkill", mcp.Description("Optional domain skill")),
		mcp.WithString("modeSkill", mcp.Description("Optional mode skill")),
		mcp.WithString("context", mcp.Description("Optional execution context JSON")),
	), o.StartChain)
	s.AddTool(mcp.NewTool("advance_chain",
		mcp.WithDescription("Advance a running chain."),
		mcp.WithString("chainId", mcp.Required(), mcp.Description("Chain run ID")),
		mcp.WithString("stepId", mcp.Required(), mcp.Description("Current step ID")),
		mcp.WithString("outcome", mcp.Required(), mcp.Description("Step outcome (success/failed/skipped)")),
		mcp.WithObject("output", mcp.Description("Step output data (object; stringified JSON accepted for legacy clients).")),
		mcp.WithObject("usage", mcp.Description("Token/cost usage (object; stringified JSON accepted for legacy clients).")),
	), o.AdvanceChain)
	s.AddTool(mcp.NewTool("build_team",
		mcp.WithDescription("Compile and start a team run."),
		mcp.WithString("team", mcp.Required(), mcp.Description("Team definition name")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description for the team")),
		mcp.WithObject("budget", mcp.Description("Optional budget policy (object; stringified JSON accepted for legacy clients).")),
	), o.BuildTeam)
	s.AddTool(mcp.NewTool("assign_team_task",
		mcp.WithDescription("Assign or claim a team task."),
		mcp.WithString("teamId", mcp.Required(), mcp.Description("Team run ID")),
		mcp.WithString("taskId", mcp.Required(), mcp.Description("Task ID to assign")),
		mcp.WithString("assignee", mcp.Description("Assignee agent name")),
		mcp.WithBoolean("claim", mcp.Description("Claim the task if true")),
	), o.AssignTeamTask)
	s.AddTool(mcp.NewTool("complete_team_task",
		mcp.WithDescription("Complete a team task."),
		mcp.WithString("teamId", mcp.Required(), mcp.Description("Team run ID")),
		mcp.WithString("taskId", mcp.Required(), mcp.Description("Task ID to complete")),
		mcp.WithString("outcome", mcp.Required(), mcp.Description("Task outcome (success/failed)")),
		mcp.WithObject("result", mcp.Description("Task result data (object; stringified JSON accepted for legacy clients).")),
		mcp.WithObject("usage", mcp.Description("Token/cost usage (object; stringified JSON accepted for legacy clients).")),
		mcp.WithObject("error", mcp.Description("Error details if failed (object; stringified JSON accepted for legacy clients).")),
	), o.CompleteTeamTask)
	s.AddTool(mcp.NewTool("start_workflow",
		mcp.WithDescription("Compile and start a workflow run."),
		mcp.WithString("workflow", mcp.Required(), mcp.Description("Workflow definition name")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description for the workflow")),
		mcp.WithString("domainSkill", mcp.Description("Optional domain skill")),
		mcp.WithString("modeSkill", mcp.Description("Optional mode skill")),
		mcp.WithObject("budget", mcp.Description("Optional budget policy (object; stringified JSON accepted for legacy clients).")),
		mcp.WithObject("context", mcp.Description("Optional execution context (object; stringified JSON accepted for legacy clients).")),
	), o.StartWorkflow)
	s.AddTool(mcp.NewTool("advance_workflow",
		mcp.WithDescription("Advance a running workflow."),
		mcp.WithString("workflowId", mcp.Required(), mcp.Description("Workflow run ID")),
		mcp.WithString("outcome", mcp.Description("Phase outcome")),
		mcp.WithObject("recovery", mcp.Description("Recovery decision (object; stringified JSON accepted for legacy clients).")),
	), o.AdvanceWorkflow)
	s.AddTool(mcp.NewTool("get_status",
		mcp.WithDescription("Get runtime status for a run."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID (chainId/teamId/workflowId)")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Run kind (chain/team/workflow)")),
	), o.GetStatus)
	s.AddTool(mcp.NewTool("get_budget",
		mcp.WithDescription("Get tracked budget state."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Run kind (chain/team/workflow)")),
	), o.GetBudget)
	s.AddTool(mcp.NewTool("retry_step",
		mcp.WithDescription("Retry a failed step."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Run kind (chain/team/workflow)")),
		mcp.WithString("stepId", mcp.Required(), mcp.Description("Step/task ID to retry")),
		mcp.WithString("reason", mcp.Description("Reason for retry")),
	), o.RetryStep)
	s.AddTool(mcp.NewTool("escalate_step",
		mcp.WithDescription("Escalate a step to a different agent."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Run kind (chain/team/workflow)")),
		mcp.WithString("stepId", mcp.Required(), mcp.Description("Step/task ID to escalate")),
		mcp.WithString("targetAgent", mcp.Required(), mcp.Description("Target agent name")),
		mcp.WithString("domainSkill", mcp.Description("Optional domain skill")),
		mcp.WithString("modeSkill", mcp.Description("Optional mode skill")),
		mcp.WithString("reason", mcp.Description("Reason for escalation")),
	), o.EscalateStep)
	s.AddTool(mcp.NewTool("handoff",
		mcp.WithDescription("Persist a resumable handoff document."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID")),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Run kind (chain/team/workflow)")),
		mcp.WithString("summary", mcp.Description("Handoff summary")),
		mcp.WithString("recipient", mcp.Description("Target recipient agent")),
		mcp.WithBoolean("includeArtifacts", mcp.Description("Include artifacts in handoff")),
	), o.Handoff)
	s.AddTool(mcp.NewTool("catalog_list",
		mcp.WithDescription("List versioned catalog definitions."),
		mcp.WithString("kind", mcp.Description("Filter by definition kind")),
	), o.CatalogList)
	s.AddTool(mcp.NewTool("catalog_get_version",
		mcp.WithDescription("Get catalog definition version."),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Definition kind")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Definition name")),
	), o.CatalogGetVersion)
	s.AddTool(mcp.NewTool("catalog_create_version",
		mcp.WithDescription("Create immutable version."),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Definition kind")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Definition name")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Definition body (JSON/YAML)")),
	), o.CatalogCreateVersion)
	s.AddTool(mcp.NewTool("catalog_set_active",
		mcp.WithDescription("Move active version pointer."),
		mcp.WithString("kind", mcp.Required(), mcp.Description("Definition kind")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Definition name")),
		mcp.WithInteger("version", mcp.Description("Version number to activate")),
	), o.CatalogSetActive)
	s.AddTool(mcp.NewTool("subscribe_run",
		mcp.WithDescription("Subscribe to run events."),
		mcp.WithString("runId", mcp.Required(), mcp.Description("Run ID to subscribe to")),
	), o.SubscribeRun)
	s.AddTool(mcp.NewTool("invoke_agent",
		mcp.WithDescription("Resolve and compose agent invocation."),
		mcp.WithString("agent", mcp.Required(), mcp.Description("Agent name")),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task description")),
		mcp.WithString("cliTool", mcp.Description("Optional CLI tool override")),
	), o.InvokeAgent)
	s.AddTool(mcp.NewTool("enqueue_job",
		mcp.WithDescription("Enqueue background job."),
		mcp.WithString("jobType", mcp.Required(), mcp.Description("Job type identifier")),
		mcp.WithObject("payload", mcp.Description("Job payload (object; stringified JSON accepted for legacy clients).")),
	), o.EnqueueJob)
	s.AddTool(mcp.NewTool("get_job",
		mcp.WithDescription("Get job status."),
		mcp.WithString("jobId", mcp.Required(), mcp.Description("Job ID")),
	), o.GetJob)
	s.AddTool(mcp.NewTool("list_jobs",
		mcp.WithDescription("List queued jobs."),
		mcp.WithString("status", mcp.Description("Filter by status (pending/claimed/completed/failed)")),
	), o.ListJobs)
}

// ───────────────── Tool implementations ─────────────────

func (o *Orchestrator) ListCatalog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	items, err := o.listCatalogItems(req)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(map[string]any{"items": items}), nil
}

func (o *Orchestrator) ComposeAgent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	base := req.GetString("base", "")
	if base == "" {
		return text("Missing required: base"), nil
	}
	spec := &types.ComposedAgentSpec{
		ID: base, Base: base, Model: "sonnet", Prompt: fmt.Sprintf("You are the %s agent.", base),
	}
	return jsonOut(spec), nil
}

func (o *Orchestrator) StartChain(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		Chain       string              `json:"chain"`
		Task        string              `json:"task"`
		DomainSkill string              `json:"domainSkill,omitempty"`
		ModeSkill   string              `json:"modeSkill,omitempty"`
		Budget      *types.BudgetPolicy `json:"budget,omitempty"`
		Context     json.RawMessage     `json:"context,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid start_chain input: %v", err)), nil
	}
	input := types.StartChainInput{
		Chain:       raw.Chain,
		Task:        raw.Task,
		DomainSkill: raw.DomainSkill,
		ModeSkill:   raw.ModeSkill,
		Budget:      raw.Budget,
	}
	if err := decodeOptionalJSONArg(raw.Context, &input.Context); err != nil {
		return text(fmt.Sprintf("Invalid start_chain context: %v", err)), nil
	}
	if input.Chain == "" || input.Task == "" {
		return text("Missing required: chain, task"), nil
	}

	plan, err := o.compileChainPlan(input)
	if err != nil {
		toolLog.Error("tool failed", "tool", "start_chain", "chain", input.Chain, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	chainState := state.CreateChainState(plan)

	if err := saveExecutionPlan(o.DB, plan); err != nil {
		toolLog.Error("tool failed", "tool", "start_chain", "chain", input.Chain, "chainId", chainState.ChainID, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveChainState(o.DB, o.projectRoot(), chainState); err != nil {
		toolLog.Error("tool failed", "tool", "start_chain", "chain", input.Chain, "chainId", chainState.ChainID, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	o.Events.Publish(chainState.ChainID, "chain.started", map[string]any{
		"definitionName": chainState.DefinitionName,
		"task":           chainState.Task,
		"currentStepId":  chainState.CurrentStepID,
		"state":          chainState.State,
	})
	toolLog.Info("tool completed", "tool", "start_chain", "chain", input.Chain, "chainId", chainState.ChainID, "currentStepId", chainState.CurrentStepID)

	return jsonOut(map[string]any{
		"chainId":         chainState.ChainID,
		"state":           chainState.State,
		"currentStep":     currentChainStepStatus(plan, chainState),
		"budget":          chainState.Budget,
		"executionPlanId": plan.ID,
	}), nil
}

func (o *Orchestrator) AdvanceChain(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		ChainID string          `json:"chainId"`
		StepID  string          `json:"stepId"`
		Outcome string          `json:"outcome"`
		Output  json.RawMessage `json:"output,omitempty"`
		Usage   json.RawMessage `json:"usage,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid advance_chain input: %v", err)), nil
	}
	input := types.AdvanceChainInput{
		ChainID: raw.ChainID,
		StepID:  raw.StepID,
		Outcome: raw.Outcome,
	}
	if err := decodeOptionalJSONArg(raw.Output, &input.Output); err != nil {
		return text(fmt.Sprintf("Invalid advance_chain output: %v", err)), nil
	}
	if err := decodeOptionalJSONArg(raw.Usage, &input.Usage); err != nil {
		return text(fmt.Sprintf("Invalid advance_chain usage: %v", err)), nil
	}
	if input.ChainID == "" || input.StepID == "" || input.Outcome == "" {
		return text("Missing: chainId, stepId, outcome"), nil
	}

	chainState, err := loadChainState(o.DB, input.ChainID)
	if err != nil {
		toolLog.Error("tool failed", "tool", "advance_chain", "chainId", input.ChainID, "stepId", input.StepID, "outcome", input.Outcome, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
	if err != nil {
		toolLog.Error("tool failed", "tool", "advance_chain", "chainId", input.ChainID, "stepId", input.StepID, "outcome", input.Outcome, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	if input.Usage != nil {
		budget.Update(&chainState.Budget, input.StepID, input.Usage)
	}

	advanced, err := state.AdvanceChain(state.AdvanceInput{
		State:   *chainState,
		Plan:    *plan,
		StepID:  input.StepID,
		Outcome: input.Outcome,
		Output:  input.Output,
	})
	if err != nil {
		toolLog.Error("tool failed", "tool", "advance_chain", "chainId", input.ChainID, "stepId", input.StepID, "outcome", input.Outcome, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	if input.Usage != nil {
		advanced.StateSnapshot.Budget = chainState.Budget
		advanced.AdvanceChainResult.Budget = chainState.Budget
	}

	if err := saveChainState(o.DB, o.projectRoot(), &advanced.StateSnapshot); err != nil {
		toolLog.Error("tool failed", "tool", "advance_chain", "chainId", input.ChainID, "stepId", input.StepID, "outcome", input.Outcome, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	nextStepID := ""
	if advanced.NextStep != nil {
		nextStepID = advanced.NextStep.StepID
	}

	o.Events.Publish(advanced.StateSnapshot.ChainID, fmt.Sprintf("chain.%s", advanced.State), map[string]any{
		"stepId":     input.StepID,
		"outcome":    input.Outcome,
		"state":      advanced.State,
		"nextStepId": nextStepID,
	})
	toolLog.Info("tool completed", "tool", "advance_chain", "chainId", advanced.StateSnapshot.ChainID, "stepId", input.StepID, "outcome", input.Outcome, "state", advanced.State, "nextStepId", nextStepID)

	return jsonOut(advanced.AdvanceChainResult), nil
}

func (o *Orchestrator) BuildTeam(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		Team   string          `json:"team"`
		Task   string          `json:"task"`
		Budget json.RawMessage `json:"budget,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid build_team input: %v", err)), nil
	}
	input := types.BuildTeamInput{Team: raw.Team, Task: raw.Task}
	if err := decodeOptionalJSONArg(raw.Budget, &input.Budget); err != nil {
		return text(fmt.Sprintf("Invalid build_team budget: %v", err)), nil
	}
	if input.Team == "" || input.Task == "" {
		return text("Missing required: team, task"), nil
	}
	plan, err := o.compileTeamPlan(input)
	if err != nil {
		toolLog.Error("tool failed", "tool", "build_team", "team", input.Team, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	def, err := decodeTeamDefinitionFromPlan(o, plan)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	teamState := state.CreateTeamState(def, plan)
	if err := saveTeamState(o.DB, o.projectRoot(), teamState); err != nil {
		toolLog.Error("tool failed", "tool", "build_team", "team", input.Team, "teamId", teamState.TeamID, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveTeamExecutionPlan(o.DB, plan); err != nil {
		toolLog.Error("tool failed", "tool", "build_team", "team", input.Team, "teamId", teamState.TeamID, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(teamState.TeamID, "team.started", map[string]any{
		"definitionName": teamState.DefinitionName,
		"task":           teamState.Task,
		"currentTaskId":  teamState.ReadyTaskIDs,
		"state":          teamState.State,
	})
	toolLog.Info("tool completed", "tool", "build_team", "team", input.Team, "teamId", teamState.TeamID, "state", teamState.State)
	return jsonOut(map[string]any{
		"teamId": teamState.TeamID,
		"state":  teamState.State,
		"tasks":  teamState.Tasks,
		"budget": teamState.Budget,
		"planId": plan.ID,
	}), nil
}

func (o *Orchestrator) AssignTeamTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input types.AssignTaskInput
	if err := bindArguments(req, &input); err != nil {
		return text(fmt.Sprintf("Invalid assign_team_task input: %v", err)), nil
	}
	if input.TeamID == "" || input.TaskID == "" {
		return text("Missing required: teamId, taskId"), nil
	}
	teamState, err := loadTeamState(o.DB, input.TeamID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	assign := input.Assignee
	claim := input.Claim
	updated, err := state.AssignTeamTask(teamState, input.TaskID, assign, claim)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveTeamState(o.DB, o.projectRoot(), updated); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(updated.TeamID, "team.task_assigned", map[string]any{"taskId": input.TaskID, "assignee": assign})
	toolLog.Info("tool completed", "tool", "assign_team_task", "teamId", updated.TeamID, "taskId", input.TaskID, "assignee", assign)
	return jsonOut(map[string]any{"teamId": updated.TeamID, "taskId": input.TaskID, "state": updated.State}), nil
}

func (o *Orchestrator) CompleteTeamTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		TeamID  string          `json:"teamId"`
		TaskID  string          `json:"taskId"`
		Outcome string          `json:"outcome"`
		Result  json.RawMessage `json:"result,omitempty"`
		Usage   json.RawMessage `json:"usage,omitempty"`
		Error   json.RawMessage `json:"error,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid complete_team_task input: %v", err)), nil
	}
	input := types.CompleteTaskInput{TeamID: raw.TeamID, TaskID: raw.TaskID, Outcome: raw.Outcome}
	if err := decodeOptionalJSONArg(raw.Result, &input.Result); err != nil {
		return text(fmt.Sprintf("Invalid complete_team_task result: %v", err)), nil
	}
	if err := decodeOptionalJSONArg(raw.Usage, &input.Usage); err != nil {
		return text(fmt.Sprintf("Invalid complete_team_task usage: %v", err)), nil
	}
	if err := decodeOptionalJSONArg(raw.Error, &input.Error); err != nil {
		return text(fmt.Sprintf("Invalid complete_team_task error: %v", err)), nil
	}
	if input.TeamID == "" || input.TaskID == "" || input.Outcome == "" {
		return text("Missing required: teamId, taskId, outcome"), nil
	}
	teamState, err := loadTeamState(o.DB, input.TeamID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if input.Usage != nil {
		budget.Update(&teamState.Budget, input.TaskID, input.Usage)
	}
	updated, err := state.CompleteTeamTask(teamState, input.TaskID, input.Outcome, input.Result, input.Usage, input.Error)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveTeamState(o.DB, o.projectRoot(), updated); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(updated.TeamID, fmt.Sprintf("team.%s", updated.State), map[string]any{"taskId": input.TaskID, "outcome": input.Outcome})
	toolLog.Info("tool completed", "tool", "complete_team_task", "teamId", updated.TeamID, "taskId", input.TaskID, "outcome", input.Outcome, "state", updated.State)
	return jsonOut(map[string]any{"teamId": updated.TeamID, "state": updated.State, "summary": updated.Summary}), nil
}

func (o *Orchestrator) StartWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		Workflow    string          `json:"workflow"`
		Task        string          `json:"task"`
		DomainSkill string          `json:"domainSkill,omitempty"`
		ModeSkill   string          `json:"modeSkill,omitempty"`
		Budget      json.RawMessage `json:"budget,omitempty"`
		Context     json.RawMessage `json:"context,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid start_workflow input: %v", err)), nil
	}
	input := types.StartWorkflowInput{
		Workflow:    raw.Workflow,
		Task:        raw.Task,
		DomainSkill: raw.DomainSkill,
		ModeSkill:   raw.ModeSkill,
	}
	if err := decodeOptionalJSONArg(raw.Budget, &input.Budget); err != nil {
		return text(fmt.Sprintf("Invalid start_workflow budget: %v", err)), nil
	}
	if err := decodeOptionalJSONArg(raw.Context, &input.Context); err != nil {
		return text(fmt.Sprintf("Invalid start_workflow context: %v", err)), nil
	}
	if input.Workflow == "" || input.Task == "" {
		return text("Missing required: workflow, task"), nil
	}
	workflowState, plan, err := startWorkflow(o, input)
	if err != nil {
		toolLog.Error("tool failed", "tool", "start_workflow", "workflow", input.Workflow, "error", err)
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	toolLog.Info("tool completed", "tool", "start_workflow", "workflow", input.Workflow, "workflowId", workflowState.WorkflowID, "state", workflowState.State)
	return jsonOut(map[string]any{
		"workflowId": workflowState.WorkflowID,
		"state":      workflowState.State,
		"phases":     workflowState.Phases,
		"budget":     workflowState.Budget,
		"planId":     plan.ID,
	}), nil
}

func (o *Orchestrator) AdvanceWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		WorkflowID string          `json:"workflowId"`
		Outcome    string          `json:"outcome,omitempty"`
		Recovery   json.RawMessage `json:"recovery,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid advance_workflow input: %v", err)), nil
	}
	input := types.AdvanceWorkflowInput{WorkflowID: raw.WorkflowID, Outcome: raw.Outcome}
	if err := decodeOptionalJSONArg(raw.Recovery, &input.Recovery); err != nil {
		return text(fmt.Sprintf("Invalid advance_workflow recovery: %v", err)), nil
	}
	if input.WorkflowID == "" {
		return text("Missing required: workflowId"), nil
	}
	workflowState, err := loadWorkflowState(o.DB, input.WorkflowID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	updated, err := state.AdvanceWorkflowState(workflowState, input.Outcome, input.Recovery)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveWorkflowState(o.DB, o.projectRoot(), updated); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(updated.WorkflowID, fmt.Sprintf("workflow.%s", updated.State), map[string]any{"outcome": input.Outcome, "state": updated.State})
	toolLog.Info("tool completed", "tool", "advance_workflow", "workflowId", updated.WorkflowID, "outcome", input.Outcome, "state", updated.State)
	return jsonOut(map[string]any{
		"workflowId": updated.WorkflowID,
		"state":      updated.State,
		"budget":     updated.Budget,
	}), nil
}

func (o *Orchestrator) GetStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID := runIDFromRequest(req)
	kind := req.GetString("kind", "")
	if runID == "" || kind == "" {
		return text("Missing: runId, kind"), nil
	}

	switch kind {
	case string(types.RunKindTeam):
		teamState, err := loadTeamState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"kind":  types.RunKindTeam,
			"runId": teamState.TeamID,
			"state": teamState.State,
			"summary": map[string]any{
				"definitionName":  teamState.DefinitionName,
				"task":            teamState.Task,
				"synthesisTaskId": teamState.SynthesisTaskID,
				"readyTaskIds":    teamState.ReadyTaskIDs,
				"totalMembers":    len(teamState.Tasks),
			},
			"tasks":   teamState.Tasks,
			"budget":  teamState.Budget,
			"history": o.Events.Replay(teamState.TeamID, 0),
		}), nil

	case string(types.RunKindWorkflow):
		workflowState, err := loadWorkflowState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"kind":  types.RunKindWorkflow,
			"runId": workflowState.WorkflowID,
			"state": workflowState.State,
			"summary": map[string]any{
				"definitionName":  workflowState.DefinitionName,
				"task":            workflowState.Task,
				"entryPhaseId":    workflowState.EntryPhaseID,
				"currentPhaseId":  workflowState.CurrentPhaseID,
				"totalPhases":     len(workflowState.Phases),
			},
			"phases":     workflowState.Phases,
			"childRuns":  workflowState.ChildRuns,
			"budget":     workflowState.Budget,
			"history":    o.Events.Replay(workflowState.WorkflowID, 0),
		}), nil

	case string(types.RunKindChain):
		chainState, err := loadChainState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"kind":  types.RunKindChain,
			"runId": chainState.ChainID,
			"state": chainState.State,
			"summary": map[string]any{
				"definitionName": chainState.DefinitionName,
				"totalSteps":     len(chainState.Steps),
				"completedSteps": len(chainState.CompletedStepIDs),
				"currentStepId":  chainState.CurrentStepID,
			},
			"current": currentChainStepStatus(plan, chainState),
			"steps":   chainState.Steps,
			"budget":  chainState.Budget,
			"history": o.Events.Replay(chainState.ChainID, 0),
		}), nil

	default:
		return text(fmt.Sprintf("Unsupported get_status kind %q", kind)), nil
	}
}

func (o *Orchestrator) GetBudget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID := runIDFromRequest(req)
	kind := req.GetString("kind", "")
	if runID == "" || kind == "" {
		return text("Missing: runId, kind"), nil
	}

	switch kind {
	case string(types.RunKindTeam):
		teamState, err := loadTeamState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"policyId":      teamState.Budget.PolicyID,
			"scope":         teamState.Budget.Scope,
			"tokens":        teamState.Budget.Tokens,
			"costUsd":       teamState.Budget.CostUsd,
			"wallClockMs":   teamState.Budget.WallClockMs,
			"retries":       teamState.Budget.Retries,
			"byStep":        teamState.Budget.ByStep,
			"lastUpdatedAt": teamState.Budget.LastUpdatedAt,
			"health":        budget.Evaluate(&teamState.Budget, &teamState.BudgetPolicy),
		}), nil

	case string(types.RunKindWorkflow):
		workflowState, err := loadWorkflowState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"policyId":      workflowState.Budget.PolicyID,
			"scope":         workflowState.Budget.Scope,
			"tokens":        workflowState.Budget.Tokens,
			"costUsd":       workflowState.Budget.CostUsd,
			"wallClockMs":   workflowState.Budget.WallClockMs,
			"retries":       workflowState.Budget.Retries,
			"byStep":        workflowState.Budget.ByStep,
			"lastUpdatedAt": workflowState.Budget.LastUpdatedAt,
			"health":        budget.Evaluate(&workflowState.Budget, &workflowState.BudgetPolicy),
		}), nil

	case string(types.RunKindChain):
		chainState, err := loadChainState(o.DB, runID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
		if err != nil {
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		return jsonOut(map[string]any{
			"policyId":      chainState.Budget.PolicyID,
			"scope":         chainState.Budget.Scope,
			"tokens":        chainState.Budget.Tokens,
			"costUsd":       chainState.Budget.CostUsd,
			"wallClockMs":   chainState.Budget.WallClockMs,
			"retries":       chainState.Budget.Retries,
			"byStep":        chainState.Budget.ByStep,
			"lastUpdatedAt": chainState.Budget.LastUpdatedAt,
			"health":        budget.Evaluate(&chainState.Budget, &plan.BudgetPolicy),
		}), nil

	default:
		return text(fmt.Sprintf("Unsupported get_budget kind %q", kind)), nil
	}
}

func (o *Orchestrator) RetryStep(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := retryStepInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" || input.StepID == "" {
		return text("Missing: runId, kind, stepId"), nil
	}

	switch input.Kind {
	case types.RunKindTeam:
		teamState, err := loadTeamState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		retried, err := state.RetryTeamTask(teamState, input.StepID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		budget.IncrementRetries(&retried.Budget, input.StepID)
		if err := saveTeamState(o.DB, o.projectRoot(), retried); err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(retried.TeamID, "team.retrying", map[string]any{"taskId": input.StepID, "state": retried.State})
		toolLog.Info("tool completed", "tool", "retry_step", "runId", retried.TeamID, "stepId", input.StepID, "kind", "team")
		return jsonOut(map[string]any{"runId": retried.TeamID, "stepId": input.StepID, "state": retried.State, "budget": retried.Budget}), nil

	case types.RunKindWorkflow:
		workflowState, err := loadWorkflowState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		recovery := &types.WorkflowRecoveryDecision{Type: "retry", TargetPhaseID: input.StepID}
		retried, err := state.AdvanceWorkflowState(workflowState, "", recovery)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		budget.IncrementRetries(&retried.Budget, input.StepID)
		if err := saveWorkflowState(o.DB, o.projectRoot(), retried); err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(retried.WorkflowID, "workflow.retrying", map[string]any{"phaseId": input.StepID, "state": retried.State})
		toolLog.Info("tool completed", "tool", "retry_step", "runId", retried.WorkflowID, "stepId", input.StepID, "kind", "workflow")
		return jsonOut(map[string]any{"runId": retried.WorkflowID, "stepId": input.StepID, "state": retried.State, "budget": retried.Budget}), nil

	case types.RunKindChain:
		chainState, err := loadChainState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		retried, attemptsRemaining, err := state.RetryChainStep(chainState, plan, input.StepID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		budget.IncrementRetries(&retried.Budget, input.StepID)
		if err := saveChainState(o.DB, o.projectRoot(), retried); err != nil {
			toolLog.Error("tool failed", "tool", "retry_step", "runId", input.RunID, "stepId", input.StepID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(retried.ChainID, "chain.retrying", map[string]any{
			"stepId": input.StepID, "attemptsRemaining": attemptsRemaining, "state": retried.State,
		})
		toolLog.Info("tool completed", "tool", "retry_step", "runId", retried.ChainID, "stepId", input.StepID, "attemptsRemaining", attemptsRemaining)
		return jsonOut(map[string]any{
			"runId": retried.ChainID, "stepId": input.StepID, "state": retried.State,
			"attemptsRemaining": attemptsRemaining, "budget": retried.Budget,
		}), nil

	default:
		return text(fmt.Sprintf("Unsupported retry_step kind %q", input.Kind)), nil
	}
}

func (o *Orchestrator) EscalateStep(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := escalateStepInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" || input.StepID == "" || input.TargetAgent == "" {
		return text("Missing: runId, kind, stepId, targetAgent"), nil
	}

	switch input.Kind {
	case types.RunKindTeam:
		teamState, err := loadTeamState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		escalated, err := state.EscalateTeamTask(teamState, input.StepID, input.TargetAgent)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveTeamState(o.DB, o.projectRoot(), escalated); err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(escalated.TeamID, "team.escalated", map[string]any{"taskId": input.StepID, "targetAgent": input.TargetAgent, "state": escalated.State})
		toolLog.Info("tool completed", "tool", "escalate_step", "runId", escalated.TeamID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "kind", "team")
		return jsonOut(map[string]any{"runId": escalated.TeamID, "stepId": input.StepID, "state": escalated.State}), nil

	case types.RunKindWorkflow:
		workflowState, err := loadWorkflowState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		recovery := &types.WorkflowRecoveryDecision{Type: "escalate", TargetPhaseID: input.StepID, Recipient: input.TargetAgent}
		escalated, err := state.AdvanceWorkflowState(workflowState, "", recovery)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveWorkflowState(o.DB, o.projectRoot(), escalated); err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(escalated.WorkflowID, "workflow.escalated", map[string]any{"phaseId": input.StepID, "targetAgent": input.TargetAgent, "state": escalated.State})
		toolLog.Info("tool completed", "tool", "escalate_step", "runId", escalated.WorkflowID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "kind", "workflow")
		return jsonOut(map[string]any{"runId": escalated.WorkflowID, "stepId": input.StepID, "state": escalated.State}), nil

	case types.RunKindChain:
		chainState, err := loadChainState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		escalatedState, escalatedPlan, err := state.EscalateChainStep(chainState, plan, input.StepID, input.TargetAgent, input.DomainSkill, input.ModeSkill)
		if err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveExecutionPlan(o.DB, escalatedPlan); err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveChainState(o.DB, o.projectRoot(), escalatedState); err != nil {
			toolLog.Error("tool failed", "tool", "escalate_step", "runId", input.RunID, "stepId", input.StepID, "targetAgent", input.TargetAgent, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(escalatedState.ChainID, "chain.escalated", map[string]any{
			"stepId": input.StepID, "targetAgent": input.TargetAgent, "state": escalatedState.State,
		})
		toolLog.Info("tool completed", "tool", "escalate_step", "runId", escalatedState.ChainID, "stepId", input.StepID, "targetAgent", input.TargetAgent)
		return jsonOut(map[string]any{
			"runId": escalatedState.ChainID, "stepId": input.StepID, "state": escalatedState.State,
			"newAssignment": currentChainStepStatus(escalatedPlan, escalatedState),
		}), nil

	default:
		return text(fmt.Sprintf("Unsupported escalate_step kind %q", input.Kind)), nil
	}
}

func (o *Orchestrator) Handoff(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := handoffInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" {
		return text("Missing: runId, kind"), nil
	}

	switch input.Kind {
	case types.RunKindTeam:
		teamState, err := loadTeamState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadTeamExecutionPlan(o.DB, teamState.ExecutionPlanID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		doc, path, err := createAndSaveTeamHandoff(o.DB, teamState, plan, input)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveTeamState(o.DB, o.projectRoot(), teamState); err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "handoffId", doc.ID, "path", path, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(teamState.TeamID, "team.handoff", map[string]any{
			"handoffId": doc.ID, "path": path, "state": teamState.State,
		})
		toolLog.Info("tool completed", "tool", "handoff", "runId", teamState.TeamID, "handoffId", doc.ID, "path", path, "kind", "team")
		return jsonOut(map[string]any{"handoffId": doc.ID, "path": path, "summary": doc.Summary, "resumable": doc.Resumable}), nil

	case types.RunKindWorkflow:
		workflowState, err := loadWorkflowState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadWorkflowExecutionPlan(o.DB, workflowState.ExecutionPlanID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		doc, path, err := createAndSaveWorkflowHandoff(o.DB, workflowState, plan, input)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveWorkflowState(o.DB, o.projectRoot(), workflowState); err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "handoffId", doc.ID, "path", path, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(workflowState.WorkflowID, "workflow.handoff", map[string]any{
			"handoffId": doc.ID, "path": path, "state": workflowState.State,
		})
		toolLog.Info("tool completed", "tool", "handoff", "runId", workflowState.WorkflowID, "handoffId", doc.ID, "path", path, "kind", "workflow")
		return jsonOut(map[string]any{"handoffId": doc.ID, "path": path, "summary": doc.Summary, "resumable": doc.Resumable}), nil

	case types.RunKindChain:
		chainState, err := loadChainState(o.DB, input.RunID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		doc, path, err := createAndSaveChainHandoff(o.DB, chainState, plan, input)
		if err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		if err := saveChainState(o.DB, o.projectRoot(), chainState); err != nil {
			toolLog.Error("tool failed", "tool", "handoff", "runId", input.RunID, "handoffId", doc.ID, "path", path, "error", err)
			return text(fmt.Sprintf("Error: %v", err)), nil
		}
		o.Events.Publish(chainState.ChainID, "chain.handoff", map[string]any{
			"handoffId": doc.ID, "path": path, "state": chainState.State,
		})
		toolLog.Info("tool completed", "tool", "handoff", "runId", chainState.ChainID, "handoffId", doc.ID, "path", path)
		return jsonOut(map[string]any{"handoffId": doc.ID, "path": path, "summary": doc.Summary, "resumable": doc.Resumable}), nil

	default:
		return text(fmt.Sprintf("Unsupported handoff kind %q", input.Kind)), nil
	}
}

func (o *Orchestrator) CatalogList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	kind := req.GetString("kind", "")
	items, err := o.Catalog.List(kind)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(items), nil
}

func (o *Orchestrator) CatalogGetVersion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	kind := req.GetString("kind", "")
	name := req.GetString("name", "")
	v, err := o.Catalog.GetVersion(kind, name, 0)
	if err != nil {
		return text(fmt.Sprintf("Not found: %s/%s", kind, name)), nil
	}
	return jsonOut(v), nil
}

func (o *Orchestrator) CatalogCreateVersion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	kind := req.GetString("kind", "")
	name := req.GetString("name", "")
	body := req.GetString("body", "")
	result, err := o.Catalog.CreateVersion(catalog.CreateVersionInput{Kind: kind, Name: name, Body: body, SetActive: true})
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(result), nil
}

func (o *Orchestrator) CatalogSetActive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	kind := req.GetString("kind", "")
	name := req.GetString("name", "")
	version := 0
	if v, err := req.RequireString("version"); err == nil {
		fmt.Sscanf(v, "%d", &version)
	}
	if err := o.Catalog.SetActive(kind, name, version); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(map[string]string{"status": "activated", "name": name}), nil
}

func (o *Orchestrator) InvokeAgent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agent := req.GetString("agent", "")
	task := req.GetString("task", "")
	cliTool := req.GetString("cliTool", "")
	result, err := o.Dispatcher.InvokeAgent(ctx, dispatch.InvocationRequest{Agent: agent, Task: task, CliTool: cliTool})
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if result == nil || result.Spec == nil {
		return text("Error: dispatcher returned no invocation spec"), nil
	}
	return jsonOut(result.Spec), nil
}

func (o *Orchestrator) SubscribeRun(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID := req.GetString("runId", "")
	ch := make(chan events.Event, 32)
	past := o.Events.Subscribe(runID, ch)
	return jsonOut(map[string]any{"subscribed": true, "runId": runID, "pastEvents": past}), nil
}

func (o *Orchestrator) EnqueueJob(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var raw struct {
		JobType string          `json:"jobType"`
		Payload json.RawMessage `json:"payload,omitempty"`
	}
	if err := bindArguments(req, &raw); err != nil {
		return text(fmt.Sprintf("Invalid enqueue_job input: %v", err)), nil
	}
	var payload map[string]any
	if err := decodeOptionalJSONArg(raw.Payload, &payload); err != nil {
		return text(fmt.Sprintf("Invalid enqueue_job payload: %v", err)), nil
	}
	job, err := o.Queue.Enqueue(queue.EnqueueInput{JobType: raw.JobType, Payload: payload})
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(job), nil
}

func (o *Orchestrator) GetJob(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := req.GetString("jobId", "")
	job, err := o.Queue.Get(jobID)
	if err != nil {
		return text(fmt.Sprintf("Not found: %s", jobID)), nil
	}
	return jsonOut(job), nil
}

func (o *Orchestrator) ListJobs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := req.GetString("status", "")
	jobs, err := o.Queue.List(status, 50)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	return jsonOut(jobs), nil
}

// ──────────────── helpers ─────────────────

func jsonOut(v any) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}}}
}

func text(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}}}
}

type ScopeContext struct {
	Scope       string
	ProjectRoot string
	GlobalRoot  string
}

func NewScopeContext(scope, projectRoot, globalRoot string) *ScopeContext {
	return &ScopeContext{Scope: scope, ProjectRoot: projectRoot, GlobalRoot: globalRoot}
}
