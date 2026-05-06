package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	orchmcp "github.com/rluisb/lazyai/packages/orchestrator/internal/mcp"
)

func TestServeRoutesMountDashboardAndPreserveDaemonRoutes(t *testing.T) {
	database := newRouteTestDB(t)
	projectRoot := t.TempDir()
	runtimeConfig := orchmcp.DefaultRuntimeConfig()
	orchestrator := orchmcp.NewOrchestrator(database, orchmcp.NewScopeContext("project", projectRoot, ""), orchmcp.WithRuntimeConfig(runtimeConfig))
	tracker := newClientTracker(time.Minute)
	idle := newIdleManager(idleManagerOptions{
		Timeout: time.Minute,
		Tracker: tracker,
		ActiveRuns: func(context.Context) (db.ActiveRunCounts, error) {
			return database.ActiveRunCounts()
		},
		Shutdown: func(string) {},
	})

	mux := http.NewServeMux()
	shutdownReason := ""
	registerServeRoutes(mux, serveRouteConfig{
		Port:          43210,
		ProjectRoot:   projectRoot,
		Scope:         "project",
		RuntimeConfig: runtimeConfig,
		StartedAt:     "2026-05-05T10:00:00Z",
		Database:      database,
		Orchestrator:  orchestrator,
		Tracker:       tracker,
		Idle:          idle,
		MCPHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Route", "mcp")
			w.WriteHeader(http.StatusNoContent)
		}),
		Shutdown: func(reason string) {
			shutdownReason = reason
		},
	})

	dashboardResponse := httptest.NewRecorder()
	mux.ServeHTTP(dashboardResponse, httptest.NewRequest(http.MethodGet, "/dashboard/", nil))
	if dashboardResponse.Code != http.StatusOK {
		t.Fatalf("GET /dashboard/ status = %d body=%s", dashboardResponse.Code, dashboardResponse.Body.String())
	}
	if !strings.HasPrefix(dashboardResponse.Header().Get("Content-Type"), "text/html") || !strings.Contains(dashboardResponse.Body.String(), `id="dashboard-app"`) {
		t.Fatalf("GET /dashboard/ did not serve dashboard shell: content-type=%q body=%s", dashboardResponse.Header().Get("Content-Type"), dashboardResponse.Body.String())
	}

	apiResponse := httptest.NewRecorder()
	mux.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/overview", nil))
	if apiResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/dashboard/overview status = %d body=%s", apiResponse.Code, apiResponse.Body.String())
	}
	if !strings.HasPrefix(apiResponse.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("GET /api/dashboard/overview content-type = %q", apiResponse.Header().Get("Content-Type"))
	}
	var overview struct {
		Health struct {
			Status      string `json:"status"`
			Name        string `json:"name"`
			Port        int    `json:"port"`
			ProjectRoot string `json:"projectRoot"`
		} `json:"health"`
	}
	if err := json.Unmarshal(apiResponse.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview: %v body=%s", err, apiResponse.Body.String())
	}
	if overview.Health.Status != "ok" || overview.Health.Name != "lazyai-orchestrator" || overview.Health.Port != 43210 || overview.Health.ProjectRoot != projectRoot {
		t.Fatalf("overview health mismatch: %+v", overview.Health)
	}

	eventsResponse := httptest.NewRecorder()
	mux.ServeHTTP(eventsResponse, httptest.NewRequest(http.MethodGet, "/api/dashboard/events", nil))
	if eventsResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/dashboard/events status = %d body=%s", eventsResponse.Code, eventsResponse.Body.String())
	}
	if !strings.HasPrefix(eventsResponse.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("GET /api/dashboard/events content-type = %q", eventsResponse.Header().Get("Content-Type"))
	}

	mcpResponse := httptest.NewRecorder()
	mux.ServeHTTP(mcpResponse, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if mcpResponse.Code != http.StatusNoContent || mcpResponse.Header().Get("X-Test-Route") != "mcp" {
		t.Fatalf("/mcp route was shadowed: status=%d headers=%v body=%s", mcpResponse.Code, mcpResponse.Header(), mcpResponse.Body.String())
	}

	healthResponse := httptest.NewRecorder()
	mux.ServeHTTP(healthResponse, httptest.NewRequest(http.MethodGet, "/health", nil))
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("GET /health status = %d body=%s", healthResponse.Code, healthResponse.Body.String())
	}
	var health daemonHealth
	if err := json.Unmarshal(healthResponse.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v body=%s", err, healthResponse.Body.String())
	}
	if health.Status != "ok" || health.Name != "lazyai-orchestrator" || health.Port != 43210 || health.ProjectRoot != projectRoot {
		t.Fatalf("health mismatch: %+v", health)
	}

	shutdownGetResponse := httptest.NewRecorder()
	mux.ServeHTTP(shutdownGetResponse, httptest.NewRequest(http.MethodGet, "/admin/shutdown", nil))
	if shutdownGetResponse.Code != http.StatusMethodNotAllowed || shutdownReason != "" {
		t.Fatalf("GET /admin/shutdown mismatch: status=%d reason=%q body=%s", shutdownGetResponse.Code, shutdownReason, shutdownGetResponse.Body.String())
	}

	shutdownPostResponse := httptest.NewRecorder()
	mux.ServeHTTP(shutdownPostResponse, httptest.NewRequest(http.MethodPost, "/admin/shutdown", nil))
	if shutdownPostResponse.Code != http.StatusOK || shutdownReason != "admin request" {
		t.Fatalf("POST /admin/shutdown mismatch: status=%d reason=%q body=%s", shutdownPostResponse.Code, shutdownReason, shutdownPostResponse.Body.String())
	}
}

func TestServeDashboardAPIRoutesRemainReadOnly(t *testing.T) {
	database := newRouteTestDB(t)
	orchestrator := orchmcp.NewOrchestrator(database, orchmcp.NewScopeContext("project", t.TempDir(), ""))
	tracker := newClientTracker(time.Minute)
	idle := newIdleManager(idleManagerOptions{
		Timeout: time.Minute,
		Tracker: tracker,
		ActiveRuns: func(context.Context) (db.ActiveRunCounts, error) {
			return database.ActiveRunCounts()
		},
		Shutdown: func(string) {},
	})

	mux := http.NewServeMux()
	registerServeRoutes(mux, serveRouteConfig{
		Port:          43210,
		ProjectRoot:   t.TempDir(),
		Scope:         "project",
		RuntimeConfig: orchmcp.DefaultRuntimeConfig(),
		StartedAt:     "2026-05-05T10:00:00Z",
		Database:      database,
		Orchestrator:  orchestrator,
		Tracker:       tracker,
		Idle:          idle,
		MCPHandler:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }),
		Shutdown:      func(string) {},
	})

	apiResponse := httptest.NewRecorder()
	mux.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodPost, "/api/dashboard/overview", nil))
	if apiResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/dashboard/overview status = %d body=%s", apiResponse.Code, apiResponse.Body.String())
	}
	if apiResponse.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("POST /api/dashboard/overview Allow = %q", apiResponse.Header().Get("Allow"))
	}

	viewResponse := httptest.NewRecorder()
	mux.ServeHTTP(viewResponse, httptest.NewRequest(http.MethodPost, "/dashboard/", nil))
	if viewResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /dashboard/ status = %d body=%s", viewResponse.Code, viewResponse.Body.String())
	}
}

func newRouteTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatal(err)
	}
	return database
}
