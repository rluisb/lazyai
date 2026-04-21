// Package plugin generates a Claude Code plugin directory from ai-setup's
// embedded library. The output follows the official plugin layout
// (.claude-plugin/plugin.json + agents/skills/commands/output-styles at the
// plugin root) and can be installed via `claude --plugin-dir <path>`.
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

// Metadata about ai-setup as a Claude plugin. Hardcoded per spec 016 Q4.
const (
	PluginName        = "ai-setup"
	PluginDescription = "ai-setup agents, skills, commands, and output styles for Claude Code"
	PluginAuthorName  = "Ricardo Borges"
	PluginAuthorURL   = "https://github.com/ricardoborges-teachable/ai-setup"
	PluginHomepage    = "https://github.com/ricardoborges-teachable/ai-setup"
	PluginRepository  = "https://github.com/ricardoborges-teachable/ai-setup"
	PluginLicense     = "MIT"
)

// Fields forbidden in plugin-shipped agent frontmatter (spec 016 §6 R2).
// Upstream docs explicitly prohibit these for security reasons.
var forbiddenAgentFields = []string{"hooks", "mcpServers", "permissionMode"}

// BuildResult reports what the generator emitted.
type BuildResult struct {
	OutDir   string
	FileCount int
}

// Build generates a Claude Code plugin directory under outDir from the given
// library FS. version is written into plugin.json (typically cmd.Version).
// Callers are responsible for pre-flight checks on outDir (e.g. --force).
func Build(libFS fs.FS, outDir, version string) (BuildResult, error) {
	if libFS == nil {
		return BuildResult{}, fmt.Errorf("plugin.Build: libFS is nil")
	}
	if outDir == "" {
		return BuildResult{}, fmt.Errorf("plugin.Build: outDir is empty")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return BuildResult{}, fmt.Errorf("create outDir: %w", err)
	}

	total := 0

	// Manifest.
	if err := buildManifest(outDir, version); err != nil {
		return BuildResult{}, err
	}
	total++

	// Agents: library/agents/*.md → outDir/agents/*.md (strip forbidden fields).
	n, err := copyAgents(libFS, "agents", filepath.Join(outDir, "agents"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("copy agents: %w", err)
	}
	total += n

	// Commands: library/claudecode/commands/*.md → outDir/commands/*.md
	n, err = copyFlat(libFS, "claudecode/commands", filepath.Join(outDir, "commands"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("copy commands: %w", err)
	}
	total += n

	// Output styles: library/claudecode/output-styles/*.md → outDir/output-styles/*.md
	n, err = copyFlat(libFS, "claudecode/output-styles", filepath.Join(outDir, "output-styles"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("copy output-styles: %w", err)
	}
	total += n

	// Skills: library/skills/*.md → outDir/skills/<name>/SKILL.md
	n, err = restructureSkills(libFS, filepath.Join(outDir, "skills"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("restructure skills: %w", err)
	}
	total += n

	return BuildResult{OutDir: outDir, FileCount: total}, nil
}

// buildManifest writes .claude-plugin/plugin.json.
func buildManifest(outDir, version string) error {
	manifestDir := filepath.Join(outDir, ".claude-plugin")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return err
	}
	manifest := map[string]any{
		"name":        PluginName,
		"version":     version,
		"description": PluginDescription,
		"author": map[string]any{
			"name": PluginAuthorName,
			"url":  PluginAuthorURL,
		},
		"homepage":   PluginHomepage,
		"repository": PluginRepository,
		"license":    PluginLicense,
		"keywords":   []string{"ai-setup", "claude-code", "agents", "skills"},
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(manifestDir, "plugin.json"), data, 0o644)
}

// copyFlat copies every *.md file from libFS/srcSubdir to destDir verbatim.
// Directories inside srcSubdir are skipped (we only emit top-level files).
func copyFlat(libFS fs.FS, srcSubdir, destDir string) (int, error) {
	entries, err := fs.ReadDir(libFS, srcSubdir)
	if err != nil {
		// Missing source dir is not an error — just emit nothing.
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		src := srcSubdir + "/" + entry.Name()
		data, err := fs.ReadFile(libFS, src)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", src, err)
		}
		if err := os.WriteFile(filepath.Join(destDir, entry.Name()), data, 0o644); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// copyAgents is like copyFlat but strips plugin-forbidden frontmatter fields
// (hooks, mcpServers, permissionMode) from each agent file before writing.
func copyAgents(libFS fs.FS, srcSubdir, destDir string) (int, error) {
	entries, err := fs.ReadDir(libFS, srcSubdir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		src := srcSubdir + "/" + entry.Name()
		data, err := fs.ReadFile(libFS, src)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", src, err)
		}
		sanitized, warned := sanitizeAgentFrontmatter(data)
		if warned {
			fmt.Fprintf(os.Stderr, "[plugin] stripped forbidden fields from %s (plugin agents cannot ship hooks/mcpServers/permissionMode)\n", entry.Name())
		}
		if err := os.WriteFile(filepath.Join(destDir, entry.Name()), sanitized, 0o644); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// sanitizeAgentFrontmatter removes plugin-forbidden keys from the YAML
// frontmatter block. Returns the rewritten content and a flag indicating
// whether any fields were stripped.
func sanitizeAgentFrontmatter(data []byte) ([]byte, bool) {
	fmBody, body := frontmatter.SplitYamlFrontmatter(string(data))
	if fmBody == "" {
		return data, false
	}
	lines := strings.Split(fmBody, "\n")
	out := make([]string, 0, len(lines))
	stripped := false
	skip := false
	for _, line := range lines {
		// Multi-line YAML block continuation (indented lines).
		if skip {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				continue
			}
			skip = false
		}
		trimmed := strings.TrimSpace(line)
		isForbidden := false
		for _, f := range forbiddenAgentFields {
			if strings.HasPrefix(trimmed, f+":") {
				isForbidden = true
				break
			}
		}
		if isForbidden {
			stripped = true
			// If the value continues on the next indented lines, skip those too.
			skip = true
			continue
		}
		out = append(out, line)
	}
	newFM := strings.Join(out, "\n")
	return []byte("---\n" + newFM + "\n---\n" + body), stripped
}

// restructureSkills converts flat library/skills/*.md into the plugin layout
// skills/<name>/SKILL.md. The directory name is derived from the source
// basename, and the frontmatter "name" field is rewritten to match (guarantees
// consistent invocation names per upstream docs).
func restructureSkills(libFS fs.FS, destDir string) (int, error) {
	entries, err := fs.ReadDir(libFS, "skills")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		src := "skills/" + entry.Name()
		data, err := fs.ReadFile(libFS, src)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", src, err)
		}
		skillName := strings.TrimSuffix(entry.Name(), ".md")
		skillDir := filepath.Join(destDir, skillName)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return count, err
		}
		rewritten := rewriteSkillName(data, skillName)
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), rewritten, 0o644); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// rewriteSkillName ensures the frontmatter "name:" field matches dirName so
// the skill's invocation name is deterministic. If there is no frontmatter
// or no name field, one is prepended.
func rewriteSkillName(data []byte, dirName string) []byte {
	fmBody, body := frontmatter.SplitYamlFrontmatter(string(data))
	if fmBody == "" {
		// No frontmatter — synthesize one.
		return []byte(fmt.Sprintf("---\nname: %s\n---\n%s", dirName, string(data)))
	}
	lines := strings.Split(fmBody, "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name:") {
			lines[i] = "name: " + dirName
			found = true
			break
		}
	}
	if !found {
		// Prepend the name line.
		lines = append([]string{"name: " + dirName}, lines...)
	}
	newFM := strings.Join(lines, "\n")
	return []byte("---\n" + newFM + "\n---\n" + body)
}
