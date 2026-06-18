package catalog

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

// Store provides versioned catalog storage backed by SQLite.
type Store struct {
	db *db.DB
}

// NewStore creates a new catalog store.
func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

// VersionRow represents a catalog definition version.
type VersionRow struct {
	ID              int    `json:"id"`
	DefinitionID    int    `json:"definitionId"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	Version         int    `json:"version"`
	FrontmatterJSON string `json:"frontmatterJson"`
	Body            string `json:"body"`
	Checksum        string `json:"checksum"`
	CreatedAt       string `json:"createdAt"`
	CreatedBy       string `json:"createdBy,omitempty"`
}

// DefinitionSummary is a lightweight catalog listing entry.
type DefinitionSummary struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"activeVersion"`
	TotalVersions int    `json:"totalVersions"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CreateVersionInput describes a new version to create.
type CreateVersionInput struct {
	Kind        string
	Name        string
	Frontmatter map[string]any
	Body        string
	CreatedBy   string
	SetActive   bool
}

// CreateVersionResult returns the outcome of creating a version.
type CreateVersionResult struct {
	Version       int    `json:"version"`
	Checksum      string `json:"checksum"`
	AlreadyExists bool   `json:"alreadyExists"`
}

// ──────────────────────── Public API ──────────────────────────────

// List returns all catalog definitions, optionally filtered by kind.
func (s *Store) List(kind string) ([]DefinitionSummary, error) {
	var query string
	var args []any
	if kind != "" {
		query = `
			SELECT d.kind, d.name, d.active_version, d.created_at, d.updated_at, COUNT(v.id)
			FROM definitions d
			LEFT JOIN definition_versions v ON v.definition_id = d.id
			WHERE d.kind = ?
			GROUP BY d.id, d.kind, d.name, d.active_version, d.created_at, d.updated_at
			ORDER BY d.name`
		args = []any{kind}
	} else {
		query = `
			SELECT d.kind, d.name, d.active_version, d.created_at, d.updated_at, COUNT(v.id)
			FROM definitions d
			LEFT JOIN definition_versions v ON v.definition_id = d.id
			GROUP BY d.id, d.kind, d.name, d.active_version, d.created_at, d.updated_at
			ORDER BY d.kind, d.name`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DefinitionSummary
	for rows.Next() {
		var d DefinitionSummary
		if err := rows.Scan(&d.Kind, &d.Name, &d.ActiveVersion, &d.CreatedAt, &d.UpdatedAt, &d.TotalVersions); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// ListVersions returns all versions for a definition.
func (s *Store) ListVersions(kind, name string) ([]VersionRow, error) {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return nil, fmt.Errorf("definition not found: %s/%s", kind, name)
	}

	rows, err := s.db.Query(
		`SELECT id, definition_id, kind, name, version, frontmatter_json, body, checksum, created_at, COALESCE(created_by,'') FROM definition_versions WHERE definition_id = ? ORDER BY version DESC`, defID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []VersionRow
	for rows.Next() {
		var v VersionRow
		if err := rows.Scan(&v.ID, &v.DefinitionID, &v.Kind, &v.Name, &v.Version, &v.FrontmatterJSON, &v.Body, &v.Checksum, &v.CreatedAt, &v.CreatedBy); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

// GetVersion returns a specific version, or the active version if version is 0.
func (s *Store) GetVersion(kind, name string, version int) (*VersionRow, error) {
	if version > 0 {
		return s.getSpecificVersion(kind, name, version)
	}

	// Get active version
	var activeVersion int
	err := s.db.QueryRow(`SELECT active_version FROM definitions WHERE kind = ? AND name = ?`, kind, name).Scan(&activeVersion)
	if err != nil {
		return nil, fmt.Errorf("definition not found: %s/%s", kind, name)
	}
	if activeVersion == 0 {
		return nil, fmt.Errorf("no active version for %s/%s", kind, name)
	}
	return s.getSpecificVersion(kind, name, activeVersion)
}

// CreateVersion creates a new immutable version with checksum deduplication.
func (s *Store) CreateVersion(input CreateVersionInput) (*CreateVersionResult, error) {
	cs := checksum(input.Frontmatter, input.Body)
	defID := s.ensureDefinition(input.Kind, input.Name)
	nextVer := s.nextVersion(defID)

	// Check for duplicate checksum
	var existingVersion int
	err := s.db.QueryRow(`SELECT version FROM definition_versions WHERE definition_id = ? AND checksum = ? ORDER BY version DESC LIMIT 1`, defID, cs).Scan(&existingVersion)
	if err == nil {
		return &CreateVersionResult{Version: existingVersion, Checksum: cs, AlreadyExists: true}, nil
	}

	fmJSON, _ := json.Marshal(input.Frontmatter)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(
		`INSERT INTO definition_versions (definition_id, kind, name, version, frontmatter_json, body, checksum, created_at, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		defID, input.Kind, input.Name, nextVer, string(fmJSON), input.Body, cs, now, input.CreatedBy)
	if err != nil {
		return nil, err
	}

	if input.SetActive {
		s.setActive(defID, nextVer, now)
	}

	return &CreateVersionResult{Version: nextVer, Checksum: cs, AlreadyExists: false}, nil
}

// SetActive moves the active version pointer.
func (s *Store) SetActive(kind, name string, version int) error {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return fmt.Errorf("definition not found: %s/%s", kind, name)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	s.setActive(defID, version, now)
	return nil
}

// Deactivate clears the active version pointer.
func (s *Store) Deactivate(kind, name string) error {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return fmt.Errorf("definition not found: %s/%s", kind, name)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`UPDATE definitions SET active_version = NULL, updated_at = ? WHERE id = ?`, now, defID)
	return err
}

// Remove deletes a definition and all its versions. Destructive.
func (s *Store) Remove(kind, name string) error {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return fmt.Errorf("definition not found: %s/%s", kind, name)
	}
	_, err := s.db.Exec(`DELETE FROM definition_versions WHERE definition_id = ?`, defID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM definitions WHERE id = ?`, defID)
	return err
}

// GetBody returns the body content of a specific version.
func (s *Store) GetBody(kind, name string, version int) (string, error) {
	v, err := s.GetVersion(kind, name, version)
	if err != nil {
		return "", err
	}
	return v.Body, nil
}

// ──────────────────────── internal helpers ──────────────────────────

func (s *Store) ensureDefinition(kind, name string) int {
	now := time.Now().UTC().Format(time.RFC3339)
	s.db.Exec(`INSERT INTO definitions (kind, name, created_at, updated_at) VALUES (?, ?, ?, ?) ON CONFLICT(kind, name) DO NOTHING`, kind, name, now, now)

	var id int
	s.db.QueryRow(`SELECT id FROM definitions WHERE kind = ? AND name = ?`, kind, name).Scan(&id)
	return id
}

func (s *Store) getDefinitionID(kind, name string) int {
	var id int
	err := s.db.QueryRow(`SELECT id FROM definitions WHERE kind = ? AND name = ?`, kind, name).Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

func (s *Store) nextVersion(defID int) int {
	var maxV int
	s.db.QueryRow(`SELECT COALESCE(MAX(version),0) FROM definition_versions WHERE definition_id = ?`, defID).Scan(&maxV)
	return maxV + 1
}

func (s *Store) setActive(defID, version int, now string) {
	s.db.Exec(`UPDATE definitions SET active_version = ?, updated_at = ? WHERE id = ?`, version, now, defID)
}

func (s *Store) countVersions(kind, name string) int {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return 0
	}
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM definition_versions WHERE definition_id = ?`, defID).Scan(&count)
	return count
}

func (s *Store) getSpecificVersion(kind, name string, version int) (*VersionRow, error) {
	defID := s.getDefinitionID(kind, name)
	if defID == 0 {
		return nil, fmt.Errorf("definition not found: %s/%s", kind, name)
	}

	var v VersionRow
	err := s.db.QueryRow(
		`SELECT id, definition_id, kind, name, version, frontmatter_json, body, checksum, created_at, COALESCE(created_by,'') FROM definition_versions WHERE definition_id = ? AND version = ?`, defID, version).
		Scan(&v.ID, &v.DefinitionID, &v.Kind, &v.Name, &v.Version, &v.FrontmatterJSON, &v.Body, &v.Checksum, &v.CreatedAt, &v.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("version %d not found for %s/%s", version, kind, name)
	}
	return &v, nil
}

func checksum(frontmatter map[string]any, body string) string {
	input := map[string]any{"frontmatter": frontmatter, "body": body}
	b, _ := json.Marshal(input)
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:8])
}
