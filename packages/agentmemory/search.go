package agentmemory

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
)

// MemorySearchResult includes a memory row plus FTS ranking metadata.
type MemorySearchResult struct {
	Memory
	Rank    float64 `json:"rank"`
	Snippet string  `json:"snippet"`
}

// ArtifactSearchResult includes an artifact row plus FTS ranking metadata.
type ArtifactSearchResult struct {
	Artifact
	Rank    float64 `json:"rank"`
	Snippet string  `json:"snippet"`
}

var safePrefixTerm = regexp.MustCompile(`^[A-Za-z0-9_]+\*$`)

func SearchMemories(ctx context.Context, db *sql.DB, namespace string, query string, limit int) ([]MemorySearchResult, error) {
	ftsQuery, err := sanitizeFTSQuery(query)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT m.id, m.namespace, m.content, COALESCE(m.source_task_id, ''), COALESCE(m.source_step_id, ''),
		       COALESCE(m.tags, ''), m.importance, m.created_at,
		       bm25(memories_fts) AS rank,
		       snippet(memories_fts, 0, '<mark>', '</mark>', '…', 12) AS snippet
		FROM memories_fts
		JOIN memories m ON memories_fts.rowid = m.rowid
		WHERE memories_fts MATCH ? AND m.namespace = ?
		ORDER BY rank LIMIT ?`, ftsQuery, namespace, normalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []MemorySearchResult{}
	for rows.Next() {
		var result MemorySearchResult
		if err := rows.Scan(&result.ID, &result.Namespace, &result.Content, &result.SourceTaskID,
			&result.SourceStepID, &result.Tags, &result.Importance, &result.CreatedAt, &result.Rank, &result.Snippet); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func SearchArtifacts(ctx context.Context, db *sql.DB, namespace string, query string, limit int) ([]ArtifactSearchResult, error) {
	ftsQuery, err := sanitizeFTSQuery(query)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT a.id, a.task_id, a.namespace, a.path, COALESCE(a.content_preview, ''), COALESCE(a.size_bytes, 0),
		       COALESCE(a.content_hash, ''), COALESCE(a.mime_type, ''), COALESCE(a.tags, ''), a.created_at,
		       bm25(artifacts_fts) AS rank,
		       snippet(artifacts_fts, 0, '<mark>', '</mark>', '…', 12) AS snippet
		FROM artifacts_fts
		JOIN artifacts a ON artifacts_fts.rowid = a.rowid
		WHERE artifacts_fts MATCH ? AND a.namespace = ?
		ORDER BY rank LIMIT ?`, ftsQuery, namespace, normalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []ArtifactSearchResult{}
	for rows.Next() {
		var result ArtifactSearchResult
		if err := rows.Scan(&result.ID, &result.TaskID, &result.Namespace, &result.Path, &result.ContentPreview,
			&result.SizeBytes, &result.ContentHash, &result.MimeType, &result.Tags, &result.CreatedAt, &result.Rank, &result.Snippet); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func sanitizeFTSQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errors.New("search query cannot be empty")
	}
	if strings.HasPrefix(query, `"`) && strings.HasSuffix(query, `"`) && len(query) > 1 {
		return quoteFTSToken(strings.Trim(query, `"`)), nil
	}

	fields := strings.Fields(query)
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if safePrefixTerm.MatchString(field) {
			parts = append(parts, field)
			continue
		}
		parts = append(parts, quoteFTSToken(field))
	}
	return strings.Join(parts, " "), nil
}

func quoteFTSToken(token string) string {
	return `"` + strings.ReplaceAll(token, `"`, `""`) + `"`
}
