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
5. **Produce Structured Findings**
   - Summarize what was found, what is uncertain, and what to do next.

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
