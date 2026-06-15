# Migration: Fortnite / orchestrator removal

This refactor removes Fortnite-specific OpenCode defaults and replaces the old orchestrator entry path with the neutral canonical adapter path.

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
| `loop-driver` was the OpenCode default agent | `primary-agent` is the default agent |
| `orchestrator` was the neutral/generated primary entry path | `primary-agent` replaces it |
| `.opencode/STARTUP.md` was generated | No startup bootstrap file is generated |
| `--plain-opencode` toggled between Fortnite and non-Fortnite installs | Removed; canonical install is now the only path |

## New default file layout

OpenCode projects now get:

- root `AGENTS.md`
- `.opencode/opencode.jsonc` with `default_agent: primary-agent`
- `.opencode/agents/primary-agent.md`
- canonical agent/skill content selected by the neutral adapter contract

OpenCode projects no longer get Fortnite-only runtime content by default:

- `loop-driver`
- generated `orchestrator` agent entry files
- `.opencode/STARTUP.md`
- Fortnite workflows, scripts, and bundled orchestration defaults

## How to migrate an existing project

1. Update managed files:

   ```bash
   lazyai-cli update
   ```

2. Verify the generated OpenCode config now points at `primary-agent`:

   ```bash
   grep -n 'default_agent' .opencode/opencode.jsonc
   ```

3. Remove any local automation that hard-codes `loop-driver`, `orchestrator`, or `.opencode/STARTUP.md`.

4. Run a health check after regeneration:

   ```bash
   lazyai-cli doctor
   ```

## Existing custom Fortnite assets

LazyAI removes managed legacy agent files during update when they still match the tracked library copy. User-edited or user-owned files are preserved. If you intentionally keep custom Fortnite-era assets, treat them as local customizations, not supported defaults.

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

This note covers the adapter/default-path migration. Later phases of Spec 025 will still remove or rewrite CLI command surfaces that depend on heavy orchestration packages. Those removals must land with separate verification and rollback records.
