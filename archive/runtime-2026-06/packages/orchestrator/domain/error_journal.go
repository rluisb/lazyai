package domain

// RunRef identifies a persisted chain/team/workflow run without coupling query
// ports to dashboard-specific row shapes.
type RunRef struct {
	Kind string
	ID   string
}

// ErrorJournalEntry represents a flattened persisted error journal entry.
type ErrorJournalEntry struct {
	ID             string `json:"id"`
	RunID          string `json:"runId,omitempty"`
	RunKind        string `json:"runKind,omitempty"`
	DefinitionName string `json:"definitionName"`
	StepID         string `json:"stepId,omitempty"`
	Category       string `json:"category"`
	Code           string `json:"code"`
	Message        string `json:"message"`
	CreatedAt      string `json:"createdAt"`
}
