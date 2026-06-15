# objective-workflow-gate Policy

## Purpose

Block only clear, observable workflow-exit contract failures: a response that explicitly claims completion but provides neither verification evidence nor a blocked/not-run reason.

## Events

- Claude Code: `Stop`
- OpenCode plugin: `session.idle` with `lastAssistantMessage`, or stored assistant text from `message.updated`
- OMP/Pi: unsupported

## Decision

- Allow when: the message does not explicitly claim completion, or it claims completion and includes a `Verification:`, `Validation:`, `Tests:`, `Checks:`, `Smoke:`, or `Evidence:` line.
- Allow when: the message states a blocker or not-run validation reason with words such as `Blocked:`, `Not run:`, `Could not run`, or `Missing prerequisite`.
- Warn when: subjective quality, completeness, or goal-achievement concerns are not machine-observable. Those remain workflow guidance, not hook decisions.
- Deny when: the runtime observes a final assistant message with an explicit completion marker such as `Done`, `Complete`, `Completed`, `Implemented`, `Fixed`, `Changed`, `Updated`, or `Shipped`, and no verification or blocked/not-run evidence marker is present.

## Runtime

- Claude Code: generated command hook runs at `Stop`, reads JSON from stdin, checks `last_assistant_message`, and returns top-level `decision: "block"` with a reason when evidence is objectively missing.
- OpenCode: generated plugin checks `session.idle` events when final assistant text is present directly or previously captured from `message.updated`; event-shape smoke tests cover the local plugin behavior. Live OpenCode runtime blocking for idle events is not claimed beyond that verified boundary.
- OMP/Pi: no project-local hook is generated.

## Fail-Closed Semantics

- Claude Code: malformed Stop hook input or missing `last_assistant_message` blocks because the supported runtime contract is unavailable.
- OpenCode: missing final assistant text does not block because the plugin cannot observe the condition; malformed observable text blocks by throwing.
- Subjective completion quality is never inferred by this hook.

## Scope

- This gate checks text shape, not truth. It requires evidence to be stated; it does not verify whether a command actually passed.
- Repo-level truth remains with `bin/doctor`, focused tests, and human review.
- The gate does not create a workflow engine, queue, ledger, memory store, or dispatcher.
