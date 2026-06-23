## Summary

Investigate and implement a shared LazyAI workflow plugin/extension runtime for host tools that do not have native workflow files, starting with OpenCode, Pi, OMP, and Antigravity.

The goal is not to make every tool understand Claude workflows directly. The goal is to compile LazyAI's canonical workflow catalog into each host's real extension/plugin mechanism when that host supports one.

## Motivation

Current state:

- LazyAI embeds workflow docs under `packages/cli/library/workflows/`.
- Workflows are marked `adapter_targets: [none]` and docs-only in `packages/cli/library/manifests/curation.yaml`.
- Most targets do not have native `workflows/` directories.
- Several targets **do** have plugin/extension mechanisms that could expose workflows as commands/tools/UI gates.

## Host extension/plugin evidence

### OpenCode

Docs: https://opencode.ai/docs/plugins/#create-a-plugin

Key points from docs:

- Project plugins: `.opencode/plugins/`
- Global plugins: `~/.config/opencode/plugins/`
- Plugins are JavaScript/TypeScript modules.
- Plugins can hook events and customize behavior.
- Config can load npm plugins via `opencode.json` `plugin` array.
- Events include command, file, permission, session, todo, tool, and TUI events.

Potential LazyAI target:

```text
.opencode/plugins/lazyai-workflows.ts
.opencode/package.json        # only if dependencies are needed
```

### Pi

Docs: https://pi.dev/docs/latest/extensions

Key points from docs:

- Project extensions: `.pi/extensions/*.ts` or `.pi/extensions/*/index.ts`
- Global extensions: `~/.pi/agent/extensions/`
- Extensions can register commands with `pi.registerCommand()`.
- Extensions can register tools with `pi.registerTool()`.
- Extensions can subscribe to lifecycle/tool/session events.
- Extensions can use `ctx.ui` for confirmation/input/notify.
- TypeScript works without compilation via Pi's loader.

Potential LazyAI target:

```text
.pi/extensions/lazyai-workflows/index.ts
```

### OMP

Docs: https://omp.sh/docs/extension-authoring and https://omp.sh/docs/slash

Need source-verified implementation research because `read` retrieved a JS-heavy shell, but OMP exposes `/extensions` and extension authoring docs. Investigate exact filesystem/package format before implementation.

Potential LazyAI target, to verify:

```text
.omp/extensions/lazyai-workflows/...
```

### Antigravity

Docs: https://antigravity.google/docs/cli-plugins

Need source-verified implementation research because initial `read` was JS-heavy. Search snippets indicate plugins are bundles with `plugin.json`, optional `mcp_config.json`, `hooks.json`, `skills/`, `agents/`, `rules/`.

Potential LazyAI target, to verify:

```text
.gemini/antigravity-cli/plugins/lazyai-workflows/plugin.json
```

or project-local equivalent if docs support one.

## Design direction

Create a shared workflow engine core in embedded assets, then generate thin host adapters:

```text
packages/cli/library/workflows/              # canonical workflow docs/catalog
packages/cli/library/opencode/plugins/lazyai-workflows.ts
packages/cli/library/pi/extensions/lazyai-workflows/index.ts
packages/cli/library/omp/extensions/lazyai-workflows/...
packages/cli/library/antigravity/plugins/lazyai-workflows/...
```

Shared semantics:

- list available LazyAI workflows
- show workflow purpose, inputs, steps, exit gate
- start a workflow by injecting/printing the right prompt/context
- optionally enforce objective gates via hooks where the host supports it
- never maintain a hidden orchestration daemon or queue

Possible commands/tools exposed by each host plugin:

- `lazyai.workflow.list`
- `lazyai.workflow.show <name>`
- `lazyai.workflow.start <name>`
- `lazyai.workflow.gate.check`

## Open questions

- Can OpenCode plugins append prompts or register slash commands directly, or only react to `tui.command.execute` / `tui.prompt.append` events?
- Can a common TypeScript core be shared across OpenCode/Pi/OMP, or do loader/runtime APIs force per-host code?
- Should Antigravity use a project-local plugin, global plugin, or existing `.agents/skills` + hooks path?
- How should plugin installation be controlled: default for full preset only, explicit flag, or per-target adapter option?

## Acceptance criteria

- [ ] Source-verify exact plugin/extension packaging and lifecycle for OpenCode, Pi, OMP, and Antigravity.
- [ ] Decide whether a shared TypeScript workflow core is feasible.
- [ ] Implement OpenCode workflow plugin only if it can expose real user-facing commands/tools without unsupported hacks.
- [ ] Implement Pi workflow extension using `.pi/extensions/...` and `pi.registerCommand()` / `pi.registerTool()` if feasible.
- [ ] Implement OMP workflow extension only after exact extension format is verified.
- [ ] Implement Antigravity workflow plugin only after exact plugin format and install location are verified.
- [ ] Add adapter install tests for each emitted plugin/extension path.
- [ ] Add negative tests proving no unsupported `workflows/` directories are emitted.
- [ ] Document that plugin-backed workflows are workflow helpers, not LazyAI becoming a runtime orchestrator.
