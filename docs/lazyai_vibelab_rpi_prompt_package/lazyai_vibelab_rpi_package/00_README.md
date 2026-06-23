# LazyAI / vibe-lab RPI Prompt Package

This package contains ready-to-copy prompts and checklists for local AI agents working inside the LazyAI repository.

The goal is to help local agents use RPI:

```text
R = Research the current implementation carefully.
P = Plan the safest implementation sequence.
I = Implement the missing improvements in small, reviewable, test-backed changes.
```

## Files

| File | Purpose |
|---|---|
| `01_MASTER_RPI_PROMPT.md` | Complete one-shot prompt for local agents. Use this when you want a single comprehensive instruction. |
| `02_AGENT_BOUNDARIES_AND_SAFETY.md` | Non-negotiable product boundaries and safety rules. Useful as a system/developer-style preamble. |
| `03_RESEARCH_PHASE_PROMPT.md` | Research-only prompt for a local agent to inspect the repo before changes. |
| `04_PLAN_PHASE_PROMPT.md` | Planning-only prompt after research is done. |
| `05_IMPLEMENTATION_BACKLOG_ORDERED.md` | Ordered backlog of all missing improvements, with acceptance criteria. |
| `06_FINAL_REPORT_TEMPLATE.md` | Required final implementation report template. |
| `07_SHORT_AGENT_PROMPT.md` | Compact version for tools with smaller context windows. |
| `08_PR_SEQUENCE.md` | Suggested PR split and implementation sequencing. |
| `lazyai_vibelab_rpi_prompt_manifest.json` | Machine-readable summary of the package. |

## Recommended usage

Start with:

```text
01_MASTER_RPI_PROMPT.md
```

For long-running work, split the work into phases:

```text
03_RESEARCH_PHASE_PROMPT.md
04_PLAN_PHASE_PROMPT.md
05_IMPLEMENTATION_BACKLOG_ORDERED.md
06_FINAL_REPORT_TEMPLATE.md
```

Keep `02_AGENT_BOUNDARIES_AND_SAFETY.md` attached to every local agent session so the agent does not accidentally expand LazyAI into a runtime/orchestrator.
