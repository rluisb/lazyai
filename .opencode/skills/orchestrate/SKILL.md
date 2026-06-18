---
name: orchestrate
description: Orchestrate multi-agent workflows and execution chains.
argument-hint: "[workflow-name]"
trigger: /orchestrate
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



# Orchestrate Skill

Procedure for driving the `@ai-setup/orchestrator` MCP server. Load this when a task asks to start a chain, coordinate a team, check run status/budget, or recover a failed step. Do not load for single-agent work.

## MCP Tools (20+)

| Tool | Run kinds | Purpose |
|------|-----------|---------|
| `list_catalog` | all | Discover chains, teams, workflows, domains, modes |
| `compose_agent` | all | Merge base + domain + mode → runtime `ComposedAgentSpec` |
| `start_chain` | chain | Begin a compiled chain; returns `chainId` and first step |
| `advance_chain` | chain | Record step outcome; returns next step, gate, recovery, or `done` |
| `build_team` | team | Compile and start a team run; returns `teamId` |
| `assign_team_task` | team | Assign or claim a task within a team |
| `complete_team_task` | team | Complete a team task with outcome, result, usage, error |
| `start_workflow` | workflow | Compile and start a workflow run; returns `workflowId` |
| `advance_workflow` | workflow | Advance a running workflow; handles child completion, gate decisions, and recovery |
| `get_status` | all | Current state, step history, progress for any run kind |
| `get_budget` | all | Remaining tokens and spend so far |
| `subscribe_run` | all | Subscribe to run events via SSE |
| `enqueue_job` | all | Enqueue a background job for deferred execution |
| `get_job` | all | Poll job status by ID |
| `list_jobs` | all | List queued or completed jobs |
| `retry_step` | all | Retry a failed step (if retries remain) — use `"chain"`, `"team"`, or `"workflow"` for `kind` |
| `escalate_step` | all | Reassign a step to a different agent |
| `handoff` | all | Persist a resumable handoff document |
| `catalog_list` | admin | List all catalog entries with kind/name/source |
| `catalog_get_version` | admin | Get a specific version of a catalog entry |
| `catalog_create_version` | admin | Upload a new version of a catalog entry |
| `catalog_set_active` | admin | Set the active version for a catalog entry |
| `invoke_agent` | all | Directly invoke an agent with a prompt |

## Canonical flows

The orchestrator supports three run kinds: **chain**, **team**, and **workflow**. Each has its own flow.

### Chain flow

1. **Discover** — `list_catalog({ kinds: ["chain"] })` to confirm the target exists and surface options.
2. **Budget gate** — call `get_budget` on a sentinel run or show an estimate from the chain definition. Present cost range and wait for explicit confirmation. Never skip.
3. **Classify dispatch** — before chain/wave dispatch, apply prompt-level Cupcake-signal-aware AFK/HITL awareness:
   - `plan_attested = true` means src/ writes can be AFK when other approvals/dependencies are satisfied.
   - `plan_attested = false` means src/ writes are HITL.
   - `gate_attested = false` means commits >20 lines are HITL.
   - Hard blocks like push to main/force push are always HITL.
   - Read-only/spec writes are AFK.
   This classification is prompt-level awareness only and does not change `ChainState`, `StepState`, `get_status`, approval outcomes, runtime state, Cupcake Rego, or pre-commit behavior.
4. **vocabulary alignment before dispatch** — when the task depends on project vocabulary, check accepted terms and `KNOWLEDGE_MAP.md` before dispatch. If a terminology decision is unresolved, pause as HITL instead of guessing.
5. **Start** — `start_chain({ chain, task, domainSkill?, modeSkill?, context? })`. Capture `chainId` and the first step.
6. **Loop**
   - Dispatch the agent named by the current step, using the composed prompt from `compose_agent` if the step needs domain/mode layering.
   - On step completion, call `advance_chain({ chainId, stepId, outcome, output?, usage? })`.
   - Repeat until the returned state is `done`.
7. **Observe** — between steps, call `get_status` when the user asks for progress, or `get_budget` to confirm spend is on track.

### Team flow

1. **Discover** — `list_catalog({ kinds: ["team"] })` to confirm the team definition exists.
2. **Budget gate** — same pattern as chain; present cost estimate to the user and wait for confirmation.
3. **Start** — `build_team({ team, task })`. Capture `teamId`.
4. **Loop** — for each ready member task:
   - Call `assign_team_task({ teamId, taskId, assignee })` to assign or claim the task.
   - Call `complete_team_task({ teamId, taskId, outcome, result, usage? })` when the member finishes.
   - Repeat until all member tasks are done and synthesis completes.
5. **Observe** — `get_status({ runId, kind: "team" })` for team state and per-member progress.

### Workflow flow

1. **Discover** — `list_catalog({ kinds: ["workflow"] })` to confirm the workflow definition exists.
2. **Budget gate** — same pattern as chain and team.
3. **Start** — `start_workflow({ workflow, task })`. Capture `workflowId`.
4. **Loop**
   - Call `advance_workflow({ workflowId, phaseId, outcome })` after each phase completes.
   - Repeat until the workflow reaches a terminal phase.
5. **Observe** — `get_status({ runId, kind: "workflow" })` for workflow phase progress.

## Lifecycle reporting vocabulary

Use lifecycle labels in status reports, handoff notes, recovery summaries, and completion reports so humans and downstream agents can interpret progress consistently. This is report vocabulary only: it does not add runtime per-agent state tracking and does not imply runtime state-machine support.

**Lifecycle label values:** `loading_context`, `planning`, `awaiting_approval`, `executing`, `verifying`, `blocked`, `handoff`, `done`, `error`.

## Recovery patterns

### Safe Auto-Recovery Policy

Use the Safe Auto-Recovery Policy before calling recovery tools. This guidance is static/prompt-only; runtime autonomous recovery is deferred and the orchestrator must not imply a runtime classifier, automatic edit loop, or changed retry semantics.

- **Auto-allowed:** re-run deterministic checks, retry transient provider/tool failures within existing retry limits, regenerate malformed report JSON from the same inputs, or create a handoff when blocked.
- **Human-gated:** code edits, dependency changes, destructive commands, migration changes, secrets/config changes, ambiguous failures, scope changes, or any action outside the approved task boundary.
- **Required evidence:** report the failure cause/evidence, retry limit and current attempt count, idempotency/safety check, and why the selected pattern is safe before acting.
- **Approval boundary:** if the failure is ambiguous or the action is not auto-allowed, pause for human confirmation before recovery.

| Pattern | Trigger | Call |
|---------|---------|------|
| Retry | Transient failure, retries remain | `retry_step({ runId, kind, stepId, reason })` — use `"chain"`, `"team"`, or `"workflow"` for `kind` |
| Fix & Resume | User fixed the issue manually | `advance_chain` / `advance_workflow` on the failed step with `outcome: "success"` |
| Escalate | Wrong approach / wrong agent | `escalate_step({ runId, kind, stepId, targetAgent, reason })` |
| Handoff | Context exhausted or fundamental block | `handoff({ runId, kind, summary, includeArtifacts: true })` |

After any failure, report to the user before acting: run kind, step/phase/task ID, agent, exact error, what succeeded so far, recommended pattern and why.

## Background job patterns

For long-running or deferrable work, use the job queue:

1. **Enqueue** — `enqueue_job({ jobType, payload })` to submit work. Returns a `jobId`.
2. **Check** — poll `get_job({ jobId })` for `"pending"`, `"claimed"`, `"completed"`, or `"failed"` status.
3. **List** — `list_jobs({ status? })` to audit queued or completed jobs.

Enqueue jobs for work that does not need to block the current run — e.g., sending notifications, generating large artifacts, running post-synthesis reports.

## When NOT to orchestrate

- Single-agent task — call the agent directly.
- No catalog entry matches — ask the user before inventing a chain, team, or workflow.
- **No team definition matches** — do not start a team run if `list_catalog({ kinds: ["team"] })` returns empty.
- **No workflow definition matches** — do not start a workflow run if `list_catalog({ kinds: ["workflow"] })` returns empty.
- User explicitly wants manual step-by-step control.

## Integration
- Primary agent: Orchestrator
- MCP server: `@ai-setup/orchestrator` (stdio)
- Output: chain artifacts written by the runtime to `.ai/orchestration/state/`
- Run events: use `subscribe_run` to receive real-time SSE event streams for long-running chain, team, or workflow executions.

## Skill Invocation Map

The following skills invoke orchestrate MCP tools to coordinate their execution:

| Skill | Tools Used | Purpose |
|-------|-----------|---------|
| `parallel-execution` | `enqueue_job`, `get_job`, `subscribe_run`, `retry_step`, `escalate_step` | Dispatch wave tasks as background jobs, poll for completion, monitor progress, handle failures |
| `process-audit` | `get_status`, `get_budget` | Check run status and budget usage during audit workflow |
| `self-improve` | `get_status`, `subscribe_run` | Monitor self-improvement task progress |
| `dynamic-compose` | `start_chain`, `advance_chain`, `get_status` | Compose and run ephemeral chains dynamically |
| `catalog-manage` | `catalog_list`, `catalog_get_version`, `catalog_create_version`, `catalog_set_active` | Version, diff, promote, deprecate, and remove catalog entries |

### Tool Coverage Summary

| Tool | Skills that use it |
|------|---------------------|
| `list_catalog` | catalog-manage |
| `compose_agent` | dynamic-compose |
| `start_chain` | dynamic-compose |
| `advance_chain` | dynamic-compose |
| `build_team` | — |
| `assign_team_task` | — |
| `complete_team_task` | — |
| `start_workflow` | — |
| `advance_workflow` | — |
| `get_status` | process-audit, self-improve |
| `get_budget` | process-audit |
| `subscribe_run` | parallel-execution, self-improve |
| `enqueue_job` | parallel-execution |
| `get_job` | parallel-execution |
| `list_jobs` | — |
| `retry_step` | parallel-execution |
| `escalate_step` | parallel-execution |
| `handoff` | — |
| `catalog_list` | catalog-manage |
| `catalog_get_version` | catalog-manage |
| `catalog_create_version` | catalog-manage |
| `catalog_set_active` | catalog-manage |
| `invoke_agent` | — |