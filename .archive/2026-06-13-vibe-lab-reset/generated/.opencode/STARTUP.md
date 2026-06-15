# LazyAI Startup Protocol

## Mandatory Startup Sequence

At the beginning of EVERY session, the orchestrator MUST execute these steps in order:

### Step 1: Health Check
```bash
lazyai-cli doctor
```
If health check fails (exit code != 0), report failures and ask user whether to proceed.

### Step 2: Session Start
```bash
lazyai-cli session start "<brief goal description>"
```
This creates a session record in SQLite and a session_start entry in the ledger.
Capture the Session ID — it's needed for all subsequent dispatches.

### Step 3: Knowledge Injection
```bash
# Load current vault context: active goals, learning focus, ideas in progress
# This is agent-specific and may involve reading memory files
```

### Step 4: Ledger Verification
```bash
lazyai-cli ledger verify
```
Confirms the hash chain is intact from previous sessions.

### Step 5: Check Messages
```bash
# Check for any unread inter-agent messages from previous sessions
# This is agent-specific
```

## Session Context

After initialization, the orchestrator should report:
- Session ID: {SID}
- Health status: {pass/fail/warn count}
- Ledger entries: {count}, chain: {intact/broken}
- Active goals from vault: {list}
- Unread messages: {count}

## Rules

- Never skip the health check — a broken system should not start work
- Always create a session — untracked work is lost work
- If ledger verification fails, escalate immediately
- If knowledge injection fails, proceed with warning (vault may be unavailable)

## Quick Reference

| Topic | File |
|-------|------|
| Agent registry, dispatch matrix, safety rules | `.opencode/AGENTS.md` |
| Fallback chains and model priority | `.opencode/MODEL-ROUTING.md` |
| Workflow definitions | `.opencode/workflows/` |
