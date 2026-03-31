<template>
  <name>ADR — Architecture Decision Record</name>
  <output>docs/adrs/NNN-title.md</output>
  <input>TechSpec decision point</input>
  <phase>Plan — Step 2 (side effect)</phase>
</template>

# ADR NNN: [Decision Title]

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded by [ADR-NNN]
**Feature:** [NNN-feature-name — or "Cross-cutting"]
**Author:** [name]

---

## Context

<!-- What situation led to this decision. Facts, not opinions. 2-3 sentences. -->

[CONTEXT]

## Decision

<!-- One clear statement. -->

We will [DECISION].

## Alignment

<!-- How this fits with what already exists. -->

- Consistent with: [existing pattern, decision, ADR, or standard — or "New direction"]
- Breaks from: [existing pattern — justify why, or "None"]

## Alternatives Considered

<!-- What else we looked at and why we rejected it.
     Most valuable section — prevents future re-litigation. -->

| Alternative | Why Rejected |
|-------------|-------------|
| [option A] | [specific reason] |
| [option B] | [specific reason] |
| Do nothing | [reason — if applicable] |

## Consequences

**Positive:**
- [consequence]

**Negative:**
- [trade-off we accept]

**Neutral:**
- [thing that changes]

---

<!-- ADR RULES
- ADRs are permanent. Never edit an accepted ADR.
- To change a decision: create a new ADR that supersedes this one.
- Mark the old ADR: Status → "Superseded by ADR-NNN"
- Keep it short. One page max. If longer, you're over-explaining.
-->
