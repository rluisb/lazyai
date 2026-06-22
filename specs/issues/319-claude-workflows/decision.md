# Issue 319 — Claude Code workflow emission decision

## Decision
Keep LazyAI's current workflow catalog docs-only for Claude Code. Do not emit `packages/cli/library/workflows/*.md` to `.claude/workflows/`.

## Evidence
Claude Code dynamic workflows are JavaScript scripts executed by Claude Code's workflow runtime, not markdown process documents:

- Claude Code docs describe a workflow as "a JavaScript script that orchestrates subagents at scale".
- Saved workflows live in `.claude/workflows/` or `~/.claude/workflows/` and become slash commands.
- Workflow runs are managed through `/workflows`; scripts coordinate agents but do not perform direct filesystem or shell access themselves.
- Claude Code requires v2.1.154+ and a paid/available workflow feature flag.

LazyAI's embedded workflow catalog is markdown:

- `packages/cli/library/workflows/*.md`
- `packages/cli/library/workflows/verified-research/templates/*.md`

Those files are workflow guidance/catalog content, not Claude workflow runtime scripts.

## Consequence
A future Claude workflow feature needs a separate source kind or generator that produces real Claude workflow JavaScript scripts. It must not copy markdown catalog files into `.claude/workflows/` and claim they are executable.

## Guardrail added
`packages/cli/internal/adapter/output_mapping_test.go` now asserts generic output mapping does not add `workflow` / `workflows` as an asset kind. This keeps the markdown catalog from being emitted as a fake native workflow directory.
