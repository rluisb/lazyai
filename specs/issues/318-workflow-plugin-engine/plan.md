# Issue 318 — Workflow plugin/extension engine plan

## Decision
Do not implement a shared workflow plugin engine until each host's packaging and command/tool APIs are source-verified. Implement host-specific thin adapters over the canonical `packages/cli/library/workflows/` catalog only where the host exposes a real extension/plugin surface.

## Verified host surfaces

| Host | Verified from docs | Safe first LazyAI target |
|---|---|---|
| OpenCode | Local plugins auto-load from `.opencode/plugins/`; plugins are JavaScript/TypeScript modules exporting plugin functions; events include `tui.command.execute`, `tui.prompt.append`, `tool.execute.*`, and session/todo events. | `.opencode/plugins/lazyai-workflows.ts`, only after confirming user-facing command behavior can be provided without unsupported TUI hacks. |
| Pi | Extensions auto-discover from `.pi/extensions/*.ts` or `.pi/extensions/*/index.ts`; TypeScript is loaded without compilation; extensions can register commands with `pi.registerCommand()` and tools with `pi.registerTool()`. | `.pi/extensions/lazyai-workflows/index.ts` exposing `lazyai.workflow.list`, `lazyai.workflow.show`, and `lazyai.workflow.start`. |
| OMP | Public docs endpoint did not return enough source text for extension packaging verification in this inspection. | No emitted plugin until exact project/global extension layout and API are verified. |
| Antigravity | Public docs endpoint did not return enough source text for CLI plugin packaging verification in this inspection. | No emitted plugin until exact plugin bundle schema and project/global install location are verified. |

## Engine boundary

LazyAI should compile workflow helpers; it should not become a workflow daemon, task queue, hidden orchestrator, or agent runtime.

A future shared core may be a static TypeScript module that:

- reads an embedded/generated workflow catalog snapshot;
- lists available workflows;
- renders workflow purpose, inputs, steps, and exit gate;
- starts a workflow by inserting or printing the host-native prompt/context;
- delegates gating to host hooks/extensions only where verified.

## Implementation sequence

1. Add a workflow delivery matrix first (#317) so unsupported `workflows/` directories remain forbidden.
2. Implement Pi first if a plugin-backed slice is desired: Pi has verified project extension placement and first-class `registerCommand` / `registerTool` APIs.
3. Implement OpenCode second only after validating whether plugins can provide a discoverable user-facing command, not only react to TUI events.
4. Re-run OMP and Antigravity documentation capture with source-verifiable pages before writing adapter assets.

## Negative constraints

- Do not emit `.opencode/workflows`, `.pi/workflows`, `.omp/workflows`, or Antigravity workflow directories.
- Do not add npm/runtime dependencies to LazyAI for plugin helpers.
- Do not claim plugin helpers are native workflow runtimes.
