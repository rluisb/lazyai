# Plan — 020: Runtime Parity Backport

**Date:** 2026-04-23
**Phase:** Plan (P of RPI) — awaiting HUMAN GATE before Implement
**Research:** `research.md` (decisions locked in §8)

---

## 1. Objective

Close the Go→TS parity gaps on the two "real bug" items identified in the audit:
1. **G1 — configmerge** — port the deep-merge + `.bak` backup module to TS, wire it into the OpenCode adapter's default-config write.
2. **G2 — opencode_validate** — port post-install CLI validation to TS, wire it into the install flow.
3. **G4 — JSONC helpers** — fold `parseJsonc` / `readJsoncFile` wrappers into the G1 work (free byproduct).

Plus one bug fix surfaced during research:
4. **MCP compiler server-merge bug** — TS `src/adapters/mcp-compiler.ts` currently replaces the entire `mcp` object on re-run, clobbering user-authored MCP server entries. Go's `mergeOpenCodeMcpServers` preserves them. Port the per-server merge logic.

Exit: `npx vitest run`, `npx tsc --noEmit`, and `go test ./... -count=1` all pass. TS `src/utils/configmerge.ts` exists and mirrors Go's exported surface; TS adapter and MCP compiler behavior match Go's on equivalent inputs.

---

## 2. Locked decisions (from research §8)

| # | Answer |
|---|---|
| Q1 | **Peer runtimes with strict mirror rule** — every feature added to one runtime must be added to the other. |
| Q2 | **Split by risk (c)** — this spec covers G1 + G2 + G4. Spec 021 (TBD) covers G3 (housekeeping) + T1 (migration parsers) + T2 (diff3). |
| Q3 | N/A — deferred to spec 021. |
| Q4 | N/A — deferred to spec 021. |
| Q5 | **Tests green + `tsc --noEmit` clean** — no cross-runtime behavioral diff in this spec. |

---

## 3. Out of scope

Explicitly excluded from spec 020 (all deferred to spec 021 or beyond):

- G3 housekeeping scaffold — needs separate scope decision on whether TS ships the full runtime consuming `sync-state.json` or just the scaffold
- G5 embedded library FS — Go-specific runtime concern; TS resolves via npm package layout, no port needed
- T1 migration parsers (claude/gemini/copilot) in Go
- T2 diff3 merging in Go
- Cross-runtime parity harness (byte-identical install outputs) — worth doing later as its own spec
- Any refactor/cleanup in either runtime not strictly required by the four items above

---

## 4. Phased breakdown

### Phase 1 — TS JSONC helpers + configmerge module

**Goal:** Mirror Go's `internal/configmerge/` and `internal/jsonc/` exported surface on the TS side with the same semantics (deep-merge, `.bak` on first touch, sorted-key idempotency, slice-wholesale-replace). No adapter wiring yet.

| Task | Deliverable | Est LOC |
|---|---|---|
| 001 | Extend `src/utils/jsonc.ts` — add `parseJsonc(input: string): Record<string, unknown>` (strips comments + `JSON.parse`) and `readJsoncFile(path: string): Record<string, unknown>` (reads from disk). Keep `stripJsonComments` unchanged. | ~15 |
| 002 | Create `src/utils/configmerge.ts` — export `mergeJsonFile(path: string, patch: Record<string, unknown>): { backupPath: string \| null }`. Internals: `deepMerge()` (recursive map merge, slices replace wholesale), `ensureBackup()` (create `.bak` sibling on first touch, never overwrite), `marshalSortedJson()` (sorted keys at every depth, 2-space indent, trailing newline). JSON-only — no TOML branch (OpenCode uses JSONC, no TOML consumers after spec 019). | ~120 |
| 003 | `src/__tests__/configmerge.test.ts` — mirror Go's test cases: new-file write (no `.bak`), preserves-user-keys + creates `.bak`, idempotent re-run produces identical bytes, slice replaced wholesale, `.bak` never overwritten on second run. Use `t.TempDir()`-equivalent via `mkdtempSync`. | ~140 |
| 004 | `src/__tests__/jsonc.test.ts` — add coverage for the new `parseJsonc` / `readJsoncFile` wrappers; keep existing `stripJsonComments` tests. | ~40 |

**Exit criteria:** `npx vitest run src/__tests__/configmerge.test.ts src/__tests__/jsonc.test.ts` passes. No production callers wired yet (module exists in isolation).

---

### Phase 2 — Wire configmerge into OpenCode adapter

**Goal:** Replace the current direct `writeFile(jsoncPath, JSON.stringify(defaultConfig, null, 2))` with `mergeJsonFile(jsoncPath, defaultConfig)` so the adapter matches Go's call path at `internal/adapter/opencode.go:87`.

| Task | Deliverable | Est LOC |
|---|---|---|
| 005 | `src/adapters/opencode.ts` — swap the default-config write in the "neither json nor jsonc exists" branch to use `mergeJsonFile`. Adjust the `fileRecords.push` to use the returned hash as before. Remove now-redundant `JSON.stringify` call. Preserve existing behavior: only writes when neither `opencode.json` nor `opencode.jsonc` exists. | ~15 |
| 006 | Update `src/__tests__/adapters-files.test.ts` — adjust the "OpenCode adapter installs agents and force-overwrites" test to verify the file-record hash is still populated correctly via the new path. No new assertions; ensure existing ones still pass. | ~5 |

**Exit criteria:** `npx vitest run` passes across the full suite (all 33 test files). Default config write behavior identical to pre-change; diff matches byte-for-byte.

---

### Phase 3 — Fix MCP compiler server-merge bug

**Goal:** Port Go's `mergeOpenCodeMcpServers` (internal/adapter/mcp_compiler.go:127-141) to TS so user-authored MCP servers are preserved on re-run.

| Task | Deliverable | Est LOC |
|---|---|---|
| 007 | `src/adapters/mcp-compiler.ts` — replace the shallow spread (`{ ...existingConfig, $schema: ..., mcp: ocMcpContent }`) with Go-equivalent logic: mutate `existingConfig` in place, set `$schema`, and set `mcp` to `mergeOpenCodeMcpServers(existingConfig.mcp, ocMcpContent)`. Add the helper function: iterate existing entries, keep ones NOT in the managed set, then overlay managed entries (managed wins on collision). | ~40 |
| 008 | `src/__tests__/mcp.test.ts` — add a test for the preservation behavior: pre-seed `.opencode/opencode.jsonc` with a user-authored `mcp.custom-server` entry and a user-authored `permission` top-level key; run `compileMcp` with a catalog containing a different managed server; assert both the user server AND the managed server end up in output, and `permission` is preserved. | ~50 |

**Exit criteria:** New test passes. Existing mcp.test.ts cases still pass. Behavior on identical inputs matches Go's `compileOpenCodeMCP` output for the MCP map shape.

---

### Phase 4 — Post-install validation port (G2)

**Goal:** Port `internal/adapter/opencode_validate.go` to TS and hook it into the install flow.

| Task | Deliverable | Est LOC |
|---|---|---|
| 009 | Create `src/adapters/opencode-validate.ts` — export `ValidationWarning` interface (`scope: 'config' \| 'agent'`, `item: string`, `reason: string`) and `validateOpenCodeInstall(ctx)` function. LookPath-gate with `which opencode` (use existing pattern from `src/adapters/opencode.ts:130`). Run `opencode debug config` → warn if fails OR output missing "mcp" substring when config file exists. For each `.md` file in `{ocDir}/agents/`, run `opencode debug agent <name>` → warn on non-zero exit. Accept an injectable `CmdRunner` type for test isolation. | ~80 |
| 010 | Hook into install flow — in `src/wizard/index.ts`, add a post-install step after the last `compileMcp` call that (a) only runs when `opencode` is in the tool list and (b) logs each warning via `console.warn`. Mirror Go's `internal/scaffold/scaffold.go:95-110` pattern. Non-fatal on all errors (never propagate). | ~25 |
| 011 | `src/__tests__/opencode-validate.test.ts` — mirror Go's 4 test cases (`BinaryAbsent`, `ConfigFails`, `AgentFails`, `AllPass`). Use the injectable `CmdRunner` to avoid real `exec` calls. | ~130 |

**Exit criteria:** New test file passes. Full install flow smoke-test (via existing `cli.test.ts` or `e2e.test.ts`) doesn't regress — validation step is non-fatal even when opencode binary is absent (the common case in CI).

---

### Phase 5 — Verification

**Goal:** Confirm all exit criteria before declaring spec done.

| Task | Deliverable |
|---|---|
| 012 | Run `npx vitest run` — expect 33+ test files passing, all tests green. |
| 013 | Run `npx tsc --noEmit` — expect zero output. |
| 014 | Run `go test ./... -count=1` — expect all Go packages passing (should be unchanged but verify no accidental regression). |
| 015 | Update `specs/KNOWLEDGE_MAP.md` — add spec 020 row to the feature table; add the new `src/utils/configmerge.ts` and `src/adapters/opencode-validate.ts` entries to the Packages Reference table. |

**Exit criteria:** All three test suites green. KNOWLEDGE_MAP reflects the new state.

---

## 5. Acceptance Criteria (summary)

- [ ] `src/utils/configmerge.ts` exists and exports `mergeJsonFile`
- [ ] `src/utils/jsonc.ts` exports `parseJsonc` and `readJsoncFile` in addition to `stripJsonComments`
- [ ] `src/adapters/opencode-validate.ts` exists and exports `validateOpenCodeInstall` + `ValidationWarning`
- [ ] `src/adapters/opencode.ts` default-config write uses `mergeJsonFile`
- [ ] `src/adapters/mcp-compiler.ts` preserves user-authored MCP entries on re-run (matches Go's `mergeOpenCodeMcpServers` behavior)
- [ ] `src/wizard/index.ts` calls `validateOpenCodeInstall` after install; warnings logged non-fatally
- [ ] `npx vitest run` passes
- [ ] `npx tsc --noEmit` clean
- [ ] `go test ./... -count=1` still passes
- [ ] `specs/KNOWLEDGE_MAP.md` updated

---

## 6. Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Sorted-key JSON output differs subtly from Go's `encoding/json` marshaling (e.g., Unicode escape differences) | Tests pass but byte-identical output between runtimes fails a future parity harness | Use `JSON.stringify` with a keys-sorted replacer; accept that we deferred the byte-diff test (Q5=a). Note divergence if any. |
| `validateOpenCodeInstall` adds real `execSync` calls that slow down CI | Test runtime increase | LookPath-gate means it no-ops in CI (binary absent). Existing CI should not invoke real opencode. Verify by running the TS test suite and timing. |
| The MCP compiler fix changes observable output format (existing users on re-run get different JSON) | User-visible diff on first re-run after upgrade | The fix only affects what's preserved (additive), not what's written. Users who never touched `.opencode/opencode.jsonc` manually see no change. Document in commit message. |
| `.bak` sidecar created for first time on existing installs pollutes repos that don't gitignore it | User-visible file appears in git status | Configmerge only creates `.bak` when file already exists and is being modified. In the current single call site, that gate is `!files.fileExists(jsoncPath)` — so `.bak` is never created during normal flow. Document this, and add `.bak` pattern to the scaffold-emitted `.gitignore` in a follow-up if we ever wire configmerge into a mutating call path. |

---

## 7. Task sequence summary

15 tasks total across 5 phases:

1. Phase 1 (001–004): JSONC + configmerge module + tests (~315 LOC new, ~180 test LOC)
2. Phase 2 (005–006): Wire into adapter (~20 LOC)
3. Phase 3 (007–008): MCP compiler bug fix + test (~90 LOC)
4. Phase 4 (009–011): Validation port + hook + tests (~235 LOC)
5. Phase 5 (012–015): Verification + docs

**Total:** ~660 LOC production + ~500 LOC tests. Comparable scope to Go's side (656 + 669 from the audit).
