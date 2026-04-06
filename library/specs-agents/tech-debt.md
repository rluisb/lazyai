<rule>
  <scope>auto</scope>
  <globs>specs/tech-debt/**</globs>
  <description>Tech debt workflow — planned risk reduction, no PRD needed</description>
</rule>

# Tech Debt Workflow Rules

## Directory Structure

```
specs/tech-debt/NNN-debt-name/
├── research.md          ← Assess the debt (Scout)
├── techspec.md          ← Fix approach using tech-debt-template (Planner)
├── tasks/               ← Only if fix is >20 lines
│   ├── tasks.md
│   └── NNN-name.md
└── progress.md          ← Trace log
```

## Flow

1. **Research** (Scout) — assess the debt, map affected code, measure impact → research.md
   - ⛔ HUMAN GATE: confirm priority and scope
2. **TechSpec** (Planner) — fix approach using tech-debt-template → techspec.md
   - **No PRD needed** — tech debt is technical, not a product requirement
   - Pass simplicity gate. Prefer incremental fix.
   - ADR if the fix changes architecture.
   - ⛔ HUMAN GATE: approve fix approach
3. **Tasks** — if fix is >20 lines or >3 files
4. **Implement** (Builder) — fix incrementally, add tests

## How Tech Debt Differs

| | Feature | Bugfix | Refactor | Tech Debt |
|---|---|---|---|---|
| **Urgency** | Planned | Urgent | Planned | Planned |
| **PRD needed** | Yes | No | Yes | No |
| **ADR needed** | If applicable | Rarely | Always | If applicable |
| **Goal** | Add value | Fix broken | Change structure | Reduce risk |
| **Approach** | Full RPI | Shortened | Full RPI + ADR | Shortened |

## Principles
- Tech debt is about RISK, not annoyance. Prioritize by impact.
- Prefer incremental fixes. Never big-bang.
- Always add tests that would have caught the debt earlier.
- If debt reveals a missing rule → add the rule via PR.
- If debt reveals a missing standard → document the pattern after fixing.

## Self-Improvement — After Every Tech Debt Resolution

Before ending the session, run the Impact Check from root AGENTS.md.

Additionally for tech debt:
- Debt caused by missing standard? → Create the standard now
- Debt caused by missing rule? → Add the rule now
- Dependency upgraded? → Update root AGENTS.md stack + any affected standards
- Fix introduced a better pattern? → Update specs/standards/ to reflect it
- Write a memory note to specs/memory/ explaining what caused the debt and how to prevent recurrence
