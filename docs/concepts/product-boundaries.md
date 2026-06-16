# Product boundaries

LazyAI is the runtime and product. It owns the `lazyai-cli` binary, the canonical `.ai/` source model, the adapter/compiler path that writes tool-native files, and the local state used by shipped runtime-adjacent commands. vibe-lab supplies principles, assets, and adapter expectations; it is not a runtime dependency of LazyAI.

This page is the source inventory for the active product boundary. It was built from the source registered under `packages/cli/cmd/` and the top-level packages under `packages/cli/internal/`.

## Categories

### `setup-core`

Included when a command, package, or asset directly creates, validates, compiles, updates, or explains the supported LazyAI setup surface: canonical `.ai/` files, manifests, embedded library assets, adapters, generated tool-native files, setup health, and workspace/scope resolution.

Excluded: optional operational state such as session records and metrics, repository-only maintenance scripts, deprecated aliases, retired orchestration/eval/taskqueue surfaces, and archived research or rollback material.

### `ops-runtime-extra`

Included when a shipped CLI command or package operates local runtime-adjacent state around a setup: sessions, ledger/audit records, memory, messages, metrics, costs, secrets, notifications, backups, and authentication probes. These commands still ship today, but they are transitional extras outside the default setup-core product boundary while explicit module lifecycle semantics are still pending.

Excluded: setup compilation itself, repo-maintainer scripts, generated-source maintenance, and retired orchestrator/eval/taskqueue command surfaces.

### `dev-harness`

Included when a file or package is for maintaining this repository or release artifact rather than being a normal LazyAI runtime dependency: repository scripts, generated catalog refreshes, token-rent enforcement, self-healing/injection helpers, and diff-review support tooling.

Excluded: shipped user setup/runtime commands unless their primary job is source/catalog maintenance, and archived historical material.

### `retired/archived`

Included when a command, package, asset, or document exists only for migration, rollback, historical context, or a deprecated hidden compatibility path.

Excluded: the active baseline-facing `implementer` canonical adapter path, current `lazyai-cli` setup/runtime commands, and active embedded library assets selected by the neutral adapter contract.

## Supported surfaces at a glance

| Surface | Status | Boundary |
|---|---|---|
| `lazyai-cli` top-level commands registered in `packages/cli/cmd/` | Shipped CLI | Categorized command-by-command below |
| `packages/cli/library/canonical/` plus active adapter-selected library content | Active embedded library assets | `setup-core` |
| `packages/cli/internal/*` packages used by the CLI | Source implementation | Categorized package-by-package below |
| `bin/doctor`, `bin/inject`, `bin/startup-self-heal` | Repository harness scripts, not shipped CLI commands | `dev-harness` |
| `archive/`, retired `docs/orchestrator-*`, and retired Fortnite/orchestrator/eval references | Historical or rollback material | `retired/archived` |

The CLI reference documents shipped `lazyai-cli` commands only. Repository scripts under `bin/` may be useful for maintainers, but they are not product commands and should not be presented as user-facing LazyAI runtime surface. Likewise, `ops-runtime-extra` commands are listed for completeness, but they are secondary/transitional surfaces rather than the default product headline.

### Default surface contract vs LazyAI extras

- Default, shipped setup outputs emit eight canonical agents across supported tool surfaces: `guide` as the front-door default plus `implementer`, `researcher`, `planner`, `reviewer`, `deployer`, `responder`, and `evidence-verifier` as specialists.
- OpenCode default setup writes a root `opencode.json` config with baseline shape only (schema, permissions, skill paths, instructions). It does not carry LazyAI runtime-only MCP content.
- LazyAI runtime-adjacent MCP/runtime extras are isolated to `.opencode/lazyai.mcp.jsonc` so the default `opencode.json` stays baseline-compatible and replaceable.
- Retired artifacts such as orchestrator/loop-driver/Startup surfaces are not part of default setup outputs and belong to explicit runtime/archival paths only.

## Top-level CLI command inventory

| Command | Category | Rationale |
|---|---|---|
| `add` | `setup-core` | Adds managed agents, skills, or tool configuration into an existing LazyAI setup. |
| `auth` | `ops-runtime-extra` | Inspects configured provider authentication and environment state around a setup. |
| `backup` | `ops-runtime-extra` | Backs up local LazyAI database, ledger, and configuration data. |
| `build-plugin` | `setup-core` | Emits a Claude Code plugin directory from the embedded LazyAI library for supported adapter distribution. |
| `compile` | `setup-core` | Regenerates per-tool MCP/config output from the canonical `.ai/` source. |
| `completion` | `setup-core` | Installs shell integration for the shipped `lazyai-cli` binary. |
| `completions` | `retired/archived` | Hidden deprecated alias retained only for compatibility; use `completion` for active docs and support. |
| `config` | `setup-core` | Reads and writes LazyAI setup configuration values used by the product. |
| `cost` | `ops-runtime-extra` | Reports runtime-adjacent model/cost analytics from local LazyAI state. |
| `create [type] [name]` | `setup-core` | Generates supported setup artifacts such as agents, skills, prompts, commands, and templates. |
| `doctor` | `setup-core` | Checks setup health, managed files, metadata, migrations, and environment drift. |
| `eject` | `setup-core` | Removes LazyAI library management while leaving generated files in place. |
| `git` | `ops-runtime-extra` | Provides optional source-control automation around a workspace, not setup compilation. |
| `import [source]` | `setup-core` | Imports configuration from another AI tool setup into LazyAI-managed form. |
| `info <name>` | `setup-core` | Explains a managed artifact from the embedded or installed setup inventory. |
| `init` | `setup-core` | Creates the canonical LazyAI setup and adapter-managed tool files. |
| `ledger` | `ops-runtime-extra` | Manages the local immutable audit trail for runtime-adjacent activity. |
| `list [category]` | `setup-core` | Lists installed or available managed setup artifacts. |
| `memory` | `ops-runtime-extra` | Operates the local long-term memory vault attached to LazyAI state. |
| `message` | `ops-runtime-extra` | Provides the SQLite-backed local agent message bus. |
| `metrics` | `ops-runtime-extra` | Lists, exports, and renders local runtime/quality metrics. |
| `migrate` | `setup-core` | Migrates older LazyAI or ai-setup configuration into the current supported format. |
| `models` | `dev-harness` | Refreshes generated model catalog source from models.dev for maintainers. |
| `notify` | `ops-runtime-extra` | Configures and sends optional local notifications around LazyAI operations. |
| `restore-runtime-db [backup-path]` | `ops-runtime-extra` | Restores local runtime database state from a backup. |
| `secret` | `ops-runtime-extra` | Stores, reads, lists, and removes local secrets used by operations. |
| `server` | `setup-core` | Manages MCP server entries that compile into supported tool-native configs. |
| `session` | `ops-runtime-extra` | Starts, lists, shows, and ends local AI agent sessions in SQLite state. |
| `setup` | `setup-core` | Inspects setup inventory and planning output for supported targets. |
| `sidecar` | `setup-core` | Configures optional sidecar locations for docs, specs, and plans used by the setup. |
| `status` | `setup-core` | Shows current manifest/setup state and selected tools. |
| `update` | `setup-core` | Refreshes managed files to match current embedded library versions. |
| `update-self` | `setup-core` | Updates the shipped CLI binary to a selected GitHub release. |
| `validate` | `setup-core` | Validates supported agent and skill file structures. |
| `workspace` | `setup-core` | Registers, selects, and reports workspace scope for setup resolution. |

Removed command surfaces such as `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` are not active top-level CLI commands. They belong in migration notes or archived documentation, not in the current CLI reference.

## Runtime-adjacent internal package inventory

| Package | Category | Rationale |
|---|---|---|
| `internal/adapter` | `setup-core` | Implements tool-specific output for OpenCode, Claude Code, Copilot, OMP/Pi, Antigravity, hook runtimes, MCP, and related adapter contracts. |
| `internal/auth` | `ops-runtime-extra` | Probes provider authentication state used by auth-aware setup and runtime-adjacent checks. |
| `internal/compiler` | `setup-core` | Compiles canonical fragments and validates agent contracts for generated setup output. |
| `internal/configmerge` | `setup-core` | Merges generated and existing config files while preserving supported user customizations. |
| `internal/conflict` | `setup-core` | Encodes overwrite/backup/conflict decisions for managed setup files. |
| `internal/db` | `ops-runtime-extra` | Provides SQLite storage and migrations for local operational state. |
| `internal/detect` | `setup-core` | Detects existing setup/tool state during initialization and update paths. |
| `internal/diff` | `dev-harness` | Builds diff structures consumed by review tooling rather than the runtime itself. |
| `internal/diffreview` | `dev-harness` | Defines the JSON review contract for the companion diffviewer workflow. |
| `internal/error` | `setup-core` | Standardizes command and boundary errors for the CLI implementation. |
| `internal/files` | `setup-core` | Provides filesystem primitives for generated and managed setup files. |
| `internal/frontmatter` | `setup-core` | Parses and validates frontmatter in agents, skills, and related library assets. |
| `internal/generator` | `setup-core` | Generates supported artifact files for the `create` command. |
| `internal/globalpaths` | `setup-core` | Resolves global and scoped LazyAI paths used by setup and adapters. |
| `internal/handoff` | `ops-runtime-extra` | Writes resumable local handoff documents from session/runtime state. |
| `internal/jsonc` | `setup-core` | Supports JSONC parsing/writing for tool-native config files. |
| `internal/library` | `setup-core` | Loads and validates the embedded library that powers setup and adapter output. |
| `internal/log` | `setup-core` | Provides shared CLI logging plumbing. |
| `internal/manifest` | `setup-core` | Reads and writes the managed setup manifest and file hashes. |
| `internal/migration` | `setup-core` | Detects and executes migrations from older LazyAI or ai-setup layouts. |
| `internal/models` | `setup-core` | Supplies the vendored model catalog and curated tiers used by adapters and validation. |
| `internal/plugin` | `setup-core` | Builds plugin output from the embedded library. |
| `internal/preset` | `setup-core` | Maps setup presets to supported features, rules, and selected artifacts. |
| `internal/reversa` | `setup-core` | Provides optional source analysis used to prefill setup context. |
| `internal/runtime` | `ops-runtime-extra` | Owns local session, ledger, and runtime database primitives used by shipped operational commands. |
| `internal/scaffold` | `setup-core` | Writes canonical files, root instructions, tool configs, and library-backed setup content. |
| `internal/setup` | `setup-core` | Builds setup context shared by add/update flows. |
| `internal/setupscan` | `setup-core` | Scans and adopts existing setup files into LazyAI-managed state. |
| `internal/sidecar` | `setup-core` | Resolves optional sidecar paths for docs, specs, and plans. |
| `internal/theme` | `setup-core` | Provides shared CLI/TUI presentation tokens for supported commands. |
| `internal/tokenrent` | `dev-harness` | Enforces the canonical library byte budget for repository maintainers. |
| `internal/types` | `setup-core` | Defines shared product types for artifacts, tools, presets, and manifests. |
| `internal/validation` | `setup-core` | Validates active artifact metadata and setup file structure. |
| `internal/version` | `setup-core` | Carries build/version metadata for the shipped CLI. |

No active `packages/cli/internal/` package is categorized as `retired/archived` as a whole in this inventory. Retired functionality found inside active packages is listed below as follow-up work rather than moved or deleted in this documentation branch.

## Active embedded library versus archived material

Active embedded library assets are the files the CLI embeds and selects through the canonical setup and adapter contracts. The supported default path is the front-door `guide` contract plus current canonical library content, with `implementer` and the other specialist agents kept behind it. LazyAI may include historical or compatibility assets in the repository, but active adapters must not treat Fortnite, the old orchestrator runtime, or eval surfaces as the default runtime foundation.

Archived and retired material is still useful for migration, rollback, and historical review. It must be described as historical context and must not be copied into current setup guidance as if it were supported runtime behavior.

## Follow-up items found during inventory

These are documented follow-ups only; this issue does not move or delete code.

- `packages/cli/internal/runtime/schema.go` still contains workflow, task queue, and eval tables even though the corresponding top-level command surfaces are removed; a future schema/migration task should decide what remains for compatibility and what becomes archived state.
- `packages/cli/internal/db/migrations.go` still includes older task queue and dispatch migration tables; a future migration audit should decide whether those are required compatibility migrations or removable historical scaffolding.
- Adapter and setup-scan tests still exercise preserved orchestrator MCP entries for adoption/compatibility scenarios; future cleanup should decide whether those tests represent supported user-owned config preservation or retired runtime coupling.
- The embedded library still contains non-default historical agent/skill names such as `orchestrator`/`orchestrate`; the library curation/provenance work should decide which assets remain active, compatibility-only, or archived.
