// Package geminiext generates a Gemini CLI extension directory from
// ai-setup's embedded library. The output conforms to the
// gemini-extension.json schema documented at
// https://github.com/google-gemini/gemini-cli/blob/main/docs/extensions/reference.md
// and is installable via `gemini extensions link <path>`.
package geminiext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/library"
)

// Extension metadata. Hardcoded per spec 017 Q6.
const (
	ExtensionName        = "ai-setup-gemini"
	ExtensionDescription = "ai-setup GEMINI.md template and custom commands for Gemini CLI"
	ContextFileName      = "GEMINI.md"
)

// geminiMdSource is the library-relative path to the raw GEMINI.md template
// shipped in the extension (locked Q4: raw, recipient fills placeholders).
const geminiMdSource = "root/GEMINI.template.md"

// envPattern matches ${VAR} placeholders in env values / headers so the
// generator can skip entries that would fail to load in a fresh extension.
var envPattern = regexp.MustCompile(`\$\{(\w+)\}`)

// McpServer is the subset of the canonical MCP catalog entry the extension
// cares about. The geminiext package accepts a pre-parsed map so callers
// can share one catalog read across multiple generators.
type McpServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// BuildResult reports what the generator emitted.
type BuildResult struct {
	OutDir            string
	FileCount         int
	SkippedMcpServers []string // server names skipped because they carry ${VAR} placeholders
}

// Build generates a Gemini extension directory under outDir from the given
// library FS. version is written into gemini-extension.json. mcpCatalog is
// optional; pass nil to omit the mcpServers key entirely.
func Build(libFS fs.FS, mcpCatalog map[string]McpServer, outDir, version string) (BuildResult, error) {
	if libFS == nil {
		return BuildResult{}, fmt.Errorf("geminiext.Build: libFS is nil")
	}
	if outDir == "" {
		return BuildResult{}, fmt.Errorf("geminiext.Build: outDir is empty")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return BuildResult{}, fmt.Errorf("create outDir: %w", err)
	}

	total := 0

	// Manifest is written last so we know the final mcpServers map after filtering.
	mcpServers, skipped := buildMcpServers(mcpCatalog)

	// GEMINI.md (raw template, no placeholder fill).
	if err := copyGeminiMd(libFS, outDir); err != nil {
		return BuildResult{}, fmt.Errorf("copy GEMINI.md: %w", err)
	}
	total++

	// Commands directory.
	nCommands, err := copyCommands(libFS, outDir)
	if err != nil {
		return BuildResult{}, fmt.Errorf("copy commands: %w", err)
	}
	total += nCommands

	// Manifest.
	if err := writeManifest(outDir, version, mcpServers); err != nil {
		return BuildResult{}, err
	}
	total++

	return BuildResult{OutDir: outDir, FileCount: total, SkippedMcpServers: skipped}, nil
}

// writeManifest writes gemini-extension.json. mcpServers may be nil.
func writeManifest(outDir, version string, mcpServers map[string]any) error {
	manifest := map[string]any{
		"name":            ExtensionName,
		"version":         version,
		"description":     ExtensionDescription,
		"contextFileName": ContextFileName,
	}
	if len(mcpServers) > 0 {
		manifest["mcpServers"] = mcpServers
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(outDir, "gemini-extension.json"), data, 0o644)
}

// copyGeminiMd copies the raw GEMINI template to <outDir>/GEMINI.md.
func copyGeminiMd(libFS fs.FS, outDir string) error {
	data, err := fs.ReadFile(libFS, geminiMdSource)
	if err != nil {
		return fmt.Errorf("read %s: %w", geminiMdSource, err)
	}
	return os.WriteFile(filepath.Join(outDir, "GEMINI.md"), data, 0o644)
}

// copyCommands walks the library's Gemini commands subdir (resolved via the
// library helper, which falls back to the legacy top-level `commands/` when
// needed) and copies each *.toml into <outDir>/commands/, preserving
// subdirectory structure so namespaced commands (e.g. `gcs/sync.toml`) stay
// namespaced in the output.
func copyCommands(libFS fs.FS, outDir string) (int, error) {
	subdir := library.ResolveGeminiCommandsSubdir(libFS)
	if subdir == "" {
		return 0, nil // no commands to ship; not an error
	}
	destDir := filepath.Join(outDir, "commands")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}

	count := 0
	err := fs.WalkDir(libFS, subdir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".toml") {
			return nil
		}
		rel, err := filepath.Rel(subdir, path)
		if err != nil {
			return fmt.Errorf("rel %s: %w", path, err)
		}
		dst := filepath.Join(destDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		data, err := fs.ReadFile(libFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

// buildMcpServers filters the canonical catalog to static entries only:
// servers whose env / headers contain ${VAR} placeholders are skipped
// because Gemini extensions load without a prompt layer for those values.
// Returns the filtered map (suitable for gemini-extension.json's mcpServers
// field) and the list of skipped server names.
func buildMcpServers(catalog map[string]McpServer) (map[string]any, []string) {
	if len(catalog) == 0 {
		return nil, nil
	}
	out := make(map[string]any)
	var skipped []string
	for name, srv := range catalog {
		if hasPlaceholders(srv) {
			skipped = append(skipped, name)
			continue
		}
		entry := map[string]any{}
		if srv.Command != "" {
			entry["command"] = srv.Command
		}
		if len(srv.Args) > 0 {
			entry["args"] = srv.Args
		}
		if srv.Cwd != "" {
			entry["cwd"] = srv.Cwd
		}
		if len(srv.Env) > 0 {
			entry["env"] = srv.Env
		}
		if srv.URL != "" {
			entry["url"] = srv.URL
		}
		if len(srv.Headers) > 0 {
			entry["headers"] = srv.Headers
		}
		out[name] = entry
	}
	if len(out) == 0 {
		return nil, skipped
	}
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "[geminiext] skipped %d MCP server(s) with ${VAR} placeholders: %s\n",
			len(skipped), strings.Join(skipped, ", "))
	}
	return out, skipped
}

// hasPlaceholders returns true if any env value or header value in srv
// contains a ${VAR} placeholder.
func hasPlaceholders(srv McpServer) bool {
	for _, v := range srv.Env {
		if envPattern.MatchString(v) {
			return true
		}
	}
	for _, v := range srv.Headers {
		if envPattern.MatchString(v) {
			return true
		}
	}
	return false
}
