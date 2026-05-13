---
name: orchestrate
description: Orchestrate multi-agent workflows and execution chains.
argument-hint: "[workflow-name]"
trigger: /orchestrate
---

# Orchestrate Skill

Procedure for driving the `@ai-setup/orchestrator` MCP server. Load this when a task asks to start a chain, coordinate a team, check run status/budget, or recover a failed step. Do not load for single-agent work.

## MCP Tools

### Chain & Team Execution

| Tool | Purpose |
|------|---------|
| `list_catalog` | Discover chains, teams, workflows, domains, modes |
| `compose_agent` | Merge base + domain + mode → runtime prompt |
| `start_chain` | Begin a compiled chain; returns `chainId` and first step |
| `advance_chain` | Record step outcome; returns next step or `done` |
| `get_status` | Current chain state, step history, progress |
| `get_budget` | Remaining tokens and spend so far |
| `retry_step` | Re-run a failed step (if retries remain) |
| `escalate_step` | Reassign a step to a different agent |
| `handoff` | Persist a resumable handoff document |

### Background Jobs

| Tool | Purpose |
|------|---------|
| `enqueue_job` | Queue a background job (e.g., agent_invoke) for async execution |
| `get_job` | Check status and result of a queued job |
| `list_jobs` | List queued jobs, optionally filtered by status |

### Run Monitoring

| Tool | Purpose |
|------|---------|
| `subscribe_run` | Subscribe to state-change events for a chain/team/workflow run |
| `unsubscribe_run` | Remove subscription for a run |

## Canonical flow

1. **Discover** — `list_catalog({ kinds: ["chain"] })` (or `["team"]`, `["domain"]`, `["mode"]`) to confirm the target exists and to surface options to the user.
2. **Budget gate** — call `get_budget` on a sentinel run or show an estimate from the chain definition. Present cost range to the user and wait for explicit confirmation. Never skip.
3. **Start** — `start_chain({ chain, task, domainSkill?, modeSkill?, context? })`. Capture `chainId` and the first step.
4. **Loop**
   - Dispatch the agent named by the current step, using the composed prompt from `compose_agent` if the step needs domain/mode layering.
   - On step completion, call `advance_chain({ chainId, stepId, outcome, output?, usage? })`.
   - Repeat until the returned state is `done`.
5. **Observe** — between steps, call `get_status` when the user asks for progress, or `get_budget` to confirm spend is on track.

## Recovery patterns

| Pattern | Trigger | Call |
|---------|---------|------|
| Retry | Transient failure, retries remain | `retry_step({ runId, kind: "chain", stepId, reason })` |
| Fix & Resume | User fixed the issue manually | `advance_chain` on the failed step with `outcome: "success"` |
| Escalate | Wrong approach / wrong agent | `escalate_step({ runId, kind: "chain", stepId, targetAgent, reason })` |
| Handoff | Context exhausted or fundamental block | `handoff({ runId, kind: "chain", summary, includeArtifacts: true })` |

After any failure, report to the user before acting: chain, step, agent, exact error, what succeeded so far, recommended pattern and why.

## When NOT to orchestrate

- Single-agent task — call the agent directly.
- User explicitly wants manual step-by-step control.

## When no catalog entry matches

- Load the `dynamic-compose` skill — it assembles an ephemeral chain, team, or workflow from available agents, domains, and modes.
- Never invent a chain or team ad-hoc yourself; let `dynamic-compose` handle composition with its guardrails (budget gate, agent inventory, ephemeral-by-default).

## Background Jobs (fire-and-forget agents)

When you need to dispatch an agent without waiting for the result (e.g., non-blocking tasks, parallel background work):

1. **Enqueue** — `enqueue_job({ jobType: "agent_invoke", payload: { agent: "<name>", task: "<description>" }, maxAttempts: 2 })`. Returns `jobId`.
2. **Check later** — `get_job({ jobId })`. Returns status: `pending`, `claimed`, `completed`, or `failed`.
3. **List all** — `list_jobs({ status: "pending" })` to see what's queued.

Use this for work that doesn't block the current chain — e.g., generating a report, running a lint pass, updating memory. The orchestrator will emit a `run_event` notification when the job completes if you have a subscription active.

## Run Monitoring (live events)

To watch a running chain/team/workflow for real-time state changes:

1. **Subscribe** — `subscribe_run({ runId })`. Returns immediately with past events, then sends MCP log notifications for each future state change (step started, step completed, step failed, run done, etc.).
2. **Process events** — each notification contains the run's current state and the event that triggered it.
3. **Unsubscribe** — `unsubscribe_run({ runId })` when you no longer need updates (e.g., run completed).

Use this when the user asks "how's it going?" outside the normal advance loop, or when monitoring a team's parallel progress.

## Integration
- Primary agent: Orchestrator
- MCP server: `@ai-setup/orchestrator` (stdio)
- Output: chain artifacts written by the runtime to `.ai/orchestration/state/`
- Fallback on no match: `dynamic-compose` skill
- Runtime overrides: `chain-customize` skill
- Catalog lifecycle: `catalog-manage` skill
