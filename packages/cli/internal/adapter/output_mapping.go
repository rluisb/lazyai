// Package adapter — output_mapping.go is the single source of truth for
// where each library asset kind (agents, skills, templates, commands,
// chatmodes, output-styles) lands in each tool's on-disk layout. Spec 022
// (Phase E2.1) added it so adapter parity can be asserted in tests instead
// of being implicit in per-adapter Install code.
//
// The mapping is declarative: each (tool, asset-kind) pair returns an
// OutputTarget describing the destination subdirectory, the per-file
// destination shape (flat / dir-per-item / extension rewrite), and an
// optional transform applied at write time. Adapters consult this table
// rather than hard-coding paths so future asset kinds plug in by adding a
// row, not by touching every adapter.
package adapter

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// AssetKind enumerates the library asset kinds we install per tool.
type AssetKind string

const (
	AssetKindAgents        AssetKind = "agents"
	AssetKindSkills        AssetKind = "skills"
	AssetKindTemplates     AssetKind = "templates"
	AssetKindCommands      AssetKind = "commands"
	AssetKindChatModes     AssetKind = "chatmodes"
	AssetKindOutputStyles  AssetKind = "output-styles"
	AssetKindPrompts       AssetKind = "prompts"
	AssetKindSystemPrompts AssetKind = "system-prompts"
)

// AllAssetKinds returns every supported asset kind in a stable order.
func AllAssetKinds() []AssetKind {
	return []AssetKind{
		AssetKindAgents,
		AssetKindSkills,
		AssetKindTemplates,
		AssetKindCommands,
		AssetKindChatModes,
		AssetKindOutputStyles,
		AssetKindPrompts,
		AssetKindSystemPrompts,
	}
}

// OutputShape describes how a single library file maps to its on-disk
// destination filename within the asset's destination directory.
type OutputShape string

const (
	// ShapeFlat copies the file under the destination directory using its
	// original basename (e.g. "researcher.md" → "<dest>/researcher.md").
	ShapeFlat OutputShape = "flat"
	// ShapeDirPerItem creates a per-item subdirectory and writes the file as
	// SKILL.md inside it (e.g. "review.md" → "<dest>/review/SKILL.md").
	ShapeDirPerItem OutputShape = "dir-per-item"
	// ShapeRewriteExt rewrites the extension (e.g. "review.md" → "<dest>/review.prompt.md").
	ShapeRewriteExt OutputShape = "rewrite-ext"
	// ShapeNone means the asset is intentionally not installed for this tool
	// (e.g. Gemini has no agents concept).
	ShapeNone OutputShape = "none"
)

// OutputTarget describes the destination shape for one (tool, asset) pair.
type OutputTarget struct {
	// Tool identifies which adapter owns this mapping.
	Tool types.ToolId
	// Kind identifies the library asset kind being mapped.
	Kind AssetKind
	// SourceSubdir is the library-relative source directory (e.g. "agents",
	// "skills", "claudecode/commands"). Empty when Shape == ShapeNone.
	SourceSubdir string
	// DestSubdir is the per-tool destination subdirectory under the tool root
	// (e.g. ".claude/agents", ".opencode/skills"). Reported as the suffix only
	// — the tool root itself is resolved at install time.
	DestSubdir string
	// Shape declares how the file basename maps to the destination filename.
	Shape OutputShape
	// RewriteSuffix is the new extension (incl. dot) when Shape == ShapeRewriteExt.
	RewriteSuffix string
	// IncludeFile is an optional filter; when non-nil, returning false skips
	// the file.
	IncludeFile func(filename string) bool
	// Notes documents *why* a particular mapping exists, surfaced in test
	// failures and in the `info` command.
	Notes string
}

// outputMappings holds the canonical (tool × asset-kind) → OutputTarget table.
// Populated lazily by buildOutputMappings so subdir constants stay in one
// place.
var outputMappings map[types.ToolId]map[AssetKind]OutputTarget

func buildOutputMappings() map[types.ToolId]map[AssetKind]OutputTarget {
	if outputMappings != nil {
		return outputMappings
	}

	canonicalAgents := func(file string) bool { return isCanonicalAgentFile(file) }

	m := map[types.ToolId]map[AssetKind]OutputTarget{
		types.ToolIdClaudeCode: {
			AssetKindAgents: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "Claude Code reads agents from .claude/agents/<name>.md",
			},
			AssetKindSkills: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
				Notes: "Claude Code reads skills as .claude/skills/<name>/SKILL.md",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindTemplates,
				SourceSubdir: "templates", DestSubdir: "templates",
				Shape: ShapeFlat,
				Notes: "Speckit templates land alongside agents at .claude/templates/",
			},
			AssetKindCommands: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindCommands,
				SourceSubdir: "claudecode/commands", DestSubdir: "commands",
				Shape: ShapeFlat,
				Notes: "Claude Code slash commands at .claude/commands/<name>.md",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindOutputStyles,
				SourceSubdir: "claudecode/output-styles", DestSubdir: "output-styles",
				Shape: ShapeFlat,
				Notes: "Output styles at .claude/output-styles/<name>.md",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindChatModes,
				Shape: ShapeNone,
				Notes: "Claude Code has no chat modes concept",
			},
			AssetKindPrompts: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindPrompts,
				Shape: ShapeNone,
				Notes: "Claude Code has no prompts directory; prompts ship as commands",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdClaudeCode, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "Claude Code has no project system-prompt file; use CLAUDE.md context",
			},
		},
		types.ToolIdOpenCode: {
			AssetKindAgents: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "OpenCode reads agents from .opencode/agents/<name>.md after frontmatter rewrite",
			},
			AssetKindSkills: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
			},
			AssetKindTemplates: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindTemplates,
				Shape: ShapeNone,
				Notes: "OpenCode has no documented template surface",
			},
			AssetKindCommands: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindCommands,
				SourceSubdir: "opencode/commands", DestSubdir: "commands",
				Shape: ShapeFlat,
			},
			AssetKindChatModes: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindChatModes,
				SourceSubdir: "opencode/modes", DestSubdir: "modes",
				Shape: ShapeFlat,
				Notes: "OpenCode chat modes at .opencode/modes/<name>.md",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
				Notes: "OpenCode has no output-styles concept",
			},
			AssetKindPrompts: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindPrompts,
				Shape: ShapeNone,
				Notes: "OpenCode prompts ship as commands",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdOpenCode, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "OpenCode has no project system-prompt file; use AGENTS.md context",
			},
		},
		types.ToolIdCopilot: {
			AssetKindAgents: {
				Tool: types.ToolIdCopilot, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "Copilot agents are generated from canonical markdown into .github/agents/<name>.agent.md",
			},
			AssetKindSkills: {
				Tool: types.ToolIdCopilot, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
				Notes: "Copilot selected skills are emitted as Agent Skills directories under .github/skills/<name>/SKILL.md (global uses .copilot/skills)",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdCopilot, Kind: AssetKindTemplates,
				SourceSubdir: "templates", DestSubdir: "instructions",
				Shape: ShapeFlat,
				Notes: "Speckit templates land in .github/instructions/ where Copilot reads them",
			},
			AssetKindCommands: {
				Tool: types.ToolIdCopilot, Kind: AssetKindCommands,
				Shape: ShapeNone,
				Notes: "Copilot has no slash commands surface",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdCopilot, Kind: AssetKindChatModes,
				SourceSubdir: "chatmodes", DestSubdir: "agents",
				Shape: ShapeFlat,
				Notes: "Copilot custom agents (migrated from chat modes) at .github/agents/<name>.agent.md",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdCopilot, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
			},
			AssetKindPrompts: {
				Tool: types.ToolIdCopilot, Kind: AssetKindPrompts,
				SourceSubdir: "prompts", DestSubdir: "prompts",
				Shape: ShapeRewriteExt, RewriteSuffix: ".prompt.md",
				Notes: "Copilot prompts at .github/prompts/<name>.prompt.md",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdCopilot, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "Copilot has no project system-prompt file; use .github/copilot-instructions.md",
			},
		},
		types.ToolIdPi: {
			AssetKindAgents: {
				Tool: types.ToolIdPi, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "Pi subagent extension reads markdown agent definitions from .pi/agents/<name>.md",
			},
			AssetKindSkills: {
				Tool: types.ToolIdPi, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
				Notes: "Pi reads skills as .pi/skills/<name>/SKILL.md",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdPi, Kind: AssetKindTemplates,
				Shape: ShapeNone,
				Notes: "Pi has no template surface",
			},
			AssetKindCommands: {
				Tool: types.ToolIdPi, Kind: AssetKindCommands,
				Shape: ShapeNone,
				Notes: "Pi has no slash command surface",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdPi, Kind: AssetKindChatModes,
				Shape: ShapeNone,
				Notes: "Pi has no chat mode surface",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdPi, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
			},
			AssetKindPrompts: {
				Tool: types.ToolIdPi, Kind: AssetKindPrompts,
				SourceSubdir: "prompts", DestSubdir: "prompts",
				Shape: ShapeFlat,
				Notes: "Pi loads prompt templates from .pi/prompts/<name>.md",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdPi, Kind: AssetKindSystemPrompts,
				SourceSubdir: "pi", DestSubdir: ".",
				Shape:       ShapeFlat,
				IncludeFile: isPiSystemPromptFile,
				Notes:       "Pi reads .pi/SYSTEM.md (replaces default prompt) and .pi/APPEND_SYSTEM.md (appends); distinct from AGENTS.md context files",
			},
		},
		types.ToolIdOmp: {
			AssetKindAgents: {
				Tool: types.ToolIdOmp, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "OMP reads task agents from .omp/agents/<name>.md",
			},
			AssetKindSkills: {
				Tool: types.ToolIdOmp, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
				Notes: "OMP reads skills as .omp/skills/<name>/SKILL.md",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdOmp, Kind: AssetKindTemplates,
				Shape: ShapeNone,
				Notes: "OMP has no template surface",
			},
			AssetKindCommands: {
				Tool: types.ToolIdOmp, Kind: AssetKindCommands,
				SourceSubdir: "canonical/commands", DestSubdir: "commands",
				Shape: ShapeFlat,
				Notes: "OMP reads slash commands from .omp/commands/<name>.md",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdOmp, Kind: AssetKindChatModes,
				Shape: ShapeNone,
				Notes: "OMP has no chat mode surface",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdOmp, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
			},
			AssetKindPrompts: {
				Tool: types.ToolIdOmp, Kind: AssetKindPrompts,
				SourceSubdir: "prompts", DestSubdir: "prompts",
				Shape: ShapeFlat,
				Notes: "OMP loads prompt templates from .omp/prompts/<name>.md",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdOmp, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "OMP has no project system-prompt file; use AGENTS.md context",
			},
		},
		types.ToolIdKiro: {
			AssetKindAgents: {
				Tool: types.ToolIdKiro, Kind: AssetKindAgents,
				SourceSubdir: "canonical/agents", DestSubdir: "agents",
				Shape: ShapeFlat, IncludeFile: canonicalAgents,
				Notes: "Kiro CLI v3 reads custom agent profiles from .kiro/agents/<name>.md",
			},
			AssetKindSkills: {
				Tool: types.ToolIdKiro, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "skills",
				Shape: ShapeDirPerItem,
				Notes: "Kiro reads skills as .kiro/skills/<name>/SKILL.md",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdKiro, Kind: AssetKindTemplates,
				Shape: ShapeNone,
				Notes: "Kiro has no template surface",
			},
			AssetKindCommands: {
				Tool: types.ToolIdKiro, Kind: AssetKindCommands,
				Shape: ShapeNone,
				Notes: "Kiro has no slash command surface",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdKiro, Kind: AssetKindChatModes,
				Shape: ShapeNone,
				Notes: "Kiro has no chat mode surface",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdKiro, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
			},
			AssetKindPrompts: {
				Tool: types.ToolIdKiro, Kind: AssetKindPrompts,
				SourceSubdir: "prompts",
				DestSubdir:   "prompts",
				Shape:        ShapeFlat,
				Notes:        "Kiro CLI reads prompts from .kiro/prompts/<name>.md",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdKiro, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "Kiro has no project system-prompt file; use .kiro/steering or agent specs",
			},
		},
		types.ToolIdAntigravity: {
			AssetKindAgents: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindAgents,
				Shape: ShapeNone,
				Notes: "Antigravity does not emit agent files",
			},
			AssetKindSkills: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindSkills,
				SourceSubdir: "skills", DestSubdir: "../.agents/skills",
				Shape: ShapeDirPerItem,
				Notes: "Antigravity CLI discovers local Agent Skills at .agents/skills/<name>/SKILL.md while LazyAI keeps settings/hooks under .gemini.",
			},
			AssetKindTemplates: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindTemplates,
				Shape: ShapeNone,
				Notes: "Antigravity has no template surface",
			},
			AssetKindCommands: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindCommands,
				Shape: ShapeNone,
				Notes: "Antigravity has no slash command surface",
			},
			AssetKindChatModes: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindChatModes,
				Shape: ShapeNone,
				Notes: "Antigravity has no chat mode surface",
			},
			AssetKindOutputStyles: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindOutputStyles,
				Shape: ShapeNone,
			},
			AssetKindPrompts: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindPrompts,
				Shape: ShapeNone,
				Notes: "Antigravity has no prompt surface",
			},
			AssetKindSystemPrompts: {
				Tool: types.ToolIdAntigravity, Kind: AssetKindSystemPrompts,
				Shape: ShapeNone,
				Notes: "Antigravity has no project system-prompt file; use GEMINI.md context",
			},
		},
	}
	outputMappings = m
	return m
}

// LookupOutputTarget returns the OutputTarget for a (tool, asset-kind) pair.
// Returns ok=false if the tool has no entry; the table is exhaustive over
// AssetKind so a missing kind for a known tool indicates a bug.
func LookupOutputTarget(tool types.ToolId, kind AssetKind) (OutputTarget, bool) {
	per := buildOutputMappings()
	tools, ok := per[tool]
	if !ok {
		return OutputTarget{}, false
	}
	target, ok := tools[kind]
	return target, ok
}

// OutputTargetsForTool returns the AssetKind→OutputTarget map for a tool, or
// an error when the tool is unknown.
func OutputTargetsForTool(tool types.ToolId) (map[AssetKind]OutputTarget, error) {
	per := buildOutputMappings()
	tools, ok := per[tool]
	if !ok {
		return nil, fmt.Errorf("output mapping: unknown tool %q", tool)
	}
	out := make(map[AssetKind]OutputTarget, len(tools))
	for k, v := range tools {
		out[k] = v
	}
	return out, nil
}

// ValidateOutputCoverage asserts every registered tool has an entry for every
// AssetKind in AllAssetKinds(). Used by tests to keep the table exhaustive.
func ValidateOutputCoverage() error {
	per := buildOutputMappings()
	for _, tool := range []types.ToolId{
		types.ToolIdClaudeCode, types.ToolIdOpenCode, types.ToolIdCopilot, types.ToolIdPi, types.ToolIdOmp, types.ToolIdKiro, types.ToolIdAntigravity,
	} {
		entries, ok := per[tool]
		if !ok {
			return fmt.Errorf("output mapping: tool %q has no entries", tool)
		}
		for _, kind := range AllAssetKinds() {
			t, has := entries[kind]
			if !has {
				return fmt.Errorf("output mapping: tool %q is missing asset kind %q", tool, kind)
			}
			if t.Shape != ShapeNone {
				if t.SourceSubdir == "" {
					return fmt.Errorf("output mapping: %q × %q has Shape=%q but empty SourceSubdir", tool, kind, t.Shape)
				}
				if t.DestSubdir == "" {
					return fmt.Errorf("output mapping: %q × %q has Shape=%q but empty DestSubdir", tool, kind, t.Shape)
				}
				if t.Shape == ShapeRewriteExt && t.RewriteSuffix == "" {
					return fmt.Errorf("output mapping: %q × %q uses ShapeRewriteExt but no RewriteSuffix", tool, kind)
				}
			}
		}
	}
	return nil
}
