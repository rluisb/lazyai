# Plan: 030-kiro-cli-v3-output-gaps

**Feature ID:** 030
**Spec:** [./spec.md](./spec.md)
**Date:** 2026-06-24
**Status:** Draft
**Owner:** rluisb
**Constitution:** [../../docs/canonical/constitution.md](../../docs/canonical/constitution.md)

> **Purpose.** How LazyAI will emit Kiro CLI v3 native hooks and correct its capability/doc metadata, while recording the specs/permissions/powers non-goals.

---

## Summary

Add native Kiro v3 hook emission (`.kiro/hooks/<name>.json`) to the Kiro adapter, mirroring the peer-adapter pattern (pre-authored hook assets copied during `Install` via `CopyLibraryDirectory`). Flip the Kiro `Hooks` capability to `true`, update the conformance tests and golden fixtures, and correct the docs to mark hooks supported while recording the Kiro non-goals (specs, repo-local permissions, direct `.kiro/powers`). No agent transform and no `output_mapping.go` change are needed. Satisfies P1, P2, and the documentation slice of P3.

---

## Technical Context

| Aspect | Decision | Rationale |
|---|---|---|
| Language | Go 1.26 (`packages/cli`) | Existing adapter codebase |
| Framework | Cobra CLI + internal `adapter`/`library`/`writer` packages | No new deps |
| Storage | none (filesystem assets) | Hooks are static JSON files |
| Hook source | New `packages/cli/library/kiro/hooks/*.json` (Approach A) | Mirrors Antigravity's pre-authored `hooks.json`; no parser to maintain |
| Emission path | `KiroAdapter.Install` → `CopyLibraryDirectory` | Same path as agents/skills/prompts; hooks are not an `AssetKind` |

**External dependencies (new):** none.
**External dependencies (rejected):** a markdown→JSON hook transform engine (Approach B) — deferred (Article V).

---

## Constitution Check

| Article | Verdict | Justification |
|---|---|---|
| I — Library-First | PASS | Reuses `CopyLibraryDirectory`, `ResolveToolRoot`, embedded library FS. |
| II — Test-First (NON-NEGOTIABLE) | PASS | Each task writes/updates a failing test first: schema-validation test, capability test flip, golden regen, adapter Install assertions. |
| III — Docs as Source of Truth | PASS | Docs (README, capability-matrix, hooks.md, tool-outputs.md) updated in the same change. |
| IV — Anti-Speculation (YAGNI) | PASS | Specs, repo-local permissions, Powers output, agent transforms, and Approach B all explicitly omitted. |
| V — Simplicity Over Abstraction | PASS | Static JSON assets over a transform engine. |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | PASS | Reuses the existing emit path; no new abstraction; one new library dir + embed entry. |

**Verdict:** APPROVED (pending human gate).

---

## Project Structure

```
packages/cli/
├── internal/adapter/
│   ├── kiro.go                         ← modified: emit .kiro/hooks via CopyLibraryDirectory
│   ├── capabilities.go                 ← modified: Kiro Hooks: true + comment
│   ├── capabilities_test.go            ← modified: assert Hooks true; keep specs/steering rejected
│   ├── kiro_adapter_test.go            ← modified: assert .kiro/hooks/*.json; drop assertMissing(hooks)
│   ├── adapter_adapters_test.go        ← modified: drop assertMissing(.kiro/hooks)
│   └── kiro_hooks_test.go              ← NEW: Kiro v3 hook JSON schema validation test
├── library/
│   ├── kiro/hooks/                     ← NEW: Kiro-native hook JSON assets
│   │   └── block-destructive-shell.json (+ script if needed)
│   └── embed.go                        ← modified: add all:kiro to //go:embed
└── testdata/golden/
    ├── kiro-only/.kiro/hooks/          ← NEW golden output (UPDATE_GOLDEN=true)
    └── full-seven-targets/.kiro/hooks/ ← NEW golden output (UPDATE_GOLDEN=true)
docs/
├── README.md                           ← modified: drop "instruction-only" line
├── adapters/capability-matrix.md       ← modified: Kiro hooks = yes; non-goals
├── reference/hooks.md                  ← modified: Kiro = supported
└── reference/tool-outputs.md           ← modified: add hooks/ to Kiro tree
specs/030-kiro-cli-v3-output-gaps/      ← spec.md, plan.md, tasks.md
```

---

## Internal Contracts

| Contract | Producer | Consumer | Shape |
|---|---|---|---|
| `.kiro/hooks/<name>.json` | `KiroAdapter.Install` | Kiro CLI v3 runtime | `{"version":"v1","hooks":[{"name":str,"trigger":enum,"action":{"type":"command"\|"agent","command"\|"prompt":str},"description"?:str,"matcher"?:str,"timeout"?:int,"enabled"?:bool}]}` |

**Hook mapping (initial, verify each trigger during implementation — FR-003):**

| Library hook | Kiro trigger | Action type | Notes |
|---|---|---|---|
| `block-destructive-shell` | `PreToolUse` | `command` | Pre-tool guard; matcher targets shell-exec tool. Verify trigger + matcher semantics in Kiro v3 hook docs before emit. |
| `objective-workflow-gate` | *(verify)* | `command` | Only emit if a Kiro trigger (e.g. `UserPromptSubmit`/`Stop`) is source-verified; else skip per FR-003. |

Only the `block-destructive-shell` hook is committed to in P1; others emit only after their trigger is verified.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| Guessed trigger/matcher diverges from real Kiro v3 schema | M | H | FR-003: emit only source-verified triggers; schema-validation test; verify against live docs in Task 1.1 | implementer |
| `command` resolution semantics (script path vs inline) unknown (A-002) | M | M | Verify in Task 1.1; prefer self-contained inline command or co-located script proven runnable | implementer |
| Golden drift across both fixtures | M | M | Regenerate both via `UPDATE_GOLDEN=true`; review diff for only-hooks additions | implementer |
| Capability flag honesty (Permissions) misread | L | L | Document as host-support metadata (OMP precedent); no code flip | implementer |

---

## Complexity Tracking

| Item | Simpler alternative | Why complexity is justified | Cost |
|---|---|---|---|
| New `library/kiro/hooks/` dir + `embed.go` entry | Reuse `library/hooks/*.md` directly | Canonical hooks are markdown policy docs, not Kiro JSON; a static JSON asset is simpler than a runtime parser | one dir + one embed token |

---

## Phases & Milestones

| Phase | Goal | Exit criterion |
|---|---|---|
| 1 — Hook assets + schema test | Author Kiro v3 hook JSON + failing schema test | `kiro_hooks_test.go` validates the asset; trigger source-verified |
| 2 — Adapter emission | Emit `.kiro/hooks/*.json` in `Install`; embed the new dir | Adapter test asserts hook files exist; idempotent |
| 3 — Capability + conformance | Flip `Hooks: true`; update capability/adapter tests | `capabilities_test.go` green; specs/steering still rejected |
| 4 — Goldens | Regenerate `kiro-only` + `full-seven-targets` | `TestCompilerGolden` green |
| 5 — Docs | Mark hooks supported; record non-goals | README/capability-matrix/hooks.md/tool-outputs.md updated; `mkdocs build --strict` clean |

---

## Out of Scope

- `.kiro/specs/` emission — deferred (user-authored artifacts).
- Repo-local Kiro permissions — forbidden by docs.
- `.kiro/powers/` output — deferred to a future importable-bundle feature.
- Agent frontmatter transform — not needed.
- Approach-B markdown→JSON hook transform — deferred.

---

## Downstream Contract

| Produces for | Filename |
|---|---|
| `speckit-tasks` | this file + spec.md → `tasks.md` |
| `speckit-implement` | this file (technical context) + task harnesses |

---

## Approvals

| Role | Name | Date | Verdict |
|---|---|---|---|
| Author | agent | 2026-06-24 | drafted |
| Constitution check | agent | 2026-06-24 | APPROVED (pending human) |
| Human gate | — | — | ⛔ awaiting approval before implementation |
