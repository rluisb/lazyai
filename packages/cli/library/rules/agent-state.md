# Rule: Agent Lifecycle State Taxonomy

**Category:** Process
**Status:** Active

---

## Rule

Use one bounded lifecycle label when reporting agent progress, handoffs, recovery summaries, or completion evidence. These labels are report vocabulary only: they describe what the agent is doing in prose and handoff artifacts.

This rule does not add runtime per-agent state tracking and does not imply runtime state-machine support. Wave 2 does not change `ChainState`, `StepState`, persistence, or `get_status`; runtime lifecycle tracking requires a separate approval decision and ADR.

## Lifecycle Vocabulary

Use these exact labels instead of ad hoc status words:

| Label | Use When |
|---|---|
| `loading_context` | Reading task contracts, instructions, required artifacts, standards, or prior handoffs before choosing an approach. |
| `planning` | Defining scope, approach, touch map, verification matrix, risks, assumptions, or approval needs. |
| `awaiting_approval` | Stopped because a human approval, clarification, budget confirmation, or scope decision is required before continuing. |
| `executing` | Performing the approved implementation, content edit, command, delegation, or other task action. |
| `verifying` | Running tests, lint, type checks, reviews, scope checks, or acceptance-evidence collection. |
| `blocked` | Unable to proceed safely because of missing information, failing prerequisites, unavailable tools, policy limits, or repeated failed attempts. |
| `handoff` | Preparing or emitting resumable context for another session, agent, or human decision. |
| `done` | Approved task/session work is complete and all required evidence has been reported, or a terminal success summary is being emitted. |
| `error` | A command, tool, test, runtime call, or validation step failed and the agent is reporting the failure or recovery path. |

## Reporting Guidance

- **Status reports:** include `Lifecycle label: <label>`, one-sentence current activity, evidence already gathered, and the next decision/action.
- **Handoffs:** include the last known `Lifecycle label`, current objective, files/artifacts touched, verification evidence, blockers, and next safe action.
- **Recovery summaries:** include `Lifecycle label: error` for the failure report, then `blocked`, `awaiting_approval`, `executing`, or `handoff` depending on the selected recovery path.
- **Completion reports:** use `Lifecycle label: done` only when acceptance criteria and verification evidence are complete; otherwise use `blocked` or `awaiting_approval` with the unmet criteria.

## Common-Term Mapping

Translate common free-form status words into the bounded vocabulary instead of introducing new labels:

- pending / queued context work → `loading_context` or `planning`
- in_progress / running → `executing` or `verifying`
- needs_approval → `awaiting_approval`
- completed / complete / success → `done`
- failed / failure → `error`
- cancelled / paused by policy → `blocked` or `awaiting_approval`, with the reason stated

## Enforcement

- Agents must use the vocabulary consistently in progress updates, handoff notes, recovery summaries, and final completion reports.
- Do not create new lifecycle labels for one-off statuses; add a rationale and ask for approval if the vocabulary is insufficient.
- Do not persist labels, add lifecycle fields, change `ChainState` or `StepState`, or change `get_status` output for this guidance layer.
