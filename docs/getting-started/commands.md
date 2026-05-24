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
- sqlite3 version
- git version
- jq availability
- bash version
- ollama responding (localhost:11434)
- openai API key configured
- disk space usage
- orchestrator binary on PATH

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

Coming soon: Skill structure validation.

## Workflow

Typical workflow using all commands:

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
