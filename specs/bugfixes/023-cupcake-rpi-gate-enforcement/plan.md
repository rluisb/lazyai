# Plan: RPI Human Gate Enforcement via Cupcake + Hardening

**Bugfix:** 023 — Cupcake RPI Gate Enforcement
**Phase:** Plan (P of RPI)
**Started:** 2026-05-04
**Confidence:** HIGH

---

## 1. Acceptance Criteria (What "Fixed" Means)

1. **A model in Claude Code auto/accept-edits mode requesting RPI will be physically prevented from writing code during research and plan phases.** The tool runtime (via cupcake hook) blocks file writes to `src/` until a plan.md with attested human gate exists.

2. **A model in Copilot agent mode requesting RPI will be caught at commit/PR time.** The pre-commit hook (calling `cupcake evaluate`) rejects commits from non-trivial changes lacking gate attestation. CI provides a second checkpoint.

3. **RPI skill text is hardened with mode-awareness.** The RPI skill detects auto/agent mode and refuses to proceed past research without explicit human approval. The skill text itself becomes a last-resort defense when cupcake is not installed.

4. **Cross-tool consistency.** A single canonical gate definition (in AGENTS.md + cupcake Rego policies) governs all tools. Tool-specific files are generated/wired from this canonical source.

5. **Forging is detectably hard.** Gate attestation markers include git authorship verification and timestamp correlation. A model cannot trivially forge "Human Gate: APPROVED" and pass checks.

---

## 2. Implementation Phases

### Phase 1 — Cupcake Installation & Policy Authoring (P1)

**Goal:** Get cupcake running with core RPI gate policies.

| Task | Description | Files | Done When |
|------|-------------|-------|-----------|
| T001 | Install cupcake CLI and verify | — | `cupcake --version` reports 0.5.x |
| T002 | Create `cupcake.yml` with signal definitions | `cupcake.yml` | Signals: `plan_attested`, `gate_attested`, `lines_changed`, `review_approved` |
| T003 | Author `policies/claude/rpi_gates.rego` | `policies/claude/rpi_gates.rego` | Policy blocks writes to `src/` when no plan.md with attestation exists. `require_review` on commits >20 lines without gate. |
| T004 | Author `policies/opencode/rpi_gates.rego` | `policies/opencode/rpi_gates.rego` | Same logic, OpenCode-specific tool names |
| T005 | Author `policies/copilot/rpi_gates.rego` (signals-based) | `policies/copilot/rpi_gates.rego` | Commit-time enforcement via signals. No native hook — relies on pre-commit/CI integration. |
| T006 | Test cupcake `allow` path | — | When plan.md HAS attestation, writes proceed normally |
| T007 | Test cupcake `require_review` path | — | When plan.md MISSING attestation, writes are blocked with feedback message |
| T008 | Test cupcake with Claude Code | — | Claude Code session: write attempt blocked, feedback message shown, agent can self-correct |

### Phase 2 — Tool Configuration Hardening (P2)

**Goal:** Wire cupcake into each tool's configuration. Update instruction files.

| Task | Description | Files | Done When |
|------|-------------|-------|-----------|
| T009 | Wire cupcake into Claude Code | `.claude/settings.json` | `customInstructions` references cupcake. Cupcake watch daemon configured. |
| T010 | Wire cupcake into OpenCode | `.opencode/opencode.jsonc`, `cupcake-plugins/opencode/` | OpenCode plugin active. Mode permissions remain. |
| T011 | Create Copilot pre-commit integration | `.husky/pre-commit`, `.github/workflows/rpi-gate-check.yml` | `cupcake evaluate` runs on pre-commit and CI |
| T012 | Harden Copilot agent YAMLs | `.github/agents/rpi-researcher.agent.yaml`, `rpi-planner.agent.yaml`, `rpi-implementor.agent.yaml` | Phase-specific tool restrictions: researcher=read-only, planner=specs-write-only, implementor=full |
| T013 | Create Copilot instructions | `.github/copilot-instructions.md`, `.github/instructions/rpi.instructions.md` | Instructions reference cupcake enforcement and gate protocol |

### Phase 3 — RPI Artifact Hardening (P3)

**Goal:** Update RPI skill text to be mode-aware. These changes are defense-in-depth — they work even when cupcake is not installed.

| Task | Description | Files | Done When |
|------|-------------|-------|-----------|
| T014 | Add mode-detection preamble to RPI skill | `.agents/skills/rpi/SKILL.md`, `.claude/skills/rpi/SKILL.md`, `.opencode/skills/rpi/SKILL.md` | Skill begins with mode check. If auto/agent mode detected: "I'm in $MODE mode but RPI requires human gates. I will proceed ONLY through Research phase and then halt for your approval." |
| T015 | Add anti-forgery guidance to rpi-workflow.md | `packages/cli/library/fragments/rpi-workflow.md` | Gate attestation section explains: "The Human Gate: APPROVED marker must be written by a human. AI-generated approval text will be rejected by cupcake signals check." |
| T016 | Update AGENTS.md with gate override block | `AGENTS.md` | Appended ⛔ HARD PROCESS GATE block with mode-awareness. References cupcake enforcement. |

### Phase 4 — Verification & Documentation (P4)

**Goal:** Verify end-to-end. Produce documentation.

| Task | Description | Files | Done When |
|------|-------------|-------|-----------|
| T017 | End-to-end test: Claude Code auto mode + RPI | — | Attempt RPI in auto mode. Verify research completes, plan blocked by cupcake until human gate approved. |
| T018 | End-to-end test: Copilot agent mode + RPI | — | Attempt RPI in agent mode. Verify pre-commit hook catches missing gate attestation. |
| T019 | End-to-end test: Interactive mode (normal flow) | — | Verify normal RPI flow is NOT blocked. Gates work as expected. No false positives. |
| T020 | Update orchestration catalog | `.ai/orchestration/chains/feature.json` | Enable orchestrator MCP. Verify `gate: "user_approval"` now programmatically enforced. |
| T021 | Produce setup docs for end users | `docs/cupcake-rpi-setup.md` | Clear instructions for installing cupcake and enabling gate enforcement |

---

## 3. File Manifest

### New Files (11)

```
cupcake.yml                                          ← Cupcake signal configuration
policies/claude/rpi_gates.rego                       ← Claude Code policy
policies/opencode/rpi_gates.rego                     ← OpenCode policy
policies/copilot/rpi_gates.rego                      ← Copilot signals-based policy
.github/agents/rpi-researcher.agent.yaml             ← Read-only research agent
.github/agents/rpi-planner.agent.yaml                ← Plan-only agent (write specs/ only)
.github/agents/rpi-implementor.agent.yaml            ← Full-access implementor (gated)
.github/copilot-instructions.md                      ← Hardened Copilot instructions
.github/instructions/rpi.instructions.md             ← Path-specific RPI gate instructions
.github/workflows/rpi-gate-check.yml                 ← CI gate verification
docs/cupcake-rpi-setup.md                            ← End-user setup guide
```

### Updated Files (8)

```
AGENTS.md                                            ← Append HARD PROCESS GATE block
.claude/settings.json                                ← Add customInstructions + permissions
.claude/skills/rpi/SKILL.md                          ← Add mode-detection preamble
.opencode/skills/rpi/SKILL.md                        ← Add mode-detection preamble
.agents/skills/rpi/SKILL.md                          ← Add mode-detection preamble
.opencode/opencode.jsonc                             ← Wire cupcake plugin
packages/cli/library/fragments/rpi-workflow.md       ← Add anti-forgery guidance
.ai/orchestration/chains/feature.json                ← Enable orchestrator MCP
.ai/mcp.json                                         ← Enable orchestrator server
```

### Spec Files (2)

```
specs/bugfixes/023-cupcake-rpi-gate-enforcement/research.md   ← This research
specs/bugfixes/023-cupcake-rpi-gate-enforcement/plan.md       ← This plan
```

---

## 4. Task Dependency Graph

```
Phase 1: Cupcake Core
  T001 (install) ──┬── T002 (signals) ──┬── T003 (claude policy) ── T006-008 (test)
                   │                    ├── T004 (opencode policy)
                   │                    └── T005 (copilot policy)
                   │
Phase 2: Tool Config (depends on Phase 1 test passing)
  T006-008 ────────┬── T009 (claude wire)
                   ├── T010 (opencode wire)
                   └── T011-013 (copilot wire + instructions)
                   
Phase 3: Artifact Hardening (parallel with Phase 2)
  T014 (mode detection ×3) ─── parallel
  T015 (anti-forgery)      ─── parallel
  T016 (AGENTS.md block)   ─── parallel

Phase 4: Verification (depends on Phase 2 + 3)
  T017-T021 ─── sequential (dependency on previous passes)
```

---

## 5. Deliberately NOT Added

Per Articles IV (YAGNI), V (Simplicity), and VI (Anti-Overengineering):

- **NOT adding** a custom Cupcake harness for Copilot — the native hook doesn't exist, and building one would be speculative. Signals-based enforcement at commit/PR time is sufficient.
- **NOT adding** a web dashboard for policy management — cupcake's CLI and git-committed policies are the simplest thing that works.
- **NOT adding** per-developer policy variants — one canonical policy set is simpler than per-user customization.
- **NOT adding** automated policy generation from AGENTS.md — the policy is stable once written; auto-generation is overengineering for a single policy.
- **NOT creating** a Claude Code `.claude/rules/rpi-gates.md` scoped rule — the cupcake hook is the enforcement mechanism. Adding a text rule would be redundant noise.

---

## 6. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-----------|--------|------------|
| Cupcake v0.5.1 has breaking changes before v1.0 | MEDIUM | HIGH | Pin version in setup docs. Test on each upgrade. |
| Copilot adds native hooks — our signals approach becomes legacy | LOW | LOW | We'd migrate to native when available. Signals approach still works. |
| `require_review` UX is confusing in Claude Code | MEDIUM | MEDIUM | T006-008 test the UX. Clear feedback message in Rego policy. |
| Pre-commit hook slows down workflow | LOW | MEDIUM | `cupcake evaluate` is sub-ms. Only runs on non-trivial changes. Trivial threshold (<20 lines) bypasses. |
| Team resistance to installing extra tool | MEDIUM | MEDIUM | Defense-in-depth: hardened RPI text works without cupcake. Cupcake is additive, not mandatory. |

---

## 7. Constitution Check

| Article | Check | Status |
|---------|-------|--------|
| I — Library-First | Using cupcake (existing policy engine) rather than building custom enforcement | PASS |
| II — Test-First | T006-008, T017-019 are explicit test tasks before declaring done | PASS |
| III — Docs as Source | Research.md + Plan.md written before implementation begins | PASS |
| IV — YAGNI | See "Deliberately NOT Added" section above | PASS |
| V — Simplicity | 21 tasks, 11 new files, 8 updates — minimal surface for the problem | PASS |
| VI — Anti-Overengineering | No custom harness, no dashboard, no auto-generation | PASS |

---

## 8. Verdict

**GO** — The plan addresses the root cause (advisory text vs. system-level mode override) with a deterministic enforcement layer (Cupcake) for supported harnesses and signals-based checkpointing for Copilot. The RPI artifact hardening provides defense-in-depth when cupcake is not installed. All constitutional articles pass.

**Estimated wall-clock:** 8-12 hours (sequential). ~4-6 hours with parallel phase execution.
**Estimated LOC:** ~400 lines (Rego policies, YAML configs, markdown hardening).

---

**⛔ Human Gate:** Plan approved? Proceed to Implement (Phase 1)?

*Approve with: APPROVE / REQUEST_CHANGES*
