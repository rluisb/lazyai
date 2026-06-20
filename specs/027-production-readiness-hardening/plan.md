# Plan: 027-production-readiness-hardening

**Feature ID:** 027
**Spec:** [./spec.md](./spec.md)
**Date:** 2026-06-20
**Status:** Draft
**Owner:** Ricardo Borges
**Constitution:** [specs/standards](../standards)

> **Purpose.** *How* the gaps in `spec.md` get closed. Each workstream maps to functional requirements and names exact files. The work is decomposed so that file-disjoint workstreams can run concurrently via subagents, while shared-file workstreams (CI/release YAML) are sequenced under a single owner.

---

## Summary

Close all Blocker and High findings from the 2026-06-20 readiness review plus the tracked Medium/deferred items. The work splits into two execution lanes: **(A) parallel, file-disjoint code fixes** safe to run as independent subagents, and **(B) sequenced workflow/release changes** that share `.github/workflows/*` and must be coordinated by one owner to avoid merge collisions.

Satisfies spec stories P1–P3 (FR-001 … FR-013).

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language(s) | Go 1.26 (`packages/cli`) | Matches existing module |
| Test framework | `go test` + `testify` | Already in `go.mod` |
| CI | GitHub Actions | Existing `.github/workflows/*` |
| Release tooling | GoReleaser + Actions | Existing `packages/*/.goreleaser.yaml` |
| Docs | mkdocs (`mkdocs build --strict`) | Existing gate; verified passing |

**External dependencies (new):** none anticipated.

**External dependencies (rejected):** new archive libs for restore — stdlib `archive/tar` + `filepath` sanitization is sufficient.

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Uses stdlib + existing testify; no new frameworks |
| II — Test-First | PASS | Each fix lands with a failing-first regression test (restore traversal, notify escaping, scope listing, validate skills) |
| III — Docs as Source of Truth | PASS | Release docs, README, KNOWLEDGE_MAP updated as part of the work |
| IV — Anti-Speculation | PASS | Scope limited to named findings; no new features/adapters |
| V — Simplicity | PASS | Prefer deleting the broken `release.yml` over maintaining two models |
| VI — Anti-Overengineering | PASS | Minimal, targeted edits to named files |

**Verdict:** APPROVED (pending human gate)

---

## Project Structure

```
packages/cli/
├── cmd/backup.go                      ← modified (FR-007)
├── cmd/backup_test.go                 ← added/extended (FR-007)
├── cmd/notify.go                      ← modified (FR-009)
├── cmd/notify_test.go                 ← added (FR-009)
├── cmd/validate.go                    ← modified (FR-008)
├── cmd/validate_test.go              ← extended (FR-008)
├── internal/setupscan/setupscan.go    ← modified (FR-010)
├── internal/setupscan/setupscan_test.go ← extended (FR-010)
├── internal/adapter/*_snapshot_test.go ← added (FR-012)
├── go.mod                             ← verified/possibly modified (FR-013)
.github/workflows/
├── release.yml                        ← fixed or removed (FR-001, FR-002, FR-003)
├── release-cli.yml                    ← canonical model (FR-002)
├── test.yml                           ← smoke enforced + coverage decision (FR-004, FR-011)
├── ci-integration.yml                 ← added: runs tests/integration/*.sh (FR-005)
├── lint.yml                           ← gates enforced/split (FR-006)
docs/wiki/Release-Process.md           ← reconciled (FR-003)
README.md                              ← validate skills honesty (FR-008)
specs/KNOWLEDGE_MAP.md                 ← deferred items closed/accepted (FR-012)
packages/cli/.goreleaser.yaml          ← asset naming reconciled (FR-003)
```

---

## Internal Contracts

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| Sanitized restore path | `backup.go` restore | filesystem | `filepath.Clean` joined + bounded under restore root |
| Scope advertisement | `setupscan.go` | CLI inventory/UX | only `(tool, scope)` pairs where `IsScopeSupported` is true |
| Release version symbol | release workflow | `cmd.Version`/`internal/version.Version` | `-X .../cmd.Version=$TAG` (matches `Makefile:8`) |

> Note: the `Makefile:8` `CLI_VERSION_LDFLAGS` already targets the correct symbols — release automation should reuse that pattern rather than `-X main.Version`.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| Deleting `release.yml` breaks an active release path | M | H | Confirm A-001 (canonical model) before removal; keep `release-cli.yml`/`release-diffviewer.yml` | Lane B owner |
| Integration scripts flaky/host-dep heavy in CI | M | M | Provision `sqlite3`/`git`; start as required only after a green baseline run | Lane B owner |
| Enforcing lint surfaces large pre-existing backlog | M | M | Triage: fix quick wins, file follow-ups, but keep gate enforcing for new code | Lane B owner |
| Restore hardening breaks legitimate entries | L | M | Regression tests for both valid + traversal cases | Lane A |
| `go 1.26.1` "fix" is a false positive | M | L | FR-013 verifies first; only change if build/parse actually breaks | Lane A |

---

## Complexity Tracking

| Item | Simpler alternative | Why complexity is justified | Cost |
|---|---|---|---|
| Separate `ci-integration.yml` | Add steps to `test.yml` | Avoids subagent merge collisions on one file; clearer required-gate boundary | One extra workflow file |

---

## Phases & Milestones

| Phase | Goal | Exit criterion |
|---|---|---|
| 1 — Parallel code fixes (Lane A) | Land file-disjoint security/correctness/honesty fixes | Each fix merged with passing regression test |
| 2 — CI & release consolidation (Lane B) | One canonical release model + enforcing CI gates | Probe PR shows failing checks block; staged release injects correct version |
| 3 — Deferred coverage | Snapshot tests + opencode CI validation or accepted exception | `KNOWLEDGE_MAP.md` carries no silent gaps |
| 4 — Docs & closeout | Reconcile README/release docs/KNOWLEDGE_MAP | `mkdocs build --strict` green; docs match automation |

> Lane A (Phase 1) and most of Phase 3 are parallel-safe. Lane B (Phase 2) is single-owner sequential because it shares `.github/workflows/*`.

---

## Out of Scope

- New runtime/adapters, performance work, MCP pipeline re-architecture — deferred indefinitely (Article IV).

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `tasks` | this file + spec.md → `tasks.md` |
| `implement` | this file (technical context) + task list |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | agent | 2026-06-20 | drafted |
| Human gate | Ricardo Borges | — | pending |
