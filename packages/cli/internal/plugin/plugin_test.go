package plugin

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	libraryembed "github.com/rluisb/lazyai/packages/cli/library"
)

// newTestLibraryFS builds an in-memory library FS that mirrors the real
// library/ layout at the subset of paths the plugin generator touches.
func newTestLibraryFS() fs.FS {
	return fstest.MapFS{
		"canonical/agents/implementer.md": &fstest.MapFile{
			Data: []byte("---\nname: Implementer\nmodel: sonnet\n---\nImplementer prompt body\n"),
		},
		"canonical/agents/planner.md": &fstest.MapFile{
			Data: []byte("---\nname: Planner\nmodel: opus\n---\nPlanner prompt body\n"),
		},
		"canonical/agents/with-forbidden.md": &fstest.MapFile{
			Data: []byte("---\nname: Foo\nhooks:\n  PostToolUse: run\nmcpServers:\n  x:\n    command: y\npermissionMode: default\n---\nbody\n"),
		},
		"skills/implement.md": &fstest.MapFile{
			Data: []byte("---\nname: implement\ntrigger: /implement\n---\nImplement body\n"),
		},
		"skills/plan.md": &fstest.MapFile{
			Data: []byte("---\nname: plan\n---\nPlan body\n"),
		},
		"skills/no-frontmatter.md": &fstest.MapFile{
			Data: []byte("Just a body, no frontmatter.\n"),
		},
		"claudecode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\nname: commit\n---\nCommit command\n"),
		},
		"claudecode/output-styles/terse.md": &fstest.MapFile{
			Data: []byte("---\nname: terse\n---\nTerse style\n"),
		},
	}
}

func TestBuild_WritesManifest(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	res, err := Build(libFS, outDir, "9.9.9")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if res.FileCount == 0 {
		t.Fatalf("expected some files, got 0")
	}

	manifestPath := filepath.Join(outDir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	if manifest["name"] != PluginName {
		t.Errorf("name = %v, want %q", manifest["name"], PluginName)
	}
	if manifest["version"] != "9.9.9" {
		t.Errorf("version = %v, want %q", manifest["version"], "9.9.9")
	}
	if manifest["license"] != PluginLicense {
		t.Errorf("license = %v, want %q", manifest["license"], PluginLicense)
	}
	if PluginName != "lazyai" {
		t.Errorf("PluginName = %q, want lazyai", PluginName)
	}
	if strings.Contains(PluginDescription, "ai-setup") {
		t.Errorf("PluginDescription still references ai-setup: %q", PluginDescription)
	}
	keywords, ok := manifest["keywords"].([]any)
	if !ok {
		t.Fatalf("keywords = %T, want []any", manifest["keywords"])
	}
	foundLazyAI := false
	for _, keyword := range keywords {
		if keyword == "ai-setup" {
			t.Fatalf("keywords still include ai-setup: %v", keywords)
		}
		if keyword == "lazyai" {
			foundLazyAI = true
		}
	}
	if !foundLazyAI {
		t.Fatalf("keywords = %v, want lazyai", keywords)
	}
}

func TestBuild_CopiesAgentsVerbatimWhenNoForbiddenFields(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	src, _ := fs.ReadFile(libFS, "canonical/agents/implementer.md")
	dst, err := os.ReadFile(filepath.Join(outDir, "agents", "implementer.md"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(src) != string(dst) {
		t.Errorf("implementer.md bytes differ:\nsrc: %q\ndst: %q", src, dst)
	}
}

func TestBuild_StripsForbiddenAgentFields(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "agents", "with-forbidden.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	for _, forbidden := range forbiddenAgentFields {
		if strings.Contains(content, forbidden+":") {
			t.Errorf("forbidden field %q not stripped from output:\n%s", forbidden, content)
		}
	}
	// Allowed fields must survive.
	if !strings.Contains(content, "name: Foo") {
		t.Errorf("allowed field 'name' was incorrectly stripped:\n%s", content)
	}
	if !strings.Contains(content, "body") {
		t.Errorf("agent body was stripped:\n%s", content)
	}
}

func TestBuild_RestructuresSkillsIntoSkillMd(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// implement skill must end up at skills/implement/SKILL.md
	dst := filepath.Join(outDir, "skills", "implement", "SKILL.md")
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	fm, body := frontmatter.SplitYamlFrontmatter(string(data))
	if fm == "" {
		t.Fatalf("no frontmatter in restructured skill")
	}
	if !strings.Contains(fm, "name: implement") {
		t.Errorf("skill frontmatter must have name: implement, got:\n%s", fm)
	}
	if !strings.Contains(body, "Implement body") {
		t.Errorf("skill body lost, got:\n%s", body)
	}
}

func TestBuild_SynthesizesNameForSkillWithoutFrontmatter(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "skills", "no-frontmatter", "SKILL.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(string(data), "---\nname: no-frontmatter\n") {
		t.Errorf("expected synthesized frontmatter, got:\n%s", string(data))
	}
}

func TestBuild_CopiesCommandsAndOutputStylesVerbatim(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Commands.
	src, _ := fs.ReadFile(libFS, "claudecode/commands/commit.md")
	dst, err := os.ReadFile(filepath.Join(outDir, "commands", "commit.md"))
	if err != nil || string(src) != string(dst) {
		t.Errorf("commands/commit.md mismatch: err=%v", err)
	}

	// Output styles.
	src, _ = fs.ReadFile(libFS, "claudecode/output-styles/terse.md")
	dst, err = os.ReadFile(filepath.Join(outDir, "output-styles", "terse.md"))
	if err != nil || string(src) != string(dst) {
		t.Errorf("output-styles/terse.md mismatch: err=%v", err)
	}
}

func TestBuild_EmptyLibFSDoesNotError(t *testing.T) {
	outDir := t.TempDir()
	libFS := fstest.MapFS{} // no files at all

	res, err := Build(libFS, outDir, "1.0.0")
	if err != nil {
		t.Fatalf("Build must tolerate empty library: %v", err)
	}
	// Only the manifest should be written.
	if res.FileCount != 1 {
		t.Errorf("expected 1 file (manifest only), got %d", res.FileCount)
	}
}

func TestBuild_RejectsNilLibFS(t *testing.T) {
	if _, err := Build(nil, t.TempDir(), "1.0.0"); err == nil {
		t.Error("expected error for nil libFS")
	}
}

func TestBuild_RejectsEmptyOutDir(t *testing.T) {
	if _, err := Build(newTestLibraryFS(), "", "1.0.0"); err == nil {
		t.Error("expected error for empty outDir")
	}
}
func TestBuild_EmbeddedLibraryEmitsSetupSkills(t *testing.T) {
	outDir := t.TempDir()

	if _, err := Build(libraryembed.FS, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build with embedded library: %v", err)
	}

	skillPath := filepath.Join(outDir, "skills", "issue-triage", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("expected embedded skill output at %s: %v", skillPath, err)
	}
	if !strings.Contains(string(data), "name: issue-triage") {
		t.Fatalf("embedded skill output missing rewritten name: %s", string(data))
	}
}
