// Package migration provides functionality to import existing AI tool
// configurations into the ai-setup canonical format with intelligent
// detection, planning, and execution.
// Ported from the TypeScript migration engine in src/migration/.
package migration

// ---------------------------------------------------------------------------
// Enum-like string types
// ---------------------------------------------------------------------------

// MergeStrategy defines how to handle file merges during migration.
type MergeStrategy string

const (
	MergeStrategySmart    MergeStrategy = "smart"
	MergeStrategyPreserve MergeStrategy = "preserve"
	MergeStrategyReplace  MergeStrategy = "replace"
	MergeStrategyAppend   MergeStrategy = "append"
)

// ActionType defines the kind of migration action.
type ActionType string

const (
	ActionTypeCreate ActionType = "create"
	ActionTypeModify ActionType = "modify"
	ActionTypeBackup ActionType = "backup"
	ActionTypeSkip   ActionType = "skip"
)

// DetectedFileType categorizes a detected file.
type DetectedFileType string

const (
	FileTypeConfig   DetectedFileType = "config"
	FileTypeAgent    DetectedFileType = "agent"
	FileTypeRule     DetectedFileType = "rule"
	FileTypeTemplate DetectedFileType = "template"
	FileTypeCommand  DetectedFileType = "command"
	FileTypeOther    DetectedFileType = "other"
)

// ---------------------------------------------------------------------------
// Core structs
// ---------------------------------------------------------------------------

// MigrationContext carries all the information needed for a migration run.
type MigrationContext struct {
	SourcePath string
	TargetPath string
	Options    MigrationOptions
}

// MigrationOptions controls migration behavior.
type MigrationOptions struct {
	Preview       bool
	MergeStrategy MergeStrategy
	Verbose       bool
	SkipBackup    bool
	Interactive   bool
}

// DetectionResult describes a detected AI tool setup.
type DetectionResult struct {
	Detected    bool
	Confidence  float64 // 0-1
	AdapterID   string
	AdapterName string
	Files       []DetectedFile
	Metadata    map[string]any
}

// DetectedFile represents a file found during detection.
type DetectedFile struct {
	Path     string
	Type     DetectedFileType
	Priority int
}

// ParsedSetup holds the parsed representation of an AI tool setup.
type ParsedSetup struct {
	ProjectName   string
	Description   string
	Agents        []AgentDefinition
	Rules         []RuleDefinition
	Commands      []CommandDefinition
	Templates     []TemplateDefinition
	CustomSections []CustomSection
	Files         []ParsedFile
	Metadata      map[string]string
}

// AgentDefinition represents a parsed agent.
type AgentDefinition struct {
	ID         string
	Name       string
	Content    string
	SourcePath string
}

// RuleDefinition represents a parsed rule.
type RuleDefinition struct {
	ID         string
	Category   string
	Content    string
	SourcePath string
	Priority   int
}

// CommandDefinition represents a parsed command/skill.
type CommandDefinition struct {
	ID         string
	Name       string
	Content    string
	SourcePath string
}

// TemplateDefinition represents a parsed template.
type TemplateDefinition struct {
	ID         string
	Name       string
	Content    string
	SourcePath string
}

// CustomSection represents a parsed custom section.
type CustomSection struct {
	ID         string
	Title      string
	Content    string
	SourcePath string
}

// ParsedFile represents a file from the source setup.
type ParsedFile struct {
	Path    string
	Content string
	Type    string
}

// MergeConflict describes a conflict during migration merging.
type MergeConflict struct {
	File            string
	LineStart       int
	LineEnd         int
	BaseContent     string
	OursContent     string
	TheirsContent   string
	Resolved        bool
	Resolution      string
	ResolvedContent string
}

// MigrationPlan describes the planned migration actions.
type MigrationPlan struct {
	SourcePath        string
	TargetPath        string
	Adapters          []string
	Actions           []MigrationAction
	Conflicts         []MergeConflict
	EstimatedFiles    int
	EstimatedConflicts int
	CanProceed        bool
}

// MigrationAction describes a single file operation in a migration plan.
type MigrationAction struct {
	Type       ActionType
	SourcePath string
	TargetPath string
	Description string
	Reason     string
}

// MigrationResult holds the outcome of a migration.
type MigrationResult struct {
	Success        bool
	Plan           *MigrationPlan
	ExecutedActions []MigrationAction
	BackupPath     string
	Errors         []string
	Warnings       []string
	Stats          MigrationStats
}

// MigrationStats holds statistics about a migration run.
type MigrationStats struct {
	FilesCreated      int
	FilesModified     int
	FilesBackedUp     int
	FilesSkipped      int
	ConflictsResolved int
	ConflictsUnresolved int
}
