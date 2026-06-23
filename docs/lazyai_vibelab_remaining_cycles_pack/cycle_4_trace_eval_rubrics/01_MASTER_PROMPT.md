# RPI Cycle 4 — Trace/eval workflow templates and rubric assets

You are running locally inside the LazyAI repository.

Repo context:

```text
Repo root: /Users/ricardo/projects/teachable/lazyai
Product: lazyai-cli, a Go CLI harness asset manager/compiler.
LazyAI owns: .ai source, manifests, schemas, validation, adapter compilation, generated output.
vibe-lab owns: process/quality/taste assets, RPI workflow, human gates, skills, hooks, rubrics, handoff discipline.
Host tools execute generated assets.
LazyAI must not become a runtime or orchestrator.
```

Cycle goal:

```text
Add trace/eval improvement-loop templates and first-class rubric assets without adding a judge, scorer, daemon, or runtime.
```

Use RPI:

```text
R = Research the current implementation before changing anything.
P = Plan a small, safe, non-breaking implementation.
I = Implement with tests and clear final reporting.
```

---

# Shared LazyAI / vibe-lab Boundaries

These apply to every remaining RPI cycle.

## Product boundary

LazyAI must remain:

```text
- a canonical .ai source manager
- a validation layer
- a manifest/schema/lockfile owner
- a compiler into host-tool-native AI harness surfaces
- an optional runtime-adjacent state manager
```

vibe-lab remains:

```text
- the process and quality layer
- the source of RPI workflow guidance
- the source of human gates, anti-slop rules, skills, hooks, rubrics, and handoff discipline
```

Host tools execute generated assets:

```text
OpenCode
Claude Code
GitHub Copilot
Pi
OMP
Gemini / Antigravity
Kiro
```

## Do not add or reintroduce

```text
lazyai run workflow
lazyai autonomous task
lazyai orchestrate
lazyai judge
lazyai eval runner
mandatory trace daemon
mandatory RAG core
background subagent runtime
LangChain/LangGraph/CrewAI dependency in core
old task/workflow/orchestration/eval commands
Codex adapter
```

## Safety checks before editing

```bash
pwd
git rev-parse --show-toplevel
git status --short
git branch --show-current
git log -1 --oneline
```

## Never run

```bash
rm -rf
git clean
git reset
git push
git commit
npm publish
go install
curl | sh
brew install
docker compose up
terraform apply
kubectl apply
lazyai update
lazyai eject
lazyai compile --write
lazyai init --force
```

## Final boundary confirmation required

Every cycle report must confirm:

```text
- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```


---

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


---

# Plan Phase — Cycle 4

Create a small plan.

Include:

```text
1. Existing eval/rubric support.
2. Missing trace taxonomy.
3. Templates to add.
4. Rubric assets to add.
5. Whether any schema updates are necessary.
6. Tests to run.
7. Boundary risks.
```

Do not implement a scoring engine.


---

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


---

# Required validation

Run targeted tests for the changed area first.

Then run if safe:

```bash
go test ./packages/cli/...
go test ./...
go vet ./...
```

If broad tests fail for unrelated existing reasons, report the exact failure and whether targeted tests passed.

Check:

```bash
git status --short
```

---

# Final Report Template

Use this final report format for the cycle.

```markdown
# LazyAI / vibe-lab RPI Cycle <N> Report — <Title>

## 1. Research summary

- Repo state:
- Existing implementation confirmed:
- Missing pieces confirmed:
- Important constraints:

## 2. Plan executed

| Item | Status | Notes |
|---|---|---|

## 3. Changes made

| Area | Files changed | Summary |
|---|---|---|

## 4. Details by priority

### Priority A

- What changed:
- Tests:
- Remaining gaps:

### Priority B

- What changed:
- Tests:
- Remaining gaps:

### Priority C

- What changed:
- Tests:
- Remaining gaps:

## 5. Tests and validation

Commands run:

```bash
...
```

Results:

```text
...
```

## 6. Changed files

```text
...
```

## 7. Risks

| Risk | Impact | Mitigation |
|---|---|---|

## 8. Remaining work

| Item | Why not completed | Suggested next step |
|---|---|---|

## 9. Product boundary confirmation

- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```
