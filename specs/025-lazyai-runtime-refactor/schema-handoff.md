# P0-4: Handoff Markdown Schema

**Status:** Complete — no-append write-on-close contract defined
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-13  
**Linked from:** `plan.md` Phase 0, P0-4

---

## Purpose

Contract for session handoff files written by the runtime to `specs/memory/handoffs/YYYY-MM-DD-[topic].md`.

## Frontmatter Keys

| Key | Required | Type | Description |
|---|---|---|---|
| `goal` | Yes | string | Current objective |
| `constraints` | Yes | string[] | Active constraints |
| `progress` | Yes | enum(done\|in-progress\|pending) | Overall status |
| `decisions` | Yes | string[] | Decisions made with rationale |
| `critical_context` | Yes | string | Context the next agent MUST know |
| `next_steps` | Yes | string[] | Next 1–2 concrete actions |
| `risks` | No | string[] | Watchouts for next agent |
| `owner` | No | string | Person/team responsible |
| `session_id` | No | string | Runtime session UUID |

## Required Sections

1. **Goal** — current objective
2. **Constraints & Preferences** — active constraints and user preferences
3. **Progress** — done / in-progress / pending items
4. **Key Decisions** — decisions made with rationale
5. **Critical Context** — facts the next agent MUST know
6. **Next Steps** — concrete next actions
7. **Open Assumptions/Questions** — unresolved items, if any
8. **Risks/Watchouts** — things that could go wrong, if any

## Path Conventions

- Directory: `specs/memory/handoffs/`
- Filename: `YYYY-MM-DD-[topic].md`
- Example: `specs/memory/handoffs/2026-06-14-lazyai-runtime-refactor.md`

## Ownership Model

- Writer and minimal reader/parser: `packages/cli/internal/handoff/writer.go`
- Round-trip tests: `packages/cli/internal/handoff/writer_test.go`
- Metadata: V2 `handoff` table stores session_id, path, created_at, status

## Round-Trip Expectations

- Write handoff on session close → read handoff → all frontmatter keys survive
- Repeated close write for the same session → atomic replace/update → no duplicate appended sections
- New session reads prior handoff → context is available

## Gate

⛔ Human must approve this schema before Phase 4 handoff implementation begins.
