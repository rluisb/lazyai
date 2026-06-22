<specialist-agent-guidance>

### Specialist Agent Guidance

When using multiple specialized agents, define clear boundaries to prevent scope drift, context bloat, and merge conflicts.

#### When to use specialist agents

- The task has distinct, sequential phases (research, plan, implement, review).
- Each phase benefits from a focused context window without cross-phase noise.
- The output of one phase is a well-defined input for the next.
- A single agent would exceed the host tool's context window.

#### When NOT to use specialist agents

- The task is small enough for one agent (< 20 lines, single file).
- Sub-tasks are tightly coupled and cannot be cleanly separated.
- The host tool does not support agent-to-agent handoff.
- The overhead of defining agent contracts exceeds the task value.

#### Agent contract requirements

Each specialist agent MUST define:

| Field | Purpose |
|---|---|
| `Role` | What the agent does — one sentence |
| `When to use` | Trigger conditions for routing to this agent |
| `When not to use` | Conditions that should route elsewhere |
| `Inputs` | What the agent reads (files, previous agent output, environment) |
| `Outputs` | What the agent produces (files, handoff document, report) |
| `Handoff` | Who receives the output and how |
| `Skills` | Skills the agent uses |
| `Evidence` | What must be proven before yielding |

#### Handoff protocol

1. The producing agent writes its output to a known location (file, handoff document, or tool-native handoff mechanism).
2. The producing agent signals completion (exit code, handoff file, or tool notification).
3. The consuming agent reads the output before starting.
4. The consuming agent verifies the output is complete and consistent.
5. If the output is incomplete or inconsistent, the consuming agent escalates — it does not guess.

#### Merge cost

Specialist agents running sequentially have **low merge cost** because only one agent is active at a time. The merge is the handoff document. No conflict resolution is needed.
Parallel specialist agents (running simultaneously on independent sub-tasks) have **high merge cost** — see the [Multi-Agent Boundary](../../../docs/concepts/multi-agent-boundary.md) concept doc for details.

#### Host-tool evidence

Before adopting specialist agents, verify the host tool supports:

1. **Agent definitions** — reads agent files from a known directory.
2. **Agent invocation** — can invoke a named agent on demand.
3. **Handoff** — supports passing context between agents.

Document the evidence in the agent contract. If the host tool lacks a required capability, fall back to a single agent + skills.

#### Anti-patterns

- **Scope creep** — An agent starts doing work outside its contract. Prevent by keeping contracts narrow and enforcing them at handoff.
- **Context bleed** — An agent carries context from a previous phase into a new phase. Prevent by starting each agent with a fresh context window.
- **Orphaned output** — An agent produces output that no agent consumes. Prevent by defining the handoff chain before implementation.
- **Fake parallelism** — Sub-tasks are not truly independent but are assigned to parallel agents anyway, causing merge conflicts. Prevent by verifying file-disjointness before splitting.

</specialist-agent-guidance>
