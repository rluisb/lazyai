# Implementation Scope — Cycle 4

## Priority A — Trace taxonomy

Add a file such as:

```text
packages/cli/library/templates/trace-taxonomy.md
docs/concepts/trace-eval-improvement-loop.md
```

Use taxonomy:

```yaml
context:
  - missing_project_context
  - stale_memory
  - bad_retrieval
  - ignored_user_constraint

tooling:
  - wrong_tool
  - missing_tool
  - unsafe_tool_use
  - unhandled_tool_error

workflow:
  - skipped_research
  - skipped_plan
  - skipped_tests
  - weak_handoff
  - missing_human_gate

quality:
  - hallucinated_api
  - overbroad_change
  - style_mismatch
  - incomplete_fix
  - no_evidence

adapter:
  - generated_output_drift
  - unsupported_surface_feature
  - broken_mcp_mapping
  - hook_not_emitted
```

Acceptance:

```text
- Trace taxonomy is file-based and inspectable.
- No trace daemon is added.
```

## Priority B — Harness improvement templates

Add templates:

```text
packages/cli/library/templates/trace-failure.md
packages/cli/library/templates/harness-change-report.md
packages/cli/library/templates/eval-promotion-checklist.md
packages/cli/library/templates/holdout-review.md
packages/cli/library/templates/evidence-report.md
```

Workflow:

```text
trace/failure
→ tagged eval case
→ one targeted harness change
→ holdout check
→ human review
→ promote asset update
```

Acceptance:

```text
- No judge/scoring runtime.
- No orchestration.
- Templates are clear and actionable.
```

## Priority C — Rubric assets

Add first-class rubrics or rubric templates.

Preferred if repo conventions allow:

```text
packages/cli/library/rubrics/
  reviewer.rubric.yaml
  evidence-verifier.rubric.yaml
  planner.rubric.yaml
  implementer.rubric.yaml
  skill-quality.rubric.yaml
  hook-quality.rubric.yaml
```

Alternative:

```text
packages/cli/library/templates/rubrics/
```

Rubrics should cover:

```text
correctness
evidence
test/lint/build verification
security
minimality
project convention fit
human gate behavior
handoff quality
uncertainty handling
```

Acceptance:

```text
- Rubrics are discoverable.
- Existing eval/rubric validation passes.
- No LLM-as-judge runtime is introduced.
```

## Priority D — Docs and manifest integration

If new templates/assets must be added to curation/provenance, do so according to repo conventions.

Acceptance:

```text
- Library manifest remains valid.
- Docs link back to harness principles.
```
