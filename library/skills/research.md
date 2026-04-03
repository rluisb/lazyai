---
name: research
description: Evidence-based codebase investigation.
trigger: /research
phase: research
---

# Research Skill

**Command:** `/research [topic]`
**Goal:** Produce a structured research document grounded in code and docs evidence.

---

## Workflow

1. **Clarify Scope**
   - Define the topic boundaries, goals, and constraints.
   - Capture assumptions and open questions up front.
2. **Search the Codebase**
   - Locate related modules, entry points, and interfaces.
   - Trace relevant flows and dependencies.
3. **Read Relevant Docs**
   - Review architecture notes, standards, ADRs, and related specs.
   - Cross-check docs against implementation reality.
4. **Identify Patterns and Gaps**
   - Document established patterns and reusable approaches.
   - Highlight inconsistencies, missing coverage, and risks.
5. **Generate Domain Knowledge**
   - Capture key patterns, dependencies, and constraints supported by evidence.
   - Discard any generated claims not grounded in code or docs.
6. **Produce Structured Findings**
   - Summarize what was found, what is uncertain, and what to do next.

## Trace Protocol (ReAct, complex investigations only)

When research is multi-step or ambiguous, keep a compact trace:

1. **Thought:** next evidence needed
2. **Action:** search/read step taken
3. **Observation:** what evidence was found
4. **Decision:** continue, pivot, or escalate question

Skip this for straightforward lookups.

## Output Format

```markdown
## Research: [Topic]

### Scope
- Objective: [what this research covers]
- Out of scope: [what is intentionally excluded]

### Sources Reviewed
- Code: [path or module]
- Docs: [doc title/path]

### Findings
- [finding with evidence]

### Risks
- [risk] — [impact]

### Recommendations
- [recommended next step]

### Open Questions
- [question needing human decision]
```

## Integration
- **Primary agent:** Scout (research phase)
- **Triggered by:** `/research` command or workflow rule step 1
- **Depends on:** Task description or PRD from user
- **Feeds into:** plan.md (Planner reads research output)
- **Related skills:** anti-speculation (prevents scope creep during research)
