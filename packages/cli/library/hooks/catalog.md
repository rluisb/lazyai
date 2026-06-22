# Hook Lifecycle Catalog

LazyAI compiles hook policy assets from the canonical library into tool-native
formats. **LazyAI does not execute hooks.** Host tools (OpenCode, Claude Code,
GitHub Copilot, Pi, OMP, Antigravity, Kiro) are the execution engines; they
read the compiled hook files and invoke them at lifecycle events.

This catalog defines the lifecycle vocabulary, maps each event to host-tool
support, and documents the boundary between compilation and execution.

---

## Lifecycle Vocabulary

The following lifecycle events form the canonical vocabulary. Each event
describes a point in a host tool's agent loop where a hook may intervene.

| Event | Trigger | Typical Use |
|---|---|---|
| `before_agent` | Before an agent/subagent invocation | Session bootstrap, context injection |
| `before_model` | Before an LLM model call | Prompt shaping, guardrail injection |
| `before_tool` | Before a tool/function call | Safety checks, input validation |
| `after_tool` | After a tool/function call returns | Output sanitization, evidence capture |
| `after_model` | After an LLM model call returns | Response validation, completion gating |
| `after_agent` | After an agent/subagent completes | Handoff preparation, memory promotion |
| `on_error` | On tool or model error | Error reporting, fallback behavior |
| `on_compaction` | On context window compaction | Memory promotion, summary preservation |
| `on_handoff` | On session handoff to another agent | Context transfer, state serialization |
| `on_human_gate` | On a human approval gate | Gate enforcement, evidence verification |

---

## Host-Tool Event Mapping

Each host tool exposes a subset of these lifecycle events through its native
hook mechanism. LazyAI compiles hook policies into the formats each tool
expects; the tool's runtime dispatches them.

| Lifecycle Event | opencode | claude | copilot | pi | omp | antigravity | kiro |
|---|---|---|---|---|---|---|---|
| `before_agent` | session.created | — | — | — | — | — | — |
| `before_model` | — | — | — | — | — | — | — |
| `before_tool` | tool.execute.before | PreToolUse | pre-exec | extension | hooks/pre/*.ts | pre-exec | — |
| `after_tool` | — | — | — | — | — | — | — |
| `after_model` | session.idle | Stop | stop | — | — | stop | — |
| `after_agent` | — | — | — | — | — | — | — |
| `on_error` | — | — | — | — | — | — | — |
| `on_compaction` | experimental.session.compacting | — | — | — | — | — | — |
| `on_handoff` | — | — | — | — | — | — | — |
| `on_human_gate` | — | — | — | — | — | — | — |

Key:
- **Named event**: the host tool's native hook point that LazyAI targets.
- **`—`**: no native hook point is targeted by any shipped LazyAI hook policy
  for this event on this tool. The lifecycle event may still be documented in
  root instructions for agent awareness.

---

## Non-AI-Tool Hooks

Some shipped hooks operate outside the AI agent lifecycle:

| Hook | Mechanism | Lifecycle |
|---|---|---|
| `pre-commit` | Git pre-commit hook (`.githooks/pre-commit`, `.husky/pre-commit`) | Pre-commit |
| `rpi-gate-check.yml` | GitHub Actions workflow | CI/CD (pull_request, push) |

These are compiled as repository-level assets, not AI-tool hook files. They
are included in the catalog because LazyAI manages them alongside AI-tool
hooks, but their execution is handled by git and GitHub Actions respectively.

---

## Compilation Boundary

LazyAI's responsibility ends when the hook files are written to disk:

1. **Source**: Hook policies live in `packages/cli/library/hooks/` and
   `packages/cli/library/canonical/hooks/`.
2. **Compilation**: `lazyai compile` transforms policies into tool-native
   formats (shell scripts, TypeScript plugins, JSON descriptors, YAML
   workflows).
3. **Execution**: The host tool reads the compiled files and invokes them at
   the mapped lifecycle events. LazyAI is not involved.

This boundary is enforced by the adapter contract: adapters emit hook files
into tool-specific directories; they do not register cron jobs, spawn
background processes, or install system-level schedulers.

---

## See Also

- [Hook Lifecycle & Capability Matrix](../docs/concepts/hook-lifecycle.md) —
  per-adapter capability classification
- [Hooks Reference](../docs/reference/hooks.md) — shipped hook inventory
- [Harness Principles](../docs/concepts/harness-principles.md) — product
  boundary between LazyAI and host tools
