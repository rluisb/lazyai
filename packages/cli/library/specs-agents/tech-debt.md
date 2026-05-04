<rule>
  <scope>auto</scope>
  <globs>specs/features/**,specs/bugfixes/**,specs/refactors/**,specs/tech-debt/**</globs>
  <description>Unified workflow rules for all work types</description>
</rule>

# Workflow Rules

## Work Types

| Type | Flow | plan.md? | spec.md? | ADR? | Task Threshold |
|------|------|----------|----------|------|---------------|
| **Feature** | Full RPI | Yes | Optional (complex) | If applicable | Always |
| **Bugfix** | Shortened | Yes (brief) | No | Rarely | >20 lines only |
| **Refactor** | Full RPI | Yes | Yes (always) | **Always** | Always |
| **Tech Debt** | Shortened | Yes (brief) | No | If applicable | >20 lines only |

## Directory Structure

All work items follow: `specs/{type}/NNN-name/`

```
specs/{type}/NNN-name/
├── research.md              ← R phase (Scout)
├── plan.md                  ← P phase (Planner) — replaces prd + techspec
├── spec.md                  ← Detailed spec (optional, complex features/refactors)
├── checklists/
│   └── requirements.md      ← Verification criteria
├── quickstart.md            ← Team onboarding (optional)
└── tasks/
    ├── 001-name.md
    └── 002-name.md
```

## RPI Flow (all types)

1. **Research** (Scout) — map what exists
   - ⛔ HUMAN GATE: approve before planning
2. **Plan** (Planner) — define approach
   - Bugfixes/tech-debt: shortened plan (root cause + fix)
   - Features/refactors: full plan with phases
   - ⛔ HUMAN GATE: approve before implementing
3. **Implement** (Builder) — one task per session
   - Follow task file. Run tests. Update progress.
   - ⛔ CHECKPOINT after each task

## Type-Specific Rules

### Features
- YAGNI: build only what's needed for the current phase
- MVP first: complete and validate before starting the next phase

### Bugfixes
- Fix the bug, nothing else. No drive-by refactors.
- Always add a regression test.

### Refactors
- ADR is MANDATORY — capture why the architecture changed
- Must be phased. No big-bang rewrites.
- Each phase leaves the codebase working.

### Tech Debt
- Prioritize by risk impact, not annoyance.
- Always add tests that would have caught the debt earlier.

## Self-Improvement — After Every Step

Run the Impact Check from root AGENTS.md. Additionally:
- New pattern? → Flag for specs/standards/ update
- Bug revealed missing rule? → Flag for specs/rules/ update
- ADR created? → Verify linked in KNOWLEDGE_MAP.md
- Feature completed? → Update KNOWLEDGE_MAP.md status
