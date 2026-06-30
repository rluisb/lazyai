# Plan: Antigravity Broad Tool-Capability Treatment

**Issue:** #575
**Epic:** #568
**Date:** 2026-06-29
**Status:** DRAFT — awaiting human gate
**Depends on:** #569 (canonical tool-capability model — implementation gate), #577 merged (foundation baseline)

---

## Scope

Broad tool-capability treatment for all Antigravity surfaces that support or mention tool usage. The decision is **not skills-only**: LazyAI must express tool capability (or the absence of it) across agents/subagents, skills, workflows, hooks, and commands. Commands are N/A for Antigravity; all other surfaces require work.

No code changes in this PR. This plan produces the task list for the implementation worktree gated behind human approval and #569.

---

## Sequencing within #568

```
#577 (merged) — tool-systems docs + matrix
    └─► #569 (P2-high, open) — canonical tool-capability model  ← BLOCKS implementation below
            └─► #575 (this) — Antigravity broad tool-capability treatment
```

All implementation tasks below are **blocked by #569**. Shape and library asset design may proceed in parallel; no Go code that reads canonical frontmatter can land before #569.

---

## Tasks

### Task 1 — Subagents: rules-based blueprint emission

**Surface:** Agents/Subagents  
**Blocker:** #569 (enable-flag values derived from canonical `tools`/`readonly` field)  
**Not blocked:** File format design, library asset authoring

**What:** Antigravity has no static agent definition file format (no `.agents/agents/` discovery path). The `define_subagent` call is runtime-programmatic. The adapter will emit a rules file that functions as a **subagent capability blueprint** — a structured instruction to the orchestrator covering how to invoke each canonical role with correct enable flags.

**Library asset to add:** `packages/cli/library/antigravity/subagents-blueprint.md`

Minimal content shape:

```markdown
# Subagent Capability Blueprint

When orchestrating subagents for LazyAI canonical roles, apply the following
`define_subagent` enable flags:

| Role | enable_write_tools | enable_mcp_tools | enable_subagent_tools |
|---|---|---|---|
| researcher | false | false | false |
| reviewer | false | false | false |
| evidence-verifier | false | false | false |
| implementer | true | true | false |
| deployer | true | true | false |
| planner | true | true | true |
| guide | true | true | true |
| responder | true | true | true |

read_url(*) and command(*) permissions follow the permissions engine defaults;
restrict via `.agents/rules/lazyai.md` policy blocks as needed.
```

**Adapter change:** `antigravity.go` Install() emits the blueprint to `.agents/rules/lazyai-subagents.md` (workspace) — skipped at global scope (rules are user-managed globally). Track as a managed file.

**Acceptance:**
- `.agents/rules/lazyai-subagents.md` emitted at workspace/project scope.
- Blueprint table correctly reflects `readonly: true` → `enable_write_tools: false` after #569 lands.
- Install test covers the new file.

---

### Task 2 — Skills: tool-capability annotation

**Surface:** Skills  
**Blocker:** None (documentation/description change only)  
**Scope:** Narrow

**What:** Antigravity skill files (`SKILL.md`) have no tool-restriction frontmatter — capability is enforced by the surrounding agent's `define_subagent` flags, not the skill itself. No structural change is needed. The action is:

1. Ensure each canonical skill's `description` field clearly states whether it is read-only or write-capable.
2. Add a `# Tool Access` section to read-only skills (researcher, reviewer, evidence-verifier families) explicitly noting they do not require `enable_write_tools`.

This is a **library asset quality improvement**, not an adapter code change.

**Files:**
- Audit `packages/cli/library/skills/*.md` for descriptions that imply write access without flagging it.
- `packages/cli/library/skills/research.md`, `packages/cli/library/skills/review.md`, `packages/cli/library/skills/diagnose.md` are likely candidates.

**Acceptance:**
- All read-only skills have descriptions that do not imply write access.
- No structural frontmatter change (Antigravity SKILL.md format unchanged).

---

### Task 3 — Hooks: write-tool matchers and PreInvocation slot

**Surface:** Hooks  
**Blocker:** None for hook expansion; #569 for per-agent capability gating  
**Priority:** Can design and add matchers now; agent-identity gating is future work

**What:** Expand `antigravity/hooks.json` and `antigravity/settings.json` to cover write-tier tools and provide a `PreInvocation` slot.

**Current emission:**
- `PreToolUse` → `run_command` (block-destructive-shell)
- `Stop` (objective-workflow-gate)
- `settings.json`: `BeforeTool`/`AfterAgent` equivalents

**Additions:**

1. **Write-tool PreToolUse matcher** (`hooks.json`):
```json
"lazyai-write-guard": {
  "PreToolUse": [
    {
      "matcher": "write_to_file|replace_file_content|multi_replace_file_content",
      "hooks": [
        {
          "type": "command",
          "command": ".gemini/hooks/lazyai/write-guard.sh",
          "timeout": 10
        }
      ]
    }
  ]
}
```
New hook script: `.gemini/hooks/lazyai/write-guard.sh` — validates write is intentional (exits 0 to allow, non-zero to deny).

2. **PreInvocation slot** (`hooks.json`): Reserved as a no-op handler for workflow step injection (Task 4):
```json
"lazyai-workflow-gate": {
  "PreInvocation": [
    { "type": "command", "command": ".gemini/hooks/lazyai/workflow-gate.sh", "timeout": 10 }
  ]
}
```

3. **settings.json alignment**: Add matching `BeforeTool` entry with `write_to_file|replace_file_content` matcher (Gemini CLI uses `BeforeTool` not `PreToolUse`).

**Note on agent-identity gap:** Hook matchers in Antigravity have no agent-name dimension — `PreToolUse` fires for all agents. Per-agent write gating is enforced by omitting `enable_write_tools` in `define_subagent` (Task 1), not by hooks. The write-guard hook is a global safety net, not a per-agent policy.

**Acceptance:**
- `hooks.json` template contains write-tool matcher and PreInvocation slot.
- `settings.json` template has corresponding `BeforeTool` entry.
- New hook scripts committed to `antigravity/hooks/lazyai/`.
- Existing install test updated to cover new tracked files.

---

### Task 4 — Workflows: emit library workflows as orchestrator skills

**Surface:** Workflows  
**Blocker:** None for asset creation; no Antigravity-native workflow file format exists  
**Dependency:** Task 3 for PreInvocation hook wiring

**What:** Antigravity has no static workflow surface. Library workflows (`packages/cli/library/workflows/*.md`) are emitted as orchestrator skills under a `workflow-` prefix.

**Emission path:**
```
packages/cli/library/workflows/<name>.md
    → .agents/skills/workflow-<name>/SKILL.md  (workspace)
    → ~/.gemini/config/skills/workflow-<name>/SKILL.md  (global)
```

The workflow SKILL.md frontmatter:
```yaml
---
name: workflow-<name>
description: >
  LazyAI orchestration workflow: <original description>. Uses write tools,
  shell execution, and may spawn subagents. Invoke with enable_write_tools=true
  and enable_subagent_tools=true.
---
```

Tool-access annotation in description is the only mechanism available (no `tools:` frontmatter in Antigravity SKILL.md). This enables the orchestrator to select correct `define_subagent` flags when invoking a workflow-executing subagent.

**Adapter change:** `antigravity.go` Install() adds a `CopyLibraryDirectory` call for `workflows/` → `skillsDir` with `workflow-` prefix in the destination path function.

**Acceptance:**
- Library workflows emitted as `workflow-<name>/SKILL.md` alongside capability skills.
- Workflow skills' descriptions include tool-access annotations.
- Install test covers workflow skills presence.

---

### Task 5 — Commands: document as N/A

**Surface:** Commands  
**No code change.**

**What:** Antigravity has no commands surface (no `.agents/commands/` discovery, no quickaction manifest field, no slash-command registration in `plugin.json`). Document this explicitly:

1. Update `docs/ai-cli-tools/tool-systems/antigravity.md` → "Readiness notes" section: add "No commands surface — Antigravity does not support slash-commands or quickactions. Closest analog is orchestrator skills."
2. Update `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` → add a Commands row or column noting Antigravity as `—`.

**Acceptance:**
- `antigravity.md` readiness notes state N/A for commands.
- Matrix updated.

---

### Task 6 — Matrix and docs update

**Surface:** Documentation  
**Blocker:** None

**What:** Update the gap-status column in `agent-tools-matrix.md` for Antigravity from "decide & document subagent stance" to reflect the planned broad treatment:

```markdown
| Antigravity | no agent/subagent files (skills-only) | n/a | **planned**: subagent blueprint (rules), workflow skills, hook expansion (#575) |
```

Update `docs/ai-cli-tools/antigravity.md` generated-structure diagram to include the new emission paths:

```text
.
├── AGENTS.md
├── .gemini/
│   ├── settings.json
│   └── hooks/lazyai/<hook>.sh
└── .agents/
    ├── hooks.json
    ├── rules/lazyai.md
    ├── rules/lazyai-subagents.md     ← new (Task 1)
    └── skills/
        ├── <skill>/SKILL.md
        └── workflow-<name>/SKILL.md  ← new (Task 4)

~/.gemini/config/mcp_config.json
```

**Acceptance:**
- Matrix row updated.
- `antigravity.md` generated-structure block updated.

---

## Verification Plan

After #569 merges and implementation is complete:

1. **`lazyai-cli init --scope project --tools antigravity --no-interactive`** — verify all new files are created.
2. **Inspect `.agents/rules/lazyai-subagents.md`** — verify enable-flag table is populated from canonical agents.
3. **Inspect `.agents/skills/workflow-*/SKILL.md`** — verify workflow skills present with tool-access annotations.
4. **Inspect `.agents/hooks.json`** — verify write-tool matcher and PreInvocation slot present.
5. **`go test ./packages/cli/internal/adapter/ -run TestAntigravity`** — all install tests pass.
6. **`mkdocs build --strict`** — docs build green.

---

## Open Questions (from research)

1. **Subagent blueprint vs. Python scaffold**: Rules-blueprint (recommended) is simpler and within the adapter's write pattern. SDK scaffold files are out of scope; revisit if Antigravity adds a static agent-definition format.
2. **write-guard hook behavior**: Should it be a soft-warn (log and allow) or a hard-deny (block)? Recommend hard-deny to match `block-destructive-shell` precedent.
3. **settings.json `BeforeTool` matcher syntax**: Verify pipe-separated matchers (`write_to_file|replace_file_content`) work in Gemini CLI OSS `BeforeTool` — may require separate entries.

---

## Human Gate

<!-- The human approver records approval here. Do NOT let an AI author this line. -->

Human Gate: APPROVED by rluisb at 2026-06-30T09:30:00-03:00

<!-- When approving, replace the line above with: Human Gate: APPROVED — [initials] [date] -->
