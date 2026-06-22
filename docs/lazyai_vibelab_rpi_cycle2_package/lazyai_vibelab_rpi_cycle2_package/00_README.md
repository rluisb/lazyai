# LazyAI / vibe-lab RPI Cycle 2 Prompt Package

This package contains ready-to-copy prompts for local AI agents working on:

> **RPI Cycle 2 — Semantic validation for skills and agent contracts**

The goal is to improve LazyAI validation so skills and agents are not only structurally valid, but also behaviorally useful, scoped, safe, evidence-driven, and aligned with vibe-lab quality principles.

## Files

| File | Purpose |
|---|---|
| `01_MASTER_RPI_CYCLE2_PROMPT.md` | Complete prompt for the local AI agent. |
| `02_BOUNDARIES_AND_SAFETY.md` | Product boundaries, forbidden work, and safe command rules. |
| `03_RESEARCH_PHASE.md` | Focused research checklist for current validation, skills, agents, and tests. |
| `04_PLAN_PHASE.md` | Planning checklist before implementation. |
| `05_IMPLEMENTATION_SCOPE.md` | Detailed implementation scope for skill validation, agent validation, output quality, docs, and templates. |
| `06_VALIDATION_RULES.md` | Rule IDs, severity model, expected checks, and output examples. |
| `07_DOCS_AND_TEMPLATES.md` | Documentation and template requirements for future asset authors. |
| `08_TESTING_AND_FINAL_REPORT.md` | Required tests and final report format. |
| `09_SHORT_AGENT_PROMPT.md` | Short version for quick local-agent execution. |
| `10_PR_SEQUENCE.md` | Suggested PR breakdown for Cycle 2. |
| `lazyai_vibelab_rpi_cycle2_manifest.json` | Machine-readable package manifest. |

## Recommended use

Use `01_MASTER_RPI_CYCLE2_PROMPT.md` for the main local AI agent.

Use `09_SHORT_AGENT_PROMPT.md` when the agent already understands the repository and you only need to steer it into the Cycle 2 scope.

## Cycle 2 focus

- Research current skill and agent formats.
- Improve semantic validation.
- Add actionable rule IDs and fix suggestions.
- Add tests for valid and invalid skill/agent cases.
- Add docs/templates for future authors.
- Preserve LazyAI’s boundary as a compiler and harness asset manager.
