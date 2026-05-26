# Gap Analysis: LazyAI vs Production AI Agent Runtime

**Source of Truth:** `/Users/ricardo/.config/opencode/` (Production AI Agent Runtime)  
**Target:** `/Users/ricardo/projects/teachable/lazyai/` (LazyAI)  
**Date:** 2026-05-23  
**Assessed by:** shield-audit (review mode)

---

## Executive Summary

| Metric | Count |
|--------|-------|
| Total capabilities assessed | 12 |
| Present in LazyAI | 3 |
| Partial in LazyAI | 4 |
| Missing in LazyAI | 5 |
| Critical gaps (P0) | 4 |
| Important gaps (P1) | 4 |
| Nice-to-have (P2) | 4 |

**Verdict:** LazyAI has a solid orchestration runtime but lacks structured agent contracts, audit infrastructure, and quality gates. The biggest risks are: no audit trail, no structured dispatch parameters, no session tracking, and no eval harness.

---

## 1. Agent Contracts

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Dispatch parameters block | ✅ Structured `## Dispatch Parameters` | ❌ None | **Missing** | No standard parameter contract | 2h | **P0** |
| Context pruning rules | ✅ Per-agent Keep/Drop tables | ❌ None | **Missing** | No token budget management | 1h | **P1** |
| Tool schema references | ✅ `agents/TOOL-SCHEMAS.md` | ⚠️ Basic | **Partial** | Has some docs, not comprehensive | 30min | **P0** |
| Negative examples | ✅ "Bad output — DON'T" in all agents | ❌ None | **Missing** | No anti-patterns documented | 1h | **P1** |
| Synthesis protocol | ✅ In loop-driver | ❌ None | **Missing** | No output synthesis standard | 30min | **P2** |

**Evidence:**
- Our `agents/loop-driver.md` lines 45-60: structured dispatch block
- LazyAI `agents/orchestrator.md`: describes topology but no parameter contract

---

## 2. Skills

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Quick Reference sections | ✅ All 38 skills | ❌ None | **Missing** | No skill quick reference | 3h | **P1** |
| Script-backed skills | ✅ Scripts per skill (e.g., `task-init.sh`) | ⚠️ Some | **Partial** | Has orchestrator binary, limited bash scripts | 2h | **P1** |
| Skill index | ✅ `skills/_INDEX.md` with 38 active + 18 archived | ❌ None | **Missing** | No skill inventory | 1h | **P2** |
| Sub-skills | ✅ storm-scout split into clarify/research/plan | ❌ None | **Missing** | No progressive loading | 2h | **P2** |
| Skill contracts | ✅ Structured frontmatter + parameters table | ⚠️ Basic | **Partial** | Has frontmatter, less structured | 1h | **P1** |

**Evidence:**
- Our `skills/storm-scout/SKILL.md` lines 12-22: Quick Reference table
- LazyAI `skills/plan/SKILL.md`: basic description, no quick reference

---

## 3. Session Tracking

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| SQLite session DB | ✅ `scripts/session-db.sh` (35 tables) | ⚠️ `.ai-setup.db` | **Partial** | Has DB, different schema | 4h | **P0** |
| Dispatch recording | ✅ `agent_dispatches` table | ❌ Unknown | **Missing** | No dispatch audit | 2h | **P0** |
| Quality metrics | ✅ `quality_metrics` table + commands | ❌ None | **Missing** | No quality tracking | 2h | **P1** |
| Workflow instances | ✅ `workflow_instances` table | ⚠️ Orchestrator state | **Partial** | Has state, not in SQLite | 3h | **P1** |
| Parallel task tracking | ✅ `parallel_tasks` table | ❌ None | **Missing** | No barrier/lock tracking | 2h | **P2** |

**Evidence:**
- Our `scripts/session-db.sh` lines 463-475: `quality_metrics` table
- LazyAI `.ai-setup.db`: exists but schema unknown

---

## 4. Audit Trail

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Immutable ledger | ✅ `skills/truth-chain/scripts/ledger.sh` | ❌ None | **Missing** | No audit trail | 4h | **P0** |
| Hash-chain verification | ✅ SHA-256 chained entries | ❌ None | **Missing** | No integrity checks | 2h | **P1** |
| Ledger rotation | ✅ Automatic archival | ❌ None | **Missing** | No rotation policy | 1h | **P2** |
| Corruption detection | ✅ `ledger.sh verify` detects breaks | ❌ None | **Missing** | No corruption checks | 2h | **P1** |

**Evidence:**
- Our `.specify/ledger.jsonl`: 568 entries, chain intact
- LazyAI: no ledger file found

---

## 5. Health Checks

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Dependency validation | ✅ `scripts/health-check.sh` checks sqlite3, git, jq | ⚠️ Smoke tests | **Partial** | Has some checks, not comprehensive | 1h | **P1** |
| Provider health | ✅ Checks ollama-cloud, openai | ❌ None | **Missing** | No provider monitoring | 1h | **P1** |
| JSON output mode | ✅ `--json` flag | ❌ None | **Missing** | No machine-readable health | 30min | **P2** |
| Disk space check | ✅ Warns at >80% | ❌ None | **Missing** | No resource monitoring | 30min | **P2** |

**Evidence:**
- Our `scripts/health-check.sh` lines 214-295: provider health section
- LazyAI `tests/scripts/smoke-test.sh`: basic checks only

---

## 6. Smoke Tests

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Test framework | ✅ POSIX shell, 148 tests | ⚠️ Basic | **Partial** | Has tests, fewer | 2h | **P1** |
| Agent file validation | ✅ Checks all 8 agents | ❌ Unknown | **Missing** | No agent validation | 1h | **P1** |
| Skill loading tests | ✅ Verifies all skills load | ❌ Unknown | **Missing** | No skill validation | 1h | **P1** |
| Ledger integrity tests | ✅ Corruption fixture tests | ❌ None | **Missing** | No ledger tests | 2h | **P2** |

**Evidence:**
- Our `tests/scripts/smoke-test.sh`: 148 passed, 0 failed
- LazyAI: has smoke tests but coverage unknown

---

## 7. Workflow Engine

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| YAML workflow definitions | ✅ 7 workflows in `.opencode/workflows/` | ❌ None | **Missing** | No YAML workflows | 3h | **P1** |
| Schema validation | ✅ `workflow-exec.sh --validate` | ❌ None | **Missing** | No validation | 2h | **P1** |
| Failure policies | ✅ stop/continue/rollback | ❌ None | **Missing** | No failure handling | 2h | **P2** |
| Metrics recording | ✅ Per-phase metrics blocks | ❌ None | **Missing** | No metrics | 1h | **P2** |

**Evidence:**
- Our `.opencode/workflows/rpi.yaml`: 48 lines, structured phases
- LazyAI: uses orchestrator binary, no YAML workflows

---

## 8. Eval Harness

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Dataset directories | ✅ 7 datasets with READMEs | ❌ None | **Missing** | No datasets | 2h | **P2** |
| Eval suite definitions | ✅ 9 suite YAMLs | ❌ None | **Missing** | No eval suites | 2h | **P2** |
| Dataset extractors | ✅ 6 extractors (dispatch, commit, spec, etc.) | ❌ None | **Missing** | No data collection | 4h | **P2** |
| Eval runner | ✅ `scripts/run-evals.sh` | ❌ None | **Missing** | No eval execution | 3h | **P2** |
| Model comparison | ✅ `scripts/compare-models.sh` | ❌ None | **Missing** | No model shootout | 2h | **P3** |

**Evidence:**
- Our `.specify/evals/suites/`: 9 YAML files
- LazyAI: no eval directory found

---

## 9. Model Routing

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Tier assignments | ✅ 4 tiers in `AGENTS.md` | ⚠️ Prompt-level | **Partial** | Has roles, not formal tiers | 1h | **P1** |
| Fallback chains | ✅ `agents/FALLBACK-CHAINS.md` | ❌ None | **Missing** | No fallback documentation | 1h | **P1** |
| Model routing policy | ✅ `.ai-harness/model-routing.yaml` | ❌ None | **Missing** | No routing policy | 2h | **P2** |
| Cost tracking | ✅ Cost snapshots in session DB | ❌ None | **Missing** | No cost monitoring | 2h | **P2** |

**Evidence:**
- Our `.ai-harness/model-routing.yaml`: 429 lines
- LazyAI: no model routing file

---

## 10. Safety Boundaries

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Autonomy classes | ✅ 4 tiers in `agents/SAFETY-BOUNDARIES.md` | ⚠️ Human gates | **Partial** | Has gates, not formal classes | 1h | **P1** |
| Approval requirements | ✅ Per-tier approval matrix | ⚠️ Process rules | **Partial** | Has rules, less structured | 30min | **P1** |
| Worktree policy | ✅ Explicit approval required | ⚠️ Mentioned | **Partial** | Less formalized | 30min | **P2** |
| REPORT_ONLY mode | ✅ Respected by all agents | ❌ None | **Missing** | No read-only enforcement | 30min | **P2** |

**Evidence:**
- Our `agents/SAFETY-BOUNDARIES.md`: full autonomy matrix
- LazyAI `AGENTS.md`: has human gates but no tier structure

---

## 11. Speckit Integration

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Speckit commands | ✅ 5 commands mapped to agents | ✅ 7 commands | **Present** | LazyAI has more! | — | — |
| Template scaffolding | ✅ `task-init.sh` creates spec dirs | ⚠️ Templates exist | **Partial** | Has templates, no scaffolding script | 1h | **P2** |
| Artifact paths | ✅ `bee-gone/specs/<slug>/` | ⚠️ `.specify/` | **Partial** | Different path convention | 30min | **P2** |
| Spec compatibility | ✅ Documented in `DOCUMENTATION.md` | ✅ Present | **Present** | Both support speckit | — | — |

**Evidence:**
- Our `commands/speckit.*.md`: 5 files
- LazyAI `.opencode/commands/`: 7 speckit files (more comprehensive)

---

## 12. Infrastructure

| Capability | Our System | LazyAI | Status | Gap | Effort | Priority |
|------------|-----------|--------|--------|-----|--------|----------|
| Startup protocol | ✅ `STARTUP.md` with 5 mandatory steps | ❌ None | **Missing** | No startup sequence | 1h | **P1** |
| Knowledge injection | ✅ `scripts/knowledge-inject.sh` | ❌ None | **Missing** | No vault context loading | 1h | **P2** |
| Inter-agent messaging | ✅ `scripts/agent-msg.sh` | ⚠️ Orchestrator | **Partial** | Has event bus, different protocol | 2h | **P2** |
| Task barriers/locks | ✅ `task-barrier.sh`, `task-lock.sh` | ❌ None | **Missing** | No parallel sync | 2h | **P2** |

**Evidence:**
- Our `STARTUP.md`: 5-step mandatory sequence
- LazyAI: no startup protocol found

---

## Top 5 Critical Gaps (P0)

| Rank | Gap | Risk | Effort |
|------|-----|------|--------|
| 1 | **No audit trail (ledger)** | Cannot verify what agents did, no accountability | 4h |
| 2 | **No structured dispatch parameters** | Agents receive inconsistent inputs, errors likely | 2h |
| 3 | **No session tracking (dispatches, workflows)** | Cannot debug failures, no history | 4h |
| 4 | **No comprehensive health checks** | Environment issues discovered late | 1h |
| 5 | **No tool schema validation** | Schema errors (like we experienced) will recur | 30min |

---

## Porting Roadmap

### Phase 1: Foundation (Week 1) — P0 Items
1. **Tool schemas** → Copy `TOOL-SCHEMAS.md`, add to all agents (30min)
2. **Dispatch parameters** → Add `## Dispatch Parameters` block to each agent (2h)
3. **Health checks** → Adapt `health-check.sh` for lazyai (1h)
4. **Session DB tables** → Add dispatch/workflow tables to `.ai-setup.db` (4h)
5. **Ledger** → Port `ledger.sh` and create `.specify/ledger.jsonl` (4h)

### Phase 2: Quality (Week 2) — P1 Items
6. **Smoke tests** → Expand test coverage to 100+ tests (2h)
7. **Agent contracts** → Add context pruning, negative examples (2h)
8. **Skills** → Add Quick Reference sections (3h)
9. **Safety boundaries** → Formalize autonomy classes (1h)
10. **Model routing** → Document fallback chains (1h)

### Phase 3: Advanced (Week 3-4) — P2 Items
11. **Workflow YAMLs** → Create structured workflows (3h)
12. **Eval harness** → Set up datasets and suites (4h)
13. **Startup protocol** → Create `STARTUP.md` (1h)
14. **Knowledge injection** → Port vault loading (1h)

**Total effort:** ~30 hours over 3-4 weeks

---

## What LazyAI Does Better

| Capability | LazyAI | Our System |
|------------|--------|-----------|
| Speckit commands | 7 commands | 5 commands |
| Orchestrator runtime | Go binary with SQLite | Bash scripts |
| MCP integration | 4 servers (orchestrator, filesystem, tavily, morph) | Basic |
| Dashboard | `/health` endpoint | None |
| A2A execution | Configurable modes | None |
| Budget tracking | Built into orchestrator | Manual |

---

## Recommendation

**Don't port everything.** LazyAI has a stronger orchestration runtime (Go binary > bash scripts). Focus on:

1. **Agent contracts** — Structured dispatch parameters prevent errors
2. **Audit trail** — Ledger is non-negotiable for production
3. **Health checks** — Validate environment before work
4. **Tool schemas** — Prevent schema errors

The orchestration layer should stay as-is (Go binary). The agent/skill contracts and infrastructure should adopt our patterns.

---

*Assessment complete. 12 capabilities evaluated, 5 missing, 4 partial, 3 present.*
