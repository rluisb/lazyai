package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/events"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

const dashboardAPIPrefix = "/api/dashboard"

// HealthSnapshot returns the current daemon health view for overview responses.
type HealthSnapshot func(context.Context) HealthView

// HandlerConfig wires the read-only dashboard HTTP API dependencies.
type HandlerConfig struct {
	ReadModel *ReadModel
	Catalog   *CatalogAdapter
	Events    *events.Bus
	Health    HealthSnapshot
}

// Handler serves the read-only dashboard JSON and per-run event APIs.
type Handler struct {
	readModel *ReadModel
	catalog   *CatalogAdapter
	events    *events.Bus
	health    HealthSnapshot
}

// NewHandler creates a dashboard API handler suitable for mounting under a ServeMux.
func NewHandler(config HandlerConfig) http.Handler {
	return &Handler{
		readModel: config.ReadModel,
		catalog:   config.Catalog,
		events:    config.Events,
		health:    config.Health,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "dashboard API endpoints are read-only and only support GET")
		return
	}
	if !strings.HasPrefix(r.URL.Path, dashboardAPIPrefix) {
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
		return
	}

	segments := splitDashboardPath(strings.TrimPrefix(r.URL.Path, dashboardAPIPrefix))
	if len(segments) == 0 {
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
		return
	}

	switch segments[0] {
	case "overview":
		if len(segments) != 1 {
			h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
			return
		}
		h.handleOverview(w, r)
	case "runs":
		h.handleRuns(w, r, segments)
	case "catalog":
		h.handleCatalog(w, r, segments)
	case "errors":
		if len(segments) != 1 {
			h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
			return
		}
		h.handleErrors(w, r)
	default:
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
	}
}

func (h *Handler) handleOverview(w http.ResponseWriter, r *http.Request) {
	if h.readModel == nil || h.catalog == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard dependencies are not configured")
		return
	}
	health := HealthView{}
	if h.health != nil {
		health = h.health(r.Context())
	}
	catalogCounts, err := h.catalogCounts(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	overview, err := h.readModel.Overview(r.Context(), health, catalogCounts)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, overview)
}

func (h *Handler) handleRuns(w http.ResponseWriter, r *http.Request, segments []string) {
	if h.readModel == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard read model is not configured")
		return
	}
	if len(segments) == 1 {
		h.handleRunList(w, r)
		return
	}
	if len(segments) != 3 && len(segments) != 4 {
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
		return
	}
	kind, ok := parseRunKind(segments[1])
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "run kind must be chain, team, or workflow")
		return
	}
	id, err := unescapePathSegment(segments[2])
	if err != nil || id == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "run id is required")
		return
	}
	if len(segments) == 3 {
		h.handleRunDetail(w, r, kind, id)
		return
	}
	switch segments[3] {
	case "budget":
		h.handleRunBudget(w, r, kind, id)
	case "events":
		h.handleRunEvents(w, r, kind, id)
	default:
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
	}
}

func (h *Handler) handleRunList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	kind, err := parseOptionalRunKind(query.Get("kind"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit, err := parseOptionalNonNegativeInt(query.Get("limit"), "limit")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit = NormalizeLimit(limit, DefaultRunLimit, MaxRunLimit)
	if err := validateOptionalNonNegativeInt(query.Get("cursor"), "cursor"); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	page, err := h.readModel.ListRuns(r.Context(), RunListOptions{
		Kind:   kind,
		State:  query.Get("state"),
		Limit:  limit,
		Cursor: query.Get("cursor"),
	})
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) handleRunDetail(w http.ResponseWriter, r *http.Request, kind types.RunKind, id string) {
	detail, err := h.readModel.GetRunDetail(r.Context(), kind, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) handleRunBudget(w http.ResponseWriter, r *http.Request, kind types.RunKind, id string) {
	budget, err := h.readModel.GetBudget(r.Context(), kind, id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, budget)
}

func (h *Handler) handleCatalog(w http.ResponseWriter, r *http.Request, segments []string) {
	if h.catalog == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard catalog adapter is not configured")
		return
	}
	if len(segments) == 1 {
		h.handleCatalogList(w, r)
		return
	}
	if len(segments) != 3 {
		h.writeError(w, http.StatusNotFound, "not_found", "dashboard API route not found")
		return
	}
	kind, err := unescapePathSegment(segments[1])
	if err != nil || kind == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "catalog kind is required")
		return
	}
	name, err := unescapePathSegment(segments[2])
	if err != nil || name == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "catalog name is required")
		return
	}
	version, err := parseOptionalNonNegativeInt(r.URL.Query().Get("version"), "version")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	detail, err := h.catalog.GetCatalogDetail(r.Context(), kind, name, version)
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) handleCatalogList(w http.ResponseWriter, r *http.Request) {
	items, err := h.catalog.ListCatalog(r.Context(), r.URL.Query().Get("kind"))
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, CatalogListResponse{Items: items})
}

func (h *Handler) handleErrors(w http.ResponseWriter, r *http.Request) {
	if h.readModel == nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard read model is not configured")
		return
	}
	limit, err := parseOptionalNonNegativeInt(r.URL.Query().Get("limit"), "limit")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	limit = NormalizeLimit(limit, DefaultErrorLimit, MaxErrorLimit)
	items, err := h.readModel.ListErrors(r.Context(), ErrorListOptions{RunID: r.URL.Query().Get("run_id"), Limit: limit})
	if err != nil {
		h.handleError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, ErrorListResponse{Items: items})
}

func (h *Handler) catalogCounts(ctx context.Context) (CatalogCounts, error) {
	items, err := h.catalog.ListCatalog(ctx, "")
	if err != nil {
		return CatalogCounts{}, err
	}
	counts := CatalogCounts{ByKind: map[string]int{}}
	for _, item := range items {
		counts.Total++
		counts.ByKind[item.Kind]++
	}
	return counts, nil
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if IsNotFound(err) {
		h.writeError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		h.writeError(w, http.StatusRequestTimeout, "request_cancelled", err.Error())
		return
	}
	h.writeError(w, http.StatusInternalServerError, "internal_error", "dashboard request failed")
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	writeStandaloneJSON(w, status, payload)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	writeStandaloneJSON(w, status, ErrorResponse(code, message))
}

func writeStandaloneError(w http.ResponseWriter, err error) {
	if IsNotFound(err) {
		writeStandaloneJSON(w, http.StatusNotFound, ErrorResponse("not_found", err.Error()))
		return
	}
	writeStandaloneJSON(w, http.StatusInternalServerError, ErrorResponse("internal_error", "dashboard request failed"))
}

func writeStandaloneJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func splitDashboardPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func parseOptionalRunKind(value string) (types.RunKind, error) {
	if value == "" {
		return "", nil
	}
	kind, ok := parseRunKind(value)
	if !ok {
		return "", fmt.Errorf("run kind must be chain, team, or workflow")
	}
	return kind, nil
}

func parseRunKind(value string) (types.RunKind, bool) {
	switch types.RunKind(value) {
	case types.RunKindChain, types.RunKindTeam, types.RunKindWorkflow:
		return types.RunKind(value), true
	default:
		return "", false
	}
}

func parseOptionalNonNegativeInt(value, field string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", field)
	}
	return parsed, nil
}

func validateOptionalNonNegativeInt(value, field string) error {
	_, err := parseOptionalNonNegativeInt(value, field)
	return err
}

func unescapePathSegment(segment string) (string, error) {
	return url.PathUnescape(segment)
}
