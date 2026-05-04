package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/budget"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/catalog"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/dispatch"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/events"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/queue"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/state"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
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
	s.AddTool(mcp.Tool{Name: "list_catalog", Description: "List orchestration catalog definitions."}, o.ListCatalog)
	s.AddTool(mcp.Tool{Name: "compose_agent", Description: "Compose a runtime agent prompt."}, o.ComposeAgent)
	s.AddTool(mcp.Tool{Name: "start_chain", Description: "Compile and start a chain execution plan."}, o.StartChain)
	s.AddTool(mcp.Tool{Name: "advance_chain", Description: "Advance a running chain."}, o.AdvanceChain)
	s.AddTool(mcp.Tool{Name: "build_team", Description: "Compile and start a team run."}, o.BuildTeam)
	s.AddTool(mcp.Tool{Name: "assign_team_task", Description: "Assign or claim a team task."}, o.AssignTeamTask)
	s.AddTool(mcp.Tool{Name: "complete_team_task", Description: "Complete a team task."}, o.CompleteTeamTask)
	s.AddTool(mcp.Tool{Name: "start_workflow", Description: "Compile and start a workflow run."}, o.StartWorkflow)
	s.AddTool(mcp.Tool{Name: "advance_workflow", Description: "Advance a running workflow."}, o.AdvanceWorkflow)
	s.AddTool(mcp.Tool{Name: "get_status", Description: "Get runtime status for a run."}, o.GetStatus)
	s.AddTool(mcp.Tool{Name: "get_budget", Description: "Get tracked budget state."}, o.GetBudget)
	s.AddTool(mcp.Tool{Name: "retry_step", Description: "Retry a failed step."}, o.RetryStep)
	s.AddTool(mcp.Tool{Name: "escalate_step", Description: "Escalate a step."}, o.EscalateStep)
	s.AddTool(mcp.Tool{Name: "handoff", Description: "Persist a resumable handoff document."}, o.Handoff)
	s.AddTool(mcp.Tool{Name: "catalog_list", Description: "List versioned catalog definitions."}, o.CatalogList)
	s.AddTool(mcp.Tool{Name: "catalog_get_version", Description: "Get catalog definition version."}, o.CatalogGetVersion)
	s.AddTool(mcp.Tool{Name: "catalog_create_version", Description: "Create immutable version."}, o.CatalogCreateVersion)
	s.AddTool(mcp.Tool{Name: "catalog_set_active", Description: "Move active version pointer."}, o.CatalogSetActive)
	s.AddTool(mcp.Tool{Name: "subscribe_run", Description: "Subscribe to run events."}, o.SubscribeRun)
	s.AddTool(mcp.Tool{Name: "invoke_agent", Description: "Resolve and compose agent invocation."}, o.InvokeAgent)
	s.AddTool(mcp.Tool{Name: "enqueue_job", Description: "Enqueue background job."}, o.EnqueueJob)
	s.AddTool(mcp.Tool{Name: "get_job", Description: "Get job status."}, o.GetJob)
	s.AddTool(mcp.Tool{Name: "list_jobs", Description: "List queued jobs."}, o.ListJobs)
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
	var input types.StartChainInput
	if err := bindArguments(req, &input); err != nil {
		return text(fmt.Sprintf("Invalid start_chain input: %v", err)), nil
	}
	if input.Chain == "" || input.Task == "" {
		return text("Missing required: chain, task"), nil
	}

	plan, err := o.compileChainPlan(input)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	chainState := state.CreateChainState(plan)

	if err := saveExecutionPlan(o.DB, plan); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveChainState(o.DB, o.projectRoot(), chainState); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	o.Events.Publish(chainState.ChainID, "chain.started", map[string]any{
		"definitionName": chainState.DefinitionName,
		"task":           chainState.Task,
		"currentStepId":  chainState.CurrentStepID,
		"state":          chainState.State,
	})

	return jsonOut(map[string]any{
		"chainId":         chainState.ChainID,
		"state":           chainState.State,
		"currentStep":     currentChainStepStatus(plan, chainState),
		"budget":          chainState.Budget,
		"executionPlanId": plan.ID,
	}), nil
}

func (o *Orchestrator) AdvanceChain(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input types.AdvanceChainInput
	if err := bindArguments(req, &input); err != nil {
		return text(fmt.Sprintf("Invalid advance_chain input: %v", err)), nil
	}
	if input.ChainID == "" || input.StepID == "" || input.Outcome == "" {
		return text("Missing: chainId, stepId, outcome"), nil
	}

	chainState, err := loadChainState(o.DB, input.ChainID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
	if err != nil {
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
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	if input.Usage != nil {
		advanced.StateSnapshot.Budget = chainState.Budget
		advanced.AdvanceChainResult.Budget = chainState.Budget
	}

	if err := saveChainState(o.DB, o.projectRoot(), &advanced.StateSnapshot); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}

	o.Events.Publish(advanced.StateSnapshot.ChainID, fmt.Sprintf("chain.%s", advanced.State), map[string]any{
		"stepId":  input.StepID,
		"outcome": input.Outcome,
		"state":   advanced.State,
		"nextStepId": func() string {
			if advanced.NextStep == nil {
				return ""
			}
			return advanced.NextStep.StepID
		}(),
	})

	return jsonOut(advanced.AdvanceChainResult), nil
}

func (o *Orchestrator) BuildTeam(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	team := req.GetString("team", "")
	task := req.GetString("task", "")
	if team == "" || task == "" {
		return text("Missing: team, task"), nil
	}
	return jsonOut(map[string]string{"teamId": "team-1", "status": "created"}), nil
}

func (o *Orchestrator) AssignTeamTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonOut(map[string]string{"status": "assigned"}), nil
}

func (o *Orchestrator) CompleteTeamTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonOut(map[string]string{"status": "completed"}), nil
}

func (o *Orchestrator) StartWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflow := req.GetString("workflow", "")
	task := req.GetString("task", "")
	if workflow == "" || task == "" {
		return text("Missing: workflow, task"), nil
	}
	return jsonOut(map[string]string{"workflowId": "wf-1", "status": "created"}), nil
}

func (o *Orchestrator) AdvanceWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonOut(map[string]string{"status": "advanced"}), nil
}

func (o *Orchestrator) GetStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID := runIDFromRequest(req)
	kind := req.GetString("kind", "")
	if runID == "" || kind == "" {
		return text("Missing: runId, kind"), nil
	}
	if kind != string(types.RunKindChain) {
		return text(fmt.Sprintf("Unsupported get_status kind %q in Go core parity slice", kind)), nil
	}
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
}

func (o *Orchestrator) GetBudget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID := runIDFromRequest(req)
	kind := req.GetString("kind", "")
	if runID == "" || kind == "" {
		return text("Missing: runId, kind"), nil
	}
	if kind != string(types.RunKindChain) {
		return text(fmt.Sprintf("Unsupported get_budget kind %q in Go core parity slice", kind)), nil
	}
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
}

func (o *Orchestrator) RetryStep(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := retryStepInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" || input.StepID == "" {
		return text("Missing: runId, kind, stepId"), nil
	}
	if input.Kind != types.RunKindChain {
		return text(fmt.Sprintf("Unsupported retry_step kind %q in Go core parity slice", input.Kind)), nil
	}
	chainState, err := loadChainState(o.DB, input.RunID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	retried, attemptsRemaining, err := state.RetryChainStep(chainState, plan, input.StepID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	budget.IncrementRetries(&retried.Budget, input.StepID)
	if err := saveChainState(o.DB, o.projectRoot(), retried); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(retried.ChainID, "chain.retrying", map[string]any{
		"stepId": input.StepID, "attemptsRemaining": attemptsRemaining, "state": retried.State,
	})
	return jsonOut(map[string]any{
		"runId": retried.ChainID, "stepId": input.StepID, "state": retried.State,
		"attemptsRemaining": attemptsRemaining, "budget": retried.Budget,
	}), nil
}

func (o *Orchestrator) EscalateStep(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := escalateStepInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" || input.StepID == "" || input.TargetAgent == "" {
		return text("Missing: runId, kind, stepId, targetAgent"), nil
	}
	if input.Kind != types.RunKindChain {
		return text(fmt.Sprintf("Unsupported escalate_step kind %q in Go core parity slice", input.Kind)), nil
	}
	chainState, err := loadChainState(o.DB, input.RunID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	escalatedState, escalatedPlan, err := state.EscalateChainStep(chainState, plan, input.StepID, input.TargetAgent, input.DomainSkill, input.ModeSkill)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveExecutionPlan(o.DB, escalatedPlan); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveChainState(o.DB, o.projectRoot(), escalatedState); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(escalatedState.ChainID, "chain.escalated", map[string]any{
		"stepId": input.StepID, "targetAgent": input.TargetAgent, "state": escalatedState.State,
	})
	return jsonOut(map[string]any{
		"runId": escalatedState.ChainID, "stepId": input.StepID, "state": escalatedState.State,
		"newAssignment": currentChainStepStatus(escalatedPlan, escalatedState),
	}), nil
}

func (o *Orchestrator) Handoff(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := handoffInputFromRequest(req)
	if input.RunID == "" || input.Kind == "" {
		return text("Missing: runId, kind"), nil
	}
	if input.Kind != types.RunKindChain {
		return text(fmt.Sprintf("Unsupported handoff kind %q in Go core parity slice", input.Kind)), nil
	}
	chainState, err := loadChainState(o.DB, input.RunID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	plan, err := loadExecutionPlan(o.DB, chainState.ExecutionPlanID)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	doc, path, err := createAndSaveChainHandoff(o.DB, chainState, plan, input)
	if err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	if err := saveChainState(o.DB, o.projectRoot(), chainState); err != nil {
		return text(fmt.Sprintf("Error: %v", err)), nil
	}
	o.Events.Publish(chainState.ChainID, "chain.handoff", map[string]any{
		"handoffId": doc.ID, "path": path, "state": chainState.State,
	})
	return jsonOut(map[string]any{"handoffId": doc.ID, "path": path, "summary": doc.Summary, "resumable": doc.Resumable}), nil
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
	result, err := o.Dispatcher.InvokeAgent(ctx, dispatch.InvocationRequest{Agent: agent, Task: task})
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
	jobType := req.GetString("jobType", "")
	job, err := o.Queue.Enqueue(queue.EnqueueInput{JobType: jobType})
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
