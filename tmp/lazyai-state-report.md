# LazyAI / vibe-lab — Current Implementation State

Inspection-only report. Repo root: `/Users/ricardo/projects/teachable/lazyai`, branch `main`, HEAD `f144afd`. Evidence cited as `path` / `path:symbol`.

---

## 1. What LazyAI is, today (proven by code)

LazyAI **is** a Go CLI (`lazyai-cli`) that manages a canonical `.ai/` source tree and **compiles it into tool-native AI-harness surfaces** for **exactly 7 host tools**. It is *not* an agent runtime — it emits configuration/assets that host tools (OpenCode, Claude Code, Copilot, Pi, OMP, Gemini/Antigravity, Kiro) then execute.

- Single Go module, `go install`, no npm/npx for normal use (`README.md:18`, `go.work`, `packages/cli/go.mod`).
- Two packages: `packages/cli` (product) and `packages/diffviewer` (companion review tool) (`go.work`).
- 7 targets hard-registered: `OpenCodeAdapter, ClaudeCodeAdapter, CopilotAdapter, PiAdapter, OmpAdapter, KiroAdapter, AntigravityAdapter` (`packages/cli/internal/adapter/registry.go:24-30`). Codex explicitly **rejected** (`internal/aimanifest/aimanifest.go`; spec 029 "Codex deprecated").

Documented boundary matches code (`docs/concepts/product-boundaries.md`):
> "LazyAI is the runtime and product. It owns the `lazyai-cli` binary, the canonical `.ai/` source model, the adapter/compiler path… vibe-lab supplies principles, assets, and adapter expectations; it is not a runtime dependency of LazyAI."

---

## 2. The compile contract (V2, manifest-driven) — implemented

| Component | File:symbol | What it does |
|---|---|---|
| Manifest `.ai/lazyai.json` | `internal/aimanifest/aimanifest.go` | Schema `version "1.0"`; `targets` (enum of 7), `profile` (`personal`/`team`), `adapters`, `source`, `safety`. Rejects `codex`, filters `EnabledTargets`. Schema: `internal/schema/lazyai.schema.json`. |
| Lockfile `.ai/lock.json` | `internal/lockfile/lockfile.go` | Version `1.0`; `LazyaiVersion`, `CompiledAt`, per-adapter records, SHA-256 `sourceHash`/`outputHash`/`managed`. Schema: `lock.schema.json`. |
| Managed-region writer | `internal/writer/writer.go` | `plan → write → lock`; idempotent skip on hash match; drift refusal unless `--force`; `--dry-run` writes nothing; atomic temp+rename; merges via `adapter.MergeManagedBlock`. |
| Compile entrypoint | `cmd/compile.go` | Manifest-first selection; contract + agent-model validation gates; adapter dispatch; beta annotation; `--tool`, `--dry-run`, `--validate-contracts`. |
| Compiler internals | `internal/compiler/{contract_validator,template,fragment,agent_validate}.go` | Cross-file frontmatter contract checks; template discovery; fragment include/conditional/variable expansion; compile-time agent-model validation. |

### Adapter capability matrix (`internal/adapter/capabilities.go`)

| Target | Support | MCP output path | Notes |
|---|---|---|---|
| opencode | stable | `.opencode/opencode.json` (legacy `.opencode/lazyai.mcp.jsonc` migrated in) | full |
| claude | stable | `.mcp.json` / `.claude/settings.local.json` / global CLI | full |
| copilot | stable | `.vscode/mcp.json`, `~/.copilot/mcp-config.json` | repo+path instructions, chatmodes |
| pi | stable | **no-op (MCP not emitted)** | skill-first |
| omp | **beta** | `.omp/mcp.json` | beta until docs snapshot (EC-006) |
| kiro | stable | `.kiro/settings/mcp.json` | omits agents/skills by design |
| antigravity | **beta** | `~/.gemini/config/mcp_config.json` + `.gemini/`, `.agents/skills` | beta (EC-006) |

Beta set pinned by `TestBetaAdaptersAreOmpAndAntigravity` (`capabilities_test.go:39`). Layouts table-driven in `internal/adapter/output_mapping.go`. MCP fan-out `internal/adapter/mcp_compiler.go`; canonical `.ai/mcp.json` scaffolded by `internal/scaffold/mcp.go`.

---

## 3. CLI command surface — verified against README

**35 visible root commands**, categories enforced by `cmd/command_category.go` + tests. README counts (`README.md:87-140`) accurate:

- **setup-core (21):** add, build-plugin, compile, completion, config, create, doctor, eject, import, info, init, list, migrate, server, setup, sidecar, status, update, update-self, validate, workspace
- **ops-runtime-extra (12):** auth, backup, cost, git, ledger, memory, message, metrics, notify, restore-runtime-db, secret, session
- **dev-harness (1):** models (`models sync`)
- **retired/archived (1):** completions (hidden alias of completion)

Excised (`task`, `workflow`, `orchestration`, `mcp-setup`, `eval`) asserted absent by `command_excision_test.go`.

Depth: setup-core real (pipelines above). `doctor` = integrity + health + security report (`cmd/doctor.go`, `doctor_security.go`, `doctor_health.go`). ops-runtime-extra genuinely DB/file-backed: session/message/metrics/cost via `internal/db`+`internal/runtime`; `ledger` hash-chained; `memory` → `.specify/memory`; `secret` OS keychain+file fallback; `backup`/`restore-runtime-db` tar `.specify/session.db`.

---

## 4. Embedded library (shipped asset pack)

`packages/cli/library/` `go:embed`-compiled (`library/embed.go`), ~23 subtrees. Curation manifest (`library/manifests/curation.yaml`, v1): **151 active + 22 excluded** entries, each with path/kind/category/adapter_targets/rationale/token_rent flags; provenance in `provenance.yaml`; validated by `internal/library/asset_manifest.go`.

| Category | Count | Examples |
|---|---|---|
| skills/ | ~51 | rpi, tdd-loop, tdd-planning, plan, implement, iterate, research, diagnose, review, chain-verify, red-team-plan, anti-speculation, handoff, parallel-execution, codebase-exploration, four-point-vibe-coding, no-workarounds, caveman, self-improve, memory-write/-promotion, speckit-* |
| canonical/agents/ | 8 | guide (front door) + implementer, researcher, planner, reviewer, deployer, responder, evidence-verifier |
| rules/ | 14 | agent-state, workflow, security, testing, tool-use, code-style, access, typescript, self-consistency, cost, review, structured-feedback, auto-recovery, agent-security |
| fragments/ | 12 | rpi-workflow, reasoning-protocol, quality-gates.xml, bug-resolution.xml, git-safety, git-conventions.xml, decision-protocol, context-discipline, harness-protocol, workspace-protocol, system-context.xml, agent-harness |
| standards/starter/ | 4 | agent-security, context-loading, error-handling, test-patterns |
| templates/ | ~17 | specs, tasks, plans, ADRs, audits, reviews, postmortems, policies, constitutions, hooks |
| hooks/ | 6 | pre-commit, rpi-gate-check.yml, caveman-memory-promotion.md, startup-self-heal.md, block-destructive-shell.md, objective-workflow-gate.md |
| mcp/catalog.json | 5 | ai-memory, filesystem, ripgrep, codegraph, obsidian (+ CLI equivalents) |
| tool surfaces | — | copilot/, claudecode/ (settings, output-styles, hook scripts), opencode/ (plugins.json, modes), pi/, omp/, antigravity/ (hooks.json/settings.json), root/ (AGENTS.md, copilot-instructions.md templates) |

vibe-lab quality/process layer ships as assets: RPI (`fragments/rpi-workflow.md`, `skills/rpi.md`, hook `rpi-gate-check.yml`), anti-slop/clean-code (`skills/anti-speculation.md`, `no-workarounds.md`, `canonical/clean-code.md`), human gates (`objective-workflow-gate.md`), memory/handoff, parallel-agent guidance, eval rubric/case schemas.

---

## 5. vibe-lab layer & dogfooding

- `canonical/`: constitution.md, clean-code.md, engineering-principles.md, prd-plan-todo.md, tdd-planning.md, clarification-levels.md, templates, speckit-vibe-lab-preset/.
- `.agents/`: agent defs (researcher/planner/implementer/reviewer/deployer/responder/evidence-verifier) + vibe-lab skills (four-point-vibe-coding, create-skill/-hook/-workflow).
- `.ai/`: repo's own canonical sources (mcp.json, constitution/, housekeeping/).
- `docs/workflows/{feature,documentation,code-review}.md`: four-point framing, TDD modes, human exit gates.

Evals = **schema + shape validation only** (`internal/evals/validate.go` + `eval-{case,holdout,rubric}.schema.json`); no scoring/judge engine (spec marks evals optional).

---

## 6. ADRs / spec status

ADRs (`specs/adrs/`): 001 TS-bootstrap→Go (superseded), 002 agent-memory-kernel (proposed), 003 LazyAI-runtime-boundary (accepted), 004 vibe-lab-alignment-contract (accepted), 005 core-vs-optional-modules (accepted), 006 manifest-compile+7-target (accepted). Active: refactor 026-vibe-lab-alignment, feature 029-lazyai-v2.

Proposed product/techspec to compare against is in-repo: `docs/lazyai-vibelab-product-spec-pack/.../` (01_PRD, 02_TECHSPEC, 03_OFFICIAL_TOOL_COMPLIANCE_MATRIX, 04_SCHEMAS_AND_EXAMPLES, 05_IMPLEMENTATION_ROADMAP). Grounded current-vs-target table: `specs/029-lazyai-v2/spec.md:25-39`.

---

## 7. Gaps & discrepancies

**Documented but not implemented / not as specified:**
- **FR-008 fixtures**: spec wants conformance fixtures under `packages/cli/testdata/` (`029/spec.md:90`), which **does not exist**. Assertions exist inline instead (e.g. `opencode_adapter_test.go:544-570` rejects deprecated `tools`/`maxSteps`). Partially met, mislocated.
- **029 status drift**: `packages/cli/KNOWLEDGE_MAP.md:22` = Done; `specs/029-lazyai-v2/spec.md:6` = Draft.
- **ADR 002 agent-memory-kernel**: proposed; no kernel in code (`memory` cmd is file/DB vault).
- **Trace/eval**: only schema+shape validation; no judge/scoring.
- **Pi MCP**: deliberate no-op (`mcp_compiler.go`, `pi.go`).
- **`server` L3 handshake**: not implemented in Go ("requires Node.js MCP SDK", `cmd/server.go:622`); L1 config checks only.
- **`init` headless populate**: AGENTS.md placeholder fill skipped (`cmd/init.go:215,392`); needs running host AI tool.

**Implemented but lightly documented:**
- Full ops-runtime-extra runtime (SQLite migrations `internal/db/migrations.go`, runtime schema `internal/runtime/schema.go`) — framed only as "transitional extras."
- `internal/tokenrent` (library byte-budget) + `internal/minimality` (minimality report).
- `internal/reversa` (deterministic Scout prefill, `cmd/helpers.go:233`), `internal/setupscan`/`absorb` (adopt native setups), `internal/sidecar` (external docs/specs workspace).
- `models sync` (`internal/models/sync.go`, models.dev) and multi-target `build-plugin` bundles (claude, copilot-cli, omp, pi — `internal/plugin/bundles.go`).

**Cross-tool harness-manager assessment:** Core delivered — canonical source, manifest+lock idempotency, managed-region safe writes, 7 native adapters, MCP fan-out, validate/doctor/migrate/eject, plugin bundles. Gaps vs "complete": (1) MCP parity (Pi no-op), (2) two beta adapters pending docs-snapshot, (3) evals validation-only, (4) conformance fixtures mislocated, (5) no headless populate (by design — not an agent runtime).

---

## 8. Security note (no values exposed)

Local state files (paths/types only): `.ai-setup.db`, `.specify/session.db`(+`.backup`) SQLite DBs; `.gemini/settings.json`(+`.bak`); `.mcp.json` / `.vscode/mcp.json` / `.ai/mcp.json` (MCP env may use `${VAR}`). Product has a secret scanner (`internal/validate/secrets.go`) and `doctor` security report (`cmd/doctor_security.go`). No secret values read or surfaced.
