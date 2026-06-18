# TASKS: LazyAI Internal Sidecar

**Version:** 1.0.0
**Status:** Draft
**Author:** turbo-crank (Fortnite multi-agent system)
**Date:** 2026-05-24

---

## Task Ordering

Tasks are ordered by dependency. Each task must be completed before the next begins. Tasks marked `[PARALLEL]` can be executed concurrently with their sibling.

---

### Task 1: Define Sidecar Types and Data Structures

**Priority:** High
**Risk:** Low — pure data structures, no logic.
**Depends on:** Nothing.

**Done when:**
- [ ] `packages/cli/internal/sidecar/types.go` exists with `SidecarConfig`, `LinkedProject`, `ResolvedPaths`, and `Scope` types.
- [ ] All struct fields have correct YAML tags (`yaml:"...,omitempty"` where appropriate).
- [ ] `WorkspaceEntry` in `packages/cli/cmd/workspace.go` has new `Sidecar *sidecar.SidecarConfig` field with `yaml:"sidecar,omitempty"`.
- [ ] Code compiles (`go build ./...` in `packages/cli`).
- [ ] Existing tests pass (backward compat check).

**Files:**
- `packages/cli/internal/sidecar/types.go` (new)
- `packages/cli/cmd/workspace.go` (modify — add field)

---

### Task 2: Implement Sidecar Loaders

**Priority:** High
**Risk:** Low — file I/O with well-defined error handling.
**Depends on:** Task 1.

**Done when:**
- [ ] `packages/cli/internal/sidecar/loader.go` exists with three functions:
  - `LoadWorkspaceSidecar() (*SidecarConfig, error)` — reads active workspace's sidecar block from `workspaces.yaml`.
  - `LoadProjectSidecar(projectRoot string) (*SidecarConfig, error)` — reads `.lazyai-sidecar.yaml`.
  - `LoadGlobalSidecar() (*SidecarConfig, error)` — reads `~/.lazyai/sidecar.yaml`.
- [ ] Missing files return `nil, nil` (not an error).
- [ ] Malformed YAML returns a parse error with file path.
- [ ] Unit tests cover: valid config, missing file, malformed YAML, workspace without sidecar block.

**Files:**
- `packages/cli/internal/sidecar/loader.go` (new)
- `packages/cli/internal/sidecar/loader_test.go` (new)

---

### Task 3: Implement Sidecar Resolver

**Priority:** High
**Risk:** Medium — resolution priority logic must be correct.
**Depends on:** Task 2.

**Done when:**
- [ ] `packages/cli/internal/sidecar/resolver.go` exists with `Resolve(scope Scope, projectRoot string) ResolvedPaths`.
- [ ] Resolution follows the priority chain: workspace > project > global > default.
- [ ] Relative paths are resolved against correct anchors:
  - Workspace: relative to `~/.lazyai/`.
  - Project: relative to `projectRoot`.
  - Global: relative to `~/.lazyai/`.
- [ ] Absolute paths pass through unchanged.
- [ ] Default `*_dir` values (`"specs"`, `"docs"`, `"plans"`) applied when fields are empty.
- [ ] `linked_projects` merged with workspace > project > global priority.
- [ ] `ConfigLevel` field correctly set to the winning level.
- [ ] Unit tests cover all priority combinations (at least 8 test cases).

**Files:**
- `packages/cli/internal/sidecar/resolver.go` (new)
- `packages/cli/internal/sidecar/resolver_test.go` (new)

---

### Task 4: Implement Sidecar Config Writers

**Priority:** High
**Risk:** Medium — must not corrupt existing config files.
**Depends on:** Task 1.

**Done when:**
- [ ] `packages/cli/internal/sidecar/writer.go` exists with three functions:
  - `WriteWorkspaceSidecar(cfg *SidecarConfig) error` — adds/updates sidecar block on active workspace entry.
  - `WriteProjectSidecar(projectRoot string, cfg *SidecarConfig) error` — writes `.lazyai-sidecar.yaml`.
  - `WriteGlobalSidecar(cfg *SidecarConfig) error` — writes `~/.lazyai/sidecar.yaml`.
- [ ] Workspace writer preserves other workspace entries and the `active` field.
- [ ] Project writer creates parent directories if needed.
- [ ] All writers create backup (`.bak`) before overwriting.
- [ ] Unit tests verify: new creation, update existing, other entries preserved.

**Files:**
- `packages/cli/internal/sidecar/writer.go` (new)
- `packages/cli/internal/sidecar/writer_test.go` (new)

---

### Task 5: Implement Sidecar Config Remover

**Priority:** Medium
**Risk:** Low — straightforward delete/update operations.
**Depends on:** Task 1.

**Done when:**
- [ ] `packages/cli/internal/sidecar/remover.go` exists with two functions:
  - `RemoveWorkspaceSidecar() error` — removes sidecar block from active workspace entry.
  - `RemoveProjectSidecar(projectRoot string) error` — deletes `.lazyai-sidecar.yaml`.
- [ ] Workspace remover preserves other entries and `active` field.
- [ ] Project remover is a no-op if file doesn't exist (returns nil).
- [ ] Unit tests cover: remove existing, remove non-existent, other entries preserved.

**Files:**
- `packages/cli/internal/sidecar/remover.go` (new)
- `packages/cli/internal/sidecar/remover_test.go` (new)

---

### Task 6: Implement Sidecar Doctor

**Priority:** Medium
**Risk:** Low — read-only validation.
**Depends on:** Task 2, Task 3.

**Done when:**
- [ ] `packages/cli/internal/sidecar/doctor.go` exists with `Doctor(scope Scope, projectRoot string) []Issue`.
- [ ] `Issue` struct has `Severity`, `Message`, and `Path` fields.
- [ ] Checks performed:
  - Sidecar `path` exists and is a directory.
  - Sidecar `path` is writable.
  - Each `*_dir` resolves to an existing or creatable path.
  - `linked_projects` paths exist.
- [ ] Returns empty slice when all checks pass.
- [ ] Unit tests cover: all valid, missing path, file-not-directory, non-writable, broken linked_projects.

**Files:**
- `packages/cli/internal/sidecar/doctor.go` (new)
- `packages/cli/internal/sidecar/doctor_test.go` (new)

---

### Task 7: Wire Sidecar Commands to CLI

**Priority:** High
**Risk:** Medium — must integrate cleanly with existing cobra command tree.
**Depends on:** Tasks 3, 4, 5, 6.

**Done when:**
- [ ] `packages/cli/cmd/sidecar.go` exists with `sidecarCmd` cobra command group.
- [ ] Five subcommands registered: `init`, `status`, `attach`, `detach`, `doctor`.
- [ ] Each subcommand calls the appropriate internal/sidecar functions.
- [ ] `sidecar init` flags: `--scope`, `--path`, `--specs-dir`, `--docs-dir`, `--plans-dir`.
- [ ] `sidecar attach` flags: `--scope`, `--path`, `--specs-dir`, `--docs-dir`, `--plans-dir`.
- [ ] `sidecar detach` flags: `--scope`, `--force`.
- [ ] `sidecar doctor` flags: `--scope`.
- [ ] `sidecar status` has no required flags.
- [ ] `sidecarCmd` registered on `rootCmd` in `init()`.
- [ ] All commands have help text.
- [ ] Code compiles and `lazyai-cli sidecar --help` works.

**Files:**
- `packages/cli/cmd/sidecar.go` (new)

---

### Task 8: Integration Tests

**Priority:** Medium
**Risk:** Medium — tests must set up and tear down config state.
**Depends on:** Task 7.

**Done when:**
- [ ] Integration test file exists (e.g., `packages/cli/cmd/sidecar_test.go` or a separate test harness).
- [ ] Test: `sidecar init --scope workspace` creates valid config.
- [ ] Test: `sidecar init --scope project` creates `.lazyai-sidecar.yaml`.
- [ ] Test: `sidecar init --scope global` creates `~/.lazyai/sidecar.yaml`.
- [ ] Test: `sidecar status` output format is correct.
- [ ] Test: `sidecar attach` + `sidecar detach` round-trip.
- [ ] Test: `sidecar doctor` exit codes (0 for clean, non-zero for issues).
- [ ] Test: backward compat — existing `workspace list/add/switch/status` still work.
- [ ] All tests use temporary directories (no pollution of real `~/.lazyai/`).

**Files:**
- `packages/cli/cmd/sidecar_test.go` (new)

---

### Task 9: Documentation

**Priority:** Low
**Risk:** Low — documentation only.
**Depends on:** Task 7.

**Done when:**
- [ ] README or docs updated with sidecar usage examples.
- [ ] Help text on all five commands is clear and complete.
- [ ] Edge cases documented in command help or a troubleshooting section.

**Files:**
- `README.md` (modify)
- `packages/cli/cmd/sidecar.go` (modify — ensure help text quality)

---

## Acceptance Criteria (Cross-Task)

- [ ] AC1: Workspace YAML with no sidecar block parses without error (Task 1, Task 2).
- [ ] AC2: `sidecar status` shows correct resolved paths (Task 3, Task 7).
- [ ] AC3: Resolution priority is correct (Task 3).
- [ ] AC4: All five commands exist and behave as specified (Task 7).
- [ ] AC5: `sidecar doctor` catches all documented edge cases (Task 6).
- [ ] AC6: No sidecar configured = behavior identical to current release (Task 8).
- [ ] AC7: No Skeeper or provider fields anywhere (all tasks — verify in review).

---

## Risk Summary

| Task | Risk | Mitigation |
|---|---|---|
| Task 3 (Resolver) | Medium — priority logic bugs | Comprehensive unit tests for all combinations |
| Task 4 (Writer) | Medium — config corruption | Backup before write; unit tests verify preservation |
| Task 7 (CLI wiring) | Medium — integration surface | Integration tests cover all commands |
| Task 8 (Integration) | Medium — test isolation | Use temp directories, never touch real config |