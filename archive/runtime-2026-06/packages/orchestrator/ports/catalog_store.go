package ports

import "github.com/rluisb/lazyai/packages/orchestrator/domain"

// CatalogStore is the persistence port for versioned orchestration catalog definitions.
type CatalogStore interface {
	List(kind string) ([]domain.CatalogDefinitionSummary, error)
	ListVersions(kind, name string) ([]domain.CatalogVersion, error)
	GetVersion(kind, name string, version int) (*domain.CatalogVersion, error)
	CreateVersion(input domain.CreateCatalogVersionInput) (*domain.CreateCatalogVersionResult, error)
	SetActive(kind, name string, version int) error
	Deactivate(kind, name string) error
	Remove(kind, name string) error
	GetBody(kind, name string, version int) (string, error)
}
