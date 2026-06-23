# CLI Reference

Command reference for shipped `lazyai-cli` commands. Repository scripts such as `bin/doctor`, `bin/inject`, and `bin/startup-self-heal` are maintainer harness tools, not CLI commands; archived Fortnite/orchestrator/eval material is historical or migration context only.

The commands under `ops-runtime-extra` are still physically shipped today, but they are transitional extras outside the default setup-core product boundary rather than the headline LazyAI story. For the source-backed command category inventory and rationales, see [Product Boundaries](../concepts/product-boundaries.md); the same category names are guarded by `packages/cli/cmd/command_category.go`.

## Top-level command surface

These are the top-level commands registered by source under `packages/cli/cmd/`:

| Category | Commands |
|---|---|
| `setup-core` | `add`, `build-plugin`, `compile`, `completion`, `config`, `create`, `doctor`, `eject`, `import`, `info`, `init`, `list`, `migrate`, `server`, `setup`, `sidecar`, `status`, `update`, `update-self`, `validate`, `workspace` |
| `ops-runtime-extra` | `auth`, `backup`, `cost`, `git`, `ledger`, `memory`, `message`, `metrics`, `notify`, `restore-runtime-db`, `secret`, `session` |
| `dev-harness` | `models` |
| `retired/archived` | `completions` (hidden deprecated alias for `completion`) |

Removed command surfaces such as `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` are not active CLI commands.

---

## Table of Contents

- [Top-level command surface](#top-level-command-surface)
- [Init](#init)
- [Add](#add)
- [Compile](#compile)
- [Update](#update)
- [Status](#status)
- [Doctor](#doctor)
- [Server](#server)
  - [server list](#server-list)
  - [server add](#server-add)
  - [server remove](#server-remove)
  - [server doctor](#server-doctor)
- [Config](#config)
- [Create](#create)
- [Import](#import)
- [Migrate](#migrate)
- [Eject](#eject)
- [Info](#info)
- [List](#list)
- [Build Plugin](#build-plugin)
- [Update Self](#update-self)
- [Setup](#setup)
- [Models](#models)
- [Session](#session)
- [Auth](#auth)
- [Cost](#cost)
- [Ledger](#ledger)
- [Validate](#validate)
  - [validate agents](#validate-agents)
  - [validate skills](#validate-skills)
  - [validate evals](#validate-evals)
- [Migration Note](#migration-note)
- [Message](#message)
- [Metrics](#metrics)
- [Memory](#memory)
- [Workspace](#workspace)
- [Completion](#completion)
- [Git](#git)
- [Backup](#backup)
- [Restore Runtime DB](#restore-runtime-db)
- [Secret](#secret)
- [Notify](#notify)
- [Sidecar](#sidecar)
- [Environment Variables](#environment-variables)
- [Exit Codes](#exit-codes)

---

## Init

### `init`

Initialize the AI development environment.

**Flags:**

| Flag | Description |
|---|---|
| `--scope` | Setup scope (`project`, `global`, `workspace`) |
| `--tools` | Tools to configure (`opencode`, `claude-code`, `copilot`, `pi`, `omp`, `kiro`, `antigravity`) |
| `--preset` | Preset configuration (`minimal`, `standard`, `full`, `custom`) |
| `--enable-servers` | MCP servers to enable (for example: `filesystem`, `ai-memory`, `ripgrep`) |
| `--name` | Project name |
| `--no-interactive` | Run without interactive prompts |
| `--force` | Overwrite existing files |
| `--dry-run` | Preview changes without writing |

**OpenCode default behavior:**

When OpenCode is selected, `init` installs the neutral canonical adapter path. Root `opencode.json` keeps the baseline config shape, `.opencode/agents/guide.md` is emitted as the front-door default agent, and Fortnite agents, `.opencode/STARTUP.md`, and `loop-driver` are not installed by default.

**Examples:**

```bash
# Interactive setup
lazyai-cli init

# Non-interactive with OpenCode
lazyai-cli init --tools opencode --preset standard --no-interactive
```

**AGENTS.md placeholders (headless init):**

`init` writes a canonical `AGENTS.md` that contains `<!-- fill-in: ... -->` markers for project-specific sections (organization, conventions, architecture, etc.). The CLI does not fill these markers itself because it cannot run your AI tool. After `init` finishes:

- **Host tool path (recommended):** Open the project in your AI tool (Claude Code, OpenCode, etc.) and run `/init` or `/populate`. The AI tool indexes the codebase and replaces each marker with a value backed by code evidence.
- **Manual path:** Edit `AGENTS.md` by hand and replace each `<!-- fill-in: ... -->` marker with a concrete value, or remove the marker if the section does not apply.
- **Validate after populate:**

  ```bash
  lazyai-cli validate agents   # confirm agent frontmatter is still valid
  lazyai-cli doctor            # full setup health check
  ```

Both `--no-interactive` and the interactive wizard leave markers unfilled on purpose; filling requires the AI tool to run.

---
## Add

### `add`

Add agents, skills, or tool configurations to an existing initialized setup, then rerun the scaffold pipeline and update the setup store.

**Flags:**

| Flag | Description |
|---|---|
| `--tools` | Tool IDs to add configuration for |
| `--agents` | Agent IDs to add |
| `--skills` | Skill IDs to add |
| `--no-interactive` | Run without prompts; at least one of `--tools`, `--agents`, or `--skills` is required |

**Tool validation:**
`--tools` is registry-backed. Current tool IDs are `opencode`, `claude-code`, `copilot`, `pi`, `omp`, `kiro`, and `antigravity`.

**Examples:**

```bash
lazyai-cli add --tools claude-code,pi --no-interactive
lazyai-cli add --agents reviewer,evidence-verifier --skills fast-feedback --no-interactive
```

---

## Compile

### `compile`

Compile `.ai/mcp.json` or `.ai/mcp.jsonc` into per-tool MCP/config files for the selected tools in the setup store.

**Flags:**

| Flag | Description |
|---|---|
| `--tool` | Compile only one tool (`opencode`, `claude-code`, `copilot`, `pi`, `omp`, `kiro`, or `antigravity`) |
| `--dry-run` | Preview the tools that would be compiled without writing |
| `--local-secrets` | Route Claude Code MCP writes to `.claude/settings.local.json` instead of committed config |
| `--validate-contracts` | Validate skill contracts before compile (default: `true`) |
| `--strict-contracts` | Turn contract warnings into compile failures |

**Example:**

```bash
lazyai-cli compile --tool claude-code --local-secrets
```

---

## Update

### `update`

Refresh managed files to match the embedded library version while applying LazyAI conflict handling.

**Flags:**

| Flag | Description |
|---|---|
| `--force` | Overwrite local changes without prompting |
| `--no-interactive` | Run without interactive prompts |
| `--dry-run` | Show what would change without writing |

**Example:**

```bash
lazyai-cli update --dry-run
```

---

## Status

### `status`

Show current setup state, selected tools, file health, git conventions, and CLI version.

**Flags:**

| Flag | Description |
|---|---|
| `--dir` | Target directory (default: current directory) |
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli status --json
```

---

## Server

Manage MCP servers from the embedded catalog and the current setup store. The current catalog includes `ai-memory`, `filesystem`, `ripgrep`, `codegraph`, and `obsidian`.

### `server list`

List cataloged MCP servers and show whether each is enabled.

**Flags:**

| Flag | Description |
|---|---|
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli server list --json
```

---

### `server add`

Enable a cataloged MCP server in the setup store. Run `lazyai-cli compile` afterward to regenerate per-tool configs.

**Arguments:**

- `name` (required): Catalog server name

**Flags:**

| Flag | Description |
|---|---|
| `--no-interactive` | Skip confirmation prompt |

**Example:**

```bash
lazyai-cli server add ai-memory
lazyai-cli compile
```

Unknown server names fail with `unknown MCP server: <name>` and include the available catalog names in the error metadata.

---

### `server remove`

Disable a cataloged MCP server in the setup store. Run `lazyai-cli compile` afterward to remove it from generated per-tool configs.

**Arguments:**

- `name` (required): Catalog server name

**Flags:**

| Flag | Description |
|---|---|
| `--no-interactive` | Skip confirmation prompt |

**Example:**

```bash
lazyai-cli server remove ripgrep
lazyai-cli compile
```

---

### `server doctor`

Validate enabled MCP server definitions in `.ai/mcp.json` and the generated per-tool MCP config files. Pass a server name to validate only one server.

**L1 config check:** Verifies the server entry exists in `.ai/mcp.json`, is enabled, and is present in each per-tool compiled MCP config file.

**L3 stdio handshake (not performed):** The Go binary does not bundle an MCP client library for the JSON-RPC handshake protocol. A TypeScript wrapper using `@modelcontextprotocol/sdk` could perform L3 checks; this is future work.

**Arguments:**

- `name` (optional): Catalog server name

**Flags:**

| Flag | Description |
|---|---|
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli server doctor ai-memory --json
```

## Config

### `config init`

Create `.opencode/config.yaml` with defaults for agent, model, database, ledger, and notifications.

```bash
lazyai-cli config init
```

### `config get`

Read one config key.

```bash
lazyai-cli config get agent.default_agent
```

### `config set`

Set one config key. Critical keys print a warning before saving.

```bash
lazyai-cli config set notifications.enabled true
```

Supported keys: `agent.default_agent`, `agent.default_model`, `database.path`, `ledger.path`, `notifications.enabled`, and `notifications.webhook`.

### `config list`

Print all current config values.

```bash
lazyai-cli config list
```

---

## Create

### `create`

Generate a new artifact using the artifact generator registry.

**Usage:**

```bash
lazyai-cli create [type] [name]
```

**Valid types:** `agent`, `skill`, `prompt`, `command`, `template`, `hook`

**Flags:**

| Flag | Description |
|---|---|
| `--force` | Overwrite an existing artifact |
| `--no-interactive` | Require type and name from arguments; do not prompt |
| `--description` | Description used by generators that support it |

**Example:**

```bash
lazyai-cli create agent release-reviewer --description "Reviews release readiness" --no-interactive
```

Invalid types fail with:

```text
invalid artifact type: <type> (valid: agent, skill, prompt, command, template, hook)
```

---

## Import

### `import`

Import configuration from another AI tool setup into LazyAI-managed form. If `source` is omitted, the current directory is scanned.

**Usage:**

```bash
lazyai-cli import [source]
```

**Flags:**

| Flag | Description |
|---|---|
| `--tool` | Source tool to import (`opencode`, `claude-code`, `copilot`, `pi`, `antigravity`, `omp`, or `kiro`); auto-detected if omitted |
| `--no-interactive` | Run without confirmation prompts |
| `--preview` | Show the import plan without writing |
| `--strategy` | Merge strategy: `smart`, `preserve`, `replace`, or `append` |
| `--verbose` | Show detailed output |
| `--skip-backup` | Do not create a backup before writing |

**Example:**

```bash
lazyai-cli import ../existing-setup --tool opencode --preview
```

---

## Migrate

### `migrate`

Migrate an older LazyAI or ai-setup configuration into the current canonical format. If `source` is omitted, the current directory is scanned.

**Usage:**

```bash
lazyai-cli migrate [source]
```

**Flags:**

| Flag | Description |
|---|---|
| `--from` | Source version hint |
| `--no-interactive` | Run without confirmation prompts |
| `--preview` | Show the migration plan without writing |
| `--strategy` | Merge strategy: `smart`, `preserve`, `replace`, or `append` |
| `--verbose` | Show detailed output |
| `--skip-backup` | Do not create a backup before writing |

**Example:**

```bash
lazyai-cli migrate ~/old-ai-setup --preview
```

---

## Eject

### `eject`

Remove `.ai-setup.json` and `.ai-setup.db` so LazyAI stops managing the generated files. The generated files remain in place.

**Flags:**

| Flag | Description |
|---|---|
| `--no-interactive` | Skip confirmation prompt |

**Example:**

```bash
lazyai-cli eject --no-interactive
```

---

## Info

### `info`

Show metadata for one tracked artifact by basename or tracked path.

**Arguments:**

- `name` (required): Artifact basename or tracked file path

**Flags:**

| Flag | Description |
|---|---|
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli info guide --json
```

---

## List

### `list`

List installed tracked artifacts, optionally filtered by category.

**Usage:**

```bash
lazyai-cli list [category]
```

**Flags:**

| Flag | Description |
|---|---|
| `--type` | Filter by artifact type (`agents`, `skills`, `templates`, `rules`, and so on) |
| `--verbose` | Include file paths |
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli list agents --verbose
```

---

## Build Plugin

### `build-plugin`

Generate a distributable plugin bundle from the embedded LazyAI library.
Supports four targets via `--target`: Claude Code (default), GitHub Copilot CLI,
OMP, and Pi.

**Flags:**

| Flag | Description |
|---|---|
| `--target` | Bundle target: `claude` (default), `copilot-cli`, `omp`, or `pi` |
| `--out` | Output directory (default: `./dist/plugin`) |
| `--force` | Overwrite the output directory if it exists and is non-empty |

**Example:**

```bash
lazyai-cli build-plugin --target claude --out ./dist/lazyai-plugin --force
lazyai-cli build-plugin --target copilot-cli --out ./dist/copilot-plugin
```

---

## Update Self

### `update-self`

Download and replace the current `lazyai-cli` binary from GitHub Releases. Development builds print `Development build — skipping self-update` unless `--version` is provided.

**Flags:**

| Flag | Description |
|---|---|
| `--check` | Only check whether an update is available; exits non-zero when one is available |
| `--force` | Skip version comparison |
| `--dry-run` | Preview without downloading |
| `--verbose` | Show detailed release and asset information |
| `--version` | Download a specific release tag |

**Example:**

```bash
lazyai-cli update-self --check
lazyai-cli update-self --version v1.2.3 --dry-run
```

---

## Setup

### `setup`

Inspect setup inventory and setup planning output. This command reports setup state; it does not start daemons or runtime execution.

**Flags:**

| Flag | Description |
|---|---|
| `--scan` | Scan known tool targets and print inventory JSON |
| `--list` | List supported setup targets and reusable setup resources |
| `--dry-run` | Show the setup plan without writing |
| `--adopt` | Mark adoptable external configs as LazyAI managed; requires `--scan` |
| `--import` | Import external configs into LazyAI reference storage; requires `--scan` |
| `--tool` | Limit planning to a tool; repeatable and only valid with `--list` or `--dry-run` |
| `--all` | Select all supported setup targets for the requested scope |
| `--global` | Use global scope/home layout where supported |

Select exactly one of `--scan`, `--list`, or `--dry-run`.

**Examples:**

```bash
lazyai-cli setup --scan
lazyai-cli setup --list --tool claude-code
lazyai-cli setup --dry-run --global --all
```

---

## Models

### `models sync`

Maintainer-only dev-harness command that refreshes `packages/cli/internal/models/catalog_gen.go` from `models.dev/api.json`.

It filters the upstream catalog to the providers each supported CLI uses, checks curated tier references, prints a diff, and writes only after confirmation unless `--yes` is passed.

**Flags:**

| Flag | Description |
|---|---|
| `--yes` | Skip the confirmation prompt and write if there are no missing curated references |
| `--output` | Write the regenerated catalog to a custom path |
| `--dry-run` | Print the diff without writing |

**Example:**

```bash
lazyai-cli models sync --dry-run
```

---

## Session

### `session start [goal]`

Start a new AI agent session.

**Arguments:**
- `goal` (required): Brief description of the session goal

**Example:**
```bash
lazyai-cli session start "Implement authentication feature"
```

**Output:**
```
✅ Session started: ses_1234567890
   Goal: Implement authentication feature
   Started: 2026-05-24T00:08:16Z
```

---

### `session list`

List all sessions.

**Example:**
```bash
lazyai-cli session list
```

**Output:**
```
Sessions:
---------
🟢 ses_1234567890 | Implement auth | 2026-05-24T00:08:16Z
🔴 ses_1234567889 | Fix bug #42 | 2026-05-23T23:45:00Z
```

---

### `session show [session-id]`

Show session details.

**Arguments:**
- `session-id` (required): Session ID (e.g., `ses_1234567890`)

**Example:**
```bash
lazyai-cli session show ses_1234567890
```

---

### `session end [session-id]`

End a session.

**Arguments:**
- `session-id` (required): Session ID to end

**Example:**
```bash
lazyai-cli session end ses_1234567890
```

---

## Doctor

### `doctor`

Validate setup integrity and run six environment health checks.

**Integrity checks:**

- Managed files exist and match tracked hashes
- Stray `AGENTS.md` files under `specs/`
- Metadata gaps in spec frontmatter

**Environment health checks:**

1. `sqlite3` dependency — fail if not found
2. `git` dependency — fail if not found
3. `jq` dependency — warn if not found
4. `bash` dependency — fail if not found
5. `ollama` provider — warn if not running on `localhost:11434`
6. Disk space — warn above 80% used, fail above 90% used

**Flags:**

| Flag | Description |
|---|---|
| `--dir` | Project directory to check (defaults to current directory) |
| `--fix` | Print fix instructions for detected issues |
| `--verbose` | Show detailed output for all files and metadata issues |
| `--json` | Output JSON |

**Example:**

```bash
lazyai-cli doctor
```

**Output:**

```text
🩺 Integrity Check

  Status ✅ All files healthy
  Health ████████████████████ 100%
  Total files 42
  Healthy 42
  Missing 0
  Modified 0
  Stray AGENTS.md 0
  Metadata gaps 0
  Health warnings 2

🩺 Environment Health Checks

  ✅ Dependency: sqlite3     | 3.39.5
  ✅ Dependency: git         | git version 2.39.0
  ⚠️  Dependency: jq          | jq not found (optional but recommended)
  ✅ Dependency: bash         | GNU bash, version 5.2.0
  ⚠️  Provider: ollama       | Ollama not running on localhost:11434
  ✅ Disk space              | 45% used
```

---

## Auth

### `auth list`

List configured AI providers from the setup store and show common authentication environment variables with masked values.

**Example:**

```bash
lazyai-cli auth list
```

Checks `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GOOGLE_API_KEY`, and `GITHUB_TOKEN`.

---

## Cost

### `cost show`

Show cost breakdown by session.

**Flags:**

| Flag | Description |
|---|---|
| `--period`, `-p` | Time period: `day`, `week`, `month`, or `all` |

```bash
lazyai-cli cost show --period week
```

### `cost agent`

Show cost and dispatch counts grouped by agent.

```bash
lazyai-cli cost agent
```

### `cost budget`

Show current spend against an optional budget limit.

**Flags:**

| Flag | Description |
|---|---|
| `--limit`, `-l` | Budget limit in USD |

```bash
lazyai-cli cost budget --limit 25
```

---

## Ledger

### `ledger init`

Initialize the immutable ledger.

**Example:**
```bash
lazyai-cli ledger init
```

---

### `ledger append [event-type] [data]`

Append an event to the ledger.

**Arguments:**
- `event-type` (required): Type of event (e.g., `dispatch`, `session_start`)
- `data` (required): Event data as key=value pairs

**Example:**
```bash
lazyai-cli ledger append dispatch "agent=implementer task=auth"
```

---

### `ledger verify`

Verify ledger integrity.

**Example:**
```bash
lazyai-cli ledger verify
```

**Output:**
```
🔍 Verifying ledger integrity...

  ✅ All 11 entries verified. Chain intact.
```

---

### `ledger show [count]`

Show recent ledger entries.

**Arguments:**
- `count` (optional): Number of entries to show (default: 10)

**Example:**
```bash
lazyai-cli ledger show 5
```

---

## Validate

The `validate` command runs structural validators over the canonical `.ai/`
asset tree. With no subcommand it prints help; pass `--all` to run every
validator in one pass, or use the `agents`, `skills`, and `evals` subcommands
for filtered per-surface checks.

**Flags:**

| Flag | Description |
|---|---|
| `--all` | Run all validators (skills, agents, hooks, MCP, secrets, path safety) over the canonical `.ai/` tree and return an error on any error-severity issue |
| `--profile` | Safety profile: `team` (inline secrets are errors) or `personal` (inline secrets are warnings). Defaults to the manifest profile, then `personal` |

### `validate agents`

Validate `.ai/agents/*.md` files using the consolidated validation engine.

The engine runs all validators over the canonical `.ai/` tree and then filters
to agent-rule issues. Counts agent files as passed, surfaces error and warning
findings, and exits with an error only when error-severity checks fail.

**Checks:**

- YAML frontmatter is present (**error**)
- `name` field is present and valid (**error**)
- `description` field is present (**warning**)
- Semantic body guidance for role, workflow, trigger, non-trigger, human gate,
  handoff, evidence, and output (**warning**)

**Example:**
```bash
lazyai-cli validate agents
```

---

### `validate skills`

Validate `.ai/skills/` using the consolidated validation engine over the
canonical `.ai/` tree. Filters to skill-rule issues, counts pass/fail across
flat `*.md` files and `<name>/SKILL.md` directories, and returns an error on
any error-severity failure.

**Checks:**

- YAML frontmatter is present (**error**)
- `name` field is present and valid (**error**)
- `description` field is present and non-empty (**error**)
- Semantic body guidance for trigger, non-trigger, evidence, and output
  (**warning**)

**Example:**
```bash
lazyai-cli validate skills
```

---

### `validate evals`

Validate `.ai/evals/cases/`, `.ai/evals/holdouts/`, and `.ai/evals/rubrics/`.
Case and holdout files (`.yaml`/`.yml`) are schema-checked by the evals
package; rubric directories are validated for structure. Counts pass/fail and
returns an error on any failure.

**Example:**
```bash
lazyai-cli validate evals
```

---

## Migration Note

The `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` command surfaces were removed during the runtime refactor.

Use:

- `docs/migration/fortnite-orchestrator-removal.md` for user-facing migration and rollback guidance
- `lazyai-cli create`, `lazyai-cli list`, and adapter-managed files for current scaffolding flows


Do not use repository harness scripts or archived assets as replacements for removed CLI commands unless you are maintaining this repository itself.

---


## Message

### `message send [to-agent] [subject] [body]`

Send a message to an agent.

**Arguments:**
- `to-agent` (required): Target agent name
- `subject` (required): Message subject
- `body` (required): Message body

**Flags:**
- `--priority, -p`: Message priority (low, normal, high, critical)

**Example:**
```bash
lazyai-cli message send implementer "Need help" "Can you review the auth code?" --priority high
```

---

### `message recv [agent]`

Receive messages for an agent.

**Arguments:**
- `agent` (required): Agent name to receive messages for

**Behavior:**
- Marks all unread messages for the agent as read
- Shows the 10 most recent messages

**Example:**
```bash
lazyai-cli message recv implementer
```

---

### `message broadcast [subject] [body]`

Broadcast a message to all agents.

**Arguments:**
- `subject` (required): Broadcast subject
- `body` (required): Broadcast body

**Flags:**
- `--priority, -p`: Message priority (low, normal, high, critical)

**Example:**
```bash
lazyai-cli message broadcast "All hands" "System update at 2pm"
```

---

## Metrics

### `metrics list`

List recent quality metrics.

**Flags:**
- `--limit, -n`: Number of metrics to show (default: 10)

**Example:**
```bash
lazyai-cli metrics list --limit 5
```

---

### `metrics export`

Export metrics to Prometheus format.

**Flags:**
- `--output, -o`: Output file path (default: `metrics.prom`)

**Safety note:** Writes a file to the current directory. Existing files are overwritten.

**Example:**
```bash
lazyai-cli metrics export --output my-metrics.prom
```

---

### `metrics dashboard`

Generate HTML dashboard.

**Flags:**
- `--output, -o`: Output file path (default: `dashboard.html`)

**Safety note:** Writes a file to the current directory. Existing files are overwritten.

**Example:**
```bash
lazyai-cli metrics dashboard --output my-dashboard.html
```

---

## Memory

### `memory save [content]`

Save a memory.

**Arguments:**
- `content` (required): Memory content

**Flags:**
- `--type, -t`: Memory type (lesson, context, decision, idea)
- `--tags`: Tags for categorization

**Example:**
```bash
lazyai-cli memory save "Always test migrations" --type lesson --tags database,migrations
```

---

### `memory list`

List all memories.

**Example:**
```bash
lazyai-cli memory list
```

---

### `memory search [query]`

Search memories.

**Arguments:**
- `query` (required): Search query

**Example:**
```bash
lazyai-cli memory search database
```

---


## Workspace


Manage multi-project workspaces.

### `workspace list`

List registered workspaces.

**Example:**
```bash
lazyai-cli workspace list
```

---

### `workspace add [path]`

Register a project path as a workspace.

**Arguments:**
- `path` (required): Path to the project directory

**Flags:**
- `--name`: Override workspace name (default: directory basename)

**Example:**
```bash
lazyai-cli workspace add /path/to/project --name my-project
```

---

### `workspace switch [name]`

Set the active workspace by name.

**Arguments:**
- `name` (required): Workspace name

**Example:**
```bash
lazyai-cli workspace switch my-project
```

---

### `workspace status`

Show active workspace details.

**Example:**
```bash
lazyai-cli workspace status
```

---

## Completion

### `completion [shell]`

Generate shell completion scripts.

**Arguments:**
- `shell` (required): One of `bash`, `zsh`, `fish`, `powershell`

**Example:**
```bash
# Bash
source <(lazyai-cli completion bash)

# Zsh
source <(lazyai-cli completion zsh)

# Fish
lazyai-cli completion fish | source
```

---

## Git

### `git sync`

Auto-commit all changes with descriptive messages.

**Safety note:** This command automatically stages (`git add -A`) and commits ALL changes in the current repository. It shows the list of files and asks for confirmation before proceeding. Use `--force` to skip the confirmation prompt.

**Flags:**
- `--message, -m`: Custom commit message
- `--force`: Skip confirmation prompt

**Example:**
```bash
lazyai-cli git sync --message "Fix auth bug" --force
```

---

### `git log`

Show recent commits, highlighting LazyAI-attributed commits.

**Flags:**
- `--limit`, `-n`: Number of commits to show (default: 10)

**Example:**
```bash
lazyai-cli git log --limit 5
```

---

### `git status`

Show `git status` with LazyAI context, including the most recent active session when one is recorded.

**Example:**
```bash
lazyai-cli git status
```

---

## Backup

### `backup create`

Create a backup of LazyAI data.

**Flags:**
- `--output, -o`: Output file path (default: `lazyai-backup-YYYYMMDD_HHMMSS.tar.gz`)

**Example:**
```bash
lazyai-cli backup create --output my-backup.tar.gz
```

---

### `backup restore [backup-file]`

Restore LazyAI data from a backup tarball.

**Arguments:**
- `backup-file` (required): Path to the backup file

**Safety note:** This overwrites current LazyAI data with the backup contents. It asks for confirmation before proceeding. Use `--force` to skip the confirmation prompt.

**Flags:**
- `--force`: Skip confirmation prompt

**Example:**
```bash
lazyai-cli backup restore my-backup.tar.gz
```

---

## Restore Runtime DB

### `restore-runtime-db`

Restore `.specify/session.db` from a backup file. The current runtime database is renamed to `.specify/session.db.pre-restore` before the backup is copied into place.

**Arguments:**

- `backup-path` (required): Path to the SQLite backup file

**Flags:**

| Flag | Description |
|---|---|
| `--force` | Skip confirmation prompt |

**Safety note:** This overwrites the current runtime database. Data written since the backup was created will be lost.

**Example:**

```bash
lazyai-cli restore-runtime-db .specify/session.db.backup --force
```

---

## Secret

### `secret set [name] [value]`

Store a secret.

**Arguments:**
- `name` (required): Secret name
- `value` (required): Secret value

**Safety note:** Secrets are stored using the OS keychain when available, or a fallback file in `~/.lazyai/secrets/` (not encrypted at rest). Do not use for production credentials.

**Example:**
```bash
lazyai-cli secret set api-key "sk-..."
```

---

### `secret get [name]`

Retrieve a secret.

**Arguments:**
- `name` (required): Secret name

**Safety note:** The secret value is printed to stdout. Be careful when running this in shared environments or logging output.

**Example:**
```bash
lazyai-cli secret get api-key
```

---

### `secret list`

List stored secrets.

**Example:**
```bash
lazyai-cli secret list
```

---

### `secret remove [name]`

Remove a secret.

**Arguments:**
- `name` (required): Secret name

**Flags:**
- `--force`: Skip confirmation prompt

**Example:**
```bash
lazyai-cli secret remove api-key
```

---

## Notify

### `notify config`

Configure notification settings.

**Flags:**
- `--webhook`: Webhook URL for notifications
- `--enabled`: Enable or disable notifications

**Safety note:** The webhook URL is stored in the local config file. Ensure the URL is trusted and uses HTTPS.

**Example:**
```bash
lazyai-cli notify config --webhook https://hooks.example.com/notify --enabled
```

---

### `notify send [message]`

Send a desktop notification.

**Arguments:**
- `message` (required): Notification message

**Flags:**
- `--title, -t`: Notification title (default: "LazyAI")

**Example:**
```bash
lazyai-cli notify send "Build complete" --title "CI"
```

---

### `notify test`

Send a test notification to verify configuration.

**Example:**
```bash
lazyai-cli notify test
```

---

## Sidecar

Manage optional sidecar directories for docs, specs, and plans.

### `sidecar init`

Initialize a sidecar configuration at the specified scope.

**Flags:**
- `--scope`: Scope level (`workspace`, `project`, `global`). Default: `workspace`.
- `--path`: Sidecar root path (required).
- `--specs-dir`: Override specs directory name. Default: `specs`.
- `--docs-dir`: Override docs directory name. Default: `docs`.
- `--plans-dir`: Override plans directory name. Default: `plans`.

**Examples:**

```bash
# Workspace sidecar (recommended)
lazyai-cli sidecar init --scope workspace --path /Users/me/kb/my-workspace

# Project sidecar
lazyai-cli sidecar init --scope project --path ../shared-docs

# Global sidecar
lazyai-cli sidecar init --scope global --path ~/kb
```

### `sidecar status`

Show resolved docs/specs/plans paths for the current scope, including which config level provided each value.

**Flags:** none

**Example:**

```bash
lazyai-cli sidecar status
# → Scope: workspace | Config level: workspace
# → Docs:  /Users/me/kb/my-workspace/docs
# → Specs: /Users/me/kb/my-workspace/specs
# → Plans: /Users/me/kb/my-workspace/plans
```

### `sidecar attach`

Attach a sidecar to the active workspace or project. Requires an existing config target.

**Flags:**
- `--scope`: Scope level (`workspace`, `project`). Default: `workspace`.
- `--path`: Sidecar root path (required).
- `--specs-dir`: Override specs directory name.
- `--docs-dir`: Override docs directory name.
- `--plans-dir`: Override plans directory name.

**Example:**

```bash
lazyai-cli sidecar attach --path /tmp/kb
```

### `sidecar detach`

Remove the sidecar configuration from the active workspace or project.

**Flags:**
- `--scope`: Scope level (`workspace`, `project`). Default: `workspace`.
- `--force`: Skip confirmation prompt.

**Example:**

```bash
lazyai-cli sidecar detach
# → Remove workspace sidecar? [y/N]
```

### `sidecar doctor`

Validate all configured sidecar paths exist and are writable. Reports issues with exit codes.

**Flags:**
- `--scope`: Scope to validate (`workspace`, `project`, `global`). Default: `workspace`.

**Exit codes:**
- `0`: All paths valid, or warnings only (e.g., missing path that will be created on first write). WARN lines are printed but the command succeeds.
- `1`: Errors found (e.g., non-writable directory, file where directory expected).

**Example:**

```bash
lazyai-cli sidecar doctor
# → ✅ Sidecar path exists and is writable
# → ✅ Docs dir: /Users/me/kb/docs
# → ✅ Specs dir: /Users/me/kb/specs
# → ✅ Plans dir: /Users/me/kb/plans
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LAZYAI_AGENT` | Current agent name | `implementer` |
| `LAZYAI_DB_PATH` | Database file path | `.specify/lazyai.db` |
| `LAZYAI_LEDGER_PATH` | Ledger file path | `.specify/ledger.jsonl` |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation error |
| 3 | Database error |
| 4 | Ledger error |
