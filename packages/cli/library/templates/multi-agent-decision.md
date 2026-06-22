---
name: multi-agent-decision
description: "Structured template for choosing an agent pattern: single agent + skills, specialist agents, parallel agents, or avoiding multi-agent"
---

# Multi-Agent Decision Template

Use this template when planning a task that could benefit from multiple agents. It guides you through the decision framework and documents the host-tool evidence required.

---

## Task overview

**Task:** <!-- describe the task -->

**Scope estimate:** <!-- small (< 20 lines), medium, large -->

**File-disjoint sub-tasks?** <!-- yes / no — if yes, list them -->

---

## Pattern selection

Use the decision tree from the [Multi-Agent Boundary](../concepts/multi-agent-boundary.md) concept doc:

```
Is the task small (< 20 lines, single file)?
  ├── Yes → Avoid multi-agent (pattern 4)
  └── No → Can sub-tasks run independently?
        ├── No → Single agent + skills (pattern 1)
        └── Yes → Are sub-tasks file-disjoint?
              ├── No → Specialist agents, sequential (pattern 2)
              └── Yes → Does parallelism save more time than merge cost?
                    ├── Yes → Parallel agents (pattern 3)
                    └── No → Specialist agents, sequential (pattern 2)
```

**Selected pattern:** <!-- pattern 1 / 2 / 3 / 4 -->

**Rationale:** <!-- why this pattern wins -->

---

## Host-tool evidence

Verify that the host tool supports the selected pattern:

| Capability | Required? | Supported? | Evidence |
|---|---|---|---|
| Agent definitions from files | Yes / No | Yes / No / Unknown | <!-- link or path --> |
| Named agent invocation | Yes / No | Yes / No / Unknown | <!-- link or path --> |
| Agent-to-agent handoff | Yes / No | Yes / No / Unknown | <!-- link or path --> |
| Parallel agent execution | Yes / No | Yes / No / Unknown | <!-- link or path --> |
| Output merging | Yes / No | Yes / No / Unknown | <!-- link or path --> |

**Fallback pattern if host tool lacks required capability:** <!-- pattern 1 / 4 -->

---

## Agent contracts

For each agent in the selected pattern, define:

### Agent: <!-- name -->

- **Role:** <!-- what this agent does -->
- **Inputs:** <!-- what this agent reads -->
- **Outputs:** <!-- what this agent produces -->
- **Handoff:** <!-- who receives the output and how -->
- **Skills:** <!-- skills this agent uses -->
- **Worktree scope:** <!-- files or directories this agent owns (for parallel agents) -->

---

## Merge protocol

**For sequential patterns (1, 2):** No merge needed — output is a single stream.

**For parallel pattern (3):**

- **Merge owner:** <!-- who reconciles the outputs -->
- **Conflict resolution:** <!-- how file conflicts are resolved -->
- **Consistency check:** <!-- how cross-agent consistency is verified -->
- **Merge cost estimate:** <!-- low / medium / high -->

---

## Verification

- [ ] Pattern matches the decision tree
- [ ] Host-tool evidence is documented and verified
- [ ] Agent contracts are complete
- [ ] Merge protocol is defined (for parallel agents)
- [ ] Fallback pattern is identified if host tool lacks support
