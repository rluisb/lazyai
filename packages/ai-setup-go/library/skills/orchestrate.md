---
name: orchestrate
description: Orchestrate multi-agent workflows and execution chains.
argument-hint: "[workflow-name]"
trigger: /orchestrate
---

# Orchestrate Skill

Procedure for driving the `@ai-setup/orchestrator` MCP server. Load this when a task asks to start a chain, coordinate a team, check run status/budget, or recover a failed step. Do not load for single-agent work.

## MCP Tools (9)

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

## Canonical flow

1. **Discover** — `list_catalog({ kinds: ["chain"] })` (or `["team"]`, `["domain"]`, `["mode"]`) to confirm the target exists and to surface options to the user.
2. **Budget gate** — call `get_budget` on a sentinel run or show an estimate from the chain definition. Present cost range to the user and wait for explicit confirmation. Never skip.
3. **Start** — `start_chain({ chain, task, domainSkill?, modeSkill?, context? })`. Capture `chainId` and the first step.
4. **Loop**
   - Dispatch the agent named by the current step, using the composed prompt from `compose_agent` if the step needs domain/mode layering.
   - On step completion, call `advance_chain({ chainId, stepId, outcome, output?, usage? })`.
   - Repeat until the returned state is `done`.
5. **Observe** — between steps, call `get_status` when the user asks for progress, or `get_budget` to confirm spend is on track.

## StructuredFeedback Relay

When a human gate rejection, review-request change, or agent report includes `StructuredFeedback`, relay it to the next assigned agent as static prompt context:

1. Preserve the feedback source (`requestedBy`), verdict, summary, and `targetPhaseOrStep`.
2. Separate required changes from suggestions; list required changes first with priority, evidence/location, target phase, target task/file, recommended next action, and whether each item blocks progress.
3. If feedback is free-form but clearly actionable, synthesize a bounded `StructuredFeedback` summary for the handoff message without changing runtime state or approval semantics.
4. If a rejected/request_changes decision lacks required changes, priority, evidence, target phase/task, or action detail, pause and ask the human for clarification; do not guess or invent fixes.
5. Treat suggestions as optional context unless the human explicitly marks them as required changes.

This relay is prompt/static guidance only. Do not claim runtime feedback persistence or propagation; T021 requires separate approval before any runtime behavior changes.

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
- No catalog entry matches — ask the user before inventing a chain or team.
- User explicitly wants manual step-by-step control.

## Integration
- Primary agent: Orchestrator
- MCP server: `@ai-setup/orchestrator` (stdio)
- Output: chain artifacts written by the runtime to `.ai/orchestration/state/`
