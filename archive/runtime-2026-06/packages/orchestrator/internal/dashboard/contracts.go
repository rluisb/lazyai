package dashboard

import (
	"errors"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

const (
	DefaultRunLimit   = 50
	MaxRunLimit       = 200
	DefaultErrorLimit = 25
	MaxErrorLimit     = 100
	DefaultEventLimit = 100
	MaxEventLimit     = 500
)

// NormalizeLimit applies dashboard default and maximum bounds to list queries.
func NormalizeLimit(requested, defaultLimit, maxLimit int) int {
	if requested <= 0 {
		return defaultLimit
	}
	if requested > maxLimit {
		return maxLimit
	}
	return requested
}

// DashboardErrorResponse is the stable JSON error envelope for dashboard APIs.
type DashboardErrorResponse struct {
	Error DashboardError `json:"error"`
}

// DashboardError describes an API error without exposing transport details.
type DashboardError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse builds a dashboard error envelope.
func ErrorResponse(code, message string) DashboardErrorResponse {
	return DashboardErrorResponse{Error: DashboardError{Code: code, Message: message}}
}

// NotFoundError marks dashboard read-model/catalog misses for handler mapping.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewNotFoundError creates a not-found error for read adapters.
func NewNotFoundError(resource, id string) error {
	return NotFoundError{Resource: resource, ID: id}
}

// IsNotFound reports whether an error represents a dashboard not-found miss.
func IsNotFound(err error) bool {
	var target NotFoundError
	return errors.As(err, &target)
}

// HealthView mirrors the existing daemon health shape used by dashboard overview.
type HealthView struct {
	Status        string                 `json:"status"`
	Name          string                 `json:"name"`
	Port          int                    `json:"port"`
	PID           int                    `json:"pid"`
	StartedAt     string                 `json:"startedAt"`
	ProjectRoot   string                 `json:"projectRoot,omitempty"`
	Scope         string                 `json:"scope,omitempty"`
	ExecutionMode string                 `json:"executionMode,omitempty"`
	ConfigPath    string                 `json:"configPath,omitempty"`
	Clients       any                    `json:"clients,omitempty"`
	Idle          any                    `json:"idle,omitempty"`
	ActiveRuns    domain.ActiveRunCounts `json:"activeRuns"`
}

// DashboardOverview is the dashboard landing-page read model.
type DashboardOverview struct {
	Health           HealthView             `json:"health"`
	ActiveRuns       domain.ActiveRunCounts `json:"activeRuns"`
	RunCountsByState map[string]int         `json:"runCountsByState"`
	RecentRuns       []RunSummary           `json:"recentRuns"`
	RecentErrors     []ErrorEntry           `json:"recentErrors"`
	CatalogCounts    CatalogCounts          `json:"catalogCounts"`
	GeneratedAt      string                 `json:"generatedAt"`
}

// CatalogCounts summarizes catalog entries by definition kind.
type CatalogCounts struct {
	Total  int            `json:"total"`
	ByKind map[string]int `json:"byKind"`
}

// RunSummary is a bounded list entry for chain/team/workflow runs.
type RunSummary struct {
	Kind              types.RunKind `json:"kind"`
	ID                string        `json:"id"`
	DefinitionName    string        `json:"definitionName"`
	DefinitionVersion string        `json:"definitionVersion,omitempty"`
	State             string        `json:"state"`
	Current           string        `json:"current,omitempty"`
	ProjectRoot       string        `json:"projectRoot"`
	CreatedAt         string        `json:"createdAt"`
	UpdatedAt         string        `json:"updatedAt"`
	BudgetHealth      string        `json:"budgetHealth,omitempty"`
	ErrorCount        int           `json:"errorCount"`
}

// RunListResponse is the read-model shape for paginated run lists.
type RunListResponse struct {
	Items      []RunSummary `json:"items"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

// RunDetail is the read-only detail view for a single run.
type RunDetail struct {
	Summary          RunSummary                 `json:"summary"`
	State            map[string]any             `json:"state,omitempty"`
	StateDecodeError string                     `json:"stateDecodeError,omitempty"`
	Steps            []types.StepState          `json:"steps,omitempty"`
	Tasks            []types.TeamTaskState      `json:"tasks,omitempty"`
	Phases           []types.WorkflowPhaseState `json:"phases,omitempty"`
	Budget           *BudgetView                `json:"budget,omitempty"`
	Events           []DashboardEvent           `json:"events"`
	Errors           []ErrorEntry               `json:"errors"`
	Handoffs         []map[string]any           `json:"handoffs,omitempty"`
	ExecutionPlan    map[string]any             `json:"executionPlan,omitempty"`
}

// RunEventsResponse is the bounded event replay response.
type RunEventsResponse struct {
	Items []DashboardEvent `json:"items"`
}

// DashboardEvent is the persisted event replay payload.
type DashboardEvent struct {
	ID        int            `json:"id"`
	RunID     string         `json:"runId"`
	EventType string         `json:"eventType"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"createdAt"`
}

// BudgetView exposes decoded budget state and health evaluation.
type BudgetView struct {
	State         *types.BudgetState         `json:"state,omitempty"`
	Evaluation    *types.BudgetEvaluation    `json:"evaluation,omitempty"`
	ByStep        map[string]types.StepUsage `json:"byStep,omitempty"`
	LastUpdatedAt string                     `json:"lastUpdatedAt,omitempty"`
	DecodeError   string                     `json:"decodeError,omitempty"`
}

// ErrorEntry is a flattened error journal entry for dashboard lists.
type ErrorEntry struct {
	ID             string        `json:"id"`
	RunID          string        `json:"runId,omitempty"`
	RunKind        types.RunKind `json:"runKind,omitempty"`
	DefinitionName string        `json:"definitionName"`
	StepID         string        `json:"stepId,omitempty"`
	Category       string        `json:"category"`
	Code           string        `json:"code"`
	Message        string        `json:"message"`
	CreatedAt      string        `json:"createdAt"`
}

// ErrorListResponse is the bounded error journal response.
type ErrorListResponse struct {
	Items []ErrorEntry `json:"items"`
}

// CatalogListResponse is the catalog list response.
type CatalogListResponse struct {
	Items []CatalogSummary `json:"items"`
}

// CatalogSummary is a read-only catalog list entry.
type CatalogSummary struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ActiveVersion *int   `json:"activeVersion,omitempty"`
	TotalVersions int    `json:"totalVersions"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CatalogVersion summarizes one immutable catalog version.
type CatalogVersion struct {
	Version   int    `json:"version"`
	Checksum  string `json:"checksum"`
	CreatedAt string `json:"createdAt"`
	CreatedBy string `json:"createdBy,omitempty"`
}

// CatalogDetail exposes active/requested catalog version metadata and body.
type CatalogDetail struct {
	Kind          string           `json:"kind"`
	Name          string           `json:"name"`
	ActiveVersion *int             `json:"activeVersion,omitempty"`
	Version       int              `json:"version"`
	Versions      []CatalogVersion `json:"versions"`
	Frontmatter   map[string]any   `json:"frontmatter"`
	Body          string           `json:"body"`
	Checksum      string           `json:"checksum"`
	CreatedAt     string           `json:"createdAt"`
	CreatedBy     string           `json:"createdBy,omitempty"`
}
