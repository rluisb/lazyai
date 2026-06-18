package domain

import (
	"errors"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// RunListFilter controls run read-store filtering and pagination.
type RunListFilter struct {
	Kind      types.RunKind
	State     string
	Search    string
	Attention string
	HasErrors bool
	Limit     int
	Cursor    string
}

// RunListPage is a bounded page of chain/team/workflow run rows.
type RunListPage struct {
	Items      []RunRow
	NextCursor string
}

// RunRow is the storage-neutral row shape for chain/team/workflow run reads.
type RunRow struct {
	Kind              types.RunKind
	ID                string
	DefinitionName    string
	DefinitionVersion string
	State             string
	Current           string
	ProjectRoot       string
	StateJSON         string
	CreatedAt         string
	UpdatedAt         string
}

// RunReadNotFoundError marks run read-store misses without coupling adapters to dashboard internals.
type RunReadNotFoundError struct {
	Resource string
	ID       string
}

func (e RunReadNotFoundError) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewRunReadNotFoundError creates a not-found error for run read-store misses.
func NewRunReadNotFoundError(resource, id string) error {
	return RunReadNotFoundError{Resource: resource, ID: id}
}

// IsRunReadNotFound reports whether an error represents a run read-store miss.
func IsRunReadNotFound(err error) bool {
	var target RunReadNotFoundError
	return errors.As(err, &target)
}
