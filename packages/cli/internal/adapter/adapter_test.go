package adapter

import (
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// createTestFS creates a memo FS with the minimum files needed for adapter tests.
func createTestFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/implementer.md": &fstest.MapFile{
			Data: []byte("---\nname: Implementer\ndescription: Test implementer agent.\nmodel: sonnet\n---\n\n# Implementer\n\nYou are an implementer."),
		},
		"agents/researcher.md": &fstest.MapFile{
			Data: []byte("---\nname: Researcher\ndescription: Test researcher agent.\nmodel: haiku\n---\n\n# Researcher\n\nYou are a researcher."),
		},
		"canonical/agents/guide.md": &fstest.MapFile{
			Data: canonicalAgentFixture("guide", "Test guide agent."),
		},
		"canonical/agents/implementer.md": &fstest.MapFile{
			Data: canonicalAgentFixture("implementer", "Test implementer agent."),
		},
		"canonical/agents/researcher.md": &fstest.MapFile{
			Data: canonicalAgentFixture("researcher", "Test researcher agent."),
		},
		"canonical/agents/deployer.md": &fstest.MapFile{
			Data: canonicalAgentFixture("deployer", "Test deployer agent."),
		},
		"canonical/agents/responder.md": &fstest.MapFile{
			Data: canonicalAgentFixture("responder", "Test responder agent."),
		},
		"canonical/agents/planner.md": &fstest.MapFile{
			Data: canonicalAgentFixture("planner", "Test planner agent."),
		},
		"canonical/agents/reviewer.md": &fstest.MapFile{
			Data: canonicalAgentFixture("reviewer", "Test reviewer agent."),
		},
		"canonical/agents/evidence-verifier.md": &fstest.MapFile{
			Data: canonicalAgentFixture("evidence-verifier", "Test evidence verifier agent."),
		},
		"agents/extra.md": &fstest.MapFile{
			Data: []byte("---\nname: Extra\ndescription: Test unselected agent.\nmodel: opus\n---\n\n# Extra\n\nYou are not selected by default."),
		},
		"skills/codebase-exploration.md": &fstest.MapFile{
			Data: canonicalSkillFixture("codebase-exploration", "Explore code paths."),
		},
		"skills/test-first-change.md": &fstest.MapFile{
			Data: canonicalSkillFixture("test-first-change", "Drive changes through tests."),
		},
		"skills/diagnose.md": &fstest.MapFile{
			Data: canonicalSkillFixture("diagnose", "Diagnose failures."),
		},
		"skills/issue-triage.md": &fstest.MapFile{
			Data: canonicalSkillFixture("issue-triage", "Triage issues."),
		},
		"tool-agents/agents-dir.md": &fstest.MapFile{
			Data: []byte("# Agents Directory\n\nThis directory contains agent definitions."),
		},
		"tool-agents/skills-dir.md": &fstest.MapFile{
			Data: []byte("# Skills Directory\n\nThis directory contains skill definitions."),
		},
		"tool-agents/root-dir.md": &fstest.MapFile{
			Data: []byte("# Root Directory\n\nProject context at root level."),
		},
		"root/AGENTS.template.md": &fstest.MapFile{
			Data: []byte("# AGENTS\n\n{{PROJECT_NAME}} project agents."),
		},
		"root/copilot-instructions.template.md": &fstest.MapFile{
			Data: []byte("# Copilot Instructions\n\nUse these instructions with Copilot."),
		},
		"prompts/preflight-task-framing.md": &fstest.MapFile{
			Data: []byte("---\nname: preflight-task-framing\n---\n\n# Task Framing\n\nFrame tasks before starting."),
		},
		"prompts/plan.md": &fstest.MapFile{
			Data: []byte("# Plan\n"),
		},
		"prompts/research.md": &fstest.MapFile{
			Data: []byte("# Research\n"),
		},
		"rules/typescript.md": &fstest.MapFile{
			Data: []byte("---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n\n- Use strict TypeScript\n- Prefer interfaces over types for objects\n"),
		},
		"pi/extensions/block-destructive-shell.ts": &fstest.MapFile{
			Data: []byte("import type { ExtensionAPI } from \"@earendil-works/pi-coding-agent\";\n\nexport default function blockDestructiveShell(pi: ExtensionAPI): void {\n  pi.on(\"tool_call\", async (event, ctx) => {\n    if (event.toolName !== \"bash\") return;\n    const cmd = String(event.input.command ?? \"\");\n    if (!/\\brm\\s+-rf\\s+\\//.test(cmd)) return;\n    if (ctx.hasUI) {\n      const allow = await ctx.ui.confirm(\"Dangerous command\", `This deletes from root:\\n${cmd}\\n\\nProceed?`);\n      if (allow) return;\n    }\n    return { block: true, reason: \"rm -rf / blocked by safety policy\" };\n  });\n}\n"),
		},
		"pi/extensions/extension-dir/index.ts": &fstest.MapFile{
			Data: []byte("import type { ExtensionAPI } from \"@earendil-works/pi-coding-agent\";\n\nexport default function (pi: ExtensionAPI): void {\n  pi.on(\"session_start\", async (_e, ctx) => ctx.ui.notify(\"loaded\", \"info\"));\n}\n"),
		},
		"pi/extensions/extension-dir/helper.ts": &fstest.MapFile{
			Data: []byte("export function helper(): string { return \"hi\"; }\n"),
		},
		"pi/extensions/extension-dir/package.json": &fstest.MapFile{
			Data: []byte("{\n  \"name\": \"extension-dir\",\n  \"version\": \"1.0.0\",\n  \"dependencies\": {}\n}\n"),
		},
		"pi/SYSTEM.md": &fstest.MapFile{
			Data: []byte("# Pi System Prompt\n\nProject-specific system prompt that replaces the default."),
		},
		"pi/APPEND_SYSTEM.md": &fstest.MapFile{
			Data: []byte("# Pi Appended System Prompt\n\nAppended to the default system prompt."),
		},
		"commands/rpi.toml": &fstest.MapFile{
			Data: []byte("name = \"rpi\"\ndescription = \"Start RPI\"\nprompt = \"Begin RPI\"\n"),
		},
		"commands/review.toml": &fstest.MapFile{
			Data: []byte("name = \"review\"\ndescription = \"Review work\"\nprompt = \"Do review\"\n"),
		},
		"commands/plan.toml": &fstest.MapFile{
			Data: []byte("name = \"plan\"\ndescription = \"Plan work\"\nprompt = \"Make plan\"\n"),
		},
		"chatmodes/architect.agent.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Architect mode\ntools: ['codebase']\n---\nArchitect instructions."),
		},
		"chatmodes/reviewer.agent.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Reviewer mode\ntools: ['search']\n---\nReviewer instructions."),
		},
		"opencode/commands/review.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Review branch\n---\n\nReview body."),
		},
		"opencode/commands/test.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Run tests\n---\n\nTest body."),
		},
		"opencode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Draft commit\n---\n\nCommit body."),
		},
		"opencode/modes/plan.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Plan mode\ntools:\n  write: false\n  read: true\n---\n\nPlan body."),
		},
		"opencode/modes/audit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Audit mode\ntools:\n  write: false\n  read: true\n---\n\nAudit body."),
		},
		"claudecode/commands/review.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Review changes\nargument-hint: \"[pr]\"\nallowed-tools: Bash Read\n---\n\nReview body."),
		},
		"claudecode/commands/test.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Run tests\nargument-hint: \"[target]\"\nallowed-tools: Bash Read\n---\n\nTest body."),
		},
		"claudecode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Draft commit\nargument-hint: \"\"\nallowed-tools: Bash Read\n---\n\nCommit body."),
		},
		"claudecode/output-styles/terse.md": &fstest.MapFile{
			Data: []byte("---\nname: Terse\ndescription: Short responses\nkeep-coding-instructions: true\n---\n\nTerse style body."),
		},
		"claudecode/output-styles/explanatory.md": &fstest.MapFile{
			Data: []byte("---\nname: Explanatory\ndescription: Detailed responses\nkeep-coding-instructions: true\n---\n\nExplanatory style body."),
		},
		"claudecode/hooks/block-destructive-shell.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"claudecode/hooks/objective-workflow-gate.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"copilot/hooks/block-destructive-shell.json": &fstest.MapFile{
			Data: []byte("{\"version\":1}"),
		},
		"copilot/hooks/block-destructive-shell.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"copilot/hooks/objective-workflow-gate.json": &fstest.MapFile{
			Data: []byte("{\"version\":1}"),
		},
		"copilot/hooks/objective-workflow-gate.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"opencode/plugins/vibe-lab-hooks.js": &fstest.MapFile{
			Data: []byte("export const VibeLabHooks = () => ({})\n"),
		},
		"antigravity/settings.json": &fstest.MapFile{
			Data: []byte("{\"hooks\":{}}\n"),
		},
		"antigravity/hooks/lazyai/block-destructive-shell.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"antigravity/hooks/lazyai/objective-workflow-gate.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
		},
		"antigravity/hooks.json": &fstest.MapFile{
			Data: []byte(`{
  "lazyai-block-destructive-shell": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "type": "command",
            "command": ".gemini/hooks/lazyai/block-destructive-shell.sh",
            "timeout": 10
          }
        ]
      }
    ]
  },
  "lazyai-objective-workflow-gate": {
    "Stop": [
      {
        "type": "command",
        "command": ".gemini/hooks/lazyai/objective-workflow-gate.sh"
      }
    ]
  }
}`),
		},
		"kiro/hooks/block-destructive-shell.json": &fstest.MapFile{
			Data: []byte(`{
  "version": "v1",
  "hooks": [
    {
      "name": "block-destructive-shell",
      "description": "Blocks destructive shell commands before they execute.",
      "trigger": "PreToolUse",
      "matcher": "shell",
      "action": {
        "type": "command",
        "command": ".kiro/hooks/block-destructive-shell.sh"
      },
      "timeout": 10,
      "enabled": true
    }
  ]
}`),
		},
		"kiro/hooks/block-destructive-shell.sh": &fstest.MapFile{
			Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
			Mode: 0o755,
		},
	}
}

// createTestAdapterContext creates an AdapterContext for testing with a temp target dir.
func createTestAdapterContext(t *testing.T) (*AdapterContext, string) {
	t.Helper()
	targetDir := t.TempDir()
	libFS := createTestFS()

	return &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: "", // empty = production mode, use LibraryFS
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"researcher"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}, targetDir
}
