# Implementation Scope — Cycle 6

## Priority A — Context compaction and handoff templates

Add/update:

```text
packages/cli/library/fragments/context-compaction.md
packages/cli/library/templates/context-handoff.md
packages/cli/library/templates/session-compaction.md
packages/cli/library/templates/recovery-handoff.md
docs/concepts/context-discipline.md
```

Rules:

```text
Compact after:
- research phase
- planning phase
- implementation phase
- failed verification
- agent/tool handoff
- before context gets too large

Never compact away:
- user constraints
- decisions
- failed attempts
- commands run
- files touched
- unresolved risks
- human approval state
```

Acceptance:

```text
- Handoff templates require evidence.
- Context compaction is guidance/assets, not runtime automation.
```

## Priority B — Multi-agent boundary

Add/update:

```text
docs/concepts/multi-agent-boundary.md
packages/cli/library/templates/multi-agent-decision.md
packages/cli/library/fragments/specialist-agent-guidance.md
```

Guidance:

```text
Use one agent + skills when:
- task is local
- context fits
- one person/agent can reason through it
- no isolation needed

Use specialist agents when:
- role separation improves quality
- review should be independent
- domain expertise differs

Use parallel agents when:
- work is read-heavy
- outputs are easy to merge
- boundaries are explicit

Avoid multi-agent when:
- agents need the same full context
- parallel writes would conflict
- merge cost exceeds benefit
- host tool cannot show evidence
```

Acceptance:

```text
- No orchestration command.
- No runtime scheduler.
- Guidance compiles as assets/templates where appropriate.
```

## Priority C — Init headless populate clarity

Inspect and update:

```text
packages/cli/cmd/init.go
README.md
docs/commands/init.md if present
```

Clarify:

```text
- why AGENTS.md placeholder fill may be skipped
- what host tool should do next
- how to complete setup manually
- how to validate after populate
```

Acceptance:

```text
- No hidden runtime behavior added.
- CLI explains what happened and next steps.
```

## Priority D — Server L1/L3 capability clarity

Inspect and update:

```text
packages/cli/cmd/server.go
README.md
docs/commands/server.md if present
doctor/status messaging if relevant
```

Clarify:

```text
- L1 config check
- L3 handshake not implemented in Go
- Node.js MCP SDK dependency if still true
- current limitation vs future work
```

Acceptance:

```text
- Users can distinguish L1 config check vs L3 handshake.
- No fake support is claimed.
```

## Priority E — Tests

Add/update tests for changed help text or docs references if repo conventions include such tests.

Acceptance:

```text
- Tests pass.
- Product boundary remains intact.
```
