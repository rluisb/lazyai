// Package types defines the core domain types for the ai-setup project.
// These types are ported from the TypeScript definitions in src/types.ts and
// src/store/schema.ts, and must maintain JSON compatibility with .ai-setup.json.
package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// Enum-like string types and constants
// ---------------------------------------------------------------------------

// SetupScope defines the scope of an ai-setup installation.
type SetupScope string

const (
	SetupScopeGlobal    SetupScope = "global"
	SetupScopeWorkspace SetupScope = "workspace"
	SetupScopeProject   SetupScope = "project"
)

// SetupPolicy defines how init should handle an existing setup.
type SetupPolicy string

const (
	SetupPolicyAbsorb     SetupPolicy = "absorb"
	SetupPolicyAdapt      SetupPolicy = "adapt"
	SetupPolicyBackupOnly SetupPolicy = "backup-only"
)

// IsValidSetupPolicy reports whether policy is a recognized setup policy.
func IsValidSetupPolicy(policy SetupPolicy) bool {
	switch policy {
	case SetupPolicyAbsorb, SetupPolicyAdapt, SetupPolicyBackupOnly:
		return true
	default:
		return false
	}
}

// ToolId identifies a supported AI coding tool.
type ToolId string

const (
	ToolIdOpenCode   ToolId = "opencode"
	ToolIdClaudeCode ToolId = "claude-code"
	ToolIdCopilot    ToolId = "copilot"
)

// AgentId identifies a specialized agent role.
type AgentId string

const (
	AgentIdBuilder      AgentId = "builder"
	AgentIdPrimaryAgent AgentId = "primary-agent"
	AgentIdPlanner      AgentId = "planner"
	AgentIdReviewer     AgentId = "reviewer"
	AgentIdScout        AgentId = "scout"
)

// SkillId identifies a workflow skill.
type SkillId string

const (
	SkillIdAntiSpeculation     SkillId = "anti-speculation"
	SkillIdBugfix              SkillId = "bugfix"
	SkillIdCodebaseExploration SkillId = "codebase-exploration"
	SkillIdDiagnose            SkillId = "diagnose"
	SkillIdExtractStandards    SkillId = "extract-standards"
	SkillIdHousekeeping        SkillId = "housekeeping"
	SkillIdImpactCheck         SkillId = "impact-check"
	SkillIdImplement           SkillId = "implement"
	SkillIdIterate             SkillId = "iterate"
	SkillIdMemoryWrite         SkillId = "memory-write"
	SkillIdParallelExecution   SkillId = "parallel-execution"
	SkillIdPlan                SkillId = "plan"
	SkillIdPrReview            SkillId = "pr-review"
	SkillIdProcessAudit        SkillId = "process-audit"
	SkillIdProofOfConcept      SkillId = "proof-of-concept"
	SkillIdResearch            SkillId = "research"
	SkillIdReview              SkillId = "review"
	SkillIdRpi                 SkillId = "rpi"
	SkillIdSelfImprove         SkillId = "self-improve"
	SkillIdSpeckitAnalyze      SkillId = "speckit-analyze"
	SkillIdSpeckitChecklist    SkillId = "speckit-checklist"
	SkillIdSpeckitClarify      SkillId = "speckit-clarify"
	SkillIdSpeckitConstitute   SkillId = "speckit-constitution"
	SkillIdSpeckitImplement    SkillId = "speckit-implement"
	SkillIdSpeckitPlan         SkillId = "speckit-plan"
	SkillIdSpeckitSpecify      SkillId = "speckit-specify"
	SkillIdSpeckitTasks        SkillId = "speckit-tasks"
	SkillIdSpike               SkillId = "spike"
	SkillIdTddLoop             SkillId = "tdd-loop"
	SkillIdTestFirstChange     SkillId = "test-first-change"
	SkillIdUpdateMemory        SkillId = "update-memory"
)

// PromptId identifies a reusable prompt.
type PromptId string

const (
	PromptIdCompact      PromptId = "compact"
	PromptIdImplement    PromptId = "implement"
	PromptIdLocalExample PromptId = "local-example"
	PromptIdPlan         PromptId = "plan"
	PromptIdResearch     PromptId = "research"
)

// ChatModeId identifies a GitHub Copilot chat mode (markdown file under
// <githubDir>/chatmodes/).
type ChatModeId string

const (
	ChatModeIdArchitect ChatModeId = "architect"
	ChatModeIdReviewer  ChatModeId = "reviewer"
)

// OpenCodeCommandId identifies an opencode slash command (markdown file
// under <opencodeRoot>/commands/).
type OpenCodeCommandId string

const (
	OpenCodeCommandIdReview            OpenCodeCommandId = "review"
	OpenCodeCommandIdTest              OpenCodeCommandId = "test"
	OpenCodeCommandIdCommit            OpenCodeCommandId = "commit"
	OpenCodeCommandIdSpeckitAnalyze    OpenCodeCommandId = "speckit.analyze"
	OpenCodeCommandIdSpeckitChecklist  OpenCodeCommandId = "speckit.checklist"
	OpenCodeCommandIdSpeckitClarify    OpenCodeCommandId = "speckit.clarify"
	OpenCodeCommandIdSpeckitConstitute OpenCodeCommandId = "speckit.constitution"
	OpenCodeCommandIdSpeckitImplement  OpenCodeCommandId = "speckit.implement"
	OpenCodeCommandIdSpeckitPlan       OpenCodeCommandId = "speckit.plan"
	OpenCodeCommandIdSpeckitSpecify    OpenCodeCommandId = "speckit.specify"
	OpenCodeCommandIdSpeckitTasks      OpenCodeCommandId = "speckit.tasks"
)

// OpenCodeModeId identifies an opencode chat mode (markdown file under
// <opencodeRoot>/modes/). Distinct keyspace from Copilot's ChatModeId.
type OpenCodeModeId string

const (
	OpenCodeModeIdPlan  OpenCodeModeId = "plan"
	OpenCodeModeIdAudit OpenCodeModeId = "audit"
)

// TemplateId identifies a document template.
type TemplateId string

const (
	TemplateIdAdr                TemplateId = "adr"
	TemplateIdBugfixRcaTemplate  TemplateId = "bugfix-rca-template"
	TemplateIdChecklistTemplate  TemplateId = "checklist-template"
	TemplateIdCodeReviewTemplate TemplateId = "code-review-template"
	TemplateIdPlanTemplate       TemplateId = "plan-template"
	TemplateIdPostmortemTemplate TemplateId = "postmortem-template"
	TemplateIdSpecTemplate       TemplateId = "spec-template"
	TemplateIdStandard           TemplateId = "standard"
	TemplateIdTask               TemplateId = "task"
	TemplateIdTechDebtTemplate   TemplateId = "tech-debt-template"
)

// RuleId identifies a project rule.
type RuleId string

const (
	RuleIdAccess        RuleId = "access"
	RuleIdAgentSecurity RuleId = "agent-security"
	RuleIdCodeStyle     RuleId = "code-style"
	RuleIdCost          RuleId = "cost"
	RuleIdReview        RuleId = "review"
	RuleIdSecurity      RuleId = "security"
	RuleIdTesting       RuleId = "testing"
	RuleIdToolUse       RuleId = "tool-use"
	RuleIdWorkflow      RuleId = "workflow"
)

// InfraId identifies an infrastructure component.
type InfraId string

const (
	InfraIdPreCommit    InfraId = "pre-commit"
	InfraIdCompliance   InfraId = "compliance"
	InfraIdKnowledgeMap InfraId = "KNOWLEDGE_MAP"
	InfraIdCodeowners   InfraId = "codeowners"
)

// ArtifactType identifies the kind of an artifact.
type ArtifactType string

const (
	ArtifactTypeAgent               ArtifactType = "agent"
	ArtifactTypeSkill               ArtifactType = "skill"
	ArtifactTypeCommand             ArtifactType = "command"
	ArtifactTypePrompt              ArtifactType = "prompt"
	ArtifactTypeTemplate            ArtifactType = "template"
	ArtifactTypeWorkflow            ArtifactType = "workflow"
	ArtifactTypeChain               ArtifactType = "chain"
	ArtifactTypeTeam                ArtifactType = "team"
	ArtifactTypeDomain              ArtifactType = "domain"
	ArtifactTypeMode                ArtifactType = "mode"
	ArtifactTypeMemoryNote          ArtifactType = "memory_note"
	ArtifactTypeMaintenanceContract ArtifactType = "maintenance_contract"
	ArtifactTypeSyncStateSnapshot   ArtifactType = "sync_state_snapshot"
)

// ApprovalScope defines the persistence scope for maintenance approvals.
type ApprovalScope string

const (
	ApprovalScopePerAction     ApprovalScope = "per_action"
	ApprovalScopeTaskScoped    ApprovalScope = "task_scoped"
	ApprovalScopeSessionScoped ApprovalScope = "session_scoped"
	ApprovalScopeStanding      ApprovalScope = "standing"
)

// Spec006Metadata captures the standardized metadata schema introduced by Spec 006.
type Spec006Metadata struct {
	SchemaVersion       int           `json:"schemaVersion" yaml:"schema_version"`
	ArtifactType        ArtifactType  `json:"artifactType" yaml:"artifact_type"`
	ID                  string        `json:"id" yaml:"id"`
	Title               string        `json:"title,omitempty" yaml:"title,omitempty"`
	TicketNumber        *string       `json:"ticketNumber,omitempty" yaml:"ticket_number,omitempty"`
	Status              string        `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt           string        `json:"createdAt" yaml:"created_at"`
	UpdatedAt           string        `json:"updatedAt" yaml:"updated_at"`
	CreatedBy           string        `json:"createdBy" yaml:"created_by"`
	UpdatedBy           string        `json:"updatedBy" yaml:"updated_by"`
	RiskLevel           string        `json:"riskLevel,omitempty" yaml:"risk_level,omitempty"`
	SizePoints          *int          `json:"sizePoints,omitempty" yaml:"size_points,omitempty"`
	ComplexityLevel     *string       `json:"complexityLevel,omitempty" yaml:"complexity_level,omitempty"`
	SessionID           *string       `json:"sessionId,omitempty" yaml:"session_id,omitempty"`
	WorkflowID          *string       `json:"workflowId,omitempty" yaml:"workflow_id,omitempty"`
	WorkflowRunID       *string       `json:"workflowRunId,omitempty" yaml:"workflow_run_id,omitempty"`
	TeamID              *string       `json:"teamId,omitempty" yaml:"team_id,omitempty"`
	ChainID             *string       `json:"chainId,omitempty" yaml:"chain_id,omitempty"`
	OwnerAgent          *string       `json:"ownerAgent,omitempty" yaml:"owner_agent,omitempty"`
	Assignee            *string       `json:"assignee,omitempty" yaml:"assignee,omitempty"`
	StepIDs             []string      `json:"stepIds,omitempty" yaml:"step_ids,omitempty"`
	WorkflowSteps       []string      `json:"workflowSteps,omitempty" yaml:"workflow_steps,omitempty"`
	RelatedDocumentRefs []string      `json:"relatedDocumentRefs,omitempty" yaml:"related_document_refs,omitempty"`
	ApprovalScope       ApprovalScope `json:"approvalScope,omitempty" yaml:"approval_scope,omitempty"`
	ApprovalExpiresAt   *string       `json:"approvalExpiresAt,omitempty" yaml:"approval_expires_at,omitempty"`
	LegacyMetadataGaps  []string      `json:"legacyMetadataGaps,omitempty" yaml:"legacy_metadata_gaps,omitempty"`
	MigrationNotes      []string      `json:"migrationNotes,omitempty" yaml:"migration_notes,omitempty"`
}

type MemoryEntry struct {
	EntryID             string   `json:"entryId" yaml:"entry_id"`
	Timestamp           string   `json:"timestamp" yaml:"timestamp"`
	Author              string   `json:"author" yaml:"author"`
	EntryType           string   `json:"entryType" yaml:"entry_type"`
	Supersedes          *string  `json:"supersedes,omitempty" yaml:"supersedes,omitempty"`
	RelatedDocumentRefs []string `json:"relatedDocumentRefs,omitempty" yaml:"related_document_refs,omitempty"`
	Content             string   `json:"content" yaml:"content"`
}

type MaintenanceContract struct {
	Spec006Metadata
	PermittedActions []string `json:"permittedActions" yaml:"permitted_actions"`
}

type SyncAcknowledgement struct {
	Fingerprint string `json:"fingerprint"`
	AckedAt     string `json:"ackedAt"`
	Reason      string `json:"reason,omitempty"`
}

type RepairProposal struct {
	Tool       string `json:"tool"`
	ProposalID string `json:"proposalId"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	Reason     string `json:"reason"`
}

type ToolSyncState struct {
	Enabled           bool   `json:"enabled"`
	IndexPath         string `json:"indexPath,omitempty"`
	DataPath          string `json:"dataPath,omitempty"`
	LastIndexTime     string `json:"lastIndexTime,omitempty"`
	SourceFingerprint string `json:"sourceFingerprint,omitempty"`
	DriftStatus       string `json:"driftStatus,omitempty"`
}

type SyncState struct {
	SchemaVersion int           `json:"schemaVersion"`
	UpdatedAt     string        `json:"updatedAt"`
	QMD           ToolSyncState `json:"qmd"`
	Codegraph     ToolSyncState `json:"codegraph"`
	Graphify      ToolSyncState `json:"graphify"`
	StaleAcked    struct {
		QMD       []SyncAcknowledgement `json:"qmd"`
		Codegraph []SyncAcknowledgement `json:"codegraph"`
		Graphify  []SyncAcknowledgement `json:"graphify"`
	} `json:"staleAcked"`
	RepairProposals []RepairProposal `json:"repairProposals,omitempty"`
}

type HousekeepingConfig struct {
	MemoryPath        string `json:"memoryPath,omitempty"`
	EnableObsidian    bool   `json:"enableObsidian,omitempty"`
	ObsidianVaultPath string `json:"obsidianVaultPath,omitempty"`
	EnableQmd         bool   `json:"enableQmd,omitempty"`
	QmdIndexPath      string `json:"qmdIndexPath,omitempty"`
	EnableCodegraph   bool   `json:"enableCodegraph,omitempty"`
	CodegraphDataPath string `json:"codegraphDataPath,omitempty"`
	EnableGraphify    bool   `json:"enableGraphify,omitempty"`
	GraphifyDataPath  string `json:"graphifyDataPath,omitempty"`
}

// PresetLevel defines the preset density level.
type PresetLevel string

const (
	PresetLevelMinimal  PresetLevel = "minimal"
	PresetLevelStandard PresetLevel = "standard"
	PresetLevelFull     PresetLevel = "full"
	PresetLevelCustom   PresetLevel = "custom"
)

// FileOwner indicates who owns a tracked file.
type FileOwner string

const (
	FileOwnerLibrary  FileOwner = "library"
	FileOwnerUser     FileOwner = "user"
	FileOwnerMigrated FileOwner = "migrated"
)

// String implements fmt.Stringer for FileOwner.
func (o FileOwner) String() string { return string(o) }

// FileStatus indicates the current status of a tracked file.
type FileStatus string

const (
	FileStatusInstalled FileStatus = "installed"
	FileStatusModified  FileStatus = "modified"
	FileStatusMissing   FileStatus = "missing"
	FileStatusConflict  FileStatus = "conflict"
)

// String implements fmt.Stringer for FileStatus.
func (s FileStatus) String() string { return string(s) }

// OperationResult indicates the outcome of an operation.
type OperationResult string

const (
	OperationResultSuccess OperationResult = "success"
	OperationResultPartial OperationResult = "partial"
	OperationResultFailure OperationResult = "failure"
)

// String implements fmt.Stringer for OperationResult.
func (r OperationResult) String() string { return string(r) }

// ConflictStrategy defines how file conflicts are handled.
type ConflictStrategy string

const (
	ConflictStrategyAlign            ConflictStrategy = "align"
	ConflictStrategyBackupAndReplace ConflictStrategy = "backup-and-replace"
	ConflictStrategySkip             ConflictStrategy = "skip"
)

// ---------------------------------------------------------------------------
// Slice constants — all valid values for each enum type
// ---------------------------------------------------------------------------

var (
	ALL_AGENTS = []AgentId{
		AgentIdPrimaryAgent,
		AgentIdBuilder,
		AgentIdPlanner,
		AgentIdReviewer,
		AgentIdScout,
	}

	ALL_SKILLS = []SkillId{
		SkillIdCodebaseExploration,
		SkillIdTestFirstChange,
		SkillIdDiagnose,
		SkillIdPrReview,
	}

	ALL_PROMPTS = []PromptId{
		PromptIdCompact,
		PromptIdImplement,
		PromptIdLocalExample,
		PromptIdPlan,
		PromptIdResearch,
	}

	ALL_CHATMODES = []ChatModeId{
		ChatModeIdArchitect,
		ChatModeIdReviewer,
	}

	ALL_OPENCODE_COMMANDS = []OpenCodeCommandId{
		OpenCodeCommandIdReview,
		OpenCodeCommandIdTest,
		OpenCodeCommandIdCommit,
		OpenCodeCommandIdSpeckitAnalyze,
		OpenCodeCommandIdSpeckitChecklist,
		OpenCodeCommandIdSpeckitClarify,
		OpenCodeCommandIdSpeckitConstitute,
		OpenCodeCommandIdSpeckitImplement,
		OpenCodeCommandIdSpeckitPlan,
		OpenCodeCommandIdSpeckitSpecify,
		OpenCodeCommandIdSpeckitTasks,
	}

	ALL_OPENCODE_MODES = []OpenCodeModeId{
		OpenCodeModeIdPlan,
		OpenCodeModeIdAudit,
	}

	ALL_TEMPLATES = []TemplateId{
		TemplateIdAdr,
		TemplateIdBugfixRcaTemplate,
		TemplateIdChecklistTemplate,
		TemplateIdCodeReviewTemplate,
		TemplateIdPlanTemplate,
		TemplateIdPostmortemTemplate,
		TemplateIdSpecTemplate,
		TemplateIdStandard,
		TemplateIdTask,
		TemplateIdTechDebtTemplate,
	}

	ALL_RULES = []RuleId{
		RuleIdAccess,
		RuleIdAgentSecurity,
		RuleIdCodeStyle,
		RuleIdCost,
		RuleIdReview,
		RuleIdSecurity,
		RuleIdTesting,
		RuleIdToolUse,
		RuleIdWorkflow,
	}

	ALL_INFRA = []InfraId{
		InfraIdPreCommit,
		InfraIdCompliance,
		InfraIdKnowledgeMap,
		InfraIdCodeowners,
	}

	ALL_SPECS_DIRS = []string{
		"features",
		"bugfixes",
		"refactors",
		"tech-debt",
		"adrs",
		"memory",
		"prompts",
		"standards",
		"templates",
		"rules",
	}
)

// ---------------------------------------------------------------------------
// Schema version
// ---------------------------------------------------------------------------

// CurrentSchemaVersion is the current version of the .ai-setup.json schema.
const CurrentSchemaVersion = 1

// ---------------------------------------------------------------------------
// Structs
// ---------------------------------------------------------------------------

// FeatureFlags controls which features are compiled into the AGENTS.md output.
type FeatureFlags struct {
	ContextEngineering bool `json:"contextEngineering"`
	RPIWorkflow        bool `json:"rpiWorkflow"`
	ChainOfThought     bool `json:"chainOfThought"`
	TreeOfThoughts     bool `json:"treeOfThoughts"`
	ADREnforcement     bool `json:"adrEnforcement"`
	QualityGates       bool `json:"qualityGates"`
	AgentHarness       bool `json:"agentHarness"`
	BugResolution      bool `json:"bugResolution"`
	PivotHandling      bool `json:"pivotHandling"`
	AdversarialDesign  bool `json:"adversarialDesign"`
}

// GitConventions defines branch and commit patterns for the project.
type GitConventions struct {
	BranchPattern string   `json:"branchPattern"`
	CommitPattern string   `json:"commitPattern"`
	Types         []string `json:"types"`
	RequireTicket bool     `json:"requireTicket"`
	TicketPattern string   `json:"ticketPattern"`
}

// FileKind indicates whether a tracked file is a regular file or a symbolic link.
// This aligns with the TypeScript FileRecord.kind field.
type FileKind string

const (
	FileKindFile    FileKind = "file"
	FileKindSymlink FileKind = "symlink"
)

// TrackedFile represents a file managed by ai-setup.
type TrackedFile struct {
	Path          string     `json:"path"`
	Hash          string     `json:"hash"`
	Source        string     `json:"source"`
	Owner         FileOwner  `json:"owner"`
	Status        FileStatus `json:"status,omitempty"`
	InstalledAt   string     `json:"installedAt,omitempty"`
	LastCheckedAt string     `json:"lastCheckedAt,omitempty"`
	Kind          FileKind   `json:"kind,omitempty"`
	LinkTarget    string     `json:"linkTarget,omitempty"`
}

// RepoInfo describes a repository reference within a workspace.
type RepoInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// Config holds the core configuration for an ai-setup installation.
type Config struct {
	SetupScope        SetupScope          `json:"setupScope"`
	SetupType         SetupScope          `json:"setupType,omitempty"`
	Tools             []ToolId            `json:"tools"`
	CLITools          []string            `json:"cliTools,omitempty"`
	EnableServers     []string            `json:"enableServers,omitempty"`
	ProjectName       string              `json:"projectName"`
	WorkspaceName     string              `json:"workspaceName,omitempty"`
	WorkspaceRoot     string              `json:"workspaceRoot,omitempty"`
	TargetDir         string              `json:"targetDir"`
	PlanningDir       string              `json:"planningDir,omitempty"`
	PlanningRepoPath  string              `json:"planningRepoPath,omitempty"`
	Repos             []RepoInfo          `json:"repos,omitempty"`
	GlobalRef         string              `json:"globalRef,omitempty"`
	Housekeeping      *HousekeepingConfig `json:"housekeeping,omitempty"`
	ProjectOverview   string              `json:"projectOverview,omitempty"`
	NamingConventions string              `json:"namingConventions,omitempty"`
	ErrorHandling     string              `json:"errorHandling,omitempty"`
	ApiConventions    string              `json:"apiConventions,omitempty"`
	ImportOrder       string              `json:"importOrder,omitempty"`
	ProtectedBranch   string              `json:"protectedBranch,omitempty"`
	TestCommand       string              `json:"testCommand,omitempty"`
	LintCommand       string              `json:"lintCommand,omitempty"`
	BuildCommand      string              `json:"buildCommand,omitempty"`
	CoverageThreshold int                 `json:"coverageThreshold,omitempty"`
}

// WizardSelections stores the choices made during the setup wizard.
type WizardSelections struct {
	Templates        []TemplateId        `json:"templates"`
	Rules            []RuleId            `json:"rules"`
	Agents           []AgentId           `json:"agents"`
	Skills           []SkillId           `json:"skills"`
	Prompts          []PromptId          `json:"prompts"`
	ChatModes        []ChatModeId        `json:"chatmodes"`
	OpenCodeCommands []OpenCodeCommandId `json:"opencodeCommands"`
	OpenCodeModes    []OpenCodeModeId    `json:"opencodeModes"`
	OpenCodePlugins  []string            `json:"opencodePlugins"`
	// OpenCodeProviders lists provider IDs (e.g., "openai", "ollama-cloud")
	// that OpenCode-side agents may resolve models against. Populated by the
	// wizard from auth.DetectAll results; never includes "anthropic" because
	// the OpenCode catalog denies it. When empty (legacy stores), adapters
	// fall back to a live auth probe at install/compile time.
	OpenCodeProviders []string        `json:"opencodeProviders,omitempty"`
	Infra             []InfraId       `json:"infra"`
	Constitution      []string        `json:"constitution"`
	Features          *FeatureFlags   `json:"features,omitempty"`
	GitConventions    *GitConventions `json:"gitConventions,omitempty"`
}

// Meta stores schema and version metadata for the store file.
type Meta struct {
	SchemaVersion int    `json:"schemaVersion"`
	CLIVersion    string `json:"cliVersion"`
	InstalledAt   string `json:"installedAt"`
	LastUpdatedAt string `json:"lastUpdatedAt"`
}

// Sync tracks the synchronization state of the store.
type Sync struct {
	LastSyncAt string `json:"lastSyncAt"`
	Dirty      bool   `json:"dirty"`
}

// Operation records a single operation performed on the store.
type Operation struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Timestamp     string          `json:"timestamp"`
	FilesAffected []string        `json:"filesAffected"`
	Result        OperationResult `json:"result"`
	BackupPaths   []string        `json:"backupPaths,omitempty"`
	Error         string          `json:"error,omitempty"`
}

// StoreData is the top-level structure for the .ai-setup.json store file.
type StoreData struct {
	Meta       Meta             `json:"meta"`
	Config     Config           `json:"config"`
	Selections WizardSelections `json:"selections"`
	Files      []TrackedFile    `json:"files"`
	Sync       Sync             `json:"sync"`
	Operations []Operation      `json:"operations"`
}

// WizardConfig extends Config with wizard selection state and interactivity flag.
// This type is not serialized directly; it is used at runtime during the wizard.
type WizardConfig struct {
	Config
	Selections  WizardSelections
	Interactive bool
}

// ---------------------------------------------------------------------------
// JSON helpers — ensure empty slices marshal as [] not null
// ---------------------------------------------------------------------------

// MarshalJSON ensures StoreData serializes empty slices as [] rather than null,
// matching the TypeScript defaultStore() behavior.
func (s StoreData) MarshalJSON() ([]byte, error) {
	type alias StoreData // avoid recursion
	a := alias(s)

	if a.Files == nil {
		a.Files = []TrackedFile{}
	}
	if a.Operations == nil {
		a.Operations = []Operation{}
	}
	if a.Config.Tools == nil {
		a.Config.Tools = []ToolId{}
	}
	if a.Selections.Templates == nil {
		a.Selections.Templates = []TemplateId{}
	}
	if a.Selections.Rules == nil {
		a.Selections.Rules = []RuleId{}
	}
	if a.Selections.Agents == nil {
		a.Selections.Agents = []AgentId{}
	}
	if a.Selections.Skills == nil {
		a.Selections.Skills = []SkillId{}
	}
	if a.Selections.Prompts == nil {
		a.Selections.Prompts = []PromptId{}
	}
	if a.Selections.Infra == nil {
		a.Selections.Infra = []InfraId{}
	}
	if a.Selections.Constitution == nil {
		a.Selections.Constitution = []string{}
	}

	return json.Marshal(a)
}

// ---------------------------------------------------------------------------
// Default constructors
// ---------------------------------------------------------------------------

// DefaultFeatureFlags returns FeatureFlags with all features enabled,
// matching the TypeScript defaults from the zod schema.
func DefaultFeatureFlags() FeatureFlags {
	return FeatureFlags{
		ContextEngineering: true,
		RPIWorkflow:        true,
		ChainOfThought:     true,
		TreeOfThoughts:     true,
		ADREnforcement:     true,
		QualityGates:       true,
		AgentHarness:       true,
		BugResolution:      true,
		PivotHandling:      true,
	}
}

// DefaultGitConventions returns GitConventions with the default patterns
// matching the TypeScript zod defaults.
func DefaultGitConventions() GitConventions {
	return GitConventions{
		BranchPattern: "{type}/{ticket}-{description}",
		CommitPattern: "{type}({scope}): {description}",
		Types: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "build", "ci", "chore", "revert",
		},
		RequireTicket: false,
		TicketPattern: "[A-Z]+-[0-9]+",
	}
}

// DefaultStoreData returns a StoreData with sensible defaults, matching the
// TypeScript defaultStore() function.
func DefaultStoreData() StoreData {
	now := time.Now().UTC().Format(time.RFC3339)

	return StoreData{
		Meta: Meta{
			SchemaVersion: CurrentSchemaVersion,
			CLIVersion:    "", // filled by the CLI at runtime
			InstalledAt:   now,
			LastUpdatedAt: now,
		},
		Config: Config{
			SetupScope:  SetupScopeProject,
			Tools:       []ToolId{},
			ProjectName: "",
			TargetDir:   "",
		},
		Selections: WizardSelections{
			Templates:    []TemplateId{},
			Rules:        []RuleId{},
			Agents:       []AgentId{},
			Skills:       []SkillId{},
			Prompts:      []PromptId{},
			Infra:        []InfraId{},
			Constitution: []string{},
		},
		Files: []TrackedFile{},
		Sync: Sync{
			LastSyncAt: now,
			Dirty:      true,
		},
		Operations: []Operation{},
	}
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

// IsValidSetupScope reports whether s is a recognized SetupScope value.
func IsValidSetupScope(s SetupScope) bool {
	switch s {
	case SetupScopeGlobal, SetupScopeWorkspace, SetupScopeProject:
		return true
	default:
		return false
	}
}

// IsValidToolId reports whether t is a recognized ToolId value.
func IsValidToolId(t ToolId) bool {
	switch t {
	case ToolIdOpenCode, ToolIdClaudeCode, ToolIdCopilot:
		return true
	default:
		return false
	}
}

// ParseOperationResult parses a string into an OperationResult, returning an
// error if the value is not recognized.
func ParseOperationResult(s string) (OperationResult, error) {
	switch OperationResult(s) {
	case OperationResultSuccess, OperationResultPartial, OperationResultFailure:
		return OperationResult(s), nil
	default:
		return "", fmt.Errorf("invalid OperationResult: %q", s)
	}
}

// ParseFileOwner parses a string into a FileOwner, returning an error if the
// value is not recognized.
func ParseFileOwner(s string) (FileOwner, error) {
	switch FileOwner(s) {
	case FileOwnerLibrary, FileOwnerUser, FileOwnerMigrated:
		return FileOwner(s), nil
	default:
		return "", fmt.Errorf("invalid FileOwner: %q", s)
	}
}

// ParseFileStatus parses a string into a FileStatus, returning an error if the
// value is not recognized.
func ParseFileStatus(s string) (FileStatus, error) {
	switch FileStatus(s) {
	case FileStatusInstalled, FileStatusModified, FileStatusMissing, FileStatusConflict:
		return FileStatus(s), nil
	default:
		return "", fmt.Errorf("invalid FileStatus: %q", s)
	}
}
