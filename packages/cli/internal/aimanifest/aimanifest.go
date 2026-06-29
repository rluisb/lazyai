// Package aimanifest loads and validates the canonical LazyAI V2 project
// manifest at .ai/lazyai.json. The manifest is the source of truth that the
// compile pipeline resolves before emitting tool-native outputs.
//
// This is distinct from package manifest, which reads/writes the legacy
// .ai-setup.json store. Manifest target tokens are user-facing short names
// (e.g. "claude"); they map to internal types.ToolId values (e.g.
// "claude-code") via ResolveTargets.
package aimanifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// FileName is the manifest filename under the canonical .ai/ directory.
const FileName = "lazyai.json"

// SchemaVersion is the current manifest schema version.
const SchemaVersion = "1.0"

// ErrNotFound is returned (wrapped) by Load when no manifest exists.
var ErrNotFound = errors.New("manifest not found")

// Source describes which canonical asset families and packs feed the compile.
type Source struct {
	Assets []string `json:"assets,omitempty"`
	Packs  []string `json:"packs,omitempty"`
}

// Library names embedded/library asset packs referenced by the manifest.
type Library struct {
	Packs []string `json:"packs,omitempty"`
}

// Safety captures the manifest safety profile flags.
type Safety struct {
	RequireDiffBeforeWrite bool   `json:"requireDiffBeforeWrite,omitempty"`
	AllowGlobalWrites      bool   `json:"allowGlobalWrites,omitempty"`
	DenyInlineSecrets      bool   `json:"denyInlineSecrets,omitempty"`
	WarnIfNoSandbox        bool   `json:"warnIfNoSandbox,omitempty"`
	GeneratedFileMode      string `json:"generatedFileMode,omitempty"`
}

// Manifest is the parsed .ai/lazyai.json document. Adapter option blocks are
// kept as open maps so unknown per-adapter keys round-trip untouched.
type Manifest struct {
	Schema   string                    `json:"$schema,omitempty"`
	Version  string                    `json:"version"`
	Profile  string                    `json:"profile,omitempty"`
	Targets  []string                  `json:"targets"`
	Source   *Source                   `json:"source,omitempty"`
	Library  *Library                  `json:"library,omitempty"`
	Adapters map[string]map[string]any `json:"adapters,omitempty"`
	Safety   *Safety                   `json:"safety,omitempty"`
}

// targetAliases maps manifest target tokens to internal tool IDs.
var targetAliases = map[string]types.ToolId{
	"opencode":    types.ToolIdOpenCode,
	"claude":      types.ToolIdClaudeCode,
	"claude-code": types.ToolIdClaudeCode,
	"copilot":     types.ToolIdCopilot,
	"pi":          types.ToolIdPi,
	"omp":         types.ToolIdOmp,
	"antigravity": types.ToolIdAntigravity,
	"kiro":        types.ToolIdKiro,
	"codex":       types.ToolIdCodex,
}

// canonicalToken is the preferred manifest token for each tool ID (the inverse
// of targetAliases, choosing the short user-facing name).
var canonicalToken = map[types.ToolId]string{
	types.ToolIdOpenCode:    "opencode",
	types.ToolIdClaudeCode:  "claude",
	types.ToolIdCopilot:     "copilot",
	types.ToolIdPi:          "pi",
	types.ToolIdOmp:         "omp",
	types.ToolIdAntigravity: "antigravity",
	types.ToolIdKiro:        "kiro",
	types.ToolIdCodex:       "codex",
}

// Path returns the manifest path for the given canonical .ai/ directory.
func Path(aiDir string) string { return filepath.Join(aiDir, FileName) }

// Load reads and parses the manifest at <aiDir>/lazyai.json. When the file is
// absent the returned error wraps ErrNotFound (test with errors.Is).
func Load(aiDir string) (*Manifest, error) {
	p := Path(aiDir)
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w at %s", ErrNotFound, p)
		}
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest %s: %w", p, err)
	}
	return &m, nil
}

// Save writes the manifest to <aiDir>/lazyai.json as pretty JSON with a
// trailing newline, creating the directory if needed.
func (m *Manifest) Save(aiDir string) error {
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", aiDir, err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	data = append(data, '\n')
	if err := files.SafeWriteFile(Path(aiDir), data, 0o644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return nil
}

// Validate checks structural invariants: a version, a non-empty target list,
// every target resolvable (Codex rejected), and a recognized profile.
func (m *Manifest) Validate() error {
	var errs []string
	version := strings.TrimSpace(m.Version)
	if version == "" {
		errs = append(errs, "version is required")
	} else if version != SchemaVersion {
		errs = append(errs, fmt.Sprintf("unsupported version %q (want %s)", m.Version, SchemaVersion))
	}
	switch m.Profile {
	case "", "personal", "team":
	default:
		errs = append(errs, fmt.Sprintf("invalid profile %q (want personal or team)", m.Profile))
	}
	if len(m.Targets) == 0 {
		errs = append(errs, "targets must not be empty")
	}
	if _, err := m.ResolveTargets(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid manifest: %s", strings.Join(errs, "; "))
	}
	return nil
}

// resolveTargetsWithTokens is like ResolveTargets but also returns a map from
// each resolved ToolId to all manifest tokens that produced it. This lets
// callers look up adapter blocks by any token that maps to that ToolId,
// even after deduplication.
func (m *Manifest) resolveTargetsWithTokens() ([]types.ToolId, map[types.ToolId][]string, error) {
	seen := make(map[types.ToolId]bool, len(m.Targets))
	tokensForID := make(map[types.ToolId][]string, len(m.Targets))
	out := make([]types.ToolId, 0, len(m.Targets))
	for _, raw := range m.Targets {
		token := strings.ToLower(strings.TrimSpace(raw))
		id, ok := targetAliases[token]
		if !ok {
			return nil, nil, fmt.Errorf("unknown target %q", raw)
		}
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
		tokensForID[id] = append(tokensForID[id], token)
	}
	return out, tokensForID, nil
}

// ResolveTargets maps manifest target tokens to internal tool IDs, preserving
// order and de-duplicating. Unknown tokens error; "codex" gets an explicit
// V2-removal message.
func (m *Manifest) ResolveTargets() ([]types.ToolId, error) {
	out, _, err := m.resolveTargetsWithTokens()
	return out, err
}

// EnabledTargets returns resolved targets whose adapter block is not explicitly
// disabled. A target with no adapter entry defaults to enabled.
func (m *Manifest) EnabledTargets() ([]types.ToolId, error) {
	resolved, tokensForID, err := m.resolveTargetsWithTokens()
	if err != nil {
		return nil, err
	}
	out := make([]types.ToolId, 0, len(resolved))
	for _, id := range resolved {
		disabled := false
		for _, token := range tokensForID[id] {
			if blk, ok := m.Adapters[token]; ok {
				if enabled, present := blk["enabled"].(bool); present && !enabled {
					disabled = true
					break
				}
			}
		}
		if !disabled {
			out = append(out, id)
		}
	}
	return out, nil
}

// Default returns a starter manifest enabling all supported targets with safe
// defaults. Used by `init` to scaffold .ai/lazyai.json.
func Default() *Manifest {
	targets := []string{"opencode", "claude", "copilot", "pi", "omp", "antigravity", "kiro", "codex"}
	sort.Strings(targets)
	return &Manifest{
		Schema:  "https://lazyai.dev/schemas/lazyai.schema.json",
		Version: SchemaVersion,
		Profile: "team",
		Targets: targets,
		Source:  &Source{Assets: []string{"agents", "skills", "rules", "hooks", "prompts", "templates", "standards"}},
		Library: &Library{Packs: []string{"vibe-lab/starter"}},
		Safety: &Safety{
			RequireDiffBeforeWrite: true,
			AllowGlobalWrites:      false,
			DenyInlineSecrets:      true,
			WarnIfNoSandbox:        true,
			GeneratedFileMode:      "managed-region",
		},
	}
}

// TokenFor returns the preferred manifest target token for a tool ID, or the
// raw ID string if unknown.
func TokenFor(id types.ToolId) string {
	if t, ok := canonicalToken[id]; ok {
		return t
	}
	return string(id)
}

// ResolveToolToken maps a single user-facing target token (alias or canonical
// tool ID) to its internal types.ToolId, applying the same alias table the
// manifest uses (targetAliases). The empty string resolves to the zero ToolId
// with ok=false so callers can treat it as "no filter".
// This lets CLI flags such as --tool accept the same short names the manifest
// accepts (e.g. "claude" -> claude-code) without duplicating the alias table.
func ResolveToolToken(raw string) (types.ToolId, bool, error) {
	token := strings.ToLower(strings.TrimSpace(raw))
	if token == "" {
		return "", false, nil
	}
	id, ok := targetAliases[token]
	if !ok {
		return "", false, fmt.Errorf("unknown target %q", raw)
	}
	return id, true, nil
}

// ForTools returns a starter manifest whose targets are the given tools. When
// tools is empty it falls back to Default() (all supported targets).
func ForTools(tools []types.ToolId) *Manifest {
	if len(tools) == 0 {
		return Default()
	}
	m := Default()
	seen := make(map[string]bool, len(tools))
	targets := make([]string, 0, len(tools))
	for _, id := range tools {
		tok := TokenFor(id)
		if !seen[tok] {
			seen[tok] = true
			targets = append(targets, tok)
		}
	}
	sort.Strings(targets)
	m.Targets = targets
	return m
}
