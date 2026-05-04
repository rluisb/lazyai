# [YOUR_PROJECT_NAME] Constitution

> **Purpose.** This document is the governing contract for every workflow, agent, and skill in this project. AI agents MUST read this file before specifying, planning, implementing, or reviewing any change. Violations are blocking findings, not suggestions.

---

## Core Principles

### Article I — Library-First Development

All solutions MUST leverage existing, proven libraries before custom implementation.

**Rationale.** LLMs produce more reliable code when working with well-documented APIs. Custom code is a long-term liability; a well-maintained dependency is a delegated problem.

**How it applies.** Before writing a new utility, helper, or abstraction:
1. Search the codebase for existing helpers (`internal/files`, `internal/configmerge`, etc.).
2. Search the dependency manifest for an installed library that already covers the use case.
3. Only build new code when no library exists or when the library is unsafe/abandoned.

### Article II — Test-First Imperative (NON-NEGOTIABLE)

All implementation MUST follow Test-Driven Development.

**The cycle is:**
1. **RED** — write a failing test that encodes the new behavior or bug.
2. **GREEN** — write the smallest amount of production code that turns the test green.
3. **REFACTOR** — improve names, structure, and duplication while the suite stays green.

**Rules.**
- Tests are written **before** the production code they cover. No "implement first, test later."
- Every bugfix MUST land with a regression test that fails on the old code and passes on the new.
- Tests serve as executable specifications: a feature without tests is not "done."
- This article is **non-negotiable**: a PR without test-first evidence MUST be rejected at review.

### Article III — Docs as Source of Truth

Documentation drives implementation, not the reverse.

**How it applies.**
- For new features, the `spec.md` is written and approved **before** the `plan.md`.
- For architecture changes, the ADR is written **before** the refactor.
- When code and docs disagree, the docs are the contract — the code is the bug, until the docs are explicitly updated.
- Public APIs, slash commands, and skill outputs MUST have documentation that matches the implementation.

### Article IV — Anti-Speculation (YAGNI)

Never implement features, options, or hooks that were not explicitly requested.

**How it applies.**
- Do not add config knobs "in case someone needs them."
- Do not write fallbacks for scenarios that cannot occur given the current call graph.
- Do not introduce abstractions to support future requirements that have not been agreed.
- Trust internal callers and framework guarantees. Validate only at system boundaries (user input, external APIs).

If you find yourself writing `if userIsLoggedIn` in a code path that is only reached by logged-in users, delete the check.

### Article V — Simplicity Over Abstraction

Prefer concrete implementations over abstract patterns. Avoid premature abstraction.

**How it applies.**
- Three similar lines of code is not duplication — it is three concrete cases. Wait for the fourth before extracting.
- A function that is called once does not need to be a function. Inline it.
- A class with a single implementation does not need an interface. Drop the interface.
- A switch statement with two cases does not need a strategy pattern. Leave the switch.

### Article VI — Anti-Overengineering (NON-NEGOTIABLE)

This article makes the previous two operational and adds the policies AI agents most often violate.

**1. YAGNI ("You Aren't Gonna Need It").** Build only what the current task requires. No speculative parameters, no "extensible" base classes, no plugin systems for the one thing that exists today.

**2. DRY — but only after the third instance.** Do not extract a helper at duplication 1. Do not extract at duplication 2. Extract at duplication 3, when the shape is stable. Extracting too early couples unrelated call sites and forces every future caller to fight the abstraction.

**3. KISS ("Keep It Simple, Stupid").** When two solutions both work, choose the simpler one. Optimize for the reader of the next change, not the cleverness of the current one.

**4. Clean Code (subset).**
- Functions ≤ 30 lines unless there is a documented reason.
- Files ≤ 300 lines unless there is a documented reason.
- One responsibility per function. One concept per file.
- Names tell the reader what — comments explain only the why.

**5. Unix Philosophy.** Each module does one thing well, communicates through clear interfaces, and composes with others. No god-objects, no mega-skills.

**Enforcement.** Reviewer agents MUST fail any change that:
- Adds an abstraction with one caller.
- Adds a function that exceeds 30 lines without a justification comment.
- Adds dead code paths reachable only by hypothetical futures.
- Re-introduces a pattern this article forbids.

This article is **non-negotiable**.

---

## Technology Constraints

The technology constraints define the boundary fence for the project. Anything outside is forbidden without an ADR.

### Approved
- **Languages:** [YOUR_APPROVED_LANGUAGES]
- **Frameworks:** [YOUR_APPROVED_FRAMEWORKS]
- **Runtimes:** [YOUR_APPROVED_RUNTIMES]
- **Test stacks:** [YOUR_APPROVED_TEST_STACKS]

### Forbidden
- No secrets committed to source control. Ever.
- No string-concatenated SQL or shell commands. Always parameterized.
- No direct DOM mutation outside the agreed UI framework.
- No circular dependencies between packages.
- No skipping pre-commit hooks (`--no-verify`) without explicit human approval.

### Confidence Markers
When approach confidence is below high, mark it inline:

- `[CONFIDENCE: HIGH]` — proven approach, proceed.
- `[CONFIDENCE: MEDIUM]` — reasonable, plan validation alongside implementation.
- `[CONFIDENCE: LOW]` — spike or PoC required before committing to the path.

Low-confidence items MUST be resolved (upgraded or abandoned) before merging to main.

---

## Quality Gates — The 5-Gate Ladder

Every change passes through five gates, in order. A failure at any gate blocks promotion to the next.

### Gate 1 — Static Integrity
*Runs:* on every save and pre-commit.
*Checks:* lint, format, type-check, dead-code analysis, dependency hygiene.
*Pass criteria:* zero errors, zero new warnings.
*Fail criteria:* any error or new warning blocks the gate.

### Gate 2 — Contract Compliance
*Runs:* before merge, against the spec/plan/task contract.
*Checks:*
- Public API matches the documented contract.
- New code obeys the constitution (Articles I-VI), especially Anti-Overengineering: no speculative options, no premature abstractions, no dead branches.
- All declared `output_schema` sections are present in produced artifacts.
*Pass criteria:* contract diff is empty; reviewer confirms YAGNI/DRY/KISS adherence.
*Fail criteria:* missing/extra sections, undeclared side effects, or contract drift.

### Gate 3 — Behavioral Validation
*Runs:* on PR, after Gates 1-2 pass.
*Checks:* unit tests, integration tests, acceptance criteria from `spec.md`, regression test for any bugfix.
*Pass criteria:* full test suite green; new behavior covered; coverage thresholds met (see below).
*Fail criteria:* any failing test, any uncovered new branch, missing regression test.

### Gate 4 — Pattern Consistency
*Runs:* in code review.
*Checks:* code uses existing helpers, naming conventions, error-handling patterns, and project standards (`specs/standards/`).
*Pass criteria:* no out-of-pattern code without a recorded ADR.
*Fail criteria:* novel pattern introduced without justification, or standards violation.

### Gate 5 — Observability Readiness
*Runs:* before deploy / release.
*Checks:* logging, metrics, error reporting, and rollback plan exist for the change.
*Pass criteria:* operator can detect and roll back the change without reading source.
*Fail criteria:* silent failure modes, no metrics, no rollback path.

### Coverage Thresholds
| Metric | Target | Minimum |
|--------|--------|---------|
| Line coverage | [YOUR_LINE_COVERAGE_TARGET] | [YOUR_LINE_COVERAGE_MINIMUM] |
| Branch coverage | [YOUR_BRANCH_COVERAGE_TARGET] | [YOUR_BRANCH_COVERAGE_MINIMUM] |
| Build time | [YOUR_BUILD_TIME_TARGET] | [YOUR_BUILD_TIME_MAXIMUM] |

---

## Governance

- **Amendments.** Constitutional amendments require explicit human approval and a recorded ADR in `specs/adrs/`.
- **Precedence.** When this constitution conflicts with a skill, agent, or template, this constitution wins.
- **Reviewer obligation.** Every reviewer (human or agent) MUST cite the article a finding is grounded in. Findings without an article reference are advisory only.
- **Living document.** The constitution evolves with the project. Stale articles are a defect — flag them for amendment.

**Version:** 1.0.0
**Ratified:** [DATE]
