package agentmemory

import (
	"context"
	"database/sql"
)

// ArtifactStore records task artifacts and previews.
type ArtifactStore struct{ db *sql.DB }

func NewArtifactStore(db *sql.DB) *ArtifactStore { return &ArtifactStore{db: db} }

func (s *ArtifactStore) RecordArtifact(ctx context.Context, artifact Artifact) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO artifacts (task_id, namespace, path, content_preview, size_bytes, content_hash, mime_type, tags, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, artifact.TaskID, artifact.Namespace, artifact.Path,
		Redact(artifact.ContentPreview), artifact.SizeBytes, artifact.ContentHash, artifact.MimeType, artifact.Tags, artifact.CreatedAt)
	return err
}

func (s *ArtifactStore) ListArtifacts(ctx context.Context, taskID string, namespace string) ([]Artifact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_id, namespace, path, COALESCE(content_preview, ''), COALESCE(size_bytes, 0),
		       COALESCE(content_hash, ''), COALESCE(mime_type, ''), COALESCE(tags, ''), created_at
		FROM artifacts WHERE task_id = ? AND namespace = ? ORDER BY id ASC`, taskID, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArtifacts(rows)
}

func scanArtifacts(rows *sql.Rows) ([]Artifact, error) {
	artifacts := []Artifact{}
	for rows.Next() {
		var artifact Artifact
		if err := rows.Scan(&artifact.ID, &artifact.TaskID, &artifact.Namespace, &artifact.Path,
			&artifact.ContentPreview, &artifact.SizeBytes, &artifact.ContentHash, &artifact.MimeType, &artifact.Tags, &artifact.CreatedAt); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}
