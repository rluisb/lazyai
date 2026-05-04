# Rule: Structured Feedback

**Category:** Process
**Status:** Active

---

## Rule

When human gate feedback, review-request changes, or rejection notes are available, express and consume them as `StructuredFeedback` so the next agent can identify what to fix, where to look, and whether progress is blocked.

This is prompt/library guidance first. Approved Wave 2 runtime support is bounded to existing rejected-gate output that includes `structuredFeedback`; it does not alter approval outcomes or add a new gate system. Outside that approved T021 path, treat feedback as static prompt context.

## Rationale

Free-form rejection notes often omit the exact required change, evidence, or target step. A small schema makes feedback actionable while preserving human control and avoiding speculative fixes.

## StructuredFeedback Contract

Emit or request one `StructuredFeedback` JSON object when feedback is intended to guide a follow-up agent:

```json
{
  "schemaVersion": "structured-feedback/v1",
  "verdict": "approved|request_changes|rejected|comment",
  "summary": "string",
  "requiredChanges": [
    {
      "id": "FB-001",
      "description": "string",
      "priority": "blocking|high|medium|low",
      "target": "step/task/file",
      "evidence": "string",
      "location": {
        "file": "repo-relative/path",
        "section": "string or null",
        "lineStart": 1,
        "lineEnd": 1
      },
      "recommendedNextAction": "string",
      "blocksProgress": true
    }
  ],
  "suggestions": [
    {
      "description": "string",
      "priority": "medium|low",
      "target": "step/task/file"
    }
  ],
  "requestedBy": "human|reviewer|red-team|planner",
  "targetPhaseOrStep": "string or null"
}
```

The Wave 2 base schema is `schemaVersion`, `verdict`, `summary`, `requiredChanges`, `suggestions`, `requestedBy`, and `targetPhaseOrStep`. Required-change entries add bounded actionability fields for evidence/location, recommended next action, and whether the item `blocksProgress`.

## Field Guidance

- **Source:** use `requestedBy` to identify the source: human, reviewer, red-team, or planner.
- **Severity / priority:** use `priority`; `blocking` and `high` items are required changes, while optional suggestions are `medium` or `low` only.
- **Finding/action:** use `description` for the finding and `recommendedNextAction` for the concrete action the next agent should take.
- **Evidence/location:** include quoted evidence or command output in `evidence`; include a repo-relative `location` when a file/section/line is known, otherwise set `location` to `null` and explain the evidence source.
- **Target:** name the target phase/task/file in `target` and use `targetPhaseOrStep` when feedback is aimed at a specific chain phase or step.
- **Progress block:** set `blocksProgress: true` when the item must be resolved before the current phase can advance.

## Clarification Path

If `verdict` is `request_changes` or `rejected` but `requiredChanges` is empty, vague, or lacks evidence/target detail, the next agent must ask for clarification rather than guessing. Treat non-blocking comments and suggestions as optional unless a human explicitly converts them into required changes.

## Consumption Rules

1. Read the summary and group feedback by required changes first, suggestions second.
2. Address blocking required changes before high/medium/low items.
3. For each required change, cite the source, priority, target phase/task/file, evidence/location, and recommended next action in the work plan.
4. If a required change is ambiguous, missing evidence, missing a target, or contradicts the approved task boundary, stop and ask for clarification.
5. Preserve the approved task scope. Do not turn suggestions into required work without human approval.

## Enforcement

- Iterate and orchestrate guidance must consume `StructuredFeedback` when present.
- Rejection feedback without required-change detail triggers clarification, not invented fixes.
- Runtime handling is limited to the approved T021 rejected-gate `output.structuredFeedback` path; no new gate engine, outcome taxonomy, measurement hooks, or broad propagation is implied.
