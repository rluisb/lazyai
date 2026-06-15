---
name: <fortnite-agent-name>
model: <primary-model>
skill: <primary-skill>
think: true
permission.task: allow
---

# <Agent Display Name>

## Role
<One sentence: what this agent does in the squad>

## Parameter Contract

| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `<value1>` / `<value2>` | `<default>` | What mode controls |
| THINK | `true` / `false` / `xhigh` | `true` | Reasoning depth |
| TOKEN_BUDGET | number | `40K` | Max context tokens |

## Fallback Chain

`<primary-model>` → `<fallback-1>` → `<fallback-2>` → escalate to loop-driver

## Capabilities

**Can do:**
- Capability 1
- Capability 2
- Dispatch to other agents (per dispatch matrix)

**Cannot do:**
- Restriction 1
- Restriction 2

## Delegation Rules

- Can dispatch to: `<agent-list>`
- Requires human approval for: `<actions>`
- Never delegates to: `<agents>` without loop-driver mediation

## Integration

- **Primary skill**: `<skill-name>` — loaded automatically
- **Secondary skills**: `<skill-list>` — loaded as needed
- **CLI tools**: `<tool-list>` — available for this agent
