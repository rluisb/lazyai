# RPI Cycle 2.1 — Link semantic validation docs, curate templates, and inventory shipped asset warnings

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
Finish Cycle 2 by linking the new semantic validation docs, deciding template curation, and producing a warning inventory for shipped skills/agents.
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

# Research Phase — Cycle 2.1

Inspect the result of Cycle 2.

## Required inspection

```text
docs/concepts/harness-principles.md
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
packages/cli/library/manifests/curation.yaml
packages/cli/library/manifests/provenance.yaml
packages/cli/internal/validate/validate.go
packages/cli/internal/validate/validate_test.go
README.md
specs/KNOWLEDGE_MAP.md
```

Answer:

```text
- Are skill-quality and agent-contract docs linked from harness-principles?
- Are the new templates referenced by curation/provenance if repo conventions require it?
- Does validate output expose semantic warnings clearly?
- What warnings appear when validating shipped library assets?
- Are any warnings accidental false positives?
- Are any docs duplicated or conflicting?
```

Run if safe:

```bash
go test -v ./packages/cli/internal/validate/...
go test ./packages/cli/...
```


---

# Plan Phase — Cycle 2.1

Produce a small plan before editing.

Include:

```text
1. Which docs need links.
2. Whether templates should be included in curation/provenance now.
3. How to capture warning inventory.
4. Whether any warning wording needs refinement.
5. Files to change.
6. Tests to run.
```

Keep this cycle small. Do not rewrite all shipped skills.


---

# Implementation Scope — Cycle 2.1

## Priority A — Link semantic validation docs

Add links from:

```text
docs/concepts/harness-principles.md
README.md if appropriate
specs/KNOWLEDGE_MAP.md if appropriate
```

To:

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
```

Acceptance:

```text
- Docs are discoverable.
- No duplicate/conflicting doctrine.
- LazyAI still described as compiler/asset manager.
```

## Priority B — Curate templates if appropriate

Inspect existing curation rules before editing.

If project conventions require templates in curation/provenance, add:

```text
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

Acceptance:

```text
- Manifest validation still passes.
- Templates are discoverable by library tooling.
```

## Priority C — Warning inventory

Run validation against shipped assets if there is an existing safe way.

Capture a warning inventory such as:

```text
rule id
asset path
count
likely true positive / likely false positive
recommended future cleanup
```

Store it in one of:

```text
docs/reports/semantic-validation-warning-inventory.md
specs/029-lazyai-v2/semantic-validation-warning-inventory.md
```

Follow existing repo conventions.

Acceptance:

```text
- No mass asset rewrite.
- Warning inventory gives future cleanup path.
- False positives are noted.
```

## Priority D — Minor wording refinements

Only adjust semantic validation wording if needed for clarity.

Acceptance:

```text
- Rule IDs stay stable.
- Tests still pass.
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
