**Epic — RPI Cycle 4: Trace/eval templates & rubrics**

Goal: Add trace taxonomy, improvement-loop templates, and rubric assets — without adding a judge/scoring/runtime engine.

### Why
Eval/rubric schemas already exist (`internal/schema/eval-rubric.schema.json`, `eval-case.schema.json`, `eval-holdout.schema.json`) and `internal/evals/validate.go`. Missing: trace taxonomy, improvement-loop templates, and first-class rubric assets (`library/rubrics/` does not exist).

### Tasks (sub-issues)
- Add trace taxonomy asset (template + concept doc)
- Add harness improvement-loop templates
- Add rubric assets validated against existing eval-rubric schema
- Docs + manifest/curation integration

### Improvement loop
trace/failure -> tagged eval case -> one targeted harness change -> holdout check -> human review -> promote asset update.

### Boundary
File-based, inspectable. No trace daemon, no LLM-as-judge runtime, no scoring engine, no orchestration.
