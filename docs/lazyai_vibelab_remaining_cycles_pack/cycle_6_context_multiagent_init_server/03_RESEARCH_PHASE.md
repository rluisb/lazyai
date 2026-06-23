# Research Phase — Cycle 6

Inspect context, handoff, multi-agent, init, and server behavior.

## Required inspection

```text
packages/cli/library/fragments/
packages/cli/library/templates/
packages/cli/library/skills/handoff*
packages/cli/library/skills/parallel-execution*
packages/cli/library/canonical/agents/
docs/concepts/
docs/workflows/
packages/cli/cmd/init.go
packages/cli/cmd/server.go
packages/cli/cmd/doctor.go
packages/cli/cmd/status.go
```

Answer:

```text
- What context/handoff templates already exist?
- Is context compaction event-driven in docs?
- What multi-agent guidance already exists?
- Does any doc imply LazyAI orchestrates agents?
- How does init headless populate behave?
- How is server L1/L3 capability explained?
- Does doctor/status communicate limitations clearly?
```
