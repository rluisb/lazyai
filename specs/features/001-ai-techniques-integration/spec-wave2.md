# Feature 001 — Wave 2 Acceptance Contract: Quality + Context

**Feature:** 001 — AI techniques integration into ai-setup installation  
**Wave:** Wave 2 — Quality + Context  
**Status:** Draft planning artifact; implementation requires human approval  
**Scope source:** `plan.md` Wave 2 roadmap: D4, N12, N2, D9, D13, D14, D5, D8  
**Date:** 2026-05-01

---

## 1. Scope Boundary

Wave 2 adds bounded quality/context improvements to the installed agent harness and, where explicitly approved, the existing sequential feature chain. It does **not** introduce retrieval, model routing, learning systems, parallel workflow execution, runtime conditionals, or a broad workflow-engine rewrite.

The original Wave 1 plan recommended observing W1 for at least two weeks before Wave 2. W1.A + W1.B are merged on `origin/main`, and the user explicitly approved moving to the next phase. This contract records the bake-period recommendation as an implementation risk: before executing Wave 2 code tasks, implementors should check W1.B gate behavior and active-run compatibility evidence.

### In Scope

| Item | Technique | Wave 2 shape |
|---|---|---|
| D4 | Chain-of-verification | `chain-verify` skill/report and optional sequential feature-chain step after review. |
| N12 | TillDone / completion enforcement | Completion checklist/report in implement/review workflow guidance. |
| N2 | Generated knowledge | Explicit knowledge-surfacing step in reasoning/planning guidance. |
| D9 | Structured feedback | Feedback schema and prompt guidance; bounded runtime propagation only behind human approval. |
| D13 | Causal reasoning | 5-Whys / causal-chain RCA method in bugfix templates and skills. |
| D14 | Environment-aware planning | Planner guidance to capture toolchain, CI, token/budget, and platform constraints before plan commitments. |
| D5 | Auto-recovery | Static safe-recovery policy and allowlist only; runtime autonomous recovery is deferred. |
| D8 | Agent state machine | Static lifecycle taxonomy and reporting vocabulary only; runtime per-agent state tracking is deferred. |

### Explicitly Out of Scope

- Wave 3 retrieval/intelligence: RAG preflight, N9 knowledge graph auto-injection, D1 retrieval pipeline, D2 model routing, N1 few-shot retrieval, N6 cross-repo orchestration.
- Wave 4 learning/measurement: continuous learning, structured evaluation benchmark, prompt optimization, agent progression scoring.
- Runtime guardrails, streaming, debate workflows, dynamic prompt optimization, provider billing/cost integrations.
- Runtime conditional steps, `{{#if ...}}` chain templates, parallel blocks, graph execution, or generic orchestration template rendering.
- Broad workflow engine rewrite or new scheduler.
- MCP server configuration changes unless a later approved task explicitly scopes them.
- Autonomous D5 recovery execution and runtime D8 per-agent lifecycle tracking without separate approval/ADR.

---

## 2. User Scenarios

### US1 — Verify artifact chains before declaring a feature done

As a human approver, I want agents to verify spec → plan → tasks → implementation → tests traceability before final documentation/completion so missing acceptance criteria are caught before work is called done.

**Acceptance:** A `ChainVerificationReport` identifies pass/warn/fail findings with repo-relative locations and does not replace human approval.

### US2 — Start plans with explicit context and causal reasoning

As a planner or bugfix agent, I want generated knowledge, environment constraints, and causal-analysis methods captured before deciding a plan or RCA so agents do not rely on unstated assumptions.

**Acceptance:** Planning/bugfix artifacts include structured knowledge, environment constraints, and causal-chain sections with clear unknowns and evidence.

### US3 — Make feedback, completion, and recovery actionable

As an orchestrator/human reviewer, I want completion claims, gate rejections, and recovery decisions to use structured formats so agents know what to fix and when to stop or ask for help.

**Acceptance:** Installed guidance defines TillDone, structured feedback, safe recovery, and lifecycle vocabulary without enabling unsafe autonomous runtime behavior.

---

## 3. Functional Requirements

### D4 — Chain-of-verification

- **FR-W2-001:** Provide a `chain-verify` skill that checks cross-artifact consistency across available artifacts: `spec.md`, `plan.md`, `tasks.md`, task files, implementation notes, tests, and verification evidence.
- **FR-W2-002:** `chain-verify` MUST emit exactly one `ChainVerificationReport` JSON object matching this contract's schema.
- **FR-W2-003:** Missing optional artifacts MUST produce `warn` findings, not parser-driven hard failures.
- **FR-W2-004:** Chain verification MAY be inserted as a sequential feature-chain step only after human approval of the chain-shape change.

### N12 — TillDone / completion enforcement

- **FR-W2-005:** Implementor and iterate guidance MUST require agents to continue until all task Done When criteria are met, or stop with a documented blocker.
- **FR-W2-006:** Reviewer guidance MUST include an early-stop check: every acceptance criterion has evidence, tests/quality gates are recorded, and out-of-scope work is absent.
- **FR-W2-007:** TillDone guidance MUST not require agents to exceed explicit session/task boundaries; one task per session remains valid.

### N2 + D14 — Generated knowledge and environment-aware planning

- **FR-W2-008:** Reasoning/planning guidance MUST require a concise Knowledge Surface for non-trivial tasks: facts, constraints, assumptions, unknowns, and evidence sources.
- **FR-W2-009:** Planning guidance MUST require an Environment Snapshot when plans depend on runtime/tooling: language/tool versions, package managers, CI latency, available tools, budget/token constraints, platform constraints, and network/secrets restrictions.
- **FR-W2-010:** Generated knowledge MUST be scoped and cited; it must not become speculative background filler.

### D9 — Structured feedback

- **FR-W2-011:** Define a `StructuredFeedback` schema for human gate rejections and review-request changes.
- **FR-W2-012:** Static guidance MUST instruct humans/agents to include required changes, suggestions, priority, evidence, and target phase/task.
- **FR-W2-013:** Runtime persistence/propagation of structured gate feedback is a separate approval gate. If approved, it MUST be a bounded extension to existing `advance_chain` output handling, not a new gate engine.

### D13 — Causal reasoning

- **FR-W2-014:** Bugfix/RCA templates and skills MUST include a causal method section using 5-Whys or a short fault-tree/causal chain.
- **FR-W2-015:** Causal reasoning MUST end at a process/standard/test gap where possible, not only a proximate code symptom.
- **FR-W2-016:** RCA guidance MUST record confidence and counterfactual checks for the selected cause.

### D5 — Auto-recovery

- **FR-W2-017:** Add a static safe-recovery policy with an allowlist of low-risk actions: re-run deterministic checks, retry transient provider/tool failures within existing retry limits, regenerate malformed report JSON from the same inputs, and create handoff when blocked.
- **FR-W2-018:** The policy MUST require human confirmation for code edits, dependency changes, destructive commands, migration changes, secrets/config changes, or ambiguous failures.
- **FR-W2-019:** No autonomous runtime recovery classifier or automatic edit loop is included in Wave 2 without separate approval.

### D8 — Agent state machine

- **FR-W2-020:** Add a static agent lifecycle vocabulary for reports/handoffs: `loading_context`, `planning`, `awaiting_approval`, `executing`, `verifying`, `blocked`, `handoff`, `done`, `error`.
- **FR-W2-021:** Guidance MUST explain how to use lifecycle labels in handoffs, status reports, and recovery summaries.
- **FR-W2-022:** Runtime per-agent state tracking in `ChainState`, `StepState`, or `get_status` is deferred pending approval/ADR.

---

## 4. Acceptance Criteria

### D4 — Chain-of-verification

- **AC-D4-001:** `chain-verify` exists with frontmatter/metadata valid under existing skill tests.
- **AC-D4-002:** `ChainVerificationReport` schema tests cover `schemaVersion`, `verdict`, `traceability`, `findings`, `checkedArtifacts`, and repo-relative locations.
- **AC-D4-003:** Fixture tests prove missing optional artifacts produce `warn`, not fail or crash.
- **AC-D4-004:** If chain integration is approved, chain-shape tests prove the feature chain remains sequential and contains no conditionals/templates/parallel blocks.

### N12 — TillDone

- **AC-N12-001:** Implement/iterate/review guidance contains an explicit completion enforcement checklist.
- **AC-N12-002:** Tests or snapshots prove the checklist requires evidence for every task Done When item before declaring done.
- **AC-N12-003:** Guidance preserves one-task/session boundaries and includes a blocker path.

### N2 + D14 — Generated knowledge and environment-aware planning

- **AC-N2-001:** Reasoning/planning guidance contains a bounded Knowledge Surface format.
- **AC-D14-001:** Planner guidance contains an Environment Snapshot format and requires environment assumptions to be marked verified/unverified.
- **AC-D14-002:** Tests or snapshots prove planning prompts mention no unapproved model routing, RAG, provider billing, or retrieval automation.

### D9 — Structured feedback

- **AC-D9-001:** `StructuredFeedback` schema is documented and tested with valid/invalid examples.
- **AC-D9-002:** Iterate/orchestrate guidance instructs agents to consume structured feedback when present and ask for clarification when required changes are missing.
- **AC-D9-003:** Runtime feedback propagation, if approved, persists rejected-gate feedback on the gate/step state and exposes it to the target step without accepting outcomes beyond existing approved/rejected semantics.

### D13 — Causal reasoning

- **AC-D13-001:** Bugfix/RCA template includes 5-Whys or causal-chain fields, evidence, confidence, and counterfactual check.
- **AC-D13-002:** Bugfix skill guidance requires causal analysis before fix planning for non-trivial bugs.

### D5 — Auto-recovery

- **AC-D5-001:** A static safe-recovery rule/policy exists and passes markdown/frontmatter or snapshot validation.
- **AC-D5-002:** Orchestrator guidance distinguishes auto-allowed low-risk retries from human-gated recovery.
- **AC-D5-003:** No Wave 2 implementation task changes runtime retry semantics unless a separate approval decision is recorded.

### D8 — Agent state machine

- **AC-D8-001:** Static lifecycle vocabulary appears in orchestrator/handoff guidance.
- **AC-D8-002:** Guidance states lifecycle labels are report vocabulary only in Wave 2 and do not imply runtime state-machine support.
- **AC-D8-003:** Runtime `ChainState`/`StepState` changes are absent unless a separate approval decision/ADR is accepted.

---

## 5. Data / Report Contracts

### ChainVerificationReport v1

```json
{
  "schemaVersion": "chain-verification-report/v1",
  "verdict": "pass|warn|fail",
  "checkedArtifacts": {
    "spec": "repo-relative/path or null",
    "plan": "repo-relative/path or null",
    "tasks": "repo-relative/path or null",
    "taskFiles": ["repo-relative/path"],
    "implementationEvidence": ["repo-relative/path or command label"],
    "tests": ["repo-relative/path or command label"]
  },
  "traceability": [
    {
      "requirementId": "FR-W2-001 or external AC id",
      "planRefs": ["repo-relative/path#section"],
      "taskRefs": ["repo-relative/path#section"],
      "evidenceRefs": ["repo-relative/path or command label"],
      "status": "covered|partial|missing|not-applicable"
    }
  ],
  "findings": [
    {
      "rule": "artifact-presence|requirement-trace|task-evidence|test-evidence|scope-boundary|rollback",
      "severity": "info|warn|fail",
      "message": "string",
      "recommendation": "string",
      "location": { "file": "repo-relative/path", "section": "string or null", "lineStart": 1, "lineEnd": 1 }
    }
  ]
}
```

### StructuredFeedback v1

```json
{
  "schemaVersion": "structured-feedback/v1",
  "verdict": "approved|request_changes|rejected|comment",
  "summary": "string",
  "requiredChanges": [
    { "id": "FB-001", "description": "string", "priority": "blocking|high|medium|low", "target": "step/task/file", "evidence": "string" }
  ],
  "suggestions": [
    { "description": "string", "priority": "medium|low", "target": "step/task/file" }
  ],
  "requestedBy": "human|reviewer|red-team|planner",
  "targetPhaseOrStep": "string or null"
}
```

### CompletionEnforcementReport v1

```json
{
  "schemaVersion": "completion-enforcement-report/v1",
  "status": "done|blocked|not-done",
  "criteria": [
    { "id": "string", "description": "string", "evidence": "string or null", "met": true }
  ],
  "blockers": ["string"],
  "outOfScopeChanges": ["repo-relative/path or note"]
}
```

---

## 6. Approval Decisions Required Before Implementation

1. **D4 chain integration:** approve adding a sequential `chain-verify` step to the feature chain, or ship `chain-verify` as a manual skill only.
2. **D9 runtime propagation:** approve bounded `advance_chain`/gate-state feedback persistence, or keep structured feedback as prompt/schema guidance only.
3. **D5 runtime auto-recovery:** deferred by default; separate approval/ADR required for autonomous recovery classifiers or automatic edit loops.
4. **D8 runtime state machine:** deferred by default; separate approval/ADR required for `ChainState`/`StepState` schema changes and `get_status` output changes.
5. **W1 bake evidence:** before code implementation, decide whether W1.B has enough observed stability or whether Wave 2 runtime-touching tasks should wait.
