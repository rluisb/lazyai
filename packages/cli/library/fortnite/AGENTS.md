# Agent Registry — Fortnite Reboot v3.1 (Multi-Agent)

8 agents with explicit parameter contracts. All agents can dispatch to each other via the `task` tool. Agent files live in: `~/.config/opencode/agents/`

## The Squad

| Agent | File | Skill | Model (Primary) | Think | Role |
|-------|------|-------|-----------------|-------|------|
| **loop-driver** | `loop-driver.md` | `battle-bus` | `ollama-cloud/kimi-k2.6:cloud` | — | Top router — dispatch with params, fallback, CLI tools |
| **engine-control** | `engine-control.md` | `workflow-engine` | `ollama-cloud/kimi-k2.6:cloud` | ✅ | Workflow orchestration — teams, workflows, step execution |
| **loot-hawk** | `loot-hawk.md` | `storm-scout` (Ph.1) | `openai/gpt-5.5-fast` | ✅ | Scout — code + vault research (`ob`, WarpGrep) |
| **turbo-crank** | `turbo-crank.md` | `storm-scout` (full) | `ollama-cloud/deepseek-v4-pro` | ✅ | Specify/Plan — speckit compatible |
| **wall-builder** | `wall-builder.md` | `build-mode` | `ollama-cloud/kimi-k2.6:cloud` | ✅ | Implementor — MODE: junior/standard/senior/tdd |
| **shield-audit** | `shield-audit.md` | `zero-point` | `openai/gpt-5.5` | xhigh | Verifier — MODE: quick/review/security/adversarial/judge (cross-model on security/adversarial) |
| **rift-deploy** | `rift-deploy.md` | none (script-based) | `ollama-cloud/nemotron-3-super` | ✅ | Ops — MODE: dry-run/staging/production. colima + dev |
| **respawn-crew** | `respawn-crew.md` | none (script-based) | `ollama-cloud/kimi-k2.6:cloud` | ✅ | SRE — SEVERITY: P1-P4 with TIMEBOX. colima status |

## Quick Reference

| Topic | File |
|-------|------|
| Dispatch matrix, parallel execution, barriers, locks | `agents/DISPATCH-MATRIX.md` |
| Fallback chains and model priority | `agents/FALLBACK-CHAINS.md` |
| Repo profiles and quality gates | `agents/REPO-PROFILES.md` |
| Parameter catalog and output schemas | `agents/OUTPUT-SCHEMAS.md` |
| Safety rules, autonomy classes, approval requirements | `agents/SAFETY-BOUNDARIES.md` |
| Tool schemas and common mistakes | `agents/TOOL-SCHEMAS.md` |

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

## Skill-to-Agent Mapping

| Skill | Primary Agent | When |
|-------|---------------|------|
| `battle-bus` | loop-driver | Workflow blueprint generation |
| `build-mode` | wall-builder | Implementation (all tiers, TDD) |
| `zero-point` | shield-audit | Verification (all modes) |
| `storm-scout` | turbo-crank | Pre-implementation pipeline (clarify→research→plan) |
| `storm-eye` | shield-audit | AI evaluation framework |
| `truth-chain` | all agents | Immutable append-only ledger |
| `workflow-engine` | engine-control | Runtime workflow orchestration |
| `slurp-juice` | all agents | Session checkpoint/handoff |
| `shield-wall` | wall-builder | Backend anti-slop guardrails |
| `build-fort` | wall-builder | UI anti-slop guardrails |
| `war-council` | engine-control | Multi-advisor debate (explicit-only) |
| `drop-ship` | rift-deploy | PR release workflow |
| `task-queue` | all agents | SQLite-backed multi-agent task queue |
| `ricochet` | wall-builder | Backpropagation from test failures |
| `drift-scope` | shield-audit | Spec vs implementation drift detection |
| `lesson-loot` | all agents | Self-improvement from corrections |

## Speckit Compatibility

| Speckit Command | Fortnite Route |
|-----------------|----------------|
| `/speckit.specify` | turbo-crank MODE=clarify |
| `/speckit.plan` | turbo-crank MODE=plan |
| `/speckit.tasks` | turbo-crank MODE=plan |
| `/speckit.implement` | wall-builder MODE=standard |
| `/speckit.analyze` | shield-audit MODE=review |

Artifacts live in `bee-gone/specs/<NNN-slug>/` as Markdown/YAML files.

## Tool Schema Quick Reference

When dispatching agents or calling tools directly, use the correct field names:

| Tool | Required Fields | Common Mistake |
|------|-----------------|----------------|
| `todowrite` | `content`, `status`, `priority` | Using `text` instead of `content` |
| `bash` | `command`, `description` | Omitting `description` |
| `task` | `description`, `prompt`, `subagent_type` | Using `mode` or `text` as top-level fields |
| `read` | `filePath` (absolute) | Using relative paths |
| `filesystem_edit_file` | `path`, `edits` (with `oldText`/`newText`) | Using `oldString`/`newString` |
| `morph-mcp_edit_file` | `path`, `instruction`, `code_edit` | Omitting `instruction` |
| `compress` | `topic`, `content` (array) | Using `text` instead of `topic` |

See `agents/TOOL-SCHEMAS.md` for full JSON schemas and validation checklist.

## Design Principles

- **Context is a budget.** Every token loaded is a token not available for reasoning. Load only what the current task needs.
- **Default to deterministic execution** when the task has stable rules. LLMs are for high-variance judgment, not for running `npm test`.
- **Prompt files are contracts, not dumping grounds.** Agent files define boundaries. Skill files define procedures.
- **Skills are on-demand procedures, not always-loaded manuals.** Quick reference first, deep mode on request.
- **The ledger is the immutable audit trail.** Append-only, hash-chained, verifiable.
- **The session DB is mutable runtime state.** Queriable, migratable, backed up.
- **Every high-risk action needs a gate.** Human approval for deploy, post-hoc verification for hotfix, dry-run default for backprop.
- **Fine-tuning is a later optimization, not a first fix.** Prompt engineering, deterministic scripts, and model selection come first.
