# Production Readiness and Project Quality

This page records the current production-readiness posture for LazyAI docs and setup outputs. It separates active product surfaces from historical or generated material so maintainers know what to ship, document, archive, or ignore.

## Current readiness snapshot

| Area | Status | Evidence |
|---|---|---|
| Supported AI CLI targets | Ready | Adapter registry includes OpenCode, Claude Code, Copilot, Pi, OMP, Kiro, and Antigravity. |
| Stable adapters | Ready | OpenCode, Claude Code, Copilot, Pi, and Kiro are marked stable in adapter capabilities. |
| Beta adapters | Watch | OMP and Antigravity are functional but beta while official host docs are still being snapshot-verified. |
| Output mapping | Ready | `output_mapping.go` is the single source of truth for seven asset kinds across seven tools. |
| Manifest contract | Ready | `.ai/lazyai.json` schema version is `1.0`; target enum is `opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, `kiro`. |
| MCP compile | Mostly ready | OpenCode, Claude Code, Copilot, OMP, Kiro, and Antigravity emit configs; Pi is intentionally no-op. |
| Docs information architecture | Improved | MkDocs now has an AI CLI Tools section with one page per supported target. |

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
| `docs/docs/` | duplicate product spec pack | consolidate with canonical spec pack |
| `docs/lazyai-vibelab-product-spec-pack.zip` | zip beside extracted content | remove after human approval |
| `docs/lazyai_vibelab_remaining_cycles_pack/*.zip` | zips beside extracted cycle packs | remove after human approval |
| `docs/reports/` | historical adversarial-review reports | keep as evidence or move to archive |
| `docs/wiki/` | manually-synced GitHub Wiki bootstrap | keep only if wiki sync remains intentional |

## Documentation source-of-truth rules

- Current user-facing setup docs live under `docs/ai-cli-tools/`, `docs/getting-started/`, `docs/concepts/`, `docs/reference/`, and `docs/adapters/`.
- Historical RPI packs and retired orchestrator docs are not production guidance.
- Generated build output such as `site/` must not be treated as source documentation.
- Adapter behavior claims should be checked against `packages/cli/internal/adapter/registry.go`, `capabilities.go`, `output_mapping.go`, and `mcp_compiler.go`.
