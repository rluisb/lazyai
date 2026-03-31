<rule>
  <scope>auto</scope>
  <globs>agents/**</globs>
  <description>Agent orchestration guidance for role-based task execution in the agents directory</description>
</rule>

# Agents Directory

This directory defines specialized agents with clear boundaries. Each agent has a focused purpose, explicit triggers, and handoff expectations so multi-agent flows remain predictable.

## Agent Inventory

| Role | Purpose | Triggers (when invoked) | Key Protocols |
|------|---------|-------------------------|---------------|
| Scout | Gather facts, constraints, and dependencies before action | New task intake, unclear scope, unknown code or docs area | Evidence-first discovery, source citation, ambiguity log |
| Planner | Convert findings into an ordered execution plan | Scout completed, implementation requested, risk identified | Milestone planning, dependency ordering, approval checkpoints |
| Builder | Execute the approved plan with minimal deltas | Plan approved, implementation phase opened | Small-batch changes, test-first where possible, rollback-ready edits |
| Reviewer | Validate correctness, quality, and policy alignment | Builder proposes completion, pre-merge checks needed | Checklist-based review, defect classification, acceptance mapping |
| Red-Team | Challenge assumptions, abuse paths, and failure modes | Security-sensitive change, high-risk logic, external input flow | Adversarial thinking, threat scenarios, mitigation recommendations |
| Documenter | Capture decisions, rationale, and operational learnings | Task closure, major decision made, notable tradeoff accepted | Decision logging, summary clarity, update of reference artifacts |

## Agent Progression Levels

| Level | Operating Mode | Expected Supervision |
|-------|----------------|----------------------|
| Level 1 | Single-task execution only | Human approval for each step |
| Level 2 | Multi-step execution within explicit guardrails | Human approval at major checkpoints |
| Level 3 | Autonomous execution within defined scope | Human notified on exceptions or risk elevation |
| Level 4 | Orchestrates sub-agents across parallel tracks | Human approves objectives and final outcomes |

Progression is earned by consistent quality, reliable handoffs, and policy compliance. If uncertainty rises, drop to a lower level and request guidance.

## Invocation and Handoff Rules

1. Start with **Scout** when requirements, boundaries, or dependencies are incomplete.
2. Move to **Planner** once sufficient facts exist; require explicit acceptance criteria.
3. Activate **Builder** only after plan approval and success checks are defined.
4. Route to **Reviewer** before task completion claims.
5. Invoke **Red-Team** for risky interfaces, trust boundaries, or security-critical workflows.
6. End with **Documenter** to preserve decisions, rationale, and follow-up actions.

### Typical Handoff Flow
- Scout → Planner: findings package with open questions and confidence levels.
- Planner → Builder: ordered steps, constraints, and verification targets.
- Builder → Reviewer: change summary, evidence, and test results.
- Reviewer → Documenter: accepted outcomes, residual risks, and next actions.

## Scope and Escalation Constraints

- Agents must remain within their designated role and avoid lateral scope creep.
- Unknowns, policy conflicts, or missing approvals must be escalated immediately.
- No agent should bypass required reviews or invent missing requirements.
- If a handoff lacks context, return it for clarification rather than guessing.

## Self-Improvement

When this directory evolves:
- Add every new agent to the inventory table with explicit triggers and protocols.
- Update invocation and handoff rules to include new routing paths.
- Reassess progression levels and supervision expectations after process changes.
- Remove outdated role definitions to keep orchestration unambiguous.
