package dashboard

import (
	"context"
	"encoding/json"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
)

// CatalogAdapter converts catalog.Store read methods into dashboard contracts.
type CatalogAdapter struct {
	store *catalog.Store
}

// NewCatalogAdapter creates a read-only dashboard catalog adapter.
func NewCatalogAdapter(store *catalog.Store) *CatalogAdapter {
	return &CatalogAdapter{store: store}
}

// ListCatalog returns catalog summaries, optionally filtered by kind.
func (a *CatalogAdapter) ListCatalog(ctx context.Context, kind string) ([]CatalogSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	items, err := a.store.List(kind)
	if err != nil {
		return nil, err
	}
	summaries := make([]CatalogSummary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, CatalogSummary{
			Kind:          item.Kind,
			Name:          item.Name,
			ActiveVersion: item.ActiveVersion,
			TotalVersions: item.TotalVersions,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return summaries, nil
}

// GetCatalogDetail returns the active or requested immutable version detail.
func (a *CatalogAdapter) GetCatalogDetail(ctx context.Context, kind, name string, version int) (*CatalogDetail, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	selected, err := a.store.GetVersion(kind, name, version)
	if err != nil {
		return nil, NewNotFoundError("catalog definition", kind+"/"+name)
	}
	versions, err := a.store.ListVersions(kind, name)
	if err != nil {
		return nil, NewNotFoundError("catalog definition", kind+"/"+name)
	}

	var activeVersion *int
	summaries, err := a.store.List(kind)
	if err != nil {
		return nil, err
	}
	for _, summary := range summaries {
		if summary.Name == name {
			activeVersion = summary.ActiveVersion
			break
		}
	}

	frontmatter := map[string]any{}
	if selected.FrontmatterJSON != "" {
		_ = json.Unmarshal([]byte(selected.FrontmatterJSON), &frontmatter)
	}

	detail := &CatalogDetail{
		Kind:          selected.Kind,
		Name:          selected.Name,
		ActiveVersion: activeVersion,
		Version:       selected.Version,
		Versions:      make([]CatalogVersion, 0, len(versions)),
		Frontmatter:   frontmatter,
		Body:          selected.Body,
		Checksum:      selected.Checksum,
		CreatedAt:     selected.CreatedAt,
		CreatedBy:     selected.CreatedBy,
	}
	for _, version := range versions {
		detail.Versions = append(detail.Versions, CatalogVersion{
			Version:   version.Version,
			Checksum:  version.Checksum,
			CreatedAt: version.CreatedAt,
			CreatedBy: version.CreatedBy,
		})
	}
	return detail, nil
}
