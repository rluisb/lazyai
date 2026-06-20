# Spec: 027-production-readiness-hardening

**Feature ID:** 027
**Feature name:** production-readiness-hardening
**Date:** 2026-06-20
**Status:** Draft
**Owner:** Ricardo Borges
**Constitution:** [specs/standards](../standards)

> **Purpose.** Close the blocking and high-risk gaps that make a LazyAI production release unsafe today. Scope is the release pipeline, CI enforcement, archive-restore safety, notification input handling, advertised-but-stub commands, scope-metadata correctness, and the deferred verification coverage already tracked in `KNOWLEDGE_MAP.md`. *What* and *why* live here; *how* lives in `plan.md`; execution lives in `tasks.md`.

---

## Context & Evidence

This spec is the remediation contract for the readiness assessment performed 2026-06-20. Each finding below cites the source file inspected.

| Finding | Severity | Evidence |
|---|---|---|
| Root release workflow triggers on `v*` and injects `-X main.Version`, but the version symbol is `cmd.Version`; docs mandate package-scoped tags | Blocker | `.github/workflows/release.yml:5-6,33-37`; `packages/cli/cmd/root.go:15-19`; `docs/wiki/Release-Process.md:7-14` |
| `backup restore` writes `header.Name` directly with `MkdirAll`/`os.Create` — no path sanitization | Blocker | `packages/cli/cmd/backup.go:132-151` |
| Integration scripts exist but no CI job runs them | Blocker | `tests/integration/*.sh`; no invocation in `.github/workflows/*` |
| Smoke test step is `continue-on-error: true` | Blocker | `.github/workflows/test.yml:56-60` |
| Snapshot tests for library assets + compiled output still deferred | High | `specs/KNOWLEDGE_MAP.md:124` |
| CI-side validation with `opencode` binary still deferred | High | `specs/KNOWLEDGE_MAP.md:127` |
| `validate skills` is an advertised stub returning "not yet implemented" | High | `README.md:180`; `packages/cli/cmd/validate.go:165-181` |
| Lint gates advisory: `staticcheck` `continue-on-error`, `yamllint ... || true` | High | `.github/workflows/lint.yml` |
| Coverage upload `fail_ci_if_error: false` | Medium | `.github/workflows/test.yml:35-39` |
| Release asset/tag naming diverges across docs, `release.yml`, and goreleaser | Medium | `docs/wiki/Release-Process.md:37-39`; `packages/cli/.goreleaser.yaml` |
| Notification helpers interpolate unsanitized title/message into shell/script payloads | Medium | `packages/cli/cmd/notify.go:120-148` |
| `setupscan` advertises all scopes for all tools; Pi/Antigravity are global-unsupported | Medium | `packages/cli/internal/setupscan/setupscan.go`; `packages/cli/internal/adapter/scope.go:30-34` |
| `go.mod` toolchain directive `go 1.26.1` flagged as a potential parse/build risk (needs verification) | Low | `packages/cli/go.mod:3` |

---

## User Scenarios

### P1 — A maintainer can cut a release that builds and versions correctly
**As a** LazyAI maintainer
**I want** a single, correct release path with consistent tags, version injection, and asset names
**So that** `go install ...@vX.Y.Z` and downloaded binaries report the right version and match the documented contract

**Acceptance criteria**
- [ ] Given a release tag created per the documented contract, when release automation runs, then binaries are built with the version correctly injected and reported by `lazyai-cli --version`.
- [ ] Given the documented tag/asset naming, when a release publishes, then asset names and tag format match `docs/wiki/Release-Process.md` exactly (docs and automation agree).
- [ ] Given the repository has both root and package release workflows, when the dust settles, then exactly one canonical release model exists with no dead/contradictory workflow.

### P1 — CI blocks regressions instead of advising
**As a** reviewer
**I want** smoke, integration, lint, and coverage signals to be enforcing gates
**So that** a red check actually blocks a merge

**Acceptance criteria**
- [ ] Given a failing smoke test, when CI runs on a PR, then the PR check fails (no `continue-on-error`).
- [ ] Given integration scripts under `tests/integration/`, when CI runs, then those scripts execute and a failure fails the pipeline.
- [ ] Given a `staticcheck`/`yamllint` violation, when CI runs, then the violation fails the pipeline (or is moved to a clearly separate, named advisory job).

### P1 — `backup restore` cannot write outside the target directory
**As a** user restoring a backup
**I want** archive entries validated before write
**So that** a malicious or malformed archive cannot escape the restore root

**Acceptance criteria**
- [ ] Given an archive entry with `..` or an absolute path, when `backup restore` runs, then the entry is rejected and no file is written outside the restore root.
- [ ] Given a well-formed archive, when `backup restore` runs, then restore behavior is unchanged for legitimate entries.

### P2 — Advertised commands are honest about behavior
**As a** user reading the docs/CLI
**I want** `validate skills` to either perform real validation or not be advertised as an active capability
**So that** documented behavior matches runtime behavior

**Acceptance criteria**
- [ ] Given `validate skills`, when it runs, then it performs real structural validation **or** the docs/CLI clearly mark it as unimplemented/hidden rather than an active command.

### P2 — Notifications handle arbitrary content safely
**As a** user with arbitrary title/message content
**I want** notification payloads constructed safely
**So that** content with quotes/special characters cannot break or inject into the underlying command

**Acceptance criteria**
- [ ] Given title/message containing quotes or shell metacharacters, when a notification is sent, then the content is delivered/escaped correctly with no command breakage.

### P2 — Scope metadata is accurate
**As an** operator inspecting setup inventory
**I want** scope listings to reflect actual adapter support
**So that** I am not shown unsupported scope combinations (Pi/Antigravity global)

**Acceptance criteria**
- [ ] Given `setupscan` output, when a tool does not support a scope, then that scope is not advertised for that tool.

### P3 — Deferred verification coverage is delivered or explicitly accepted
**As a** release owner
**I want** snapshot tests and CI `opencode` validation either added or formally accepted as release exceptions
**So that** `KNOWLEDGE_MAP.md` no longer carries silent open readiness gaps

**Acceptance criteria**
- [ ] Given the two deferred items in `KNOWLEDGE_MAP.md:124,127`, when this feature closes, then each is either implemented or recorded as an accepted exception with rationale.

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | Release automation MUST inject the version into the symbol the binary actually reads (`cmd.Version` / `internal/version.Version`), not `main.Version`. | P1 | P1 release |
| FR-002 | The repository MUST expose exactly one canonical release model; contradictory/dead release workflows MUST be removed or fixed. | P1 | P1 release |
| FR-003 | Release tag format and published asset names MUST match `docs/wiki/Release-Process.md`; docs and automation MUST agree. | P1 | P1 release |
| FR-004 | The smoke-test CI step MUST be an enforcing gate (no `continue-on-error`). | P1 | P1 CI |
| FR-005 | CI MUST execute the `tests/integration/*.sh` scripts and fail the pipeline on their failure. | P1 | P1 CI |
| FR-006 | `staticcheck` and YAML lint MUST be enforcing, or moved into an explicitly named advisory job separate from required gates. | P1 | P1 CI |
| FR-007 | `backup restore` MUST reject archive entries that resolve outside the restore root (`..`, absolute paths, symlink escape) before any write. | P1 | P1 restore |
| FR-008 | `validate skills` MUST perform real validation, OR be removed from advertised active commands and clearly marked unimplemented. | P2 | P2 validate |
| FR-009 | Notification helpers MUST construct payloads without unsafe string interpolation of user content. | P2 | P2 notify |
| FR-010 | `setupscan` MUST NOT advertise scopes a tool's adapter does not support. | P2 | P2 scope |
| FR-011 | Coverage upload failure handling MUST be a deliberate, documented choice (enforce or explicitly advisory). | P2 | P1 CI |
| FR-012 | The two deferred items (`KNOWLEDGE_MAP.md:124,127`) MUST be implemented or recorded as accepted exceptions. | P3 | P3 coverage |
| FR-013 | The `go.mod` toolchain directive MUST be verified valid for the supported toolchain and corrected if it breaks build/parse. | P3 | P1 release |

---

## Success Criteria

- **SC-001 — Release correctness:** A dry/staged release produces a binary whose `--version` matches the tag. Measured by: release workflow run + `lazyai-cli --version`.
- **SC-002 — Enforcing CI:** An intentionally failing smoke/integration/lint check fails a PR. Measured by: CI run status on a probe PR.
- **SC-003 — Restore safety:** A crafted traversal archive cannot write outside the restore root. Measured by: regression test asserting rejection.
- **SC-004 — Honest surface:** No advertised active command returns "not yet implemented". Measured by: `validate skills` behavior + docs review.
- **SC-005 — No silent readiness gaps:** `KNOWLEDGE_MAP.md` has no unchecked readiness items without an accepted-exception note. Measured by: doc review.

---

## Edge Cases

- **EC-001 — Traversal archive:** Entry `../../etc/evil` → rejected, no write, non-zero/warned outcome.
- **EC-002 — Absolute-path archive entry:** `/tmp/evil` → rejected.
- **EC-003 — Symlink entry escape:** symlink pointing outside root → not followed for writes outside root.
- **EC-004 — Notification with quotes:** title `a"b` / message `$(x)` → delivered literally, no command breakage.
- **EC-005 — Empty integration suite host deps:** integration scripts requiring `sqlite3`/`git` absent → CI provisions deps or the job fails loudly (not silently skipped).
- **EC-006 — Tool without a scope:** Pi at global scope → not listed by `setupscan`.

---

## Assumptions

- **A-001:** The package-scoped tag model in `docs/wiki/Release-Process.md` is the intended canonical release model — confidence: MEDIUM (confirm before deleting `release.yml`).
- **A-002:** `tests/integration/*.sh` are runnable on `ubuntu-latest` with documented deps — confidence: MEDIUM.
- **A-003:** `go 1.26.1` directive concern may be a false positive (patch versions are valid in modern Go) — confidence: MEDIUM; FR-013 verifies rather than assumes.

> Low/medium-confidence assumptions MUST be resolved before the dependent task is marked done.

---

## Out of Scope

- New runtime features or new adapters.
- Broad refactors of command surfaces beyond the named files.
- Performance work.
- Re-architecting the MCP compile pipeline.

---

## Clarifications

| Date | Question | Answer | Decided by |
|---|---|---|---|
| 2026-06-20 | Use spec number 024 or 027? | 027 — 024/025/026 already assigned in `KNOWLEDGE_MAP.md` to distinct features | agent + human approval pending |

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `plan` | this file |
| `tasks` | indirectly via plan |
