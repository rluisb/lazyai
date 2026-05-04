package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

const (
	defaultModel         = "sonnet"
	defaultMCPServerName = "lazyai-orchestrator"
	handoffPathURIPrefix = "sqlite://handoffs/"
)

type catalogListItem struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"activeVersion,omitempty"`
	TotalVersions int    `json:"totalVersions"`
	Description   string `json:"description"`
	Path          string `json:"path"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type retryStepInput struct {
	RunID  string        `json:"runId"`
	Kind   types.RunKind `json:"kind"`
	StepID string        `json:"stepId"`
	Reason string        `json:"reason,omitempty"`
}

type escalateStepInput struct {
	RunID         string        `json:"runId"`
	Kind          types.RunKind `json:"kind"`
	StepID        string        `json:"stepId"`
	TargetAgent   string        `json:"targetAgent"`
	TargetPhaseID string        `json:"targetPhaseId,omitempty"`
	DomainSkill   string        `json:"domainSkill,omitempty"`
	ModeSkill     string        `json:"modeSkill,omitempty"`
	Reason        string        `json:"reason,omitempty"`
}

type handoffInput struct {
	RunID            string        `json:"runId"`
	Kind             types.RunKind `json:"kind"`
	Summary          string        `json:"summary,omitempty"`
	Recipient        string        `json:"recipient,omitempty"`
	IncludeArtifacts bool          `json:"includeArtifacts,omitempty"`
}

func bindArguments(req mcp.CallToolRequest, target any) error {
	return req.BindArguments(target)
}

func runIDFromRequest(req mcp.CallToolRequest) string {
	if id := req.GetString("runId", ""); id != "" {
		return id
	}
	return req.GetString("chainId", "")
}

func retryStepInputFromRequest(req mcp.CallToolRequest) retryStepInput {
	var input retryStepInput
	_ = bindArguments(req, &input)
	if input.RunID == "" {
		input.RunID = req.GetString("chainId", "")
	}
	return input
}

func escalateStepInputFromRequest(req mcp.CallToolRequest) escalateStepInput {
	var input escalateStepInput
	_ = bindArguments(req, &input)
	if input.RunID == "" {
		input.RunID = req.GetString("chainId", "")
	}
	return input
}

func handoffInputFromRequest(req mcp.CallToolRequest) handoffInput {
	var input handoffInput
	_ = bindArguments(req, &input)
	if input.RunID == "" {
		input.RunID = req.GetString("chainId", "")
	}
	return input
}

func (o *Orchestrator) projectRoot() string {
	if o.Scope == nil {
		return ""
	}
	return o.Scope.ProjectRoot
}

func (o *Orchestrator) listCatalogItems(req mcp.CallToolRequest) ([]catalogListItem, error) {
	requestedKinds := requestedCatalogKinds(req)
	query := strings.ToLower(req.GetString("query", ""))

	summaries, err := o.Catalog.List("")
	if err != nil {
		return nil, err
	}

	items := make([]catalogListItem, 0, len(summaries))
	for _, summary := range summaries {
		if len(requestedKinds) > 0 && !requestedKinds[summary.Kind] {
			continue
		}

		item := catalogListItem{
			Kind:          summary.Kind,
			Name:          summary.Name,
			ActiveVersion: summary.ActiveVersion,
			TotalVersions: summary.TotalVersions,
			Path:          fmt.Sprintf("catalog://%s/%s", summary.Kind, summary.Name),
			CreatedAt:     summary.CreatedAt,
			UpdatedAt:     summary.UpdatedAt,
		}
		if summary.ActiveVersion != nil {
			item.Path = fmt.Sprintf("catalog://%s/%s/%d", summary.Kind, summary.Name, *summary.ActiveVersion)
			if version, err := o.Catalog.GetVersion(summary.Kind, summary.Name, *summary.ActiveVersion); err == nil {
				item.Description = descriptionFromVersion(version)
			}
		}

		if query != "" {
			haystack := strings.ToLower(fmt.Sprintf("%s %s %s", item.Kind, item.Name, item.Description))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func requestedCatalogKinds(req mcp.CallToolRequest) map[string]bool {
	if kind := req.GetString("kind", ""); kind != "" {
		return map[string]bool{kind: true}
	}

	args := req.GetArguments()
	raw, ok := args["kinds"]
	if !ok {
		return nil
	}

	kinds := map[string]bool{}
	switch values := raw.(type) {
	case []string:
		for _, kind := range values {
			kinds[kind] = true
		}
	case []any:
		for _, value := range values {
			if kind, ok := value.(string); ok {
				kinds[kind] = true
			}
		}
	}
	return kinds
}

func descriptionFromVersion(version *catalog.VersionRow) string {
	var frontmatter map[string]any
	if err := json.Unmarshal([]byte(version.FrontmatterJSON), &frontmatter); err == nil {
		if description, ok := frontmatter["description"].(string); ok {
			return description
		}
	}
	var metadata struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(version.Body), &metadata); err == nil {
		return metadata.Description
	}
	return ""
}

func (o *Orchestrator) compileChainPlan(input types.StartChainInput) (*types.ExecutionPlan, error) {
	version, err := o.getActiveOrLatestVersion(string(types.KindChain), input.Chain)
	if err != nil {
		return nil, fmt.Errorf("unknown chain definition %q: %w", input.Chain, err)
	}

	definition, err := decodeChainDefinition(version)
	if err != nil {
		return nil, err
	}
	if err := validateChainDefinition(definition); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	budgetPolicy := buildBudgetPolicy(input.Budget)
	compiledSteps := make([]types.CompiledStepPlan, 0, len(definition.Steps))
	for _, step := range definition.Steps {
		compiled, err := o.compileChainStep(definition, step, input)
		if err != nil {
			return nil, err
		}
		compiledSteps = append(compiledSteps, compiled)
	}

	plan := &types.ExecutionPlan{
		ID:   uuid.NewString(),
		Kind: string(types.RunKindChain),
		Definition: types.DefinitionRef{
			Kind:    string(types.KindChain),
			Name:    definition.Name,
			Version: definition.Version,
			Source:  definition.Source,
			Path:    definition.Path,
		},
		Cli:           cliContext(input),
		Project:       types.ProjectStackContext{RootPath: o.projectRoot()},
		BudgetPolicy:  budgetPolicy,
		Entrypoint:    definition.Entry,
		CompiledSteps: compiledSteps,
		CreatedAt:     now,
		Task:          input.Task,
	}
	if input.Context != nil {
		plan.RootContext = &input.Context.RootContext
	}
	return plan, nil
}

func decodeChainDefinition(version *catalog.VersionRow) (*types.ChainDefinition, error) {
	var definition types.ChainDefinition
	if err := json.Unmarshal([]byte(version.Body), &definition); err != nil {
		return nil, fmt.Errorf("active chain %s/%s version %d body must be a JSON chain definition: %w", version.Kind, version.Name, version.Version, err)
	}
	if definition.Name == "" {
		definition.Name = version.Name
	}
	if definition.Kind == "" {
		definition.Kind = string(types.KindChain)
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

func (o *Orchestrator) getActiveOrLatestVersion(kind, name string) (*catalog.VersionRow, error) {
	version, err := o.Catalog.GetVersion(kind, name, 0)
	if err == nil {
		return version, nil
	}
	versions, listErr := o.Catalog.ListVersions(kind, name)
	if listErr != nil || len(versions) == 0 {
		return nil, err
	}
	return &versions[0], nil
}

func validateChainDefinition(definition *types.ChainDefinition) error {
	if definition.Entry == "" {
		return fmt.Errorf("chain %q must define an entry step", definition.Name)
	}
	if len(definition.Steps) == 0 {
		return fmt.Errorf("chain %q must define at least one step", definition.Name)
	}

	stepIDs := map[string]bool{}
	for _, step := range definition.Steps {
		if step.ID == "" {
			return fmt.Errorf("chain %q contains a step without an id", definition.Name)
		}
		stepIDs[step.ID] = true
	}
	if !stepIDs[definition.Entry] {
		return fmt.Errorf("chain entry step %q does not exist", definition.Entry)
	}

	terminalTargets := map[string]bool{"done": true, "handoff": true, "abandon": true}
	for _, step := range definition.Steps {
		if step.Agent == "" {
			return fmt.Errorf("chain step %q must define an agent", step.ID)
		}
		if len(step.Transitions) == 0 {
			return fmt.Errorf("chain step %q must define at least one transition", step.ID)
		}
		for outcome, transition := range step.Transitions {
			target := transition.Then
			if target == "" {
				return fmt.Errorf("chain step %q outcome %q must define a transition target", step.ID, outcome)
			}
			if !stepIDs[target] && !terminalTargets[target] {
				return fmt.Errorf("chain step %q outcome %q references unknown target %q", step.ID, outcome, target)
			}
		}
	}
	return nil
}

func (o *Orchestrator) compileChainStep(definition *types.ChainDefinition, step types.ChainStepDefinition, input types.StartChainInput) (types.CompiledStepPlan, error) {
	agent := o.resolveBaseAgent(step.Agent)
	stepType := inferStepType(step)
	outputContract := outputContractFor(stepType)
	domainSkill := injectedSkill(definition.DomainSkillInject, step, input.DomainSkill)
	modeSkill := injectedSkill(definition.ModeSkillInject, step, input.ModeSkill)
	instructions := buildStepInstructions(step, outputContract)
	tools := step.AllowedTools
	if len(tools) == 0 {
		tools = agent.AllowedTools
	}
	model := step.Model
	if model == "" {
		model = agent.ModelHint
	}
	if model == "" {
		model = defaultModel
	}
	approvalPolicy := types.ApprovalMinimal
	if step.Gate != "" {
		approvalPolicy = types.ApprovalStrict
	}

	composed := types.ComposedAgentSpec{
		ID:             step.ID,
		Base:           agent.Name,
		DomainSkill:    domainSkill,
		ModeSkill:      modeSkill,
		Model:          model,
		Tools:          tools,
		ApprovalPolicy: approvalPolicy,
		Constraints:    agent.Constraints,
		Prompt:         strings.TrimSpace(strings.Join([]string{agent.Prompt, instructions}, "\n\n")),
		OutputContract: &outputContract,
		MergedFrom: []types.PromptLayer{
			{Source: "base", Name: agent.Name, Prompt: agent.Prompt, AllowedTools: agent.AllowedTools, ModelHint: agent.ModelHint, Constraints: agent.Constraints, ApprovalPolicy: types.ApprovalMinimal},
			{Source: "step", Name: step.ID, Prompt: instructions, AllowedTools: step.AllowedTools, ModelHint: step.Model, ApprovalPolicy: approvalPolicy},
		},
	}

	return types.CompiledStepPlan{
		ID:             step.ID,
		Kind:           "step",
		Agent:          step.Agent,
		Skills:         step.Skills,
		TaskType:       step.TaskType,
		StepType:       stepType,
		DomainSkill:    domainSkill,
		ModeSkill:      modeSkill,
		Instructions:   instructions,
		AllowedTools:   tools,
		Model:          model,
		OutputContract: outputContract,
		Transitions:    step.Transitions,
		Gate:           step.Gate,
		ComposedAgent:  composed,
		ExecutionMode:  types.ExecutionNative,
	}, nil
}

func (o *Orchestrator) resolveBaseAgent(name string) types.BaseAgentDefinition {
	agent := types.BaseAgentDefinition{
		DefinitionMetadata: types.DefinitionMetadata{Name: name, Source: types.SourceDB, Path: fmt.Sprintf("catalog://agent/%s", name)},
		Kind:               string(types.KindAgent),
		Prompt:             fmt.Sprintf("Execute as the %s agent.", name),
		AllowedTools:       []string{},
		Constraints:        []string{},
		DisplayName:        name,
	}

	version, err := o.Catalog.GetVersion(string(types.KindAgent), name, 0)
	if err != nil {
		return agent
	}
	var parsed types.BaseAgentDefinition
	if err := json.Unmarshal([]byte(version.Body), &parsed); err == nil && parsed.Prompt != "" {
		if parsed.Name == "" {
			parsed.Name = name
		}
		if parsed.Source == "" {
			parsed.Source = types.SourceDB
		}
		if parsed.Path == "" {
			parsed.Path = fmt.Sprintf("catalog://agent/%s/%d", name, version.Version)
		}
		if parsed.DisplayName == "" {
			parsed.DisplayName = parsed.Name
		}
		return parsed
	}

	agent.Prompt = version.Body
	agent.Path = fmt.Sprintf("catalog://agent/%s/%d", name, version.Version)
	return agent
}

func buildBudgetPolicy(overrides *types.BudgetPolicy) types.BudgetPolicy {
	policy := types.BudgetPolicy{
		ID:                   uuid.NewString(),
		Scope:                string(types.RunKindChain),
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
	policy.RequireUserApprovalMultiplier = overrides.RequireUserApprovalMultiplier
	return policy
}

func cliContext(input types.StartChainInput) types.CliContext {
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

func inferStepType(step types.ChainStepDefinition) types.StepType {
	identifiers := append([]string{step.ID, step.TaskType}, step.Skills...)
	for _, identifier := range identifiers {
		value := strings.ToLower(identifier)
		switch {
		case strings.Contains(value, "research"):
			return types.StepResearch
		case strings.Contains(value, "plan"):
			return types.StepPlan
		case strings.Contains(value, "review"):
			return types.StepReview
		case strings.Contains(value, "document"):
			return types.StepDocument
		case strings.Contains(value, "implement") || strings.Contains(value, "fix") || strings.Contains(value, "iterate"):
			return types.StepImplement
		}
	}
	return types.StepCustom
}

func outputContractFor(stepType types.StepType) types.StepOutputContract {
	switch stepType {
	case types.StepResearch:
		return contractFor(stepType, []string{"summary", "status", "findings"})
	case types.StepPlan:
		return contractFor(stepType, []string{"summary", "status", "plan", "tasks"})
	case types.StepImplement:
		return contractFor(stepType, []string{"summary", "status", "files_changed", "tests_passed"})
	case types.StepReview:
		return contractFor(stepType, []string{"summary", "status", "verdict", "findings"})
	case types.StepDocument:
		return contractFor(stepType, []string{"summary", "status", "files_created"})
	default:
		return contractFor(types.StepCustom, []string{})
	}
}

func contractFor(stepType types.StepType, required []string) types.StepOutputContract {
	contract := types.StepOutputContract{
		StepType:                  stepType,
		RequiredFields:            required,
		AllowAdditionalProperties: true,
		Schema: map[string]any{
			"type":     "object",
			"required": required,
		},
	}
	contract.OnValidationFailure.Category = types.ErrorValidation
	contract.OnValidationFailure.DefaultRecovery = types.RecoveryAction{
		Type:     string(types.RecoveryRetry),
		Guidance: fmt.Sprintf("Return the required structured fields: %s.", strings.Join(required, ", ")),
	}
	if len(required) == 0 {
		contract.OnValidationFailure.DefaultRecovery.Guidance = "Return a structured JSON object for the custom step."
	}
	return contract
}

func injectedSkill(mode string, step types.ChainStepDefinition, selected string) string {
	if selected == "" || mode == "none" {
		return ""
	}
	if mode == "builder_steps_only" && step.Agent != "builder" {
		return ""
	}
	return selected
}

func buildStepInstructions(step types.ChainStepDefinition, outputContract types.StepOutputContract) string {
	sections := []string{
		fmt.Sprintf("Step: %s", step.ID),
		step.Description,
		step.Prompt,
	}
	if step.TaskType != "" {
		sections = append(sections, fmt.Sprintf("Task Type: %s", step.TaskType))
	}
	if len(step.Skills) > 0 {
		sections = append(sections, fmt.Sprintf("Apply supporting skills: %s.", strings.Join(step.Skills, ", ")))
	}
	sections = append(sections,
		fmt.Sprintf("Return structured output with required fields: %s.", strings.Join(outputContract.RequiredFields, ", ")),
		fmt.Sprintf("Valid outcomes: %s.", strings.Join(sortedTransitionKeys(step.Transitions), ", ")),
	)
	if step.Gate != "" {
		sections = append(sections, fmt.Sprintf("A %s gate must be satisfied before the chain can continue.", step.Gate))
	}

	filtered := make([]string, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section) != "" {
			filtered = append(filtered, section)
		}
	}
	return strings.Join(filtered, "\n\n")
}

func sortedTransitionKeys(transitions map[string]types.StepTransition) []string {
	keys := make([]string, 0, len(transitions))
	for key := range transitions {
		keys = append(keys, key)
	}
	// Transition maps are tiny; insertion sort avoids another import in this helper-heavy file.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func saveExecutionPlan(database *db.DB, plan *types.ExecutionPlan) error {
	encoded, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshal execution plan: %w", err)
	}
	_, err = database.Exec(`
		INSERT INTO execution_plans (id, kind, definition_name, definition_version, project_root, plan_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			definition_name = excluded.definition_name,
			definition_version = excluded.definition_version,
			project_root = excluded.project_root,
			plan_json = excluded.plan_json
	`, plan.ID, plan.Kind, plan.Definition.Name, plan.Definition.Version, plan.Project.RootPath, string(encoded), plan.CreatedAt)
	if err != nil {
		return fmt.Errorf("save execution plan %s: %w", plan.ID, err)
	}
	return nil
}

func loadExecutionPlan(database *db.DB, id string) (*types.ExecutionPlan, error) {
	var planJSON string
	err := database.QueryRow(`SELECT plan_json FROM execution_plans WHERE id = ?`, id).Scan(&planJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution plan not found: %s", id)
		}
		return nil, err
	}
	var plan types.ExecutionPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, fmt.Errorf("decode execution plan %s: %w", id, err)
	}
	return &plan, nil
}

func saveChainState(database *db.DB, projectRoot string, chainState *types.ChainState) error {
	encoded, err := json.Marshal(chainState)
	if err != nil {
		return fmt.Errorf("marshal chain state: %w", err)
	}
	_, err = database.Exec(`
		INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state = excluded.state,
			current_step_id = excluded.current_step_id,
			state_json = excluded.state_json,
			updated_at = excluded.updated_at
	`, chainState.ChainID, chainState.DefinitionName, chainState.DefinitionVersion, chainState.State, nullableString(chainState.CurrentStepID), projectRoot, string(encoded), chainState.CreatedAt, chainState.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save chain state %s: %w", chainState.ChainID, err)
	}
	return nil
}

func loadChainState(database *db.DB, id string) (*types.ChainState, error) {
	var stateJSON string
	err := database.QueryRow(`SELECT state_json FROM chain_runs WHERE id = ?`, id).Scan(&stateJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain state not found: %s", id)
		}
		return nil, err
	}
	var chainState types.ChainState
	if err := json.Unmarshal([]byte(stateJSON), &chainState); err != nil {
		return nil, fmt.Errorf("decode chain state %s: %w", id, err)
	}
	return &chainState, nil
}

func createAndSaveChainHandoff(database *db.DB, chainState *types.ChainState, plan *types.ExecutionPlan, input handoffInput) (*types.HandoffDocument, string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	chainState.State = types.ChainHandoff
	chainState.UpdatedAt = now

	summary := input.Summary
	if summary == "" {
		summary = fmt.Sprintf("Handoff for chain %s", chainState.ChainID)
	}
	doc := &types.HandoffDocument{
		ID:        uuid.NewString(),
		RunID:     chainState.ChainID,
		Kind:      types.RunKindChain,
		Summary:   summary,
		Recipient: input.Recipient,
		CreatedAt: now,
		Resumable: true,
		Status:    chainState,
		Plan:      plan,
	}
	encoded, err := json.Marshal(doc)
	if err != nil {
		return nil, "", fmt.Errorf("marshal handoff: %w", err)
	}
	_, err = database.Exec(`INSERT OR IGNORE INTO handoffs (id, run_id, run_kind, doc_json, created_at) VALUES (?, ?, ?, ?, ?)`, doc.ID, doc.RunID, doc.Kind, string(encoded), doc.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("save handoff %s: %w", doc.ID, err)
	}
	path := handoffPathURIPrefix + doc.ID
	chainState.HandoffPath = path
	return doc, path, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func currentChainStepStatus(plan *types.ExecutionPlan, chainState *types.ChainState) *types.ChainStepStatus {
	if chainState.CurrentStepID == "" {
		return nil
	}
	var runtimeStep *types.StepState
	for i := range chainState.Steps {
		if chainState.Steps[i].StepID == chainState.CurrentStepID {
			runtimeStep = &chainState.Steps[i]
			break
		}
	}
	if runtimeStep == nil {
		return nil
	}
	var compiled *types.CompiledStepPlan
	for i := range plan.CompiledSteps {
		if plan.CompiledSteps[i].ID == chainState.CurrentStepID {
			compiled = &plan.CompiledSteps[i]
			break
		}
	}
	if compiled == nil {
		return nil
	}
	return &types.ChainStepStatus{
		StepID:         compiled.ID,
		Agent:          runtimeStep.Agent,
		Skills:         runtimeStep.Skills,
		StepType:       runtimeStep.StepType,
		State:          runtimeStep.State,
		Model:          compiled.Model,
		Tools:          compiled.AllowedTools,
		Instructions:   compiled.Instructions,
		OutputContract: compiled.OutputContract,
		Gate:           runtimeStep.Gate,
		ComposedAgent:  compiled.ComposedAgent,
	}
}
