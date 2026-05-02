# Rule: Safe Auto-Recovery Policy

**Category:** Process
**Status:** Active

---

## Rule

Use bounded recovery only when the failure cause/evidence is known, the action is idempotent or read-only, the retry limit is explicit, and the action stays inside the approved task boundary.

This is static prompt/library guidance only. Wave 2 does not add runtime autonomous recovery, failure classifiers, runtime retry semantic changes, or automatic edit loops. Runtime autonomous recovery is deferred until a separate approval decision and ADR authorize it.

## Safe Auto-Recovery Allowlist

The following actions are auto-allowed only after recording the failure cause/evidence, confirming the idempotency/safety check, and respecting the retry limit:

1. **Safe retry:** re-run deterministic checks that do not mutate code, data, dependencies, secrets, or external systems.
2. **Transient retry:** retry transient provider/tool failures within existing retry limits when the same request or command can be safely repeated.
3. **Report repair:** regenerate malformed report JSON from the same inputs when the prior output is syntactically invalid and no new facts or decisions are invented.
4. **Blocked handoff:** create handoff when blocked, context is exhausted, or the next safe action requires a human decision.

If an action is not on this allowlist, treat it as human-gated.

## Human Confirmation Required

Require explicit human confirmation before any recovery action involving:

- code edits;
- dependency changes;
- destructive commands;
- migration changes;
- secrets/config changes;
- ambiguous failures;
- changing task scope, acceptance criteria, approvals, or runtime behavior.

No destructive recovery without explicit human approval. If the blast radius is unclear, stop and ask before proceeding.

## Recovery Pattern Guidance

| Pattern | Use When | Required Safeguards |
|---|---|---|
| Safe retry | Deterministic check failed due to a known transient or environmental condition | Record failure cause/evidence; verify read-only/idempotent action; stay within retry limit |
| Fix-and-resume | A human or approved agent has already applied the fix | Cite the approved fix/evidence; resume only the failed boundary; do not infer unapproved edits |
| Escalation | The current agent/tool/approach is likely wrong or lacks authority | Summarize evidence, attempted actions, recommended next owner, and approval needed |
| Handoff | Context, budget, tool access, or policy blocks further safe progress | Include current state, failure cause/evidence, retry count, open decisions, and next safe action |

## Decision Checklist

Before any recovery action, answer:

1. What failed, and what failure cause/evidence supports that diagnosis?
2. Is the proposed action on the Safe Auto-Recovery Allowlist?
3. Is the action idempotent, read-only, or otherwise safe to repeat?
4. What retry limit applies, and how many attempts have already occurred?
5. Could the action edit code, alter dependencies, run destructive commands, change migrations, touch secrets/config, or handle an ambiguous failure?
6. If yes to any human-gated condition, has the human explicitly confirmed this recovery path?

## Enforcement

- Orchestrator guidance must distinguish auto-allowed low-risk actions from human-gated recovery.
- Recovery summaries must include failure cause/evidence, selected pattern, retry count/limit, and idempotency/safety check.
- Do not change orchestrator runtime retry semantics, chain-machine behavior, workflow-machine behavior, or state schemas for this policy.
