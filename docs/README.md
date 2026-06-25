# LazyAI + vibe-lab — Full Product Spec Pack

Generated: 2026-06-21

This pack turns the LazyAI/vibe-lab concept into a full product PRD and technical specification. It is designed as a product-wide foundation, not a feature spec.

## Files

- `01_PRD.md` — complete product requirements document.
- `02_TECHSPEC.md` — architecture, schemas, compiler, adapters, validation, testing.
- `03_OFFICIAL_TOOL_COMPLIANCE_MATRIX.md` — official docs to adapter requirements.
- `04_SCHEMAS_AND_EXAMPLES.md` — canonical schema and generated-output examples.
- `05_IMPLEMENTATION_ROADMAP.md` — milestone plan and exit criteria.
- `SOURCES.md` — official docs source registry.

## Core product stance

LazyAI is the harness asset manager and compiler.

vibe-lab is the quality/taste/process layer.

Host tools execute.

Optional runtime-adjacent features remain optional.

## Key compliance decisions

- Claude Code adapter must emit `CLAUDE.md`; current `AGENTS.md` alone is not enough.
- OpenCode adapter must use current permissions and `steps`, not deprecated fields.
- Copilot adapter must support `.github/copilot-instructions.md`, path instructions, skills, MCP, hooks, and optional plugin packaging.
- Pi adapter must respect project trust and explicitly warn that Pi has no built-in sandbox.
- Kiro adapter emits native `.kiro/hooks/*.json` (Kiro v3 hook schema); specs and steering remain absent (user-authored / non-goal), repo-local permissions are forbidden, and direct `.kiro/powers/` is not emitted (Powers is a future importable-package direction only).
- Antigravity adapter should use `.agents/skills` and `.agents/rules` as primary outputs, with global settings changes only by explicit consent.
- OMP adapter should support `.omp/mcp.json`, plugins, skills, commands, hooks, compaction, and handoff; stable since 2026-06-23 (all emitted surfaces verified against authoritative OMP docs, #486). `.omp/tools/` and `.omp/extensions/` are documented OMP-native discovery surfaces but are not emitted by LazyAI — executable-module generation is out of current product scope (user-managed; #524).