# Hook Lifecycle & Capability Matrix

This document defines the canonical hook lifecycle vocabulary and classifies
each adapter's hook support using a five-level capability scale. It is the
single source of truth for what each host tool can do with hooks.

---

## Lifecycle Vocabulary

The canonical lifecycle covers ten events in the agent loop:

| # | Event | Phase | Description |
|---|---|---|---|
| 1 | `before_agent` | Pre-invocation | Before an agent or subagent is invoked |
| 2 | `before_model` | Pre-LLM | Before an LLM model call |
| 3 | `before_tool` | Pre-tool | Before a tool or function call |
| 4 | `after_tool` | Post-tool | After a tool call returns |
| 5 | `after_model` | Post-LLM | After an LLM model call returns |
| 6 | `after_agent` | Post-invocation | After an agent completes |
| 7 | `on_error` | Error | On tool or model error |
| 8 | `on_compaction` | Compaction | On context window compaction |
| 9 | `on_handoff` | Handoff | On session handoff between agents |
| 10 | `on_human_gate` | Human gate | On a human approval gate |

---

## Capability Scale

Each adapter's hook support is classified with one of five levels:

| Level | Meaning |
|---|---|
| `supported` | Full native hook mechanism; verified runtime with multiple lifecycle events |
| `partial` | Some lifecycle events supported; limited event coverage or scope |
| `instruction_only` | Capability declared in adapter metadata; no runtime hook files emitted; hooks documented in root instructions only |
| `unsupported` | No hook capability; no hook files emitted |
| `not_applicable` | Concept does not apply to this tool |

---

## Adapter Hook Capability Matrix

| Adapter | Support Level | Hook Capability | Classification | Evidence |
|---|---|---|---|---|
| opencode | stable | `Hooks: true` | **supported** | Full plugin runtime (`vibe-lab-hooks.js`); events: `session.created`, `tool.execute.before`, `session.idle`, `experimental.session.compacting` |
| claude | stable | `Hooks: true` | **supported** | Shell command hooks at `.claude/hooks/*.sh`; events: `PreToolUse`, `Stop`; wired via `.claude/settings.json` |
| copilot | stable | `Hooks: true` | **partial** | JSON descriptors + shell scripts at `.github/hooks/*.{json,sh}`; project scope only; limited to pre-exec and stop events |
| pi | stable | `Hooks: true` | **instruction_only** | No `.pi/hooks` directory emitted; only `block-destructive-shell` has an extension runtime at `.pi/extensions/*.ts`; all other hooks are markdown-only |
| omp | beta | `Hooks: true` | **partial** | TypeScript hook factories at `.omp/hooks/pre/*.ts`; only `before_tool` surface (pre hooks); beta support level |
| antigravity | beta | `Hooks: true` | **partial** | Shell scripts at `.gemini/hooks/lazyai/*.sh` + `.agents/hooks.json` + `.gemini/settings.json`; limited to pre-exec and stop events; beta support level |
| kiro | stable | `Hooks: false` | **instruction_only** | Adapter does not declare runtime hook support; no `.kiro/hooks` directory is emitted; hook guidance is markdown-only |

### Notes

- **Pi** declares `Hooks: true` but is `instruction_only`: most hook guidance is
  documented for agent awareness, and only the destructive-shell guard has an
  extension runtime.
- **Kiro** does not declare runtime hook support (`Hooks: false`) and emits no
  `.kiro/hooks` files; hook guidance remains markdown-only until a native output
  contract is source-verified.
- **OMP** and **Antigravity** are classified as `partial` because they emit hook
  files but are at `beta` support level and cover a limited subset of lifecycle
  events.
- **Copilot** is `partial` because hooks are project-scope only (not global) and
  limited to pre-exec and stop events.

---

## Lifecycle Event Coverage by Adapter

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

**`—`**: No native hook point is targeted by any shipped LazyAI hook policy
for this event on this tool.

---

## Compilation Boundary

LazyAI compiles hook policies into tool-native formats. It does not execute
hooks. The host tool is responsible for:

1. Reading the compiled hook files from their tool-specific directories.
2. Dispatching them at the mapped lifecycle events.
3. Handling hook output (block, warn, allow).

See [Harness Principles](../concepts/harness-principles.md) for the full product
boundary definition.

---

## See Also

- [Hook Catalog](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/hooks/catalog.md) — lifecycle vocabulary, event mapping, compilation boundary
- [Hooks Reference](../reference/hooks.md) — shipped hook inventory
- [Harness Principles](../concepts/harness-principles.md) — product boundary
