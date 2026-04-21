# 004 — Orchestrator Watch & Catalog DB: Research

## Date
2026-04-20

## Method

Two parallel exploration agents:
1. Mapped `orchestrator/src/` file-by-file, surfacing extension points, persistence boundaries, and test contracts.
2. Audited host CLI configuration directories (opencode, claude-code, codex) for how each stores agents/skills/commands/modes and whether they support project-level overrides.

## Findings — orchestrator internals

### Persistence is already cleanly seamed

`orchestrator/src/persistence.ts` exposes ~16 functions (`saveChainState`, `loadChainState`, etc.) that are the only file-I/O in the system. State machines (`chain-machine.ts`, `team-machine.ts`, `workflow-machine.ts`) are pure: they accept state in, return new state out, and the caller persists. A SQLite swap is a drop-in replacement of these 16 function bodies; nothing else needs to change.

### Catalog is loaded once at startup

`loader.ts` reads agents (`.md` with YAML frontmatter), skills (directory with `SKILL.md`), and chains/teams/workflows (JSON) from a configurable library root and a project override root. Project entries shadow library entries by name. The "library + project override" pattern already exists; the DB becomes a third resolution layer (or first, depending on phase 3 priority order).

### State shapes are JSON-native

`ChainState`, `TeamState`, `WorkflowState`, `ExecutionPlan`, and `BudgetState` contain deeply nested arrays (`steps[]`, `tasks[]`, `phases[]`, `byStep` map). Normalizing them into relational tables would create dozens of tables for marginal benefit. Decision: store nested structures as `JSON` columns with indexes on hot lookup fields (`state`, `updated_at`, `definition_name`).

### State machines don't emit events

The machines mutate state and return; they don't notify anyone. Adding watch/subscribe means inserting an emit point at every transition (or wrapping the machine functions). Phase 5 inserts these emit points; phase 0–4 leave the machines alone.

### Test surface

`orchestrator/src/__tests__/*.case.ts` covers loader merge semantics, persistence round-trips, chain/team/workflow state machine transitions, budget evaluation, error-journal append/read, and tool-handler entry points. Treat all of these as contracts. Specific invariants to preserve:

- All IDs are UUIDs.
- All timestamps are ISO 8601 strings.
- Gates have three states: `pending`, `approved`, `rejected`.
- Retry transitions are objects: `{ retry: number, then: string }`.
- `byStep` keys are step IDs (chains), task IDs (teams), or `<childId>:<stepId>` (workflows).
- `ErrorJournalEntry` rows are append-only — never updated.

## Findings — host CLI configuration

### opencode (`~/.config/opencode/`)

```
agents/         single .md files, name = filename
skills/         directory per skill containing SKILL.md
commands/       single .md files
modes/          single .md files
opencode.jsonc  global config; no project-override path declared
```

Frontmatter (agent example):

```yaml
name: architect
model: ollama-cloud/gemma4:31b
description: ...
mode: subagent
temperature: 0.2
steps: 18
tools: { write: true, bash: false }
permissions: { write: { allow: ["bee-gone/specs/*"] } }
```

opencode reads on startup; no file-watcher. **No project-level catalog directory is searched** — global only.

### claude-code (`~/.claude/` + `<project>/.claude/`)

```
~/.claude/skills/   directory per skill containing SKILL.md (often symlinked to ~/.agents/skills/)
~/.claude/agents/   not always present at user level
<project>/.claude/  full parallel structure: agents/, skills/, commands/, settings.json, settings.local.json, CLAUDE.md
```

Frontmatter (agent):

```yaml
name: Builder
model: sonnet
```

Frontmatter (skill):

```yaml
name: implement
trigger: /implement
phase: implement
```

claude-code is the **only host with a real project-override pattern**: project `.claude/` shadows user `~/.claude/`. Settings merge via `settings.json` + `settings.local.json`.

### codex (`~/.codex/`)

```
~/.codex/state_5.sqlite          state DB (not user-editable as files)
~/.codex/skills/.system/         read-only system skills (skill-creator, plugin-creator)
~/.codex/config.toml             trust-level config; no project-path redirection
```

codex skills/agents are not file-resolvable in the same way — its catalog is internal to its SQLite. **The orchestrator cannot resolve codex-owned definitions from disk**; treat as opaque, fall through to internal DB catalog.

### Resolution priority (synthesized from above)

For a step that says `agent: reviewer`:

1. **Explicit version pin** (`reviewer@7` or `reviewer/internal@7`) → DB lookup.
2. **Project host override** (only meaningful for claude-code today): `<project>/.claude/agents/reviewer.md`.
3. **User-global host override**: `~/.config/opencode/agents/reviewer.md`, `~/.claude/agents/reviewer.md`.
4. **Internal DB catalog** at active version.

Codex steps skip steps 2 and 3 (not file-resolvable).

## Decisions captured (from spec.md, recorded here for traceability)

| Decision | Choice | Rationale |
|---|---|---|
| DB location | `~/.local/share/ai-setup-orchestrator/orchestrator.db` (cross-project) | Required for cross-client features (watch, agent registry) |
| Versioning scheme | Monotonic integer per `(kind, name)` | Internal artifacts; semver overhead unwarranted |
| HTTP transport | Unix socket (mode 0600) | No port collisions, simpler than bearer tokens |
| Frontmatter validation | Strict per-kind Zod schemas | Matches per-host shapes documented above |
| Spec format | `specs/004-.../{spec,plan,research}.md` | Matches repo convention from 002 and 003 |

## Risks identified

- **Hundreds of host agents** could make scanning slow. Mitigation: mtime-based memoization per directory.
- **Concurrent SQLite writes** under HTTP transport. Mitigation: single writer thread + WAL mode; revisit at phase 4.
- **Existing on-disk state** must not be lost during phase 2 cutover. Mitigation: importer reads JSON files into DB rows; original files retained until user opts to delete.
- **Codex opacity** means features that depend on host-resolved agents won't work for codex users. Mitigation: codex always falls through to internal catalog; documented behavior, not a bug.

## References

- `orchestrator/src/persistence.ts` — current persistence layer (the seam being replaced)
- `orchestrator/src/loader.ts` — current catalog loader (extended in phase 3)
- `orchestrator/src/__tests__/` — test contracts to preserve
- `~/.config/opencode/`, `~/.claude/`, `~/.codex/` — host config layouts audited above
- `specs/002-simplification-and-restructure/plan.md`, `specs/003-post-install-automation-and-integrations/research.md` — convention reference for this spec's structure
