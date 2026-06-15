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

func TestOutputTargetsAllKnownTools(t *testing.T) {
	tools := []types.ToolId{
		types.ToolIdClaudeCode, types.ToolIdOpenCode, types.ToolIdCopilot, types.ToolIdPi, types.ToolIdAntigravity,
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
	if !target.IncludeFile("scout.md") {
		t.Error("scout should be included in bulk agent copy")
	}
}

func TestOutputMappingCopilotSkillsRewritesExt(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdCopilot, AssetKindSkills)
	if !ok {
		t.Fatal("copilot has no skills target")
	}
	if target.Shape != ShapeRewriteExt {
		t.Errorf("copilot skills Shape=%q, want %q", target.Shape, ShapeRewriteExt)
	}
	if target.RewriteSuffix != ".agent.yaml" {
		t.Errorf("copilot skills RewriteSuffix=%q, want %q", target.RewriteSuffix, ".agent.yaml")
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

func TestOutputMappingAntigravitySkillsNone(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdAntigravity, AssetKindSkills)
	if !ok {
		t.Fatal("antigravity has no skills target")
	}
	if target.Shape != ShapeNone {
		t.Errorf("antigravity skills Shape=%q, want %q", target.Shape, ShapeNone)
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
