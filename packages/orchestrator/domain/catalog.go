package domain

// CatalogVersion represents one immutable catalog definition version.
type CatalogVersion struct {
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

// CatalogDefinitionSummary is a lightweight catalog listing entry.
type CatalogDefinitionSummary struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"activeVersion"`
	TotalVersions int    `json:"totalVersions"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CreateCatalogVersionInput describes a new catalog version to create.
type CreateCatalogVersionInput struct {
	Kind        string
	Name        string
	Frontmatter map[string]any
	Body        string
	CreatedBy   string
	SetActive   bool
}

// CreateCatalogVersionResult returns the outcome of creating a catalog version.
type CreateCatalogVersionResult struct {
	Version       int    `json:"version"`
	Checksum      string `json:"checksum"`
	AlreadyExists bool   `json:"alreadyExists"`
}
