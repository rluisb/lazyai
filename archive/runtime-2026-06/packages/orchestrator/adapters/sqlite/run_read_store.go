package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.RunReadStore = (*RunReadStore)(nil)

const (
	defaultRunReadLimit = 50
	maxRunReadLimit     = 200
)

// RunReadStore queries chain/team/workflow dashboard run rows from SQLite.
type RunReadStore struct {
	database *db.DB
}

// NewRunReadStore creates a SQLite-backed run read store adapter.
func NewRunReadStore(database *db.DB) *RunReadStore {
	return &RunReadStore{database: database}
}

// ListRuns returns a bounded page of chain/team/workflow run rows.
func (s *RunReadStore) ListRuns(ctx context.Context, filter domain.RunListFilter) (domain.RunListPage, error) {
	limit := normalizeRunReadLimit(filter.Limit)
	offset := parseRunReadCursor(filter.Cursor)

	clauses := []string{}
	args := []any{}
	if filter.Kind != "" {
		clauses = append(clauses, "kind = ?")
		args = append(args, string(filter.Kind))
	}
	if filter.State != "" {
		clauses = append(clauses, "state = ?")
		args = append(args, filter.State)
	}
	if filter.Search != "" {
		pattern := "%" + likeEscape(filter.Search) + "%"
		clauses = append(clauses, `(LOWER(id) LIKE LOWER(?) ESCAPE '\' OR LOWER(definition_name) LIKE LOWER(?) ESCAPE '\')`)
		args = append(args, pattern, pattern)
	}
	switch filter.Attention {
	case "running":
		clauses = append(clauses, "state = 'running'")
	case "failed":
		clauses = append(clauses, "state = 'failed'")
	case "gated":
		clauses = append(clauses, "state IN ('gated','paused','awaiting_recovery','waiting_on_child')")
	case "recent":
		clauses = append(clauses, "updated_at >= datetime('now', '-1 hour')")
	case "":
		// no attention filter
	default:
		return domain.RunListPage{}, fmt.Errorf("unknown attention filter: %q", filter.Attention)
	}
	if filter.HasErrors {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM error_journal e WHERE e.run_id = runs.id AND e.run_kind = runs.kind)")
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	query := `
		SELECT kind, id, definition_name, definition_version, state, current, project_root, state_json, created_at, updated_at
		FROM (` + runUnionQuery() + `) AS runs
		` + where + `
		ORDER BY updated_at DESC, id DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit+1, offset)
	rows, err := s.database.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.RunListPage{}, err
	}
	runRows, hasMore, err := scanRunRows(rows, limit)
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return domain.RunListPage{}, err
	}

	page := domain.RunListPage{Items: runRows}
	if hasMore {
		page.NextCursor = strconv.Itoa(offset + limit)
	}
	return page, nil
}

// CountRunsByState returns counts by lifecycle state across chain/team/workflow runs.
func (s *RunReadStore) CountRunsByState(ctx context.Context) (map[string]int, error) {
	rows, err := s.database.QueryContext(ctx, `
		SELECT state, COUNT(*)
		FROM (`+runUnionQuery()+`)
		GROUP BY state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, err
		}
		counts[state] = count
	}
	return counts, rows.Err()
}

// FindRunRow returns one chain/team/workflow run row by kind and ID.
func (s *RunReadStore) FindRunRow(ctx context.Context, kind types.RunKind, id string) (domain.RunRow, error) {
	table, currentColumn, err := runTable(kind)
	if err != nil {
		return domain.RunRow{}, err
	}
	query := fmt.Sprintf(`SELECT '%s' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, %s AS current, project_root, state_json, created_at, updated_at FROM %s WHERE id = ?`, kind, currentColumn, table)
	row, err := scanRunRow(s.database.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.RunRow{}, domain.NewRunReadNotFoundError("run", string(kind)+"/"+id)
		}
		return domain.RunRow{}, err
	}
	return row, nil
}

func runUnionQuery() string {
	return `
		SELECT 'chain' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, COALESCE(current_step_id, '') AS current, project_root, state_json, created_at, updated_at FROM chain_runs
		UNION ALL
		SELECT 'team' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, '' AS current, project_root, state_json, created_at, updated_at FROM team_runs
		UNION ALL
		SELECT 'workflow' AS kind, id, definition_name, COALESCE(definition_version, '') AS definition_version, state, COALESCE(current_phase_id, '') AS current, project_root, state_json, created_at, updated_at FROM workflow_runs
	`
}

func scanRunRows(rows *sql.Rows, limit int) ([]domain.RunRow, bool, error) {
	items := make([]domain.RunRow, 0, limit)
	hasMore := false
	for rows.Next() {
		row, err := scanRunRow(rows)
		if err != nil {
			return nil, false, err
		}
		if len(items) < limit {
			items = append(items, row)
		} else {
			hasMore = true
		}
	}
	return items, hasMore, rows.Err()
}

func scanRunRow(scanner interface{ Scan(dest ...any) error }) (domain.RunRow, error) {
	var row domain.RunRow
	if err := scanner.Scan(&row.Kind, &row.ID, &row.DefinitionName, &row.DefinitionVersion, &row.State, &row.Current, &row.ProjectRoot, &row.StateJSON, &row.CreatedAt, &row.UpdatedAt); err != nil {
		return domain.RunRow{}, err
	}
	return row, nil
}

func runTable(kind types.RunKind) (table string, currentColumn string, err error) {
	switch kind {
	case types.RunKindChain:
		return "chain_runs", "COALESCE(current_step_id, '')", nil
	case types.RunKindTeam:
		return "team_runs", "''", nil
	case types.RunKindWorkflow:
		return "workflow_runs", "COALESCE(current_phase_id, '')", nil
	default:
		return "", "", domain.NewRunReadNotFoundError("run kind", string(kind))
	}
}

// likeEscape escapes SQL LIKE wildcards so user search terms cannot inject pattern metacharacters.
func likeEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func normalizeRunReadLimit(requested int) int {
	if requested <= 0 {
		return defaultRunReadLimit
	}
	if requested > maxRunReadLimit {
		return maxRunReadLimit
	}
	return requested
}

func parseRunReadCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	offset, err := strconv.Atoi(cursor)
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}
