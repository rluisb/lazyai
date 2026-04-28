# Spike: [###-spike-slug]

**Spike ID:** ###
**Date:** YYYY-MM-DD
**Status:** Draft | Investigating | Concluded | Cancelled
**Owner:** [name]
**Time-box:** [N hours / days] *(spikes MUST have a time-box)*

> **Purpose.** A spike reduces uncertainty before committing to a plan. It is **research, not code**. The output is a recommendation with evidence — never production code. If the spike produces code, that code is throwaway (Anti-Slope, Rule 5).

---

## 1. Question

The spike answers a single, sharp question. If you cannot phrase it as one question, run two spikes.

**Question:** [Can we …? Should we …? Which of A vs B …?]

**Why this matters now:** [why postponing this question would be more expensive than answering it].

---

## 2. Definition of Done

The spike is **CONCLUDED** when:

- [ ] The question is answered with a recommendation: **YES / NO / DEPENDS**.
- [ ] The recommendation cites at least three pieces of objective evidence.
- [ ] Risks and unknowns are explicit, with confidence levels.
- [ ] A follow-up plan exists (either: write a spec, abandon the path, or run a deeper spike).

---

## 3. Approaches Considered (Tree of Thoughts)

At least two viable paths MUST be explored. Single-path spikes are biased.

### Path A — [Approach name]
- **Hypothesis:** [what we expect this to deliver]
- **Method:** [how we tested it: prototype / read source / benchmark / docs]
- **Findings:** [what we learned, with citations]
- **Cost / risk:** [complexity / time / dependency footprint]

### Path B — [Approach name]
- **Hypothesis:** [...]
- **Method:** [...]
- **Findings:** [...]
- **Cost / risk:** [...]

### Path C — [Approach name] *(optional)*
[…]

---

## 4. Generated Knowledge

Background information that frames the answer. Pull from documentation, ADRs, prior spikes, or external research. Cite everything.

| Source | Relevance | Citation |
|---|---|---|
| [doc / ADR / paper] | [what it tells us] | [link] |
| [code path] | [what it tells us] | [file:line] |

---

## 5. Evidence

Objective signals that drive the recommendation. **No vibes** — only measurable or directly observable evidence.

| Evidence | Path | Method | Result |
|---|---|---|---|
| [evidence name] | [Path A / B / C] | [benchmark / unit test / code read] | [number / observation] |
| [evidence name] | [...] | [...] | [...] |

---

## 6. Self-Consistency Check

Run the question through at least two independent reasoning passes (different perspectives). Record any disagreement.

- **Perspective 1 — [name]:** [conclusion]
- **Perspective 2 — [name]:** [conclusion]
- **Agreement:** YES / PARTIAL / NO. If NO or PARTIAL: [explain what differs and which we trust].

---

## 7. Reflexion — what surprised us

Pattern memory for the next spike. What did we expect to find that we didn't, or vice versa?

- **Surprise:** [observation]
- **Lesson for future spikes:** [generalizable rule]

---

## 8. Recommendation

```
RECOMMENDATION: YES / NO / DEPENDS
PATH:           A / B / C / hybrid
CONFIDENCE:     HIGH / MEDIUM / LOW
```

**Justification (3 bullets):**
- [evidence ref + reasoning]
- [evidence ref + reasoning]
- [evidence ref + reasoning]

**Conditions / caveats:** [what must be true for the recommendation to hold]

---

## 9. Follow-Up

The spike's purpose is to unblock the next step. Write that next step here.

- [ ] **Write spec:** [pointer to where `speckit-specify` will start, with the spike's findings as Generated Knowledge].
- [ ] **Abandon path:** [why and what we will do instead].
- [ ] **Deeper spike:** [next spike question].

---

## 10. Throwaway Inventory (Anti-Slope)

The spike code is research. Track what gets discarded.

| Artifact | Path | Action |
|---|---|---|
| [prototype] | [branch / dir] | DELETED on YYYY-MM-DD |
| [scratch test] | [path] | DELETED on YYYY-MM-DD |

> No spike artifact reaches `main`. The lessons reach `main` via the spec or the standards file.

---

## 11. Memory Update

- [ ] Append spike summary to `.specify/memory/repos/<repo>/ledger.md` (one line + link to this file).
- [ ] If a generalizable lesson emerged, propose a `specs/standards/` entry via `extract-standards`.
