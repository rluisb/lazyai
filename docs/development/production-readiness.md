# Production Readiness and Project Quality

This page records the current production-readiness posture for LazyAI docs and setup outputs. It separates active product surfaces from historical or generated material so maintainers know what to ship, document, archive, or ignore.

## Current readiness snapshot

| Area | Status | Evidence |
|---|---|---|
| Supported AI CLI targets | Ready | Adapter registry includes OpenCode, Claude Code, Copilot, Pi, OMP, Kiro, and Antigravity. |
| Stable adapters | Ready | All seven adapters (OpenCode, Claude Code, Copilot, Pi, Kiro, OMP, Antigravity) are marked stable in adapter capabilities. OMP and Antigravity were promoted from beta on 2026-06-23 after docs verification (#486). |
| Beta adapters | Ready | None. The last beta adapter (Antigravity/Gemini) was promoted to stable on 2026-06-23 once its two gaps — global-scope skills path and root-instructions discovery — were closed and pinned by conformance tests (#486). |
| Output mapping | Ready | `output_mapping.go` is the single source of truth for seven asset kinds across seven tools. |
| Manifest contract | Ready | `.ai/lazyai.json` schema version is `1.0`; target enum is `opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, `kiro`. |
| MCP compile | Mostly ready | OpenCode, Claude Code, Copilot, OMP, Kiro, and Antigravity emit configs; Pi is intentionally no-op. |
| Docs information architecture | Improved | MkDocs now has an AI CLI Tools section with one page per supported target. |

## Tracked watchouts

- Beta adapter exit criteria tracked in [#486](https://github.com/rluisb/lazyai/issues/486) are fully cleared: OMP and Antigravity are both stable as of 2026-06-23 (Antigravity's global-skills path and root-instructions discovery gaps are closed and pinned by tests). No adapter remains below stable. See [adapter snapshot](../adapters/snapshots/beta-adapter-verification-2026-06.md).
- Workflow helper ownership is tracked in [#487](https://github.com/rluisb/lazyai/issues/487): ADR-007 must be accepted or superseded before LazyAI emits workflow helpers, and LazyAI must not become a workflow runtime.

## Quality gates to run before release

```bash
# CLI package tests
cd packages/cli
go test ./...

# Documentation build
cd ../..
mkdocs build --strict
```

Run targeted tests for adapter changes first, then the full package suite before release.

## Stale or historical docs inventory

These files are excluded from MkDocs or absent from navigation, but they still live under `docs/`. Do not use them as source-of-truth product docs.

| Path | Status | Recommended action |
|---|---|---|
| `docs/orchestrator-design.md` | retired orchestrator runtime | archive or remove after human approval |
| `docs/orchestrator-research.md` | retired orchestrator runtime | archive or remove after human approval |
| `docs/orchestrator-blueprint.md` | retired orchestrator runtime | archive or remove after human approval |
| `docs/orchestration-usage.md` | redirect/stub for retired runtime | remove after human approval |
| `docs/AI-Agentic-Setup-Implementation-Plan.md` | predecessor ai-setup material | archive or remove after human approval |
| `docs/AI-Agentic-Setup-Playbook.md` | predecessor ai-setup material | archive or remove after human approval |
| `docs/AI-Agentic-Setup-Templates/` | predecessor template tree | archive or remove after human approval |
| `docs/docs/` | duplicate product spec pack | removed — canonical spec pack at `docs/lazyai-vibelab-product-spec-pack/` |
| `docs/lazyai-vibelab-product-spec-pack.zip` | zip beside extracted content | remove after human approval |
| `docs/lazyai_vibelab_remaining_cycles_pack/*.zip` | zips beside extracted cycle packs | remove after human approval |
| `docs/reports/` | historical adversarial-review reports | keep as evidence or move to archive |
| `docs/wiki/` | manually-synced GitHub Wiki bootstrap | keep only if wiki sync remains intentional |

## Documentation source-of-truth rules

- Current user-facing setup docs live under `docs/ai-cli-tools/`, `docs/getting-started/`, `docs/concepts/`, `docs/reference/`, and `docs/adapters/`.
- Historical RPI packs and retired orchestrator docs are not production guidance.
- Generated build output such as `site/` must not be treated as source documentation.
- Adapter behavior claims should be checked against `packages/cli/internal/adapter/registry.go`, `capabilities.go`, `output_mapping.go`, and `mcp_compiler.go`.
