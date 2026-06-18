# Fortnite System Startup Protocol

This file is loaded by OpenCode at session start via the `instructions` field in opencode.jsonc.
The loop-driver agent reads this and executes the initialization sequence.

## Mandatory Startup Sequence

At the beginning of EVERY session, loop-driver MUST execute these steps in order:

### Step 1: Health Check
```bash
scripts/health-check.sh --quick
```
If health check fails (exit code != 0), report failures and ask user whether to proceed.

### Step 2: Session Start
```bash
SID=$(scripts/session-db.sh start "<brief goal description>" "<repo-name>")
```
This creates a session record in SQLite and a session_start entry in the ledger.
Capture the SID — it's needed for all subsequent dispatches.

### Step 3: Knowledge Injection
```bash
scripts/knowledge-inject.sh light
```
Loads current vault context: active goals, learning focus, ideas in progress.

### Step 4: Ledger Verification
```bash
skills/truth-chain/scripts/ledger.sh verify
```
Confirms the hash chain is intact from previous sessions.

### Step 5: Check Messages
```bash
scripts/agent-msg.sh recv loop-driver
```
Process any unread inter-agent messages from previous sessions.

## Session Context

After initialization, loop-driver should report:
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
| Agent registry, dispatch matrix, safety rules | `AGENTS.md` |
| Fallback chains and model priority | `agents/FALLBACK-CHAINS.md` |
| Repo profiles and quality gates | `agents/REPO-PROFILES.md` |
| Parameter catalog and output schemas | `agents/OUTPUT-SCHEMAS.md` |
| Safety boundaries and approval requirements | `agents/SAFETY-BOUNDARIES.md` |
