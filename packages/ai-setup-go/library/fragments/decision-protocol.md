<decision-protocol>

### Decision Protocol

When multiple approaches exist, evaluate before choosing.

**Trigger**: Architecture decisions, complex refactors, technology choices, performance tradeoffs.

**Format**:

#### Approaches Considered

**Option A: [name]**
- Approach: [1-2 sentences]
- Pros: [list]
- Cons: [list]
- Effort: [low/medium/high]

**Option B: [name]**
- [same structure]

#### Decision: [chosen option]
- Rationale: [why this one wins now]
- Tradeoff accepted: [what we give up]
- Record as ADR: [yes/no — yes if it affects architecture]

**Example**:
- A: Keep sync workflow (low complexity, poor performance under load)
- B: Queue + worker (higher complexity, better scalability)
- Decision: **B** — latency/SLO risk outweighs implementation cost
- Tradeoff: Added operational surface (queue monitoring)

**Skip for**: Single obvious approach, bug fixes, style changes.

</decision-protocol>
