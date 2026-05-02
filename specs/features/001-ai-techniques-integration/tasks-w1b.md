# Tasks: Feature 001 — Phase W1.B

## Task Index

| # | ID | Name | User Story | Files | Dependencies | [P] |
|---|----|------|------------|-------|--------------|-----|
| 8 | T8 | Chain gate/conditional capability tests | P3 | Orchestrator chain-machine tests, Go/TS orchestration scaffold tests | W1.A complete | — |
| 9 | T9 | Inline plan-quality checklist + report tests | P3 / D6 | Planner skill source, plan-quality report tests/fixtures | T8 | — |
| 10 | T10 | Sequential `plan-gate` chain step | P3 / D6 | Source `feature.json`, chain-shape tests | T9 | — |
| 11 | T11 | `red-team-plan` skill + soft-fail tests | P3 / D17 | New red-team skill source, skill/report tests | T8 | — |
| 12 | T12 | `adversarialDesign` flag + generated chain tests | P3 / D17 | Presets/wizard/schema, orchestration generation, source `feature.json` conditional/variant support | T10, T11 | — |
| 13 | T13 | W1.B gate report integration tests | P3 / D6+D17 | End-to-end chain/gate/report tests | T10, T12 | — |

No W1.B task is marked [P]. T9 and T11 have partially disjoint file scopes after T8, but the report contracts converge at `plan-gate`; keeping the implementation order sequential reduces coordination risk and matches the W1.B chain constraint.

## Dependency Graph

```text
W1.A merged complete
  │
  ▼
T8 chain gate/conditional capability tests
  ├──► T9 inline plan-quality checklist + report tests ───► T10 sequential plan-gate chain step ─┐
  │                                                                                              │
  └──► T11 red-team-plan skill + soft-fail tests ─────────► T12 adversarial flag + chain tests ──┤
                                                                                                 ▼
                                                                         T13 W1.B gate integration tests
```

Hard dependencies:
- T8 blocks all `feature.json` edits because conditional/gate behavior must be characterized first.
- T9 blocks T10 because `plan-quality` must have a tested report contract before the chain points at it.
- T10 blocks T12 in this worktree because both tasks may touch `packages/ai-setup-go/library/orchestration/chains/feature.json`.
- T11 blocks T12 because red-team inclusion must target an existing, tested skill/report contract.
- T13 depends on T10 and T12 because it verifies the final human gate report with plan-quality and optional adversarial findings.

## W1.B Acceptance Coverage

| AC | Covered by |
|---|---|
| AC-D6-001 | T9, T10, T13 |
| AC-D6-002 | T10, T13 |
| AC-D6-003 | T9 |
| AC-D6-004 | T9, T10, T13 |
| AC-D6-005 | T9 |
| AC-D6-006 | T9, T13 |
| AC-D17-001 | T12 |
| AC-D17-002 | T12 |
| AC-D17-003 | T11, T12, T13 |
| AC-D17-004 | T8, T12 |
| AC-D17-005 | T11, T13 |
| AC-D17-006 | T10, T13 |

## Scope Guardrails

- Worktree only: `/Users/ricardo/projects/teachable/ai-setup/.worktrees/feature-001-ai-techniques-w1b`.
- Docs created by this planner are the only changes in this planning task.
- W1.B implementation must not modify W1.A task files or re-open W1.A scope unless a later approved task explicitly changes the contract.
- The source chain is `packages/ai-setup-go/library/orchestration/chains/feature.json`; generated installs are expected at `.ai/orchestration/chains/feature.json`.
- The planner skill source is `packages/ai-setup-go/library/skills/plan.md`.
- The red-team plan skill source is `packages/ai-setup-go/library/skills/red-team-plan.md` and does not exist yet.
- `specs/templates/` does not exist; do not invent template deliverables for W1.B.

## Constitution Gate

- **Article I — Library-First:** W1.B reuses the existing sequential chain machine, existing planner skill, existing red-team role/skill format, existing preset/schema/wizard surfaces, and existing test harnesses.
- **Article II — Test-First:** Every task starts by adding failing tests/fixtures before production library or chain changes.
- **Article III — Docs as Source of Truth:** All tasks cite `spec.md` AC-D6/AC-D17 and `plan.md` W1.B sequencing; task files are the implementor contract.
- **Article IV — YAGNI:** No parallel blocks, no standalone `plan-validate` skill, no automatic fail loop, no speculative runtime condition system without tests.
- **Article V — Simplicity:** The chain remains sequential with one explicit `plan-gate`; conditional red-team support is either proven or handled by static generated variants.
- **Article VI — Anti-Overengineering:** Avoid one-caller abstractions and broad validators; keep report schemas concrete and bounded to W1.B.