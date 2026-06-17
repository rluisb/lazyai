# LazyAI Commands

## Session Management

Track AI agent sessions with full audit trail.

### Start a Session

```bash
lazyai-cli session start "Implement user authentication"
```

Output:
```
✅ Session started: ses_1779583480
   Goal: Implement user authentication
   Started: 2026-05-24T00:44:40Z
```

### List Sessions

```bash
lazyai-cli session list
```

Output:
```
Sessions:
---------
🟢 ses_1779583480 | Implement user authentication | 2026-05-24T00:44:40Z
🔴 ses_1779582575 | Fix login bug | 2026-05-24T00:29:34Z
```

### Show Session Details

```bash
lazyai-cli session show ses_1779583480
```

Output:
```
Session: ses_1779583480
Goal: Implement user authentication
Status: active
Started: 2026-05-24T00:44:40Z

Dispatches:
  [completed] implementer: Implement auth middleware
  [completed] reviewer: Review auth implementation
```

### End a Session

```bash
lazyai-cli session end ses_1779583480
```

Output:
```
✅ Session ended: ses_1779583480
   Ended: 2026-05-24T00:54:25Z
```

## Health Checks

Validate your environment before starting work.

### Run Doctor

```bash
lazyai-cli doctor
```

Doctor reports setup integrity first, then these 6 environment checks:

1. `sqlite3` dependency
2. `git` dependency
3. `jq` dependency (optional warning)
4. `bash` dependency
5. `ollama` provider on `localhost:11434` (warning if unavailable)
6. Disk space usage

### JSON Output

```bash
lazyai-cli doctor --json
```

Useful for CI/CD pipelines and automated monitoring.

## Setup and Configuration Commands

These setup-core commands cover common maintenance tasks around an initialized LazyAI project.

### Manage MCP Servers

Enable or disable cataloged MCP servers, then compile per-tool configs.

```bash
lazyai-cli server list
lazyai-cli server add ai-memory
lazyai-cli compile
```

### Configure Local Defaults

Create and update `.opencode/config.yaml`.

```bash
lazyai-cli config init
lazyai-cli config set notifications.enabled true
```

### Create an Artifact

Generate agents, skills, prompts, commands, templates, or hooks.

```bash
lazyai-cli create agent release-reviewer --description "Reviews release readiness" --no-interactive
```

### Manage Sidecars

Attach optional docs/specs/plans directories to workspace, project, or global scope.

```bash
lazyai-cli sidecar init --scope project --path ../shared-docs
lazyai-cli sidecar status
```

### Import an Existing Setup

Import an existing tool setup into LazyAI-managed form.

```bash
lazyai-cli import ../existing-setup --tool opencode --preview
```

### Migrate Older LazyAI or ai-setup Config

Move older LazyAI or ai-setup files into the current canonical format.

```bash
lazyai-cli migrate ~/old-ai-setup --preview
```

### Update the CLI Binary

Check or install a GitHub Release for `lazyai-cli`.

```bash
lazyai-cli update-self --check
```

### Inspect Setup Inventory

Scan or plan supported setup targets without starting runtime execution.

```bash
lazyai-cli setup --scan
lazyai-cli setup --dry-run --global --all
```


## Audit Trail

Immutable hash-chained ledger for tracking all actions.

### Initialize Ledger

```bash
lazyai-cli ledger init
```

Creates `.specify/ledger.jsonl` with a genesis entry.

### Append Events

```bash
lazyai-cli ledger append dispatch "agent=implementer task=auth"
lazyai-cli ledger append session_start "goal=Implement auth"
lazyai-cli ledger append validation "status=pass agents=8"
```

### Verify Integrity

```bash
lazyai-cli ledger verify
```

Verifies:
- Hash chain continuity
- Entry hash correctness
- No tampering detected

### Show Entries

```bash
lazyai-cli ledger show 10
```

Shows last 10 entries with truncated hashes.

## Validation

Check agent and skill file structure.

### Validate Agents

```bash
lazyai-cli validate agents
```

Checks:
- YAML frontmatter present (**error**)
- `# System Prompt` heading present (**error**)
- Managed marker present (**warning**)
- Section heading after `# System Prompt` present (**warning**)

### Validate Skills

```bash
lazyai-cli validate skills
```

Current behavior:
- Confirms `.opencode/skills/` exists
- Prints `Skill validation not yet implemented`
- Lists planned checks for quick reference, frontmatter, and script references

## Workspace

Manage multi-project workspaces.

### List Workspaces

```bash
lazyai-cli workspace list
```

### Add a Workspace

```bash
lazyai-cli workspace add /path/to/project --name my-project
```

### Switch Active Workspace

```bash
lazyai-cli workspace switch my-project
```

### Show Workspace Status

```bash
lazyai-cli workspace status
```

## Runtime migration note

Legacy `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` CLI commands were removed during the runtime refactor. Repository harness scripts are not replacements for those product commands.

See `../migration/fortnite-orchestrator-removal.md` and `../concepts/product-boundaries.md` for replacements, rollback guidance, and current command ownership.

## Agent Message Bus

SQLite-based messaging between agents.

### Send a Message

```bash
lazyai-cli message send implementer "Need help" "Can you review the auth code?" --priority high
```

### Receive Messages

```bash
lazyai-cli message recv implementer
```

**Note:** This marks all unread messages for the agent as read and shows the 10 most recent messages.

### Broadcast

```bash
lazyai-cli message broadcast "All hands" "System update at 2pm"
```

## Git Integration

### Sync Changes

```bash
lazyai-cli git sync
```

**Safety note:** This command stages and commits ALL changes in the current repository after showing the file list and asking for confirmation. Use `--force` to skip the confirmation prompt.

## Backup

### Create Backup

```bash
lazyai-cli backup create
```

### Restore from Backup

```bash
lazyai-cli backup restore <backup-file>
```

**Safety note:** This overwrites current LazyAI data with the backup contents. Use `--force` to skip the confirmation prompt.

## Secrets

### Store a Secret

```bash
lazyai-cli secret set <name> <value>
```

### Retrieve a Secret

```bash
lazyai-cli secret get <name>
```

**Safety note:** Secrets are stored locally using the OS keychain when available, or a fallback file in `~/.lazyai/secrets/` (not encrypted at rest). Do not use for production credentials.

## Notifications

### Configure Webhook

```bash
lazyai-cli notify config --webhook https://hooks.example.com/notify
```

**Safety note:** The webhook URL is stored in the local config file. Ensure the URL is trusted and uses HTTPS.

### Send a Test Notification

```bash
lazyai-cli notify test
```

## Metrics

### Export Metrics

```bash
lazyai-cli metrics export
```

Writes `metrics.prom` (Prometheus format) in the current directory. Use `-o` to specify a different path.

### Generate Dashboard

```bash
lazyai-cli metrics dashboard
```

Writes `dashboard.html` in the current directory. Use `-o` to specify a different path.

## Completion

Generate shell completion scripts.

```bash
# Bash
source <(lazyai-cli completion bash)

# Zsh
source <(lazyai-cli completion zsh)

# Fish
lazyai-cli completion fish | source
```

## Typical Workflow

```bash
# 1. Check environment
lazyai-cli doctor

# 2. Start session
lazyai-cli session start "Implement feature X"

# 3. Work with agents...
# (agent dispatches are automatically logged)

# 4. Validate changes
lazyai-cli validate agents

# 5. Check audit trail
lazyai-cli ledger verify

# 6. End session
lazyai-cli session end ses_1234567890
```
