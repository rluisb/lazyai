# RPI Cycle 6 — Context/handoff, multi-agent boundary, init headless clarity, and server L3 clarity

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
Complete the remaining guidance layer: context compaction, handoff, multi-agent boundary, init populate clarity, and server capability levels.
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

# Research Phase — Cycle 6

Inspect context, handoff, multi-agent, init, and server behavior.

## Required inspection

```text
packages/cli/library/fragments/
packages/cli/library/templates/
packages/cli/library/skills/handoff*
packages/cli/library/skills/parallel-execution*
packages/cli/library/canonical/agents/
docs/concepts/
docs/workflows/
packages/cli/cmd/init.go
packages/cli/cmd/server.go
packages/cli/cmd/doctor.go
packages/cli/cmd/status.go
```

Answer:

```text
- What context/handoff templates already exist?
- Is context compaction event-driven in docs?
- What multi-agent guidance already exists?
- Does any doc imply LazyAI orchestrates agents?
- How does init headless populate behave?
- How is server L1/L3 capability explained?
- Does doctor/status communicate limitations clearly?
```


---

# Plan Phase — Cycle 6

Create a small plan.

Include:

```text
1. Context/handoff docs/templates to add/update.
2. Multi-agent boundary docs/templates to add/update.
3. Init headless populate help/docs updates.
4. Server L1/L3 clarity updates.
5. Tests to add/update.
6. Boundary risks.
```

Do not add runtime behavior.


---

# Implementation Scope — Cycle 6

## Priority A — Context compaction and handoff templates

Add/update:

```text
packages/cli/library/fragments/context-compaction.md
packages/cli/library/templates/context-handoff.md
packages/cli/library/templates/session-compaction.md
packages/cli/library/templates/recovery-handoff.md
docs/concepts/context-discipline.md
```

Rules:

```text
Compact after:
- research phase
- planning phase
- implementation phase
- failed verification
- agent/tool handoff
- before context gets too large

Never compact away:
- user constraints
- decisions
- failed attempts
- commands run
- files touched
- unresolved risks
- human approval state
```

Acceptance:

```text
- Handoff templates require evidence.
- Context compaction is guidance/assets, not runtime automation.
```

## Priority B — Multi-agent boundary

Add/update:

```text
docs/concepts/multi-agent-boundary.md
packages/cli/library/templates/multi-agent-decision.md
packages/cli/library/fragments/specialist-agent-guidance.md
```

Guidance:

```text
Use one agent + skills when:
- task is local
- context fits
- one person/agent can reason through it
- no isolation needed

Use specialist agents when:
- role separation improves quality
- review should be independent
- domain expertise differs

Use parallel agents when:
- work is read-heavy
- outputs are easy to merge
- boundaries are explicit

Avoid multi-agent when:
- agents need the same full context
- parallel writes would conflict
- merge cost exceeds benefit
- host tool cannot show evidence
```

Acceptance:

```text
- No orchestration command.
- No runtime scheduler.
- Guidance compiles as assets/templates where appropriate.
```

## Priority C — Init headless populate clarity

Inspect and update:

```text
packages/cli/cmd/init.go
README.md
docs/commands/init.md if present
```

Clarify:

```text
- why AGENTS.md placeholder fill may be skipped
- what host tool should do next
- how to complete setup manually
- how to validate after populate
```

Acceptance:

```text
- No hidden runtime behavior added.
- CLI explains what happened and next steps.
```

## Priority D — Server L1/L3 capability clarity

Inspect and update:

```text
packages/cli/cmd/server.go
README.md
docs/commands/server.md if present
doctor/status messaging if relevant
```

Clarify:

```text
- L1 config check
- L3 handshake not implemented in Go
- Node.js MCP SDK dependency if still true
- current limitation vs future work
```

Acceptance:

```text
- Users can distinguish L1 config check vs L3 handshake.
- No fake support is claimed.
```

## Priority E — Tests

Add/update tests for changed help text or docs references if repo conventions include such tests.

Acceptance:

```text
- Tests pass.
- Product boundary remains intact.
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
