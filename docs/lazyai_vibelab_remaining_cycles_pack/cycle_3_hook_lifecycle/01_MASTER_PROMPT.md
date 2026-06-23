# RPI Cycle 3 — Hook lifecycle catalog and adapter capability mapping

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
Define a neutral hook lifecycle and map current hook assets honestly to each supported adapter.
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

# Research Phase — Cycle 3

Inspect current hook and adapter behavior.

## Required inspection

```text
packages/cli/library/hooks/
packages/cli/library/claudecode/
packages/cli/library/opencode/
packages/cli/library/antigravity/
packages/cli/library/cupcake/
packages/cli/internal/adapter/
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
docs/concepts/harness-principles.md
```

Answer:

```text
- What hook assets exist?
- Which hooks block actions?
- Which hooks require human approval?
- Which hooks capture evidence?
- Which adapters support executable hooks?
- Which adapters are instruction-only?
- Which adapters do not support hooks?
- Is there already a neutral hook lifecycle model?
```

Confirm current hook-related files emitted for:

```text
OpenCode
Claude Code
Gemini / Antigravity
Copilot
Pi
OMP
Kiro
```


---

# Plan Phase — Cycle 3

Create a small plan.

Include:

```text
1. Existing hook assets.
2. Proposed neutral lifecycle model.
3. Adapter capability mapping.
4. Whether implementation is docs-only, schema-backed, or code-backed.
5. Files to change.
6. Tests to add/update.
7. Compatibility risks.
```

Prefer docs + validation + capability mapping over runtime behavior.


---

# Implementation Scope — Cycle 3

## Priority A — Neutral hook lifecycle catalog

Define this lifecycle:

```text
before_agent
before_model
before_tool
after_tool
after_model
after_agent
on_error
on_compaction
on_handoff
on_human_gate
```

Add or update:

```text
packages/cli/library/hooks/catalog.md
docs/concepts/hook-lifecycle.md
```

If code support is appropriate:

```text
packages/cli/internal/hooks/lifecycle.go
packages/cli/internal/hooks/capabilities.go
```

Acceptance:

```text
- Lifecycle is documented.
- It does not imply LazyAI executes hooks itself.
- Host-tool execution boundary is explicit.
```

## Priority B — Hook capability matrix

Map support per adapter:

```text
supported
partial
instruction_only
unsupported
not_applicable
```

For each surface:

```text
opencode
claude
copilot
pi
omp
antigravity
kiro
```

Acceptance:

```text
- Weak surfaces use instruction-only fallbacks instead of fake hooks.
- Capability matrix matches adapter code.
```

## Priority C — Classify existing hooks

Classify:

```text
pre-commit
rpi-gate-check
caveman-memory-promotion
startup-self-heal
block-destructive-shell
objective-workflow-gate
```

For each hook record:

```yaml
name:
lifecycle:
purpose:
blocks_actions:
requires_human_approval:
captures_evidence:
surfaces:
```

Acceptance:

```text
- Every shipped hook has a lifecycle classification.
- Unsupported claims are not made.
```

## Priority D — Validation/tests

If validation exists for hook assets, add checks for:

```text
- unknown lifecycle
- unsupported adapter claim
- missing purpose
- missing safety behavior for blocking hooks
```

If validation does not exist yet, add docs and tests around capability mapping only.

Acceptance:

```text
- Tests pass.
- No runtime hook scheduler is added.
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
