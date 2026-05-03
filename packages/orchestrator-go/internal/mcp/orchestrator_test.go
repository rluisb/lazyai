package mcp

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/catalog"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

func TestChainLifecycleStartAdvanceStatus(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	versions, err := orchestrator.Catalog.ListVersions("chain", "simple-chain")
	if err != nil || len(versions) == 0 {
		t.Fatalf("expected stored catalog version, versions=%+v err=%v", versions, err)
	}

	startResult, err := orchestrator.StartChain(context.Background(), toolRequest(map[string]any{
		"chain": "simple-chain",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("start chain: %v", err)
	}
	var started struct {
		ChainID     string `json:"chainId"`
		State       string `json:"state"`
		CurrentStep struct {
			StepID string `json:"stepId"`
			State  string `json:"state"`
		} `json:"currentStep"`
	}
	decodeToolResult(t, startResult, &started)
	if started.ChainID == "" {
		t.Fatalf("expected chain id in start result")
	}
	if started.State != "running" || started.CurrentStep.StepID != "research" || started.CurrentStep.State != "running" {
		t.Fatalf("unexpected start result: %+v", started)
	}

	advanceResult, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
		"chainId": started.ChainID,
		"stepId":  "research",
		"outcome": "success",
		"output": map[string]any{
			"summary":  "researched",
			"status":   "done",
			"findings": []any{},
		},
	}))
	if err != nil {
		t.Fatalf("advance chain: %v", err)
	}
	var advanced struct {
		State    string `json:"state"`
		NextStep struct {
			StepID string `json:"stepId"`
			State  string `json:"state"`
		} `json:"nextStep"`
	}
	decodeToolResult(t, advanceResult, &advanced)
	if advanced.State != "running" || advanced.NextStep.StepID != "implement" || advanced.NextStep.State != "running" {
		t.Fatalf("unexpected advance result: %+v", advanced)
	}

	statusResult, err := orchestrator.GetStatus(context.Background(), toolRequest(map[string]any{
		"runId": started.ChainID,
		"kind":  "chain",
	}))
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	var status struct {
		State   string `json:"state"`
		Summary struct {
			CompletedSteps int    `json:"completedSteps"`
			CurrentStepID  string `json:"currentStepId"`
		} `json:"summary"`
		History []any `json:"history"`
	}
	decodeToolResult(t, statusResult, &status)
	if status.State != "running" || status.Summary.CompletedSteps != 1 || status.Summary.CurrentStepID != "implement" {
		t.Fatalf("unexpected status result: %+v", status)
	}
	if len(status.History) == 0 {
		t.Fatalf("expected persisted run history")
	}
}

func TestStartChainUnknownDefinitionDoesNotCreateRunState(t *testing.T) {
	orchestrator := newTestOrchestrator(t)

	result, err := orchestrator.StartChain(context.Background(), toolRequest(map[string]any{
		"chain": "missing-chain",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("start chain returned transport error: %v", err)
	}
	if text := decodeToolText(t, result); !strings.Contains(text, "unknown chain definition") {
		t.Fatalf("expected unknown chain error, got %q", text)
	}
	if count := countRows(t, orchestrator.DB, "chain_runs"); count != 0 {
		t.Fatalf("expected no chain run rows, got %d", count)
	}
	if count := countRows(t, orchestrator.DB, "execution_plans"); count != 0 {
		t.Fatalf("expected no execution plan rows, got %d", count)
	}
}

func TestAdvanceChainRejectsPendingStepAndPreservesState(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)
	before, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load chain before advance: %v", err)
	}

	result, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
		"chainId": started.ChainID,
		"stepId":  "implement",
		"outcome": "success",
		"output": map[string]any{
			"summary":       "done out of order",
			"status":        "done",
			"files_changed": []any{},
			"tests_passed":  true,
		},
		"usage": map[string]any{"totalTokens": 50},
	}))
	if err != nil {
		t.Fatalf("advance chain returned transport error: %v", err)
	}
	if text := decodeToolText(t, result); !strings.Contains(text, "is not active") {
		t.Fatalf("expected pending step rejection, got %q", text)
	}
	after, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load chain after advance: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("chain state changed after rejected advance\nbefore=%+v\nafter=%+v", before, after)
	}
}

func TestRetryStepRejectsExhaustedRetryLimitAndPreservesStateAndBudget(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	failChainStep(t, orchestrator, started.ChainID, "research")
	retryResult, err := orchestrator.RetryStep(context.Background(), toolRequest(map[string]any{
		"runId":  started.ChainID,
		"kind":   "chain",
		"stepId": "research",
		"reason": "try again",
	}))
	if err != nil {
		t.Fatalf("retry step returned transport error: %v", err)
	}
	var retried struct {
		State             string `json:"state"`
		AttemptsRemaining int    `json:"attemptsRemaining"`
		Budget            struct {
			Retries struct {
				Consumed int `json:"consumed"`
			} `json:"retries"`
		} `json:"budget"`
	}
	decodeToolResult(t, retryResult, &retried)
	if retried.State != "running" || retried.AttemptsRemaining != 0 || retried.Budget.Retries.Consumed != 1 {
		t.Fatalf("unexpected retry result: %+v", retried)
	}

	failChainStep(t, orchestrator, started.ChainID, "research")
	before, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load chain before exhausted retry: %v", err)
	}

	exhaustedResult, err := orchestrator.RetryStep(context.Background(), toolRequest(map[string]any{
		"runId":  started.ChainID,
		"kind":   "chain",
		"stepId": "research",
		"reason": "one too many",
	}))
	if err != nil {
		t.Fatalf("retry step returned transport error: %v", err)
	}
	if text := decodeToolText(t, exhaustedResult); !strings.Contains(text, "no retries remaining") {
		t.Fatalf("expected exhausted retry rejection, got %q", text)
	}
	after, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load chain after exhausted retry: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("chain state/budget changed after rejected retry\nbefore=%+v\nafter=%+v", before, after)
	}
}

func TestEscalateStepPersistsTargetAgentInStateAndPlan(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	result, err := orchestrator.EscalateStep(context.Background(), toolRequest(map[string]any{
		"runId":       started.ChainID,
		"kind":        "chain",
		"stepId":      "research",
		"targetAgent": "senior-builder",
		"reason":      "needs senior help",
	}))
	if err != nil {
		t.Fatalf("escalate step returned transport error: %v", err)
	}
	var escalated struct {
		State         string `json:"state"`
		NewAssignment struct {
			StepID string `json:"stepId"`
			Agent  string `json:"agent"`
			State  string `json:"state"`
		} `json:"newAssignment"`
	}
	decodeToolResult(t, result, &escalated)
	if escalated.State != "running" || escalated.NewAssignment.StepID != "research" || escalated.NewAssignment.Agent != "senior-builder" || escalated.NewAssignment.State != "running" {
		t.Fatalf("unexpected escalation result: %+v", escalated)
	}
	chainState, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load escalated chain: %v", err)
	}
	if chainState.State != types.ChainRunning || chainState.CurrentStepID != "research" {
		t.Fatalf("expected escalated run to remain resumable/running, got state=%q current=%q", chainState.State, chainState.CurrentStepID)
	}
	if step := findStep(t, chainState, "research"); step.Agent != "senior-builder" || step.State != types.StepRunning {
		t.Fatalf("target agent was not persisted in state: %+v", step)
	}
	plan, err := loadExecutionPlan(orchestrator.DB, chainState.ExecutionPlanID)
	if err != nil {
		t.Fatalf("load escalated plan: %v", err)
	}
	if compiled := findCompiledStep(t, plan, "research"); compiled.Agent != "senior-builder" {
		t.Fatalf("target agent was not persisted in plan: %+v", compiled)
	}
}

func TestHandoffPersistsResumableArtifactWithStatusPlanAndPath(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	result, err := orchestrator.Handoff(context.Background(), toolRequest(map[string]any{
		"runId":     started.ChainID,
		"kind":      "chain",
		"summary":   "Continue later",
		"recipient": "next-agent",
	}))
	if err != nil {
		t.Fatalf("handoff returned transport error: %v", err)
	}
	var handoffResult struct {
		HandoffID string `json:"handoffId"`
		Path      string `json:"path"`
		Summary   string `json:"summary"`
		Resumable bool   `json:"resumable"`
	}
	decodeToolResult(t, result, &handoffResult)
	if handoffResult.HandoffID == "" || !handoffResult.Resumable || handoffResult.Summary != "Continue later" || !strings.HasPrefix(handoffResult.Path, handoffPathURIPrefix) {
		t.Fatalf("unexpected handoff result: %+v", handoffResult)
	}

	chainState, err := loadChainState(orchestrator.DB, started.ChainID)
	if err != nil {
		t.Fatalf("load handed off chain: %v", err)
	}
	if chainState.State != types.ChainHandoff || chainState.HandoffPath != handoffResult.Path {
		t.Fatalf("handoff state/path not persisted: state=%q path=%q resultPath=%q", chainState.State, chainState.HandoffPath, handoffResult.Path)
	}

	var docJSON string
	if err := orchestrator.DB.QueryRow(`SELECT doc_json FROM handoffs WHERE id = ?`, handoffResult.HandoffID).Scan(&docJSON); err != nil {
		t.Fatalf("load handoff document: %v", err)
	}
	var doc struct {
		RunID     string               `json:"runId"`
		Kind      types.RunKind        `json:"kind"`
		Summary   string               `json:"summary"`
		Recipient string               `json:"recipient"`
		Resumable bool                 `json:"resumable"`
		Status    types.ChainState     `json:"status"`
		Plan      *types.ExecutionPlan `json:"plan"`
	}
	if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
		t.Fatalf("decode handoff document: %v", err)
	}
	if doc.RunID != started.ChainID || doc.Kind != types.RunKindChain || doc.Summary != "Continue later" || doc.Recipient != "next-agent" || !doc.Resumable {
		t.Fatalf("unexpected handoff document metadata: %+v", doc)
	}
	if doc.Status.State != types.ChainHandoff || doc.Status.ChainID != started.ChainID {
		t.Fatalf("handoff document did not include resumable chain status: %+v", doc.Status)
	}
	if doc.Plan == nil || doc.Plan.ID != chainState.ExecutionPlanID || doc.Plan.Definition.Name != "simple-chain" {
		t.Fatalf("handoff document did not include execution plan: %+v", doc.Plan)
	}
}

func toolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func newTestOrchestrator(t *testing.T) *Orchestrator {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewOrchestrator(database, NewScopeContext("project", "/tmp/project", ""))
}

func createSimpleChainCatalog(t *testing.T, orchestrator *Orchestrator) {
	t.Helper()
	created, err := orchestrator.Catalog.CreateVersion(catalog.CreateVersionInput{
		Kind:      "chain",
		Name:      "simple-chain",
		Body:      simpleChainDefinitionJSON(),
		SetActive: true,
	})
	if err != nil {
		t.Fatalf("create chain catalog version: %v", err)
	}
	if created.AlreadyExists {
		t.Fatalf("expected new catalog version, got duplicate: %+v", created)
	}
}

func startSimpleChain(t *testing.T, orchestrator *Orchestrator) struct {
	ChainID     string `json:"chainId"`
	State       string `json:"state"`
	CurrentStep struct {
		StepID string `json:"stepId"`
		State  string `json:"state"`
	} `json:"currentStep"`
} {
	t.Helper()
	result, err := orchestrator.StartChain(context.Background(), toolRequest(map[string]any{
		"chain": "simple-chain",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("start simple chain: %v", err)
	}
	var started struct {
		ChainID     string `json:"chainId"`
		State       string `json:"state"`
		CurrentStep struct {
			StepID string `json:"stepId"`
			State  string `json:"state"`
		} `json:"currentStep"`
	}
	decodeToolResult(t, result, &started)
	if started.ChainID == "" {
		t.Fatalf("expected chain id in start result")
	}
	return started
}

func failChainStep(t *testing.T, orchestrator *Orchestrator, chainID, stepID string) {
	t.Helper()
	result, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
		"chainId": chainID,
		"stepId":  stepID,
		"outcome": "failure",
	}))
	if err != nil {
		t.Fatalf("fail chain step: %v", err)
	}
	if text := decodeToolText(t, result); strings.HasPrefix(text, "Error:") {
		t.Fatalf("fail chain step returned error: %s", text)
	}
}

func decodeToolText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatalf("tool result had no content")
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	return content.Text
}

func decodeToolResult(t *testing.T, result *mcp.CallToolResult, target any) {
	t.Helper()
	text := decodeToolText(t, result)
	if err := json.Unmarshal([]byte(text), target); err != nil {
		t.Fatalf("decode result %q: %v", text, err)
	}
}

func countRows(t *testing.T, database *db.DB, table string) int {
	t.Helper()
	queries := map[string]string{
		"chain_runs":      `SELECT COUNT(*) FROM chain_runs`,
		"execution_plans": `SELECT COUNT(*) FROM execution_plans`,
	}
	query, ok := queries[table]
	if !ok {
		t.Fatalf("unsupported test table %q", table)
	}
	var count int
	if err := database.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("count %s rows: %v", table, err)
	}
	return count
}

func findStep(t *testing.T, chainState *types.ChainState, stepID string) types.StepState {
	t.Helper()
	for _, step := range chainState.Steps {
		if step.StepID == stepID {
			return step
		}
	}
	t.Fatalf("step %q not found in chain state", stepID)
	return types.StepState{}
}

func findCompiledStep(t *testing.T, plan *types.ExecutionPlan, stepID string) types.CompiledStepPlan {
	t.Helper()
	for _, step := range plan.CompiledSteps {
		if step.ID == stepID {
			return step
		}
	}
	t.Fatalf("step %q not found in execution plan", stepID)
	return types.CompiledStepPlan{}
}

func simpleChainDefinitionJSON() string {
	return `{
  "kind": "chain",
  "name": "simple-chain",
  "description": "Simple test chain",
  "version": "1.0.0",
  "source": "db",
  "path": "catalog://chain/simple-chain/1",
  "entry": "research",
  "steps": [
    {
      "id": "research",
      "agent": "researcher",
      "skills": ["research"],
      "description": "Research the task",
      "transitions": {
        "success": "implement",
        "failure": { "retry": 1, "then": "handoff" }
      }
    },
    {
      "id": "implement",
      "agent": "builder",
      "skills": ["implement"],
      "description": "Implement the task",
      "transitions": {
        "success": "done",
        "failure": "handoff"
      }
    }
  ]
}`
}
