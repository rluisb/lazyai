---
name: war-council
description: Multi-advisor debate system for high-impact decisions. 6 archetype subagents debate architecture, technology, and product choices. 7-phase process with steel-manning, dissent preservation, and position evolution tracking.
trigger: /war-council
triggers:
  - "convene the council"
  - "debate this decision"
  - "council review"
  - "architecture decision"
  - "evaluate trade-offs"
skill_path: skills/war-council
---

## Quick Reference

| | |
|---|---|
| **Use when** | High-impact architecture decisions, technology trade-offs |
| **Do not use when** | Routine implementation, low-stakes choices |
| **Primary agent** | engine-control |
| **Runtime risk** | High — expensive multi-advisor loading |
| **Outputs** | Synthesis document, dissent preservation, trade-off log |
| **Validation** | Steel-manning, evidence requirements |
| **Deep mode trigger** | `/war-council` with MODE=premium — **explicit-only, never auto-loaded** |
| **Invocation**        | Tier 4 expensive/debate mode — must be explicitly invoked via engine-control |

# War Council — Multi-Advisor Debate System

> *"No victory is won by a single voice. The storm respects only those who have seen every angle of the battlefield."*

## Purpose

When the stakes are high and the storm is closing in, unilateral decisions get you eliminated. The War Council exists to **stress-test critical choices** before you commit materials, time, or reputation.

Convene the council when:
- Choosing between competing architectures or technologies
- Evaluating high-risk product decisions with long-term consequences
- Facing trade-offs where every option has significant downsides
- Making decisions that affect multiple squads, systems, or stakeholders
- You need **documented rationale** for future reference or audit

The council does not replace human judgment — it **sharpens** it. Six archetypes debate, dissent, and evolve their positions so you enter the battlefield with eyes wide open.

---

## Council Members

| Archetype | Perspective | Strengths |
|-----------|-------------|-----------|
| **pragmatic-engineer** | "What can we ship this week?" | Delivery velocity, incremental value, technical debt awareness, feasibility under constraints |
| **architect-advisor** | "What does this look like at 10x scale?" | System design, long-term maintainability, coupling/decoupling, evolutionary architecture |
| **security-advocate** | "How does this get us eliminated?" | Threat modeling, blast radius, compliance, least privilege, defense in depth |
| **product-mind** | "Does the player actually want this?" | User value, market fit, feature complexity vs. impact, onboarding friction |
| **devils-advocate** | "Why is this a terrible idea?" | Cognitive bias detection, hidden assumptions, worst-case scenarios, sunk cost fallacy |
| **the-thinker** | "What are we not seeing?" | Second-order effects, emergent behavior, cross-domain patterns, philosophical consistency |

Each member speaks from their archetype. No member claims to be "right" — they claim to be **heard**.

---

## The 7-Phase Process

### Phase 1: Scope
**Goal:** Define the decision boundary.

- What exactly are we deciding?
- What is in scope? What is explicitly out of scope?
- What constraints are non-negotiable?
- What would "success" look like in 30/90/365 days?

*Output:* Scoped decision brief, constraints list, success criteria.

### Phase 2: Select
**Goal:** Identify the options on the table.

- Brainstorm all viable alternatives (minimum 2, ideally 3+)
- Include the "do nothing" and "defer" options
- Tag each option with initial assumptions

*Integration:* Uses **storm-scout** to research and surface options from the vault, codebase, and external sources.

*Output:* Options list with assumptions and preliminary evidence.

### Phase 3: Opening Statements
**Goal:** Each council member presents their initial position.

- Each archetype speaks once, uninterrupted
- They declare their preferred option and why
- Evidence is required — gut feelings are noted but flagged

*Output:* Six initial positions with supporting rationale.

### Phase 4: Tensions
**Goal:** Surface conflicts between positions.

- Identify pairwise disagreements (e.g., pragmatic-engineer vs. architect-advisor on timeline)
- Map tensions to specific trade-offs
- No resolution yet — only **naming** the conflict

*Output:* Tension map showing who disagrees with whom and why.

### Phase 5: Rebuttals (Steel-Manning Required)
**Goal:** Strengthen the opposing view before attacking it.

- **Mandatory steel-manning:** Before criticizing a position, you must restate it in its strongest form
- Only then may you present counter-evidence or expose weaknesses
- The goal is not to "win" — it is to find the **strongest version** of every option

*Output:* Steel-manned positions, identified weaknesses, and surviving arguments.

### Phase 6: Position Evolution
**Goal:** Track how each member's view changes.

- Each archetype declares: "I have changed my position" or "I stand firm"
- If changed: explain what evidence or argument caused the shift
- If firm: explain what would need to be true to change your mind

*Output:* Position evolution log with delta tracking.

### Phase 7: Synthesis
**Goal:** Produce a final recommendation without false consensus.

- Majority position is identified (if one exists)
- Dissenting opinions are **preserved verbatim**
- Trade-offs are explicitly documented
- Revisit triggers are defined: "Reconvene if X happens"

*Output:* Final synthesis document (see format below).

---

## Modes

| Mode | Phases | Use When | Output |
|------|--------|----------|--------|
| **full** | All 7 phases | High-stakes, irreversible, or cross-squad decisions | Complete synthesis with full debate record |
| **quick** | Scope → Opening → Synthesis | Time-boxed, lower-stakes, or urgent decisions | Condensed synthesis with key tensions noted |
| **embedded** | Synthesis only | The council has already debated; you need the recommendation | Synthesis document only, no debate replay |

---

## Synthesis Output Format

Every council session produces a structured recommendation:

```
## Decision: [Title]

### Majority Position
[The option favored by the majority of council members]

### Rationale
[Why this option won — evidence, trade-offs, and key arguments]

### Dissenting Opinions
- **[Archetype]:** [Their position and why they disagree]
- **[Archetype]:** [Their position and why they disagree]

### Trade-offs Accepted
- [Trade-off 1]: [What we gain vs. what we sacrifice]
- [Trade-off 2]: [What we gain vs. what we sacrifice]

### Risks and Mitigations
- [Risk]: [Mitigation strategy]

### Revisit Triggers
- Reconvene if: [Condition 1]
- Reconvene if: [Condition 2]

### Evidence Log
- [Source/argument and who presented it]
```

---

## Integration

| System | Integration Point |
|--------|-----------------|
| **storm-scout** | Phase 2 (Select) — research options, gather evidence, surface prior art from the vault |
| **workflow-engine** | Council sessions can be registered as workflow steps; synthesis feeds into downstream implementation tasks |
| **truth-chain** | Every council session is logged as an immutable ledger entry with SHA-256 hash; decisions are auditable |
| **battle-bus** | Council outputs feed into battle-bus blueprints as pre-implementation gates; parallel waves may require council sign-off |

---

## Rules of Engagement

1. **Steel-manning is mandatory.** No member may attack a position they have not first strengthened. Weak opponents make weak decisions.

2. **Dissent is preserved.** The synthesis document must include every dissenting opinion, even from a single archetype. Silence dissent and you silence truth.

3. **No false consensus.** If the council is split, say so. A documented disagreement is better than a unanimous lie.

4. **Evidence is required.** "I feel" is noted but flagged. "Here is the data" is weighted. "Here is a prior failure with the same pattern" is gold.

5. **Positions must evolve or explain why not.** If no evidence changes any mind, the debate was shallow. If minds change, log what changed them.

6. **The human decides.** The council advises. The human builder, architect, or squad lead makes the final call — but they do so with full visibility into what they are accepting and rejecting.

---

## Invocation

```
/war-council [mode] [decision topic]
```

Examples:
- `/war-council full "Should we migrate from REST to GraphQL?"`
- `/war-council quick "Evaluate three auth providers"`
- `/war-council embedded` (returns last synthesis)

---

*The storm does not care about your confidence. It cares about your preparation.*
