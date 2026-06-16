# LazyAI

A setup engine for AI toolchains. `lazyai-cli` defines your AI assistant setup once in a canonical format, then compiles and updates it across every supported AI tool — OpenCode, Claude Code, GitHub Copilot, Pi, Antigravity, and friends.

`lazyai-cli` follows a **canonical source → compile** model:

1. `init` scaffolds a tool-agnostic canonical layer under `.ai/`
2. You edit rules, agents, and templates in one place
3. `compile` generates tool-native files (`.opencode/`, `.claude/`, `.github/`, `.vscode/`)
4. `update` refreshes managed files from the bundled library
5. `doctor` checks health and drift

**Setup-core is the default product.** `init`, `compile`, `update`, `doctor`, `add`, `build-plugin`, `validate`, and the workspace/sidecar commands are the engine. Runtime-adjacent commands (sessions, ledger, memory, auth, cost, metrics, notify, secret, backup, restore-runtime-db, git) still ship in the same binary today, but they are transitional extras outside setup-core on the path to explicit opt-in modules — see [ADR-005](specs/adrs/005-core-vs-optional-modules.md) and [Product Boundaries](docs/concepts/product-boundaries.md).

Learn more in [How It Works](docs/concepts/how-it-works.md).


---

## What is supported

- **Default product (setup-core):** the shipped `lazyai-cli` binary, canonical `.ai/` setup model, adapter/compiler output for OpenCode, Claude Code, GitHub Copilot, Pi, and Antigravity, and the active embedded library assets. See [ADR-005](specs/adrs/005-core-vs-optional-modules.md) for the framing.
- **Transitional runtime extras:** runtime-adjacent command families (sessions, ledger, memory, auth, cost, metrics, notify, secret, backup, restore-runtime-db, git) still ship today, but they are outside the default setup-core product boundary and are being staged toward explicit opt-in module semantics in a follow-up phase.
- **Not a runtime dependency:** vibe-lab supplies principles, assets, and adapter expectations that inform LazyAI, but LazyAI owns the Go runtime and product surface. See [ADR-004](specs/adrs/004-vibe-lab-alignment-contract.md) for the alignment contract.
- **Repository harness only:** scripts such as `bin/doctor`, `bin/inject`, and `bin/startup-self-heal` support maintainers of this repository; they are not shipped LazyAI CLI commands.
- **Retired/archived:** Fortnite defaults, the old orchestrator runtime, obsolete eval/task/workflow surfaces, and files under `archive/` are historical or migration material, not supported active runtime.

See [Product Boundaries](docs/concepts/product-boundaries.md) for the command and internal-package inventory.

---

## Commands

### Setup-core examples

Core setup commands are the default product surface:

```bash
lazyai-cli doctor
# → Checks file integrity, metadata, migrations, env dependencies, and setup drift

lazyai-cli validate skills
# → Checks skill structure and common mistakes

lazyai-cli workspace add /path/to/project --name my-project
lazyai-cli workspace switch my-project
lazyai-cli workspace status

source <(lazyai-cli completion bash)
```

### Transitional runtime extras (selected examples)

These commands still ship today, but they are secondary surfaces outside setup-core and are not required by generated adapter output. The examples below are intentionally brief; see the [CLI Reference](docs/cli/reference.md) for the full categorized inventory.

```bash
lazyai-cli session start "Implement auth feature"
lazyai-cli ledger verify
lazyai-cli message send implementer "Need help" "Can you review the auth code?"
lazyai-cli git sync
lazyai-cli backup create
lazyai-cli secret set api-key "sk-..."
lazyai-cli notify test
lazyai-cli metrics dashboard
lazyai-cli memory list
```

See [CLI Reference](docs/cli/reference.md) and [Product Boundaries](docs/concepts/product-boundaries.md) for the full categorized command inventory.

### Runtime migration note

Legacy workflow/orchestration/taskqueue surfaces were removed in the runtime refactor.
See `docs/migration/fortnite-orchestrator-removal.md`.

---

## Sidecar (Optional)

LazyAI can keep your docs, specs, and plans in a dedicated **sidecar** directory instead of inside each project. This is useful when you want a single knowledge base shared across workspaces, or when you prefer to keep planning artifacts outside version control.

### What sidecar means in LazyAI

A sidecar is a separate directory on disk that stores:
- `docs/` — documentation and guides
- `specs/` — feature specifications and ADRs
- `plans/` — execution plans and task breakdowns

When a sidecar is configured, LazyAI resolves these directories from the sidecar path instead of the project/workspace root. If no sidecar is configured, LazyAI falls back to its default behavior (docs/specs/plans live in the current scope root).

### Scope behavior

Sidecar configuration can live at three levels, with **workspace** as the primary use case:

| Scope | Config file | Priority |
|---|---|---|
| **Workspace** | `~/.lazyai/workspaces.yaml` (active workspace entry) | Highest |
| **Project** | `<project-root>/.lazyai-sidecar.yaml` | Middle |
| **Global** | `~/.lazyai/sidecar.yaml` | Lowest |

Resolution follows the chain: **workspace → project → global → default**. A workspace sidecar always wins over a project sidecar; a project sidecar wins over global; if none are configured, LazyAI uses the scope default.

**Workspace scope (recommended):**
- Best for multi-repo teams with a planning repo
- The active workspace entry in `workspaces.yaml` carries the sidecar block
- All projects in that workspace share the same sidecar by default

**Project scope:**
- Best when one repo needs its own isolated docs/specs/plans
- Create `.lazyai-sidecar.yaml` in the project root

**Global scope:**
- Best for personal defaults across all projects
- Set once in `~/.lazyai/sidecar.yaml`

### Commands

```bash
# Initialize a sidecar at a scope
lazyai-cli sidecar init --scope workspace --path /Users/me/kb/my-workspace

# Show resolved paths for the current scope
lazyai-cli sidecar status
# → Scope: workspace | Config Level: workspace
# → Docs:  /Users/me/kb/my-workspace/docs
# → Specs: /Users/me/kb/my-workspace/specs
# → Plans: /Users/me/kb/my-workspace/plans

# Attach a sidecar to the active workspace or project
lazyai-cli sidecar attach --path /tmp/kb

# Detach (remove) the sidecar configuration
lazyai-cli sidecar detach

# Validate sidecar paths exist and are writable
lazyai-cli sidecar doctor
```

### Optional fallback behavior

Sidecar is **always optional**. If you never run `sidecar init`, LazyAI behaves exactly as it does today:
- `project` scope → docs/specs/plans live in the project root
- `workspace` scope → docs/specs/plans live in the workspace (planning repo) root
- `global` scope → docs/specs/plans live in `~/.lazyai/`

No sidecar configured = no errors, no warnings, no behavior change.

### Explicit exclusions

- **No Skeeper integration.** The sidecar is purely local. There is no `skeeper` field, no provider abstraction, and no remote sync.
- **No content migration.** `sidecar init` does not move existing docs/specs/plans.
- **No multi-sidecar.** One sidecar per scope level.
- **No auto-discovery.** Sidecars are explicitly configured, not detected from parent directories or environment variables.

---

## Supported Tools

- [OpenCode](docs/concepts/tools.md#opencode)
- [Claude Code](docs/concepts/tools.md#claude-code)
- [GitHub Copilot](docs/concepts/tools.md#github-copilot)
- [OMP/Pi](docs/concepts/tools.md#omppi)
- [Antigravity](docs/concepts/tools.md#antigravity)

> **Note:** OpenCode now installs the neutral canonical adapter path with `implementer` as the default agent. Fortnite-era OpenCode defaults are retired from the active default path and are not installed by default.

---

## Documentation

- **Official docs:** <https://rluisb.github.io/lazyai/>
- **GitHub Wiki:** <https://github.com/rluisb/lazyai/wiki>

| Topic | Link |
|---|---|
| Quick Start | [docs/getting-started/quick-start.md](docs/getting-started/quick-start.md) |
| Installation | [docs/getting-started/installation.md](docs/getting-started/installation.md) |
| How It Works | [docs/concepts/how-it-works.md](docs/concepts/how-it-works.md) |
| Product Boundaries | [docs/concepts/product-boundaries.md](docs/concepts/product-boundaries.md) |
| Scopes | [docs/concepts/scopes.md](docs/concepts/scopes.md) |
| Presets | [docs/concepts/presets.md](docs/concepts/presets.md) |
| Tools | [docs/concepts/tools.md](docs/concepts/tools.md) |
| CLI Reference | [docs/cli/reference.md](docs/cli/reference.md) |
| MCP Integration | [docs/integration/mcp.md](docs/integration/mcp.md) |
| Runtime removal note | [docs/integration/orchestration.md](docs/integration/orchestration.md) |
| Contributing | [docs/development/contributing.md](docs/development/contributing.md) |
| Release Process | [docs/development/release.md](docs/development/release.md) |
| FAQ | [docs/troubleshooting/faq.md](docs/troubleshooting/faq.md) |

---

## Development

Requirements:

- Go 1.26+

```bash
cd packages/cli && go test ./...
cd ../diffviewer && go test ./...
```

Read the full [Contributing guide](docs/development/contributing.md).

---

## License

MIT. See [LICENSE](LICENSE).
