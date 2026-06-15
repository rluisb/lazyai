package sqlite

import (
	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.CatalogStore = (*CatalogStore)(nil)

// CatalogStore persists versioned orchestration catalog definitions in SQLite.
type CatalogStore struct {
	store *catalog.Store
}

// NewCatalogStore creates a SQLite-backed catalog store adapter.
func NewCatalogStore(database *db.DB) *CatalogStore {
	return &CatalogStore{store: catalog.NewStore(database)}
}

// List returns all catalog definitions, optionally filtered by kind.
func (s *CatalogStore) List(kind string) ([]domain.CatalogDefinitionSummary, error) {
	items, err := s.store.List(kind)
	if err != nil {
		return nil, err
	}
	result := make([]domain.CatalogDefinitionSummary, 0, len(items))
	for _, item := range items {
		result = append(result, domain.CatalogDefinitionSummary{
			Kind:          item.Kind,
			Name:          item.Name,
			ActiveVersion: item.ActiveVersion,
			TotalVersions: item.TotalVersions,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return result, nil
}

// ListVersions returns all versions for a definition.
func (s *CatalogStore) ListVersions(kind, name string) ([]domain.CatalogVersion, error) {
	versions, err := s.store.ListVersions(kind, name)
	if err != nil {
		return nil, err
	}
	result := make([]domain.CatalogVersion, 0, len(versions))
	for _, version := range versions {
		result = append(result, catalogVersionToDomain(version))
	}
	return result, nil
}

// GetVersion returns a specific version, or the active version if version is 0.
func (s *CatalogStore) GetVersion(kind, name string, version int) (*domain.CatalogVersion, error) {
	row, err := s.store.GetVersion(kind, name, version)
	if err != nil {
		return nil, err
	}
	result := catalogVersionToDomain(*row)
	return &result, nil
}

// CreateVersion creates a new immutable version with checksum deduplication.
func (s *CatalogStore) CreateVersion(input domain.CreateCatalogVersionInput) (*domain.CreateCatalogVersionResult, error) {
	result, err := s.store.CreateVersion(catalog.CreateVersionInput{
		Kind:        input.Kind,
		Name:        input.Name,
		Frontmatter: input.Frontmatter,
		Body:        input.Body,
		CreatedBy:   input.CreatedBy,
		SetActive:   input.SetActive,
	})
	if err != nil {
		return nil, err
	}
	return &domain.CreateCatalogVersionResult{
		Version:       result.Version,
		Checksum:      result.Checksum,
		AlreadyExists: result.AlreadyExists,
	}, nil
}

// SetActive moves the active version pointer.
func (s *CatalogStore) SetActive(kind, name string, version int) error {
	return s.store.SetActive(kind, name, version)
}

// Deactivate clears the active version pointer.
func (s *CatalogStore) Deactivate(kind, name string) error {
	return s.store.Deactivate(kind, name)
}

// Remove deletes a definition and all its versions.
func (s *CatalogStore) Remove(kind, name string) error {
	return s.store.Remove(kind, name)
}

// GetBody returns the body content of a specific version.
func (s *CatalogStore) GetBody(kind, name string, version int) (string, error) {
	return s.store.GetBody(kind, name, version)
}

func catalogVersionToDomain(version catalog.VersionRow) domain.CatalogVersion {
	return domain.CatalogVersion{
		ID:              version.ID,
		DefinitionID:    version.DefinitionID,
		Kind:            version.Kind,
		Name:            version.Name,
		Version:         version.Version,
		FrontmatterJSON: version.FrontmatterJSON,
		Body:            version.Body,
		Checksum:        version.Checksum,
		CreatedAt:       version.CreatedAt,
		CreatedBy:       version.CreatedBy,
	}
}
