> **HISTORICAL (2026-06-20) — parts of this plan target the removed `packages/ai-setup-ts` package; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Task 024: Wave 2 Integration Validation + Scope Audit

**Phase:** W2.D — Integration validation  
**User Story:** All Wave 2 stories  
**Status:** DONE — Wave 2 scope audit and validation evidence recorded  
**Depends on:** T015-T023 approved/completed subset  
**Parallel with:** none

---

## Objective

Validate that completed Wave 2 tasks satisfy `spec-wave2.md`, remain within approved scope, and introduce no Wave 3/4 or unsupported runtime/chain constructs.

## Spec References

- All `spec-wave2.md` acceptance criteria.
- `tasks-wave2.md` Acceptance Trace.
- `plan-wave2.md` Runtime/Deferred Capability Ledger.

## Files to Change/Create

- Existing integration/snapshot tests that validate installed library content.
- Existing chain-shape tests if T018 was approved.
- Existing orchestrator runtime tests if T021 was approved.
- A Wave 2 acceptance trace note only if the repo convention uses evidence notes.

## Files NOT to Touch

- Any unapproved deferred runtime files for D5/D8.
- Wave 3/4 implementation files.
- MCP server configuration.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing acceptance-trace checks mapping each Wave 2 AC to at least one test/snapshot/evidence item.
2. Add failing scope-audit assertions that no RAG/model-routing/eval/learning/debate/MCP artifacts were introduced.
3. Add failing chain-shape assertions that no chain contains runtime conditionals, template markers, or parallel blocks.
4. If T021 was approved, add regression tests that gate feedback remains backward compatible.
5. Run package-specific test suites for changed surfaces and record commands/evidence.

## Done When

- [x] Every Wave 2 AC has test/snapshot/evidence trace.
- [x] Approved runtime/chain changes are explicitly identified; unapproved D5/D8 runtime changes are absent.
- [x] No unsupported chain constructs or Wave 3/4 scope is present.
- [x] Test evidence is recorded for all changed packages.
- [x] Human decisions from T014 are reflected in the final scope audit.

## Acceptance Trace Evidence

| Spec AC | Evidence |
|---|---|
| AC-D4-001 | `packages/ai-setup-ts/src/__tests__/chain-verification-report.test.ts` validates `chain-verify` skill frontmatter/metadata and read-only scope. |
| AC-D4-002 | `packages/ai-setup-ts/src/__tests__/chain-verification-report.test.ts` validates `ChainVerificationReport` schema fields: `schemaVersion`, `verdict`, `checkedArtifacts`, `traceability`, `findings`, and repo-relative locations. |
| AC-D4-003 | `packages/ai-setup-ts/src/__tests__/chain-verification-report.test.ts` fixture `warn.json` validates missing optional artifacts produce warn-only output; `ambiguous-warn.json` validates malformed/ambiguous parsing warns instead of failing. |
| AC-D4-004 | `packages/ai-setup-ts/src/__tests__/orchestration.test.ts` and `packages/ai-setup-go/internal/scaffold/orchestration_test.go` validate approved sequential `chain-verify` placement and reject conditionals/templates/parallel blocks. |
| AC-N12-001 | `packages/ai-setup-ts/src/__tests__/completion-enforcement.test.ts` validates implement/iterate/review completion checklist guidance. |
| AC-N12-002 | `packages/ai-setup-ts/src/__tests__/completion-enforcement.test.ts` validates `CompletionEnforcementReport` rejects done claims without evidence for every criterion. |
| AC-N12-003 | `packages/ai-setup-ts/src/__tests__/completion-enforcement.test.ts` validates one-task/session boundaries and blocker handling. |
| AC-N2-001 | `packages/ai-setup-ts/src/__tests__/generated-knowledge-env-planning.test.ts` validates bounded Knowledge Surface guidance. |
| AC-D14-001 | `packages/ai-setup-ts/src/__tests__/generated-knowledge-env-planning.test.ts` validates Environment Snapshot guidance with verified/unverified assumptions. |
| AC-D14-002 | `packages/ai-setup-ts/src/__tests__/generated-knowledge-env-planning.test.ts` validates no model routing, RAG, provider billing, or retrieval automation promises. |
| AC-D9-001 | `packages/ai-setup-ts/src/__tests__/structured-feedback.test.ts` validates `StructuredFeedback` schema fixtures and invalid examples. |
| AC-D9-002 | `packages/ai-setup-ts/src/__tests__/structured-feedback.test.ts` validates iterate/orchestrate consumption guidance and clarification path. |
| AC-D9-003 | `packages/orchestrator/src/__tests__/chain-machine.case.ts` validates approved bounded gate-feedback propagation for rejected gates and backward-compatible approved/rejected routing. |
| AC-D13-001 | `packages/ai-setup-go/internal/library/bugfix_rca_content_test.go` validates RCA template 5-Whys/causal-chain, evidence, confidence, and counterfactual fields. |
| AC-D13-002 | `packages/ai-setup-go/internal/library/bugfix_rca_content_test.go` validates bugfix skill causal analysis before non-trivial fix planning. |
| AC-D5-001 | `packages/ai-setup-go/internal/library/auto_recovery_content_test.go` validates static safe-recovery policy content. |
| AC-D5-002 | `packages/ai-setup-go/internal/library/auto_recovery_content_test.go` validates orchestrator guidance distinguishes auto-allowed low-risk retries from human-gated recovery. |
| AC-D5-003 | `packages/ai-setup-go/internal/library/auto_recovery_content_test.go` plus this T024 audit validate no autonomous runtime recovery support was introduced. |
| AC-D8-001 | `packages/ai-setup-go/internal/library/agent_state_content_test.go` validates lifecycle vocabulary labels. |
| AC-D8-002 | `packages/ai-setup-go/internal/library/agent_state_content_test.go` validates labels are report vocabulary only and do not imply runtime state-machine support. |
| AC-D8-003 | `packages/ai-setup-go/internal/library/agent_state_content_test.go` plus this T024 audit validate no `ChainState`/`StepState` lifecycle-label fields or `get_status` lifecycle output were introduced. |

## Scope Audit

- **Approved bounded changes:** T018 default feature-chain integration was approved in the T018 implementation request and remains a single sequential `chain-verify` step after review. T021 runtime feedback propagation was approved in the T021 implementation request and remains limited to rejected-gate `output.structuredFeedback` persistence without new outcomes or a gate engine.
- **D5 remains deferred:** Wave 2 added only static safe-recovery policy/guidance. No recovery classifier, autonomous edit loop, retry semantic change, or runtime auto-recovery artifact was introduced.
- **D8 remains deferred:** Wave 2 added only static lifecycle reporting vocabulary. No runtime lifecycle-label fields were added to `ChainState`/`StepState`, persistence, or `get_status` output.
- **Unsupported chain constructs absent:** `feature.json` and `feature-adversarial.json` remain explicit sequential `steps` chains with no `parallel` blocks, runtime `condition` fields, `optionalByFeature`, or `{{#if}}` template markers.
- **Wave 3/4 scope absent:** No RAG/model-routing/eval/learning/debate artifacts, telemetry integration, MCP server configuration change, or provider routing/billing automation was introduced by Wave 2.
- **T014 decisions reflected:** T014 recorded that D5/D8 runtime automation was deferred and that T018/T021 needed explicit approval before runtime/chain work. T024 confirms T018/T021 proceeded only through bounded approved changes, while D5/D8 runtime automation remains absent.

## Validation Evidence

- Red check for T024 trace documentation: `pnpm --filter @ai-setup/cli test -- --run src/__tests__/wave2-integration-validation.test.ts` initially failed because this task still had `**Status:** TODO` and no `## Acceptance Trace Evidence` section; after this evidence note, the test guards the final trace and scope audit.
- Wave 2 CLI content/orchestration suite: `pnpm --filter @ai-setup/cli test -- --run src/__tests__/chain-verification-report.test.ts src/__tests__/completion-enforcement.test.ts src/__tests__/generated-knowledge-env-planning.test.ts src/__tests__/structured-feedback.test.ts src/__tests__/orchestration.test.ts src/__tests__/wave2-integration-validation.test.ts` — PASS; 6 test files and 39 tests passed.
- Go/content/scaffold suite: `go test ./...` from `packages/ai-setup-go` — PASS; all Go package tests passed, including `internal/library` and `internal/scaffold`.
- Orchestrator runtime regression: `pnpm --filter @ai-setup/orchestrator test -- --run src/__tests__/chain-machine.case.ts` — PASS; 15 tests passed.
- Orchestrator typecheck: `pnpm --filter @ai-setup/orchestrator typecheck` — PASS.
- CLI/TS typecheck: `pnpm --filter @ai-setup/cli typecheck` — PASS.

## Risks

- **Partial Wave 2 appears complete:** mitigate with explicit AC trace.
- **Unapproved runtime creep:** mitigate with negative scope assertions.

## Constitution Check

- **Article I:** Reuse existing integration/snapshot tests.
- **Article II:** Scope/trace tests precede final glue changes.
- **Article III:** `spec-wave2.md` is the acceptance source.
- **Article IV:** Audit rejects Wave 3/4 and unapproved runtime work.
- **Article V:** A trace audit is simpler than new governance tooling.
- **Article VI:** No new frameworks; validation only.
