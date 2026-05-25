---
name: storm-plan
description: Spec creation and task ordering. Phase 2 of the storm-scout pipeline. Transforms research findings and clarified requirements into an executable spec and ordered task list.
trigger: /storm-plan
skill_path: skills/storm-plan
---

## Quick Reference

| | |
|---|---|
| **Use when** | Requirements are clear and research is done; ready to create executable spec |
| **Do not use when** | Requirements are ambiguous (use storm-clarify) or research is incomplete (use storm-research) |
| **Primary agent** | turbo-crank |
| **Runtime risk** | Medium — spec ambiguity can propagate to implementation |
| **Outputs** | Executable spec, ordered task list, constitution (if needed) |
| **Deep mode trigger** | `/storm-plan` or explicit plan request |

## Purpose

This is **Phase 2** of the storm-scout pre-implementation pipeline. Transform research findings and clarified requirements into an executable spec and ordered task list. This is the Spec-Driven Development artifact phase.

**Planning is cheap. Implementation is expensive.**

> **Note:** This is a sub-skill of storm-scout. For the full pipeline (Clarify → Research → Plan), use `storm-scout` with `MODE=full`.

---

<!-- LOAD: MODE=plan MODE=full -->
## Phase 2: Plan

Transform research findings and clarified requirements into an executable spec and ordered task list. This is the Spec-Driven Development artifact phase.

**Planning is cheap. Implementation is expensive.**

### Tooling — Use `codebase_search` (WarpGrep)

Use `codebase_search` (WarpGrep) during **Step 2.0: Read Inputs** and **Step 2.2: Specification** to explore existing patterns before writing the spec.

### Step 2.0: Read Inputs
Before writing anything, read:
1. Research findings (from Phase 1)
2. Any existing constitution, spec, or ADR for this area
3. Relevant codebase patterns (what already exists this plan must fit into)
4. **Vault methodology patterns** — query your Obsidian vault for Harness Engineering, Spec-Driven Development, and quality gate patterns:
   ```bash
   qmd query $'lex: "harness engineering" "spec-driven" "quality gate" "design-by-contract"\nvec: how to structure specifications with quality gates and verification checkpoints' -c second-brain -l 5
   ```
   This loads your researched methodologies into context so plans are grounded in your own engineering standards, not generic patterns.

Do not plan in a vacuum. If inputs are missing, note it and surface it.

### Step 2.1: Constitution (for new domains or significant changes)
For new features, subsystems, or significant changes — write a one-page constitution first:

```markdown
## Constitution: [feature/change name]

**Problem:** [What problem does this solve?]
**Solution:** [What are we building at a high level?]
**Non-negotiable constraints:** [What cannot change?]
**Success looks like:** [How will we know it worked?]
**Out of scope:** [What are we explicitly not doing?]
```

For small, well-understood changes: skip the constitution, go directly to spec.

### Step 2.2: Specification
Write the spec. Every spec must have:

```markdown
## Spec: [feature/task name]

**Goal:** [One sentence]

**Requirements:**
1. [Numbered, testable, unambiguous requirement]
2. ...

**Non-goals:**
- [Explicit exclusion — what this will NOT do]

**Acceptance criteria:**
- [ ] [Checkable condition — passes/fails, not subjective]
- [ ] ...

**Dependencies:**
- [What must exist or be true before this can be implemented]
```

If a requirement cannot be made testable: flag it and return to Phase 0 Clarify for that specific point. Do not write untestable requirements.

### Step 2.3: Task Breakdown
Break the spec into ordered, independently executable tasks:

```markdown
## Tasks

### Task 1: [name]
**Done when:** [specific, checkable condition]
**Files likely affected:** [if known]
**Risk:** [low / medium / high — one sentence why]

### Task 2: [name]
...
```

Each task must:
- Have a clear done-condition (not "implement X" — "X returns Y given Z")
- Be small enough to complete in one focused session
- Be ordered by dependency (what must come before what)

Flag tasks with high uncertainty as `[SPIKE]` — a spike is an investigation task, not an implementation task.

### Step 2.4: Approval Gate
Present spec and tasks to human for approval before any implementation begins.

### Output Artifacts
Save to `.specify/`:
- `spec.md` — the specification
- `tasks.md` — the ordered task list

### Plan Rules
- No code in planning phase
- Every requirement must be testable — flag and escalate if not
- Every task must have a done-condition — no open-ended tasks
- If the spec cannot be written without making assumptions: return to Phase 0 Clarify
- Constitution is required for new domains; optional for small changes
- Present spec to human for approval before advancing to implementation
- If scope grows during planning: load the `zero-point` skill (YAGNI pre-flight phase)
- The spec defines the boundary. Anything outside it is out of scope until deliberately added via a new spec revision

<!-- /LOAD -->
