package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// ──────────────────────────────────────────────────────────────
// Definition metadata (shared across all catalog definitions)
// ──────────────────────────────────────────────────────────────

type DefinitionMetadata struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version,omitempty"`
	Source      DefinitionSource `json:"source"`
	Path        string           `json:"path"`
}

// ──────────────────────────────────────────────────────────────
// Agent definitions
// ──────────────────────────────────────────────────────────────

type BaseAgentDefinition struct {
	DefinitionMetadata
	Kind         string   `json:"kind"`
	Prompt       string   `json:"prompt"`
	ModelHint    string   `json:"modelHint,omitempty"`
	AllowedTools []string `json:"allowedTools"`
	Constraints  []string `json:"constraints"`
	DisplayName  string   `json:"displayName"`
}

type SkillDefinition struct {
	DefinitionMetadata
	Kind           string         `json:"kind"`
	Prompt         string         `json:"prompt"`
	AllowedTools   []string       `json:"allowedTools,omitempty"`
	Constraints    []string       `json:"constraints"`
	ModelHint      string         `json:"modelHint,omitempty"`
	ApprovalPolicy ApprovalPolicy `json:"approvalPolicy,omitempty"`
	AppliesTo      []string       `json:"appliesTo,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Agent execution configuration (new for A2A)
// ──────────────────────────────────────────────────────────────

type AgentFallbackConfig struct {
	Provider       ProviderId  `json:"provider"`
	Model          string      `json:"model"`
	ThinkingBudget int         `json:"thinkingBudget,omitempty"`
	Effort         EffortLevel `json:"effort,omitempty"`
	Temperature    float64     `json:"temperature,omitempty"`
	OnError        []string    `json:"onError,omitempty"`
	MaxRetries     int         `json:"maxRetries,omitempty"`
}

type AgentExecutionConfig struct {
	Provider       ProviderId            `json:"provider"`
	Model          string                `json:"model"`
	ThinkingBudget int                   `json:"thinkingBudget,omitempty"`
	Effort         EffortLevel           `json:"effort,omitempty"`
	Temperature    float64               `json:"temperature,omitempty"`
	TopP           float64               `json:"topP,omitempty"`
	MaxTokens      int                   `json:"maxTokens,omitempty"`
	Fallback       []AgentFallbackConfig `json:"fallback,omitempty"`
}

type A2AAgentConfig struct {
	Enabled  bool                 `json:"enabled"`
	Mode     string               `json:"mode"` // "embedded", "http", "grpc"
	Endpoint string               `json:"endpoint,omitempty"`
	Config   AgentExecutionConfig `json:"config"`
}

type NativeAgentToolConfig struct {
	Model          string            `json:"model"`
	Provider       ProviderId        `json:"provider,omitempty"`
	Effort         EffortLevel       `json:"effort,omitempty"`
	Tools          []string          `json:"tools,omitempty"`
	PermissionMode string            `json:"permissionMode,omitempty"`
	Permission     map[string]string `json:"permission,omitempty"`
	Mode           string            `json:"mode,omitempty"`
	Hidden         bool              `json:"hidden,omitempty"`
	MaxTurns       int               `json:"maxTurns,omitempty"`
	Background     bool              `json:"background,omitempty"`
	Isolation      string            `json:"isolation,omitempty"`
	Skills         []string          `json:"skills,omitempty"`
	Color          string            `json:"color,omitempty"`
	UserInvocable  bool              `json:"userInvocable,omitempty"`
}

type AgentExecution struct {
	A2A      *A2AAgentConfig                  `json:"a2a,omitempty"`
	Native   map[string]NativeAgentToolConfig `json:"native,omitempty"`
	Embedded *AgentExecutionConfig            `json:"embedded,omitempty"`
}

// AgentDefinition is the full definition for an agent in the catalog.
type AgentDefinition struct {
	DefinitionMetadata
	ID             string              `json:"id"`
	Kind           string              `json:"kind"`
	SystemPrompt   string              `json:"systemPrompt"`
	Skills         []string            `json:"skills"`
	AllowedTools   []string            `json:"allowedTools"`
	Constraints    []string            `json:"constraints"`
	ApprovalPolicy ApprovalPolicy      `json:"approvalPolicy"`
	Execution      AgentExecution      `json:"execution"`
	OutputContract *StepOutputContract `json:"outputContract,omitempty"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
	CreatedBy      string              `json:"createdBy,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Chain definitions
// ──────────────────────────────────────────────────────────────

type StepTransition struct {
	Retry int    `json:"retry,omitempty"`
	Then  string `json:"then,omitempty"`
}

// UnmarshalJSON accepts both TypeScript catalog transition shapes:
// a string target ("next-step") and an object ({"retry": 2, "then": "fallback"}).
func (t *StepTransition) UnmarshalJSON(data []byte) error {
	var target string
	if err := json.Unmarshal(data, &target); err == nil {
		t.Then = target
		t.Retry = 0
		return nil
	}

	type transitionAlias StepTransition
	var parsed transitionAlias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("step transition must be a string target or retry object: %w", err)
	}
	*t = StepTransition(parsed)
	return nil
}

type ChainStepDefinition struct {
	ID           string                    `json:"id"`
	Agent        string                    `json:"agent"`
	Skills       []string                  `json:"skills"`
	Description  string                    `json:"description"`
	TaskType     string                    `json:"taskType,omitempty"`
	Gate         string                    `json:"gate,omitempty"`
	Prompt       string                    `json:"prompt,omitempty"`
	Transitions  map[string]StepTransition `json:"transitions"`
	AllowedTools []string                  `json:"allowedTools,omitempty"`
	Model        string                    `json:"model,omitempty"`
}

type ChainDefinition struct {
	DefinitionMetadata
	Kind              string                `json:"kind"`
	Entry             string                `json:"entry"`
	Steps             []ChainStepDefinition `json:"steps"`
	DomainSkillInject string                `json:"domain_skill_injection,omitempty"`
	ModeSkillInject   string                `json:"mode_skill_injection,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Team definitions
// ──────────────────────────────────────────────────────────────

type TeamMemberDefinition struct {
	Role   string   `json:"role"`
	Agent  string   `json:"agent"`
	Skills []string `json:"skills"`
	Focus  string   `json:"focus"`
}

type TeamSynthesizeDefinition struct {
	Agent       string `json:"agent"`
	Description string `json:"description"`
}

type TeamDefinition struct {
	DefinitionMetadata
	Kind                     string                   `json:"kind"`
	BudgetMultiplier         *int                     `json:"budget_multiplier,omitempty"`
	UserConfirmationRequired *bool                    `json:"user_confirmation_required,omitempty"`
	Parallel                 []TeamMemberDefinition   `json:"parallel"`
	Synthesize               TeamSynthesizeDefinition `json:"synthesize"`
}

// ──────────────────────────────────────────────────────────────
// Workflow definitions
// ──────────────────────────────────────────────────────────────

type WorkflowPhaseDefinition struct {
	ID     string            `json:"id"`
	Kind   string            `json:"kind"` // "chain", "team", "gate", "terminal"
	Ref    string            `json:"ref,omitempty"`
	Gate   string            `json:"gate,omitempty"`
	Prompt string            `json:"prompt,omitempty"`
	When   string            `json:"when,omitempty"`
	On     map[string]string `json:"on,omitempty"`
}

type WorkflowDefinition struct {
	DefinitionMetadata
	Kind   string                    `json:"kind"`
	Entry  string                    `json:"entry"`
	Phases []WorkflowPhaseDefinition `json:"phases"`
}

// ──────────────────────────────────────────────────────────────
// Orchestration catalog
// ──────────────────────────────────────────────────────────────

type OrchestrationCatalog struct {
	Agents    map[string]BaseAgentDefinition `json:"agents"`
	Domains   map[string]SkillDefinition     `json:"domains"`
	Modes     map[string]SkillDefinition     `json:"modes"`
	Chains    map[string]ChainDefinition     `json:"chains"`
	Teams     map[string]TeamDefinition      `json:"teams"`
	Workflows map[string]WorkflowDefinition  `json:"workflows"`
}

// ──────────────────────────────────────────────────────────────
// Prompt composition
// ──────────────────────────────────────────────────────────────

type PromptLayer struct {
	Source         string         `json:"source"`
	Name           string         `json:"name"`
	Prompt         string         `json:"prompt"`
	AllowedTools   []string       `json:"allowedTools,omitempty"`
	ModelHint      string         `json:"modelHint,omitempty"`
	Constraints    []string       `json:"constraints,omitempty"`
	ApprovalPolicy ApprovalPolicy `json:"approvalPolicy,omitempty"`
}

type RootContextLayer struct {
	Prompt         string         `json:"prompt,omitempty"`
	Constraints    []string       `json:"constraints,omitempty"`
	AllowedTools   []string       `json:"allowedTools,omitempty"`
	ModelHint      string         `json:"modelHint,omitempty"`
	ApprovalPolicy ApprovalPolicy `json:"approvalPolicy,omitempty"`
}

type StepOutputContract struct {
	StepType                  StepType       `json:"stepType"`
	RequiredFields            []string       `json:"requiredFields"`
	AllowAdditionalProperties bool           `json:"allowAdditionalProperties"`
	Schema                    map[string]any `json:"schema"`
	OnValidationFailure       struct {
		Category        ErrorCategory  `json:"category"`
		DefaultRecovery RecoveryAction `json:"defaultRecovery"`
	} `json:"onValidationFailure"`
}

type ComposedAgentSpec struct {
	ID             string              `json:"id"`
	Base           string              `json:"base"`
	DomainSkill    string              `json:"domainSkill,omitempty"`
	ModeSkill      string              `json:"modeSkill,omitempty"`
	Model          string              `json:"model"`
	Tools          []string            `json:"tools"`
	ApprovalPolicy ApprovalPolicy      `json:"approvalPolicy"`
	Constraints    []string            `json:"constraints"`
	Prompt         string              `json:"prompt"`
	OutputContract *StepOutputContract `json:"outputContract,omitempty"`
	MergedFrom     []PromptLayer       `json:"mergedFrom"`
}

// ──────────────────────────────────────────────────────────────
// Context types
// ──────────────────────────────────────────────────────────────

type CliContext struct {
	Host                     HostCli      `json:"host"`
	DispatchMode             DispatchMode `json:"dispatchMode"`
	SupportsSubagents        bool         `json:"supportsSubagents"`
	SupportsParallelTeams    bool         `json:"supportsParallelTeams"`
	SupportsStructuredOutput bool         `json:"supportsStructuredOutput"`
	MCPServerName            string       `json:"mcpServerName"`
}

type ProjectStackContext struct {
	RootPath            string `json:"rootPath"`
	RootInstructionFile string `json:"rootInstructionFile,omitempty"`
	Language            string `json:"language,omitempty"`
	Framework           string `json:"framework,omitempty"`
	Database            string `json:"database,omitempty"`
	PackageManager      string `json:"packageManager,omitempty"`
	TestCommand         string `json:"testCommand,omitempty"`
	BuildCommand        string `json:"buildCommand,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Budget types
// ──────────────────────────────────────────────────────────────

type BudgetThreshold struct {
	Limit    int  `json:"limit,omitempty"`
	WarnAt   int  `json:"warnAt,omitempty"`
	PauseAt  int  `json:"pauseAt,omitempty"`
	HardStop bool `json:"hardStop,omitempty"`
}

type BudgetPolicy struct {
	ID                            string           `json:"id"`
	Scope                         string           `json:"scope"`
	Tokens                        *BudgetThreshold `json:"tokens,omitempty"`
	CostUsd                       *BudgetThreshold `json:"costUsd,omitempty"`
	WallClockMs                   *BudgetThreshold `json:"wallClockMs,omitempty"`
	Retries                       *BudgetThreshold `json:"retries,omitempty"`
	RequireUserApprovalMultiplier *int             `json:"requireUserApprovalForTeamMultiplier,omitempty"`
	DefaultActionOnLimit          string           `json:"defaultActionOnLimit"`
}

type BudgetDimensionState struct {
	Limit            int  `json:"limit,omitempty"`
	Consumed         int  `json:"consumed"`
	Remaining        int  `json:"remaining,omitempty"`
	WarningTriggered bool `json:"warningTriggered"`
	PausedAtLimit    bool `json:"pausedAtLimit"`
}

type StepUsage struct {
	InputTokens  int     `json:"inputTokens,omitempty"`
	OutputTokens int     `json:"outputTokens,omitempty"`
	TotalTokens  int     `json:"totalTokens,omitempty"`
	CostUsd      float64 `json:"costUsd,omitempty"`
	WallClockMs  int     `json:"wallClockMs,omitempty"`
}

type BudgetState struct {
	PolicyID      string               `json:"policyId"`
	Scope         string               `json:"scope"`
	Tokens        BudgetDimensionState `json:"tokens"`
	CostUsd       BudgetDimensionState `json:"costUsd"`
	WallClockMs   BudgetDimensionState `json:"wallClockMs"`
	Retries       BudgetDimensionState `json:"retries"`
	ByStep        map[string]StepUsage `json:"byStep"`
	LastUpdatedAt string               `json:"lastUpdatedAt"`
}

type BudgetEvaluation struct {
	Overall    BudgetHealth `json:"overall"`
	Dimensions struct {
		Tokens      BudgetHealth `json:"tokens"`
		CostUsd     BudgetHealth `json:"costUsd"`
		WallClockMs BudgetHealth `json:"wallClockMs"`
		Retries     BudgetHealth `json:"retries"`
	} `json:"dimensions"`
	RecommendedAction string `json:"recommendedAction"`
	ShouldPause       bool   `json:"shouldPause"`
}

// ──────────────────────────────────────────────────────────────
// Recovery types
// ──────────────────────────────────────────────────────────────

type RecoveryAction struct {
	Type         string `json:"type"`
	MaxAttempts  int    `json:"maxAttempts,omitempty"`
	Guidance     string `json:"guidance,omitempty"`
	Instructions string `json:"instructions,omitempty"`
	TargetAgent  string `json:"targetAgent,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Recipient    string `json:"recipient,omitempty"`
	Summary      string `json:"summary,omitempty"`
}

type StructuredError struct {
	Category ErrorCategory `json:"category"`
	Code     string        `json:"code"`
	Message  string        `json:"message"`
	StepID   string        `json:"stepId"`
	Agent    string        `json:"agent"`
	Skills   []string      `json:"skills"`
	Context  struct {
		RunID          string         `json:"runId"`
		RunKind        RunKind        `json:"runKind"`
		Task           string         `json:"task"`
		Attempt        int            `json:"attempt"`
		HostCli        HostCli        `json:"hostCli"`
		BudgetSnapshot *BudgetState   `json:"budgetSnapshot,omitempty"`
		RawOutput      map[string]any `json:"rawOutput,omitempty"`
		Notes          []string       `json:"notes,omitempty"`
	} `json:"context"`
	SuggestedRecovery RecoveryAction `json:"suggestedRecovery"`
	Timestamp         string         `json:"timestamp"`
}

type ErrorJournalEntry struct {
	ID             string          `json:"id"`
	RunID          string          `json:"runId"`
	RunKind        RunKind         `json:"runKind"`
	DefinitionName string          `json:"definitionName"`
	StepID         string          `json:"stepId,omitempty"`
	Error          StructuredError `json:"error"`
	Resolution     *struct {
		Action     RecoveryAction `json:"action"`
		ResolvedAt string         `json:"resolvedAt,omitempty"`
		Notes      string         `json:"notes,omitempty"`
	} `json:"resolution,omitempty"`
	Lesson *struct {
		Summary    string   `json:"summary"`
		Prevention []string `json:"prevention"`
		Tags       []string `json:"tags"`
	} `json:"lesson,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Execution plan types
// ──────────────────────────────────────────────────────────────

type DefinitionRef struct {
	Kind    string           `json:"kind"`
	Name    string           `json:"name"`
	Version string           `json:"version"`
	Source  DefinitionSource `json:"source"`
	Path    string           `json:"path"`
}

type CompiledStepPlan struct {
	ID             string                    `json:"id"`
	Kind           string                    `json:"kind"`
	Agent          string                    `json:"agent"`
	Skills         []string                  `json:"skills"`
	TaskType       string                    `json:"taskType,omitempty"`
	StepType       StepType                  `json:"stepType"`
	DomainSkill    string                    `json:"domainSkill,omitempty"`
	ModeSkill      string                    `json:"modeSkill,omitempty"`
	Instructions   string                    `json:"instructions"`
	AllowedTools   []string                  `json:"allowedTools"`
	Model          string                    `json:"model"`
	OutputContract StepOutputContract        `json:"outputContract"`
	Transitions    map[string]StepTransition `json:"transitions"`
	Gate           string                    `json:"gate,omitempty"`
	ComposedAgent  ComposedAgentSpec         `json:"composedAgent"`
	ExecutionMode  ExecutionMode             `json:"executionMode,omitempty"`
}

type ExecutionPlan struct {
	ID            string              `json:"id"`
	Kind          string              `json:"kind"`
	Definition    DefinitionRef       `json:"definition"`
	Cli           CliContext          `json:"cli"`
	Project       ProjectStackContext `json:"project"`
	BudgetPolicy  BudgetPolicy        `json:"budgetPolicy"`
	Entrypoint    string              `json:"entrypoint"`
	CompiledSteps []CompiledStepPlan  `json:"compiledSteps,omitempty"`
	CreatedAt     string              `json:"createdAt"`
	Task          string              `json:"task"`
	RootContext   *RootContextLayer   `json:"rootContext,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Runtime state types
// ──────────────────────────────────────────────────────────────

type GateState struct {
	Type      GateType   `json:"type"`
	Prompt    string     `json:"prompt"`
	Status    GateStatus `json:"status"`
	DecidedAt string     `json:"decidedAt,omitempty"`
}

type StepState struct {
	StepID      string             `json:"stepId"`
	Order       int                `json:"order"`
	Agent       string             `json:"agent"`
	Skills      []string           `json:"skills"`
	StepType    StepType           `json:"stepType"`
	DomainSkill string             `json:"domainSkill,omitempty"`
	ModeSkill   string             `json:"modeSkill,omitempty"`
	State       StepLifecycleState `json:"state"`
	Attempts    int                `json:"attempts"`
	MaxRetries  int                `json:"maxRetries"`
	StartedAt   string             `json:"startedAt,omitempty"`
	CompletedAt string             `json:"completedAt,omitempty"`
	Output      map[string]any     `json:"output,omitempty"`
	OutputValid bool               `json:"outputValid,omitempty"`
	Usage       StepUsage          `json:"usage"`
	Gate        *GateState         `json:"gate,omitempty"`
	LastOutcome string             `json:"lastOutcome,omitempty"`
	Error       *StructuredError   `json:"error,omitempty"`
	NextStepID  string             `json:"nextStepId,omitempty"`
}

type ChainState struct {
	ChainID           string              `json:"chainId"`
	DefinitionName    string              `json:"definitionName"`
	DefinitionVersion string              `json:"definitionVersion"`
	ExecutionPlanID   string              `json:"executionPlanId"`
	State             ChainLifecycleState `json:"state"`
	Task              string              `json:"task"`
	CurrentStepID     string              `json:"currentStepId,omitempty"`
	EntryStepID       string              `json:"entryStepId"`
	Steps             []StepState         `json:"steps"`
	CompletedStepIDs  []string            `json:"completedStepIds"`
	Budget            BudgetState         `json:"budget"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
	HandoffPath       string              `json:"handoffPath,omitempty"`
}

type TeamTaskState struct {
	TaskID      string                 `json:"taskId"`
	Kind        string                 `json:"kind"`
	Role        string                 `json:"role"`
	Agent       string                 `json:"agent"`
	Skills      []string               `json:"skills"`
	Focus       string                 `json:"focus"`
	State       TeamTaskLifecycleState `json:"state"`
	Order       int                    `json:"order"`
	DependsOn   []string               `json:"dependsOn"`
	Assignee    string                 `json:"assignee,omitempty"`
	ClaimedBy   string                 `json:"claimedBy,omitempty"`
	AssignedAt  string                 `json:"assignedAt,omitempty"`
	ClaimedAt   string                 `json:"claimedAt,omitempty"`
	CompletedAt string                 `json:"completedAt,omitempty"`
	Result      map[string]any         `json:"result,omitempty"`
	Usage       StepUsage              `json:"usage"`
	Error       *StructuredError       `json:"error,omitempty"`
}

type TeamState struct {
	TeamID            string             `json:"teamId"`
	DefinitionName    string             `json:"definitionName"`
	DefinitionVersion string             `json:"definitionVersion"`
	ExecutionPlanID   string             `json:"executionPlanId,omitempty"`
	State             TeamLifecycleState `json:"state"`
	Task              string             `json:"task"`
	Tasks             []TeamTaskState    `json:"tasks"`
	ReadyTaskIDs      []string           `json:"readyTaskIds"`
	SynthesisTaskID   string             `json:"synthesisTaskId"`
	BudgetPolicy      BudgetPolicy       `json:"budgetPolicy"`
	Budget            BudgetState        `json:"budget"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	Summary           map[string]any     `json:"summary,omitempty"`
}

type WorkflowChildRun struct {
	PhaseID        string `json:"phaseId"`
	RunID          string `json:"runId"`
	RunKind        string `json:"runKind"`
	DefinitionName string `json:"definitionName"`
	LaunchedAt     string `json:"launchedAt"`
	CompletedAt    string `json:"completedAt,omitempty"`
	Outcome        string `json:"outcome,omitempty"`
}

type WorkflowPhaseState struct {
	PhaseID     string                      `json:"phaseId"`
	Kind        string                      `json:"kind"`
	State       WorkflowPhaseLifecycleState `json:"state"`
	Ref         string                      `json:"ref,omitempty"`
	Gate        string                      `json:"gate,omitempty"`
	Prompt      string                      `json:"prompt,omitempty"`
	StartedAt   string                      `json:"startedAt,omitempty"`
	CompletedAt string                      `json:"completedAt,omitempty"`
	LastOutcome string                      `json:"lastOutcome,omitempty"`
	ChildRun    *WorkflowChildRun           `json:"childRun,omitempty"`
}

type WorkflowRecoveryDecision struct {
	Type          string `json:"type"`
	TargetPhaseID string `json:"targetPhaseId,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Recipient     string `json:"recipient,omitempty"`
	Summary       string `json:"summary,omitempty"`
}

type WorkflowRuntime struct {
	DomainSkill string `json:"domainSkill,omitempty"`
	ModeSkill   string `json:"modeSkill,omitempty"`
}

type WorkflowState struct {
	WorkflowID        string                 `json:"workflowId"`
	DefinitionName    string                 `json:"definitionName"`
	DefinitionVersion string                 `json:"definitionVersion"`
	ExecutionPlanID   string                 `json:"executionPlanId,omitempty"`
	State             WorkflowLifecycleState `json:"state"`
	Task              string                 `json:"task"`
	EntryPhaseID      string                 `json:"entryPhaseId"`
	CurrentPhaseID    string                 `json:"currentPhaseId,omitempty"`
	Phases            []WorkflowPhaseState   `json:"phases"`
	ChildRuns         []WorkflowChildRun     `json:"childRuns"`
	BudgetPolicy      BudgetPolicy           `json:"budgetPolicy"`
	Budget            BudgetState            `json:"budget"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
	LastError         *StructuredError       `json:"lastError,omitempty"`
	HandoffSummary    string                 `json:"handoffSummary,omitempty"`
	Runtime           *WorkflowRuntime       `json:"runtime,omitempty"`
}

// HandoffDocument represents a persisted resumable state artifact.
type HandoffDocument struct {
	ID        string         `json:"id"`
	RunID     string         `json:"runId"`
	Kind      RunKind        `json:"kind"`
	Summary   string         `json:"summary"`
	Recipient string         `json:"recipient,omitempty"`
	CreatedAt string         `json:"createdAt"`
	Resumable bool           `json:"resumable"`
	Status    any            `json:"status"`
	Plan      *ExecutionPlan `json:"plan,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// API input/output types
// ──────────────────────────────────────────────────────────────

type StartChainInput struct {
	Chain       string        `json:"chain"`
	Task        string        `json:"task"`
	DomainSkill string        `json:"domainSkill,omitempty"`
	ModeSkill   string        `json:"modeSkill,omitempty"`
	Budget      *BudgetPolicy `json:"budget,omitempty"`
	Context     *struct {
		CliTool     HostCli          `json:"cliTool,omitempty"`
		RootContext RootContextLayer `json:"rootContext,omitempty"`
	} `json:"context,omitempty"`
}

type AdvanceChainInput struct {
	ChainID string         `json:"chainId"`
	StepID  string         `json:"stepId"`
	Outcome string         `json:"outcome"`
	Output  map[string]any `json:"output,omitempty"`
	Usage   *StepUsage     `json:"usage,omitempty"`
}

type AdvanceChainResult struct {
	State    ChainLifecycleState `json:"state"`
	NextStep *ChainStepStatus    `json:"nextStep,omitempty"`
	Gate     *GateState          `json:"gate,omitempty"`
	Recovery *RecoveryAction     `json:"recovery,omitempty"`
	Budget   BudgetState         `json:"budget"`
	Error    *StructuredError    `json:"error,omitempty"`
}

type ChainStepStatus struct {
	StepID         string             `json:"stepId"`
	Agent          string             `json:"agent"`
	Skills         []string           `json:"skills"`
	StepType       StepType           `json:"stepType"`
	State          StepLifecycleState `json:"state"`
	Model          string             `json:"model"`
	Tools          []string           `json:"tools"`
	Instructions   string             `json:"instructions"`
	OutputContract StepOutputContract `json:"outputContract"`
	Gate           *GateState         `json:"gate,omitempty"`
	ComposedAgent  ComposedAgentSpec  `json:"composedAgent"`
}

type BuildTeamInput struct {
	Team   string        `json:"team"`
	Task   string        `json:"task"`
	Budget *BudgetPolicy `json:"budget,omitempty"`
}

type AssignTaskInput struct {
	TeamID   string `json:"teamId"`
	TaskID   string `json:"taskId"`
	Assignee string `json:"assignee"`
	Claim    bool   `json:"claim,omitempty"`
}

type CompleteTaskInput struct {
	TeamID  string           `json:"teamId"`
	TaskID  string           `json:"taskId"`
	Outcome string           `json:"outcome"`
	Result  map[string]any   `json:"result,omitempty"`
	Usage   *StepUsage       `json:"usage,omitempty"`
	Error   *StructuredError `json:"error,omitempty"`
}

type StartWorkflowInput struct {
	Workflow    string        `json:"workflow"`
	Task        string        `json:"task"`
	DomainSkill string        `json:"domainSkill,omitempty"`
	ModeSkill   string        `json:"modeSkill,omitempty"`
	Budget      *BudgetPolicy `json:"budget,omitempty"`
	Context     *struct {
		CliTool     HostCli          `json:"cliTool,omitempty"`
		RootContext RootContextLayer `json:"rootContext,omitempty"`
	} `json:"context,omitempty"`
}

type AdvanceWorkflowInput struct {
	WorkflowID string                    `json:"workflowId"`
	Outcome    string                    `json:"outcome,omitempty"`
	Recovery   *WorkflowRecoveryDecision `json:"recovery,omitempty"`
}

// CatalogItem is a summary entry for catalog listings.
type CatalogItem struct {
	Kind        CatalogKind      `json:"kind"`
	Name        string           `json:"name"`
	Source      DefinitionSource `json:"source"`
	Description string           `json:"description"`
	Version     string           `json:"version,omitempty"`
	Path        string           `json:"path"`
}
