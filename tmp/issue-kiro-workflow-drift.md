## Summary

Kiro installs do not show workflows because LazyAI currently does not emit workflow assets for Kiro. Deeper inspection found a more important Kiro adapter drift: Kiro capabilities declare Specs/Steering/Hook surfaces and intentionally omit agents/skills, while the Kiro adapter actually installs agents/skills/prompts and no specs/steering.

## Evidence

Current Kiro install behavior:

- `packages/cli/internal/adapter/kiro.go`
  - creates `.kiro/agents`, `.kiro/prompts`, `.kiro/skills`
  - copies canonical agents to `.kiro/agents/<name>.md`
  - copies skills to `.kiro/skills/<name>/SKILL.md`
  - copies prompts to `.kiro/prompts/<name>.md`
- `packages/cli/internal/adapter/adapter_adapters_test.go:212-234`
  - test explicitly expects `.kiro/agents/guide.md`, `.kiro/agents/reviewer.md`, and `.kiro/skills/*/SKILL.md`

Declared Kiro capabilities:

- `packages/cli/internal/adapter/capabilities.go:161-174`
  - comments say Kiro surfaces are `steering`, `specs`, `hooks`, `MCP`, permissions, global steering
  - comments say it deliberately omits agents and skills to avoid unsupported `.kiro/agents`
- `packages/cli/internal/adapter/capabilities_test.go:63-73`
  - test says Kiro must declare Specs and Steering and must not declare Agents or Skills

Workflow catalog state:

- `packages/cli/library/workflows/*.md` exists and is embedded by `packages/cli/library/embed.go`
- `packages/cli/library/manifests/curation.yaml:1011-1098` marks workflows as `adapter_targets: [none]` and docs-only
- No `.kiro/workflows` emission exists, and there is no `workflow` asset kind in `packages/cli/internal/adapter/output_mapping.go`

External/native format finding:

- Kiro docs expose Specs, Steering, Hooks, and MCP as native concepts, not a separate `.kiro/workflows` directory.
- Relevant docs:
  - https://kiro.dev/docs/specs/
  - https://kiro.dev/docs/steering/
  - https://kiro.dev/docs/getting-started/first-project/

## Problem

LazyAI's Kiro adapter contract is internally inconsistent:

1. Capabilities and tests say Kiro should use Specs/Steering and not Agents/Skills.
2. Installer and install tests emit Agents/Skills/Prompts.
3. Workflow guidance has no Kiro-native projection.

This explains why a Kiro user does not see workflows, but the root issue is broader: LazyAI needs to map vibe-lab workflow intent into Kiro-native Specs/Steering rather than inventing `.kiro/workflows`.

## Proposed direction

For Kiro, do not emit `.kiro/workflows`.

Instead:

- Map reusable workflow/process guidance into `.kiro/steering/`.
- Map workflow templates/scenario flows into `.kiro/specs/` templates or examples.
- Keep hooks and MCP in Kiro-native locations.
- Reconcile whether Kiro actually supports `.kiro/agents` and `.kiro/skills`; if unsupported, stop emitting them and update tests.

Potential target shape, subject to docs verification:

```text
.kiro/
  steering/
    lazyai-workflows.md
    rpi.md
    bugfix.md
  specs/
    templates/
      feature/
      bugfix/
      verified-research/
  hooks/
  settings/mcp.json
```

## Acceptance criteria

- [ ] Resolve Kiro adapter contract drift: capabilities, install behavior, docs, and tests agree.
- [ ] Kiro install emits native Kiro Specs/Steering artifacts for LazyAI workflow guidance where supported.
- [ ] Kiro install does **not** emit `.kiro/workflows` unless official Kiro docs add that native surface.
- [ ] Tests cover the expected Kiro paths and negative path `.kiro/workflows`.
- [ ] `docs/concepts/tools.md` includes Kiro and describes Specs/Steering behavior accurately.
- [ ] Migration note documents any removal/relocation of `.kiro/agents` or `.kiro/skills` if those are no longer emitted.
