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

func TestDashboardViewCatalogControlsExposeAllKindsSearchAndSort(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		`<option value="agent">Agents</option>`,
		`<option value="domain">Domains</option>`,
		`<option value="mode">Modes</option>`,
		`<option value="chain">Chains</option>`,
		`<option value="team">Teams</option>`,
		`<option value="workflow">Workflows</option>`,
		`id="catalog-search"`,
		`name="search"`,
		`id="catalog-sort"`,
		`name="sort"`,
		`value="name"`,
		`value="kind"`,
		`value="updated"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard catalog controls missing %q in:\n%s", want, body)
		}
	}
}

func TestDashboardViewCatalogAssetContracts(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.js", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard js status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		`fetchJSON("/catalog/detail"`,
		"catalog-search",
		"catalog-sort",
		"applyCatalogFilters",
		"groupCatalogItems",
		"catalog-kind-group",
		"No active version",
		"Definition has no active version",
		"button.disabled = true",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard js missing catalog contract %q", want)
		}
	}
	for _, forbidden := range []string{"catalog/create", "catalog/delete", "catalog/deactivate", "data-action=\"delete"} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Fatalf("dashboard js contains write/admin affordance %q", forbidden)
		}
	}
}

func TestDashboardViewCSSLayoutAssetContracts(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.css", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard css status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		"@media (max-width: 800px)",
		"@media (max-width: 480px)",
		"align-items: start;",
		"h4,\nh5",
		"list-style: none;",
		"min-width: 12rem;",
		"overflow-wrap: anywhere;",
		"minmax(0, 1fr)",
		"max-width: 100%;",
		"button:disabled",
		".catalog-kind-items",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard css missing layout contract %q", want)
		}
	}
}

func TestDashboardViewLayoutStatusAccessibilityHooks(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		`id="status-message" class="status-message" role="status" aria-live="polite"`,
		`<form id="catalog-filters" class="toolbar" aria-label="Catalog filters">`,
		`<ul id="catalog-list" class="catalog-list" data-empty="No catalog definitions found.">`,
		`<article id="catalog-detail" class="catalog-detail empty-state"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard shell missing layout/accessibility hook %q", want)
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

func TestDashboardViewProvidesNavShellAndPlannedBlock(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()

	for _, want := range []string{
		// body data attrs drive theme/density/nav switching
		`data-theme=`,
		`data-density=`,
		`data-nav=`,
		// JetBrains Mono Nerd Font hookup
		"JetBrainsMono Nerd Font",
		// Nav landmark + primary items
		`aria-label="Dashboard sections"`,
		`data-route="#/overview"`,
		`data-route="#/runs"`,
		`data-route="#/catalog"`,
		`data-route="#/errors"`,
		// Run detail nav is conditional ("select first" hint)
		`data-nav-run-detail`,
		`Select a run first`,
		// Errors first-class screen
		`id="errors-panel"`,
		`data-dashboard-section="errors"`,
		`id="errors-list"`,
		// Other sections expose data-dashboard-section hooks for hash routing
		`data-dashboard-section="overview"`,
		`data-dashboard-section="runs"`,
		`data-dashboard-section="run-detail"`,
		`data-dashboard-section="catalog"`,
		// Tweaks toggle exists
		`id="tweaks-toggle"`,
		`id="tweaks-panel"`,
		`name="theme"`,
		`name="density"`,
		`name="nav"`,
		// Planned block — exact inline copy from nav-gating-plan.md
		`Planned after global dashboard event stream support.`,
		`Requires #174 structured logging dashboard contract.`,
		`Not scoped; read-only daemon config would require separate approval.`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard shell missing nav/theme hook %q in body", want)
		}
	}

	// Planned items must NOT be clickable: no <a href> or <button> for live/logs/settings routes.
	for _, forbidden := range []string{
		`href="#/live"`,
		`href="#/logs"`,
		`href="#/settings"`,
		`data-route="#/live"`,
		`data-route="#/logs"`,
		`data-route="#/settings"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("dashboard shell exposes clickable future route %q", forbidden)
		}
	}
}

func TestDashboardJSHashRoutingAndErrorsScreen(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.js", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard js status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		"parseHash",
		"navigateTo",
		"activateSection",
		"updateNavState",
		"hashchange",
		`"#/overview"`,
		`"#/runs"`,
		`"#/catalog"`,
		`"#/errors"`,
		// Run-open path constructs run-detail hash route
		`#/runs/${`,
		// Errors screen is hooked up via existing /errors API
		`fetchJSON("/errors"`,
		"errors-list",
		// Tweaks state writes to body data attrs (no DB)
		"document.body.dataset.theme",
		"document.body.dataset.density",
		"document.body.dataset.nav",
		"localStorage",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard js missing route/tweak/errors contract %q", want)
		}
	}
	// Logs/Settings remain forbidden (no backend); Live activity primary nav stays forbidden
	// even though the underlying /api/dashboard/events stream now exists — the IA decision
	// is to surface activity on Overview, not as a primary nav destination.
	for _, forbidden := range []string{
		`"#/live"`,
		`"#/logs"`,
		`"#/settings"`,
		`"/logs"`,
		`"/settings"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("dashboard js references forbidden path %q", forbidden)
		}
	}
}

func TestDashboardViewLiveActivityFeedOnOverview(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		`id="live-activity"`,
		`id="activity-feed"`,
		`id="activity-pause-toggle"`,
		`id="activity-status"`,
		`data-activity-empty`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard shell missing live activity hook %q", want)
		}
	}
}

func TestDashboardJSConsumesGlobalEventStream(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.js", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard js status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		"connectGlobalEvents",
		"disconnectGlobalEvents",
		"appendActivityEvent",
		"handleGlobalEvent",
		"showToastForEvent",
		"scheduleOverviewRefresh",
		`"/events"`,
		"activity-feed",
		"activity-pause-toggle",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard js missing global-events contract %q", want)
		}
	}
}

func TestDashboardCSSActivityFeedVisuals(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.css", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard css status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		".activity-feed",
		".activity-item",
		".activity-item-glyph",
		".activity-item.entering",
		".toasts",
		".toast",
		"@keyframes",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard css missing activity feed visual %q", want)
		}
	}
}

func TestDashboardViewRunDetailHasHeroTimelineAndBudgetCards(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard shell status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		`id="run-detail-hero"`,
		`id="run-detail-state-chip"`,
		`id="run-detail-copy-id"`,
		`id="run-timeline"`,
		`id="run-budget-cards"`,
		`<details `, // collapsibles for raw JSON / execution plan / handoffs
		`id="run-detail-raw-state"`,
		`id="run-detail-execution-plan"`,
		`id="run-detail-handoffs"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard run detail shell missing %q", want)
		}
	}
}

func TestDashboardJSRunDetailRenderersAndCopy(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.js", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard js status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		"renderRunHero",
		"renderTimeline",
		"renderBudgetCards",
		"renderRunStateChip",
		"navigator.clipboard",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard js missing run-detail renderer %q", want)
		}
	}
}

func TestDashboardCSSRunDetailVisuals(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.css", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard css status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		".run-hero",
		".run-timeline",
		".timeline-node",
		".timeline-marker",
		".budget-card",
		".budget-bar",
		".budget-bar-fill",
		"summary",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard css missing run-detail visual %q", want)
		}
	}
}

func TestDashboardCSSCatppuccinAndNavLayout(t *testing.T) {
	handler := NewViewHandler(ViewConfig{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dashboard/assets/dashboard.css", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard css status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, want := range []string{
		// Theme tokens (Catppuccin Latte default + Macchiato dark)
		"--ctp-base",
		"--ctp-mantle",
		"--ctp-text",
		"--ctp-peach",
		"--accent",
		// Theme/density/nav switching selectors
		`body[data-theme="dark"]`,
		`body[data-density="compact"]`,
		`body[data-density="comfortable"]`,
		`body[data-nav="sidebar"]`,
		`body[data-nav="top"]`,
		// JetBrains Mono Nerd Font font-family
		"JetBrainsMono Nerd Font",
		// Nav primitives
		".nav-item",
		`[aria-current="page"]`,
		".planned-item",
		".nav-hint",
		// Section visibility for hash routing
		`[data-dashboard-section]`,
		`[hidden]`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard css missing nav/theme contract %q", want)
		}
	}
}

func assertContentTypePrefix(t *testing.T, got string, wantPrefix string) {
	t.Helper()
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("content type = %q, want prefix %q", got, wantPrefix)
	}
}
