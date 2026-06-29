# Plan: Cross-CLI Agent-Tools Alignment

**Epic:** #568
**Date:** 2026-06-29
**Status:** awaiting human gate

## Scope of this PR (Foundation, docs-only)

Land the upstream tool-systems reference docs + the cross-CLI compatibility matrix
that every implementation issue (#569–#573) cites as the canonical capability mapping.

- `docs/ai-cli-tools/tool-systems/{index,kiro,claude-code,antigravity,pi,omp}.md` — per-target upstream tool-system reference (offline-readable, `verified_on` frontmatter + source URLs).
- `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` — canonical capability → per-target native model + casing + gap status.
- `mkdocs.yml` — "Tool Systems (upstream)" nav group.
- `specs/KNOWLEDGE_MAP.md` — Standards row pointing at the matrix.

No code changes. No adapter behavior change.

## Sequencing (whole epic, for context)

1. **Foundation (this PR):** reference docs + matrix.
2. **#569 (gate):** machine-readable tool-capability model on canonical agents + Go parser. Blocks #570–#573.
3. **Adapters (parallel, after #569):** #571 (copilot.go), #573 (omp.go), and #570→#572 sequenced (shared `agent_transform.go`).
4. **Reconcile:** #574 (Kiro verify + reconcile), #575 (Antigravity broad tool emission).

## Verification

- `mkdocs build --strict` green (CI `docs.yml`).
- Each PR carries its own `research.md` + `plan.md` for the RPI gate.

## Acceptance (this PR)

- Reference docs + matrix present and rendered by mkdocs strict build.
- mkdocs nav + KNOWLEDGE_MAP wired.
- No code or adapter changes.

## Human Gate

<!-- The human approver records approval here. Do NOT let an AI author this line. -->

Human Gate:APPROVED
