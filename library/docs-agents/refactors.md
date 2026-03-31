<rule>
  <scope>auto</scope>
  <globs>docs/refactors/**</globs>
  <description>Refactor workflow — full RPI flow, ADR mandatory</description>
</rule>

# Refactor Workflow Rules

## Directory Structure

```
docs/refactors/NNN-refactor-name/
├── research.md          ← Map current state (Scout)
├── prd.md               ← Why refactor, goals, scope (Planner)
├── techspec.md          ← New design + migration path (Planner)
├── tasks/
│   ├── tasks.md
│   └── NNN-name.md
└── progress.md          ← Trace log
```

## Full Flow (same as features, plus ADR)

1. **Research** — map current state of code being refactored
2. **PRD** — why this refactor, what's the goal, what's the scope
3. **TechSpec** — new design, migration path, rollback plan
   - **ADR is MANDATORY** — capture the architectural decision in docs/adrs/
4. **Tasks** — phased approach. Never big-bang refactors.
5. **Implement** — one phase at a time. Tests pass after each phase.

## Principles
- Refactors MUST be phased. No "rewrite everything in one go."
- Each phase must leave the codebase in a working state.
- ADR captures WHY we changed the architecture.
- If the refactor introduces a new pattern → update docs/standards/.
- Keep existing tests passing at every step. Add new tests for new patterns.

## Decision-Making Protocol (Required for structural refactors)

Before committing to a refactor direction:

1. Generate **2+ viable alternatives**.
2. Evaluate each option using:
   - complexity
   - consistency with current code patterns
   - reversibility / rollback safety
   - performance impact
   - team familiarity
3. Choose one path and state why it is preferred now.
4. Record tradeoffs and what risk is accepted by rejecting other options.

If the decision changes architecture or boundaries, capture it in an ADR.

## Self-Improvement — After Every Refactor

Before ending the session, run the Impact Check from root AGENTS.md.

Refactors are the HIGHEST impact on project knowledge. Additionally:
- Module structure changed? → Update root AGENTS.md codebase map immediately
- Code patterns changed? → Update ALL affected docs/standards/ files
- Old pattern replaced? → Update standards to show new pattern, mark old as deprecated
- ADR is MANDATORY → verify it exists and is linked in KNOWLEDGE_MAP.md
- Cross-module boundaries changed? → Update docs/standards/architecture/
