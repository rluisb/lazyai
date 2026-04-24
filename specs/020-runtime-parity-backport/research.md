# Spec 020: Runtime Parity Backport — Research

**Date:** 2026-04-23
**Status:** Research — awaiting human gate before Plan
**Predecessor:** Spec 019 (OpenCode-only consolidation) — both runtimes now target OpenCode exclusively
**Goal:** Identify and assess bidirectional feature gaps between the Go and TypeScript codebases so we can decide what to backport, reverse-port, or accept as permanent divergence.

---

## §1 — Context

After spec 019, both runtimes support only OpenCode. A follow-up audit confirmed scope alignment but surfaced capability drift in both directions:

- **Go-only:** 5 capabilities (configmerge, opencode_validate, housekeeping, jsonc helpers, embedded library FS)
- **TypeScript-only:** 2 capabilities (claude/gemini/copilot migration parsers, diff3 merging)

The TypeScript codebase is still actively shipped (npm `@ricardoborges-teachable/ai-setup` binary published from `bin/ai-setup.js`), so parity is not purely academic — users pick whichever entry point matches their install pathway.

---

## §2 — Go-only capabilities

### G1 — `internal/configmerge/` (245 LOC + 246 LOC tests)

**What it does:** Deep-merge JSON/JSONC/TOML patches into existing config files, writing a `.bak` sidecar on first touch, with idempotent (sorted-key) output.

**Key semantics:**
- Maps recurse; patch wins on leaf collisions
- Slices replace wholesale (prevents duplicate accumulation on re-run)
- `.bak` created once, never overwritten (preserves original forever)
- Deterministic output via sorted keys (JSON and TOML)

**Call sites in Go:**
- `internal/adapter/opencode.go:87` — writing default `opencode.jsonc`

**TS equivalent (partial):**
- `src/adapters/mcp-compiler.ts` uses **shallow spread** (`{ ...existingConfig, $schema: ..., mcp: ocMcpContent }`)
- `src/adapters/opencode.ts` writes default config **only when absent** (no merge at all)
- No `.bak` sidecar anywhere in the write path
- TS migration flow (`src/migration/plan.ts`) writes backups under `.ai-setup-backup/`, but that's a separate migration concept, not runtime safety

**Actual TS-side risk:** MCP compiler clobbers user-authored nested `mcp.<server>` entries on re-run, because the shallow spread replaces the entire `mcp` object. Go's deep-merge would preserve them.

**Porting cost:** ~150 LOC (JSON-only; TOML branch unneeded since no TOML configs in OpenCode-only world). Can wrap `jsonc-parser` (already transitively available via TS ecosystem) or use existing `stripJsonComments`.

---

### G2 — `internal/adapter/opencode_validate.go` (84 LOC + 153 LOC tests)

**What it does:** Post-install sanity check via the `opencode` CLI. Runs `opencode debug config` and `opencode debug agent <name>` for each installed agent; collects non-fatal `ValidationWarning` entries; LookPath-gated so it no-ops when binary is absent.

**Call sites in Go:**
- `internal/scaffold/scaffold.go:104` — end of install flow, warnings logged

**TS equivalent:** None. TS runs no post-install CLI probes.

**Value:** Catches "MCP entries didn't register" and broken agent frontmatter at install time instead of first use. Purely additive UX.

**Porting cost:** ~60 LOC. Needs an `execSync` / `spawn` wrapper + LookPath gate. TS already has `which opencode` pattern in `src/adapters/opencode.ts:130` for plugin install.

---

### G3 — `internal/scaffold/housekeeping.go` (55 LOC, no direct tests)

**What it does:** Scaffolds `.ai/housekeeping/sync-state.json` with a v1 schema tracking QMD/CodeGraph drift status, ACK'd stale entries, and pending repair proposals. Only runs when `ctx.Housekeeping` config is set.

**Call sites in Go:**
- `internal/scaffold/scaffold.go:48` — early in install flow

**Dependencies:**
- `types.HousekeepingConfig` struct with `MemoryPath`, `EnableQmd`, `QmdIndexPath`, `EnableCodegraph`, `CodegraphDataPath` fields

**TS equivalent:** None. `HousekeepingConfig` type doesn't exist in `src/types.ts`. The wizard has no housekeeping phase.

**Value:** Unclear without human context. This scaffolds **state tracking for external integrations** (QMD search index + CodeGraph). If no TS user has those external tools, the empty sync-state.json accomplishes nothing.

**Porting cost:** ~80 LOC (file emitter + type + wizard-phase hook + CLI flag wiring). But the cost scales with how deep the integration goes — if housekeeping is meant to include the drift-detection runtime that writes to this file, it's a much larger lift (spec 006 in Go is "Housekeeping, memory, and bootstrap" — presumably a multi-phase spec).

---

### G4 — `internal/jsonc/` (96 LOC + 133 LOC tests)

**What it does:** Provides `StripComments`, `ParseJSONC`, `ReadJSONCFile`. Header comment says *"Ported from the TypeScript utilities in src/utils/jsonc.ts."*

**TS equivalent:** `src/utils/jsonc.ts` has `stripJsonComments` — the string→string primitive. Missing the `ParseJSONC` (string→object) and `ReadJSONCFile` (path→object) wrappers.

**Porting cost:** ~10 LOC — two tiny wrappers. Not worth a dedicated task; roll into G1.

---

### G5 — `internal/library/embed.go` (176 LOC + 137 LOC tests)

**What it does:** Resolves the library directory, switching between dev-mode (walks up from cwd looking for `library/mcp/catalog.json`) and production-mode (uses `go:embed`-populated `fs.FS`).

**Why Go needs it:** Go binaries are standalone — they have no concept of "npm package resources". The embed tag bakes `library/` into the binary; runtime decides whether to read from disk or the embedded FS.

**TS equivalent status:** TS doesn't need this. The `package.json` declares `"files": ["dist", "bin", "library"]`, so npm ships `library/` as files under `node_modules/@ricardoborges-teachable/ai-setup/library/`. TS resolves it relative to the installed package location via `import.meta.url` or `__dirname`. A single path-resolution helper (if not already present) is all that's needed, not an embed abstraction.

**Porting cost:** **Don't port.** This is a runtime-specific concern Go handles because it must. TS should keep its current npm-native resolution.

---

## §3 — TypeScript-only capabilities

### T1 — Migration parsers for Claude / Gemini / Copilot (`src/migration/parsers/`)

**What it does:** Reads existing `~/.claude/`, `~/.gemini/`, `.github/copilot-*` configs and translates them into the canonical `.ai/` format. Called by `ai-setup migrate` and `ai-setup import`.

**Key insight from spec 019 summary:** These were deliberately kept on the TS side because *importing from* other tools is a legitimate feature even when we don't *install for* those tools. The user may have pre-existing Claude/Gemini/Copilot setups they want to consolidate into OpenCode-compatible form.

**Go equivalent:** `internal/migration/parser.go` — only `parseOpenCodeSetup()`.

**Decision needed:** If a user on the Go binary runs `ai-setup migrate` against an existing Claude/Gemini/Copilot project, does that flow work? If not, this is a real gap on the Go side.

**Porting cost (TS→Go):** Significant — ~200-400 LOC per parser depending on complexity. Would need to re-implement YAML frontmatter extraction, directory walking, and per-tool format mapping in Go.

---

### T2 — Diff3 merging (`src/migration/diff/diff3.ts`)

**What it does:** Three-way merge algorithm (base + ours + theirs) for migration conflict resolution.

**Go equivalent:** Simple two-way diff in `internal/diff/`.

**Decision needed:** How often does the migration flow actually hit a three-way conflict that a two-way diff can't handle? If rare, skip. If common, Go needs diff3.

**Porting cost (TS→Go):** ~200 LOC for the algorithm + wiring into the migration plan executor.

---

## §4 — Alignment confirmed (no action needed)

All of these were confirmed aligned by the audit agent:

| Subsystem | Go | TS |
|---|---|---|
| CLI commands (16 subcommands) | ✓ | ✓ |
| Scope resolution | ✓ | ✓ |
| Global paths abstraction | ✓ | ✓ |
| SQLite store + migrations | ✓ | ✓ (lowdb-backed equivalent) |
| Orchestrator catalog | ✓ | ✓ |
| Compiler (fragment + template) | ✓ | ✓ |
| Conflict strategy | ✓ | ✓ |
| Frontmatter parsing | ✓ | ✓ |
| Error boundary | ✓ | ✓ |
| Manifest | ✓ | ✓ |
| Presets | ✓ | ✓ |

---

## §5 — Options matrix

Four coherent scopes the human could pick from. Each row is a self-contained option; pick one (or mix).

| Option | Work | Value | Risk |
|---|---|---|---|
| **A. Safety-only (minimal)** | Port G1 (configmerge with `.bak` + deep-merge) to TS. Fix the MCP-compiler clobber bug as a side effect. | High — closes a real data-loss bug. | Low — well-scoped, ~150 LOC. |
| **B. Safety + validation** | A + port G2 (post-install validation). | A's value + nicer UX on TS installs. | Low — G2 is isolated and non-fatal. |
| **C. Full Go→TS backport** | A + B + assess G3 (housekeeping) scope; skip G4 (fold into A) and G5 (don't port). | Complete catch-up on user-facing Go features. | Medium — G3 may balloon if integration is deep. |
| **D. Bidirectional** | C + port T1 (migration parsers) to Go. | Full symmetry. | High — T1 is 3× parsers, each non-trivial. Probably warrants its own spec. |

---

## §6 — Questions for human gate

Before committing to a plan, please answer:

- **Q1: What's the strategic status of the TypeScript codebase?**
  - (a) Peer implementation — keep at parity indefinitely
  - (b) Legacy — maintain passing tests but don't invest
  - (c) Deprecating — migration path exists; avoid new TS investment

- **Q2: Which option from §5 do you want?** (A, B, C, or D)

- **Q3 (only if Q2 includes G3 housekeeping):** Is the TS side expected to ship the *full* housekeeping runtime (QMD/CodeGraph drift detection + repair) or just the scaffold file? The scaffold alone is pointless without the runtime that consumes it.

- **Q4 (only if Q2 = D):** Does the Go binary need to support `migrate`/`import` for Claude/Gemini/Copilot sources, or is that strictly a TS-side user flow? (If strictly TS, Go doesn't need T1.)

- **Q5: Acceptance criteria for "parity" — when do we call this done?**
  - (a) Tests green + `tsc --noEmit` clean
  - (b) Tests green + manual install smoke-test on both runtimes produces byte-identical outputs for the same inputs
  - (c) Something stricter

---

## §7 — Recommendation

**Default to Option A (safety-only).**

The configmerge backport is the only item with a concrete, demonstrable bug (MCP compiler clobbers user-authored server entries on re-run). Everything else is polish. If the user is shipping the TS binary to anyone who customizes their `opencode.jsonc`, A is load-bearing. Bundle G4's tiny JSONC wrappers into A since they fall out of the same refactor for free.

B is a reasonable stretch if the user wants validation parity too — it's isolated and cheap.

C and D warrant their own follow-up specs rather than bundling here; G3's scope is ambiguous and T1 is 3× the size of A–B combined.

---

## §8 — Decisions (filled in at human gate)

| # | Answer |
|---|---|
| Q1 | **Peer** — Go and TS are two applications with the same features and logic. Users install/use either; both MUST reflect each other. Every feature added to one runtime must be added to the other in the same change set. |
| Q2 | **Split by risk (c)** — Spec 020 closes the "real bug" items (G1 configmerge + G2 validation, port Go→TS). Structural items (G3 housekeeping, T1 migration parsers, T2 diff3) deferred to spec 021. |
| Q3 | Deferred to spec 021 (housekeeping out of scope for 020). |
| Q4 | Deferred to spec 021 (migration parsers out of scope for 020). |
| Q5 | **Tests green + `tsc --noEmit` clean (a)** — `go test ./... -count=1`, `npx vitest run`, and `npx tsc --noEmit` all pass. No cross-runtime behavioral diff this spec; a dedicated parity-harness spec can come later if needed. |
