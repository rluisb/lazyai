# LazyAI / vibe-lab — RPI Plan Phase Prompt

You have completed the Research phase inside the LazyAI repository.

Your job now is to produce a safe, ordered implementation plan. Do not edit files until the plan is complete.

## Required plan format

```markdown
# LazyAI / vibe-lab Implementation Plan

## 1. What I found

Summarize evidence from the research phase.

## 2. What is missing

List missing or partial pieces, grouped by priority.

## 3. What I will implement now

Be explicit about the scope of this pass.

## 4. What I will intentionally not implement

Must include:

- no runtime/orchestration surface
- no old task/workflow/eval command reintroduction
- no mandatory judge/scoring engine
- no mandatory trace daemon
- no mandatory RAG core
- no Codex adapter

## 5. Files likely to change

| Area | Likely files |
|---|---|

## 6. Tests to add/update

| Area | Tests |
|---|---|

## 7. Risks

| Risk | Impact | Mitigation |
|---|---|---|

## 8. Validation commands

```bash
go test ./...
go vet ./...
```

Add targeted tests for the changed areas.

## 9. Implementation sequence

Use the priority order from `05_IMPLEMENTATION_BACKLOG_ORDERED.md`.
```

## Planning rules

Prefer small, reviewable changes.

Suggested patch groups:

```text
1. Conformance fixtures
2. Harness/boundary docs
3. Status drift cleanup
4. Skill/agent validation
5. Hook lifecycle catalog
6. Trace/eval templates
7. Rubrics
8. MCP examples
9. Adapter docs
10. Context/multi-agent docs
```

Use existing conventions before adding new architecture.

Do not make breaking schema changes unless unavoidable.
