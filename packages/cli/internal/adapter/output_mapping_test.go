package adapter

import (
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestOutputCoverageIsExhaustive(t *testing.T) {
	if err := ValidateOutputCoverage(); err != nil {
		t.Fatalf("output mapping coverage failed: %v", err)
	}
}

func TestOutputMappingDoesNotEmitWorkflowDirectories(t *testing.T) {
	for _, kind := range AllAssetKinds() {
		if kind == AssetKind("workflows") || kind == AssetKind("workflow") {
			t.Fatalf("workflow catalog must not be emitted through generic output mapping: %q", kind)
		}
	}
}

func TestOutputTargetsAllKnownTools(t *testing.T) {
	tools := []types.ToolId{
		types.ToolIdClaudeCode, types.ToolIdOpenCode, types.ToolIdCopilot, types.ToolIdPi, types.ToolIdOmp, types.ToolIdKiro, types.ToolIdAntigravity,
	}
	for _, tool := range tools {
		entries, err := OutputTargetsForTool(tool)
		if err != nil {
			t.Errorf("OutputTargetsForTool(%q) returned error: %v", tool, err)
			continue
		}
		if len(entries) != len(AllAssetKinds()) {
			t.Errorf("OutputTargetsForTool(%q) has %d kinds, want %d",
				tool, len(entries), len(AllAssetKinds()))
		}
	}
}

func TestOutputMappingSpeckitTemplatesShipForEveryTool(t *testing.T) {
	for _, tool := range []types.ToolId{
		types.ToolIdClaudeCode, types.ToolIdOpenCode, types.ToolIdCopilot,
	} {
		target, ok := LookupOutputTarget(tool, AssetKindTemplates)
		if !ok {
			t.Errorf("tool %q has no AssetKindTemplates entry", tool)
			continue
		}
		if target.Shape == ShapeNone {
			t.Errorf("tool %q has Shape=ShapeNone for templates — speckit alignment requires templates land somewhere", tool)
		}
		if target.SourceSubdir != "templates" {
			t.Errorf("tool %q templates SourceSubdir=%q, want %q", tool, target.SourceSubdir, "templates")
		}
		if !strings.HasSuffix(target.DestSubdir, "templates") &&
			!strings.Contains(target.DestSubdir, "instructions") {
			t.Errorf("tool %q templates DestSubdir=%q does not look like a templates dir",
				tool, target.DestSubdir)
		}
	}
}

func TestOutputMappingClaudeCodeAgents(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdClaudeCode, AssetKindAgents)
	if !ok {
		t.Fatal("claude-code has no agents target")
	}
	if target.Shape != ShapeFlat {
		t.Errorf("claude agents Shape=%q, want %q", target.Shape, ShapeFlat)
	}
	if target.IncludeFile == nil {
		t.Fatal("claude agents must have an IncludeFile filter")
	}
	if target.IncludeFile("orchestrator.md") {
		t.Error("orchestrator must be excluded from neutral agent mappings")
	}
	if !target.IncludeFile("researcher.md") {
		t.Error("researcher should be included in bulk agent copy")
	}
}

func TestOutputMappingCopilotAgentsWritesMd(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdCopilot, AssetKindAgents)
	if !ok {
		t.Fatal("copilot has no agents target")
	}
	if target.Shape != ShapeFlat {
		t.Errorf("copilot agents Shape=%q, want %q", target.Shape, ShapeFlat)
	}
	if target.RewriteSuffix != "" {
		t.Errorf("copilot agents RewriteSuffix=%q, want empty", target.RewriteSuffix)
	}
	if target.IncludeFile == nil {
		t.Fatal("copilot agents must have an IncludeFile filter")
	}
	if target.IncludeFile("orchestrator.md") {
		t.Error("orchestrator must be excluded from copilot agent mappings")
	}
	if !target.IncludeFile("researcher.md") {
		t.Error("researcher should be included in copilot agent copy")
	}
}

func TestOutputMappingCopilotSkillsRewritesToSkillDirectories(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdCopilot, AssetKindSkills)
	if !ok {
		t.Fatal("copilot has no skills target")
	}
	if target.Shape != ShapeDirPerItem {
		t.Errorf("copilot skills Shape=%q, want %q", target.Shape, ShapeDirPerItem)
	}
	if target.RewriteSuffix != "" {
		t.Errorf("copilot skills RewriteSuffix=%q, want %q", target.RewriteSuffix, "")
	}
}

func TestOutputMappingPiSkillsDirPerItem(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdPi, AssetKindSkills)
	if !ok {
		t.Fatal("pi has no skills target")
	}
	if target.Shape != ShapeDirPerItem {
		t.Errorf("pi skills Shape=%q, want %q", target.Shape, ShapeDirPerItem)
	}
	if target.SourceSubdir != "skills" {
		t.Errorf("pi skills SourceSubdir=%q, want skills", target.SourceSubdir)
	}
}

func TestOutputMappingOmpSkillsDirPerItem(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdOmp, AssetKindSkills)
	if !ok {
		t.Fatal("omp has no skills target")
	}
	if target.Shape != ShapeDirPerItem {
		t.Errorf("omp skills Shape=%q, want %q", target.Shape, ShapeDirPerItem)
	}
	if target.SourceSubdir != "skills" {
		t.Errorf("omp skills SourceSubdir=%q, want skills", target.SourceSubdir)
	}
}

func TestOutputMappingKiroSkillsDirPerItem(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdKiro, AssetKindSkills)
	if !ok {
		t.Fatal("kiro has no skills target")
	}
	if target.Shape != ShapeDirPerItem {
		t.Errorf("kiro skills Shape=%q, want %q", target.Shape, ShapeDirPerItem)
	}
	if target.SourceSubdir != "skills" {
		t.Errorf("kiro skills SourceSubdir=%q, want skills", target.SourceSubdir)
	}
}

func TestOutputMappingAntigravitySkillsDirPerItem(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdAntigravity, AssetKindSkills)
	if !ok {
		t.Fatal("antigravity has no skills target")
	}
	if target.Shape != ShapeDirPerItem {
		t.Errorf("antigravity skills Shape=%q, want %q", target.Shape, ShapeDirPerItem)
	}
	if target.SourceSubdir != "skills" {
		t.Errorf("antigravity skills SourceSubdir=%q, want skills", target.SourceSubdir)
	}
	if target.DestSubdir != "../.agents/skills" {
		t.Errorf("antigravity skills DestSubdir=%q, want ../.agents/skills", target.DestSubdir)
	}
}

func TestOutputMappingDirPerItemShape(t *testing.T) {
	for _, tool := range []types.ToolId{
		types.ToolIdClaudeCode, types.ToolIdOpenCode,
	} {
		target, _ := LookupOutputTarget(tool, AssetKindSkills)
		if target.Shape != ShapeDirPerItem {
			t.Errorf("tool %q skills Shape=%q, want %q (skills must be <name>/SKILL.md)",
				tool, target.Shape, ShapeDirPerItem)
		}
	}
}
