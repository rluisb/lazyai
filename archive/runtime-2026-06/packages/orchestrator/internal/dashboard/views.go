package dashboard

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

const (
	defaultDashboardTitle = "LazyAI Orchestrator Dashboard"
	defaultDashboardBase  = "/dashboard/"
	dashboardAssetPrefix  = "/dashboard/assets/"
)

//go:embed templates/dashboard.html assets/dashboard.css assets/dashboard.js
var dashboardFiles embed.FS

var dashboardTemplate = template.Must(template.ParseFS(dashboardFiles, "templates/dashboard.html"))

// ViewConfig controls embedded dashboard shell rendering.
type ViewConfig struct {
	Title     string
	BasePath  string
	APIPrefix string
}

type viewHandler struct {
	title     string
	basePath  string
	apiPrefix string
	assets    http.Handler
}

type dashboardViewData struct {
	Title     string
	BasePath  string
	AssetPath string
	APIPrefix string
}

// NewViewHandler serves the embedded dashboard shell and static assets.
func NewViewHandler(config ViewConfig) http.Handler {
	basePath := config.BasePath
	if basePath == "" {
		basePath = defaultDashboardBase
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	title := config.Title
	if title == "" {
		title = defaultDashboardTitle
	}
	apiPrefix := config.APIPrefix
	if apiPrefix == "" {
		apiPrefix = dashboardAPIPrefix
	}

	assetFS, err := fs.Sub(dashboardFiles, "assets")
	if err != nil {
		panic(err)
	}

	return &viewHandler{
		title:     title,
		basePath:  basePath,
		apiPrefix: apiPrefix,
		assets:    http.StripPrefix(basePath+"assets/", http.FileServer(http.FS(assetFS))),
	}
}

// RegisterViewRoutes mounts only the dashboard shell/assets routes on mux.
func RegisterViewRoutes(mux *http.ServeMux, config ViewConfig) {
	basePath := config.BasePath
	if basePath == "" {
		basePath = defaultDashboardBase
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	handler := NewViewHandler(ViewConfig{Title: config.Title, BasePath: basePath, APIPrefix: config.APIPrefix})
	mux.Handle(basePath, handler)
	mux.Handle(strings.TrimSuffix(basePath, "/"), handler)
}

func (h *viewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "dashboard views are read-only and only support GET", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == strings.TrimSuffix(h.basePath, "/") {
		http.Redirect(w, r, h.basePath, http.StatusMovedPermanently)
		return
	}
	if r.URL.Path == h.basePath {
		h.renderShell(w, r)
		return
	}
	assetPrefix := h.basePath + "assets/"
	if strings.HasPrefix(r.URL.Path, assetPrefix) && r.URL.Path != assetPrefix {
		h.assets.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *viewHandler) renderShell(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_ = dashboardTemplate.ExecuteTemplate(w, "dashboard.html", dashboardViewData{
		Title:     h.title,
		BasePath:  h.basePath,
		AssetPath: h.basePath + "assets/",
		APIPrefix: h.apiPrefix,
	})
}
