package scout

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DetectEntryPoints detects likely application, server, framework, and command entry points.
func DetectEntryPoints(targetDir string) []EntryPoint {
	var entries []EntryPoint
	seen := map[string]struct{}{}
	add := func(path, typ string) {
		if len(entries) >= 10 || path == "" || typ == "" {
			return
		}
		path = filepath.ToSlash(path)
		key := path + "\x00" + typ
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		entries = append(entries, EntryPoint{Path: path, Type: typ})
	}

	if mainPath := packageJSONMain(targetDir); mainPath != "" {
		add(mainPath, "package_main")
	}

	excluded := excludedDirSet(DefaultExcludedDirs)
	var found []EntryPoint
	_ = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != targetDir && isExcludedDir(d.Name(), excluded) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(targetDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		name := d.Name()
		entryType := ""
		switch {
		case rel == "src/app/layout.tsx":
			entryType = "nextjs_layout"
		case rel == "src/app/page.tsx":
			entryType = "nextjs_page"
		case rel == "pages/index.tsx":
			entryType = "nextjs_pages"
		case rel == "pages/_app.tsx":
			entryType = "nextjs_app"
		case rel == "manage.py":
			entryType = "django_manage"
		case name == "wsgi.py":
			entryType = "wsgi_entry"
		case strings.HasPrefix(rel, "cmd/") && strings.HasSuffix(rel, "/main.go"):
			entryType = "go_cmd"
		case name == "server.ts" || name == "server.js":
			entryType = "server_entry"
		case name == "main.go" || name == "main.ts" || name == "main.js" || name == "index.ts" || name == "index.js" || name == "app.ts" || name == "bootstrap.ts":
			entryType = "app_entry"
		}
		if entryType != "" {
			found = append(found, EntryPoint{Path: rel, Type: entryType})
		}
		return nil
	})
	sort.Slice(found, func(i, j int) bool {
		if found[i].Type == found[j].Type {
			return found[i].Path < found[j].Path
		}
		return found[i].Type < found[j].Type
	})
	for _, entry := range found {
		add(entry.Path, entry.Type)
	}
	if len(entries) == 0 {
		entries = append(entries, EntryPoint{Path: ".", Type: "project_root"})
	}
	return entries
}

func packageJSONMain(targetDir string) string {
	data, err := os.ReadFile(filepath.Join(targetDir, "package.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Main string `json:"main"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ""
	}
	return strings.TrimSpace(manifest.Main)
}
