package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	budgetpkg "github.com/rluisb/lazyai/packages/orchestrator/internal/budget"
	oconfig "github.com/rluisb/lazyai/packages/orchestrator/internal/config"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
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

func TestStartChainAcceptsContextShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	cases := []struct {
		name           string
		context        any
		wantTextPrefix string
		wantHost       types.HostCli
		wantPrompt     string
	}{
		{name: "omitted", context: sentinelOmit, wantHost: types.HostOpenCode},
		{name: "null", context: nil, wantHost: types.HostOpenCode},
		{name: "empty_string", context: "", wantHost: types.HostOpenCode},
		{
			name: "object",
			context: map[string]any{
				"cliTool":     "claude-code",
				"rootContext": map[string]any{"prompt": "be careful"},
			},
			wantHost:   types.HostClaudeCode,
			wantPrompt: "be careful",
		},
		{
			name:       "stringified_json",
			context:    `{"cliTool":"claude-code","rootContext":{"prompt":"be careful"}}`,
			wantHost:   types.HostClaudeCode,
			wantPrompt: "be careful",
		},
		{name: "invalid_string", context: "not-json", wantTextPrefix: "Invalid start_chain"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			createSimpleChainCatalog(t, orchestrator)

			args := map[string]any{
				"chain": "simple-chain",
				"task":  "ship the feature",
			}
			if s, ok := tc.context.(string); !ok || s != sentinelOmit {
				args["context"] = tc.context
			}

			result, err := orchestrator.StartChain(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("start chain returned transport error: %v", err)
			}

			if tc.wantTextPrefix != "" {
				text := decodeToolText(t, result)
				if !strings.HasPrefix(text, tc.wantTextPrefix) {
					t.Fatalf("expected text to start with %q, got %q", tc.wantTextPrefix, text)
				}
				if count := countRows(t, orchestrator.DB, "chain_runs"); count != 0 {
					t.Fatalf("expected no chain run rows for invalid context, got %d", count)
				}
				return
			}

			var started struct {
				ChainID         string `json:"chainId"`
				State           string `json:"state"`
				ExecutionPlanID string `json:"executionPlanId"`
			}
			decodeToolResult(t, result, &started)
			if started.ChainID == "" || started.ExecutionPlanID == "" {
				t.Fatalf("expected chain id and plan id, got %+v", started)
			}

			plan, err := orchestrator.ExecutionPlans.LoadExecutionPlan(started.ExecutionPlanID)
			if err != nil {
				t.Fatalf("load execution plan: %v", err)
			}
			if plan.Cli.Host != tc.wantHost {
				t.Fatalf("expected Cli.Host=%q, got %q", tc.wantHost, plan.Cli.Host)
			}
			switch tc.wantPrompt {
			case "":
				if plan.RootContext != nil {
					t.Fatalf("expected nil RootContext, got %+v", plan.RootContext)
				}
			default:
				if plan.RootContext == nil || plan.RootContext.Prompt != tc.wantPrompt {
					t.Fatalf("expected RootContext.Prompt=%q, got %+v", tc.wantPrompt, plan.RootContext)
				}
			}
		})
	}
}

func TestAdvanceChainAcceptsJSONArgShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	validOutput := map[string]any{
		"summary":  "researched",
		"status":   "done",
		"findings": []any{},
	}
	validUsage := map[string]any{"totalTokens": 50}

	cases := []struct {
		name           string
		output         any
		usage          any
		wantTextPrefix string
		wantTokens     int
	}{
		{name: "object_object", output: validOutput, usage: validUsage, wantTokens: 50},
		{name: "stringified_output", output: `{"summary":"researched","status":"done","findings":[]}`, usage: validUsage, wantTokens: 50},
		{name: "stringified_usage", output: validOutput, usage: `{"totalTokens":50}`, wantTokens: 50},
		{name: "empty_string_output", output: "", usage: validUsage, wantTokens: 50},
		{name: "empty_string_usage", output: validOutput, usage: "", wantTokens: 0},
		{name: "null_both", output: nil, usage: nil, wantTokens: 0},
		{name: "omitted_both", output: sentinelOmit, usage: sentinelOmit, wantTokens: 0},
		{name: "invalid_output", output: "not-json", usage: sentinelOmit, wantTextPrefix: "Invalid advance_chain output"},
		{name: "invalid_usage", output: validOutput, usage: "not-json", wantTextPrefix: "Invalid advance_chain usage"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			createSimpleChainCatalog(t, orchestrator)
			started := startSimpleChain(t, orchestrator)

			args := map[string]any{
				"chainId": started.ChainID,
				"stepId":  "research",
				"outcome": "success",
			}
			if s, ok := tc.output.(string); !ok || s != sentinelOmit {
				args["output"] = tc.output
			}
			if s, ok := tc.usage.(string); !ok || s != sentinelOmit {
				args["usage"] = tc.usage
			}

			result, err := orchestrator.AdvanceChain(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("advance chain transport error: %v", err)
			}

			if tc.wantTextPrefix != "" {
				gotText := decodeToolText(t, result)
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected text to start with %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}

			var advanced struct {
				State    string `json:"state"`
				NextStep struct {
					StepID string `json:"stepId"`
				} `json:"nextStep"`
				Budget struct {
					Tokens struct {
						Consumed int `json:"consumed"`
					} `json:"tokens"`
				} `json:"budget"`
			}
			decodeToolResult(t, result, &advanced)
			if advanced.State != "running" || advanced.NextStep.StepID != "implement" {
				t.Fatalf("unexpected advance result: %+v", advanced)
			}
			if advanced.Budget.Tokens.Consumed != tc.wantTokens {
				t.Fatalf("expected tokens=%d, got %d", tc.wantTokens, advanced.Budget.Tokens.Consumed)
			}
		})
	}
}

func TestBuildTeamAcceptsBudgetShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	cases := []struct {
		name           string
		budget         any
		wantTextPrefix string
	}{
		{name: "object", budget: map[string]any{"id": "default", "scope": "team", "defaultActionOnLimit": "warn"}},
		{name: "stringified", budget: `{"id":"default","scope":"team","defaultActionOnLimit":"warn"}`},
		{name: "empty_string", budget: ""},
		{name: "null", budget: nil},
		{name: "omitted", budget: sentinelOmit},
		{name: "invalid", budget: "not-json", wantTextPrefix: "Invalid build_team budget"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			args := map[string]any{
				"team": "missing-team",
				"task": "do work",
			}
			if s, ok := tc.budget.(string); !ok || s != sentinelOmit {
				args["budget"] = tc.budget
			}
			result, err := orchestrator.BuildTeam(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("build team transport error: %v", err)
			}
			gotText := decodeToolText(t, result)
			if tc.wantTextPrefix != "" {
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected prefix %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}
			if strings.HasPrefix(gotText, "Invalid build_team input") || strings.HasPrefix(gotText, "Invalid build_team budget") {
				t.Fatalf("expected decode to succeed, got %q", gotText)
			}
		})
	}
}

func TestCompleteTeamTaskAcceptsJSONArgShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	validResult := map[string]any{"summary": "done"}
	validUsage := map[string]any{"totalTokens": 5}
	validError := map[string]any{"category": "validation", "code": "X", "message": "fail", "stepId": "s", "agent": "a", "skills": []any{}}

	cases := []struct {
		name           string
		result         any
		usage          any
		errorArg       any
		wantTextPrefix string
	}{
		{name: "all_object", result: validResult, usage: validUsage, errorArg: validError},
		{name: "stringified_result", result: `{"summary":"done"}`, usage: sentinelOmit, errorArg: sentinelOmit},
		{name: "stringified_usage", result: sentinelOmit, usage: `{"totalTokens":5}`, errorArg: sentinelOmit},
		{name: "stringified_error", result: sentinelOmit, usage: sentinelOmit, errorArg: `{"category":"validation","code":"X","message":"fail","stepId":"s","agent":"a","skills":[]}`},
		{name: "empty_strings", result: "", usage: "", errorArg: ""},
		{name: "all_null", result: nil, usage: nil, errorArg: nil},
		{name: "all_omitted", result: sentinelOmit, usage: sentinelOmit, errorArg: sentinelOmit},
		{name: "invalid_result", result: "not-json", usage: sentinelOmit, errorArg: sentinelOmit, wantTextPrefix: "Invalid complete_team_task result"},
		{name: "invalid_usage", result: sentinelOmit, usage: "not-json", errorArg: sentinelOmit, wantTextPrefix: "Invalid complete_team_task usage"},
		{name: "invalid_error", result: sentinelOmit, usage: sentinelOmit, errorArg: "not-json", wantTextPrefix: "Invalid complete_team_task error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			args := map[string]any{
				"teamId":  "missing-team",
				"taskId":  "missing-task",
				"outcome": "success",
			}
			if s, ok := tc.result.(string); !ok || s != sentinelOmit {
				args["result"] = tc.result
			}
			if s, ok := tc.usage.(string); !ok || s != sentinelOmit {
				args["usage"] = tc.usage
			}
			if s, ok := tc.errorArg.(string); !ok || s != sentinelOmit {
				args["error"] = tc.errorArg
			}
			result, err := orchestrator.CompleteTeamTask(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("complete team task transport error: %v", err)
			}
			gotText := decodeToolText(t, result)
			if tc.wantTextPrefix != "" {
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected prefix %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}
			if strings.HasPrefix(gotText, "Invalid complete_team_task input") ||
				strings.HasPrefix(gotText, "Invalid complete_team_task result") ||
				strings.HasPrefix(gotText, "Invalid complete_team_task usage") ||
				strings.HasPrefix(gotText, "Invalid complete_team_task error") {
				t.Fatalf("expected decode to succeed, got %q", gotText)
			}
		})
	}
}

func TestStartWorkflowAcceptsJSONArgShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	validBudget := map[string]any{"id": "default", "scope": "workflow", "defaultActionOnLimit": "warn"}
	validContext := map[string]any{"cliTool": "claude-code", "rootContext": map[string]any{"prompt": "x"}}

	cases := []struct {
		name           string
		budget         any
		contextArg     any
		wantTextPrefix string
	}{
		{name: "object_both", budget: validBudget, contextArg: validContext},
		{name: "stringified_budget", budget: `{"id":"default","scope":"workflow","defaultActionOnLimit":"warn"}`, contextArg: sentinelOmit},
		{name: "stringified_context", budget: sentinelOmit, contextArg: `{"cliTool":"claude-code","rootContext":{"prompt":"x"}}`},
		{name: "empty_strings", budget: "", contextArg: ""},
		{name: "all_null", budget: nil, contextArg: nil},
		{name: "all_omitted", budget: sentinelOmit, contextArg: sentinelOmit},
		{name: "invalid_budget", budget: "not-json", contextArg: sentinelOmit, wantTextPrefix: "Invalid start_workflow budget"},
		{name: "invalid_context", budget: sentinelOmit, contextArg: "not-json", wantTextPrefix: "Invalid start_workflow context"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			args := map[string]any{
				"workflow": "missing-workflow",
				"task":     "do work",
			}
			if s, ok := tc.budget.(string); !ok || s != sentinelOmit {
				args["budget"] = tc.budget
			}
			if s, ok := tc.contextArg.(string); !ok || s != sentinelOmit {
				args["context"] = tc.contextArg
			}
			result, err := orchestrator.StartWorkflow(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("start workflow transport error: %v", err)
			}
			gotText := decodeToolText(t, result)
			if tc.wantTextPrefix != "" {
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected prefix %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}
			if strings.HasPrefix(gotText, "Invalid start_workflow input") ||
				strings.HasPrefix(gotText, "Invalid start_workflow budget") ||
				strings.HasPrefix(gotText, "Invalid start_workflow context") {
				t.Fatalf("expected decode to succeed, got %q", gotText)
			}
		})
	}
}

func TestAdvanceWorkflowAcceptsRecoveryShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	validRecovery := map[string]any{"type": "skip", "reason": "noop"}

	cases := []struct {
		name           string
		recovery       any
		wantTextPrefix string
	}{
		{name: "object", recovery: validRecovery},
		{name: "stringified", recovery: `{"type":"skip","reason":"noop"}`},
		{name: "empty_string", recovery: ""},
		{name: "null", recovery: nil},
		{name: "omitted", recovery: sentinelOmit},
		{name: "invalid", recovery: "not-json", wantTextPrefix: "Invalid advance_workflow recovery"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			args := map[string]any{"workflowId": "missing-workflow"}
			if s, ok := tc.recovery.(string); !ok || s != sentinelOmit {
				args["recovery"] = tc.recovery
			}
			result, err := orchestrator.AdvanceWorkflow(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("advance workflow transport error: %v", err)
			}
			gotText := decodeToolText(t, result)
			if tc.wantTextPrefix != "" {
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected prefix %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}
			if strings.HasPrefix(gotText, "Invalid advance_workflow input") ||
				strings.HasPrefix(gotText, "Invalid advance_workflow recovery") {
				t.Fatalf("expected decode to succeed, got %q", gotText)
			}
		})
	}
}

func TestEnqueueJobAcceptsPayloadShapes(t *testing.T) {
	const sentinelOmit = "<<omit>>"

	cases := []struct {
		name           string
		payload        any
		wantPayload    map[string]any
		wantTextPrefix string
	}{
		{name: "object", payload: map[string]any{"key": "value"}, wantPayload: map[string]any{"key": "value"}},
		{name: "stringified", payload: `{"key":"value"}`, wantPayload: map[string]any{"key": "value"}},
		{name: "empty_string", payload: "", wantPayload: nil},
		{name: "null", payload: nil, wantPayload: nil},
		{name: "omitted", payload: sentinelOmit, wantPayload: nil},
		{name: "invalid", payload: "not-json", wantTextPrefix: "Invalid enqueue_job payload"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchestrator := newTestOrchestrator(t)
			args := map[string]any{"jobType": "test-job"}
			if s, ok := tc.payload.(string); !ok || s != sentinelOmit {
				args["payload"] = tc.payload
			}
			result, err := orchestrator.EnqueueJob(context.Background(), toolRequest(args))
			if err != nil {
				t.Fatalf("enqueue job transport error: %v", err)
			}
			if tc.wantTextPrefix != "" {
				gotText := decodeToolText(t, result)
				if !strings.HasPrefix(gotText, tc.wantTextPrefix) {
					t.Fatalf("expected prefix %q, got %q", tc.wantTextPrefix, gotText)
				}
				return
			}
			var job struct {
				ID      string         `json:"id"`
				Payload map[string]any `json:"payload"`
			}
			decodeToolResult(t, result, &job)
			if job.ID == "" {
				t.Fatalf("expected job id, got %+v", job)
			}
			if !reflect.DeepEqual(job.Payload, tc.wantPayload) {
				t.Fatalf("expected payload %+v, got %+v", tc.wantPayload, job.Payload)
			}
		})
	}
}

func TestAdvanceChainRejectsPendingStepAndPreservesState(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)
	before, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
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
	after, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
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
	before, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
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
	after, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
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
	chainState, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
	if err != nil {
		t.Fatalf("load escalated chain: %v", err)
	}
	if chainState.State != types.ChainRunning || chainState.CurrentStepID != "research" {
		t.Fatalf("expected escalated run to remain resumable/running, got state=%q current=%q", chainState.State, chainState.CurrentStepID)
	}
	if step := findStep(t, chainState, "research"); step.Agent != "senior-builder" || step.State != types.StepRunning {
		t.Fatalf("target agent was not persisted in state: %+v", step)
	}
	plan, err := orchestrator.ExecutionPlans.LoadExecutionPlan(chainState.ExecutionPlanID)
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

	chainState, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
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

func TestDefaultRuntimeConfigIsNative(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	if orchestrator.Runtime.ExecutionMode != types.ExecutionNative {
		t.Fatalf("expected default execution mode native, got %q", orchestrator.Runtime.ExecutionMode)
	}
}

func TestInvokeAgentNativePreservesComposedSpecShape(t *testing.T) {
	orchestrator := newTestOrchestrator(t)

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent": "builder",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	var spec types.ComposedAgentSpec
	decodeToolResult(t, result, &spec)
	if spec.ID != "builder" || spec.Base != "builder" || spec.Model != "sonnet" {
		t.Fatalf("unexpected native invoke_agent spec identity/model: %+v", spec)
	}
	if spec.Prompt != "Task: ship the feature\nExecute as the builder agent." {
		t.Fatalf("native invoke_agent prompt shape changed: %q", spec.Prompt)
	}
}

func TestAgentInvokerOptionRoutesInvokeAgentThroughPort(t *testing.T) {
	invoker := &recordingAgentInvoker{
		result: &domain.InvocationResult{
			ExecutionMode: types.ExecutionNative,
			Spec: &types.ComposedAgentSpec{
				ID:     "custom-builder",
				Base:   "custom-builder",
				Model:  "test-model",
				Prompt: "custom prompt",
			},
		},
	}
	orchestrator := newTestOrchestratorWithOptions(t, WithAgentInvoker(invoker))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent":   "builder",
		"task":    "ship the feature",
		"cliTool": "opencode",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}

	if invoker.request.Agent != "builder" || invoker.request.Task != "ship the feature" || invoker.request.CliTool != "opencode" {
		t.Fatalf("invoke_agent did not route request through agent invoker port: %+v", invoker.request)
	}
	var spec types.ComposedAgentSpec
	decodeToolResult(t, result, &spec)
	if spec.ID != "custom-builder" || spec.Model != "test-model" || spec.Prompt != "custom prompt" {
		t.Fatalf("invoke_agent did not return port result spec: %+v", spec)
	}
}

func TestStartChainCompiledStepsRemainNativeByDefault(t *testing.T) {
	orchestrator := newTestOrchestrator(t)
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)
	chainState, err := orchestrator.ChainStates.LoadChainState(started.ChainID)
	if err != nil {
		t.Fatalf("load chain state: %v", err)
	}
	plan, err := orchestrator.ExecutionPlans.LoadExecutionPlan(chainState.ExecutionPlanID)
	if err != nil {
		t.Fatalf("load execution plan: %v", err)
	}
	if len(plan.CompiledSteps) == 0 {
		t.Fatalf("expected compiled chain steps")
	}
	for _, step := range plan.CompiledSteps {
		if step.ExecutionMode != types.ExecutionNative {
			t.Fatalf("compiled step %q should remain native by default, got %q", step.ID, step.ExecutionMode)
		}
	}
}

func TestInvokeAgentA2AModeReturnsNotConfiguredError(t *testing.T) {
	orchestrator := newTestOrchestratorWithOptions(t, WithRuntimeConfig(RuntimeConfig{ExecutionMode: types.ExecutionA2A}))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent": "builder",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	message := decodeToolText(t, result)
	for _, want := range []string{"A2A execution mode", "not configured", "orchestrator config"} {
		if !strings.Contains(message, want) {
			t.Fatalf("A2A not-configured message %q missing %q", message, want)
		}
	}
}

func TestInvokeAgentA2AConfiguredCliToolTargetReturnsPhase3Error(t *testing.T) {
	orchestrator := newTestOrchestratorWithOptions(t, WithRuntimeConfig(RuntimeConfig{
		ExecutionMode: types.ExecutionA2A,
		A2AConfig:     testRuntimeA2AConfig(types.ExecutionA2A),
	}))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent":   "builder",
		"task":    "ship the feature",
		"cliTool": "opencode",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	message := decodeToolText(t, result)
	for _, want := range []string{"A2A provider", "builder", "not implemented until Phase 3"} {
		if !strings.Contains(message, want) {
			t.Fatalf("configured A2A message %q missing %q", message, want)
		}
	}
}

func TestInvokeAgentA2AConfiguredCliToolMismatchFailsBeforePhase3(t *testing.T) {
	orchestrator := newTestOrchestratorWithOptions(t, WithRuntimeConfig(RuntimeConfig{
		ExecutionMode: types.ExecutionA2A,
		A2AConfig:     testRuntimeA2AConfig(types.ExecutionA2A),
	}))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent":   "builder",
		"task":    "ship the feature",
		"cliTool": "copilot",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	if message := decodeToolText(t, result); !strings.Contains(message, "cliTool") {
		t.Fatalf("expected cliTool targeting error, got %q", message)
	}
}

func TestInvokeAgentHybridWithoutA2AConfigFallsBackToNative(t *testing.T) {
	orchestrator := newTestOrchestratorWithOptions(t, WithRuntimeConfig(RuntimeConfig{ExecutionMode: types.ExecutionHybrid}))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent": "builder",
		"task":  "ship the feature",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	var spec types.ComposedAgentSpec
	decodeToolResult(t, result, &spec)
	if spec.ID != "builder" || spec.Base != "builder" {
		t.Fatalf("hybrid without A2A config should fall back to native spec, got %+v", spec)
	}
	if spec.Prompt != "Task: ship the feature\nExecute as the builder agent." {
		t.Fatalf("hybrid fallback prompt shape changed: %q", spec.Prompt)
	}
}

func TestInvokeAgentHybridWithConfiguredA2ATargetFallsBackToNative(t *testing.T) {
	orchestrator := newTestOrchestratorWithOptions(t, WithRuntimeConfig(RuntimeConfig{
		ExecutionMode: types.ExecutionHybrid,
		A2AConfig:     testRuntimeA2AConfig(types.ExecutionHybrid),
	}))

	result, err := orchestrator.InvokeAgent(context.Background(), toolRequest(map[string]any{
		"agent":   "builder",
		"task":    "ship the feature",
		"cliTool": "opencode",
	}))
	if err != nil {
		t.Fatalf("invoke agent returned transport error: %v", err)
	}
	var spec types.ComposedAgentSpec
	decodeToolResult(t, result, &spec)
	if spec.ID != "builder" || spec.Base != "builder" || spec.Prompt != "Task: ship the feature\nExecute as the builder agent." {
		t.Fatalf("hybrid configured fallback should preserve native shape, got %+v", spec)
	}
}

func TestNewRuntimeConfigRejectsInvalidExecutionMode(t *testing.T) {
	if _, err := NewRuntimeConfig("bogus"); err == nil {
		t.Fatalf("expected invalid execution mode error")
	}
}

func TestBudgetTrackerOptionRoutesUsageAndEvaluation(t *testing.T) {
	tracker := &recordingBudgetTracker{
		delegate: budgetpkg.NewTracker(),
		evaluation: types.BudgetEvaluation{
			Overall:           types.HealthWarning,
			RecommendedAction: "warn",
		},
	}
	orchestrator := newTestOrchestratorWithOptions(t, WithBudgetTracker(tracker))
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	advanceResult, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
		"chainId": started.ChainID,
		"stepId":  "research",
		"outcome": "success",
		"usage":   map[string]any{"totalTokens": 50},
	}))
	if err != nil {
		t.Fatalf("advance chain returned transport error: %v", err)
	}
	var advanced struct {
		Budget types.BudgetState `json:"budget"`
	}
	decodeToolResult(t, advanceResult, &advanced)
	if len(tracker.updatedStepIDs) != 1 || tracker.updatedStepIDs[0] != "research" || advanced.Budget.Tokens.Consumed != 50 {
		t.Fatalf("budget tracker update was not routed through port: steps=%v budget=%+v", tracker.updatedStepIDs, advanced.Budget)
	}

	budgetResult, err := orchestrator.GetBudget(context.Background(), toolRequest(map[string]any{
		"runId": started.ChainID,
		"kind":  "chain",
	}))
	if err != nil {
		t.Fatalf("get budget returned transport error: %v", err)
	}
	var budgetView struct {
		Health types.BudgetEvaluation `json:"health"`
	}
	decodeToolResult(t, budgetResult, &budgetView)
	if tracker.evaluateCalls != 1 || budgetView.Health.Overall != types.HealthWarning {
		t.Fatalf("budget tracker evaluation was not routed through port: calls=%d health=%+v", tracker.evaluateCalls, budgetView.Health)
	}
}

func TestChainStateStoreOptionRoutesLifecyclePersistence(t *testing.T) {
	store := newRecordingChainStateStore()
	orchestrator := newTestOrchestratorWithOptions(t, WithChainStateStore(store))
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	if len(store.savedIDs) != 1 || store.savedIDs[0] != started.ChainID {
		t.Fatalf("start_chain did not save through chain state store port: saved=%v chain=%q", store.savedIDs, started.ChainID)
	}

	_, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
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
		t.Fatalf("advance chain returned transport error: %v", err)
	}
	if len(store.loadedIDs) != 1 || store.loadedIDs[0] != started.ChainID {
		t.Fatalf("advance_chain did not load through chain state store port: loaded=%v chain=%q", store.loadedIDs, started.ChainID)
	}
	if len(store.savedIDs) != 2 || store.savedIDs[1] != started.ChainID {
		t.Fatalf("advance_chain did not save through chain state store port: saved=%v chain=%q", store.savedIDs, started.ChainID)
	}
}

func TestTeamStateStoreOptionRoutesLifecyclePersistence(t *testing.T) {
	store := newRecordingTeamStateStore()
	orchestrator := newTestOrchestratorWithOptions(t, WithTeamStateStore(store))
	createSimpleTeamCatalog(t, orchestrator)
	started := startSimpleTeam(t, orchestrator)

	if len(store.savedIDs) != 1 || store.savedIDs[0] != started.TeamID {
		t.Fatalf("build_team did not save through team state store port: saved=%v team=%q", store.savedIDs, started.TeamID)
	}

	_, err := orchestrator.AssignTeamTask(context.Background(), toolRequest(map[string]any{
		"teamId":   started.TeamID,
		"taskId":   "researcher-0",
		"assignee": "agent-a",
	}))
	if err != nil {
		t.Fatalf("assign team task returned transport error: %v", err)
	}
	if len(store.loadedIDs) != 1 || store.loadedIDs[0] != started.TeamID {
		t.Fatalf("assign_team_task did not load through team state store port: loaded=%v team=%q", store.loadedIDs, started.TeamID)
	}
	if len(store.savedIDs) != 2 || store.savedIDs[1] != started.TeamID {
		t.Fatalf("assign_team_task did not save through team state store port: saved=%v team=%q", store.savedIDs, started.TeamID)
	}
}

func TestWorkflowStateStoreOptionRoutesLifecyclePersistence(t *testing.T) {
	store := newRecordingWorkflowStateStore()
	orchestrator := newTestOrchestratorWithOptions(t, WithWorkflowStateStore(store))
	createSimpleWorkflowCatalog(t, orchestrator)
	started := startSimpleWorkflow(t, orchestrator)

	if len(store.savedIDs) != 1 || store.savedIDs[0] != started.WorkflowID {
		t.Fatalf("start_workflow did not save through workflow state store port: saved=%v workflow=%q", store.savedIDs, started.WorkflowID)
	}

	_, err := orchestrator.AdvanceWorkflow(context.Background(), toolRequest(map[string]any{
		"workflowId": started.WorkflowID,
		"outcome":    "approved",
	}))
	if err != nil {
		t.Fatalf("advance workflow returned transport error: %v", err)
	}
	if len(store.loadedIDs) != 1 || store.loadedIDs[0] != started.WorkflowID {
		t.Fatalf("advance_workflow did not load through workflow state store port: loaded=%v workflow=%q", store.loadedIDs, started.WorkflowID)
	}
	if len(store.savedIDs) != 2 || store.savedIDs[1] != started.WorkflowID {
		t.Fatalf("advance_workflow did not save through workflow state store port: saved=%v workflow=%q", store.savedIDs, started.WorkflowID)
	}
}

func TestExecutionPlanStoreOptionRoutesLifecyclePersistence(t *testing.T) {
	store := newRecordingExecutionPlanStore()
	orchestrator := newTestOrchestratorWithOptions(t, WithExecutionPlanStore(store))
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	if len(store.savedIDs) != 1 || store.savedIDs[0] == "" {
		t.Fatalf("start_chain did not save through execution plan store port: saved=%v", store.savedIDs)
	}
	planID := store.savedIDs[0]

	_, err := orchestrator.AdvanceChain(context.Background(), toolRequest(map[string]any{
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
		t.Fatalf("advance chain returned transport error: %v", err)
	}
	if len(store.loadedIDs) != 1 || store.loadedIDs[0] != planID {
		t.Fatalf("advance_chain did not load through execution plan store port: loaded=%v plan=%q", store.loadedIDs, planID)
	}
}

func TestHandoffStoreOptionRoutesHandoffPersistence(t *testing.T) {
	store := &recordingHandoffStore{}
	orchestrator := newTestOrchestratorWithOptions(t, WithHandoffStore(store))
	createSimpleChainCatalog(t, orchestrator)
	started := startSimpleChain(t, orchestrator)

	result, err := orchestrator.Handoff(context.Background(), toolRequest(map[string]any{
		"runId":   started.ChainID,
		"kind":    "chain",
		"summary": "Continue through port",
	}))
	if err != nil {
		t.Fatalf("handoff returned transport error: %v", err)
	}
	var handoffResult struct {
		HandoffID string `json:"handoffId"`
		Path      string `json:"path"`
	}
	decodeToolResult(t, result, &handoffResult)

	if len(store.saved) != 1 {
		t.Fatalf("handoff did not save through handoff store port: saved=%d", len(store.saved))
	}
	saved := store.saved[0]
	if saved.ID != handoffResult.HandoffID || saved.RunID != started.ChainID || saved.Kind != types.RunKindChain || saved.Summary != "Continue through port" || !saved.Resumable {
		t.Fatalf("unexpected handoff saved through port: result=%+v saved=%+v", handoffResult, saved)
	}
	if !strings.HasPrefix(handoffResult.Path, handoffPathURIPrefix) {
		t.Fatalf("unexpected handoff path: %q", handoffResult.Path)
	}
}

func toolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func newTestOrchestrator(t *testing.T) *Orchestrator {
	return newTestOrchestratorWithOptions(t)
}

func newTestOrchestratorWithOptions(t *testing.T, options ...OrchestratorOption) *Orchestrator {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewOrchestrator(database, NewScopeContext("project", "/tmp/project", ""), options...)
}

func testRuntimeA2AConfig(mode types.ExecutionMode) *oconfig.Config {
	enabled := true
	return &oconfig.Config{
		Version:   1,
		Execution: oconfig.ExecutionConfig{Mode: mode},
		Providers: map[string]oconfig.ProviderConfig{
			"local": {Endpoint: "https://a2a.example.test/rpc", Auth: oconfig.AuthConfig{Type: oconfig.AuthNone}},
		},
		Agents: map[string]oconfig.AgentConfig{
			"builder": {Provider: "local", Enabled: &enabled, Tools: []string{"opencode"}},
		},
	}
}

func createSimpleChainCatalog(t *testing.T, orchestrator *Orchestrator) {
	t.Helper()
	created, err := orchestrator.Catalog.CreateVersion(domain.CreateCatalogVersionInput{
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

func createSimpleTeamCatalog(t *testing.T, orchestrator *Orchestrator) {
	t.Helper()
	created, err := orchestrator.Catalog.CreateVersion(domain.CreateCatalogVersionInput{
		Kind:      "team",
		Name:      "simple-team",
		Body:      simpleTeamDefinitionJSON(),
		SetActive: true,
	})
	if err != nil {
		t.Fatalf("create team catalog version: %v", err)
	}
	if created.AlreadyExists {
		t.Fatalf("expected new team catalog version, got duplicate: %+v", created)
	}
}

func startSimpleTeam(t *testing.T, orchestrator *Orchestrator) struct {
	TeamID string                   `json:"teamId"`
	State  types.TeamLifecycleState `json:"state"`
	Tasks  []types.TeamTaskState    `json:"tasks"`
} {
	t.Helper()
	result, err := orchestrator.BuildTeam(context.Background(), toolRequest(map[string]any{
		"team": "simple-team",
		"task": "ship the feature",
	}))
	if err != nil {
		t.Fatalf("start simple team: %v", err)
	}
	var started struct {
		TeamID string                   `json:"teamId"`
		State  types.TeamLifecycleState `json:"state"`
		Tasks  []types.TeamTaskState    `json:"tasks"`
	}
	decodeToolResult(t, result, &started)
	if started.TeamID == "" {
		t.Fatalf("expected team id in build result")
	}
	return started
}

func createSimpleWorkflowCatalog(t *testing.T, orchestrator *Orchestrator) {
	t.Helper()
	created, err := orchestrator.Catalog.CreateVersion(domain.CreateCatalogVersionInput{
		Kind:      "workflow",
		Name:      "simple-workflow",
		Body:      simpleWorkflowDefinitionJSON(),
		SetActive: true,
	})
	if err != nil {
		t.Fatalf("create workflow catalog version: %v", err)
	}
	if created.AlreadyExists {
		t.Fatalf("expected new workflow catalog version, got duplicate: %+v", created)
	}
}

func startSimpleWorkflow(t *testing.T, orchestrator *Orchestrator) struct {
	WorkflowID string                       `json:"workflowId"`
	State      types.WorkflowLifecycleState `json:"state"`
	Phases     []types.WorkflowPhaseState   `json:"phases"`
	Budget     types.BudgetState            `json:"budget"`
	PlanID     string                       `json:"planId"`
} {
	t.Helper()
	result, err := orchestrator.StartWorkflow(context.Background(), toolRequest(map[string]any{
		"workflow": "simple-workflow",
		"task":     "ship the feature",
	}))
	if err != nil {
		t.Fatalf("start simple workflow: %v", err)
	}
	var started struct {
		WorkflowID string                       `json:"workflowId"`
		State      types.WorkflowLifecycleState `json:"state"`
		Phases     []types.WorkflowPhaseState   `json:"phases"`
		Budget     types.BudgetState            `json:"budget"`
		PlanID     string                       `json:"planId"`
	}
	decodeToolResult(t, result, &started)
	if started.WorkflowID == "" {
		t.Fatalf("expected workflow id in start result")
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

func simpleTeamDefinitionJSON() string {
	return `{
  "kind": "team",
  "name": "simple-team",
  "description": "Simple test team",
  "version": "1.0.0",
  "source": "db",
  "path": "catalog://team/simple-team/1",
  "parallel": [
    {"role":"researcher","agent":"researcher","skills":["research"],"focus":"research"}
  ],
  "synthesize": {"agent":"builder","description":"Synthesize results"}
}`
}

func simpleWorkflowDefinitionJSON() string {
	return `{
  "kind": "workflow",
  "name": "simple-workflow",
  "description": "Simple test workflow",
  "version": "1.0.0",
  "source": "db",
  "path": "catalog://workflow/simple-workflow/1",
  "entry": "approve",
  "phases": [
    {"id":"approve","kind":"gate","gate":"human","prompt":"Approve completion"},
    {"id":"done","kind":"terminal"}
  ]
}`
}

type recordingAgentInvoker struct {
	request domain.InvocationRequest
	result  *domain.InvocationResult
	err     error
}

func (i *recordingAgentInvoker) InvokeAgent(ctx context.Context, req domain.InvocationRequest) (*domain.InvocationResult, error) {
	i.request = req
	if i.err != nil {
		return nil, i.err
	}
	return i.result, nil
}

type recordingBudgetTracker struct {
	delegate       ports.BudgetTracker
	evaluation     types.BudgetEvaluation
	updates        []string
	retries        []string
	updatedStepIDs []string
	evaluateCalls  int
}

func (t *recordingBudgetTracker) Update(state *types.BudgetState, stepID string, usage *types.StepUsage) {
	t.updates = append(t.updates, stepID)
	t.updatedStepIDs = append(t.updatedStepIDs, stepID)
	if t.delegate != nil {
		t.delegate.Update(state, stepID, usage)
	}
}

func (t *recordingBudgetTracker) IncrementRetries(state *types.BudgetState, stepID string) {
	t.retries = append(t.retries, stepID)
	if t.delegate != nil {
		t.delegate.IncrementRetries(state, stepID)
	}
}

func (t *recordingBudgetTracker) Evaluate(state *types.BudgetState, policy *types.BudgetPolicy) types.BudgetEvaluation {
	t.evaluateCalls++
	if t.evaluation.Overall != "" {
		return t.evaluation
	}
	if t.delegate != nil {
		return t.delegate.Evaluate(state, policy)
	}
	return types.BudgetEvaluation{}
}

type recordingExecutionPlanStore struct {
	plans     map[string]*types.ExecutionPlan
	savedIDs  []string
	loadedIDs []string
}

func newRecordingExecutionPlanStore() *recordingExecutionPlanStore {
	return &recordingExecutionPlanStore{plans: make(map[string]*types.ExecutionPlan)}
}

func (s *recordingExecutionPlanStore) SaveExecutionPlan(plan *types.ExecutionPlan) error {
	s.savedIDs = append(s.savedIDs, plan.ID)
	clone, err := cloneExecutionPlan(plan)
	if err != nil {
		return err
	}
	s.plans[plan.ID] = clone
	return nil
}

func (s *recordingExecutionPlanStore) LoadExecutionPlan(id string) (*types.ExecutionPlan, error) {
	s.loadedIDs = append(s.loadedIDs, id)
	plan, ok := s.plans[id]
	if !ok {
		return nil, fmt.Errorf("execution plan not found: %s", id)
	}
	return cloneExecutionPlan(plan)
}

func cloneExecutionPlan(plan *types.ExecutionPlan) (*types.ExecutionPlan, error) {
	encoded, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}
	var clone types.ExecutionPlan
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

type recordingHandoffStore struct {
	saved []*types.HandoffDocument
}

func (s *recordingHandoffStore) SaveHandoffDocument(doc *types.HandoffDocument) error {
	s.saved = append(s.saved, doc)
	return nil
}

type recordingChainStateStore struct {
	states    map[string]*types.ChainState
	savedIDs  []string
	loadedIDs []string
}

func newRecordingChainStateStore() *recordingChainStateStore {
	return &recordingChainStateStore{states: make(map[string]*types.ChainState)}
}

func (s *recordingChainStateStore) SaveChainState(projectRoot string, chainState *types.ChainState) error {
	s.savedIDs = append(s.savedIDs, chainState.ChainID)
	clone, err := cloneChainState(chainState)
	if err != nil {
		return err
	}
	s.states[chainState.ChainID] = clone
	return nil
}

func (s *recordingChainStateStore) LoadChainState(id string) (*types.ChainState, error) {
	s.loadedIDs = append(s.loadedIDs, id)
	chainState, ok := s.states[id]
	if !ok {
		return nil, fmt.Errorf("chain state not found: %s", id)
	}
	return cloneChainState(chainState)
}

func cloneChainState(chainState *types.ChainState) (*types.ChainState, error) {
	encoded, err := json.Marshal(chainState)
	if err != nil {
		return nil, err
	}
	var clone types.ChainState
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

type recordingTeamStateStore struct {
	states    map[string]*types.TeamState
	savedIDs  []string
	loadedIDs []string
}

func newRecordingTeamStateStore() *recordingTeamStateStore {
	return &recordingTeamStateStore{states: make(map[string]*types.TeamState)}
}

func (s *recordingTeamStateStore) SaveTeamState(projectRoot string, teamState *types.TeamState) error {
	s.savedIDs = append(s.savedIDs, teamState.TeamID)
	clone, err := cloneTeamState(teamState)
	if err != nil {
		return err
	}
	s.states[teamState.TeamID] = clone
	return nil
}

func (s *recordingTeamStateStore) LoadTeamState(id string) (*types.TeamState, error) {
	s.loadedIDs = append(s.loadedIDs, id)
	teamState, ok := s.states[id]
	if !ok {
		return nil, fmt.Errorf("team state not found: %s", id)
	}
	return cloneTeamState(teamState)
}

func cloneTeamState(teamState *types.TeamState) (*types.TeamState, error) {
	encoded, err := json.Marshal(teamState)
	if err != nil {
		return nil, err
	}
	var clone types.TeamState
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

type recordingWorkflowStateStore struct {
	states    map[string]*types.WorkflowState
	savedIDs  []string
	loadedIDs []string
}

func newRecordingWorkflowStateStore() *recordingWorkflowStateStore {
	return &recordingWorkflowStateStore{states: make(map[string]*types.WorkflowState)}
}

func (s *recordingWorkflowStateStore) SaveWorkflowState(projectRoot string, workflowState *types.WorkflowState) error {
	s.savedIDs = append(s.savedIDs, workflowState.WorkflowID)
	clone, err := cloneWorkflowState(workflowState)
	if err != nil {
		return err
	}
	s.states[workflowState.WorkflowID] = clone
	return nil
}

func (s *recordingWorkflowStateStore) LoadWorkflowState(id string) (*types.WorkflowState, error) {
	s.loadedIDs = append(s.loadedIDs, id)
	workflowState, ok := s.states[id]
	if !ok {
		return nil, fmt.Errorf("workflow state not found: %s", id)
	}
	return cloneWorkflowState(workflowState)
}

func cloneWorkflowState(workflowState *types.WorkflowState) (*types.WorkflowState, error) {
	encoded, err := json.Marshal(workflowState)
	if err != nil {
		return nil, err
	}
	var clone types.WorkflowState
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}
