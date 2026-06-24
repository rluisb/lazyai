package adapter

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestCopilotAdapter_Install_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents:  types.ALL_AGENTS[:],
		Skills:  types.ALL_SKILLS[:],
		Prompts: []types.PromptId{"preflight-task-framing"},
	}

	agentsDir := filepath.Join(targetDir, ".github", "agents")
	skillsDir := filepath.Join(targetDir, ".github", "skills", "diagnose")
	legacyMarkdownPath := filepath.Join(agentsDir, "diagnose.agent.md")
	legacyYamlPath := filepath.Join(agentsDir, "diagnose.agent.yaml")
	legacySkillSource, err := fs.ReadFile(ctx.LibraryFS, filepath.ToSlash(filepath.Join("skills", "diagnose.md")))
	if err != nil {
		t.Fatalf("reading legacy skill source fixture failed: %v", err)
	}
	legacyMarkdown, err := skillToCopilotAgentMarkdown(ctx, filepath.ToSlash(filepath.Join("skills", "diagnose.md")), string(legacySkillSource))
	if err != nil {
		t.Fatalf("construct legacy copilot skill content failed: %v", err)
	}
	_, legacyBody, err := frontmatter.ExtractFrontmatter(legacySkillSource)
	if err != nil {
		t.Fatalf("extracting legacy skill frontmatter failed: %v", err)
	}
	legacyYaml := "name: diagnose\nmodel: claude-sonnet-4.6\nprompt: |\n" + indentLines(string(legacyBody), "  ")
	if err := files.EnsureDir(agentsDir); err != nil {
		t.Fatalf("creating agent output dir for legacy cleanup fixture failed: %v", err)
	}
	if err := os.WriteFile(legacyMarkdownPath, []byte(legacyMarkdown), 0o644); err != nil {
		t.Fatalf("writing legacy markdown fixture failed: %v", err)
	}
	if err := os.WriteFile(legacyYamlPath, []byte(legacyYaml), 0o644); err != nil {
		t.Fatalf("writing legacy yaml fixture failed: %v", err)
	}

	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Copilot Install failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// --- Prompts copied to .github/prompts/*.prompt.md ---
	promptsDir := filepath.Join(targetDir, ".github", "prompts")
	promptFile := filepath.Join(promptsDir, "preflight-task-framing.prompt.md")
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Error("prompt .prompt.md was not created in .github/prompts/")
	}

	// --- Selected skills are emitted to Agent Skills directories ---
	skillDir := filepath.Join(skillsDir, "SKILL.md")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("selected skill was not created at .github/skills/diagnose/SKILL.md")
	}
	data, _ := os.ReadFile(skillDir)
	content := string(data)
	if !strings.Contains(content, "---\nname: diagnose\n") {
		t.Error("skill output missing canonical source frontmatter")
	}
	if !strings.Contains(content, "# diagnose") {
		t.Error("skill output missing canonical skill body")
	}

	// Migration cleanup removed legacy skill-as-agent artifacts.
	if _, err := os.Stat(legacyMarkdownPath); err == nil {
		t.Error("legacy diagnose.agent.md should be removed during migration")
	}
	if _, err := os.Stat(legacyYamlPath); err == nil {
		t.Error("legacy diagnose.agent.yaml should be removed during migration")
	}

	// --- Default agents are compiled as Markdown files ---
	defaultAgentFile := filepath.Join(agentsDir, "guide.agent.md")
	if _, err := os.Stat(defaultAgentFile); os.IsNotExist(err) {
		t.Error("default agent .agent.md was not created in .github/agents/")
	}
	legacyDefaultAgentFile := filepath.Join(agentsDir, "guide.agent.yaml")
	if _, err := os.Stat(legacyDefaultAgentFile); err == nil {
		t.Error("guide.agent.yaml should not be emitted for default copilot agents")
	}

	// Root AGENTS.md and .github/copilot-instructions.md are emitted by
	// scaffold.ScaffoldCompiledRoot (scope-aware) rather than the adapter;
	// asserting them here would test the wrong layer.

	// --- Tracked file records created (prompts + agents + skills) ---
	if len(ctx.FileRecords) < 3 {
		t.Errorf("expected at least 3 tracked file records, got %d", len(ctx.FileRecords))
	}
	hasPreFlight := false
	hasDiagnose := false
	hasSkillDir := false
	for _, rec := range ctx.FileRecords {
		switch rec.Path {
		case ".github/prompts/preflight-task-framing.prompt.md":
			hasPreFlight = true
		case ".github/skills/diagnose/SKILL.md":
			hasDiagnose = true
			hasSkillDir = true
		}
	}
	if !hasPreFlight {
		t.Error("no tracked record for preflight-task-framing.prompt.md")
	}
	if !hasDiagnose || !hasSkillDir {
		t.Error("no tracked record for diagnose skill output at .github/skills/diagnose/SKILL.md")
	}
}

func TestClaudeCodeAdapter_Install_CopiesHookScriptsAndSettings(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &ClaudeCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("ClaudeCode Install failed: %v", err)
	}

	for _, rel := range []string{
		".claude/hooks/block-destructive-shell.sh",
		".claude/hooks/objective-workflow-gate.sh",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}

	settings, err := os.ReadFile(filepath.Join(targetDir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	for _, want := range []string{
		"block-destructive-shell.sh",
		"objective-workflow-gate.sh",
	} {
		if !strings.Contains(string(settings), want) {
			t.Fatalf("settings.json missing hook reference %q: %s", want, string(settings))
		}
	}
	if strings.Contains(string(settings), "startup-self-heal.sh") {
		t.Fatalf("settings.json should not reference startup-self-heal.sh: %s", string(settings))
	}
}

func TestOpenCodeAdapter_Install_CopiesHookPlugin(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &OpenCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	pluginPath := filepath.Join(targetDir, ".opencode", "plugins", "vibe-lab-hooks.js")
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read hook plugin: %v", err)
	}
	if !strings.Contains(string(data), "VibeLabHooks") {
		t.Fatalf("hook plugin missing export: %s", string(data))
	}
}

func TestPiAdapter_Install_AgentsAndSkills(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
		Agents: []types.AgentId{types.AgentIdResearcher},
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}
	for _, rel := range []string{
		".pi/agents/researcher.md",
		".pi/skills/diagnose/SKILL.md",
		".pi/skills/issue-triage/SKILL.md",
		".pi/extensions/block-destructive-shell.ts",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	// Pi has no .pi/hooks path.
	assertMissing(t, filepath.Join(targetDir, ".pi", "hooks"))
}

func TestKiroAdapter_Install_AgentsAndSkills(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdReviewer},
		Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
	}

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	for _, rel := range []string{
		".kiro/agents/guide.md",
		".kiro/agents/reviewer.md",
		".kiro/skills/diagnose/SKILL.md",
		".kiro/skills/issue-triage/SKILL.md",
		".kiro/hooks/block-destructive-shell.json",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	assertMissing(t, filepath.Join(targetDir, ".kiro", "specs"))
	assertMissing(t, filepath.Join(targetDir, ".kiro", "steering"))
}

func TestAntigravityAdapter_Install_MinimalSurface(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, ".gemini", "settings.json")); err != nil {
		t.Fatalf("expected .gemini/settings.json: %v", err)
	}
	for _, rel := range []string{
		".gemini/hooks/lazyai/block-destructive-shell.sh",
		".gemini/hooks/lazyai/objective-workflow-gate.sh",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	assertMissing(t, filepath.Join(targetDir, ".gemini", "agents"))
	assertMissing(t, filepath.Join(targetDir, ".gemini", "skills"))
}
