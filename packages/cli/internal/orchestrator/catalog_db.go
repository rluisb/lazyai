// Package orchestrator provides catalog versioning via the orchestrator SQLite DB.
package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
)

// ---------------------------------------------------------------------------
// Catalog DB types
// ---------------------------------------------------------------------------

// CatalogDefSummary mirrors the TypeScript DefinitionSummary.
type CatalogDefSummary struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"activeVersion,omitempty"`
	TotalVersions int    `json:"totalVersions"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CatalogVersionRow mirrors the TypeScript DefinitionVersionRow (simplified).
type CatalogVersionRow struct {
	ID              int            `json:"id"`
	DefinitionID    int            `json:"definitionId"`
	Kind            string         `json:"kind"`
	Name            string         `json:"name"`
	Version         int            `json:"version"`
	FrontmatterJSON string         `json:"frontmatterJson"`
	Frontmatter     map[string]any `json:"frontmatter"`
	Body            string         `json:"body"`
	Checksum        string         `json:"checksum"`
	CreatedAt       string         `json:"createdAt"`
	CreatedBy       *string        `json:"createdBy,omitempty"`
	IsActive        bool           `json:"isActive,omitempty"`
}

// CreateVersionResult mirrors the TypeScript CreateVersionResult.
type CreateVersionResult struct {
	Version       int    `json:"version"`
	Checksum      string `json:"checksum"`
	AlreadyExists bool   `json:"alreadyExists"`
}

// DiffResult holds two versions for comparison.
type DiffResult struct {
	Kind string             `json:"kind"`
	Name string             `json:"name"`
	From *CatalogVersionRow `json:"from"`
	To   *CatalogVersionRow `json:"to"`
}

// ---------------------------------------------------------------------------
// kind mapping helpers
// ---------------------------------------------------------------------------

// listCategoryToKind maps Go ListCategory to the orchestrator DB kind string.
func listCategoryToKind(cat ListCategory) string {
	switch cat {
	case CategoryChains:
		return "chain"
	case CategoryTeams:
		return "team"
	case CategoryWorkflows:
		return "workflow"
	case CategoryDomains:
		return "skill"
	case CategoryModes:
		return "mode"
	default:
		return string(cat)
	}
}

// kindToListCategory maps orchestrator DB kind string to Go ListCategory.
func kindToListCategory(kind string) ListCategory {
	switch kind {
	case "chain":
		return CategoryChains
	case "team":
		return CategoryTeams
	case "workflow":
		return CategoryWorkflows
	case "skill":
		return CategoryDomains
	case "mode":
		return CategoryModes
	default:
		return ListCategory(kind)
	}
}

// allCatalogKinds returns all orchestrator DB kinds relevant to the Go CLI.
var allCatalogKinds = []string{"chain", "team", "workflow", "skill", "mode", "agent", "command"}

// ---------------------------------------------------------------------------
// CatalogDB — thin wrapper over db.DB
// ---------------------------------------------------------------------------

// CatalogDB provides catalog versioning CRUD backed by the orchestrator SQLite DB.
type CatalogDB struct {
	*db.DB
}

// DefaultCatalogDBPath returns the default path of the orchestrator DB.
func DefaultCatalogDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".local", "share", "lazyai-orchestrator", "orchestrator.db"), nil
}

// OpenCatalogDB opens the orchestrator DB at the given path.
func OpenCatalogDB(dbPath string) (*CatalogDB, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open catalog db: %w", err)
	}
	return &CatalogDB{DB: database}, nil
}

// checksumPayload ensures canonical JSON key ordering matching TypeScript's
// JSON.stringify({ frontmatter, body }).
type checksumPayload struct {
	Frontmatter map[string]any `json:"frontmatter"`
	Body        string         `json:"body"`
}

// catalogChecksum matches the TypeScript orchestrator's checksum algorithm:
// JSON.stringify({ frontmatter, body }) then SHA-256 hex, first 16 chars.
func catalogChecksum(fm map[string]any, body string) string {
	payload := checksumPayload{Frontmatter: fm, Body: body}
	canonical, _ := json.Marshal(payload)
	h := sha256Hash(canonical)
	return h[:16]
}

// sha256Hash computes a SHA-256 hex digest (pure Go stdlib).
func sha256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ListDefinitions lists all catalog definitions, optionally filtered by kind.
func (cdb *CatalogDB) ListDefinitions(kind string) ([]CatalogDefSummary, error) {
	where := ""
	params := []any{}
	if kind != "" {
		where = "WHERE d.kind = ?"
		params = append(params, kind)
	}

	query := fmt.Sprintf(`
		SELECT d.kind, d.name,
			   (SELECT dv.version FROM definition_versions dv WHERE dv.id = d.active_version_id) AS active_version,
			   (SELECT COUNT(*) FROM definition_versions dv WHERE dv.definition_id = d.id) AS total_versions,
			   d.created_at, d.updated_at
		FROM definitions d
		%s
		ORDER BY d.kind, d.name
	`, where)

	rows, err := cdb.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query definitions: %w", err)
	}
	defer rows.Close()

	var results []CatalogDefSummary
	for rows.Next() {
		var s CatalogDefSummary
		var activeVersion *int
		var totalVersions int
		var createdAt, updatedAt string

		if err := rows.Scan(&s.Kind, &s.Name, &activeVersion, &totalVersions, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan definition: %w", err)
		}
		s.ActiveVersion = activeVersion
		s.TotalVersions = totalVersions
		s.CreatedAt = createdAt
		s.UpdatedAt = updatedAt
		results = append(results, s)
	}
	if results == nil {
		results = []CatalogDefSummary{}
	}
	return results, rows.Err()
}

// ListVersions lists all versions for a given kind+name.
func (cdb *CatalogDB) ListVersions(kind, name string) ([]CatalogVersionRow, error) {
	query := `
		SELECT dv.id, dv.definition_id, d.kind, d.name, dv.version,
			   dv.frontmatter_json, dv.body, dv.checksum, dv.created_at, dv.created_by,
			   (dv.id = d.active_version_id) AS is_active
		FROM definition_versions dv
		JOIN definitions d ON d.id = dv.definition_id
		WHERE d.kind = ? AND d.name = ?
		ORDER BY dv.version ASC
	`

	rows, err := cdb.DB.Query(query, kind, name)
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var results []CatalogVersionRow
	for rows.Next() {
		var r CatalogVersionRow
		var fmJSON string
		var createdBy *string

		if err := rows.Scan(&r.ID, &r.DefinitionID, &r.Kind, &r.Name, &r.Version,
			&fmJSON, &r.Body, &r.Checksum, &r.CreatedAt, &createdBy, &r.IsActive); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}

		r.FrontmatterJSON = fmJSON
		if fmJSON != "" {
			var fm map[string]any
			if err := json.Unmarshal([]byte(fmJSON), &fm); err == nil {
				r.Frontmatter = fm
			} else {
				r.Frontmatter = map[string]any{}
			}
		} else {
			r.Frontmatter = map[string]any{}
		}
		r.CreatedBy = createdBy

		results = append(results, r)
	}
	if results == nil {
		results = []CatalogVersionRow{}
	}
	return results, rows.Err()
}

// CreateVersion creates a new version from a file path. Returns the new version info.
func (cdb *CatalogDB) CreateVersion(kind, name, filePath, createdBy string, setActive bool) (*CreateVersionResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}
	body := string(data)

	// Build frontmatter from file metadata.
	fm := map[string]any{
		"name": name,
	}

	return cdb.createVersionInternal(kind, name, fm, body, createdBy, setActive)
}

func (cdb *CatalogDB) createVersionInternal(kind, name string, fm map[string]any, body, createdBy string, setActive bool) (*CreateVersionResult, error) {
	cs := catalogChecksum(fm, body)
	now := time.Now().UTC().Format(time.RFC3339)

	// Ensure definition exists.
	defID, err := cdb.ensureDefinition(kind, name, now)
	if err != nil {
		return nil, fmt.Errorf("ensure definition: %w", err)
	}

	// Check for dedup by checksum.
	existingQuery := `SELECT id, version FROM definition_versions WHERE definition_id = ? AND checksum = ?`
	row := cdb.DB.QueryRow(existingQuery, defID, cs)
	var existingID, existingVersion int
	if err := row.Scan(&existingID, &existingVersion); err == nil {
		return &CreateVersionResult{Version: existingVersion, Checksum: cs, AlreadyExists: true}, nil
	}

	// Get next version.
	nextV, err := cdb.nextVersion(defID)
	if err != nil {
		return nil, fmt.Errorf("next version: %w", err)
	}

	fmJSON, err := json.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	var createdByParam any = createdBy
	if createdBy == "" {
		createdByParam = nil
	}

	insertQuery := `
		INSERT INTO definition_versions (definition_id, version, frontmatter_json, body, checksum, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := cdb.DB.Exec(insertQuery, defID, nextV, string(fmJSON), body, cs, now, createdByParam)
	if err != nil {
		return nil, fmt.Errorf("insert version: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	// Set as active if requested, or if this is the first version for the definition
	if setActive || !cdb.hasActiveVersion(defID) {
		updateQuery := `UPDATE definitions SET active_version_id = ?, updated_at = ? WHERE id = ?`
		if _, err := cdb.DB.Exec(updateQuery, newID, now, defID); err != nil {
			return nil, fmt.Errorf("set active version: %w", err)
		}
	}

	return &CreateVersionResult{Version: nextV, Checksum: cs, AlreadyExists: false}, nil
}

// SetActive sets the active version for a definition.
func (cdb *CatalogDB) SetActive(kind, name string, version int) error {
	query := `
		SELECT dv.id FROM definition_versions dv
		JOIN definitions d ON d.id = dv.definition_id
		WHERE d.kind = ? AND d.name = ? AND dv.version = ?
	`
	row := cdb.DB.QueryRow(query, kind, name, version)
	var dvID int
	if err := row.Scan(&dvID); err != nil {
		return fmt.Errorf("version %d of %s/%s not found: %w", version, kind, name, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	updateQuery := `UPDATE definitions SET active_version_id = ?, updated_at = ? WHERE kind = ? AND name = ?`
	if _, err := cdb.DB.Exec(updateQuery, dvID, now, kind, name); err != nil {
		return fmt.Errorf("set active: %w", err)
	}
	return nil
}

// DiffVersions returns two versions for comparison.
func (cdb *CatalogDB) DiffVersions(kind, name string, fromV, toV int) (*DiffResult, error) {
	from, err := cdb.getVersion(kind, name, fromV)
	if err != nil {
		return nil, fmt.Errorf("get from version %d: %w", fromV, err)
	}
	to, err := cdb.getVersion(kind, name, toV)
	if err != nil {
		return nil, fmt.Errorf("get to version %d: %w", toV, err)
	}

	return &DiffResult{
		Kind: kind,
		Name: name,
		From: from,
		To:   to,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (cdb *CatalogDB) ensureDefinition(kind, name, now string) (int, error) {
	// Try insert.
	insertQuery := `INSERT OR IGNORE INTO definitions (kind, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	if _, err := cdb.DB.Exec(insertQuery, kind, name, now, now); err != nil {
		return 0, fmt.Errorf("insert definition: %w", err)
	}

	query := `SELECT id FROM definitions WHERE kind = ? AND name = ?`
	row := cdb.DB.QueryRow(query, kind, name)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("get definition id: %w", err)
	}
	return id, nil
}

func (cdb *CatalogDB) nextVersion(defID int) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM definition_versions WHERE definition_id = ?`
	row := cdb.DB.QueryRow(query, defID)
	var maxV int
	if err := row.Scan(&maxV); err != nil {
		return 0, fmt.Errorf("max version: %w", err)
	}
	return maxV + 1, nil
}

func (cdb *CatalogDB) hasActiveVersion(defID int) bool {
	query := `SELECT active_version_id FROM definitions WHERE id = ?`
	row := cdb.DB.QueryRow(query, defID)
	var avID *int
	if err := row.Scan(&avID); err != nil || avID == nil {
		return false
	}
	return true
}

func (cdb *CatalogDB) getVersion(kind, name string, version int) (*CatalogVersionRow, error) {
	query := `
		SELECT dv.id, dv.definition_id, d.kind, d.name, dv.version,
			   dv.frontmatter_json, dv.body, dv.checksum, dv.created_at, dv.created_by,
			   (dv.id = d.active_version_id) AS is_active
		FROM definition_versions dv
		JOIN definitions d ON d.id = dv.definition_id
		WHERE d.kind = ? AND d.name = ? AND dv.version = ?
	`

	row := cdb.DB.QueryRow(query, kind, name, version)
	var r CatalogVersionRow
	var fmJSON string
	var createdBy *string

	if err := row.Scan(&r.ID, &r.DefinitionID, &r.Kind, &r.Name, &r.Version,
		&fmJSON, &r.Body, &r.Checksum, &r.CreatedAt, &createdBy, &r.IsActive); err != nil {
		return nil, err
	}

	r.FrontmatterJSON = fmJSON
	if fmJSON != "" {
		var fm map[string]any
		if err := json.Unmarshal([]byte(fmJSON), &fm); err == nil {
			r.Frontmatter = fm
		} else {
			r.Frontmatter = map[string]any{}
		}
	} else {
		r.Frontmatter = map[string]any{}
	}
	r.CreatedBy = createdBy

	return &r, nil
}

// formatDate truncates an ISO date string to just the date part.
func formatDate(iso string) string {
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

// FormatCatalogDate is the public wrapper for formatDate used by CLI formatting.
func FormatCatalogDate(iso string) string {
	return formatDate(iso)
}

// countBodySteps counts steps/agents/phases in the body based on kind.
func countBodySteps(kind string, body string) int {
	var data map[string]any
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return 0
	}

	switch kind {
	case "chain":
		if steps, ok := data["steps"].([]any); ok {
			return len(steps)
		}
	case "team":
		if parallel, ok := data["parallel"].([]any); ok {
			return len(parallel)
		}
	case "workflow":
		if phases, ok := data["phases"].([]any); ok {
			return len(phases)
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Formatting helpers for CLI output
// ---------------------------------------------------------------------------

// FormatCatalogList formats a list of definitions for display.
func FormatCatalogList(defs []CatalogDefSummary) string {
	if len(defs) == 0 {
		return "  (no catalog definitions found)\n"
	}

	// Group by kind for cleaner output
	byKind := make(map[string][]CatalogDefSummary)
	for _, d := range defs {
		byKind[d.Kind] = append(byKind[d.Kind], d)
	}

	kindOrder := []string{"chain", "team", "workflow", "skill", "mode", "agent", "command"}
	var lines []string
	for _, kind := range kindOrder {
		entries, ok := byKind[kind]
		if !ok {
			continue
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

		for _, e := range entries {
			activeStr := "none"
			if e.ActiveVersion != nil {
				activeStr = fmt.Sprintf("v%d", *e.ActiveVersion)
			}
			line := fmt.Sprintf("  %-12s  %-20s  (active: %s)", kind, e.Name, activeStr)
			if e.TotalVersions > 0 {
				line += fmt.Sprintf("  [%d version(s)]", e.TotalVersions)
			}
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// FormatVersionList formats a list of versions for display.
func FormatVersionList(versions []CatalogVersionRow) string {
	if len(versions) == 0 {
		return "  (no versions)\n"
	}

	var lines []string
	for _, v := range versions {
		marker := " "
		if v.IsActive {
			marker = "*"
		}
		date := formatDate(v.CreatedAt)
		steps := countBodySteps(v.Kind, v.Body)
		stepsStr := ""
		if steps > 0 {
			stepsStr = fmt.Sprintf(" (%d steps)", steps)
		}
		line := fmt.Sprintf("  %s v%d  — %s%s", marker, v.Version, date, stepsStr)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}
