# Safety Boundaries

## Universal Safety Rules

1. **Never push, merge/rebase, or create branches/worktrees without explicit approval.**
2. **Respect `REPORT_ONLY` and `PLAN_ONLY` modes.**
3. **Record every dispatch to `session-db.sh` for auditability.**
4. **Worktree/branch operations require explicit policy approval.**
5. **No model shared between opposing roles** (implement≠review≠plan).

## Autonomy Classes

| Class | Agents | Can Do | Needs Approval For |
|-------|--------|--------|-------------------|
| **Tier 0 — Deterministic** | scripts | `/test`, `/commit`, health checks | None (no LLM) |
| **Tier 1 — Router** | loop-driver | Dispatch, routing, status | Implementation, deploy |
| **Tier 2 — Read-only** | loot-hawk, shield-audit | Research, review, verify | Writes, edits, commits |
| **Tier 3 — Implement** | wall-builder, engine-control | Code changes, workflow steps | Deploy, merge, branch ops |
| **Tier 4 — Sensitive** | rift-deploy, respawn-crew | Deploy, incident response | Production deploy, rollback |

## Approval Requirements

- **Branch/worktree/checkout/merge/rebase/commit** → explicit user approval
- **Edits** → explicit user approval (or PLAN_ONLY mode)
- **Migrations** → explicit user approval
- **Deploys** → explicit user approval (rift-deploy MODE=production requires human gate)
- **GitHub operations** → explicit user approval

## Deterministic Bypass Policy

Scripts, schemas, SQLite, YAML, locks, and gates handle deterministic work. LLMs handle reasoning, planning, reviewing, and high-variance judgment.

Tier 0 commands (`/test`, `/commit`, health checks) bypass the AI stack entirely.

## Max Dispatch Depth

Maximum dispatch depth: **5**. Circular dispatch (A → B → A) escalates to loop-driver.

## Protected Resources

- `.specify/session.db` — backup before migration
- `.specify/ledger.jsonl` — append-only, hash-chained
- `bee-gone/specs/<slug>/` — durable artifacts, never the only source of truth

## Fine-tuning Policy

- **Do not fine-tune any model** in this implementation.
- Collect datasets only. Block all training until data thresholds are met.
- Use context retrieval, RAG, codegraph, docs, and runtime context instead.

## Ledger and Session DB Policy

- **Ledger**: immutable append-only JSONL with SHA-256 hash chain. Every significant action leaves a trace.
- **Session DB**: mutable SQLite runtime state. WAL mode, migratable, backed up.
- **SQLite is operational/cache/index only** and must never be the only source of truth for durable project artifacts.
