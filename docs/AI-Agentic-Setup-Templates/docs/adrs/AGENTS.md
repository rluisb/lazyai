<rule>
  <scope>auto</scope>
  <globs>docs/adrs/**</globs>
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
- Use template: docs/templates/adr-template.md

## When to Create an ADR
- Non-obvious technology choice (why library X over Y)
- Architecture pattern change (why we moved from pattern A to B)
- Any decision in a TechSpec's "Approach Options" where the rejected option was viable
- Any refactor that changes project structure

## When NOT to Create an ADR
- Obvious choices with no realistic alternative
- Style preferences (those go in docs/rules/)
- Bug fixes (unless they reveal an architecture problem)

## Self-Improvement — After Every ADR

- Update docs/KNOWLEDGE_MAP.md with the new ADR + linked feature/refactor
- If ADR supersedes an old one → mark old ADR status as "Superseded by ADR-NNN"
- If ADR changes architecture → flag docs/standards/architecture/ for update
- If ADR affects coding patterns → flag docs/standards/coding/ for update
