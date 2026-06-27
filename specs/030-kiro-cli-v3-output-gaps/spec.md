# Spec: 030-kiro-cli-v3-output-gaps

**Feature ID:** 030
**Feature name:** kiro-cli-v3-output-gaps
**Date:** 2026-06-24
**Status:** Done
**Owner:** rluisb
**Constitution:** [../../docs/canonical/constitution.md](../../docs/canonical/constitution.md)
**Tracking issue:** [#517](https://github.com/rluisb/lazyai/issues/517)

> **Purpose.** Close the gaps between Kiro CLI v3's documented native surfaces and what LazyAI's Kiro adapter emits. This describes *what* changes and *why*; the *how* lives in `plan.md`.

---

## Context & Evidence

LazyAI is a canonical `.ai/` source compiler that emits tool-native config; it is not an agent runtime. The Kiro adapter today emits five surfaces, verified against the codebase:

| Surface | Path | Source |
|---|---|---|
| Agents | `.kiro/agents/<name>.md` | `packages/cli/internal/adapter/kiro.go` |
| Skills | `.kiro/skills/<name>/SKILL.md` | `packages/cli/internal/adapter/kiro.go` |
| Prompts | `.kiro/prompts/<name>.md` | `packages/cli/internal/adapter/kiro.go` |
| MCP | `.kiro/settings/mcp.json` | `packages/cli/internal/adapter/mcp_compiler.go:compileKiroMCP` |
| Hooks | `.kiro/hooks/<name>.json` | `packages/cli/internal/adapter/kiro.go` |

Kiro CLI v3 docs (verified 2026-06-24) introduced native surfaces the adapter did not cover; Hooks are now implemented (see table above):

- **Hooks** — `https://kiro.dev/docs/cli/v3/hooks/`: standalone `.kiro/hooks/<name>.json`, schema `{"version":"v1","hooks":[…]}`, 11 triggers, `command`/`agent` action types. Feature-overview lists hooks as "Enhanced" with a new JSON schema. Implemented: `KiroAdapter.Capabilities().Hooks` is `true`; `.kiro/hooks/<name>.json` is emitted by `packages/cli/internal/adapter/kiro.go`.
- **Specs** — `https://kiro.dev/docs/cli/v3/specs/`: `.kiro/specs/<name>/{requirements,design,tasks}.md`, created interactively via `/spec new <name>`. These are user-authored workflow artifacts, not compiler output.
- **Permissions** — `https://kiro.dev/docs/cli/v3/permissions/`: permission files live at `~/.kiro/settings/permissions.yaml` and `~/.kiro/workspace-roots/<hash>/permissions.yaml`. Docs state: "A cloned repo cannot inject permission rules." LazyAI must not emit a repo-local permissions file.
- **Powers** — `https://kiro.dev/docs/powers/`: importable packages (`POWER.md` + optional `mcp.json` + `steering/`), installed via marketplace, GitHub URL, or local folder import. No auto-discovered `.kiro/powers/` repo path is documented.

Verified design facts driving scope:
- Canonical agent frontmatter is already valid Kiro v3 (unknown YAML keys tolerated); **no agent transform is needed** (`ResearchAgentFrontmatter` finding; `kiro.go` copies agents verbatim via `ShapeFlat`).
- Hooks are emitted in adapter `Install` via `CopyLibraryDirectory`, **not** through the `output_mapping.go` `AssetKind` table (peer pattern: Claude `claudecode/hooks/`, OMP `omp/hooks/`, Antigravity `antigravity/hooks.json`).
- Golden tests (`internal/compiler/golden_test.go:TestCompilerGolden`) compare output byte-for-byte against `testdata/golden/<project>/`.

---

## User Scenarios

### P1 — Kiro users get native runtime hooks
**As a** developer who runs `lazyai-cli compile` with `kiro` selected
**I want** LazyAI to emit Kiro v3 native hook files
**So that** my Kiro CLI actually enforces the LazyAI safety/workflow hooks at runtime instead of only describing them in prose.

**Acceptance criteria**
- [x] Given a project with `kiro` selected and hook assets chosen, when I run `compile`, then `.kiro/hooks/<name>.json` files are written.
- [x] Given an emitted hook file, when parsed as JSON, then it conforms to the Kiro v3 schema: top-level `version: "v1"` and a `hooks` array whose entries carry a valid `trigger` (one of the 11 documented triggers) and an `action` with `type` ∈ {`command`,`agent`}.
- [x] Given an emitted `command` hook, when inspected, then its referenced command/script is present and runnable (no dangling reference).
- [x] Given a re-run of `compile` with no source change, then the hook write is a no-op (idempotent; no drift).

### P2 — Adapter metadata and docs match reality
**As a** maintainer reading capability metadata and docs
**I want** the Kiro adapter's declared capabilities and the docs to reflect what is actually emitted
**So that** capability/conformance tests are honest and contributors are not misled.

**Acceptance criteria**
- [x] Given `KiroAdapter.Capabilities()`, when read, then `Hooks` is `true`.
- [x] Given `capabilities_test.go`, when run, then it asserts Kiro declares `Hooks` and still rejects `Specs` and `Steering`.
- [x] Given `docs/README.md`, `docs/adapters/capability-matrix.md`, `docs/reference/hooks.md`, and `docs/reference/tool-outputs.md`, when read, then Kiro hooks are documented as supported (not "instruction-only"), and the Kiro non-goals (specs, repo-local permissions, direct `.kiro/powers`) are stated.

### P3 — Powers packaging direction is recorded (no code yet)
**As a** maintainer
**I want** a recorded decision on Kiro Powers
**So that** we do not guess an unverified `.kiro/powers/` output path.

**Acceptance criteria**
- [x] Given the docs, when read, then Powers is documented as an importable-package future option, not an emitted directory.

---

## Functional Requirements

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The Kiro adapter MUST emit `.kiro/hooks/<name>.json` files for selected hook assets during `Install`. | P1 | P1 |
| FR-002 | Each emitted hook file MUST be valid JSON conforming to the Kiro v3 hook schema (`version: "v1"`, `hooks[]` with valid `trigger` + `action.type`). | P1 | P1 |
| FR-003 | The adapter MUST only emit a hook whose `trigger` is a source-verified Kiro v3 trigger; canonical hooks without a verified Kiro equivalent MUST NOT be emitted as guessed mappings. | P1 | P1 |
| FR-004 | Hook emission MUST be idempotent and participate in the existing managed-write/drift model (no spurious rewrites). | P1 | P1 |
| FR-005 | `KiroAdapter.Capabilities().Hooks` MUST be `true`; `Specs` and `Steering` MUST remain `false`. | P1 | P2 |
| FR-006 | The adapter MUST NOT emit a repo-local Kiro permissions file. | P1 | P2 |
| FR-007 | Capability/conformance tests MUST be updated to assert Kiro hooks are present and `.kiro/specs`/`.kiro/steering` remain absent. | P1 | P2 |
| FR-008 | Golden fixtures (`kiro-only`, `full-seven-targets`) MUST be regenerated to include the new hook output. | P1 | P1 |
| FR-009 | Docs MUST be updated to mark Kiro hooks supported and to record Kiro non-goals (specs, repo-local permissions, direct `.kiro/powers`). | P1 | P2 |
| FR-010 | The Kiro `Permissions` capability flag SHOULD be documented as host-support metadata (no emitted repo file), consistent with the OMP metadata-flag precedent. | P2 | P2 |
| FR-011 | Docs SHOULD record Powers as an importable-package future option, not an emitted `.kiro/powers/` path. | P2 | P3 |

> **MUST** = non-negotiable, **SHOULD** = strongly preferred.

---

## Key Entities

| Entity | Description | Lifecycle |
|---|---|---|
| Kiro hook asset | A Kiro v3 hook definition (`.kiro/hooks/<name>.json`) emitted from a library source | created on `compile` when `kiro` + hooks selected |
| `library/kiro/hooks/` | New library source dir holding Kiro-native hook JSON (Approach A) | created by this feature |

---

## Success Criteria

- **SC-001 — Hooks parse:** 100% of emitted `.kiro/hooks/*.json` files parse as JSON and validate against the Kiro v3 schema. Measured by: a Go test that unmarshals each emitted hook and checks required fields.
- **SC-002 — Idempotent compile:** a second `compile` with no source change produces zero hook writes/drift. Measured by: golden test + plan no-op assertion.
- **SC-003 — Metadata honesty:** capability flags equal emitted surfaces for Kiro. Measured by: `capabilities_test.go` + golden output diff.

---

## Edge Cases

- **EC-001 — No hooks selected:** When a project selects `kiro` but no hook assets, then no `.kiro/hooks/` directory noise is created beyond what selection dictates (match peer-adapter behavior — only emit selected assets).
- **EC-002 — Unmappable canonical hook:** When a canonical hook has no source-verified Kiro v3 trigger, then it is skipped (not emitted with a guessed trigger) — FR-003.
- **EC-003 — Global scope:** When scope is global, hooks resolve under `~/.kiro/hooks/` via the same `ResolveToolRoot` path used by agents/skills today.
- **EC-004 — User-owned hook file drift:** When a user hand-edits an emitted hook, then the existing drift/`--force` model governs the rewrite (no silent clobber).

---

## Assumptions

- **A-001:** Kiro v3 hook JSON schema is `{"version":"v1","hooks":[{name,trigger,action:{type,command|prompt},…}]}` with 11 triggers — HIGH (source-verified via docs).
- **A-002:** Kiro command hooks can invoke a co-located script or inline command — MEDIUM (verify exact command-resolution semantics during implementation before finalizing the `command` value).
- **A-003:** Canonical agent frontmatter needs no Kiro transform — HIGH (verified).
- **A-004:** `output_mapping.go` needs no change because hooks are emitted in `Install`, not via `AssetKind` — HIGH (verified against peer adapters).

> A-002 is MEDIUM and MUST be verified (or restated as a risk) before the hook `command` field is finalized.

---

## Out of Scope

- Emitting `.kiro/specs/` — specs are user-authored `/spec new` artifacts.
- Emitting a repo-local Kiro `permissions.yaml` — docs forbid repo-injected permission rules.
- Emitting a direct `.kiro/powers/` directory — no documented auto-discovery path.
- Any Kiro agent frontmatter transform — canonical frontmatter is already valid Kiro v3.
- Inline per-agent `mcpServers`/`permissions` in Kiro agent profiles — optional future enhancement.
- A markdown→JSON hook transform engine (Approach B) — deferred; Approach A (pre-authored Kiro JSON) is the initial path.

---

## Clarifications

| Date | Question | Answer | Decided by |
|---|---|---|---|
| 2026-06-24 | Spec number? | 030 — highest in KNOWLEDGE_MAP Feature Specs table is 029; no collision. | agent |
| 2026-06-24 | Pre-author Kiro hook JSON (Approach A) vs markdown→JSON transform (Approach B)? | Approach A for this feature; B deferred. Mirrors Antigravity's pre-authored `hooks.json`. | agent (pending human gate) |
| 2026-06-24 | Flip Kiro `Permissions` to false or keep as metadata? | Keep `true` as host-support metadata + doc note (OMP precedent: capabilities.go documents metadata-only flags). | agent (pending human gate) |
| 2026-06-24 | Which canonical hooks map to Kiro triggers? | Only those with a source-verified trigger; mapping table fixed in plan.md and verified during implementation. | agent (pending human gate) |

---

## Constitutional Notes

- **Article I — Library-First:** reuses `CopyLibraryDirectory`, `ResolveToolRoot`, and the embedded library FS; no new infra.
- **Article IV — YAGNI:** deliberately omits specs, repo-local permissions, Powers output, agent transforms, and the Approach-B transform engine.
- **Article V — Simplicity:** Approach A (static JSON assets) chosen over a markdown-parsing transform engine; the simpler path emits verifiable bytes with no parser to maintain.

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-plan` | this file → `plan.md` |
| `speckit-tasks` | indirectly via plan → `tasks.md` |
