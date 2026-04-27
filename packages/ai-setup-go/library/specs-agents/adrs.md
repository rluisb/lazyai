<rule>
  <scope>auto</scope>
  <globs>specs/adrs/**</globs>
  <description>ADRs are permanent architectural decisions. Never edit accepted ADRs.</description>
</rule>

# ADR Rules

## What Goes Here
Architecture Decision Records — short documents that capture WHY a technical decision was made.

## Rules
- ADRs are **permanent**. Never edit an accepted ADR.
- To change a decision: create a NEW ADR that supersedes the old one.
- Mark the old ADR: `Status: Superseded by ADR-NNN`
- One page max. If it's longer, you're over-explaining.
- Number sequentially across the entire project: 001, 002, 003...
- Use template: specs/templates/adr-template.md

## Decision-Making Protocol (Required)

Before finalizing any ADR decision:

1. Generate **2+ viable alternatives**.
2. Evaluate each alternative using these criteria:
   - complexity
   - consistency with existing patterns
   - reversibility
   - performance
   - team familiarity
3. Select one path and explain why it is best *now*.
4. Record key tradeoffs and the risk of rejected options.

Use concise scoring/notes if helpful, but always keep rationale explicit.

## When to Create an ADR
- Non-obvious technology choice (why library X over Y)
- Architecture pattern change (why we moved from pattern A to B)
- Any decision in a TechSpec's "Approach Options" where the rejected option was viable
- Any refactor that changes project structure

## When NOT to Create an ADR
- Obvious choices with no realistic alternative
- Style preferences (those go in specs/rules/)
- Bug fixes (unless they reveal an architecture problem)

## Self-Improvement — After Every ADR

- Update specs/KNOWLEDGE_MAP.md with the new ADR + linked feature/refactor
- If ADR supersedes an old one → mark old ADR status as "Superseded by ADR-NNN"
- If ADR changes architecture → flag specs/standards/architecture/ for update
- If ADR affects coding patterns → flag specs/standards/coding/ for update
