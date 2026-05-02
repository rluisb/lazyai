# Feature 001 — Wave 2 Task Index

**Wave:** Wave 2 — Quality + Context  
**Contract:** `spec-wave2.md`  
**Plan:** `plan-wave2.md`  
**Status:** Planning only; implementation pending human approval

---

## User Stories

- **[US1] Verification + completion:** D4 chain-of-verification and N12 TillDone enforcement.
- **[US2] Context-aware reasoning:** N2 generated knowledge, D14 environment-aware planning, D13 causal reasoning.
- **[US3] Actionable feedback/recovery/state:** D9 structured feedback, D5 safe recovery policy, D8 lifecycle vocabulary.

---

## Master Task Index

| ID | Story | Task | Parallel | Depends on | Approval Gate |
|---|---|---|---|---|---|
| T014 | Setup | Wave 2 baseline capability + W1 bake audit | No | none | Records decisions |
| T015 | [US2] | Generated knowledge + environment-aware planning guidance | [P] | T014 | No |
| T016 | [US2] | Causal reasoning RCA guidance | [P] | T014 | No |
| T017 | [US1] | `chain-verify` skill + report contract | [P] | T014 | No |
| T018 | [US1] | Optional sequential feature-chain `chain-verify` integration | No | T017 | **Yes** |
| T019 | [US1] | TillDone completion enforcement guidance | [P] | T014 | No |
| T020 | [US3] | Structured feedback schema + static guidance | [P] | T014 | No |
| T021 | [US3] | Structured feedback runtime propagation | No | T020 | **Yes** |
| T022 | [US3] | Safe auto-recovery policy guidance | [P] | T014 | No runtime automation |
| T023 | [US3] | Agent lifecycle/state taxonomy guidance | [P] | T014 | No runtime state tracking |
| T024 | All | Wave 2 integration validation + scope audit | No | T015-T023 approved/completed subset | No |

`[P]` means tasks can proceed in parallel after T014 because they touch different library files and do not share runtime state. T018 and T021 must not start until the named human approval is recorded.

---

## Dependency Graph

```text
T014
├── T015 [P] ┐
├── T016 [P] ├── T024
├── T017 [P] ├── T018 ┐
├── T019 [P] ┤       ├── T024
├── T020 [P] ├── T021┘
├── T022 [P] ┤
└── T023 [P] ┘
```

---

## Runtime/Deferred Capability Ledger

| Technique | Wave 2 task | Runtime work? | Decision |
|---|---|---|---|
| D4 | T017, T018 | Chain JSON only if approved; no engine changes | Human approval before T018 |
| D9 | T020, T021 | Optional gate feedback persistence | Human approval before T021 |
| D5 | T022 | No runtime auto-recovery | Deferred separate ADR |
| D8 | T023 | No `ChainState`/`StepState` lifecycle fields | Deferred separate ADR |

---

## Acceptance Trace

| Spec AC | Tasks |
|---|---|
| AC-D4-001..004 | T017, T018, T024 |
| AC-N12-001..003 | T019, T024 |
| AC-N2-001, AC-D14-001..002 | T015, T024 |
| AC-D9-001..003 | T020, T021, T024 |
| AC-D13-001..002 | T016, T024 |
| AC-D5-001..003 | T022, T024 |
| AC-D8-001..003 | T023, T024 |
