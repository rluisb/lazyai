<template>
  <name>Progress — Feature Trace Log</name>
  <output>docs/features/NNN-feature-name/progress.md</output>
  <input>Auto-populated by agents after each step</input>
  <phase>All phases — living document</phase>
</template>

# Progress: [Feature Name]

**Feature:** NNN-feature-name
**Started:** YYYY-MM-DD
**Current phase:** Research | Plan | Implement | Review | Done

---

## Session Log

<!-- Every agent appends an entry after completing its step.
     This is the audit trail. Never delete entries. Only append.
     This file IS your observability for AI-assisted work. -->

### [YYYY-MM-DD HH:MM] — [Step Name] ([Agent Name])
- **Agent:** [scout | planner | builder | reviewer | red-team | documenter]
- **Session:** new | continued
- **Context loaded:** [files read at session start]
- **Files read:** [count, key paths]
- **Files changed:** [paths — implementation only, or "N/A"]
- **Output:** [artifact produced — e.g. research.md, prd.md]
- **Decisions:** [choices made — or "None"]
- **Blockers:** [if any — or "None"]
- **Status:** ✅ Complete | ⏳ In Progress | 🚫 Blocked

---

## ADRs Created

<!-- Track architectural decisions produced during this work.
     Bidirectional: Feature → ADR (here). ADR → Feature (in the ADR's "Feature" field). -->

- [docs/adrs/NNN-title.md — created during TechSpec — or "None"]

---

## Current State

- **Phase:** [Research | PRD | TechSpec | Tasks | Implementing | Review | Done]
- **Task progress:** [N/M complete]
- **Tests:** [N passing, N failing]
- **Next step:** [what comes next]
- **Blockers:** [active blockers — or "None"]
