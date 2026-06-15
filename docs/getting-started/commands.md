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
  [completed] builder: Implement auth middleware
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

Checks:
- File integrity (managed files present and unmodified)
- Stray AGENTS.md files
- Metadata gaps in specs
- Stale Claude MCP entries
- Dependency health (sqlite3, git, jq, bash)
- Provider status (ollama, openai)
- Disk space
- Orchestrator binary on PATH

### JSON Output

```bash
lazyai-cli doctor --json
```

Useful for CI/CD pipelines and automated monitoring.

## Audit Trail

Immutable hash-chained ledger for tracking all actions.

### Initialize Ledger

```bash
lazyai-cli ledger init
```

Creates `.specify/ledger.jsonl` with a genesis entry.

### Append Events

```bash
lazyai-cli ledger append dispatch "agent=builder task=auth"
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
- Dispatch Parameters section present
- Tool Schema Quick Reference present
- Common mistakes (text vs content, mode misuse)

### Validate Skills

```bash
lazyai-cli validate skills
```

Checks:
- Skills directory exists
- Skill structure validation (basic checks; expanding)

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

Legacy `workflow` and orchestration CLI commands were removed during the runtime refactor.

See `../migration/fortnite-orchestrator-removal.md` for replacements and rollback guidance.

## Agent Message Bus

SQLite-based messaging between agents.

### Send a Message

```bash
lazyai-cli message send builder "Need help" "Can you review the auth code?" --priority high
```

### Receive Messages

```bash
lazyai-cli message recv builder
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
