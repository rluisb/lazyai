# Plan: 029-lazyai-v2

**Feature ID:** 029
**Spec:** [./spec.md](./spec.md)
**Date:** 2026-06-22
**Status:** Draft
**Owner:** maintainers
**Constitution:** [specs/001-store-and-errors/constitution.md](../001-store-and-errors/constitution.md)

> **Purpose.** How V2 is built on top of the existing Go CLI. The load-bearing change is moving the compile source of truth from the implicit embedded `library/` root to an explicit **`.ai/lazyai.json` manifest + `.ai/lock.json`**, with a formal `discover → parse → normalize → resolve → validate → plan → write` pipeline, an adapter capabilities model, docs-conformance fixtures, validation hardening, migration/eject, plugin bundles, and an optional trace/eval loop.

---

## Summary

V2 turns `.ai/` into the canonical source. `compile` parses `.ai/lazyai.json`, resolves assets (default pack = embedded `library/`), validates, plans a managed-region diff, writes native files for the 7 targets, and records `.ai/lock.json`. Codex is removed; the binary stays `lazyai-cli`. The plan reuses the current adapter registry, `configmerge`, and `managed_block` substrates rather than rebuilding them. Work is phased so each phase is independently mergeable; P1 (manifest + lock + pipeline) is the nucleus.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language | Go (existing `packages/cli`) | Repo is Go; no new runtime |
| Framework | Cobra (existing `cmd/`) | All commands already Cobra |
| Storage | JSON files (`.ai/lazyai.json`, `.ai/lock.json`, `.ai/mcp.json`); existing SQLite for runtime-adjacent only | Spec wants inspectable, committable source |
| Schemas | Embedded JSON Schema under `packages/cli/internal/schema/` | No hosted `lazyai.dev` dependency (YAGNI) |
| Adapters | Existing `internal/adapter` registry (7 targets) | Already aligned to pack target set |
| Hashing | `crypto/sha256` over normalized asset bytes | Deterministic drift detection |
| Testing | `go test ./...`, golden + docs-conformance fixtures under `packages/cli/testdata/` | Matches pack conformance model |

**External dependencies (new):** none required for P1–P3. A JSON Schema validator (e.g. `santhosh-tekuri/jsonschema`) MAY be added in P2 if hand-rolled validation proves brittle — evaluate during P2, prefer stdlib first.

**External dependencies (rejected):** any agent-orchestration / tracing SDK (violates PRD §4 non-goals).

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Reuses Cobra, adapter registry, configmerge, managed_block, setupscan |
| II — Test-First (NON-NEGOTIABLE) | PASS | Each task writes failing fixtures/golden first; conformance tests precede adapter changes |
| III — Docs as Source of Truth | PASS | Spec pack normalized; KNOWLEDGE_MAP + AGENTS notes updated in cleanup phase |
| IV — Anti-Speculation (YAGNI) | PASS | Trace/eval + extra plugin bundles gated to P3/P4; no binary rename; no hosted schema registry |
| V — Simplicity Over Abstraction | PASS | Managed-region writer extends existing block logic; no new templating engine |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Manifest is plain JSON; capabilities model is a struct, not a plugin DSL |

**Verdict:** APPROVED (pending human gate).

---

## Project Structure

```
packages/cli/
├── cmd/
│   ├── compile.go                 ← modified: read manifest, plan→write→lock
│   ├── init.go                    ← modified: scaffold .ai/lazyai.json + lock
│   ├── validate*.go               ← modified/added: validate --all
│   ├── doctor.go                  ← modified: security report
│   ├── import.go                  ← new: migration engine entrypoint
│   ├── eject.go                   ← new
│   └── build_plugin.go            ← modified: OMP/Copilot-CLI/Pi bundles
├── internal/
│   ├── manifest/                  ← new: lazyai.json parse/resolve
│   ├── lockfile/                  ← new: lock.json read/write + hashing
│   ├── plan/                      ← new: diff planner over managed regions
│   ├── writer/                    ← new (or extend configmerge): safe managed writer
│   ├── schema/                    ← new: embedded JSON schemas
│   ├── adapter/                   ← modified: capabilities model; remove codex
│   ├── validate/                  ← modified: secret scanner, path/symlink safety
│   └── migrate/ + eject/          ← new (may wrap setupscan)
├── testdata/
│   ├── golden/                    ← per-adapter expected outputs
│   ├── fixtures/                  ← starter .ai/ trees
│   └── docs-snapshots/            ← conformance expectations
└── library/                       ← default asset pack (referenced by manifest)

specs/029-lazyai-v2/
├── spec.md   ← contract
├── plan.md   ← this file
└── tasks.md  ← phased tasks
```

---

## Data Model

| Entity | Storage | Fields | Constraints |
|---|---|---|---|
| Manifest | `.ai/lazyai.json` | version, profile, targets[], adapters{}, source{}, safety{} | targets ⊆ 7 supported; profile ∈ {personal,team} |
| Lockfile | `.ai/lock.json` | version, lazyaiVersion, compiledAt, adapters{version,docsSnapshot}, generated[]{path,target,sourceHash,outputHash,managed} | hashes sha256 |
| MCP catalog | `.ai/mcp.json` | servers{}, env refs only | no inline secrets |

**Migrations:** none (file-based). Existing SQLite schema untouched.
**Backfill:** `import` generates `.ai/` from native config; no automatic data movement otherwise.

---

## Internal Contracts

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| `manifest.Load(dir) (*Manifest, error)` | `internal/manifest` | `cmd/compile`, `doctor` | parsed manifest struct |
| `Adapter.Capabilities() Capability` | each adapter | compile, doctor, validate | surfaces + stability |
| `plan.Build(resolved, lock) Plan` | `internal/plan` | writer | ordered writes w/ managed bounds |
| `writer.Apply(plan, opts) (Lock, error)` | `internal/writer` | compile | writes + new lock entries |
| `lockfile.Load/Save` | `internal/lockfile` | compile, plan | lock struct |

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| Manifest model breaks existing `compile` callers | M | H | Keep adapter output functions stable; change only the source-resolution layer; golden tests guard output | lead |
| Codex removal breaks build/tests/specs | M | M | Inventory all Codex refs first (EC-005); remove behind one task with green gates | lead |
| Managed-region writer corrupts hand edits | L | H | Reuse `managed_block`; round-trip tests; refuse-on-drift without `--force` | lead |
| Beta adapters (OMP/Antigravity) fail conformance | M | M | Label beta; gate bundles; warnings not hard fail | lead |
| Scope creep into runtime/eval | M | M | P3/P4 gated; non-goals enforced in spec | maintainers |

---

## Complexity Tracking

| Item | Simpler alternative | Why justified | Cost |
|---|---|---|---|
| New `manifest`/`lockfile`/`plan`/`writer` packages | Inline in `cmd/compile.go` | Pipeline stages must be unit-testable and reused by doctor/import | moderate cognitive |
| Adapter capabilities model | Hard-coded per-tool switches | Drives compile/doctor/validate uniformly; removes scattered branching | low |
| Docs-conformance fixtures | Rely on existing golden tests | Pack ties official-doc requirements to named tests (FR-008) | test maintenance |

---

## Phases & Milestones

| Phase | Goal | Exit criterion |
|---|---|---|
| A — Manifest nucleus | `.ai/lazyai.json` + `lock.json` + manifest-driven `compile` with plan→write→lock | Starter `.ai/` compiles for 7 targets; second run = empty diff; golden tests green |
| B — Capabilities & conformance | Adapter capabilities model + docs-conformance fixtures; remove Codex | Every stable adapter passes conformance; Codex removed, gates green |
| C — Validation & safety | `validate --all` + secret scanner + path/symlink safety + `doctor` security report | Negative-case tests pass; Pi/Kiro no-sandbox caveats reported |
| D — Migration & eject | `import` + `eject` engines | Round-trip import→eject integration test passes, no native loss |
| E — Bundles & eval (optional) | OMP/Copilot-CLI/Pi plugin bundles + eval schemas + `validate evals` | Bundles pass native validation; an eval case validates; no cloud dep |
| F — Cleanup | Docs/KNOWLEDGE_MAP/AGENTS normalization, schema freeze | Full `make test`/`vet`/`lint`/`mkdocs --strict` green; pack `lazyai`→`lazyai-cli` normalized |

> Phase A is the V2 nucleus and the minimum mergeable V2. B–F layer on. Tasks live in `tasks.md`.

---

## Out of Scope

- Binary rename to `lazyai` — deferred/declined.
- Codex support — removed.
- Hosted schema registry — schemas embedded.
- Runtime/orchestration/cloud-tracing features — PRD §4 non-goals.

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | this file + spec.md → `tasks.md` |
| `speckit-implement` | this file + task harnesses (after human gate) |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | agent | 2026-06-22 | drafted |
| Human gate | — | — | **pending — implementation blocked until approved** |
