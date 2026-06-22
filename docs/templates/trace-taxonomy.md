# Trace Taxonomy

> **Canonical source:** `packages/cli/library/templates/trace-taxonomy.md`
>
> This page is a documentation reference. The canonical template lives in the embedded library and is the single source of truth.

The trace taxonomy is a classification vocabulary for agent-session evidence traces. It defines five mutually exclusive categories and granular tags for structured review, failure analysis, and improvement-loop feed-in.

**Boundary:** This is an **evidence taxonomy for human and host-tool review** — not a trace daemon, not a runtime capture engine, not a scoring/judge runtime.

## Categories

| Category | Definition | What it captures |
|---|---|---|
| `context` | What the agent knew before acting | Loaded files, environment snapshot, instructions, spec references, prior decisions, assumptions recorded |
| `tooling` | How the agent interacted with tools | Tool invocations, parameters, results, errors, timeouts, retries, fallback chains |
| `workflow` | The process the agent followed | Phase transitions, gate checks, handoffs, human-gate stops, lifecycle label changes |
| `quality` | What the agent produced and how it verified | Test results, lint/build output, evidence records, diff quality, acceptance-criterion satisfaction |
| `adapter` | How the agent adapted to project constraints | Pattern following, convention application, constraint honoring, speculative-code avoidance, escalation decisions |

## Tags

### Context tags
`spec_loaded`, `codebase_explored`, `env_snapshot`, `constraints_known`, `assumptions_recorded`, `prior_decision`, `instructions_loaded`

### Tooling tags
`tool_invoked`, `tool_result`, `tool_error`, `tool_timeout`, `tool_retry`, `tool_fallback`, `tool_chain`

### Workflow tags
`phase_start`, `phase_complete`, `gate_passed`, `gate_failed`, `human_gate`, `handoff`, `lifecycle_change`, `escalation`

### Quality tags
`test_passed`, `test_failed`, `lint_ok`, `lint_error`, `build_ok`, `build_error`, `evidence_recorded`, `ac_satisfied`, `ac_failed`, `regression_detected`

### Adapter tags
`pattern_followed`, `convention_applied`, `constraint_honored`, `speculative_avoided`, `escalation`, `minimal_change`

## Evidence Record Format

```json
{
  "category": "context | tooling | workflow | quality | adapter",
  "tags": ["tag1", "tag2"],
  "timestamp": "YYYY-MM-DDTHH:MM:SSZ",
  "phase": "research | plan | implement | verify | cleanup",
  "summary": "One-line description of what happened",
  "evidence": "Path, command output, or reference to the supporting artifact",
  "outcome": "success | failure | partial | blocked"
}
```

Minimal form: `[category:tag1,tag2] summary — evidence: <path or ref>`

## Related

- [Trace/Eval Improvement Loop](../concepts/trace-eval-improvement-loop.md)
- [Canonical template: trace-taxonomy.md](../../packages/cli/library/templates/trace-taxonomy.md)
