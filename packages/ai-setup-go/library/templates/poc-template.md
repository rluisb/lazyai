# Proof of Concept: [###-poc-slug]

**PoC ID:** ###
**Date:** YYYY-MM-DD
**Status:** Draft | Building | Validated | Failed
**Owner:** [name]
**Time-box:** [N hours / days]
**Discard date:** [target date for branch deletion]

> **Purpose.** A PoC validates **feasibility**, not architecture. It is the smallest possible implementation that answers the question "can this work in our world?" — and it will be **deleted** when the answer is reached. The lessons feed `speckit-specify`; the code does not feed `main`.

---

## 1. Feasibility Question

What ONE thing must we prove possible before committing to a feature spec?

**Question:** [Can we integrate X with Y under constraints Z?]

**Why a PoC and not a spike?** [Spike = research; PoC = minimal running code. Justify the code-level investigation.]

---

## 2. Success Criteria

Concrete signals that the PoC has **succeeded** or **failed**. No middle ground.

- [ ] **SC-1:** [observable, binary signal — e.g., "request returns 200 with payload shape Z"].
- [ ] **SC-2:** [observable, binary signal].
- [ ] **SC-3:** [observable, binary signal].

**Failure criteria** *(when to stop):*
- [ ] [signal — e.g., "library X cannot run in our runtime"].
- [ ] [signal — e.g., "throughput below 100 req/s on baseline hardware"].

---

## 3. Scope (Mini-RPI)

A PoC is the lightest possible Research → Plan → Implement loop.

### Research (5-15% of time-box)
- [ ] Read [docs / source / examples] to understand the surface.
- [ ] Identify the smallest end-to-end path that exercises the feasibility question.

### Plan (5-10% of time-box)
- [ ] List the 3-5 steps needed to reach a yes/no answer.
- [ ] Name the throwaway scaffolding needed (mocks, fixtures, dummy data).

### Implement (rest of time-box)
- [ ] Wire the smallest path. **No tests beyond the success criterion check.**
- [ ] Stop on first conclusive signal.

---

## 4. What This PoC Does NOT Do

Anti-Speculation (Article IV) plus PoC-specific anti-scope:

- ❌ No production-quality error handling.
- ❌ No abstractions for "future flexibility."
- ❌ No tests beyond the success-criteria check.
- ❌ No styling, no UX polish, no logging unless required for the signal.
- ❌ No code review, no merge to main.

> If you find yourself adding any of the above, the PoC has graduated — **stop**, archive findings, and start `speckit-specify` instead.

---

## 5. Findings

Filled in as the PoC runs.

| Date | Observation | Implication |
|---|---|---|
| YYYY-MM-DD | [what was tried, what happened] | [what this means for the question] |
| YYYY-MM-DD | [...] | [...] |

---

## 6. Verdict

```
VERDICT:    FEASIBLE / NOT FEASIBLE / FEASIBLE WITH CAVEATS
CONFIDENCE: HIGH / MEDIUM / LOW
TIME SPENT: [actual time vs. time-box]
```

**Evidence backing the verdict:**
- [SC-1 result + path]
- [SC-2 result + path]
- [SC-3 result + path]

---

## 7. Lessons Extracted

What does the team know now that it didn't before? These lessons are the PoC's permanent product.

- **Constraint discovered:** [statement] — must appear in the future spec.
- **Risk discovered:** [statement] — must appear in the future plan's Risks table.
- **Pattern that worked:** [pointer] — candidate for `specs/standards/`.
- **Pattern that failed:** [pointer] — candidate for `specs/standards/` as a "rejected" entry.

---

## 8. Anti-Slope — Discard Plan (NON-NEGOTIABLE)

The PoC code MUST be discarded. The lessons survive; the code does not.

- [ ] PoC branch identified: `[branch name]`.
- [ ] Discard date set: `YYYY-MM-DD`.
- [ ] No code from this branch has been cherry-picked into other branches.
- [ ] On discard date: branch deleted, no orphaned files in `main`.

> Reviewers: if PoC code reaches a non-PoC branch without a fresh implementation, REJECT the change. Promotion of PoC code to production is a structural anti-pattern this template blocks.

---

## 9. Follow-Up

What happens after the verdict? The PoC's output is the *next workflow's* input.

- [ ] **If FEASIBLE:** start `speckit-specify` using lessons from §7 as Generated Knowledge.
- [ ] **If NOT FEASIBLE:** record decision in `ledger.md`; consider a deeper spike of an alternative path.
- [ ] **If FEASIBLE WITH CAVEATS:** spec MUST address each caveat as a constraint or risk.

---

## 10. Memory Update

- [ ] Append PoC verdict to `.specify/memory/repos/<repo>/ledger.md` (one line + link).
- [ ] Lessons in §7 surfaced to next spec's Generated Knowledge.
- [ ] Discard executed on the planned date.
