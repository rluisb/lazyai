package types

// DefinitionSource tracks where a catalog definition was loaded from.
type DefinitionSource string

const (
	SourceLibrary     DefinitionSource = "library"
	SourceProject     DefinitionSource = "project"
	SourceDB          DefinitionSource = "db"
	SourceUserProject DefinitionSource = "user_project"
	SourceUserGlobal  DefinitionSource = "user_global"
)

// CatalogKind enumerates the kinds of catalog definitions.
type CatalogKind string

const (
	KindAgent    CatalogKind = "agent"
	KindDomain   CatalogKind = "domain"
	KindMode     CatalogKind = "mode"
	KindChain    CatalogKind = "chain"
	KindTeam     CatalogKind = "team"
	KindWorkflow CatalogKind = "workflow"
)

// HostCli identifies which AI CLI tool is hosting the orchestrator session.
type HostCli string

const (
	HostClaudeCode HostCli = "claude-code"
	HostOpenCode   HostCli = "opencode"
	HostCopilot    HostCli = "copilot"
)

// DispatchMode describes how an agent step is executed.
type DispatchMode string

const (
	DispatchTaskTool        DispatchMode = "task-tool"
	DispatchNativeSubagent  DispatchMode = "native-subagent"
	DispatchSDKSession      DispatchMode = "sdk-session"
	DispatchInstructionOnly DispatchMode = "instruction-only"
	DispatchA2A             DispatchMode = "a2a"
	DispatchEmbedded        DispatchMode = "embedded"
)

// ApprovalPolicy controls how strictly approvals are enforced.
type ApprovalPolicy string

const (
	ApprovalMinimal ApprovalPolicy = "minimal"
	ApprovalNormal  ApprovalPolicy = "normal"
	ApprovalStrict  ApprovalPolicy = "strict"
)

// RunKind identifies the type of orchestration run.
type RunKind string

const (
	RunKindChain    RunKind = "chain"
	RunKindTeam     RunKind = "team"
	RunKindWorkflow RunKind = "workflow"
)

// ExecutionMode selects how agents are executed.
type ExecutionMode string

const (
	ExecutionNative   ExecutionMode = "native"
	ExecutionA2A      ExecutionMode = "a2a"
	ExecutionHybrid   ExecutionMode = "hybrid"
	ExecutionEmbedded ExecutionMode = "embedded"
)

// StepType classifies what kind of work a chain step performs.
type StepType string

const (
	StepResearch  StepType = "research"
	StepPlan      StepType = "plan"
	StepImplement StepType = "implement"
	StepReview    StepType = "review"
	StepDocument  StepType = "document"
	StepCustom    StepType = "custom"
)

// Lifecycle states for chains.
type ChainLifecycleState string

const (
	ChainCreated   ChainLifecycleState = "created"
	ChainRunning   ChainLifecycleState = "running"
	ChainGated     ChainLifecycleState = "gated"
	ChainPaused    ChainLifecycleState = "paused"
	ChainCompleted ChainLifecycleState = "completed"
	ChainAbandoned ChainLifecycleState = "abandoned"
	ChainHandoff   ChainLifecycleState = "handoff"
)

// Lifecycle states for chain steps.
type StepLifecycleState string

const (
	StepPending   StepLifecycleState = "pending"
	StepRunning   StepLifecycleState = "running"
	StepCompleted StepLifecycleState = "completed"
	StepFailed    StepLifecycleState = "failed"
	StepRetrying  StepLifecycleState = "retrying"
	StepEscalated StepLifecycleState = "escalated"
	StepSkipped   StepLifecycleState = "skipped"
	StepAbandoned StepLifecycleState = "abandoned"
)

// Lifecycle states for teams.
type TeamLifecycleState string

const (
	TeamCreated      TeamLifecycleState = "created"
	TeamRunning      TeamLifecycleState = "running"
	TeamSynthesizing TeamLifecycleState = "synthesizing"
	TeamCompleted    TeamLifecycleState = "completed"
	TeamFailed       TeamLifecycleState = "failed"
	TeamPaused       TeamLifecycleState = "paused"
	TeamHandoff      TeamLifecycleState = "handoff"
)

// Lifecycle states for team tasks.
type TeamTaskLifecycleState string

const (
	TaskPending   TeamTaskLifecycleState = "pending"
	TaskAssigned  TeamTaskLifecycleState = "assigned"
	TaskClaimed   TeamTaskLifecycleState = "claimed"
	TaskBlocked   TeamTaskLifecycleState = "blocked"
	TaskCompleted TeamTaskLifecycleState = "completed"
	TaskFailed    TeamTaskLifecycleState = "failed"
)

// Lifecycle states for workflows.
type WorkflowLifecycleState string

const (
	WorkflowCreated          WorkflowLifecycleState = "created"
	WorkflowWaitingOnChild   WorkflowLifecycleState = "waiting_on_child"
	WorkflowRunning          WorkflowLifecycleState = "running"
	WorkflowGated            WorkflowLifecycleState = "gated"
	WorkflowAwaitingRecovery WorkflowLifecycleState = "awaiting_recovery"
	WorkflowCompleted        WorkflowLifecycleState = "completed"
	WorkflowFailed           WorkflowLifecycleState = "failed"
	WorkflowHandoff          WorkflowLifecycleState = "handoff"
	WorkflowPaused           WorkflowLifecycleState = "paused"
)

// Lifecycle states for workflow phases.
type WorkflowPhaseLifecycleState string

const (
	PhasePending        WorkflowPhaseLifecycleState = "pending"
	PhaseWaitingOnChild WorkflowPhaseLifecycleState = "waiting_on_child"
	PhaseRunning        WorkflowPhaseLifecycleState = "running"
	PhaseGated          WorkflowPhaseLifecycleState = "gated"
	PhaseCompleted      WorkflowPhaseLifecycleState = "completed"
	PhaseFailed         WorkflowPhaseLifecycleState = "failed"
	PhaseSkipped        WorkflowPhaseLifecycleState = "skipped"
)

// BudgetHealth represents the overall budget health assessment.
type BudgetHealth string

const (
	HealthOK           BudgetHealth = "ok"
	HealthWarning      BudgetHealth = "warning"
	HealthLimitReached BudgetHealth = "limit_reached"
)

// ErrorCategory classifies errors for recovery decisions.
type ErrorCategory string

const (
	ErrorTransient  ErrorCategory = "transient"
	ErrorLogical    ErrorCategory = "logical"
	ErrorBudget     ErrorCategory = "budget"
	ErrorPermission ErrorCategory = "permission"
	ErrorValidation ErrorCategory = "validation"
	ErrorFatal      ErrorCategory = "fatal"
)

// RecoveryActionType describes how to recover from a failed step.
type RecoveryActionType string

const (
	RecoveryRetry     RecoveryActionType = "retry"
	RecoveryFixResume RecoveryActionType = "fix_and_resume"
	RecoveryEscalate  RecoveryActionType = "escalate"
	RecoveryPause     RecoveryActionType = "pause"
	RecoveryHandoff   RecoveryActionType = "handoff"
	RecoveryAbort     RecoveryActionType = "abort"
)

// GateType enumerates gate kinds (user approval, severity, cost).
type GateType string

const (
	GateUserApproval         GateType = "user_approval"
	GateSeverityConfirmation GateType = "severity_confirmation"
	GateCostConfirmation     GateType = "cost_confirmation"
)

// GateStatus tracks whether a gate is pending, approved, or rejected.
type GateStatus string

const (
	GatePending  GateStatus = "pending"
	GateApproved GateStatus = "approved"
	GateRejected GateStatus = "rejected"
)

// DriftStatus indicates whether tracked artifacts are in sync.
type DriftStatus string

const (
	DriftFresh       DriftStatus = "fresh"
	DriftStale       DriftStatus = "stale"
	DriftStaleAcked  DriftStatus = "stale_acked"
	DriftDisabled    DriftStatus = "disabled"
	DriftUnavailable DriftStatus = "unavailable"
	DriftUnknown     DriftStatus = "unknown"
)

// MaintenanceApprovalScope controls the lifetime of maintenance approvals.
type MaintenanceApprovalScope string

const (
	MaintenancePerAction     MaintenanceApprovalScope = "per_action"
	MaintenanceTaskScoped    MaintenanceApprovalScope = "task_scoped"
	MaintenanceSessionScoped MaintenanceApprovalScope = "session_scoped"
	MaintenanceStanding      MaintenanceApprovalScope = "standing"
)

// ProviderId identifies the LLM provider powering an agent.
type ProviderId string

const (
	ProviderAnthropic ProviderId = "anthropic"
	ProviderOpenAI    ProviderId = "openai"
	ProviderGoogle    ProviderId = "google"
	ProviderOllama    ProviderId = "ollama"
	ProviderDeepSeek  ProviderId = "deepseek"
	ProviderCustom    ProviderId = "custom"
)

// EffortLevel represents the thinking/reasoning effort for an LLM call.
type EffortLevel string

const (
	EffortLow    EffortLevel = "low"
	EffortMedium EffortLevel = "medium"
	EffortHigh   EffortLevel = "high"
	EffortXHigh  EffortLevel = "xhigh"
	EffortMax    EffortLevel = "max"
)
