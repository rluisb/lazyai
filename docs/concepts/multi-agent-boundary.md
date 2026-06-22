# Multi-Agent Boundary

LazyAI compiles guidance assets — agent definitions, skills, prompts, templates, and fragments — that describe how agents should behave. It does **not** schedule, orchestrate, or run agents at runtime. The host tool (Claude Code, OpenCode, Copilot, Pi, OMP, Kiro, Antigravity) owns agent lifecycle: invocation, context management, parallel execution, and output merging.

This page defines the boundary between what LazyAI provides (compiled guidance) and what the host tool owns (agent runtime). It also provides a decision framework for choosing the right agent pattern.

---

## What LazyAI provides

LazyAI's `compile` command reads canonical `.ai/` sources and writes tool-native files. For agents, this means:

| Asset | What it is | How it compiles |
|---|---|---|
| Agent definitions | Role, trigger, workflow, contract | Written as tool-native agent files (e.g., `.opencode/agents/*.md`, `.claude/agents/*.md`) |
| Skills | Reusable capability modules | Written to the tool's skills directory |
| Fragments | Reusable XML/markdown guidance | Embedded inline or referenced by agent definitions |
| Templates | Structured decision frameworks | Written as tool-native templates or standalone files |
| Prompts | Context-injection snippets | Written to the tool's prompts directory |

The host tool reads these files and decides when and how to invoke agents. LazyAI never spawns a process, opens a socket, or manages a runtime schedule.

---

## Agent patterns

There are four patterns for structuring agent work. Each has different merge costs, context implications, and host-tool evidence requirements.

### 1. Single agent + skills

One agent with a broad role that delegates sub-tasks to skills.

**When to use:**
- The task is linear and fits one context window.
- Sub-tasks are well-defined and independent enough to be skills.
- Merge cost of parallel agents would exceed the time saved.

**Host-tool evidence:**
- The agent definition references skills by name.
- Skills are registered in the tool's skills directory.
- The agent's workflow includes skill-invocation steps.

**Merge cost:** None — single output stream, no conflict resolution.

**Example:** A `guide` agent that calls `search-company-knowledge`, `plan`, and `implement` skills sequentially.

### 2. Specialist agents

Multiple agents, each with a narrow, well-defined role. They run sequentially or in a handoff chain.

**When to use:**
- The task has distinct phases (research, plan, implement, review).
- Each phase benefits from a focused context window.
- The output of one phase is the input of the next.

**Host-tool evidence:**
- Each agent has a contract defining inputs, outputs, and handoff.
- Handoff protocol is documented in agent definitions or fragments.
- The host tool supports agent-to-agent handoff (varies by tool).

**Merge cost:** Low — sequential handoff means one active agent at a time. The merge is the handoff document.

**Example:** `researcher` → `planner` → `implementer` → `reviewer`, each reading the previous agent's output.

### 3. Parallel agents

Multiple agents working on independent sub-tasks simultaneously.

**When to use:**
- Sub-tasks are file-disjoint and truly independent.
- The time savings from parallelism outweigh the merge cost.
- The host tool supports parallel agent execution.

**Host-tool evidence:**
- The host tool's parallel execution mechanism is documented and tested.
- Each parallel agent has a scoped worktree or file assignment.
- A merge protocol exists for combining outputs.

**Merge cost:** High — requires conflict resolution, context merging, and consistency checks. Each parallel agent produces a separate output stream that must be reconciled.

**Example:** Three agents fixing unrelated bugs in different packages, each in its own worktree.

### 4. Avoiding multi-agent

Using a single agent without skills, or using the host tool's built-in capabilities directly.

**When to use:**
- The task is small (single file, few lines).
- The overhead of defining agents and skills exceeds the task value.
- The host tool's default behavior is sufficient.

**Host-tool evidence:**
- No agent definitions beyond the default.
- No custom skills registered.
- The task is completed within a single session.

**Merge cost:** None.

---

## Decision framework

Use this decision tree when choosing an agent pattern:

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

### Cost estimation

| Pattern | Context overhead | Merge cost | Host-tool dependency |
|---|---|---|---|
| Single agent + skills | Low | None | Low — skills are well-supported |
| Specialist agents | Medium | Low (handoff) | Medium — handoff varies by tool |
| Parallel agents | High | High (conflict resolution) | High — parallel execution varies |
| Avoiding multi-agent | None | None | None |

---

## Host-tool evidence requirements

Before adopting a multi-agent pattern, verify that the host tool supports it:
- Template: `packages/cli/library/templates/multi-agent-decision.md` — structured template for choosing an agent pattern
1. **Agent definitions** — Does the tool read agent files from a known directory?
2. **Agent invocation** — Can the tool invoke a named agent on demand?
3. **Handoff** — Can one agent pass context to another?
4. **Parallel execution** — Can the tool run multiple agents concurrently?
5. **Output merging** — Does the tool have a mechanism for combining parallel outputs?

Document the evidence in the agent contract or a fragment. If the host tool does not support a required capability, fall back to a simpler pattern.

---

## LazyAI's role: compiler, not scheduler

LazyAI compiles agent definitions, skills, fragments, and templates into the host tool's native format. It does not:

- Schedule agent execution
- Manage agent lifecycles
- Resolve parallel output conflicts
- Provide a runtime for agent-to-agent communication
- Orchestrate multi-step workflows at runtime

These responsibilities belong to the host tool. LazyAI's output is static guidance — files the host tool reads and interprets.

This boundary is intentional. By staying out of the runtime, LazyAI remains compatible with any host tool that reads files from standard locations. Adding runtime orchestration would couple LazyAI to specific host tools and create a maintenance burden that outweighs the benefit.

---

## Related

- [Agent Contracts](agent-contracts.md) — defining agent scope, inputs, outputs, and handoff
- [Product Boundaries](product-boundaries.md) — what LazyAI ships vs what it does not
- [Harness Principles](harness-principles.md) — the engineering principles behind agent guidance
- Template: `packages/cli/library/templates/multi-agent-decision.md` — structured template for choosing an agent pattern
