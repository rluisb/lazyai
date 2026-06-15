---
name: storm-research
description: Codebase and domain research. Phase 1 of the storm-scout pipeline. Gathers evidence, maps constraints, and produces structured findings before planning begins.
trigger: /storm-research
skill_path: skills/storm-research
---

## Quick Reference

| | |
|---|---|
| **Use when** | You need to understand the codebase, domain, or existing patterns before planning |
| **Do not use when** | Requirements are unclear (use storm-clarify first) or ready to plan (use storm-plan) |
| **Primary agent** | turbo-crank |
| **Runtime risk** | Low — read-only exploration, no changes |
| **Outputs** | Structured findings document, constraint map, pattern inventory |
| **Deep mode trigger** | `/storm-research` or explicit research request |

## Purpose

This is **Phase 1** of the storm-scout pre-implementation pipeline. Gather, organise, and document what is known before suggesting anything. Research produces a **findings artifact** that feeds forward into planning. Never skip to conclusions.

> **Note:** This is a sub-skill of storm-scout. For the full pipeline (Clarify → Research → Plan), use `storm-scout` with `MODE=full`.

---

<!-- LOAD: MODE=research MODE=full -->
## Phase 1: Research

Gather, organise, and document what is known before suggesting anything. Research produces a **findings artifact** that feeds forward into planning. Never skip to conclusions.

### Tooling — Use `codebase_search` (WarpGrep)

Use `codebase_search` (WarpGrep) as the primary tool for **Step 1.2: Codebase Exploration**. It runs parallel grep + file reads for fast semantic queries.

Best for: "Find the auth flow", "How does X work?", "Where is Y handled?", domain-oriented exploration by concern.
Not for: exact keyword matches — use ripgrep for those, reading known files — use OpenCode `Read`.

### Step 1.1: Scope Definition
Write down what you are researching before you search anything:
- **Primary question:** what must you find out?
- **Secondary questions:** what would be useful to know?
- **Out of scope:** what are you explicitly not researching?

This prevents research from sprawling.

### Step 1.2: Codebase Exploration (if applicable)
Explore by **domain and concern**, not by file or directory name.

For each relevant area:
- What does this part of the system do?
- Where are the entry points?
- What are the key dependencies and constraints?
- What patterns does the codebase already use for this kind of problem?
- Where are the quality gates?

Use `loot-hawk` agent for deep codebase traversal when the scope is broad.

### Step 1.3: Domain Research (if applicable)
For external knowledge (APIs, libraries, framework patterns):
- What do the official docs say?
- What are the known gotchas and version-specific constraints?
- What is the recommended approach and why?

### Step 1.4: Constraint Mapping
Before writing findings, explicitly map constraints:
- What are the **hard constraints** (cannot change, must respect)?
- What are the **soft constraints** (prefer to respect, but could change with justification)?
- What would invalidate the original goal if discovered?

If you find a constraint that invalidates the goal → surface it immediately before writing findings.

### Step 1.5: Findings Document
Produce a structured findings document **before making any suggestions**:

```markdown
## Research Findings

### Goal
[What was researched — one sentence]

### Key Findings
[Bullet points — facts only, no opinions or recommendations yet]

### Constraints
[Hard and soft constraints that will shape the plan]

### Patterns in Use
[How the existing codebase handles similar problems]

### Open Questions
[Anything still unclear — flagged as risk, not assumption]

### Recommended Starting Point
[One sentence: where should planning begin?]
```

### Research Rules
- No implementation suggestions during research phase
- Findings are facts — label opinions clearly if you include any
- If a constraint invalidates the goal: surface it before writing findings, do not bury it
- If research raises more questions than it answers: note them as open questions, don't speculate
- Research without a scope definition is exploration, not research — always define scope first

<!-- /LOAD -->
