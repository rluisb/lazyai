package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sqliteadapter "github.com/rluisb/lazyai/packages/orchestrator/adapters/sqlite"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/events"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

func TestDashboardHandlerOverviewAndMethodRejection(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-running", "release", "1", "running", "build", chainStateJSON(t, "chain-running", "release", "1", "running", "build"), "2026-05-05T10:00:00Z")
	seedError(t, database, "err-1", "chain-running", types.RunKindChain, "release", "build", "transient", "dispatch_failed", "dispatch failed", "2026-05-05T10:01:00Z")
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"description": "Release chain"}, "body", true)
	createCatalogVersion(t, store, "team", "launch", map[string]any{"description": "Launch team"}, "team body", true)

	handler := newDashboardHTTPHandler(t, database, store)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dashboard/overview", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("overview status = %d body=%s", response.Code, response.Body.String())
	}
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var overview DashboardOverview
	decodeResponse(t, response, &overview)
	if overview.Health.Status != "ok" || overview.ActiveRuns.Chains != 1 || overview.CatalogCounts.Total != 2 || overview.CatalogCounts.ByKind["chain"] != 1 {
		t.Fatalf("overview mismatch: %+v", overview)
	}
	if len(overview.RecentRuns) != 1 || len(overview.RecentErrors) != 1 {
		t.Fatalf("overview recent data mismatch: %+v", overview)
	}

	for _, path := range []string{
		"/api/dashboard/overview",
		"/api/dashboard/runs",
		"/api/dashboard/runs/chain/chain-running",
		"/api/dashboard/runs/chain/chain-running/budget",
		"/api/dashboard/runs/chain/chain-running/events",
		"/api/dashboard/catalog",
		"/api/dashboard/catalog/chain/release",
		"/api/dashboard/errors",
		"/api/dashboard/events",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, nil))
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s status = %d body=%s", path, response.Code, response.Body.String())
		}
		if response.Header().Get("Allow") != http.MethodGet {
			t.Fatalf("POST %s Allow header = %q", path, response.Header().Get("Allow"))
		}
		var errResponse DashboardErrorResponse
		decodeResponse(t, response, &errResponse)
		if errResponse.Error.Code != "method_not_allowed" || errResponse.Error.Message == "" {
			t.Fatalf("POST %s error mismatch: %+v", path, errResponse)
		}
	}
}

func TestDashboardHandlerRunsEndpointFiltersBoundsAndValidatesInputs(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-running", "release", "1", "running", "build", chainStateJSON(t, "chain-running", "release", "1", "running", "build"), "2026-05-05T10:00:00Z")
	seedRun(t, database, types.RunKindChain, "chain-complete", "release", "1", "completed", "done", `{}`, "2026-05-05T10:01:00Z")
	seedRun(t, database, types.RunKindTeam, "team-running", "launch", "1", "running", "", `{}`, "2026-05-05T10:02:00Z")
	handler := newDashboardHTTPHandler(t, database, sqliteadapter.NewCatalogStore(database))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs?kind=chain&state=running&limit=500", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("runs status = %d body=%s", response.Code, response.Body.String())
	}
	var runs RunListResponse
	decodeResponse(t, response, &runs)
	if len(runs.Items) != 1 || runs.Items[0].Kind != types.RunKindChain || runs.Items[0].State != "running" {
		t.Fatalf("filtered runs mismatch: %+v", runs)
	}

	for _, path := range []string{
		"/api/dashboard/runs?kind=bogus",
		"/api/dashboard/runs?limit=not-a-number",
		"/api/dashboard/runs?cursor=-1",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d body=%s", path, response.Code, response.Body.String())
		}
		var errResponse DashboardErrorResponse
		decodeResponse(t, response, &errResponse)
		if errResponse.Error.Code != "invalid_request" {
			t.Fatalf("%s error mismatch: %+v", path, errResponse)
		}
	}
}

func TestDashboardHandlerRunsAttentionSearchHasErrorsParams(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-running", "release", "1", "running", "build", `{}`, "2026-05-05T10:05:00Z")
	seedRun(t, database, types.RunKindChain, "chain-failed", "release", "1", "failed", "build", `{}`, "2026-05-05T10:04:00Z")
	seedRun(t, database, types.RunKindChain, "chain-gated", "release", "1", "gated", "build", `{}`, "2026-05-05T10:03:00Z")
	seedRun(t, database, types.RunKindTeam, "team-launch-running", "launch", "1", "running", "", `{}`, "2026-05-05T10:02:00Z")
	seedError(t, database, "err-failed", "chain-failed", types.RunKindChain, "release", "build", "fatal", "boom", "boom", "2026-05-05T10:04:30Z")

	handler := newDashboardHTTPHandler(t, database, sqliteadapter.NewCatalogStore(database))

	cases := []struct {
		name        string
		path        string
		wantStatus  int
		wantIDs     []string
		wantErrCode string
	}{
		{name: "attention=running", path: "/api/dashboard/runs?attention=running", wantStatus: http.StatusOK, wantIDs: []string{"chain-running", "team-launch-running"}},
		{name: "attention=failed", path: "/api/dashboard/runs?attention=failed", wantStatus: http.StatusOK, wantIDs: []string{"chain-failed"}},
		{name: "attention=gated", path: "/api/dashboard/runs?attention=gated", wantStatus: http.StatusOK, wantIDs: []string{"chain-gated"}},
		{name: "search by definition name", path: "/api/dashboard/runs?search=launch", wantStatus: http.StatusOK, wantIDs: []string{"team-launch-running"}},
		{name: "search by id substring", path: "/api/dashboard/runs?search=GATED", wantStatus: http.StatusOK, wantIDs: []string{"chain-gated"}},
		{name: "has_errors=true", path: "/api/dashboard/runs?has_errors=true", wantStatus: http.StatusOK, wantIDs: []string{"chain-failed"}},
		{name: "has_errors=false is no-op", path: "/api/dashboard/runs?has_errors=false&attention=running", wantStatus: http.StatusOK, wantIDs: []string{"chain-running", "team-launch-running"}},
		{name: "invalid attention", path: "/api/dashboard/runs?attention=bogus", wantStatus: http.StatusBadRequest, wantErrCode: "invalid_request"},
		{name: "invalid has_errors", path: "/api/dashboard/runs?has_errors=maybe", wantStatus: http.StatusBadRequest, wantErrCode: "invalid_request"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d, body=%s", response.Code, tc.wantStatus, response.Body.String())
			}
			if tc.wantErrCode != "" {
				var errResponse DashboardErrorResponse
				decodeResponse(t, response, &errResponse)
				if errResponse.Error.Code != tc.wantErrCode {
					t.Fatalf("error code = %q, want %q", errResponse.Error.Code, tc.wantErrCode)
				}
				return
			}
			var page RunListResponse
			decodeResponse(t, response, &page)
			gotIDs := map[string]bool{}
			for _, item := range page.Items {
				gotIDs[item.ID] = true
			}
			for _, want := range tc.wantIDs {
				if !gotIDs[want] {
					t.Fatalf("missing id %q in response: %+v", want, page.Items)
				}
			}
			if len(page.Items) != len(tc.wantIDs) {
				t.Fatalf("got %d items, want %d: %+v", len(page.Items), len(tc.wantIDs), page.Items)
			}
		})
	}
}

func TestDashboardHandlerDetailBudgetCatalogAndErrors(t *testing.T) {
	database := newDashboardTestDB(t)
	seedExecutionPlan(t, database, "plan-warning", types.BudgetPolicy{ID: "policy-warning", Scope: "chain", Tokens: &types.BudgetThreshold{Limit: 100, WarnAt: 50}, DefaultActionOnLimit: "pause"})
	seedRun(t, database, types.RunKindChain, "chain-detail", "release", "1", "running", "build", chainBudgetStateJSON(t, "chain-detail", "plan-warning", 75), "2026-05-05T10:00:00Z")
	seedEvent(t, database, "chain-detail", "step_started", `{"stepId":"build"}`, "2026-05-05T10:01:00Z")
	seedError(t, database, "err-detail", "chain-detail", types.RunKindChain, "release", "build", "transient", "retry", "retry requested", "2026-05-05T10:02:00Z")
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"owner": "ops"}, "body v1", true)

	handler := newDashboardHTTPHandler(t, database, store)

	detailResponse := httptest.NewRecorder()
	handler.ServeHTTP(detailResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/chain/chain-detail", nil))
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("detail status = %d body=%s", detailResponse.Code, detailResponse.Body.String())
	}
	var detail RunDetail
	decodeResponse(t, detailResponse, &detail)
	if detail.Summary.ID != "chain-detail" || len(detail.Events) != 1 || len(detail.Errors) != 1 || detail.Budget == nil {
		t.Fatalf("detail mismatch: %+v", detail)
	}

	budgetResponse := httptest.NewRecorder()
	handler.ServeHTTP(budgetResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/chain/chain-detail/budget", nil))
	if budgetResponse.Code != http.StatusOK {
		t.Fatalf("budget status = %d body=%s", budgetResponse.Code, budgetResponse.Body.String())
	}
	var budget BudgetView
	decodeResponse(t, budgetResponse, &budget)
	if budget.Evaluation == nil || budget.Evaluation.Overall != types.HealthWarning {
		t.Fatalf("budget mismatch: %+v", budget)
	}

	catalogListResponse := httptest.NewRecorder()
	handler.ServeHTTP(catalogListResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/catalog?kind=chain", nil))
	if catalogListResponse.Code != http.StatusOK {
		t.Fatalf("catalog list status = %d body=%s", catalogListResponse.Code, catalogListResponse.Body.String())
	}
	var catalogList CatalogListResponse
	decodeResponse(t, catalogListResponse, &catalogList)
	if len(catalogList.Items) != 1 || catalogList.Items[0].Name != "release" {
		t.Fatalf("catalog list mismatch: %+v", catalogList)
	}

	catalogDetailResponse := httptest.NewRecorder()
	handler.ServeHTTP(catalogDetailResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/catalog/chain/release?version=1", nil))
	if catalogDetailResponse.Code != http.StatusOK {
		t.Fatalf("catalog detail status = %d body=%s", catalogDetailResponse.Code, catalogDetailResponse.Body.String())
	}
	var catalogDetail CatalogDetail
	decodeResponse(t, catalogDetailResponse, &catalogDetail)
	if catalogDetail.Name != "release" || catalogDetail.Version != 1 || catalogDetail.Body != "body v1" {
		t.Fatalf("catalog detail mismatch: %+v", catalogDetail)
	}

	errorsResponse := httptest.NewRecorder()
	handler.ServeHTTP(errorsResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/errors?run_id=chain-detail&limit=500", nil))
	if errorsResponse.Code != http.StatusOK {
		t.Fatalf("errors status = %d body=%s", errorsResponse.Code, errorsResponse.Body.String())
	}
	var errorsList ErrorListResponse
	decodeResponse(t, errorsResponse, &errorsList)
	if len(errorsList.Items) != 1 || errorsList.Items[0].RunID != "chain-detail" {
		t.Fatalf("errors mismatch: %+v", errorsList)
	}

	missingResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/runs/chain/missing", nil))
	if missingResponse.Code != http.StatusNotFound {
		t.Fatalf("missing run status = %d body=%s", missingResponse.Code, missingResponse.Body.String())
	}

	badVersionResponse := httptest.NewRecorder()
	handler.ServeHTTP(badVersionResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/catalog/chain/release?version=bad", nil))
	if badVersionResponse.Code != http.StatusBadRequest {
		t.Fatalf("bad version status = %d body=%s", badVersionResponse.Code, badVersionResponse.Body.String())
	}
}

func TestDashboardHandlerCatalogDetailQueryRouteSupportsSlashNamesAndLegacyRoute(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "agent", "agents/platform/reviewer", map[string]any{"owner": "platform"}, "agent body", true)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"owner": "ops"}, "chain body", true)

	handler := newDashboardHTTPHandler(t, database, store)

	queryPath := "/api/dashboard/catalog/detail?kind=agent&name=" + url.QueryEscape("agents/platform/reviewer") + "&version=1"
	queryResponse := httptest.NewRecorder()
	handler.ServeHTTP(queryResponse, httptest.NewRequest(http.MethodGet, queryPath, nil))
	if queryResponse.Code != http.StatusOK {
		t.Fatalf("query detail status = %d body=%s", queryResponse.Code, queryResponse.Body.String())
	}
	var queryDetail CatalogDetail
	decodeResponse(t, queryResponse, &queryDetail)
	if queryDetail.Kind != "agent" || queryDetail.Name != "agents/platform/reviewer" || queryDetail.Version != 1 || queryDetail.Body != "agent body" {
		t.Fatalf("query detail mismatch: %+v", queryDetail)
	}

	legacyResponse := httptest.NewRecorder()
	handler.ServeHTTP(legacyResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/catalog/chain/release", nil))
	if legacyResponse.Code != http.StatusOK {
		t.Fatalf("legacy detail status = %d body=%s", legacyResponse.Code, legacyResponse.Body.String())
	}
	var legacyDetail CatalogDetail
	decodeResponse(t, legacyResponse, &legacyDetail)
	if legacyDetail.Kind != "chain" || legacyDetail.Name != "release" || legacyDetail.Body != "chain body" {
		t.Fatalf("legacy detail mismatch: %+v", legacyDetail)
	}
}

func TestDashboardHandlerCatalogDetailQueryRouteValidatesInputs(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"owner": "ops"}, "body", true)
	handler := newDashboardHTTPHandler(t, database, store)

	for _, path := range []string{
		"/api/dashboard/catalog/detail?name=release",
		"/api/dashboard/catalog/detail?kind=chain",
		"/api/dashboard/catalog/detail?kind=bogus&name=release",
		"/api/dashboard/catalog/detail?kind=chain&name=release&version=-1",
		"/api/dashboard/catalog/detail?kind=chain&name=release&version=bad",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d body=%s", path, response.Code, response.Body.String())
		}
		var errResponse DashboardErrorResponse
		decodeResponse(t, response, &errResponse)
		if errResponse.Error.Code != "invalid_request" || errResponse.Error.Message == "" {
			t.Fatalf("%s error mismatch: %+v", path, errResponse)
		}
	}
}

func TestDashboardHandlerCatalogDetailQueryRouteReportsNoActiveVersion(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "agent", "draft-agent", map[string]any{"description": "Draft"}, "agent body", false)
	handler := newDashboardHTTPHandler(t, database, store)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dashboard/catalog/detail?kind=agent&name=draft-agent", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("no-active detail status = %d body=%s", response.Code, response.Body.String())
	}
	var errResponse DashboardErrorResponse
	decodeResponse(t, response, &errResponse)
	if errResponse.Error.Code != "not_found" || !strings.Contains(errResponse.Error.Message, "agent/draft-agent") {
		t.Fatalf("no-active detail error mismatch: %+v", errResponse)
	}
}

func newDashboardHTTPHandler(t *testing.T, database *db.DB, store ports.CatalogStore) http.Handler {
	t.Helper()
	handler, _ := newDashboardHTTPHandlerWithBus(t, database, store)
	return handler
}

func newDashboardHTTPHandlerWithBus(t *testing.T, database *db.DB, store ports.CatalogStore) (http.Handler, *events.Bus) {
	t.Helper()
	bus := events.NewBus(sqliteadapter.NewRunEventStore(database))
	return NewHandler(HandlerConfig{
		ReadModel: NewReadModel(sqliteadapter.NewRunReadStore(database), sqliteadapter.NewActivityStore(database), sqliteadapter.NewHandoffStore(database), sqliteadapter.NewRunEventStore(database), sqliteadapter.NewExecutionPlanStore(database), sqliteadapter.NewErrorJournalStore(database)),
		Catalog:   NewCatalogAdapter(store),
		Events:    bus,
		Health: func(context.Context) HealthView {
			return HealthView{Status: "ok", Name: "lazyai-orchestrator", Port: 4321, PID: 99}
		},
	}), bus
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response %q: %v", response.Body.String(), err)
	}
}

func assertJSONContentType(t *testing.T, contentType string) {
	t.Helper()
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
}
