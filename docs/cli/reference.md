# CLI Reference

Complete reference for all `lazyai-cli` commands.

---

## Table of Contents

- [Init](#init)
- [Session Management](#session-management)
- [Health Checks](#health-checks)
- [Audit Trail](#audit-trail)
- [Validation](#validation)
- [Task Queue](#task-queue)
- [Agent Message Bus](#agent-message-bus)
- [Metrics Dashboard](#metrics-dashboard)
- [Memory Vault](#memory-vault)
- [Evaluation Harness](#evaluation-harness)
- [Workflow Execution](#workflow-execution)
- [Workspace](#workspace)
- [Completion](#completion)

---

## Init

### `init`

Initialize the AI development environment.

**Flags:**

| Flag | Description |
|---|---|
| `--scope` | Setup scope (`project`, `global`, `workspace`) |
| `--tools` | Tools to configure (`opencode`, `claude-code`, `copilot`) |
| `--preset` | Preset configuration (`minimal`, `standard`, `full`, `custom`) |
| `--enable-servers` | MCP servers to enable (`orchestrator`, `filesystem`, `memory`) |
| `--name` | Project name |
| `--no-interactive` | Run without interactive prompts |
| `--plain-opencode` | Use plain OpenCode scaffolding without Fortnite multi-agent mode |
| `--force` | Overwrite existing files |
| `--dry-run` | Preview changes without writing |

**OpenCode default behavior:**

When OpenCode is selected, `init` defaults to the **Fortnite/OpenCode runtime**. The default install includes Fortnite agents, skills, scripts, and workflows, plus `AGENTS.md`, `.opencode/STARTUP.md`, and `opencode.jsonc` with `default_agent: loop-driver`.

To opt out and install legacy/generic OpenCode assets, use `--plain-opencode`.

**Examples:**

```bash
# Interactive setup
lazyai-cli init

# Non-interactive with OpenCode (Fortnite runtime by default)
lazyai-cli init --tools opencode --preset standard --no-interactive

# Opt out of Fortnite runtime
lazyai-cli init --tools opencode --plain-opencode --no-interactive
```

---

## Session Management

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

## Health Checks

### `doctor`

Run health checks on the environment.

**Checks:**
- File integrity (managed files present and unmodified)
- Stray AGENTS.md files in specs/
- Metadata gaps in spec frontmatter
- Stale Claude MCP entries referencing legacy orchestrator
- Dependencies: sqlite3, git, jq, bash
- Providers: ollama (localhost:11434), openai (API key)
- Disk space usage
- Orchestrator binary on PATH

**Example:**
```bash
lazyai-cli doctor
```

**Output:**
```
🩺 Integrity Check

  Status ✅ All files healthy
  Health ████████████████████ 100%
  Total files 42
  Healthy 42
  Missing 0
  Modified 0
  Stray AGENTS.md 0
  Metadata gaps 0
  Stale MCP entries 0

🩺 Environment Health Checks

  ✅ Dependency: sqlite3     | 3.39.5
  ✅ Dependency: git         | git version 2.39.0
  ⚠️  Dependency: jq          | jq not found (optional but recommended)
  ✅ Dependency: bash         | GNU bash, version 5.2.0
  ⚠️  Provider: ollama       | Ollama not running on localhost:11434
  ✅ Provider: openai        | API key configured
  ✅ Disk space              | 45% used
  ⚠️  Orchestrator binary     | lazyai-orchestrator not on PATH
```

---

## Audit Trail

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
lazyai-cli ledger append dispatch "agent=builder task=auth"
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

## Validation

### `validate agents`

Validate agent file structure.

**Checks:**
- Dispatch parameters present
- Tool schemas correct
- Common mistakes (text vs content, mode misuse)

**Example:**
```bash
lazyai-cli validate agents
```

---

### `validate skills`

Validate skill file structure.

**Checks:**
- Skills directory exists under `.opencode/skills/`
- Basic structure validation (expanding)

**Example:**
```bash
lazyai-cli validate skills
```

---

## Task Queue

### `task create [description]`

Create a new task.

**Arguments:**
- `description` (required): Task description

**Example:**
```bash
lazyai-cli task create "Implement login page"
```

---

### `task list`

List all tasks.

**Example:**
```bash
lazyai-cli task list
```

**Output:**
```
Tasks:
───────────────────────────────────────────────────────────────
  task_1234567890 [pending] Implement login page
  task_1234567889 [claimed] Fix navigation bug | Claimed by: builder
```

---

### `task claim [task-id]`

Claim a task for processing.

**Arguments:**
- `task-id` (required): Task ID to claim

**Example:**
```bash
lazyai-cli task claim task_1234567890
```

---

### `task complete [task-id]`

Mark a task as completed.

**Arguments:**
- `task-id` (required): Task ID to complete

**Example:**
```bash
lazyai-cli task complete task_1234567890
```

---

## Agent Message Bus

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
lazyai-cli message send builder "Need help" "Can you review the auth code?" --priority high
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
lazyai-cli message recv builder
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

## Metrics Dashboard

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

## Memory Vault

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

## Evaluation Harness

### `eval list`

List available evaluation suites.

**Example:**
```bash
lazyai-cli eval list
```

---

### `eval run [suite-name]`

Run an evaluation suite.

**Arguments:**
- `suite-name` (required): Name of the evaluation suite

**Note:** Evaluation logic is currently a stub. The command validates the suite exists and records the attempt to the ledger, but does not yet execute the suite.

**Example:**
```bash
lazyai-cli eval run agent-quality
```

---

## Workflow Execution

Workflows are defined in `.opencode/workflows/*.yaml`.

### `workflow list`

List available workflows.

**Example:**
```bash
lazyai-cli workflow list
```

---

### `workflow show [workflow-name]`

Show workflow details.

**Arguments:**
- `workflow-name` (required): Name of the workflow

**Example:**
```bash
lazyai-cli workflow show rpi
```

---

### `workflow run [workflow-name]`

Execute a workflow.

**Arguments:**
- `workflow-name` (required): Name of the workflow

**Flags:**
- `--dry-run`: Show what would be executed without running (default: `true`)

**Note:** By default, `workflow run` operates in dry-run mode. Use `--dry-run=false` to actually execute phases.

**Example:**
```bash
# Dry run (default)
lazyai-cli workflow run rpi

# Actual execution
lazyai-cli workflow run rpi --dry-run=false
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

## Git Integration

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

## Secrets

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

## Notifications

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

## Sidecar Commands

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
| `LAZYAI_AGENT` | Current agent name | `orchestrator` |
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
