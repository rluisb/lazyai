# Subagent Capability Blueprint

When orchestrating subagents for LazyAI canonical roles, apply the following
`define_subagent` enable flags:

| Role | enable_write_tools | enable_mcp_tools | enable_subagent_tools |
|---|---|---|---|
| researcher | false | false | false |
| reviewer | false | false | false |
| evidence-verifier | false | false | false |
| implementer | true | true | false |
| deployer | true | true | false |
| planner | true | true | true |
| guide | true | true | true |
| responder | true | true | true |

## Rationale

Read-only roles (`researcher`, `reviewer`, `evidence-verifier`) declare
`tools: [read, search]` in LazyAI canonical frontmatter (#569). They must not
be granted `enable_write_tools` — omitting it from `define_subagent` enforces
that constraint at the Antigravity capability layer.

Write-capable roles (`implementer`, `deployer`) require `enable_write_tools`
and `enable_mcp_tools` to access file-mutation and MCP server tools.

Orchestrator roles (`planner`, `guide`, `responder`) additionally need
`enable_subagent_tools` to spawn and coordinate sub-agents.

## Permissions Engine

`read_url(*)` and `command(*)` permissions follow the Antigravity permissions
engine defaults. Restrict further via `.agents/rules/lazyai.md` policy blocks
as needed. The `lazyai-write-guard` hook (`hooks.json`) provides a global
safety net for write-tier tool calls.

## Example

```python
# Spawn a read-only researcher subagent
agent.define_subagent(
    name="researcher",
    instructions=agent.load_skill("researcher"),
    enable_write_tools=False,
    enable_mcp_tools=False,
    enable_subagent_tools=False,
)

# Spawn a write-capable implementer subagent
agent.define_subagent(
    name="implementer",
    instructions=agent.load_skill("implementer"),
    enable_write_tools=True,
    enable_mcp_tools=True,
    enable_subagent_tools=False,
)
```
