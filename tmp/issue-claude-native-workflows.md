## Summary

Evaluate and optionally implement native Claude Code workflow emission under `.claude/workflows/`.

Unlike most other LazyAI targets, Claude Code appears to have a documented native dynamic workflow surface. LazyAI currently treats workflow catalog assets as docs-only and does not emit `.claude/workflows/`.

## Current LazyAI state

Claude adapter installs:

- `.claude/settings.json`
- `.claude/rules/`
- `.claude/hooks/`
- `.claude/skills/`
- `.claude/agents/`
- `.claude/commands/`
- `.claude/output-styles/`

Evidence:

- `packages/cli/internal/adapter/claudecode.go`
- `packages/cli/internal/adapter/output_mapping.go:102-143`

No workflow support today:

- `packages/cli/internal/adapter/output_mapping.go:21-45` has no workflow asset kind.
- `packages/cli/cmd/create.go:92-103` does not allow `workflow`.
- `packages/cli/library/manifests/curation.yaml:1011-1098` marks workflows as `adapter_targets: [none]`.

## External docs to verify

Claude Code workflow docs mentioned during investigation:

- https://code.claude.com/docs/en/workflows.md
- https://code.claude.com/docs/en/commands.md
- https://code.claude.com/docs/en/skills.md
- https://code.claude.com/docs/en/hooks.md

Investigation finding:

- Claude Code workflows are scripts/pipelines that orchestrate subagents and can be saved/rerun from `.claude/workflows/`.
- Saved workflows may become slash commands.
- Skills, commands, and hooks remain distinct surfaces.

This must be source-verified before implementation because it is a runtime workflow surface, not just markdown guidance.

## Design options

### Option A — Keep docs-only

Continue to project workflow behavior via existing Claude-native surfaces:

- skills for reusable procedures
- commands for entrypoints
- hooks for objective gates
- rules/CLAUDE.md for always-on guidance
- templates for reusable prompts

### Option B — Add native Claude workflow support

Add a distinct capability and install path:

```go
Capability.Workflows bool
```

Add asset mapping:

```text
packages/cli/library/workflows/*.md -> .claude/workflows/*.md or required Claude workflow format
```

Only do this if Claude Code docs define the format and semantics clearly.

## Non-goals

- Do not make all tools emit workflow directories just because Claude supports one.
- Do not reintroduce LazyAI's retired workflow/task/orchestrator runtime.
- Do not claim markdown workflow docs are executable unless Claude Code actually executes them in that format.

## Acceptance criteria

- [ ] Source-verify Claude Code `.claude/workflows/` format and lifecycle.
- [ ] Decide Option A vs Option B with evidence.
- [ ] If Option B, add `Workflows` capability only for Claude Code.
- [ ] If Option B, add output mapping/install code and tests for `.claude/workflows/`.
- [ ] If Option B, add `validate` checks for Claude workflow files.
- [ ] Add docs explaining the difference between native Claude workflows and LazyAI docs-only workflow catalog.
- [ ] Add negative tests proving other targets do not receive fake workflow directories.
