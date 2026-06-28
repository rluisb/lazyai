# Migration: Fortnite / orchestrator removal

This refactor removes Fortnite-specific OpenCode defaults and the old orchestrator runtime path from the active LazyAI product surface, replacing them with the neutral canonical adapter path.

LazyAI owns the shipped runtime/product. vibe-lab supplies principles, assets, and adapter expectations only; it is not a runtime dependency or fallback runtime.

## Who is affected

You are affected if you relied on any of these behaviors:

- `lazyai-cli init --tools opencode` installing Fortnite agents, skills, workflows, and scripts by default
- `.opencode/opencode.jsonc` using `default_agent: loop-driver`
- generated `orchestrator` agent files as the neutral/default entry point
- `.opencode/STARTUP.md` bootstrap instructions
- `--plain-opencode` as an opt-out toggle

## What changed

| Old behavior | New behavior |
|---|---|
| OpenCode default install used Fortnite runtime assets | OpenCode default install uses the neutral canonical library |
| `loop-driver` was the OpenCode default agent | `guide` is the front-door default agent |
| `orchestrator` was the neutral/generated primary entry path | `guide` replaces it as the default entry point; `implementer` remains a specialist |
| `.opencode/STARTUP.md` was generated | Startup self-heal is handled by the generated OpenCode plugin without a STARTUP.md bootstrap file |
| `--plain-opencode` toggled between Fortnite and non-Fortnite installs | Removed; canonical install is now the only path |

## New default file layout

OpenCode projects now get:

- root `AGENTS.md`
- root `opencode.json`
- `.opencode/agents/guide.md`
- canonical specialist agents and skills selected by the neutral adapter contract

OpenCode projects no longer get Fortnite-only runtime content by default:

- `loop-driver`
- generated `orchestrator` agent entry files
- `.opencode/STARTUP.md`
- Fortnite workflows, scripts, bundled orchestration defaults, and obsolete eval/task/workflow runtime surfaces

## How to migrate an existing project

1. Update managed files:

   ```bash
   lazyai-cli update
   ```

2. Verify the generated OpenCode config now points at `implementer`:

   ```bash
   grep -n 'default_agent' .opencode/opencode.json
   ```

3. Remove any local automation that hard-codes `loop-driver`, `orchestrator`, or `.opencode/STARTUP.md`.

4. Run a health check after regeneration:

   ```bash
   lazyai-cli doctor
   ```

## Existing custom Fortnite assets

LazyAI removes managed legacy agent files during update when they still match the tracked library copy. User-edited or user-owned files are preserved. If you intentionally keep custom Fortnite-era assets, treat them as local customizations, not supported defaults.

For boundary categories and current command ownership, see [Product Boundaries](../concepts/product-boundaries.md).

## Rollback / pinning

`lazyai-cli update-self --version <tag>` now accepts exact release tags, including slash-containing tags such as:

- `packages/orchestrator/v1.0.0`
- `pre-refactor-025-phase-2`

Examples:

```bash
lazyai-cli update-self --version packages/orchestrator/v1.0.0
lazyai-cli update-self --version pre-refactor-025-phase-2
```

## Scope of this note

This note covers migration away from retired Fortnite/orchestrator defaults. Current supported setup and runtime-adjacent surfaces are listed in [Product Boundaries](../concepts/product-boundaries.md) and [CLI Reference](../cli/reference.md); archived orchestrator/eval/taskqueue material remains historical context unless a future issue explicitly reclassifies it.
