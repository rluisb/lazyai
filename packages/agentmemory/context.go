package agentmemory

import (
	"context"
	"database/sql"
	"strings"
)

// ContextBundle is the deterministic memory context for resuming a task.
type ContextBundle struct {
	Task             Task                 `json:"task"`
	LatestCheckpoint *Checkpoint          `json:"latest_checkpoint,omitempty"`
	RecentEvents     []TaskEvent          `json:"recent_events"`
	RecentMemories   []MemorySearchResult `json:"recent_memories"`
	Artifacts        []Artifact           `json:"artifacts"`
}

// BuildContext assembles recent task continuity data from deterministic stores.
func BuildContext(ctx context.Context, db *sql.DB, taskID string) (ContextBundle, error) {
	task, err := NewTaskStore(db).GetTask(ctx, taskID)
	if err != nil {
		return ContextBundle{}, err
	}

	bundle := ContextBundle{Task: task, RecentEvents: []TaskEvent{}, RecentMemories: []MemorySearchResult{}, Artifacts: []Artifact{}}
	checkpoint, err := NewCheckpointStore(db).LatestCheckpoint(ctx, taskID)
	if err != nil && err != sql.ErrNoRows {
		return ContextBundle{}, err
	}
	if err == nil {
		bundle.LatestCheckpoint = &checkpoint
	}

	events, err := NewEventStore(db).ListEvents(ctx, taskID, 0, 20)
	if err != nil {
		return ContextBundle{}, err
	}
	bundle.RecentEvents = events

	artifacts, err := NewArtifactStore(db).ListArtifacts(ctx, taskID, task.Namespace)
	if err != nil {
		return ContextBundle{}, err
	}
	bundle.Artifacts = artifacts

	memoryQuery := strings.TrimSpace(task.Goal + " " + task.Tags)
	if memoryQuery != "" {
		memories, err := SearchMemories(ctx, db, task.Namespace, memoryQuery, 5)
		if err != nil {
			return ContextBundle{}, err
		}
		bundle.RecentMemories = memories
	}

	return bundle, nil
}
