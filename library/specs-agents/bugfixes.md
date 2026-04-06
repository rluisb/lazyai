<rule>
  <scope>auto</scope>
  <globs>specs/bugfixes/**</globs>
  <description>Bugfix workflow — shortened RPI, skip PRD</description>
</rule>

# Bugfix Workflow Rules

## Directory Structure

```
specs/bugfixes/NNN-bug-name/
├── research.md          ← Investigate the bug (Scout)
├── techspec.md          ← Root cause + fix approach (Planner)
├── tasks/               ← Only if fix is >20 lines
│   ├── tasks.md
│   └── 001-name.md
└── progress.md          ← Trace log
```

## Shortened Flow

1. **Research** (Scout) — investigate the bug, map affected code → research.md
   - ⛔ HUMAN GATE: confirm root cause before planning fix
2. **TechSpec** (Planner) — root cause analysis + fix approach → techspec.md
   - **No PRD needed for bugfixes**
   - Pass simplicity gate. Reference existing patterns.
   - ⛔ HUMAN GATE: approve fix approach
3. **Tasks** — only if fix touches >20 lines or >3 files
   - Small fixes: implement directly from techspec
4. **Implement** (Builder) — fix the bug, add regression test
   - ⛔ CHECKPOINT: tests pass, bug is fixed

## Observability

Same as features — every step updates progress.md. Not optional.

## Principles
- Fix the bug, nothing else. No drive-by refactors.
- Always add a regression test.
- If the bug reveals a missing rule → add the rule via PR.

## Self-Improvement — After Every Bugfix

Before ending the session, run the Impact Check from root AGENTS.md.

Additionally for bugfixes:
- Bug caused by missing rule? → Flag specs/rules/ update
- Bug caused by unclear standard? → Flag specs/standards/ improvement
- Regression test uses a new testing pattern? → Flag for specs/standards/testing/ update
- Bug reveals architectural weakness? → Consider ADR + specs/standards/architecture/ update
- Write a memory note to specs/memory/ explaining what went wrong and how to prevent it
