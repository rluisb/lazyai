# TASKS: LazyAI Internal Sidecar

**Version:** 1.1.0
**Status:** Approved
**Author:** turbo-crank (Fortnite multi-agent system)
**Date:** 2026-06-17

---

## Task Ordering

This v1.1.0 remediation list replaces the original implementation tasks. The original feature is shipped; these tasks close residual defects and spec drift identified by the adversarial review. Tasks are ordered by dependency — each task must be completed and verified before the next begins.

---

### Task 1: Update Spec Package to v1.1.0 Approved

**Priority:** High
**Risk:** Low — documentation only.
**Depends on:** Nothing.

**Done when:**
- [x] `SPEC.md`, `TECH-SPEC.md`, `TASKS.md` have `Version: 1.1.0`, `Status: Approved`, `Date: 2026-06-17`.
- [x] R1 strengthens empty-path to an error (Resolve returns error, status non-zero, doctor ERROR).
- [x] R2 specifies whole-config replacement and level-specific relative-path anchors.
- [x] R3/R4 specify doctor scope-chain validation (global / project / workspace) and detach no-op exit 0.
- [x] R5 marks `linked_projects` as reserved/ignored; `.bak` is single-slot.
- [x] Durability policy section present.
- [x] TECH-SPEC architecture decision updated from "thin wrappers" to "boundary validation + delegate".
- [x] TECH-SPEC resolution algorithm uses "apply/replace" not "merge"; doctor flags are `workspace|project|global`.
- [x] Testing strategy removes linked-project tests and adds empty-path, anchor, scope-chain, atomic write, lock, root-error, and typo-prevention tests.

**Files:**
- `specs/001-internal-sidecar/SPEC.md` (modify)
- `specs/001-internal-sidecar/TECH-SPEC.md` (modify)
- `specs/001-internal-sidecar/TASKS.md` (modify)

---

### Task 2: Fix Resolver Validation and Level-Specific Relative-Path Anchors

**Priority:** High
**Risk:** Medium — changes resolution behavior for relative paths.
**Depends on:** Task 1.

**Done when:**
- [x] `resolveSidecar` returns `fmt.Errorf("sidecar path is required but empty")` when `cfg.Path == ""`.
- [x] `Resolve` computes anchors by config level: global → `globalDir`, project → `projectRoot`, workspace → `globalDir`.
- [x] `scopeAnchor` deleted if no remaining caller; otherwise renamed/behavior-aligned.
- [x] `ConfigLevel` values unchanged; whole-config replacement preserved.
- [x] Tests added in `resolver_test.go`: empty path rejection, global relative path at project scope, project relative path at workspace scope, workspace relative path at workspace scope, higher-priority sidecar replaces whole tuple.

**Files:**
- `packages/cli/internal/sidecar/resolver.go` (modify)
- `packages/cli/internal/sidecar/resolver_test.go` (modify)

---

### Task 3: Rewrite Doctor as Scope-Chain Validator

**Priority:** High
**Risk:** Medium — changes validation behavior.
**Depends on:** Task 1.

**Done when:**
- [x] `Doctor` validates all configured levels in the requested scope chain, not only the highest-priority active config.
- [x] `sidecarCandidate` private type and `validateSidecarCandidate` helper added.
- [x] Level-specific anchors: global/workspace → `globalDir`, project → `projectRoot`.
- [x] Linked-project validation block removed (no TODO, no dead branch).
- [x] `Doctor(scope, projectRoot) ([]Issue, error)` signature unchanged.
- [x] `doctor_test.go` added: no sidecars returns no issues, validates global+project for project scope, validates global+project+workspace for workspace scope, level-specific anchors, empty path ERROR, missing root WARN, path-is-file ERROR, linked projects ignored.

**Files:**
- `packages/cli/internal/sidecar/doctor.go` (modify)
- `packages/cli/internal/sidecar/doctor_test.go` (new)

---

### Task 4: Remove LinkedProject Code from the v1.1 Surface

**Priority:** Medium
**Risk:** Low — removing dead half-feature.
**Depends on:** Task 3.

**Done when:**
- [x] `LinkedProjects []LinkedProject` removed from `SidecarConfig` in `types.go`.
- [x] `LinkedProject` type deleted from `types.go`.
- [x] `type LinkedProject = sidecarpkg.LinkedProject` alias deleted from `cmd/workspace.go`.
- [x] No Go code references `LinkedProject` (only spec text saying reserved/ignored and tests asserting tolerance).

**Files:**
- `packages/cli/internal/sidecar/types.go` (modify)
- `packages/cli/cmd/workspace.go` (modify)

---

### Task 5: Add Atomic Write and File-Lock Helpers

**Priority:** High
**Risk:** Low — new exported helpers, no existing callers changed.
**Depends on:** Nothing.

**Done when:**
- [x] `files.AtomicWriteFile(path, data, perm) (backupPath, err)` implemented: EnsureDir, single-slot `.bak` backup, temp file in same dir, `Sync`, `os.Rename`, temp cleanup on error.
- [x] `files.WithFileLock(lockPath, timeout, staleAfter, fn) error` implemented: `O_CREATE|O_EXCL|O_WRONLY` mode 0600, PID+timestamp payload, stale-lock removal, 25ms retry, timeout error.
- [x] `files_test.go` extended: creates-without-backup, replaces-and-backs-up, overwrites-single-slot-bak, serializes-concurrent, times-out-when-held, removes-stale-lock.

**Files:**
- `packages/cli/internal/files/files.go` (modify)
- `packages/cli/internal/files/files_test.go` (modify)

---

### Task 6: Unify Workspace Config Persistence Around Locked Read-Modify-Write

**Priority:** High
**Risk:** High — changes persistence behavior and command wiring.
**Depends on:** Tasks 4, 5.

**Done when:**
- [x] `SaveWorkspaceConfig(cfg) error` and `UpdateWorkspaceConfig(update func(*WorkspaceConfig) error) error` added to `writer.go`.
- [x] `UpdateWorkspaceConfig` acquires `workspaces.yaml.lock` with `WithFileLock`, loads inside lock, calls update, saves inside lock.
- [x] `LoadWorkspaceConfig` kept as public read-only loader (no lock).
- [x] `WriteWorkspaceSidecar` and `RemoveWorkspaceSidecar` use `UpdateWorkspaceConfig`.
- [x] `WriteProjectSidecar` validates `projectRoot` exists and is a directory; deletes `MkdirAll(projectRoot)`.
- [x] Project/global sidecar writes use `files.AtomicWriteFile`.
- [x] `cmd/workspace.go` duplicate helpers (`getGlobalConfigDir`, `getWorkspacesConfigPath`, `loadWorkspaceConfig`, `saveWorkspaceConfig`) deleted; callsites use `sidecarpkg.LoadWorkspaceConfig` / `UpdateWorkspaceConfig`.
- [x] `cmd/sidecar.go` `determineScope` returns workspace config load errors instead of swallowing them.
- [x] `cmd/sidecar.go` workspace mutations use `UpdateWorkspaceConfig` closures.
- [x] `writer_test.go` added: rejects missing/file project root, writes atomically with backup, SaveWorkspaceConfig backup, UpdateWorkspaceConfig concurrency.
- [x] `remover_test.go` added: removes active sidecar preserving others, no-active no-op, no-sidecar no-op, active-missing error, project remove, project-missing no-op, backup through SaveWorkspaceConfig.
- [x] `cmd/workspace_test.go` added: add creates+auto-activates, rejects duplicate, switch updates active, status reports missing path, list uses sidecar store.
- [x] `cmd/sidecar_test.go` added: init/attach/detach use locked workspace update.

**Files:**
- `packages/cli/internal/sidecar/writer.go` (modify)
- `packages/cli/internal/sidecar/remover.go` (modify)
- `packages/cli/cmd/workspace.go` (modify)
- `packages/cli/cmd/sidecar.go` (modify)
- `packages/cli/internal/sidecar/writer_test.go` (new)
- `packages/cli/internal/sidecar/remover_test.go` (new)
- `packages/cli/cmd/workspace_test.go` (new)
- `packages/cli/cmd/sidecar_test.go` (new)

---

### Task 7: Fix Command Root Handling and Project Attach Typo Prevention

**Priority:** Medium
**Risk:** Low — error path fixes.
**Depends on:** Task 6.

**Done when:**
- [x] `runSidecarStatus` propagates `getProjectRoot` errors.
- [x] `runSidecarDoctor` propagates `getProjectRoot` errors.
- [x] `runSidecarAttach` project scope validates project path exists and is a directory before writing.
- [x] `runSidecarInit` project scope validates project root before writing.
- [x] `cmd/sidecar_test.go` added: status propagates root error, doctor propagates root error, attach rejects missing project path without creating it, determineScope returns workspace config load error.

**Files:**
- `packages/cli/cmd/sidecar.go` (modify)
- `packages/cli/cmd/sidecar_test.go` (modify)

---

### Task 8: Final Verification

**Priority:** High
**Risk:** None — verification only.
**Depends on:** Tasks 1–7.

**Done when:**
- [x] `go build ./packages/cli/...` passes.
- [x] `go test ./packages/cli/... -count=1` passes (existing + new tests).
- [x] Focused domain tests pass: `go test ./packages/cli/internal/sidecar -run 'Test(Resolve|Doctor|Write|Save|Update|Remove)' -count=1`.
- [x] Files helper tests pass: `go test ./packages/cli/internal/files -run 'Test(AtomicWriteFile|WithFileLock)' -count=1`.
- [x] Focused command tests pass: `go test ./packages/cli/cmd -run 'Test(Sidecar|Workspace|DetermineScope)' -count=1`.
- [x] No test touches real `HOME`; all use `t.TempDir` + `t.Setenv("HOME")`.
- [x] `docs/reports/sidecar-workspaces-adversarial-review-2026-06-17.md` addendum added.

**Files:**
- `docs/reports/sidecar-workspaces-adversarial-review-2026-06-17.md` (modify)

---

## Acceptance Criteria (Cross-Task)

- [x] AC1: Spec package is v1.1.0 Approved with all decisions encoded.
- [x] AC2: Empty sidecar path is rejected by Resolve, doctor, and status.
- [x] AC3: Doctor validates all configured levels in the scope chain.
- [x] AC4: Relative paths anchor by config level, not requested scope.
- [x] AC5: Workspace mutations are locked and atomic; project/global writes are atomic.
- [x] AC6: `linked_projects` is reserved — no Go code references it.
- [x] AC7: Command root errors and project attach typo prevention verified by tests.
- [x] AC8: No regressions — existing tests pass alongside new tests.

---

## Risk Summary

| Task | Risk | Mitigation |
|---|---|---|
| Task 2 (Resolver) | Medium — anchor behavior change | Explicit tests for each level/scope combination |
| Task 3 (Doctor) | Medium — validation behavior change | Scope-chain tests for all three scope levels |
| Task 6 (Persistence) | High — persistence + wiring change | Atomic write tests, lock tests, command tests; reuse sidecar store everywhere |
| Task 7 (Commands) | Low — error path fixes | Root-error and typo-prevention tests |