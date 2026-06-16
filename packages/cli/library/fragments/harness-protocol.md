<harness-protocol>

# Harness Engineering Protocol

The harness is the surrounding system that turns an LLM agent from a one-shot guesser into a reliable software engineer. The constitution defines *what* is correct; the harness defines *how* correctness is enforced across the lifecycle of work.

Five rules govern every workflow in this library:

1. **Feed Forward** — every step is constrained by an upstream blueprint.
2. **The Contract** — work is verified by a separate evaluator, not the author.
3. **Feedback & Sensors** — quality is measured against objective reality, not vibes.
4. **Memory & State** — durable records survive context loss and session boundaries.
5. **Anti-Slope** — regressions and prototype rot are prevented by structural gates.

Skills and agents reference this fragment instead of restating the rules.

---

## Rule 1 — Feed Forward (the Blueprint)

**Definition.** No phase begins without a written, approved blueprint from the previous phase. Constraints flow downstream; they never appear out of thin air mid-implementation.

**The chain.**
```
constitution.md  →  spec.md  →  plan.md  →  tasks.md  →  task-harness.md  →  code
   (governing)      (what)      (how)        (steps)      (this step)       (output)
```

**Concrete case study — Spec-Driven Development.** Before any task is written, `spec.md` exists and is human-approved. Before any code is written, `plan.md` and `tasks.md` exist and are human-approved. Each task carries a `task-harness.md` that snapshots the state of the world (tool versions, file paths, patterns to follow) so the implementing agent never improvises constraints.

**How agents apply it.**
- Before writing: read every upstream document and quote the constraints you intend to honor.
- During writing: when a constraint is unclear or contradictory, **stop** and escalate — do not invent a resolution.
- After writing: confirm every produced section maps to an upstream requirement.

**Failure mode this prevents.** "Whatever-the-LLM-felt-like" implementations that pass tests but break the spec.

---

## Rule 2 — The Contract (Dual-Agent Simulation)

**Definition.** The agent that produces an artifact MUST NOT be the agent that approves it. Generator and verifier are separate roles, even when run by the same model — context is split, prompts are distinct, and the verifier reads only the artifact and the contract, never the generator's reasoning.

**The pattern.**
```
generator  →  artifact  →  verifier  →  pass / fail
                ↑                          │
                └──── feedback loop ◄──────┘
```

**Concrete case study — LLM-as-Judge for code review.**
Multiple review lenses (Test Quality, Contract, Patterns, Performance/Security, Simplicity Audit) run in parallel against the same diff. A judge agent synthesizes their findings, deduplicates conflicts, prioritizes by severity, and produces one structured report citing the constitution articles each finding violates.

**How agents apply it.**
- Generator agents (planner, implementer) MUST emit artifacts that name their contract (the spec and articles they implement).
- Verifier agents (reviewer, red-team) MUST evaluate against the named contract — never against private opinion.
- Verifiers escalate to a human gate when generator and verifier disagree after one feedback cycle.

**Failure mode this prevents.** Agent self-approval, motivated reasoning, and the "looks good to me" rubber stamp.

---

## Rule 3 — Feedback & Sensors (Objective Reality)

**Definition.** Every claim of "done" is backed by a measurable signal: a passing test, a green CI run, a captured metric, a human approval. Self-report is not evidence.

**The 5-gate ladder is the canonical sensor stack.**
| Gate | Sensor | Signal |
|---|---|---|
| Static Integrity | linter, type-checker | exit code 0 |
| Contract Compliance | reviewer / verifier | structured pass/fail per article |
| Behavioral Validation | test runner | suite green + coverage threshold |
| Pattern Consistency | reviewer + standards file | no novel patterns without ADR |
| Observability Readiness | runtime telemetry plan | logs/metrics/rollback exist |

**Concrete case study — Bugfix RCA.**
A bugfix is "done" only when:
1. A regression test exists that **fails** on the unfixed code (pre-fix evidence).
2. The same test **passes** on the fixed code (post-fix evidence).
3. CI is green across the rest of the suite (no-regression evidence).

**How agents apply it.**
- Cite the sensor for every "done" claim. "Tests pass" → quote the command and the result.
- When a sensor is missing, build it before claiming success (telemetry, regression test, etc.).
- Treat a green sensor as necessary but not sufficient — pair with the Contract from Rule 2.

**Failure mode this prevents.** "I think it works" and "looks good to me" closing tickets.

---

## Rule 4 — Memory & State (Beating Amnesia)

**Definition.** Context windows are ephemeral; project state is not. Every session ends with a durable, versioned record that the next session can resume from without re-reading the entire history.

**The persistence layers.**
| Layer | Lifespan | What lives here |
|---|---|---|
| Conversation | minutes | active reasoning, scratchpad |
| Plan / tasks | days | current iteration's roadmap |
| Ledger (`.specify/memory/repos/<name>/ledger.md`) | weeks-months | per-repo decisions, completed work, open follow-ups |
| Last-known-state (`.specify/memory/repos/<name>/last-known-state.md`) | weeks-months | branch, dirty files, paused work |
| Constitution / standards | indefinite | governing rules |

**Concrete case study — Multi-session feature.**
A feature spanning 4 sessions never loses state because:
- Session 1 writes spec.md, ledger entry "spec drafted", last-known-state "branch feat/X".
- Session 2 reads ledger + last-known-state first, writes plan.md, appends ledger entry.
- Session 3 implements task 1, appends ledger, updates last-known-state with progress.
- Session 4 reads forward from session 3 with zero re-derivation.

**How agents apply it.**
- Every workflow MUST update the relevant ledger on completion (skill: `update-memory`).
- Every session MUST start by reading the latest `last-known-state.md` for the active repo.
- Decisions made in conversation are not real until written to a durable file.

**Failure mode this prevents.** Re-deriving the same plan three sessions in a row, contradicting last week's decision because nobody wrote it down.

---

## Rule 5 — Anti-Slope Protocol

**Definition.** Quality erodes silently. The harness must actively prevent regression — old patterns sneaking back, prototypes shipping to production, fixed bugs reappearing.

**The structural gates.**
- **Branch protection.** Main is read-only; merges require passing the 5-gate ladder.
- **Regression tests.** Every fixed bug carries a permanent test that would re-fail if the bug returned.
- **Throw away PoCs.** A proof-of-concept is research, not code. The PoC branch is deleted after the lessons are extracted into a spec.
- **Standards as memory.** When a pattern is rejected at review, it is recorded in `specs/standards/` so the same pattern is rejected next time without re-debating.
- **Constitution amendments are versioned.** Articles are not silently re-interpreted; they are amended with an ADR.

**Concrete case study — PoC discipline.**
After a PoC validates feasibility, the PoC code is **discarded** and a fresh implementation is written under the full SDD workflow. The PoC's value is the lessons learned, not the lines of code. Shipping PoC code is a structural anti-pattern this rule blocks.

**How agents apply it.**
- Reviewers MUST flag re-emergence of patterns the standards file rejects.
- Implementors MUST NOT promote PoC code to main; they re-implement under TDD.
- When a sensor degrades (flaky test, declining coverage), it is fixed, not silenced.

**Failure mode this prevents.** Slow drift, prototype rot, fixed bugs reappearing, and "we've always done it this way" overriding the constitution.

---

## How skills and agents declare harness compliance

Every skill in `library/skills/` declares which harness rules it enforces, in its frontmatter:

```yaml
harness:
  feed_forward: [constitution.md, spec.md]
  contract: [reviewer-judge]
  sensors: [gate-1, gate-3]
  memory: [ledger.md, last-known-state.md]
  anti_slope: [regression-test-required]
```

Agents reading this fragment can verify a workflow is harness-compliant by checking that every rule appears in the chain at least once between specify and merge.

</harness-protocol>
