# Agent Contracts

Agents in LazyAI are specialized actors. Each agent must have a clearly defined contract that dictates its scope, inputs, outputs, and handoff behavior.

## Semantic Validation

When `lazyai-cli validate --all` runs, agents are audited for both structural validity and semantic completeness.

### Structural Requirements (Errors)
- Must be a readable file.
- Must contain valid YAML frontmatter.
- Must declare a `name` in frontmatter.
- The body must not be empty.

### Contract Requirements (Warnings)
An agent's markdown body should define its role and workflow completely:
- **Role/Purpose:** What the agent does.
- **Trigger/Misuse:** `When to use` and `When not to use`.
- **Workflow:** The steps the agent takes.
- **Evidence Requirements:** What must be proven before yielding.
- **Human Gates:** When the agent must stop and ask the user (e.g., `⛔ Human gate`).
- **Handoff:** What to do when finished (e.g., pass output to another agent or user).
- **Output:** The format of the final deliverable.

Agents that lack these explicit contract boundaries are likely to drift, hallucinate scope, or fail to interact correctly with the rest of the multi-agent system.