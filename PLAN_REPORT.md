# LazyAI / vibe-lab Implementation Plan

## 1. What I found

- **Compile Contract & Adapter Support**: LazyAI reliably parses targets, tracks locks, and halts on drift. It supports 7 targets (`opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, `kiro`), with capabilities mapped per adapter.
- **Test/Conformance State**: `packages/cli/testdata/` is entirely missing. Golden outputs are absent; existing tests rely on inline assertions and `fstest.MapFS`.
- **Docs/Spec Status Drift**: Status markers are inconsistent (e.g. `KNOWLEDGE_MAP.md` claims `specs/029-lazyai-v2/` is `Done`, but `specs/029-lazyai-v2/spec.md` states `Draft`).
- **Validation**: `contract_validator.go` checks structure well, but `skill_validate.go` and `agent_validate.go` lack semantic/quality linting for missing instructions, misuse guidance, or required tools.

## 2. What is missing

**Priority 1**: Adapter conformance fixtures (`packages/cli/testdata/`) and golden testing.
**Priority 2**: Consolidated harness principles document (`docs/concepts/harness-principles.md`).
**Priority 3**: Status drift resolution for spec 029.
**Priority 4**: Semantic skill validation (warnings for missing trigger guidance/evidence).
**Priority 5**: Agent role contract validation (warnings for missing roles/workflows).

## 3. What I will implement now

I will implement Priorities 1, 2, and 3 in this pass. Priority 1 is the most critical as it introduces the `testdata/` structure for golden tests and conformance scenarios. Priority 2 and 3 involve document creation/updates.

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
| Fixtures | `packages/cli/testdata/projects/*`, `packages/cli/testdata/golden/*`, `packages/cli/internal/compiler/golden_test.go` |
| Docs | `docs/concepts/harness-principles.md`, `README.md`, `packages/cli/KNOWLEDGE_MAP.md`, `specs/029-lazyai-v2/spec.md` |

## 6. Tests to add/update

| Area | Tests |
|---|---|
| Conformance | Golden tests for manifest resolution, target outputs, drift refusal, and adapter specifics. |

## 7. Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Breaking existing tests | Low | Fixtures will be additive; existing inline tests will remain intact initially. |
| Incomplete golden coverage | Medium | Cover the core 14 fixture scenarios required by the backlog first. |

## 8. Validation commands

```bash
cd packages/cli && go test ./...
cd packages/cli && go vet ./...
```

## 9. Implementation sequence

1. **Conformance fixtures (Priority 1)**: Create the `testdata/` folder structure, populate with minimal manifests, and write the Go test runner for golden outputs.
2. **Harness/boundary docs (Priority 2)**: Draft `docs/concepts/harness-principles.md` and link from `README.md`.
3. **Status drift cleanup (Priority 3)**: Unify `029-lazyai-v2` status.