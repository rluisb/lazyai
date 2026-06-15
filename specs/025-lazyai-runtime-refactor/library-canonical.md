# P0-5: Canonical Library Specification

**Status:** Complete — target inventory defined; implementation pending
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-5

---

## Purpose

Inventory of every agent, skill, hook, and command that survives the refactor into `packages/cli/library/canonical/`. Must name surviving agents (including `primary-agent`) so Phase 1 adapter test fixtures can reference canonical names before files are populated in Phase 5.


Current code status: only `packages/cli/library/canonical/agents/primary-agent.md` is materialized. `packages/cli/library/embed.go` still embeds `all:fortnite`; canonical embedding and the remaining inventory are Phase 5 work.

## Surviving Inventory

### Agents

| Agent | Path | Justification |
|---|---|---|
| `primary-agent` | `agents/primary-agent.md` | Replaces orchestrator; redesigned primary-agent path per Clarify Q1. Single entry point for non-Fortnite adapters. |
| `builder` | `agents/builder.md` | General-purpose implementation agent; used across adapters |
| `planner` | `agents/planner.md` | Planning agent; used in Claude Code and Copilot workflows |
| `reviewer` | `agents/reviewer.md` | Code review agent; used across adapters |
| `scout` | `agents/scout.md` | Research/exploration agent; used in OpenCode workflows |

### Skills

| Skill | Path | Justification |
|---|---|---|
| `codebase-exploration` | `skills/codebase-exploration.md` | Used by scout agent across adapters |
| `test-first-change` | `skills/test-first-change.md` | TDD workflow; adapter-agnostic |
| `diagnose` | `skills/diagnose.md` | Debugging skill; adapter-agnostic |
| `pr-review` | `skills/pr-review.md` | PR review workflow; adapter-agnostic |

### Hooks

| Hook | Path | Justification |
|---|---|---|
| `session-start` | `hooks/session-start.md` | Session initialization; adapter-agnostic |
| `pre-commit` | `hooks/pre-commit.md` | Pre-commit validation; adapter-agnostic |

### Commands

| Command | Path | Justification |
|---|---|---|
| `graphify` | `commands/graphify.md` | Knowledge graph generation; adapter-agnostic |
| `handoff` | `commands/handoff.md` | Session handoff; new for Phase 4 |

## Excluded Items (Fortnite/Orchestrator-Specific)

| Item | Reason |
|---|---|
| `loop-driver` agent | Fortnite-specific default agent |
| `orchestrator` agent | Removed; replaced by `primary-agent` |
| `implementor` agent | Fortnite-specific |
| `documenter` agent | Fortnite-specific |
| `red-team` agent | Fortnite-specific |
| All `fortnite/` library content | Fortnite-specific; archived to `archive/fortnite-2026-06/` |
| All `orchestration/` library content | Orchestrator-dependent |
| `STARTUP.md` | Fortnite-specific instructions file |
| Fortnite workflows/chains/teams | Orchestrator-dependent |

## Test-First Requirement

P0-5 names the surviving agents (including `primary-agent`) and their expected tool sets so that Phase 1 adapter test fixtures can reference canonical agent names before the library files are physically populated in Phase 5.

**Canonical agent names for Phase 1 fixtures:**
- `primary-agent` (replaces `orchestrator`/`loop-driver`)
- `builder`
- `planner`
- `reviewer`
- `scout`

## Directory Structure

```
packages/cli/library/canonical/
├── agents/
│   ├── primary-agent.md
│   ├── builder.md
│   ├── planner.md
│   ├── reviewer.md
│   └── scout.md
├── skills/
│   ├── codebase-exploration.md
│   ├── test-first-change.md
│   ├── diagnose.md
│   └── pr-review.md
├── hooks/
│   ├── session-start.md
│   └── pre-commit.md
└── commands/
    ├── graphify.md
    └── handoff.md
```

## Gate

⛔ Human must approve this inventory before Phase 5 library curation begins.
