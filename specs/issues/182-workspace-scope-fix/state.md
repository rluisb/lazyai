# State: 182-workspace-scope-fix

**Task ID**: T182
**Status**: DONE

**What was implemented**: Workspace scope now treats `TargetDir`/`PlanningRepoPath` as the planning repo and routes adapter/tool config installs through `WorkspaceRoot`. Global scope skips project-specific planning artifacts (`.specify/`, `specs/`, `KNOWLEDGE_MAP.md`, `CODEOWNERS`, compliance/spec scaffolding) while preserving global tool installs through `HomeDir`.

**Tests**: Targeted packages passed: `./cmd`, `./internal/scaffold`, `./internal/adapter`.

**Quality gates**: all applicable gates passed. Gate 5 not applicable (no new endpoints, background jobs, or external calls).

**Deviations from harness**: None.

**Patterns followed**: Reused existing `PlanningRepoPath`, adapter `WorkspaceRoot`, setup-scope routing, conflict strategy, and scaffold helper patterns already present in `cmd/helpers.go`, `internal/adapter/scope.go`, and `internal/scaffold/scaffold.go`.
