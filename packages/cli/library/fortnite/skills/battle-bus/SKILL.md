---
name: battle-bus
description: Human-gated workflow orchestrator and parallel execution engine. Generates phase-by-phase blueprints mapped to specific agents, then executes independent waves concurrently. The Battle Bus deploys everyone at the right place and time — you decide when to jump.
trigger: /battle-bus
skill_path: skills/battle-bus
scripts:
  - name: agent-dispatch.sh
    description: Record dispatches and export STATE.md
    path: scripts/agent-dispatch.sh
  - name: agent-msg.sh
    description: Inter-agent message bus (send, recv, broadcast)
    path: scripts/agent-msg.sh
  - name: squad-map.sh
    description: Full visual dispatch map of agent assignments
    path: scripts/squad-map.sh
  - name: task-barrier.sh
    description: Parallel sync barriers (create, arrive, wait)
    path: scripts/task-barrier.sh
  - name: task-lock.sh
    description: Exclusive resource locks (acquire, release, try)
    path: scripts/task-lock.sh
  - name: dispatch-wave.sh
    description: Auto-dispatch parallel tasks from wave definition with collision check
    path: ../../scripts/dispatch-wave.sh
  - name: wait-barrier.sh
    description: Non-blocking barrier wait with timeout and status polling
    path: ../../scripts/wait-barrier.sh
  - name: wave-summary.sh
    description: Aggregate results from all parallel tasks in a wave
    path: ../../scripts/wave-summary.sh
  - name: check-file-collision.sh
    description: Pre-flight check for parallel write conflicts
    path: ../../scripts/check-file-collision.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | workflow blueprint generation, parallel execution planning, team dispatch |
| **Do not use when** | single-agent task, implementation, review |
| **Primary agent** | loop-driver |
| **Runtime risk** | Medium — parallel execution coordination |
| **Outputs** | Phase-by-phase blueprint, agent assignments, wave definitions |
| **Validation** | Barrier sync, collision checks |
| **Deep mode trigger** | `/battle-bus` or complex multi-phase workflow |

# Battle Bus

## Purpose
Deploy your squad at the right place and time. This skill does two things:

1. **Generates a phase-by-phase execution blueprint** — each phase maps to a specific agent with feedforward context from the previous phase.
2. **Enables safe parallel execution** — identifies independent work units that can run concurrently via the Wave Model without file conflicts.

**This skill outputs a plan. It does not dispatch agents.**
The human decides which phases to run, in what order, and whether to skip or modify any.

---

## Scripts

This skill owns the following scripts:

| Script | Purpose |
|--------|---------|
| `agent-dispatch.sh` | Record dispatches and export STATE.md |
| `agent-msg.sh` | Inter-agent message bus (send, recv, broadcast) |
| `squad-map.sh` | Full visual dispatch map of agent assignments |
| `task-barrier.sh` | Parallel sync barriers (create, arrive, wait) |
| `task-lock.sh` | Exclusive resource locks (acquire, release, try) |
| `dispatch-wave.sh` | Auto-dispatch parallel tasks from wave definition |
| `wait-barrier.sh` | Non-blocking barrier wait with timeout |
| `wave-summary.sh` | Aggregate results from all parallel tasks |
| `check-file-collision.sh` | Pre-flight check for parallel write conflicts |

Run from skill directory: `./scripts/<script-name>.sh <command>`

---

# Part 1: Workflow Blueprint

## How to use
1. Load this skill
2. Describe your task in one sentence
3. Battle Bus selects the best template (or you specify one)
4. Battle Bus fills in a blueprint with your context
5. You review, adjust, and approve
6. Dispatch each phase manually via `session` tool or by instructing the orchestrator

### Blueprint Output Format

Battle-bus produces a structured blueprint containing:

| Field | Type | Description |
|-------|------|-------------|
| phases | array | Ordered list of phases with agent assignments |
| waves | array | Independent task groups for parallel execution |
| gates | array | Human approval checkpoints |
| fallback | object | Failure recovery strategy |

**Example:**
```yaml
phases:
  - name: "Clarify"
    agent: turbo-crank
    mode: clarify
waves:
  - id: wave1
    tasks: [task-a, task-b]
    barrier: wave1-sync
gates:
  - phase: 2
    condition: "Human approves spec before implementation"
fallback:
  strategy: "escalate"
  target: "loop-driver"
```

---

## Templates

### rpi — Research → Plan → Implement → Verify
*Use for: new features, significant changes, anything with unclear scope.*

```yaml
template: rpi
phases:
  - phase: clarify
    agent: turbo-crank
    skill: storm-scout (Phase 0: Clarify)
    goal: Resolve all unknowns before any work begins
    feedforward: "Task: {GOAL}. Identify and resolve critical unknowns. MAX_QUESTIONS=3. Output: Confirmed Understanding block."
    gate: human confirms understanding

  - phase: research
    agent: loot-hawk
    skill: storm-scout (Phase 1: Research)
    goal: Map existing codebase and domain knowledge relevant to the task
    feedforward: "Confirmed Understanding: {CLARIFY_OUTPUT}. Research the codebase and domain. Output: structured findings document."
    gate: findings document complete

  - phase: plan
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Produce spec and ordered task list
    feedforward: "Research findings: {RESEARCH_OUTPUT}. Produce spec.md and tasks.md. No code yet."
    gate: human approves spec + task list

  - phase: implement
    agent: wall-builder
    mode: standard   # swap → senior for complex/multi-module, junior for tiny fixes
    skill: build-mode
    goal: Write code against the approved spec, task by task
    feedforward: "Spec: {SPEC}. Current task: {TASK}. Done-condition: {DONE_CONDITION}. Follow existing patterns."
    gate: task done-condition met

  - phase: verify
    agent: shield-audit
    mode: review
    skill: zero-point
    goal: Independent check — spec compliance + quality gates
    feedforward: "Spec: {SPEC}. Implementation complete. Run quality gates and verify spec compliance. Output: Verification Report."
    gate: PASS verdict
```

---

### bug-investigation — Diagnose → Fix → Verify
*Use for: unknown bugs, production issues, failing tests with unclear cause.*

```yaml
template: bug-investigation
phases:
  - phase: diagnose
    agent: loot-hawk
    skill: reboot-van (diagnose mode)
    goal: Identify root cause through hypothesis-driven investigation
    feedforward: "Bug: {BUG_DESCRIPTION}. Symptoms: {SYMPTOMS}. Build feedback loop, reproduce, hypothesise, instrument. Output: confirmed root cause + hypothesis evidence."
    gate: root cause confirmed by evidence

  - phase: fix
    agent: wall-builder
    mode: standard   # swap → senior for systemic/complex bugs
    skill: build-mode
    goal: Apply minimal targeted fix for the confirmed root cause
    feedforward: "Confirmed root cause: {ROOT_CAUSE}. Apply minimal fix. Do not refactor unrelated code. Output: changed files + regression test."
    gate: fix in place + regression test written

  - phase: verify
    agent: shield-audit
    mode: review
    skill: zero-point
    goal: Confirm fix resolves the bug and quality gates pass
    feedforward: "Bug: {BUG_DESCRIPTION}. Fix: {FIX_DESCRIPTION}. Run quality gates. Verify fix via regression test. Output: Verification Report."
    gate: PASS verdict
```

---

### tdd — Clarify → Plan → Implement (TDD) → Verify
*Use for: well-scoped new behaviour implemented test-first.*

```yaml
template: tdd
phases:
  - phase: clarify
    agent: turbo-crank
    skill: storm-scout (Phase 0: Clarify)
    goal: Confirm exact behaviour to be tested
    feedforward: "Task: {GOAL}. Clarify exact expected behaviour for TDD. Output: Confirmed Understanding with testable acceptance criteria."
    gate: human confirms acceptance criteria

  - phase: plan
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Produce TDD spec and ordered test list
    feedforward: "Acceptance criteria: {CLARIFY_OUTPUT}. Produce spec with testable requirements. Break into ordered test cases — each test case is one unit of behaviour."
    gate: human approves test plan

  - phase: implement
    agent: wall-builder
    mode: tdd
    skill: build-mode (TDD mode)
    goal: Implement behaviour test-first using RED→GREEN→REFACTOR cycle
    feedforward: "Spec: {SPEC}. Test plan: {TEST_PLAN}. One test at a time. Output: passing tests + clean implementation."
    gate: all tests pass, quality gates pass

  - phase: verify
    agent: shield-audit
    mode: review
    skill: zero-point
    goal: Independent spec compliance check
    feedforward: "Spec: {SPEC}. TDD implementation complete. Verify all acceptance criteria met. Run full quality gates. Output: Verification Report."
    gate: PASS verdict
```

---

### code-review — Research → Review → Report
*Use for: reviewing PRs, auditing a module, assessing quality of a change.*

```yaml
template: code-review
phases:
  - phase: research
    agent: loot-hawk
    skill: storm-scout (Phase 1: Research)
    goal: Understand the change and its context before reviewing
    feedforward: "Change: {CHANGE_DESCRIPTION}. Files/PR: {FILES_OR_PR}. Map what changed, why, and what it touches. Output: context summary."
    gate: context summary complete

  - phase: review
    agent: shield-audit
    mode: review
    skill: zero-point
    goal: Structured review — correctness, spec compliance, quality, scope. Produces final report.
    feedforward: "Context: {RESEARCH_OUTPUT}. Spec/requirements: {SPEC_IF_ANY}. Review for correctness, missing tests, scope creep, patterns. Output: structured findings + pass/fail report."
    gate: human receives report
```

---

### system-design — Clarify → Research → Architecture → Spec
*Use for: new systems, major architectural decisions, multi-team design work.*

```yaml
template: system-design
phases:
  - phase: clarify
    agent: turbo-crank
    skill: storm-scout (Phase 0: Clarify)
    goal: Understand problem space, constraints, and non-negotiables before designing
    feedforward: "Design problem: {GOAL}. MAX_QUESTIONS=5. Focus on: constraints, scale, existing systems to integrate with, non-goals. Output: Confirmed Understanding."
    gate: human confirms problem space

  - phase: research
    agent: loot-hawk
    skill: storm-scout (Phase 1: Research)
    goal: Map existing architecture, patterns, and relevant external knowledge
    feedforward: "Problem: {CLARIFY_OUTPUT}. Research existing codebase architecture, relevant patterns, and external precedents. Output: findings document."
    gate: findings document complete

  - phase: architecture
    agent: turbo-crank
    goal: Produce architecture decision and constitution
    feedforward: "Problem: {CLARIFY_OUTPUT}. Research: {RESEARCH_OUTPUT}. Produce: 1) Architecture decision with options considered and rationale, 2) Constitution (what/why/constraints/non-goals)."
    gate: human approves architecture + constitution

  - phase: spec
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Convert architecture into an executable spec and task breakdown
    feedforward: "Architecture: {ARCHITECTURE_OUTPUT}. Constitution: {CONSTITUTION}. Produce spec.md and tasks.md ready for implementation."
    gate: human approves spec
```

---

### incident-response — Diagnose → Fix → Verify → Post-mortem
*Use for: production incidents, P1/P2 issues, system failures.*

```yaml
template: incident-response
phases:
  - phase: diagnose
    agent: respawn-crew
    skill: reboot-van (diagnose mode)
    goal: Fast root-cause identification under time pressure
    feedforward: "Incident: {INCIDENT_DESCRIPTION}. Symptoms: {SYMPTOMS}. Affected systems: {AFFECTED_SYSTEMS}. Time-boxed investigation — prioritise recovery speed. Output: confirmed or best-hypothesis root cause."
    gate: root cause identified (confirmed or high-confidence hypothesis)

  - phase: fix
    agent: wall-builder
    mode: senior
    skill: build-mode
    goal: Apply targeted fix — minimise blast radius, maximise recovery speed
    feedforward: "Root cause: {ROOT_CAUSE}. Affected system: {SYSTEM}. Apply minimal targeted fix. Document every change. Output: fix applied + verification steps."
    gate: human approves fix before deploy

  - phase: verify
    agent: shield-audit
    mode: review
    skill: zero-point
    goal: Confirm system restored and quality gates pass
    feedforward: "Incident: {INCIDENT_DESCRIPTION}. Fix applied: {FIX}. Verify: 1) Original symptom resolved, 2) No regressions, 3) Quality gates pass. Output: Verification Report."
    gate: PASS verdict

  - phase: post-mortem
    agent: respawn-crew
    goal: Write post-mortem document
    feedforward: "Incident timeline: {TIMELINE}. Root cause: {ROOT_CAUSE}. Fix: {FIX}. Write post-mortem: timeline, root cause, impact, fix, action items to prevent recurrence."
    gate: post-mortem published
```

---

### spec-design-judge — Clarify → Fork Architectures → Judge → Spec
*Use for: new subsystems with multiple valid architectural approaches.*

```yaml
template: spec-design-judge
phases:
  - phase: clarify
    agent: turbo-crank
    skill: storm-scout (Phase 0: Clarify)
    goal: Understand problem space and constraints before designing
    feedforward: "Design problem: {GOAL}. MAX_QUESTIONS=5. Focus on: constraints, scale, non-goals. Output: Confirmed Understanding."
    gate: human confirms understanding

  - phase: fork-architecture-a
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Produce architecture option A with rationale
    feedforward: "Design problem: {CLARIFY_OUTPUT}. Produce architecture A: options considered, trade-offs, rationale. Output: architecture document A."
    gate: architecture A complete

  - phase: fork-architecture-b
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Produce architecture option B with rationale
    feedforward: "Design problem: {CLARIFY_OUTPUT}. Produce architecture B: different approach, trade-offs, rationale. Output: architecture document B."
    gate: architecture B complete

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate architecture A vs B against weighted rubric
    feedforward: "INPUTS: {ARCH_A_PATH},{ARCH_B_PATH}. CONTEXT: {CLARIFY_OUTPUT}. Judge which architecture better fits the problem. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: spec-from-winner
    agent: turbo-crank
    skill: storm-scout (Phase 2: Plan)
    goal: Convert winning architecture into executable spec and task breakdown
    feedforward: "Winning architecture: {WINNER}. Verdict: {VERDICT}. Produce spec.md and tasks.md ready for implementation."
    gate: human approves spec
```

---

### implementation-judge — Fork Implementations → Judge → Apply Winner
*Use for: non-trivial algorithmic choices or structural pattern decisions.*

```yaml
template: implementation-judge
phases:
  - phase: fork-impl-a
    agent: wall-builder
    mode: senior
    skill: build-mode
    goal: Implement approach A for the chosen algorithm/pattern
    feedforward: "Spec: {SPEC}. Implement approach A. Output: changed files + test results."
    gate: implementation A complete

  - phase: fork-impl-b
    agent: wall-builder
    mode: senior
    skill: build-mode
    goal: Implement approach B for the chosen algorithm/pattern
    feedforward: "Spec: {SPEC}. Implement approach B. Output: changed files + test results."
    gate: implementation B complete

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate implementation A vs B against weighted rubric
    feedforward: "INPUTS: {IMPL_A_PATH},{IMPL_B_PATH}. CONTEXT: {SPEC}. Judge which is cleaner, faster, more maintainable. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: apply-winner
    agent: wall-builder
    mode: senior
    skill: build-mode
    goal: Apply the winning implementation to the main branch
    feedforward: "Winner: {WINNER}. Verdict: {VERDICT}. Apply winner's implementation. Discard loser."
    gate: winner applied + quality gates pass
```

---

### incident-mitigation-judge — Diagnose → Fork Mitigations → Judge → Apply
*Use for: P2/P3 incidents with multiple remediation paths and unclear trade-offs.*

```yaml
template: incident-mitigation-judge
phases:
  - phase: diagnose
    agent: respawn-crew
    skill: reboot-van (diagnose mode)
    goal: Confirm root cause and identify multiple remediation paths
    feedforward: "Incident: {INCIDENT}. Symptoms: {SYMPTOMS}. Confirm root cause. Identify at least two viable remediation paths with trade-offs. Output: confirmed root cause + remediation options."
    gate: root cause confirmed + options identified

  - phase: fork-mitigation-a
    agent: respawn-crew
    mode: mitigate
    skill: reboot-van (iterate mode)
    goal: Detail mitigation strategy A
    feedforward: "Root cause: {ROOT_CAUSE}. Strategy A: detailed steps, blast radius, rollback plan. Output: mitigation plan A."
    gate: plan A complete

  - phase: fork-mitigation-b
    agent: respawn-crew
    mode: mitigate
    skill: reboot-van (iterate mode)
    goal: Detail mitigation strategy B
    feedforward: "Root cause: {ROOT_CAUSE}. Strategy B: detailed steps, blast radius, rollback plan. Output: mitigation plan B."
    gate: plan B complete

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate mitigation A vs B on safety and speed
    feedforward: "INPUTS: {MITIGATION_A_PATH},{MITIGATION_B_PATH}. Judge which is safer and faster. Add 'Blast radius safety' criterion. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: apply-winner
    agent: respawn-crew
    mode: mitigate
    goal: Apply the winning mitigation strategy
    feedforward: "Winner: {WINNER}. Verdict: {VERDICT}. Apply mitigation. Verify symptom resolution."
    gate: human approves + mitigation applied
```

---

### deploy-strategy-judge — Fork Deploy Strategies → Judge → Execute
*Use for: production deploys with ambiguous rollout strategy or competing patterns.*

```yaml
template: deploy-strategy-judge
phases:
  - phase: fork-strategy-a
    agent: rift-deploy
    mode: production
    goal: Design deployment strategy A (e.g., blue/green, canary, rolling)
    feedforward: "Deploy target: {TARGET}. Strategy A: rollout pattern, rollback speed, blast radius. Output: deploy strategy document A."
    gate: strategy A complete

  - phase: fork-strategy-b
    agent: rift-deploy
    mode: production
    goal: Design deployment strategy B (alternative rollout pattern)
    feedforward: "Deploy target: {TARGET}. Strategy B: alternative rollout pattern, rollback speed, blast radius. Output: deploy strategy document B."
    gate: strategy B complete

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate deploy strategy A vs B on operational risk
    feedforward: "INPUTS: {STRATEGY_A_PATH},{STRATEGY_B_PATH}. Add 'Rollback speed' and 'Blast radius' criteria. Judge which minimizes risk. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: execute-winner
    agent: rift-deploy
    mode: production
    goal: Execute the winning deployment strategy with human approval
    feedforward: "Winner: {WINNER}. Verdict: {VERDICT}. Execute deploy. Human must approve each step."
    gate: deploy complete + post-deploy health check passed
```

---

### model-fallback-judge — Capture Outputs → Judge → Resume Routing
*Use for: resolving disagreements between model fallback outputs on critical routing decisions.*

```yaml
template: model-fallback-judge
phases:
  - phase: capture-outputs
    agent: loop-driver
    goal: Capture the two conflicting model outputs as files
    feedforward: "Primary model output: {OUTPUT_A}. Fallback model output: {OUTPUT_B}. Save both to temporary files. Record disagreement point."
    gate: outputs captured

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate which model output better aligns with spec intent
    feedforward: "INPUTS: {OUTPUT_A_PATH},{OUTPUT_B_PATH}. CONTEXT: {SPEC_OR_TASK}. Add 'Spec/plan alignment' as CRITICAL weight. Judge which routing decision is correct. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: resume-routing
    agent: loop-driver
    goal: Resume routing using the winning model output, with human confirmation
    feedforward: "Winner: {WINNER}. Verdict: {VERDICT}. Resume dispatch with winning output. Human confirms before executing high-stakes dispatches."
    gate: dispatch resumed
```

---

### research-synthesis-judge — Fork Research → Judge → Synthesize Findings
*Use for: MODE=exhaustive research in unfamiliar domains with competing directions.*

```yaml
template: research-synthesis-judge
phases:
  - phase: fork-research-a
    agent: loot-hawk
    mode: exhaustive
    skill: storm-scout (Phase 1: Research)
    goal: Explore research direction A (e.g., breadth-first, specific tool focus)
    feedforward: "Domain: {DOMAIN}. Direction A: approach, tools, expected findings. Output: structured findings document A."
    gate: findings A complete

  - phase: fork-research-b
    agent: loot-hawk
    mode: exhaustive
    skill: storm-scout (Phase 1: Research)
    goal: Explore research direction B (alternative approach)
    feedforward: "Domain: {DOMAIN}. Direction B: alternative approach, different tools, expected findings. Output: structured findings document B."
    gate: findings B complete

  - phase: judge
    agent: shield-audit
    mode: judge
    skill: zero-point
    goal: Evaluate which research synthesis is more actionable and accurate
    feedforward: "INPUTS: {FINDINGS_A_PATH},{FINDINGS_B_PATH}. Add 'Actionability' as CRITICAL criterion. Judge which findings are more useful for downstream planning. Output: structured verdict JSON."
    gate: verdict delivered

  - phase: synthesize-winner
    agent: loot-hawk
    mode: deep
    goal: Finalize findings using the winning research direction
    feedforward: "Winner: {WINNER}. Verdict: {VERDICT}. Synthesize final findings document for downstream agents."
    gate: final findings delivered to planner
```

---

## How to fill a template

1. **State your task** — one sentence: what are you trying to do?
2. **Battle Bus selects a template** — or you say `template: rpi` etc.
3. **Battle Bus produces a filled blueprint** — replacing `{GOAL}`, `{REPO}`, `{CONSTRAINTS}` with your actual context
4. **You review the blueprint** — adjust agents, skip phases, change parameters
5. **You dispatch phases** — one by one, using `session` tool or orchestrator instruction

## Spec Format — Compact Blueprint

Specs in battle-bus templates should use **compact-blueprint format** (§G/§C/§I/§V/§T/§B) for token efficiency (~75% fewer tokens than prose):

```markdown
## §G Goal
feature: auth mw refresh path. scope: token expiry handling.

## §C Constraints
| id | rule | priority |
|----|------|----------|
| C1 | no cloud calls | critical |

## §I Interfaces
| id | signature | returns |
|----|-----------|---------|
| I1 | refresh(token) | {valid, expiry} |

## §V Invariants
| id | rule | evidence |
|----|------|----------|
| V1 | auth mw throws 401 @ expired | test: auth_test.go:42 |

## §T Tasks
| id | desc | done | files |
|----|------|------|-------|
| T1 | add refresh path | ☐ | auth.go |

## §B Bugs
| id | symptom | fix | status |
|----|---------|-----|--------|
| B1 | null user @ profile | add guard L42 | fixed |
```

Use `./skills/compact-blueprint/scripts/spec-encode.sh --input <prose> --output <compact>` to convert existing prose specs. Ricochet backprop updates §V/§B sections automatically on test failures.

## Agent swap guide

| Situation | Default | Swap to |
|-----------|---------|---------|
| Complex/multi-module implementation | `wall-builder` (standard) | `wall-builder` (senior mode) |
| Tiny 1-file fix | `wall-builder` (standard) | `wall-builder` (junior mode) |
| Test-first implementation | `wall-builder` (standard) | `wall-builder` (TDD mode) |
| Parallel implementation waves | — | See Part 2: Wave Model below |
| Security-sensitive change | `shield-audit` (review mode) | `shield-audit` (security mode) |
| Adversarial testing needed | `shield-audit` (review mode) | `shield-audit` (adversarial mode) |
| Fast quality check only | `shield-audit` (review mode) | `shield-audit` (quick mode) |
| Judge parallel fork outputs | `shield-audit` (review mode) | `shield-audit` (judge mode) |
| Performance/SLO concern | `shield-audit` (review) | `respawn-crew` |
| Infrastructure deployment | — | `rift-deploy` |

## Integration with Other Skills

- **storm-scout**: Feeds research findings into plan phase
- **compact-blueprint**: Spec format for all templates (§G/§C/§I/§V/§T/§B)
- **ricochet**: Updates spec §V/§B on test failures during implement phase
- **drift-scope**: Validates implementation against spec during verify phase
- **zero-point**: Runs as verify phase — spec compliance + quality gates
- **build-mode**: Runs as implement phase — writes code against spec
- **truth-chain**: All dispatches, barriers, parallel tasks, and decisions recorded to immutable ledger
- **slurp-juice**: Checkpoints at each gate transition

## Workflow rules
- This skill outputs a plan only — never dispatches automatically
- Every phase has an explicit gate — human decides when to advance
- Feedforward context must be filled in before dispatching a phase
- If a phase produces nothing useful, diagnose before advancing
- Phases can be skipped if already complete (note it in your blueprint)

---

# Part 2: Parallel Execution (Wave Model)

Enable safe parallel task execution within implementation phases by identifying independent work units that can run concurrently without conflicts.

## The Wave Model

Group tasks into waves based on dependencies. Tasks within a wave run in parallel; waves run sequentially.

```
Wave 0 (No dependencies):
  ┌─────────────┐  ┌─────────────┐
  │ Task A      │  │ Task B      │  ← Run in parallel
  └─────────────┘  └─────────────┘
         │                │
         └────────┬───────┘
                  ▼
Wave 1 (Depends on Wave 0):
  ┌─────────────────────────┐
  │ Task C (needs A + B)    │  ← Must wait for Wave 0
  └─────────────────────────┘
                  │
                  ▼
Wave 2 (Depends on Wave 1):
  ┌─────────────┐  ┌─────────────┐
  │ Task D      │  │ Task E      │  ← Parallel again
  └─────────────┘  └─────────────┘
```

## Coordination Scripts

Use the scripts in this skill for parallel execution:

```bash
# Create a barrier for wave synchronization
./scripts/task-barrier.sh create "wave-0-sync" 2

# Each parallel task arrives when done
./scripts/task-barrier.sh arrive "wave-0-sync"

# Coordinator waits for all
./scripts/task-barrier.sh wait "wave-0-sync" 120

# Acquire exclusive lock before writing shared resource
./scripts/task-lock.sh acquire "spec-write" "turbo-crank"

# Release when done
./scripts/task-lock.sh release "spec-write"

# Send message between agents
./scripts/agent-msg.sh send <session-id> "wall-builder" "shield-audit" "Review ready" "Implementation complete"
```

## Dependency Analysis

### File Touch-Map

Before parallelizing, create a file touch-map — which files each task modifies:

```
Task A: src/api/handlers.ts, src/api/routes.ts
Task B: src/db/queries.ts, src/db/schema.sql
Task C: src/api/handlers.ts, src/db/queries.ts   ← conflicts with A and B
Task D: src/services/user.ts
Task E: src/services/order.ts
```

### Conflict Detection

Two tasks **conflict** if they modify the same file:
- Task A and Task C both touch `src/api/handlers.ts` → **CONFLICT** — different waves
- Task A and Task B have no overlap → **SAFE** to parallelize in same wave

### Wave Assignment Rules

1. Tasks with no file conflicts can be in the same wave
2. If Task X depends on Task Y's output, X must be in a later wave
3. Minimize total waves while respecting all dependencies

## Execution

```
Wave 0: dispatch agent(A) + agent(B) concurrently → wait for both
Wave 1: dispatch agent(C) → wait for completion
Wave 2: dispatch agent(D) + agent(E) concurrently → wait for both
```

## Safety Guarantees

| Guarantee | How Enforced |
|-----------|--------------|
| No file conflicts | File touch-map analysis before dispatch |
| No race conditions | Each agent works in isolated worktree or branch |
| Deterministic merge | Waves complete fully before next wave starts |
| Rollback safety | Each wave's commits are atomic |
| Approval gates | User approves wave plan before execution |

## Worktree Isolation

Each parallel task gets its own worktree or branch:

```
bee-gone/worktrees/
├── <task-slug>-a/   ← Fork A workspace
├── <task-slug>-b/   ← Fork B workspace
└── ...
```

**Fork outputs**: Completed fork results are written to `<worktree>/result.md` (or other agreed artifact path) before calling `task-barrier.sh arrive`. The judge phase references these output files via the `INPUTS` parameter.

**CRITICAL**: Agents never push from worktrees. Merging and remote operations require explicit user approval.

## Merge Strategy

1. Complete Wave 0 → user merges all wave worktrees to main
2. Rebase Wave 1 worktrees on updated main
3. Complete Wave 1 → user merges
4. Repeat until all waves complete

## Example

```
Feature: Add user reviews to products

Wave 0 (Independent):
├── Task A: Add review data model + migrations
│           Files: src/models/review.ts, migrations/
└── Task B: Add product rating calculation
            Files: src/services/rating.ts

Wave 1 (Depends on A + B):
└── Task C: Add review API endpoints
            Files: src/api/reviews.ts, src/api/routes.ts

Wave 2 (Depends on C):
├── Task D: Add review UI components
│           Files: src/ui/review_form.tsx
└── Task E: Add review notification emails
            Files: src/services/notification.ts
```

## Anti-Patterns

- Parallelizing tasks that share mutable files
- Starting Wave 1 before Wave 0 fully completes
- Forgetting to rebase on updated main between waves
- Multiple agents working in the same worktree
- Agents pushing or merging without user approval

## Output Format

```markdown
## Parallel Execution Report
- **Waves completed:** N/N
- **Files touched:** [list]
- **Conflicts detected:** [none | list]
- **Merge status:** [clean | required manual resolution]
```

## Parallel Execution Integration
- **Planner agent**: Creates file touch-map and wave assignments during planning
- **Builder agent**: Executes tasks within assigned worktrees
- **Reviewer agent**: Reviews each wave's output before merge approval

---

## Judge Wave — LLM-as-Judge for Parallel Forks

After a parallel fork wave completes (barrier resolved), use a Judge Wave to select the best output before proceeding.

### Template

```
## Wave N+1: Judge
AGENT: shield-audit
MODE: judge
THINK: xhigh
INPUTS: <path-to-fork-a-output>,<path-to-fork-b-output>
CONTEXT: <path-to-spec-or-plan>

## Task
Evaluate the two fork outputs against the weighted rubric. Return a structured verdict.
```

### Full Fork → Judge → Proceed Pattern

```
Wave 1 (parallel forks):
  fork A: wall-builder → output: bee-gone/worktrees/<task-slug>-a/result.md
  fork B: wall-builder → output: bee-gone/worktrees/<task-slug>-b/result.md
  barrier: fork-sync (count: 2)

Wave 2 (judge):
  shield-audit MODE=judge
  INPUTS: bee-gone/worktrees/<task-slug>-a/result.md,bee-gone/worktrees/<task-slug>-b/result.md
  CONTEXT: bee-gone/specs/NNN-slug/SPEC.md

Wave 3 (human gate):
  loop-driver presents verdict to user
  user selects winner (or requests re-fork if "neither")

Wave 4 (proceed):
  wall-builder applies winner's implementation
```

### When to Use a Judge Wave
- Two competing implementations of the same feature
- Uncertain which approach better fits the codebase
- After adversarial forks (one conservative, one aggressive)
- When spec allows multiple valid solutions

### When NOT to Use a Judge Wave
- Single implementation (no fork) — use shield-audit MODE=review instead
- Forks targeting different files — no meaningful comparison
- Time-critical hotfixes — judge adds latency, use MODE=review
