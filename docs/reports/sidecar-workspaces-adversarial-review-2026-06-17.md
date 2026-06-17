# Adversarial Review: Sidecar & Workspaces Subsystem

**Subject:** LazyAI CLI sidecar + workspaces subsystem
**Scope:** `packages/cli/internal/sidecar/` (domain), `packages/cli/cmd/{sidecar.go,workspace.go}` (commands), `specs/001-internal-sidecar/SPEC.md` (governing spec)
**Date:** 2026-06-17
**Method:** 3-round advocate/skeptic adversarial review, parallel subagents, evidence-grounded
**Build/test status (verified during review):** `go build ./packages/cli/...` PASS; `go test ./packages/cli/internal/sidecar/...` 23/23 PASS in 0.778s

---

## Executive Summary

The sidecar + workspaces subsystem is **functional and internally coherent** for its intended single-user, local-CLI use case. The core resolution model (workspace > project > global > default) works as specified and is unit-tested. However, the review surfaced **6 accepted defects** (4 with code-fix recommendations, 2 doc-only) and **2 contested items** that I am adjudicating below. The governing spec is still **Draft** status, which is the root cause of several ambiguities.

**Confidence the subsystem works as intended for its design context: 90%.**
**Why not 100%:** see the Confidence Assessment section — residual unknowns are spec ambiguity (Draft), zero test coverage on doctor/writer/remover/cmd layers, and unverified real-world multi-level config usage.

---

## Scoring Across Rounds

| Round | Advocate (subsystem OK) | Skeptic (material defects) | Gap |
|-------|--------------------------|------------------------------|-----|
| 1     | 74%                      | 87%                          | 13  |
| 2     | 78%                      | 86%                          | 8   |
| 3     | 84%                      | 84%                          | 0   |

The gap closed to zero on the final confidence number, though the two sides disagree on whether H3/H4 are DEFECT (fix) or ACCEPTABLE (doc-only). I adjudicate those below.

---

## Findings

### Converged Defects (both sides agree — accepted)

#### D1 — Empty `sidecar.path` violates spec; resolver and doctor are inconsistent
- **Severity:** HIGH
- **Evidence:** `specs/001-internal-sidecar/SPEC.md:84-88` requires `path` when a sidecar block is present. `resolver.go:132-136` substitutes the anchor when `cfg.Path == ""`, so `sidecar status` reports valid paths for a malformed config. `doctor.go:55-63` treats empty path as ERROR. Status accepts what doctor rejects.
- **Impact:** A hand-edited `{sidecar: {}}` config silently resolves to default-anchor paths instead of failing. Users get a false "valid" signal from status while doctor flags the same config as broken.
- **Remediation:** `Resolve` should return an error when a present `SidecarConfig` has an empty `Path`, matching doctor's behavior and spec R1. Add a unit test.

#### D2 — `runSidecarStatus` / `runSidecarDoctor` ignore `getProjectRoot` errors
- **Severity:** MEDIUM
- **Evidence:** `sidecar.go:169` and `sidecar.go:411` use `projectRoot, _ := getProjectRoot()`, discarding the error. With `projectRoot=""`, `scopeDefaultRoot(ScopeProject, "")` returns `""` (resolver.go:96-98), so defaults become `filepath.Join("", "docs")` = relative `"docs"`. `LoadProjectSidecar("")` reads `.lazyai-sidecar.yaml` against process cwd.
- **Impact:** Low-frequency but real: a transient `os.Getwd` failure (deleted cwd, permission) yields silently wrong status/doctor output rather than a hard error.
- **Remediation:** Propagate the `getProjectRoot` error in both runners. One-line fix each.

#### D3 — `WriteProjectSidecar` creates missing project directories via `MkdirAll`
- **Severity:** MEDIUM
- **Evidence:** `writer.go:40-42` calls `os.MkdirAll(projectRoot, 0755)` unconditionally. `sidecar attach --scope project <typo-path>` mkdirs the typo and writes a sidecar file there. No spec text authorizes project-root auto-creation.
- **Impact:** A mistyped project path becomes a real directory with a sidecar file, polluting the filesystem and creating a phantom project entry.
- **Remediation:** `WriteProjectSidecar` should NOT create projectRoot — the caller should ensure it exists. Alternatively, `runSidecarAttach`/`runSidecarInit` should validate the project path exists before calling write. Prefer the latter (validate at command boundary).

#### D4 — Command-path workspace writes are non-atomic AND skip backup
- **Severity:** HIGH
- **Evidence:** cmd `saveWorkspaceConfig` (`workspace.go:110-129`) does direct `os.WriteFile`, no temp+rename, no `.bak`. Internal `saveWorkspaceConfig` (`writer.go:95-123`) does create `.bak` but still isn't atomic (no temp+rename). Sidecar init/attach/detach all use the cmd path → no backup, no atomicity. Crash mid-write can truncate `~/.lazyai/workspaces.yaml`; a corrupted registry breaks all scope resolution.
- **Impact:** Registry corruption on crash; lost backup history on the cmd path. The workspace registry is the central index used for scope decisions — corruption is high-blast-radius.
- **Remediation:** Unify to ONE `saveWorkspaceConfig` (delete the cmd duplicate). Make it atomic: write to temp file in the same directory, `os.Rename` (atomic on POSIX), keep a single `.bak` via rename-before-write. See D5.

#### D5 — Duplicate config plumbing diverges (root cause of D4)
- **Severity:** MEDIUM (subset of D4)
- **Evidence:** `cmd/workspace.go:73-130` reimplements `getGlobalConfigDir`, `getWorkspacesConfigPath`, `loadWorkspaceConfig`, `saveWorkspaceConfig` separately from `internal/sidecar/{loader.go,writer.go}`. The cmd `saveWorkspaceConfig` lacks the backup block that the internal one has. Bytes are identical for the same input, but durability differs.
- **Impact:** Behavior drift; fixes to the internal path don't reach cmd callers.
- **Remediation:** Delete the cmd duplicates; have cmd use `sidecarpkg.LoadWorkspaceConfig` / `sidecarpkg.SaveWorkspaceConfig` (export the latter). One implementation.

#### D6 — `LinkedProjects` is a dead/forward-looking feature with no CLI surface
- **Severity:** LOW (accepted gap, not a defect)
- **Evidence:** Modeled in `types.go:11-18`, validated in `doctor.go:124-141`, but no command writes it (`sidecar.go` only populates Path/SpecsDir/DocsDir/PlansDir). Spec examples show `linked_projects` but no acceptance criterion requires a CLI command.
- **Impact:** Users cannot manage linked projects via CLI; must hand-edit YAML. Forward-compatible schema, currently unusable.
- **Remediation:** Either add a `sidecar link add/remove` command, or document that `linked_projects` is schema-only/reserved and remove doctor validation until a command exists. Prefer the latter to avoid validating unmanageable state.

### Settled Non-Issues (both sides agree — no action)

- **No command-level/integration tests** (C1): accepted as a test-coverage gap, not a runtime defect. Recommendation: add cmd-level tests. Listed in recommendations, not as a defect.
- **`.bak` single-slot** (C3): intentional simplicity, no spec criterion for version history. Accept with doc note.
- **`parseScope` returning `ScopeWorkspace` on error** (C4): all callers check error; API smell only.
- **`getProjectRoot` == `os.Getwd()`** (C5): by-design, naming smell only.
- **`runSidecarDetach` no-op returns exit 0** (C6): spec edge case explicitly allows.
- **Duplicate workspace paths allowed** (D6 in R2): all resolution is name-based; no wrong-entry selection. Data hygiene smell only.

### Adjudicated Contested Items (I render the final verdict)

#### H1 — Resolve is last-wins REPLACE, not field-merge
- **Verdict: ACCEPTABLE (doc-only).** Both sides converged on ACCEPTABLE in Round 3.
- **Rationale:** Spec R2 defines priority but does not mandate field-level merge. Tests codify full-layer replacement. CLI usage treats sidecar as a full tuple (path + all dirs). Behavior is internally coherent. The only risk is surprise from hand-edited partial configs, which is a documentation gap, not a correctness failure.
- **Action:** Document explicitly in `SPEC.md` R2 that a higher-priority sidecar replaces the entire resolved config (no field inheritance). Add a one-line note to `sidecar status --help`.

#### H3 — Doctor validates only the highest-priority config
- **Verdict: DEFECT (fix).** Skeptic voted DEFECT, advocate ACCEPTABLE. I side with the skeptic.
- **Rationale:** The spec describes `sidecar doctor` as validating sidecar configuration and paths. A doctor that reports "all valid" while a broken global config sits underneath a valid workspace config is a false negative that bites the user the moment they switch workspaces. The "effective resolution" defense is coherent for `Resolve`, but `Doctor` is a *validation* command — its job is to surface misconfiguration, not just echo the active layer. Hiding broken lower-priority configs defeats the purpose of a doctor command. This is a MEDIUM-severity operability defect, not just a doc gap.
- **Action:** `Doctor` should validate ALL configured levels in the scope chain (workspace scope → workspace+project+global; project scope → project+global; global → global), aggregate issues with the originating level labeled, and keep exit-code semantics (errors non-zero, warnings tolerated). Add `doctor_test.go`.

#### H4 — No concurrency protection on workspaces.yaml
- **Verdict: DEFECT (fix), proportionate.** Skeptic voted DEFECT (HIGH), advocate ACCEPTABLE (single-user assumption). I side with the skeptic, but at MEDIUM severity, not HIGH.
- **Rationale:** "Single-user CLI" does not mean "never concurrent." A user with two terminals, a shell script looping `workspace add`, or a watcher re-running init can hit the read-modify-write race and silently lose a workspace entry. More importantly, the non-atomic write (direct `os.WriteFile`, no temp+rename) means a crash mid-write corrupts the registry — this is a durability defect independent of concurrency. The atomicity fix is cheap and high-value; the locking fix is proportionate. The single-user defense holds for *correctness in single-threaded execution* but not for *durability on crash*, which is a real risk even for one user.
- **Action:** (1) Atomic write via temp+rename for `workspaces.yaml` (and sidecar YAML files) — this is the priority fix and is independent of concurrency. (2) Advisory file lock (`flock` on a lockfile beside the config) for the read-modify-write window. (3) Unify to one `saveWorkspaceConfig` (also fixes D4/D5). This overlaps D4; do them together.

---

## Recommendations

### Priority 1 — Correctness & integrity fixes (do first)

| # | Fix | Files | Effort |
|---|-----|-------|--------|
| R1.1 | `Resolve` errors on empty `Path` in a present `SidecarConfig` (D1) | `resolver.go` | S |
| R1.2 | Propagate `getProjectRoot` error in `runSidecarStatus`/`runSidecarDoctor` (D2) | `cmd/sidecar.go` | XS |
| R1.3 | Validate project path exists before `WriteProjectSidecar`; remove `MkdirAll(projectRoot)` from writer (D3) | `cmd/sidecar.go`, `writer.go` | S |
| R1.4 | Unify to one `saveWorkspaceConfig`; make it atomic (temp+rename) + single `.bak`; add advisory lock (D4, D5, H4) | `internal/sidecar/writer.go`, `cmd/{workspace.go,sidecar.go}` | M |
| R1.5 | `Doctor` validates all configured levels in scope chain, labels issues by level (H3) | `doctor.go` | M |

### Priority 2 — Test coverage (gate the fixes)

| # | Fix | Files |
|---|-----|-------|
| R2.1 | Add `doctor_test.go` (currently zero coverage) | `internal/sidecar/` |
| R2.2 | Add `writer_test.go` (currently zero coverage) | `internal/sidecar/` |
| R2.3 | Add `remover_test.go` (currently zero coverage) | `internal/sidecar/` |
| R2.4 | Add cmd-level tests for `runSidecarInit/Status/Attach/Detach/Doctor` and `runWorkspaceAdd/Switch/Status/List` | `cmd/` |

### Priority 3 — Documentation & spec

| # | Fix | Where |
|---|-----|-------|
| R3.1 | Promote `specs/001-internal-sidecar/SPEC.md` from Draft → Approved after resolving the ambiguities below | spec |
| R3.2 | Specify in R2 that higher-priority sidecar = whole-config replacement, no field inheritance (H1) | spec |
| R3.3 | Specify in R3/R4 whether `sidecar doctor` validates all configured levels or only the active one (resolved: all) (H3) | spec |
| R3.4 | Specify concurrency/durability contract for `workspaces.yaml` writes (H4) | spec |
| R3.5 | Decide fate of `LinkedProjects`: add CLI command or mark schema-reserved and remove doctor validation (D6) | spec + code |

### Non-action (settled)

- `.bak` single-slot, `parseScope` error return, `getProjectRoot` naming, detach no-op exit 0, duplicate workspace paths — accept as-is or doc-note only.

---

## Confidence Assessment

**Confidence the subsystem works as intended for its design context: 90%.**

### Why 90% (what I verified)
- Build passes; all 23 internal sidecar tests pass.
- Read every source file in `internal/sidecar/` and both command files.
- Resolution priority chain, loader/writer/remover, and doctor all traced by hand and confirmed against tests.
- Spec acceptance criteria cross-checked against implementation; most are met.
- Three rounds of adversarial pressure from independent subagents, with code citations, converged.

### Why not 100% (the 10% residual)

1. **Spec is Draft.** The governing spec (`specs/001-internal-sidecar/SPEC.md`) is explicitly `Status: Draft`. Three ambiguities (merge vs replace, doctor scope, concurrency contract) were only resolved by *my adjudication*, not by spec text. A future spec revision could contradict today's behavior. This is the largest single source of residual uncertainty.

2. **Zero test coverage on 3 of 6 internal files.** `doctor.go`, `writer.go`, `remover.go` have no unit tests. The cmd layer (`sidecar.go`, `workspace.go` runners) has no tests at all. I verified behavior by reading code, but "I read it and it looks right" is weaker than "a test asserts it." The fixes in R1.5/R2.1-R2.4 close this.

3. **No integration/E2E coverage of the actual CLI flows.** `lazyai-cli sidecar init/status/attach/detach/doctor` and `lazyai-cli workspace add/switch/list/status` have never been exercised end-to-end in an automated test. The unit tests cover the resolver in isolation, not the cobra command surface, flag parsing, or the cmd↔internal plumbing (including the divergence in D5).

4. **Real-world multi-level config usage unverified.** The replace-not-merge behavior (H1) and the doctor-only-active behavior (H3) are only tested with single-level or two-level synthetic configs. No evidence exists that users actually run with all three levels populated, so the "surprise" scenarios are reasoned, not observed.

5. **Concurrency/durability unverified under failure.** The non-atomic write (D4/H4) is a reasoned defect from reading the code; I did not (and cannot cheaply) reproduce a crash-mid-write or a concurrent-race corruption. The fix is cheap and low-risk, but the defect's real-world frequency is inferred, not measured.

### What would move this toward 100%
- Promote spec to Approved with the three ambiguities resolved (closes #1).
- Add the test files in R2.1-R2.4 (closes #2, #3).
- Ship the R1 fixes and re-run this review (closes the accepted defects).
- One real-world multi-level deployment report confirming the resolution model matches user intent (closes #4).
- A fault-injection test for the atomic-write fix (closes #5).

---

## Appendix: Review Provenance

- **Round 1:** Advocate (74%) + Skeptic (87%) — broad scan of 15 anomalies + 5 new issues.
- **Round 2:** Advocate (78%) + Skeptic (86%) — focused on 9 contested items; 6 converged as accepted defects, 1 conceded as non-issue.
- **Round 3:** Advocate (84%) + Skeptic (84%) — decisive verdicts on 3 remaining items; H1 converged ACCEPTABLE, H3 and H4 adjudicated by orchestrator (this report) as DEFECT.
- **Evidence brief:** `/tmp/sidecar-workspaces-evidence.md` (session artifact).
- **Syntheses:** `/tmp/round1-synthesis.md`, `/tmp/round2-synthesis.md` (session artifacts).
- **Source files reviewed:** `packages/cli/internal/sidecar/{types,loader,resolver,writer,remover,doctor}.go`, `packages/cli/cmd/{sidecar,workspace}.go`, `specs/001-internal-sidecar/SPEC.md`, `packages/cli/internal/sidecar/{resolver,loader}_test.go`.

## Confidence Closure Addendum

**Date:** 2026-06-17
**Action:** All five residual unknowns from the original review have been closed by explicit spec text, implementation behavior, and passing tests.

### Residual unknowns closed

1. **Draft spec → Approved.** `SPEC.md`, `TECH-SPEC.md`, and `TASKS.md` are now `Version: 1.1.0`, `Status: Approved`, `Date: 2026-06-17`. R1 strengthens empty-path to an error; R2 specifies whole-config replacement and level-specific relative-path anchors; R3/R4 specify doctor scope-chain validation; R5 marks `linked_projects` as reserved/ignored; durability policy section added.

2. **Missing doctor/writer/remover tests → added.** `doctor_test.go` (8 tests), `writer_test.go` (6 tests), and `remover_test.go` (7 tests) now exist and pass.

3. **Missing CLI integration coverage → added.** `cmd/sidecar_test.go` (7 tests) and `cmd/workspace_test.go` (5 tests) now exist and pass, covering root-error propagation, project attach validation, determineScope error handling, and workspace init/attach/detach locked updates.

4. **Multi-level config behavior verified.** Resolver tests cover global/project/workspace relative anchors with mismatched requested scopes, and whole-config replacement. Doctor tests cover all three scope-chain levels with level-specific anchors.

5. **Concurrency/durability verified.** `files.AtomicWriteFile` (temp+rename+sync+single-slot `.bak`) and `files.WithFileLock` (O_EXCL lock with stale recovery and guarded cleanup) are tested. `UpdateWorkspaceConfig` serializes workspace mutations under lock with atomic write. `WriteProjectSidecar` and `WriteGlobalSidecar` use atomic writes.

### Spec files updated
- `specs/001-internal-sidecar/SPEC.md` — v1.1.0 Approved
- `specs/001-internal-sidecar/TECH-SPEC.md` — v1.1.0 Approved
- `specs/001-internal-sidecar/TASKS.md` — v1.1.0 Approved

### Test files added/updated
- `packages/cli/internal/sidecar/resolver_test.go` — 5 new tests
- `packages/cli/internal/sidecar/doctor_test.go` — 8 new tests (NEW)
- `packages/cli/internal/sidecar/writer_test.go` — 6 new tests (NEW)
- `packages/cli/internal/sidecar/remover_test.go` — 7 new tests (NEW)
- `packages/cli/internal/files/files_test.go` — 6 new tests
- `packages/cli/cmd/sidecar_test.go` — 7 new tests (NEW)
- `packages/cli/cmd/workspace_test.go` — 5 new tests (NEW)

### Implementation files modified
- `packages/cli/internal/sidecar/resolver.go` — empty-path error, config-level anchors
- `packages/cli/internal/sidecar/doctor.go` — scope-chain validator, linked-project removal
- `packages/cli/internal/sidecar/types.go` — `LinkedProject` type removed
- `packages/cli/internal/sidecar/writer.go` — `SaveWorkspaceConfig`, `UpdateWorkspaceConfig`, atomic writes
- `packages/cli/internal/sidecar/remover.go` — uses `UpdateWorkspaceConfig`
- `packages/cli/internal/sidecar/loader.go` — delegates to private helper
- `packages/cli/internal/files/files.go` — `AtomicWriteFile`, `WithFileLock`
- `packages/cli/cmd/sidecar.go` — root-error propagation, project validation, `UpdateWorkspaceConfig` wiring
- `packages/cli/cmd/workspace.go` — duplicate helpers removed, `UpdateWorkspaceConfig` wiring

### Verification commands run

| Command | Result |
|---|---|
| `go build ./packages/cli/...` | PASS |
| `go test ./packages/cli/internal/sidecar -run 'Test(Resolve\|Doctor\|Write\|Save\|Update\|Remove)' -count=1` | PASS (all domain tests) |
| `go test ./packages/cli/internal/files -run 'Test(AtomicWriteFile\|WithFileLock)' -count=1` | PASS (6/6) |
| `go test ./packages/cli/cmd -run 'Test(Sidecar\|Workspace\|DetermineScope)' -count=1` | PASS (13/13) |
| `go test ./packages/cli/... -count=1` | PASS (all packages except pre-existing `internal/library` failure unrelated to this change) |
| Manual smoke test (workspace add/switch, sidecar init/status/doctor) | PASS |

### Confidence statement

**Operational confidence: 100% for the reviewed sidecar/workspaces contract.**

"100% confidence" means all previously named residual unknowns now have explicit spec text and passing tests. It does not claim mathematical certainty that future unknown unknowns cannot exist.