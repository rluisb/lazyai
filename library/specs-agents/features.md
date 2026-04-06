<rule>
  <scope>auto</scope>
  <globs>specs/features/**</globs>
  <description>Feature workflow — RPI flow with PRD, TechSpec, Tasks, and observability</description>
</rule>

# Feature Workflow Rules

## Directory Structure

Each feature gets a numbered directory:
```
specs/features/NNN-feature-name/
├── research.md          ← R phase (Scout)
├── prd.md               ← P step 1 (Planner)
├── techspec.md          ← P step 2 (Planner)
├── tasks/
│   ├── tasks.md         ← P step 3 (Planner)
│   ├── 001-name.md      ← P step 4 (Planner)
│   ├── 002-name.md
│   └── 003-name.md
└── progress.md          ← Trace log (all agents)
```

## RPI Flow

1. **Research** (Scout) — map what exists → research.md
   - ⛔ HUMAN GATE: approve before planning
2. **PRD** (Planner) — define WHAT/WHY → prd.md
   - Ask clarifying questions FIRST. Never skip.
   - ⛔ HUMAN GATE: approve before techspec
3. **TechSpec** (Planner) — define HOW → techspec.md
   - Pass simplicity gate. Explore ≥2 approaches. Reference standards.
   - Create ADR if non-obvious decision made.
   - ⛔ HUMAN GATE: approve before tasks
4. **Tasks** (Planner) — ordered breakdown → tasks/tasks.md
   - Show HIGH-LEVEL list FIRST → human approves → then generate task files
   - ⛔ HUMAN GATE: approve task list
5. **Task Breakdown** (Planner) — individual files → tasks/NNN-*.md
   - ⛔ HUMAN GATE: review task files
6. **Implement** (Builder) — one task per session → code
   - Follow task file. Check boxes. Run tests. Update progress.md.
   - ⛔ CHECKPOINT after each task

## Observability — MANDATORY

After completing ANY step, the agent MUST append to progress.md:

```
### [YYYY-MM-DD HH:MM] — [Step] ([Agent])
- Agent: [name]
- Session: new | continued
- Context loaded: [files]
- Files read: [count, paths]
- Files changed: [paths — or "N/A"]
- Output: [artifact]
- Decisions: [choices — or "None"]
- Status: ✅ Complete | ⏳ In Progress | 🚫 Blocked
```

This is NOT optional. Every step updates progress.md.

## Principles
- YAGNI: build only what P1 needs
- Simplest thing that works
- Respect existing patterns (check specs/standards/)
- One task = one session = clean context
- MVP first: P1 complete and validated before P2 starts

## Self-Improvement — After Every Step

Before ending any session within a feature, run the Impact Check from root AGENTS.md.

Additionally for features:
- New pattern introduced? → Flag for specs/standards/ update
- ADR created during TechSpec? → Verify it's linked in progress.md AND KNOWLEDGE_MAP.md
- Feature completed? → Update KNOWLEDGE_MAP.md status
- New test type created? → Check if specs/standards/testing/ needs a new standard
- New module created? → Update root AGENTS.md codebase map
