package agentmemory

// Task types.
const (
	TaskTypeResearch  = "research"
	TaskTypePlan      = "plan"
	TaskTypeImplement = "implement"
	TaskTypeReview    = "review"
	TaskTypeBugfix    = "bugfix"
	TaskTypeHandoff   = "handoff"
)

// Task states.
const (
	TaskStatePending   = "pending"
	TaskStateRunning   = "running"
	TaskStateCompleted = "completed"
	TaskStateFailed    = "failed"
)

// Event types.
const (
	EventTypeTaskCreated   = "task.created"
	EventTypeTaskStarted   = "task.started"
	EventTypeTaskCompleted = "task.completed"
	EventTypeTaskFailed    = "task.failed"
	EventTypeStepStarted   = "step.started"
	EventTypeStepCompleted = "step.completed"
	EventTypeCheckpoint    = "checkpoint.created"
	EventTypeArtifact      = "artifact.recorded"
	EventTypeMemory        = "memory.saved"
)

// Memory importance values.
const (
	ImportanceLow      = "low"
	ImportanceNormal   = "normal"
	ImportanceHigh     = "high"
	ImportanceCritical = "critical"
)

// Task mirrors the tasks table.
type Task struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	ProjectRoot string `json:"project_root"`
	TaskType    string `json:"task_type"`
	State       string `json:"state"`
	CurrentStep string `json:"current_step,omitempty"`
	StateJSON   string `json:"state_json,omitempty"`
	Goal        string `json:"goal,omitempty"`
	Tags        string `json:"tags,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TaskEvent mirrors the task_events table.
type TaskEvent struct {
	ID          int64  `json:"id"`
	TaskID      string `json:"task_id"`
	Namespace   string `json:"namespace"`
	RunID       string `json:"run_id,omitempty"`
	EventType   string `json:"event_type"`
	PayloadJSON string `json:"payload_json,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// Checkpoint mirrors the checkpoints table.
type Checkpoint struct {
	ID        int64  `json:"id"`
	TaskID    string `json:"task_id"`
	Namespace string `json:"namespace"`
	StepID    string `json:"step_id,omitempty"`
	Summary   string `json:"summary,omitempty"`
	StateJSON string `json:"state_json,omitempty"`
	CreatedAt string `json:"created_at"`
}

// Artifact mirrors the artifacts table.
type Artifact struct {
	ID             int64  `json:"id"`
	TaskID         string `json:"task_id"`
	Namespace      string `json:"namespace"`
	Path           string `json:"path"`
	ContentPreview string `json:"content_preview,omitempty"`
	SizeBytes      int64  `json:"size_bytes,omitempty"`
	ContentHash    string `json:"content_hash,omitempty"`
	MimeType       string `json:"mime_type,omitempty"`
	Tags           string `json:"tags,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// Memory mirrors the memories table.
type Memory struct {
	ID           int64  `json:"id"`
	Namespace    string `json:"namespace"`
	Content      string `json:"content"`
	SourceTaskID string `json:"source_task_id,omitempty"`
	SourceStepID string `json:"source_step_id,omitempty"`
	Tags         string `json:"tags,omitempty"`
	Importance   string `json:"importance"`
	CreatedAt    string `json:"created_at"`
}
