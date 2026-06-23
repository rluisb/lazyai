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
- Kiro adapter must prioritize `.kiro/steering`, `.kiro/specs`, hooks, MCP, protected paths, and warnings that Supervised mode is not a sandbox.
- Antigravity adapter should use `.agents/skills` and `.agents/rules` as primary outputs, with global settings changes only by explicit consent.
- OMP adapter should support `.omp/mcp.json`, plugins, skills, commands, hooks, compaction, and handoff; keep beta until docs snapshots are fully captured.
