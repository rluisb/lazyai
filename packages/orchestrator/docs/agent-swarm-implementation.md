# Agent Swarms in `@ai-setup/orchestrator`

## 1. How Swarms are Mapped to the Codebase (The "Team" Abstraction)

In the `ai-setup` monorepo, **Agent Swarms** are implemented as **"Teams"**. While linear, dependent AI work is organized as a *Chain*, multi-agent swarms working in parallel on the same artifact are orchestrated as a *Team*.

### The JSON Definition Layer
Swarms are defined in JSON templates located in `library/orchestration/teams/`. Each definition acts as the blueprint for an Agent Swarm and contains:
- **`parallel` (Specialized Roles)**: An array of agents with highly specific roles, goals, and skills (e.g., `context-researcher`, `quality-challenger`).
- **`synthesize` (Handoff Mechanism)**: A specific agent tasked with receiving the concurrent outputs of the parallel workers and merging them into a single, cohesive result.
- **Budgeting & Guardrails**: Directives like `budget_multiplier` and `user_confirmation_required` that ensure swarms (which are token-intensive) don't spiral out of control without human approval.

### Execution Logic (`team-machine.ts`)
The `packages/orchestrator/src/team-machine.ts` file acts as the engine for the swarm:
1. **Decomposition**: `createTeamState` splits the task into individual `TeamTaskState` tasks that have no dependencies (`dependsOn: []`), allowing them to be executed completely concurrently.
2. **Synthesis Blocking**: It dynamically creates a `synthesisTask` that is marked as `blocked`. The `dependsOn` array of this synthesis task maps to every parallel worker.
3. **Triggering Handoff**: As workers complete their jobs via `completeTeamTask`, the `updateSynthesisReadiness` function listens. Once all parallel tasks are marked `completed`, the synthesis task is unlocked, passing the shared memory/context to the synthesizer.

---

## 2. Are Swarms Running in the Orchestrator MCP? (Current Limitation)

**Short Answer:** The engine and logic are fully built, but **they are completely unreachable via the MCP Server.**

### The Missing MCP Link
If we look at `packages/orchestrator/src/tool-handlers.ts`, the logic to support Swarms/Teams already exists! The class exposes methods like `buildTeam`, `assignTask`, `completeTask`, `startWorkflow`, and `advanceWorkflow`.

However, the actual MCP exposure layer (`packages/orchestrator/src/server.ts`) **completely omits them**. 

#### Specific Deficiencies in `server.ts`:
1. **No Entry Points**: There are no registered MCP tools to trigger a swarm. We have `start_chain` and `advance_chain`, but **no `build_team`, `assign_team_task`, or `start_workflow`**.
2. **Hardcoded Chain Restrictions**: Tools that *should* support polymorphic runs are hardcoded to only accept chains. For example, `get_status` and `get_budget` use a schema `CHAIN_KIND_SCHEMA = z.literal('chain')`, ignoring teams entirely.
3. **Handoff & Retries**: `handoff`, `retry_step`, and `escalate_step` also hardcode their `kind` parameters to `'chain'`.

Because of this, if a user is interacting with an AI assistant (like Claude Code) that connects to this MCP server, the assistant **cannot dispatch a parallel swarm/team** because the tools don't exist in the MCP layer.

---

## 3. How We Can Improve the MCP to Reach a Better State of Agent Swarm

To make Agent Swarms operational via MCP, we need to bridge the gap between `server.ts` and `tool-handlers.ts`. 

Here is the roadmap for improvement:

### Step 1: Register Team-Specific MCP Tools
We need to register the missing tools in `server.ts`:
- **`build_team`**: Accept a team name, task, and budget constraints. Return the team ID and `readyTaskIds`.
- **`assign_team_task`**: Allow an agent to pull/claim a specific task from a swarm.
- **`complete_team_task`**: Allow an agent to push its result and usage. If this triggers the synthesis phase, the MCP response should indicate that the `synthesisTaskId` is now ready.

### Step 2: Update Existing Tools to be Polymorphic
We need to remove the `z.literal('chain')` restrictions on generic state tools.
- Refactor `CHAIN_KIND_SCHEMA` in `server.ts` to `RUN_KIND_SCHEMA = z.enum(['chain', 'team', 'workflow'])`.
- Update `get_status`, `get_budget`, and `handoff` schemas to accept the new `kind` and correctly route to the team variants in `tool-handlers.ts`.

### Step 3: Enable Workflows via MCP (Optional but Recommended)
Workflows (`start_workflow`, `advance_workflow`) are essentially state machines that can compose both Chains and Teams. If we expose Workflows via MCP, a human could trigger a complex task that automatically launches a Research Chain, waits for a gate, and then launches an Assessment Swarm (Team).

### Step 4: Budget Transparency via MCP
Since Swarms multiply token usage (`budget_multiplier: 3`), the MCP tools must explicitly surface this to the host CLI. When `build_team` is called, the MCP server should preemptively throw a `user_approval` gate demanding human confirmation before the parallel tasks are assigned.