---
name: storm-scout
description: Pre-implementation specification pipeline. Clarify → Research → Plan in one skill. Grill Me first, gather evidence, then produce an executable spec and task list. Storm-scout ahead, map the safe zone, and plan the drop before the bus leaves.
trigger: /storm-scout
skill_path: skills/storm-scout
scripts:
  - name: task-init.sh
    description: Scaffold .specify/<slug>/ directory structure for new tasks
    path: scripts/task-init.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | pre-implementation clarification, research, planning |
| **Do not use when** | implementation, review, deterministic execution |
| **Primary agent** | turbo-crank |
| **Runtime risk** | Medium — spec ambiguity resolution |
| **Outputs** | Executable spec, task list, research findings |
| **Validation** | Grill Me depth, evidence links |
| **Deep mode trigger** | `/storm-scout` or MODE=full |

# Storm-Scout


## Tool Selection

Use the right tool for each job. See skills/_tool-hierarchy.md for full decision tree.

| Task | Tool |
|------|------|
| Read known file | OpenCode Read |
| Find code by description | morph codebase_search |
| Symbol analysis | codegraph MCP |
| Vault search | qmd MCP |
| Architecture overview | graphify CLI |


## Purpose
Scout ahead of the storm. You are the pre-implementation specification pipeline — three phases in one skill:

1. **Clarify** — Grill Me interrogation. Resolve unknowns before anything else.
2. **Research** — Gather evidence. Map the codebase and domain.
3. **Plan** — Produce an executable spec and ordered task list.

**Feed Forward principle:** ambiguity resolved here costs minutes. The same ambiguity resolved during implementation costs hours (or a rewrite).

## Sub-Skills

storm-scout is the compatibility wrapper. For focused work, load sub-skills directly:

| Mode | Sub-Skill | File |
|------|-----------|------|
| clarify | storm-clarify | `skills/storm-clarify/SKILL.md` |
| research | storm-research | `skills/storm-research/SKILL.md` |
| plan | storm-plan | `skills/storm-plan/SKILL.md` |
| full | storm-scout (this file) | All phases loaded |

> **Backward compatibility:** storm-scout remains the primary entry point. Existing triggers (`/storm-scout`, `MODE=clarify|research|plan|full`) continue to work exactly as before.

## Scripts

This skill owns the following scripts:

| Script | Purpose |
|--------|---------|
| `task-init.sh` | Scaffold `.specify/<slug>/` directory with spec.md, tasks.md templates |

Run from skill directory: `./scripts/task-init.sh <slug>`

## Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| MAX_QUESTIONS | 3 | Maximum clarifying questions per pass |

Override at invocation: include `MAX_QUESTIONS=N` in your prompt.
- Standard scope: `MAX_QUESTIONS=3`
- High ambiguity / architectural decisions: `MAX_QUESTIONS=5`

---

## Phase Loading Guide

Load phases conditionally based on MODE. The Purpose and Scripts sections above are always loaded.

| MODE | Load Phase 0 | Load Phase 1 | Load Phase 2 |
|------|-------------|-------------|-------------|
| clarify | ✅ | ❌ | ❌ |
| research | ❌ | ✅ | ❌ |
| plan | ❌ | ❌ | ✅ |
| full | ✅ | ✅ | ✅ |

---

<!-- LOAD: MODE=clarify MODE=full -->
## Phase 0: Clarify

You are a clarification agent. Your job is to identify and resolve critical unknowns **before** any research, planning, or implementation begins. Ambiguity is cheap to fix here. It is expensive everywhere else.

This is the Grill Me pattern: structured interrogation before action.

### Step 0.0: Load Existing Knowledge (Grill Me With Docs)
Before interrogating the human, read what already exists. Don't ask questions that are already answered.

1. Check `.specify/memory/*.md` — any relevant lessons, constraints, or gotchas?
2. Check `bee-gone/specs/` — any existing specs, ADRs, or research for this domain?
3. Search the vault with `qmd` — use hybrid lex+vec search for existing knowledge:
   ```bash
   qmd query $'lex: domain-keywords\nvec: what are the key decisions and tradeoffs for {topic}' -c second-brain -l 5
   ```

4. Check the codebase itself — any patterns or conventions that answer your questions?

**Self-filter your question list**: for each candidate question, ask "Is this answerable from existing docs/code/specs/vault?" If yes → read it, don't ask it. If no → keep it in your question list.

This prevents the most expensive kind of clarification question: the one whose answer already exists on disk.

### Step 0.1: Analyse
Read the request carefully. List ALL ambiguities, missing context, and unstated assumptions. Do not proceed past this step until you have an exhaustive list.

### Step 0.2: Prioritise
Rank unknowns by impact. Ask yourself: which of these, if left unresolved, would cause me to build the wrong thing entirely?

Select the **top `MAX_QUESTIONS`** highest-impact unknowns only. Ignore the rest for now — you can ask in a follow-up pass if needed.

### Step 0.3: Ask
Present questions as a numbered list. Each question must be:
- **Specific** — not open-ended ("which X?" not "tell me about X")
- **Bounded** — answerable in 1–3 sentences
- **Actionable** — the answer must change what you do next

Do NOT ask about things you can determine from context.
Do NOT combine multiple questions into one.

### Step 0.4: Wait
Stop completely. Do not proceed. Do not guess. Do not assume "probably X."
Wait for the human's answers.

### Step 0.5: Confirm Understanding
After receiving answers, produce a **Confirmed Understanding** block:

```
## Confirmed Understanding

**Goal:** [what we're doing — one sentence]
**Constraints:** [what we cannot or must not do]
**Scope:** [what's in, and explicitly what's out]
**Open risks:** [anything still unclear, flagged as a risk not an assumption]
```

### Step 0.6: Gate
Present the Confirmed Understanding block and ask: "Does this accurately capture the intent?"

Only after the human confirms: proceed to Phase 1.

### Clarify Rules
- Never ask more than `MAX_QUESTIONS` per pass
- Never proceed with an unconfirmed assumption — turn it into a flagged risk instead
- If the human's answers introduce new ambiguities, one targeted follow-up is allowed per pass
- The Confirmed Understanding block is mandatory — skipping it invalidates this entire skill
- If scope is too large to clarify in MAX_QUESTIONS, say so and ask which part to start with

<!-- /LOAD -->

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
   qmd query $'lex: "harness engineering" "spec-driven" "quality gate" "design-by-contract"
vec: how to structure specifications with quality gates and verification checkpoints' -c second-brain -l 5
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

---

## Integration with Other Skills

- **sidecar**: On session start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs before Phase 0 (Grill Me). If no sidecar config exists or the query returns no specs, proceed normally without failing.
