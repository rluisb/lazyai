package domain

// RunEvent represents a persisted and streamed lifecycle event for a run.
type RunEvent struct {
	ID        int            `json:"id"`
	RunID     string         `json:"runId"`
	Type      string         `json:"eventType"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"createdAt"`
}
