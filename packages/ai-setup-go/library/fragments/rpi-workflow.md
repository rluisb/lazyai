<rpi-workflow>

# RPI Workflow — Harness-Aligned Research → Plan → Implement

RPI is the lighter sibling of Spec-Driven Development (SDD). Use it for non-trivial but bounded work that does not need a full `spec.md` (e.g., refactoring a module, adding a small feature, paying down tech debt). For new product features, use SDD instead.

Every RPI run passes through five phases. Each phase has an entry blueprint, a sensor, a memory update, and a human gate at its boundary.

---

## Phase 0 — Feed Forward

**Goal.** Establish the constraints before reasoning starts.

**Inputs (read first):**
- `<workspace>/.specify/memory/constitution.md` (or repo-local equivalent).
- `<workspace>/.specify/memory/repos/<active-repo>/last-known-state.md`.
- Relevant standards in `specs/standards/`.

**Outputs.**
- One paragraph stating the task in your own words.
- A list of constitutional articles that constrain the task.
- Confidence level: HIGH / MEDIUM / LOW.

**Sensor.** The task statement is a faithful re-statement of the request — confirmed by the human or escalated.

**Skip rule.** Trivial changes (<20 lines, typos, dependency bumps, doc-only) skip RPI entirely.

⛔ **Human gate:** confirm scope before research begins.

---

## Phase 1 — Research

**Goal.** Understand the system before proposing changes.

**Activities.**
- Identify affected files, packages, callers, and tests.
- Read related ADRs in `specs/adrs/`.
- Search for existing helpers, patterns, and standards (Article I — Library-First).
- Note unknowns and mark each as HIGH / MEDIUM / LOW confidence.

**Output:** `specs/<task>/research.md` with: Affected Surface, Existing Patterns, Unknowns, Risks.

**Sensor.** Every claim about existing code cites a file path and (where useful) line number.

**Anti-Slope check.** If research reveals a pattern this project explicitly rejected (see `specs/standards/`), abort and escalate — do not "rediscover" rejected solutions.

⛔ **Human gate:** approve research before planning.

---

## Phase 2 — Plan

**Goal.** Define the approach, the acceptance criteria, and the task breakdown.

**Activities.**
- Define acceptance criteria as bullet-list "the system shall…" statements.
- Break work into tasks of ≤ 1 session each. Order by dependency.
- For each task, name: inputs, outputs, tests to write first (Article II), risks.
- If multiple approaches exist, run the Decision Protocol (ToT) and record the rejected options.

**Output:** `specs/<task>/plan.md` and `specs/<task>/tasks.md`. Optional: a task-harness file per task.

**Sensor.** Every acceptance criterion has at least one task that satisfies it.

**Constitution check.** The plan MUST cite Articles IV (YAGNI), V (Simplicity), and VI (Anti-Overengineering) explicitly — listing what was deliberately *not* added and why.

⛔ **Human gate:** approve the plan before implementation begins.

---

## Phase 3 — Implement

**Goal.** Execute the plan one task at a time with continuous sensor feedback.

**Per-task TDD loop (Article II).**
1. **RED** — write the failing test described in the task.
2. **GREEN** — write the smallest production change that turns it green.
3. **REFACTOR** — improve names, structure, duplication; keep the suite green.
4. **GATE** — run the 5-gate ladder (lint, contract, behavior, patterns, observability).
5. **COMMIT** — small, conventional, scoped (`feat(scope):`, `fix(scope):`).

**Anti-Overengineering hold.** Before each commit, ask:
- Is anything in this commit speculative? Delete it.
- Is anything extracted before the third instance? Inline it.
- Is anything > 30 lines / > 300 lines without a documented reason? Split it.

**Pivot handling.** If implementation reveals the plan is wrong:
1. STOP the current task.
2. Document why in the ledger.
3. If architecture is affected, write an ADR.
4. Return to Phase 1 with new information.

**Sensor.** All five gates green for the task; tests authored before code (RED-first git history).

⛔ **Checkpoint** after each task: confirm progress, allow human to redirect.

---

## Phase 4 — Feedback

**Goal.** Verify objective reality matches the plan, then persist what was learned.

**Activities.**
- Run the full test suite one more time from a clean state.
- Run a `review` skill against the change (LLM-as-Judge synthesizer over multiple lenses).
- Record findings; address blockers, log non-blockers as follow-ups.
- Update memory:
  - Append a `ledger.md` entry: date, who, what, plan link, verified status.
  - Rewrite `last-known-state.md` for the repo.
- Run `self-improve` on the session's transcript: did any pattern emerge that should become a standard?

**Sensor.** Reviewer report is structured (per-finding severity + cited article); ledger entry exists; last-known-state is current.

**Anti-Slope check.** No PoC code shipped to main. Every fixed bug carries a regression test that fails on the pre-fix code.

⛔ **Final human gate:** approve before merging.

---

## Phase summary table

| Phase | Input | Output | Sensor | Memory write |
|---|---|---|---|---|
| 0 Feed Forward | constitution, last-known-state | task statement, confidence | restated faithfully | (read only) |
| 1 Research | task statement | research.md | every claim cites a path | (read only) |
| 2 Plan | research.md | plan.md, tasks.md | every AC has a task | (read only) |
| 3 Implement | tasks.md | code + tests | 5 gates green | per-commit |
| 4 Feedback | code + tests | review report | reviewer pass | ledger + last-known-state |

</rpi-workflow>
