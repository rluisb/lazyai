# ADR-[NNN]: [Decision Title]

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-[NNN]
**Deciders:** [Names]
**Constitution:** [link to active `constitution.md`]

> **Purpose.** Capture an architectural decision with the context it was made in, the alternatives considered, and the consequences accepted. ADRs are the **record of how the constitution evolves under load** — they document why a project chose one path when several were valid.

---

## Context

What is the situation that requires a decision? What forces are at play? Cite related specs, plans, spikes, prior ADRs.

[2-4 paragraphs.]

**Related artifacts:**
- Spec(s): [link]
- Spike(s): [link]
- Prior ADR(s): [link, especially any this supersedes]

---

## Constitution Alignment

Which articles bear on this decision? Cite explicitly.

| Article | Bearing | Note |
|---|---|---|
| I — Library-First | [bears / N/A] | [note] |
| IV — Anti-Speculation | [bears / N/A] | [note] |
| V — Simplicity | [bears / N/A] | [note] |
| VI — Anti-Overengineering | [bears / N/A] | [note] |

> If this ADR proposes to **amend** an article, mark "amends" and require explicit human approval before the status moves to Accepted.

---

## Options Considered (Tree of Thoughts)

At least two viable alternatives MUST be evaluated against the same axes.

### Option A — [Approach]
- **Summary:** [one sentence]
- **Complexity:** [low / med / high]
- **Reversibility:** [high / med / low]
- **Performance impact:** [observation]
- **Team familiarity:** [high / med / low]
- **Constitution fit:** [which articles support / oppose]

### Option B — [Approach]
[same axes]

### Option C — [Approach] *(optional)*
[same axes]

---

## Decision

[The chosen option, stated unambiguously.]

---

## Rationale

Why this option won today. Be specific — "it's simpler" without naming the alternative cost is not a reason.

- [reason citing evidence]
- [reason citing trade-off]

**Why the rejected options were rejected:**
- Option [X]: [specific drawback]
- Option [Y]: [specific drawback]

---

## Consequences

**Positive:**
- [benefit]

**Negative / accepted trade-offs:**
- [cost — and the threshold at which it would force re-evaluation]

**Neutral:**
- [side effect]

---

## Reversal Conditions

This ADR is not eternal. Name the signal that would prompt re-opening.

- [trigger — e.g., "if request latency p95 exceeds X for 7 consecutive days"]
- [trigger — e.g., "if the chosen library is abandoned by upstream"]

---

## Implementation Pointer

Where the decision was carried out (link to the plan / commits / PR).

- Plan: [link]
- PRs: [list]
- Standards updated: [list]

---

## Memory Update

- [ ] Append ADR ratification to `.specify/memory/repos/<repo>/ledger.md`.
- [ ] Update `KNOWLEDGE_MAP.md` (if a new architectural pattern is now canonical).
- [ ] If this supersedes a prior ADR: mark the prior one **Superseded by ADR-NNN**.
