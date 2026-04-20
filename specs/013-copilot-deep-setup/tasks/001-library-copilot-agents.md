# Task 001 — `library/copilot/agents/*.agent.yaml` starter set

**Phase:** 1 (library content)
**Estimated LOC:** ~160 (mostly content)

## Goal

Author ai-setup-flavored custom agents as `.agent.yaml` files using the zod-confirmed schema from research §2.2. These files feed both Copilot surfaces: VS Code reads them from `.github/agents/`, the standalone CLI reads them from both `.github/agents/` and `~/.copilot/agents/`.

## Files to create

| File | Purpose |
|---|---|
| `library/copilot/agents/planner.agent.yaml` | Plans approaches, breaks down tasks (port `library/agents/planner.md` prompt) |
| `library/copilot/agents/builder.agent.yaml` | Executes task files, writes code + tests |
| `library/copilot/agents/scout.agent.yaml` | Reads-only research / codebase mapping |
| `library/copilot/agents/reviewer.agent.yaml` | Code review, standards conformance |
| `library/copilot/agents/orchestrator.agent.yaml` | Multi-step chain runner (wired in later phase, gated on `EnableServers`) |

## Schema (from research §2.2)

```yaml
name: <lowercase-id>           # required
displayName: <Title Case>
description: >
  <1-3 sentences>
model: claude-sonnet-4.5       # optional; pick from --help choices
tools:
  - "*"                        # or explicit list
promptParts:
  includeAISafety: true
  includeToolInstructions: true
  includeParallelToolCalling: true
  includeCustomAgentInstructions: false
prompt: |
  <system prompt body>
```

## Acceptance criteria

- [ ] Five YAML files exist under `library/copilot/agents/`
- [ ] Each has `name`, `description`, `prompt` at minimum (other keys optional)
- [ ] Prompts are ports of existing `library/agents/*.md` content, adapted to Copilot's single-prompt format
- [ ] `tools: ["*"]` unless the role is explicitly read-only (scout uses a narrower list)
- [ ] YAML parses cleanly (validated in task 003)

## Test plan

Task 003 covers the parse + schema unit test. No new tests in this task.

## Notes

- Match `displayName` against our existing naming (`Planner`, `Builder`, etc.).
- Orchestrator file exists from this task but is only wired at install time when `EnableServers` includes it — adapter wiring lives in task 004.
