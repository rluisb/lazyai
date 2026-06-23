# Spec: 029-lazyai-v2

**Feature ID:** 029
**Feature name:** lazyai-v2
**Date:** 2026-06-22
**Status:** Done
**Owner:** maintainers
**Constitution:** [specs/001-store-and-errors/constitution.md](../001-store-and-errors/constitution.md)

> **Purpose.** This spec re-baselines the [LazyAI + vibe-lab Product Spec Pack](../../docs/lazyai-vibelab-product-spec-pack/) against the current repository and defines the V2 product slice: a **canonical `.ai/` manifest-driven compile pipeline** with a lockfile, an explicit adapter capabilities model, docs-conformance fixtures, validation hardening, migration/eject, multi-tool plugin bundles, and an optional trace/eval loop. The product stance is unchanged: **LazyAI is a...

The authoritative source documents are the spec pack:
- PRD: `docs/lazyai-vibelab-product-spec-pack/01_PRD.md`
- TechSpec: `…/02_TECHSPEC.md`
- Compliance matrix: `…/03_OFFICIAL_TOOL_COMPLIANCE_MATRIX.md`
- Schemas/examples: `…/04_SCHEMAS_AND_EXAMPLES.md`
- Roadmap: `…/05_IMPLEMENTATION_ROADMAP.md`

This spec narrows the pack to what V2 actually delivers given the existing codebase.

---

## Context / Evidence

| Area | Current repo state (grounded) | V2 target |
|---|---|---|
| Adapter registry | `packages/cli/internal/adapter/registry.go` registers OpenCode, ClaudeCode, Copilot, Pi, Omp, Kiro, Antigravity | Same 7 targets, formalized capabilities model |
| Compile source | `cmd/compile.go` + `internal/compiler/compiler.go` compile from embedded `packages/cli/library/` | Compile reads **`.ai/lazyai.json` manifest**; library becomes the default pack |
| Canonical manifest | `.ai/` holds `mcp.json`, `housekeeping/`, `constitution/`; **no `lazyai.json`** | `.ai/lazyai.json` is the source of truth |
| Lockfile | absent | `.ai/lock.json` records versions + source/output hashes + managed regions |
| Diff/safe write | `internal/configmerge` (backup-on-touch), `adapter/managed_block.go` | Formal `plan → write → lock` loop with managed-region writer |
| Codex / Gemini-extension | Codex adapter + `build-gemini-extension` exist (specs 017/018) | **Codex deprecated** (decision); Gemini served only via Antigravity adapter |
| Binary name | `lazyai-cli` (Makefile, KNOWLEDGE_MAP) | **Unchanged** — keep `lazyai-cli` (decision) |
| Validation | `validate skills`, `doctor` | Hardened: secret scanner, path/symlink safety, drift detector, doctor security report |
| Migration/eject | `internal/setupscan`, `.ai-setup-backup/` | Formal import + `eject` engine |
| Plugin bundles | `build-plugin` (Claude) | Add OMP, Copilot-CLI, Pi bundles |
| Trace/eval | referenced in pack only | Optional schemas + `validate evals` |

Decisions recorded in Clarifications: target set matches the pack's 7 (Codex dropped), manifest-driven compile adopted, binary name retained.

---

## User Scenarios

### P1 — Compile a canonical `.ai/` into native outputs
**As a** solo senior engineer using multiple AI tools
**I want** to define agents/skills/rules/MCP once in `.ai/` and run `lazyai-cli compile`
**So that** every enabled target receives correct, native, non-destructive files derived from one manifest.

**Acceptance criteria**
- [ ] Given a repo with `.ai/lazyai.json` listing `targets`, when `lazyai-cli compile` runs, then each enabled adapter emits its native files (per the supported-surface table in PRD §8.1) into the repo.
- [ ] Given a prior compile, when `compile` runs again with no source change, then no file is rewritten (lockfile hash match) and exit is clean.
- [ ] Given a source asset changed, when `compile` runs, then only outputs whose source hash changed are rewritten, and `.ai/lock.json` is updated.
- [ ] Given a generated file edited by hand outside its managed region, when `compile` runs, then the hand edits are preserved and only the managed region is rewritten.
- [ ] Given `--dry-run`, when `compile` runs, then a diff plan is printed and no file is written.

### P2 — Validate before execution
**As a** security reviewer / team lead
**I want** `lazyai-cli validate --all` to fail on unsafe or non-conformant assets
**So that** broken or dangerous setup never reaches a host tool.

**Acceptance criteria**
- [ ] Given a skill with a non-kebab-case `name` or missing `description`, validation fails with a precise message.
- [ ] Given an inline secret in any asset or MCP env value, validation fails under the `team` profile.
- [ ] Given a hook that runs a destructive command, validation surfaces it in the doctor security report.
- [ ] Given an adapter output that would use a deprecated field (e.g. OpenCode `maxSteps`), the docs-conformance test fails.

### P3 — Migrate in, eject out
**As a** platform engineer adopting or leaving LazyAI
**I want** to import existing native config into `.ai/` and to `eject` later
**So that** adoption is reversible and never destroys existing setup.

**Acceptance criteria**
- [ ] Given a repo with existing `AGENTS.md`/`CLAUDE.md`/`.opencode`/`.claude`/`.github`/`.pi`/`.omp`/`.kiro`/`.agents`, when `lazyai-cli import` runs, then a `.ai/` source is produced with a confidence-scored migration report and no native file is deleted.
- [ ] Given a LazyAI-managed repo, when `lazyai-cli eject` runs, then generated native files remain usable by host tools and LazyAI management metadata is removed.

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The CLI MUST read `.ai/lazyai.json` as the compile source of truth, resolving `targets`, `adapters`, `source`, and `safety`. | P1 | P1 |
| FR-002 | The embedded `packages/cli/library/` MUST act as the default asset pack referenced by the manifest, not the implicit compile root. | P1 | P1 |
| FR-003 | The CLI MUST write `.ai/lock.json` recording lazyai version, adapter versions/docs-snapshot dates, and per-output `sourceHash`/`outputHash`/`managed` flags. | P1 | P1 |
| FR-004 | `compile` MUST skip rewriting any output whose `sourceHash` and `outputHash` match the lockfile. | P1 | P1 |
| FR-005 | The writer MUST preserve content outside managed regions and only rewrite managed regions of generated files. | P1 | P1 |
| FR-006 | `compile --dry-run` MUST print a diff plan and write nothing. | P1 | P1 |
| FR-007 | The adapter interface MUST expose a capabilities model (supported surfaces, beta/stable status) consumed by compile and `doctor`. | P1 | P1 |
| FR-008 | Each adapter MUST have docs-conformance fixtures under `packages/cli/testdata/` asserting native shape and absence of deprecated fields. | P1 | P1 |
| FR-009 | `validate --all` MUST run skill, agent, hook, and MCP validators plus a secret scanner and path/symlink safety check. | P2 | P2 |
| FR-010 | Validation MUST fail on inline secrets under the `team` profile and warn under `personal`. | P2 | P2 |
| FR-011 | `doctor` MUST emit a security report covering MCP inventory, hook risks, and trust/sandbox caveats (Pi/Kiro have no sandbox). | P2 | P2 |
| FR-012 | The compiler MUST emit `CLAUDE.md` for the Claude target; `AGENTS.md` alone MUST NOT satisfy Claude conformance. | P1 | P1 |
| FR-013 | The OpenCode adapter MUST use `permission` and `steps`, never deprecated `tools`/`maxSteps`. | P1 | P1 |
| FR-014 | The product MUST support exactly the 7 targets {opencode, claude, copilot, pi, omp, antigravity, kiro}; the Codex adapter MUST be removed from the supported set. | P1 | P1 |
| FR-015 | `import` MUST produce a `.ai/` source from existing native config with a confidence-scored report and never delete native files. | P3 | P3 |
| FR-016 | `eject` MUST leave generated native files host-usable and remove LazyAI management metadata. | P3 | P3 |
| FR-017 | `build-plugin` MUST support OMP and Copilot-CLI bundles and a Pi package bundle in addition to the existing Claude bundle. | P2 | P2 |
| FR-018 | The CLI SHOULD provide eval-case/holdout/rubric schemas and `validate evals`, with no cloud tracing dependency. | P3 | P3 |
| FR-019 | The binary name MUST remain `lazyai-cli`; docs in the pack referencing `lazyai` are normalized to `lazyai-cli`. | P1 | — |

---

## Key Entities

| Entity | Description | Lifecycle |
|---|---|---|
| Manifest (`.ai/lazyai.json`) | Declares targets, adapter config, source packs, safety profile | created by `init`, edited by user |
| Lockfile (`.ai/lock.json`) | Records versions and source/output hashes + managed flags | written by `compile`, read for drift |
| MCP catalog (`.ai/mcp.json`) | Canonical MCP server definitions compiled per target | created by `init`/`import` |
| Adapter capability | Per-target descriptor of supported surfaces + stability | static, owned by each adapter |
| Diff plan | Ordered set of pending writes with managed-region boundaries | ephemeral per `compile` |
| Migration report | Confidence-scored import findings | produced by `import` |

---

## Success Criteria

- **SC-001 — Manifest compile:** A starter `.ai/` compiles cleanly for all 7 targets with zero unsafe overwrites. Measured by: golden tests + `compile --dry-run` showing empty diff on second run.
- **SC-002 — Conformance:** Every stable adapter passes its docs-conformance fixtures in CI. Measured by: `make test` green over `packages/cli/testdata/`.
- **SC-003 — Safety:** `validate --all` fails on inline secrets and dangerous hooks; `doctor` reports Pi/Kiro no-sandbox caveats. Measured by: negative-case tests.
- **SC-004 — Reversibility:** A multi-tool repo can `import` then `eject` with no native file loss. Measured by: round-trip integration test under `tests/`.

---

## Edge Cases

- **EC-001 — No manifest:** When `.ai/lazyai.json` is absent, `compile` errors with guidance to run `init` (no implicit library-root compile).
- **EC-002 — Hand-edited generated file:** When a managed file has user edits outside the managed region, the writer preserves them and rewrites only the region.
- **EC-003 — Lockfile drift:** When an output file was changed on disk but lockfile says managed, `compile` reports drift and (without `--force`) refuses to clobber.
- **EC-004 — Unsupported surface:** When an adapter cannot express a requested feature (e.g. Antigravity hook schema unverified), it emits a warning, not a silent omission.
- **EC-005 — Codex references:** When existing config/specs reference Codex, removal must not break the build; remaining adapters and tests stay green.
- **EC-006 — Beta adapter:** When OMP/Antigravity are beta, `doctor` and compile output label them beta and gate plugin bundles accordingly.

---

## Assumptions

- **A-001:** The 7-target set in the pack supersedes the current Codex/Gemini-extension surface — confidence: HIGH (explicit human decision).
- **A-002:** The existing `library/` content can serve as the default pack without restructure for P1 — confidence: MEDIUM (verify during plan).
- **A-003:** `internal/configmerge` + `adapter/managed_block.go` are reusable substrates for the managed-region writer — confidence: MEDIUM.
- **A-004:** Go toolchain availability in CI is unchanged; local Go may be absent (per spec-authoring skill) — confidence: HIGH.

---

## Out of Scope

- Becoming an agent runtime, orchestration framework, RAG/memory daemon, or cloud tracing platform (PRD §4 non-goals).
- Renaming the binary to `lazyai` (decision: keep `lazyai-cli`).
- Reviving or extending Codex support.
- Implementing a hosted schema registry at `lazyai.dev` (schemas shipped locally/embedded).
- Antigravity/OMP plugin bundles graduating from beta until docs snapshots are captured.

---

## Clarifications

| Date | Question | Answer | Decided by |
|---|---|---|---|
| 2026-06-22 | Spec number? | 029 (KNOWLEDGE_MAP max was 028) | agent + human approval |
| 2026-06-22 | Disposition of Codex/Gemini-extension? | Match pack exactly — drop Codex; Gemini via Antigravity only | human |
| 2026-06-22 | Compile model? | Manifest-driven (`.ai/lazyai.json` + `lock.json`) as source of truth | human |
| 2026-06-22 | Binary name? | Keep `lazyai-cli`; normalize pack's `lazyai` references | human |

---

## Constitutional Notes

- **Article I — Library-First:** Reuses Cobra, existing adapter registry, `configmerge`, `managed_block`, `setupscan`.
- **Article IV — YAGNI:** Trace/eval loop kept optional (P3); no cloud deps; no binary rename.
- **Article V — Simplicity:** Manifest-driven compile chosen over a parallel plugin-DSL; managed-region writer reuses existing block logic rather than a new templating engine.

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-plan` | this file → `plan.md` |
| `speckit-tasks` | indirectly via plan → `tasks.md` |
