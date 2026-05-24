# CLI Reference

Complete reference for all `lazyai-cli` commands.

---

## Table of Contents

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
- sqlite3: Database engine
- git: Version control
- jq: JSON processor
- bash: Shell
- ollama: Local LLM runtime
- openai: API access
- disk: Disk usage
- orchestrator: MCP runtime

**Example:**
```bash
lazyai-cli doctor
```

**Output:**
```
LazyAI Health Check
═══════════════════════════════════════════════════════════════
✅ sqlite3     | SQLite database engine
✅ git         | Git version control
✅ jq          | JSON processor
✅ bash        | Bash shell
⚠️  ollama     | Local LLM runtime (not running)
✅ openai      | API access configured
✅ disk        | 45% usage
✅ orchestrator| MCP runtime
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
- Common mistakes

**Example:**
```bash
lazyai-cli validate agents
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
- `--output, -o`: Output file path (default: metrics.prom)

**Example:**
```bash
lazyai-cli metrics export --output my-metrics.prom
```

---

### `metrics dashboard`

Generate HTML dashboard.

**Flags:**
- `--output, -o`: Output file path (default: dashboard.html)

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

**Example:**
```bash
lazyai-cli eval run agent-quality
```

---

## Workflow Execution

Workflows are defined in `.opencode/workflows/*.yaml`.

### Workflow Structure

```yaml
name: rpi
version: "1.0"
description: "Research → Plan → Implement"

phases:
  - name: research
    agent: scout
    inputs:
      - task_description
    outputs:
      - findings_document

  - name: plan
    agent: planner
    inputs:
      - findings_document
    outputs:
      - spec_document

  - name: implement
    agent: builder
    inputs:
      - spec_document
    outputs:
      - implementation
```

---


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
