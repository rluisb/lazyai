// Package validate provides the consolidated `validate --all` engine for the
// canonical .ai/ asset tree: skill, agent, hook, and MCP structural validators
// plus a secret scanner (secrets.go) and a path/symlink safety check (paths.go).
//
// All validators append to a single Report so the CLI can render one unified
// result and decide pass/fail from Report.HasErrors. Severity is profile-aware:
// inline secrets are errors under the team profile and warnings under personal
// (FR-010).
package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/validation"
)

// Severity classifies an issue as a hard failure or an advisory.
type Severity string

const (
	// SeverityError fails `validate --all`.
	SeverityError Severity = "error"
	// SeverityWarning is reported but does not fail validation.
	SeverityWarning Severity = "warning"
)

// Profile selects how strict secret handling is (FR-010).
type Profile string

const (
	// ProfileTeam treats inline secrets as errors.
	ProfileTeam Profile = "team"
	// ProfilePersonal treats inline secrets as warnings.
	ProfilePersonal Profile = "personal"
)

// NormalizeProfile maps a raw manifest/flag string to a Profile. Unknown or
// empty values default to personal (the safe, non-blocking default).
func NormalizeProfile(raw string) Profile {
	if strings.EqualFold(strings.TrimSpace(raw), string(ProfileTeam)) {
		return ProfileTeam
	}
	return ProfilePersonal
}

// Issue is a single validation finding tied to a file and rule.
type Issue struct {
	File     string   `json:"file"`
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// Report accumulates issues across every validator.
type Report struct {
	Issues []Issue `json:"issues"`
}

func (r *Report) add(file, rule string, sev Severity, format string, args ...any) {
	r.Issues = append(r.Issues, Issue{
		File:     file,
		Rule:     rule,
		Severity: sev,
		Message:  fmt.Sprintf(format, args...),
	})
}

// Errors returns the number of error-severity issues.
func (r Report) Errors() int { return r.count(SeverityError) }

// Warnings returns the number of warning-severity issues.
func (r Report) Warnings() int { return r.count(SeverityWarning) }

func (r Report) count(sev Severity) int {
	n := 0
	for _, issue := range r.Issues {
		if issue.Severity == sev {
			n++
		}
	}
	return n
}

// HasErrors reports whether any error-severity issue was recorded.
func (r Report) HasErrors() bool { return r.Errors() > 0 }

// Options configures an `All` run.
type Options struct {
	// Root is the project root containing the .ai/ directory.
	Root string
	// Profile controls secret-handling strictness.
	Profile Profile
}

// All runs every validator over the canonical .ai/ tree and returns the
// combined report. A missing .ai/ directory yields a single error issue.
func All(opts Options) Report {
	var r Report
	aiDir := filepath.Join(opts.Root, ".ai")
	if info, err := os.Stat(aiDir); err != nil || !info.IsDir() {
		r.add(".ai", "structure", SeverityError, "canonical .ai/ directory not found; run 'lazyai-cli init' first")
		return r
	}

	validateManifest(opts.Root, aiDir, &r)
	validateSkills(aiDir, &r)
	validateAgents(aiDir, &r)
	validateHooks(aiDir, &r)
	validateMCP(aiDir, opts.Profile, &r)
	scanSecrets(aiDir, opts.Profile, &r)
	checkPaths(opts.Root, aiDir, &r)

	sort.SliceStable(r.Issues, func(i, j int) bool {
		if r.Issues[i].File != r.Issues[j].File {
			return r.Issues[i].File < r.Issues[j].File
		}
		return r.Issues[i].Rule < r.Issues[j].Rule
	})
	return r
}

// validateManifest enforces the canonical manifest contract: .ai/lazyai.json
// must parse and satisfy aimanifest.Validate, which freezes the schema version
// (FR schema-freeze). A missing manifest is tolerated so `validate --all` still
// works on a bare or partial .ai/ tree; a present-but-invalid manifest is a
// hard error, matching `compile`.
func validateManifest(root, aiDir string, r *Report) {
	mf, err := aimanifest.Load(aiDir)
	if err != nil {
		if errors.Is(err, aimanifest.ErrNotFound) {
			return
		}
		r.add(relForReport(aiDir, aimanifest.Path(aiDir)), "manifest", SeverityError, "%v", err)
		return
	}
	if err := mf.Validate(); err != nil {
		r.add(relForReport(aiDir, aimanifest.Path(aiDir)), "manifest", SeverityError, "%v", err)
	}
}

// markdownAssets returns the markdown asset files under aiDir/<category>,
// supporting both flat "<name>.md" files and "<name>/SKILL.md" directories.
func markdownAssets(aiDir, category string) []assetFile {
	dir := filepath.Join(aiDir, category)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var assets []assetFile
	for _, entry := range entries {
		if entry.IsDir() {
			candidate := filepath.Join(dir, entry.Name(), "SKILL.md")
			if _, statErr := os.Stat(candidate); statErr == nil {
				assets = append(assets, assetFile{
					rel:  filepath.Join(category, entry.Name(), "SKILL.md"),
					abs:  candidate,
					name: entry.Name(),
				})
			}
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		assets = append(assets, assetFile{
			rel:  filepath.Join(category, entry.Name()),
			abs:  filepath.Join(dir, entry.Name()),
			name: strings.TrimSuffix(entry.Name(), ".md"),
		})
	}
	return assets
}

type assetFile struct {
	rel  string // path relative to project root for reporting
	abs  string // absolute path on disk
	name string // derived artifact name (frontmatter name takes precedence)
}

// validateSkills checks .ai/skills assets: frontmatter present, a valid name,
// and a non-empty description (FR-009).
func validateSkills(aiDir string, r *Report) {
	for _, asset := range markdownAssets(aiDir, "skills") {
		content, err := os.ReadFile(asset.abs)
		if err != nil {
			r.add(asset.rel, "skill", SeverityError, "unreadable: %v", err)
			continue
		}
		fm, _, err := frontmatter.ExtractFrontmatter(content)
		if err != nil || fm == nil {
			r.add(asset.rel, "skill", SeverityError, "missing or invalid YAML frontmatter")
			continue
		}
		name := frontmatter.ExtractField(fm, "name")
		if name == "" {
			r.add(asset.rel, "skill", SeverityError, "frontmatter missing 'name' field")
		} else if nameErr := validation.ValidateArtifactName(name); nameErr != nil {
			r.add(asset.rel, "skill", SeverityError, "invalid skill name %q: %v", name, nameErr)
		}
		if strings.TrimSpace(frontmatter.ExtractField(fm, "description")) == "" {
			r.add(asset.rel, "skill", SeverityError, "frontmatter missing 'description' field")
		}
	}
}

// validateAgents checks .ai/agents assets: frontmatter present, a valid name,
// and a non-empty description.
func validateAgents(aiDir string, r *Report) {
	for _, asset := range markdownAssets(aiDir, "agents") {
		content, err := os.ReadFile(asset.abs)
		if err != nil {
			r.add(asset.rel, "agent", SeverityError, "unreadable: %v", err)
			continue
		}
		fm, _, err := frontmatter.ExtractFrontmatter(content)
		if err != nil || fm == nil {
			r.add(asset.rel, "agent", SeverityError, "missing or invalid YAML frontmatter")
			continue
		}
		name := frontmatter.ExtractField(fm, "name")
		if name == "" {
			r.add(asset.rel, "agent", SeverityError, "frontmatter missing 'name' field")
		} else if nameErr := validation.ValidateArtifactName(name); nameErr != nil {
			r.add(asset.rel, "agent", SeverityError, "invalid agent name %q: %v", name, nameErr)
		}
		if strings.TrimSpace(frontmatter.ExtractField(fm, "description")) == "" {
			r.add(asset.rel, "agent", SeverityWarning, "frontmatter missing 'description' field")
		}
	}
}

// dangerousHookPatterns are shell fragments that make a hook script unsafe to
// ship. Matches are reported as errors.
var dangerousHookPatterns = []string{
	"rm -rf /",
	"rm -rf ~",
	":(){:|:&};:", // fork bomb
	"mkfs",
	"dd if=",
	"curl | sh",
	"curl | bash",
	"wget | sh",
	"wget | bash",
	"eval \"$(curl",
	"chmod -R 777 /",
}

// validateHooks checks .ai/hooks scripts: must declare a shebang and must not
// contain destructive shell fragments (FR-009; SC-003 dangerous hooks fail).
func validateHooks(aiDir string, r *Report) {
	hooksDir := filepath.Join(aiDir, "hooks")
	_ = filepath.WalkDir(hooksDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel := relForReport(aiDir, path)
		ext := strings.ToLower(filepath.Ext(d.Name()))
		// Only shell scripts get shebang/safety treatment; .json/.ts policy
		// files are config, not executables.
		isShell := ext == ".sh" || ext == ".bash" || ext == ""
		content, err := os.ReadFile(path)
		if err != nil {
			r.add(rel, "hook", SeverityError, "unreadable: %v", err)
			return nil
		}
		text := string(content)
		normalized := strings.Join(strings.Fields(text), " ")
		for _, bad := range dangerousHookPatterns {
			needle := strings.Join(strings.Fields(bad), " ")
			if strings.Contains(normalized, needle) {
				r.add(rel, "hook", SeverityError, "contains dangerous command %q", bad)
			}
		}
		if isShell && !strings.HasPrefix(text, "#!") {
			r.add(rel, "hook", SeverityWarning, "shell hook missing shebang line")
		}
		return nil
	})
}

// mcpServers extracts the server map from a parsed mcp.json, accepting either
// the "servers" or "mcpServers" key.
func mcpServers(doc map[string]any) map[string]any {
	if m, ok := doc["servers"].(map[string]any); ok && m != nil {
		return m
	}
	if m, ok := doc["mcpServers"].(map[string]any); ok && m != nil {
		return m
	}
	return nil
}

// validateMCP checks .ai/mcp.json structure: parseable, a servers map exists,
// and every server entry declares a command or url (FR-009).
func validateMCP(aiDir string, _ Profile, r *Report) {
	path := filepath.Join(aiDir, "mcp.json")
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join(aiDir, "mcp.jsonc")
		if _, altErr := os.Stat(alt); altErr != nil {
			return // no MCP config is valid (MCP is optional)
		}
		path = alt
	}
	rel := relForReport(aiDir, path)
	doc, err := jsonc.ReadJSONCFile(path)
	if err != nil {
		r.add(rel, "mcp", SeverityError, "not valid JSON: %v", err)
		return
	}
	servers := mcpServers(doc)
	if servers == nil {
		r.add(rel, "mcp", SeverityError, "missing 'servers' map")
		return
	}
	for name, raw := range servers {
		entry, ok := raw.(map[string]any)
		if !ok {
			r.add(rel, "mcp", SeverityError, "server %q is not an object", name)
			continue
		}
		_, hasCmd := entry["command"]
		_, hasURL := entry["url"]
		if !hasCmd && !hasURL {
			r.add(rel, "mcp", SeverityError, "server %q has neither 'command' nor 'url'", name)
		}
	}
}

// relForReport renders a path relative to the parent of aiDir (the project
// root) so report paths look like ".ai/hooks/foo.sh".
func relForReport(aiDir, path string) string {
	root := filepath.Dir(aiDir)
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}
