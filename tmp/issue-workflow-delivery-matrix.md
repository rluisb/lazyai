## Summary

LazyAI embeds workflow catalog assets, but currently marks them docs-only and does not install them for any target. We need a per-target workflow delivery matrix that respects each AI tool's native format instead of inventing unsupported `workflows/` directories.

## Current state

Workflow assets exist:

- `packages/cli/library/workflows/*.md`
- `packages/cli/library/workflows/verified-research/templates/*.md`
- embedded by `packages/cli/library/embed.go`

But they are intentionally docs-only today:

- `packages/cli/library/manifests/curation.yaml:1011-1098`
  - `kind: workflow`
  - `adapter_targets: [none]`
  - `reason_kept: Workflow reference retained for documentation-only parity`
- `packages/cli/library/workflows/feature.md:53-57`
  - says workflows are consumed as markdown context and no runtime is claimed for OpenCode/OMP/Pi
- `packages/cli/library/skills/create-workflow.md:18-30`
  - says workflow artifacts are documentation/catalog surfaces unless a supported runtime surface is explicitly approved

Output mapping lacks workflow support:

- `packages/cli/internal/adapter/output_mapping.go:21-45`
  - asset kinds are only `agents`, `skills`, `templates`, `commands`, `chatmodes`, `output-styles`, `prompts`
- `packages/cli/cmd/create.go:92-103`
  - `create` allows only `agent`, `skill`, `prompt`, `command`, `template`, `hook`

## Per-target findings

| Target | Native workflow surface? | Correct LazyAI delivery |
|---|---|---|
| OpenCode | No documented native workflow surface found | Use plugin/commands/skills/modes/instructions; no `.opencode/workflows` unless docs prove it |
| Claude Code | Yes: Claude Code dynamic workflows can be saved under `.claude/workflows/` | Candidate for native workflow support, gated by capability |
| Copilot | No native workflow surface | Use `.github/copilot-instructions.md`, `.github/instructions/`, `.github/prompts/`, `.github/agents/`, hooks |
| Pi | Extension-backed; no `.pi/workflows` | Use `.pi/extensions`, skills, prompts |
| OMP | Extension/plugin-backed; no verified `.omp/workflows` | Use extension package plus skills/commands/prompts/hooks/agents |
| Kiro | Specs/Steering/Hooks/MCP; no workflow dir | Map workflows to `.kiro/steering` and `.kiro/specs` |
| Antigravity | Plugin-backed; no verified workflow dir | Use plugin layout / `.agents/skills` / hooks / rules |

## Proposed design

Add an explicit workflow delivery model instead of a universal `AssetKindWorkflow` installed everywhere.

Options:

1. Add `Capability.Workflows` only for truly native workflow runtimes.
   - Currently likely only Claude Code.
2. Add a separate `WorkflowDelivery` map:
   - `native`: Claude `.claude/workflows`
   - `steering/specs`: Kiro
   - `plugin`: OpenCode/Pi/OMP/Antigravity
   - `guidance-only`: Copilot or fallback
3. Keep workflow catalog canonical in `packages/cli/library/workflows/`, but project each workflow into supported surfaces.

## Non-goals

- Do not install fake `.opencode/workflows`, `.pi/workflows`, `.omp/workflows`, `.kiro/workflows`, or `.gemini/workflows` directories without official support.
- Do not reintroduce the retired LazyAI runtime workflow/task/orchestrator system.
- Do not claim plugin/extension support is equivalent to native workflows unless the plugin actually provides executable workflow commands/tools.

## Acceptance criteria

- [ ] Add a documented workflow delivery matrix to code/docs.
- [ ] Update `docs/concepts/tools.md` to include all seven targets and the workflow delivery behavior for each.
- [ ] Workflow catalog remains canonical under `packages/cli/library/workflows/`.
- [ ] Unsupported native workflow dirs are asserted absent in adapter tests.
- [ ] If native Claude workflow support is implemented, it is gated by a distinct capability and tests.
- [ ] If plugin/extension workflow support is implemented, it is tracked separately from native workflows.
