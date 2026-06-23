# Research Phase — Cycle 4

Inspect current eval, trace, rubric, runtime-adjacent, and template support.

## Required inspection

```text
packages/cli/internal/evals/
packages/cli/internal/schema/
packages/cli/library/templates/
packages/cli/library/skills/
packages/cli/library/rules/
packages/cli/internal/runtime/
packages/cli/internal/db/
packages/cli/cmd/session*
packages/cli/cmd/ledger*
packages/cli/cmd/memory*
packages/cli/cmd/metrics*
packages/cli/cmd/cost*
docs/concepts/harness-principles.md
docs/concepts/trace-evidence-over-vibes.md
```

Answer:

```text
- Are eval cases supported?
- Are holdouts supported?
- Are rubrics supported?
- Are there schemas?
- Is there a trace taxonomy?
- Is there a harness-change report template?
- Is there LLM-as-judge behavior?
- Are session/ledger/memory/metrics/cost optional?
- Which docs already explain trace evidence?
```
