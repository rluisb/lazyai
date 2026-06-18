---
name: truth-chain
description: Immutable append-only ledger with hash-chain verification. Tamper-proof audit trail for all agent actions, decisions, and state changes. Every entry links to the previous via SHA-256.
trigger: /truth-chain, /ledger
skill_path: skills/truth-chain
scripts:
  - name: ledger.sh
    description: Append-only ledger with hash-chain verification, replay, and audit
    path: skills/truth-chain/scripts/ledger.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | Ledger verification, audit trail, immutable logging |
| **Do not use when** | Mutable runtime state, session management |
| **Primary agent** | all agents |
| **Runtime risk** | Low — append-only ledger |
| **Outputs** | Hash-chained entries, verification reports |
| **Validation** | SHA-256 chain integrity |
| **Deep mode trigger** | `/truth-chain` or ledger audit |

# Truth-Chain — Immutable Append-Only Ledger

## Purpose
Session-db is mutable — records can be edited or deleted. Truth-Chain is **append-only** with a **SHA-256 hash chain**. Every entry cryptographically links to the previous one. Tamper with any entry and the entire chain breaks. Ground truth for the Fortnite squad.

**Use when:**
- "Record this decision immutably" — `ledger.sh append`
- "Verify nothing was tampered with" — `ledger.sh verify`
- "Replay session state at entry N" — `ledger.sh replay`
- "Show me the audit trail" — `ledger.sh audit`

## Ledger File

```
.specify/ledger.jsonl
```

One JSON object per line. Each entry includes:
- `seq` — Sequential number (auto-increment)
- `ts` — ISO 8601 timestamp
- `type` — Entry type (see below)
- `session_id` — Session identifier
- `data` — Payload (varies by type)
- `prev_hash` — SHA-256 of previous entry (genesis = "0000...")
- `hash` — SHA-256 of this entry (computed after prev_hash set)

## Entry Types

| Type | Agent | Data Fields |
|------|-------|-------------|
| `session_start` | loop-driver | goal, repo, model, worktree |
| `session_end` | loop-driver | status, token_total, summary |
| `dispatch` | loop-driver | agent, task, phase, mode, workflow |
| `dispatch_complete` | any | seq, result, summary, files_touched |
| `dispatch_fail` | any | seq, error_message |
| `decision` | any | title, decision, rationale, alternatives, tags |
| `memory` | any | title, content, tags, importance |
| `artifact` | any | path, action |
| `barrier_create` | loop-driver | barrier_id, expected_count |
| `barrier_resolve` | loop-driver | barrier_id, arrived_count |
| `lock_acquire` | any | lock_name, held_by |
| `lock_release` | any | lock_name, held_by |
| `msg_send` | any | from_agent, to_agent, subject, priority |
| `msg_read` | any | msg_id, from_agent, subject |
| `quality_gate` | shield-audit | gate_type, repo, passed, duration_ms, errors, warnings |
| `drift_violation` | shield-audit | spec_section, claim, actual, severity |
| `backprop_update` | wall-builder | spec_section, old_value, new_value, trigger_test |
| `token_log` | any | token_count, context_pct |
| `ptask_register` | loop-driver | wave_id, agent, task |
| `ptask_complete` | any | ptask_id, result, output_path |
| `ptask_fail` | any | ptask_id, error_message |
| `workflow_start` | engine-control | workflow_name, total_steps |
| `workflow_step_start` | engine-control | step_order, agent, task, mode |
| `workflow_step_done` | engine-control | step_order, result, output_path |
| `workflow_fail` | engine-control | step_order, error_message |

## Ledger v2 Fields

Backward-compatible additions to the base entry schema:

| Field | Type | Description |
|-------|------|-------------|
| `workflow_run_id` | string | Links entry to a workflow instance |
| `agent` | string | Agent that produced the entry |
| `risk` | string | `low`, `medium`, `high`, `critical` |
| `input_hash` | string | SHA-256 of input material |
| `output_hash` | string | SHA-256 of output material |
| `redactions` | JSON | Metadata about redacted fields (optional) |

Hash policy: `hash = SHA-256(seq + ts + type + session_id + JSON(data) + prev_hash)` where `JSON(data)` preserves Unicode characters (`ensure_ascii=False`). This matches the historical append policy and prevents canonicalization mismatches.

## Secret Redaction

Use `--redact` when appending entries that may contain secrets:

```bash
./skills/truth-chain/scripts/ledger.sh append --type decision \
  --data '{"api_key":"sk-live-xxx","password":"hunter2","title":"infra"}' \
  --session ses_abc --redact
```

Redacted keys (case-insensitive): `token`, `access_token`, `refresh_token`, `auth_token`, `bearer_token`, `api_token`, `jwt`, `secret`, `password`, `api_key`, `apikey`, `key`, `private_key`, `client_secret`, `client_id`, `auth_code`, `credential`, `credentials`.

Values are replaced with `[REDACTED]`. The hash is computed over the redacted payload so the chain remains verifiable.

## Locking

Lock priority (automatic, no manual steps):
1. `flock` — Linux/util-linux (advisory file lock)
2. `shlock` — macOS/BSD (PID-based file lock)
3. `mkdir` — Portable atomic test-and-set fallback

## Replay to Session DB

```bash
./skills/truth-chain/scripts/ledger.sh replay-to-session-db
```

Prefers `./scripts/session-db.sh record-ledger-ref` when available. Falls back to direct SQL with a warning if the command is missing. Skips entries already present in `ledger_refs`.

## Backup / Rotate / Prune / Archive

```bash
./skills/truth-chain/scripts/ledger.sh backup    # Copy to .specify/backups/
./skills/truth-chain/scripts/ledger.sh rotate    # Rotate when >10MB
./skills/truth-chain/scripts/ledger.sh prune     # Keep last N entries, archive rest
./skills/truth-chain/scripts/ledger.sh archive   # Compress to .jsonl.gz
```

## Fixtures / Tests

Smoke tests in `tests/scripts/smoke-test.sh` cover:
- Ledger verify (real + fixture variants: corrupted hash, prev_hash, gap, truncated)
- Secret redaction temp-ledger test (proves raw secrets absent)
- Replay-to-session-db temp DB test (proves command path and insert count)

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `ledger.sh` | Append-only ledger with hash-chain verification | `append`, `verify`, `replay`, `query`, `audit`, `head`, `tail`, `chain`, `backup`, `rotate`, `prune`, `archive`, `replay-to-session-db` |

## Workflow

### Step 1: Initialize Ledger
```bash
./skills/truth-chain/scripts/ledger.sh init
# Creates .specify/ledger.jsonl with genesis entry
```

### Step 2: Append Entries
```bash
# Session start
./skills/truth-chain/scripts/ledger.sh append session_start '{"goal":"Add auth middleware","repo":"fedora","model":"ollama-cloud/deepseek-v4-pro"}'

# Dispatch
./skills/truth-chain/scripts/ledger.sh append dispatch '{"agent":"wall-builder","task":"Implement auth middleware","phase":"implement","mode":"standard"}'

# Decision
./skills/truth-chain/scripts/ledger.sh append decision '{"title":"Use JWT","decision":"Bearer tokens","rationale":"Stateless auth","alternatives":"Session cookies","tags":"auth"}'

# Quality gate
./skills/truth-chain/scripts/ledger.sh append quality_gate '{"gate_type":"test","repo":"fedora","passed":1,"duration_ms":4500,"errors":0,"warnings":2}'

# With v2 metadata and redaction
./skills/truth-chain/scripts/ledger.sh append --type decision \
  --data '{"api_key":"sk-live-xxx","password":"hunter2","title":"infra"}' \
  --session ses_abc --agent wall-builder --risk medium --redact
```

### Step 3: Verify Integrity
```bash
./skills/truth-chain/scripts/ledger.sh verify
# Checks entire hash chain. Returns 0 if valid, 1 if tampered.
```

### Step 4: Query Entries
```bash
# All dispatches by wall-builder
./skills/truth-chain/scripts/ledger.sh query --type dispatch --agent wall-builder

# All decisions tagged "auth"
./skills/truth-chain/scripts/ledger.sh query --type decision --tag auth

# All entries for a session
./skills/truth-chain/scripts/ledger.sh query --session ses_abc123
```

### Step 5: Replay State
```bash
# Reconstruct state at entry 15
./skills/truth-chain/scripts/ledger.sh replay 15
```

### Step 6: Replay to Session DB
```bash
./skills/truth-chain/scripts/ledger.sh replay-to-session-db
```

### Step 7: Audit Report
```bash
# Full audit for a session
./skills/truth-chain/scripts/ledger.sh audit --session ses_abc123 --output audit-report.json
```

## Integration with Other Skills

- **battle-bus**: Records all dispatches, barriers, parallel tasks to ledger
- **build-mode**: Logs implementation actions, quality gates, ricochet backprop updates
- **zero-point**: Records verification results, drift violations, audit outcomes
- **slurp-juice**: Checkpoint events logged to ledger for tamper-proof session state
- **workflow-engine**: Workflow instances, step transitions, failures all recorded
- **ricochet**: Backpropagation updates logged with trigger test reference
- **drift-scope**: Drift violations recorded with spec section and severity

## Hash Chain Verification

Each entry's `hash` is computed as:
```
hash = SHA-256(seq + ts + type + session_id + JSON(data) + prev_hash)
```

The next entry's `prev_hash` = this entry's `hash`.

If any entry is modified, its hash changes, breaking the chain for all subsequent entries. `ledger.sh verify` detects this immediately.

## Tips

- Ledger is append-only — no delete, no update, no overwrite
- Genesis entry has `prev_hash: "0000000000000000000000000000000000000000000000000000000000000000"`
- Run `ledger.sh verify` after any session to confirm integrity
- Ledger file grows linearly — compress old sessions with `ledger.sh archive`
- Use `ledger.sh head` / `ledger.sh tail` for quick inspection
- Ledger and session-db are complementary: session-db for queries, ledger for proof
