<agent-harness>

### Agent Coordination

When using multiple specialized agents, follow these coordination rules:

**Agent roles**:
| Agent | When to use | What it does | What it does NOT do |
|-------|-------------|-------------|---------------------|
| Scout | Research phase | Maps codebase, identifies patterns | Does not suggest, plan, or write code |
| Planner | Plan phase | Creates plans, asks clarifying questions | Does not implement code |
| Builder | Implement phase | Executes plan, writes code and tests | Does not add unrequested features |
| Reviewer | After implementation | Finds issues, rates severity | Does not fix code |
| Documenter | After completion | Writes docs, updates standards | Does not modify source code |

**Handoff protocol**:
- Each agent reads the previous agent's output before starting.
- Each agent writes its output to the designated location (specs/ or code).
- If an agent is blocked, it STOPS and describes the blocker — does not guess.

**Escalation**: When confidence is low, return to the human with:
1. What you know.
2. What you are uncertain about.
3. What you need to proceed.

</agent-harness>
