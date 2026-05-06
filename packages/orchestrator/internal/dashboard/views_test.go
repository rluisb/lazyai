package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardViewHandlerServesShellWithSemanticHooks(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	assertContentTypePrefix(t, response.Header().Get("Content-Type"), "text/html")
	body := response.Body.String()
	for _, want := range []string{
		`<main id="dashboard-app"`,
		`data-api-prefix="/api/dashboard"`,
		`href="/dashboard/assets/dashboard.css"`,
		`src="/dashboard/assets/dashboard.js"`,
		`id="status-message"`,
		`id="overview-panel"`,
		`id="health-status"`,
		`id="active-runs"`,
		`id="run-counts"`,
		`id="recent-runs"`,
		`id="recent-errors"`,
		`id="catalog-counts"`,
		`id="run-browser"`,
		`id="run-kind-filter"`,
		`id="run-state-filter"`,
		`id="run-limit"`,
		`id="run-list"`,
		`id="run-list-empty"`,
		`id="run-detail-panel"`,
		`id="run-detail-empty"`,
		`id="run-summary"`,
		`id="run-state-json"`,
		`id="budget-panel"`,
		`id="budget-state"`,
		`id="budget-evaluation"`,
		`id="event-timeline"`,
		`id="run-errors"`,
		`id="catalog-browser"`,
		`id="catalog-kind-filter"`,
		`id="catalog-list"`,
		`id="catalog-detail"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard shell missing %q in:\n%s", want, body)
		}
	}
	for _, forbidden := range []string{"admin/shutdown", "data-action=\"delete", "data-action=\"retry", "data-action=\"start"} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Fatalf("dashboard shell contains write/admin affordance %q", forbidden)
		}
	}
}

func TestDashboardViewHandlerServesEmbeddedAssets(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	tests := []struct {
		path        string
		contentType string
		contains    []string
	}{
		{
			path:        "/dashboard/assets/dashboard.css",
			contentType: "text/css",
			contains:    []string{".dashboard-shell", ".status-message", ".run-list", ".event-timeline", ".catalog-detail"},
		},
		{
			path:        "/dashboard/assets/dashboard.js",
			contentType: "text/javascript",
			contains: []string{
				"fetchOverview", "fetchRuns", "openRunDetail", "connectRunEvents", "EventSource",
				"fetchCatalog", "renderCatalogDetail", "renderEmptyState", "renderErrorState",
				"/overview", "/runs", "/catalog", "/errors",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("asset status = %d body=%s", response.Code, response.Body.String())
			}
			assertContentTypePrefix(t, response.Header().Get("Content-Type"), tt.contentType)
			body := response.Body.String()
			if strings.TrimSpace(body) == "" {
				t.Fatalf("asset body is empty")
			}
			for _, want := range tt.contains {
				if !strings.Contains(body, want) {
					t.Fatalf("asset %s missing %q", tt.path, want)
				}
			}
		})
	}
}

func TestDashboardViewRoutesDoNotShadowDashboardAPI(t *testing.T) {
	mux := http.NewServeMux()
	RegisterViewRoutes(mux, ViewConfig{})
	mux.HandleFunc("/api/dashboard/overview", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("api handler"))
	})

	apiResponse := httptest.NewRecorder()
	mux.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/overview", nil))
	if apiResponse.Code != http.StatusAccepted || apiResponse.Body.String() != "api handler" {
		t.Fatalf("dashboard views shadowed API route: status=%d body=%q", apiResponse.Code, apiResponse.Body.String())
	}

	shellResponse := httptest.NewRecorder()
	mux.ServeHTTP(shellResponse, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if shellResponse.Code != http.StatusOK {
		t.Fatalf("dashboard shell status through mux = %d body=%s", shellResponse.Code, shellResponse.Body.String())
	}
}

func TestDashboardViewHandlerRejectsUnknownAndNonReadOnlyRoutes(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	for _, path := range []string{"/api/dashboard/overview", "/mcp", "/dashboard/api/dashboard/overview", "/dashboard/assets/missing.js"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("GET %s status = %d body=%s", path, response.Code, response.Body.String())
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/dashboard/", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /dashboard/ status = %d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("POST /dashboard/ Allow = %q", response.Header().Get("Allow"))
	}
}

func assertContentTypePrefix(t *testing.T, got string, wantPrefix string) {
	t.Helper()
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("content type = %q, want prefix %q", got, wantPrefix)
	}
}
