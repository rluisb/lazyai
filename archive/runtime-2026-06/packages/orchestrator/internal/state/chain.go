package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// ──────────────────────────────────────────────────────────────
// Public API
// ──────────────────────────────────────────────────────────────

// CreateChainState builds the initial ChainState from an ExecutionPlan.
func CreateChainState(plan *types.ExecutionPlan) *types.ChainState {
	steps := make([]types.StepState, 0, len(plan.CompiledSteps))
	for i, cs := range plan.CompiledSteps {
		state := types.StepPending
		attempts := 0
		if cs.ID == plan.Entrypoint {
			state = types.StepRunning
			attempts = 1
		}
		s := types.StepState{
			StepID:      cs.ID,
			Order:       i,
			Agent:       cs.Agent,
			Skills:      cs.Skills,
			StepType:    cs.StepType,
			DomainSkill: cs.DomainSkill,
			ModeSkill:   cs.ModeSkill,
			State:       state,
			Attempts:    attempts,
			MaxRetries:  getMaxRetries(cs.Transitions),
			Usage:       types.StepUsage{},
		}
		if state == types.StepRunning {
			s.StartedAt = plan.CreatedAt
		}
		steps = append(steps, s)
	}

	return &types.ChainState{
		ChainID:           uuid.NewString(),
		DefinitionName:    plan.Definition.Name,
		DefinitionVersion: plan.Definition.Version,
		ExecutionPlanID:   plan.ID,
		State:             types.ChainRunning,
		Task:              plan.Task,
		CurrentStepID:     plan.Entrypoint,
		EntryStepID:       plan.Entrypoint,
		Steps:             steps,
		CompletedStepIDs:  []string{},
		Budget:            createEmptyBudget(plan),
		CreatedAt:         plan.CreatedAt,
		UpdatedAt:         plan.CreatedAt,
	}
}

// AdvanceInput is the input to AdvanceChain.
type AdvanceInput struct {
	State           types.ChainState
	Plan            types.ExecutionPlan
	StepID          string
	Outcome         string
	Output          map[string]any
	ValidationError *types.StructuredError
}

// AdvanceResult is the output from AdvanceChain.
type AdvanceResult struct {
	types.AdvanceChainResult
	StateSnapshot types.ChainState
}

// AdvanceChain advances a chain to its next step based on the outcome.
func AdvanceChain(input AdvanceInput) (AdvanceResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	state := cloneChainState(&input.State)
	step, err := requireStep(state, input.StepID)
	if err != nil {
		return AdvanceResult{}, err
	}
	compiledStep, err := requireCompiledStep(&input.Plan, input.StepID)
	if err != nil {
		return AdvanceResult{}, err
	}

	// Gate decision
	if state.State == types.ChainGated && step.Gate != nil && step.Gate.Status == types.GatePending {
		return applyGateDecision(state, &input.Plan, compiledStep, step, input.Outcome, now, input.Output)
	}

	// Only active steps can be advanced
	if step.State != types.StepRunning && step.State != types.StepEscalated && step.State != types.StepRetrying {
		return AdvanceResult{}, fmt.Errorf("step %q is not active", step.StepID)
	}

	// Validation error
	if input.ValidationError != nil {
		step.State = types.StepFailed
		step.Error = input.ValidationError
		step.LastOutcome = input.Outcome
		state.UpdatedAt = now

		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State:    state.State,
				NextStep: toChainStepStatus(&input.Plan, step.StepID, step.State),
				Recovery: &input.ValidationError.SuggestedRecovery,
				Budget:   state.Budget,
				Error:    input.ValidationError,
			},
			StateSnapshot: *state,
		}, nil
	}

	// Store output
	if input.Output != nil {
		step.Output = input.Output
		step.OutputValid = true
	}

	// Gate on success
	if compiledStep.Gate != "" && input.Outcome == "success" {
		step.State = types.StepCompleted
		step.CompletedAt = now
		step.LastOutcome = input.Outcome
		gate := buildPendingGate(compiledStep.Gate, compiledStep.ID, state)
		step.Gate = gate
		markCompleted(state, step.StepID)
		state.State = types.ChainGated
		state.UpdatedAt = now

		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State:    state.State,
				NextStep: toChainStepStatus(&input.Plan, step.StepID, step.State),
				Gate:     gate,
				Budget:   state.Budget,
			},
			StateSnapshot: *state,
		}, nil
	}

	// Failure transition
	if input.Outcome == "failure" {
		return handleFailureTransition(state, compiledStep, step, now, &input.Plan)
	}

	// Normal transition
	transition, ok := compiledStep.Transitions[input.Outcome]
	if !ok {
		return AdvanceResult{}, fmt.Errorf("outcome %q is not valid for step %q", input.Outcome, input.StepID)
	}
	if transition.Then == "" && transition.Retry == 0 {
		return AdvanceResult{}, fmt.Errorf("outcome %q has no string transition for step %q", input.Outcome, input.StepID)
	}

	step.State = types.StepCompleted
	step.CompletedAt = now
	step.LastOutcome = input.Outcome
	markCompleted(state, step.StepID)

	// Only string transitions are valid here (no retry objects for non-failure outcomes)
	if transition.Then != "" {
		return moveToTarget(state, &input.Plan, transition.Then, now)
	}
	return AdvanceResult{}, fmt.Errorf("invalid transition for outcome %q on step %q", input.Outcome, input.StepID)
}

// RetryChainStep retries a failed chain step.
func RetryChainStep(state *types.ChainState, plan *types.ExecutionPlan, stepID string) (*types.ChainState, int, error) {
	next := cloneChainState(state)
	step, err := requireStep(next, stepID)
	if err != nil {
		return nil, 0, err
	}
	if _, err := requireCompiledStep(plan, stepID); err != nil {
		return nil, 0, err
	}

	if step.State != types.StepFailed {
		return nil, 0, fmt.Errorf("step %q is not eligible for retry", stepID)
	}

	remaining := getAttemptsRemaining(step)
	if remaining <= 0 {
		return nil, 0, fmt.Errorf("step %q has no retries remaining", stepID)
	}

	step.State = types.StepRunning
	step.Attempts++
	step.StartedAt = time.Now().UTC().Format(time.RFC3339)
	step.CompletedAt = ""
	step.Error = nil
	step.Gate = nil
	step.Output = nil
	step.OutputValid = false
	step.LastOutcome = "retry"
	next.State = types.ChainRunning
	next.CurrentStepID = stepID
	next.UpdatedAt = step.StartedAt

	return next, getAttemptsRemaining(step), nil
}

// EscalateChainStep escalates a chain step to a different agent.
func EscalateChainStep(state *types.ChainState, plan *types.ExecutionPlan, stepID, targetAgent, domainSkill, modeSkill string) (*types.ChainState, *types.ExecutionPlan, error) {
	nextState := cloneChainState(state)
	nextPlan := cloneExecutionPlan(plan)

	step, err := requireStep(nextState, stepID)
	if err != nil {
		return nil, nil, err
	}
	compiled, err := requireCompiledStep(nextPlan, stepID)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	step.State = types.StepRunning
	step.Agent = targetAgent
	step.Attempts++
	step.StartedAt = now
	step.CompletedAt = ""
	step.Error = nil
	step.Gate = nil
	step.LastOutcome = "escalated"
	if domainSkill != "" {
		step.DomainSkill = domainSkill
	}
	if modeSkill != "" {
		step.ModeSkill = modeSkill
	}

	compiled.Agent = targetAgent
	if domainSkill != "" {
		compiled.DomainSkill = domainSkill
	}
	if modeSkill != "" {
		compiled.ModeSkill = modeSkill
	}
	nextState.State = types.ChainRunning
	nextState.CurrentStepID = stepID
	nextState.UpdatedAt = now

	return nextState, nextPlan, nil
}

// ──────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────

func handleFailureTransition(state *types.ChainState, compiled *types.CompiledStepPlan, step *types.StepState, now string, plan *types.ExecutionPlan) (AdvanceResult, error) {
	step.State = types.StepFailed
	step.LastOutcome = "failure"

	transition, ok := compiled.Transitions["failure"]
	if !ok {
		state.UpdatedAt = now
		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State:    state.State,
				NextStep: toChainStepStatus(plan, step.StepID, step.State),
				Budget:   state.Budget,
			},
			StateSnapshot: *state,
		}, nil
	}

	// Object transition with retry count
	if transition.Retry > 0 {
		if getAttemptsRemaining(step) > 0 {
			recovery := types.RecoveryAction{
				Type:        "retry",
				MaxAttempts: step.MaxRetries,
				Guidance:    fmt.Sprintf("Retry step %q before routing to %q.", step.StepID, transition.Then),
			}
			state.UpdatedAt = now
			return AdvanceResult{
				AdvanceChainResult: types.AdvanceChainResult{
					State:    state.State,
					NextStep: toChainStepStatus(plan, step.StepID, step.State),
					Recovery: &recovery,
					Budget:   state.Budget,
				},
				StateSnapshot: *state,
			}, nil
		}
		return moveToTarget(state, plan, transition.Then, now)
	}

	// String transition
	if transition.Then != "" {
		return moveToTarget(state, plan, transition.Then, now)
	}

	return moveToTarget(state, plan, transition.Then, now)
}

func applyGateDecision(state *types.ChainState, plan *types.ExecutionPlan, compiled *types.CompiledStepPlan, step *types.StepState, outcome, now string, output map[string]any) (AdvanceResult, error) {
	if step.Gate == nil {
		return AdvanceResult{}, fmt.Errorf("step %q is not waiting on a gate", step.StepID)
	}
	if outcome != "approved" && outcome != "rejected" {
		return AdvanceResult{}, fmt.Errorf("gate outcome must be approved or rejected")
	}

	step.Gate.Status = types.GateStatus(outcome)
	step.Gate.DecidedAt = now
	if outcome == "rejected" && output != nil {
		if _, ok := output["structuredFeedback"]; ok {
			step.Output = output
			step.OutputValid = true
		}
	}

	transition, ok := compiled.Transitions[outcome]
	if !ok || transition.Then == "" {
		return AdvanceResult{}, fmt.Errorf("gate outcome %q is not valid for step %q", outcome, step.StepID)
	}

	return moveToTarget(state, plan, transition.Then, now)
}

func moveToTarget(state *types.ChainState, plan *types.ExecutionPlan, target, now string) (AdvanceResult, error) {
	switch target {
	case "done":
		state.State = types.ChainCompleted
		state.CurrentStepID = ""
		state.UpdatedAt = now
		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State:  state.State,
				Budget: state.Budget,
			},
			StateSnapshot: *state,
		}, nil

	case "handoff":
		state.State = types.ChainHandoff
		state.CurrentStepID = ""
		state.UpdatedAt = now
		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State: state.State,
				Recovery: &types.RecoveryAction{
					Type:    "handoff",
					Summary: "Definition requested a handoff transition.",
				},
				Budget: state.Budget,
			},
			StateSnapshot: *state,
		}, nil

	case "abandon":
		state.State = types.ChainAbandoned
		state.CurrentStepID = ""
		state.UpdatedAt = now
		return AdvanceResult{
			AdvanceChainResult: types.AdvanceChainResult{
				State: state.State,
				Recovery: &types.RecoveryAction{
					Type:   "abort",
					Reason: "Definition requested an abandon transition.",
				},
				Budget: state.Budget,
			},
			StateSnapshot: *state,
		}, nil
	}

	nextStep, err := requireStep(state, target)
	if err != nil {
		return AdvanceResult{}, err
	}
	nextStep.State = types.StepRunning
	nextStep.StartedAt = now
	if nextStep.Attempts == 0 {
		nextStep.Attempts = 1
	} else {
		nextStep.Attempts++
	}
	nextStep.Error = nil
	nextStep.Gate = nil
	state.CurrentStepID = target
	state.State = types.ChainRunning
	state.UpdatedAt = now

	return AdvanceResult{
		AdvanceChainResult: types.AdvanceChainResult{
			State:    state.State,
			NextStep: toChainStepStatus(plan, target, nextStep.State),
			Budget:   state.Budget,
		},
		StateSnapshot: *state,
	}, nil
}

func toChainStepStatus(plan *types.ExecutionPlan, stepID string, st types.StepLifecycleState) *types.ChainStepStatus {
	if stepID == "" {
		return nil
	}
	compiled, err := requireCompiledStep(plan, stepID)
	if err != nil {
		return nil
	}
	var gate *types.GateState
	if compiled.Gate != "" {
		g := buildPendingGate(compiled.Gate, compiled.ID, nil)
		gate = g
	}
	return &types.ChainStepStatus{
		StepID:         compiled.ID,
		Agent:          compiled.Agent,
		Skills:         compiled.Skills,
		StepType:       compiled.StepType,
		State:          st,
		Model:          compiled.Model,
		Tools:          compiled.AllowedTools,
		Instructions:   compiled.Instructions,
		OutputContract: compiled.OutputContract,
		Gate:           gate,
		ComposedAgent:  compiled.ComposedAgent,
	}
}

// ──────────────────────────────────────────────────────────────
// Plan quality / red-team gate report types
// ──────────────────────────────────────────────────────────────

type planQualityVerdict string

const (
	verdictPass planQualityVerdict = "pass"
	verdictWarn planQualityVerdict = "warn"
	verdictFail planQualityVerdict = "fail"
)

type redTeamStatus string

const (
	rtStatusOK       redTeamStatus = "ok"
	rtStatusSoftFail redTeamStatus = "soft_fail"
	rtStatusSkipped  redTeamStatus = "skipped"
)

type reportLocation struct {
	File      string  `json:"file"`
	Section   *string `json:"section"`
	LineStart *int    `json:"lineStart"`
	LineEnd   *int    `json:"lineEnd"`
}

type planQualityFinding struct {
	Rule     string         `json:"rule"`
	Severity string         `json:"severity"`
	Message  string         `json:"message"`
	Location reportLocation `json:"location"`
}

type planQualityReport struct {
	SchemaVersion  string               `json:"schemaVersion"`
	Verdict        planQualityVerdict   `json:"verdict"`
	Findings       []planQualityFinding `json:"findings"`
	CheckedAgainst map[string]any       `json:"checkedAgainst"`
}

type redTeamFinding struct {
	Category       string         `json:"category"`
	Severity       string         `json:"severity"`
	Message        string         `json:"message"`
	Recommendation string         `json:"recommendation"`
	Location       reportLocation `json:"location"`
}

type redTeamPlanReport struct {
	SchemaVersion string           `json:"schemaVersion"`
	Status        redTeamStatus    `json:"status"`
	Findings      []redTeamFinding `json:"findings"`
}

type mergedGateReport struct {
	SchemaVersion string `json:"schemaVersion"`
	Summary       struct {
		PlanVerdict   planQualityVerdict `json:"planVerdict"`
		RedTeamStatus redTeamStatus      `json:"redTeamStatus"`
		BlockingCount int                `json:"blockingCount"`
		WarningCount  int                `json:"warningCount"`
	} `json:"summary"`
	PlanQuality       *planQualityReport `json:"planQuality"`
	AdversarialReview *redTeamPlanReport `json:"adversarialReview"`
}

func buildPendingGate(gateType, stepID string, state *types.ChainState) *types.GateState {
	var prompt string
	if stepID == "plan-gate" && state != nil {
		if report := buildMergedGateReport(state); report != nil {
			prompt = renderMergedGatePrompt(gateType, stepID, report)
		}
	}
	if prompt == "" {
		prompt = fmt.Sprintf("Awaiting %s for step %q.", gateType, stepID)
	}
	return &types.GateState{
		Type:   types.GateType(gateType),
		Prompt: prompt,
		Status: types.GatePending,
	}
}

func buildMergedGateReport(state *types.ChainState) *mergedGateReport {
	pq := findPlanQualityReport(state)
	if pq == nil {
		return nil
	}
	rt := findRedTeamPlanReport(state)

	var allFindings int
	for _, f := range pq.Findings {
		if isBlockingSeverity(f.Severity) {
			allFindings++
		}
	}
	if rt != nil {
		for _, f := range rt.Findings {
			if isBlockingSeverity(f.Severity) {
				allFindings++
			}
		}
	}

	rtStatus := rtStatusSkipped
	if rt != nil {
		rtStatus = rt.Status
	}

	return &mergedGateReport{
		SchemaVersion: "plan-gate-report/v1",
		Summary: struct {
			PlanVerdict   planQualityVerdict `json:"planVerdict"`
			RedTeamStatus redTeamStatus      `json:"redTeamStatus"`
			BlockingCount int                `json:"blockingCount"`
			WarningCount  int                `json:"warningCount"`
		}{
			PlanVerdict:   pq.Verdict,
			RedTeamStatus: rtStatus,
			BlockingCount: allFindings,
		},
		PlanQuality:       pq,
		AdversarialReview: rt,
	}
}

func findPlanQualityReport(state *types.ChainState) *planQualityReport {
	for i := range state.Steps {
		step := &state.Steps[i]
		for _, key := range []string{"planQualityReport", "PlanQualityReport", "report"} {
			if r := extractNested(step.Output, key); r != nil {
				var report planQualityReport
				if err := mapToStruct(r, &report); err == nil && report.SchemaVersion == "plan-quality-report/v1" &&
					isValidVerdict(report.Verdict) && report.Findings != nil {
					return &report
				}
			}
		}
		if step.Output != nil {
			var report planQualityReport
			if err := mapToStruct(step.Output, &report); err == nil && report.SchemaVersion == "plan-quality-report/v1" &&
				isValidVerdict(report.Verdict) && report.Findings != nil {
				return &report
			}
		}
	}
	return nil
}

func findRedTeamPlanReport(state *types.ChainState) *redTeamPlanReport {
	for i := range state.Steps {
		step := &state.Steps[i]
		for _, key := range []string{"redTeamPlanReport", "RedTeamPlanReport", "adversarialReview", "report"} {
			if r := extractNested(step.Output, key); r != nil {
				var report redTeamPlanReport
				if err := mapToStruct(r, &report); err == nil && report.SchemaVersion == "red-team-plan-report/v1" &&
					isValidRTStatus(report.Status) && report.Findings != nil {
					return &report
				}
			}
		}
	}
	return nil
}

func extractNested(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if mv, ok := v.(map[string]any); ok {
			return mv
		}
	}
	return nil
}

func mapToStruct(m map[string]any, target any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func isValidVerdict(v planQualityVerdict) bool {
	return v == verdictPass || v == verdictWarn || v == verdictFail
}

func isValidRTStatus(s redTeamStatus) bool {
	return s == rtStatusOK || s == rtStatusSoftFail || s == rtStatusSkipped
}

func isBlockingSeverity(s string) bool {
	return s == "fail" || s == "high" || s == "critical"
}

func renderMergedGatePrompt(gateType, stepID string, report *mergedGateReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Awaiting %s for step %q.\n\n", gateType, stepID))
	b.WriteString("## Plan Gate Report\n")
	b.WriteString(fmt.Sprintf("PlanQualityReport verdict: %s\n", report.Summary.PlanVerdict))
	b.WriteString(fmt.Sprintf("RedTeamPlanReport status: %s\n", report.Summary.RedTeamStatus))
	b.WriteString(fmt.Sprintf("Blocking findings: %d\n", report.Summary.BlockingCount))
	b.WriteString(fmt.Sprintf("Warning findings: %d\n\n", report.Summary.WarningCount))

	if len(report.PlanQuality.Findings) == 0 {
		b.WriteString("PlanQualityReport findings: none.\n\n")
	} else {
		b.WriteString("PlanQualityReport findings:\n")
		for _, f := range report.PlanQuality.Findings {
			b.WriteString(fmt.Sprintf("- %s %s: %s (%s)\n",
				strings.ToUpper(f.Severity), f.Rule, f.Message, formatLocation(f.Location)))
		}
		b.WriteString("\n")
	}

	if report.AdversarialReview == nil {
		b.WriteString("RedTeamPlanReport findings: skipped.\n\n")
	} else if len(report.AdversarialReview.Findings) == 0 {
		b.WriteString(fmt.Sprintf("RedTeamPlanReport findings: %s; none.\n\n", report.AdversarialReview.Status))
	} else {
		b.WriteString(fmt.Sprintf("RedTeamPlanReport findings (%s):\n", report.AdversarialReview.Status))
		for _, f := range report.AdversarialReview.Findings {
			b.WriteString(fmt.Sprintf("- %s %s: %s Recommendation: %s (%s)\n",
				strings.ToUpper(f.Severity), f.Category, f.Message, f.Recommendation, formatLocation(f.Location)))
		}
		b.WriteString("\n")
	}

	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	b.WriteString("\n```json\n")
	b.WriteString(string(reportJSON))
	b.WriteString("\n```\n")
	return b.String()
}

func formatLocation(loc reportLocation) string {
	var lineRange string
	if loc.LineStart != nil {
		if loc.LineEnd == nil || *loc.LineEnd == *loc.LineStart {
			lineRange = fmt.Sprintf(":%d", *loc.LineStart)
		} else {
			lineRange = fmt.Sprintf(":%d-%d", *loc.LineStart, *loc.LineEnd)
		}
	}
	section := ""
	if loc.Section != nil {
		section = fmt.Sprintf(" § %s", *loc.Section)
	}
	return fmt.Sprintf("%s%s%s", loc.File, lineRange, section)
}

// ──────────────────────────────────────────────────────────────
// Utility functions
// ──────────────────────────────────────────────────────────────

func getMaxRetries(transitions map[string]types.StepTransition) int {
	if t, ok := transitions["failure"]; ok {
		return t.Retry
	}
	return 0
}

func getAttemptsRemaining(step *types.StepState) int {
	return step.MaxRetries - max(step.Attempts-1, 0)
}

func requireStep(state *types.ChainState, stepID string) (*types.StepState, error) {
	for i := range state.Steps {
		if state.Steps[i].StepID == stepID {
			return &state.Steps[i], nil
		}
	}
	return nil, fmt.Errorf("unknown step state: %s", stepID)
}

func requireCompiledStep(plan *types.ExecutionPlan, stepID string) (*types.CompiledStepPlan, error) {
	for i := range plan.CompiledSteps {
		if plan.CompiledSteps[i].ID == stepID {
			return &plan.CompiledSteps[i], nil
		}
	}
	return nil, fmt.Errorf("unknown compiled step: %s", stepID)
}

func createEmptyBudget(plan *types.ExecutionPlan) types.BudgetState {
	scaledCostUsd := plan.BudgetPolicy.CostUsd
	if scaledCostUsd != nil {
		s := *scaledCostUsd
		s.Limit *= 100
		scaledCostUsd = &s
	}
	return types.BudgetState{
		PolicyID:      plan.BudgetPolicy.ID,
		Scope:         "chain",
		Tokens:        buildDimension(plan.BudgetPolicy.Tokens),
		CostUsd:       buildDimension(scaledCostUsd),
		WallClockMs:   buildDimension(plan.BudgetPolicy.WallClockMs),
		Retries:       buildDimension(plan.BudgetPolicy.Retries),
		ByStep:        map[string]types.StepUsage{},
		LastUpdatedAt: plan.CreatedAt,
	}
}

func buildDimension(threshold *types.BudgetThreshold) types.BudgetDimensionState {
	d := types.BudgetDimensionState{Consumed: 0, WarningTriggered: false, PausedAtLimit: false}
	if threshold != nil {
		d.Limit = threshold.Limit
		d.Remaining = threshold.Limit
	}
	return d
}

func markCompleted(state *types.ChainState, stepID string) {
	for _, id := range state.CompletedStepIDs {
		if id == stepID {
			return
		}
	}
	state.CompletedStepIDs = append(state.CompletedStepIDs, stepID)
}

func cloneChainState(s *types.ChainState) *types.ChainState {
	b, _ := json.Marshal(s)
	var c types.ChainState
	json.Unmarshal(b, &c)
	return &c
}

func cloneExecutionPlan(p *types.ExecutionPlan) *types.ExecutionPlan {
	b, _ := json.Marshal(p)
	var c types.ExecutionPlan
	json.Unmarshal(b, &c)
	return &c
}
