# Go vs TS Parity Verification Report

## Summary
Using a shared fixture under `/tmp/ai-setup-compare-1eqgfz2s`, Go and TypeScript were compared across `setup --scan`, `--list`, `--dry-run`, and `--scan --adopt --import`, with normalization based on `specs/contracts/setup-engine/conformance/parity-rules.md` and `normalization.md`. `--scan` and all required `--dry-run` cases matched after path normalization, but `--list`, adopt/import behavior, wizard defaults, and MCP compilation still drift from Go. Two additional packaging/interface issues were found up front: the supplied Go artifact at `/tmp/ai-setup-go` was an `ar` archive rather than an executable, and neither CLI currently supports the user-requested `--target-dir` flag shape.

## 1. Scan Output Comparison
- Status: PASS
- Issues found:
  - No semantic drift found after normalization.
- Differences:
  - None. Normalized Go and TS `setup --scan` JSON matched structurally and semantically.
  - Shared paths, target detections, states, versions, observed files, and reusable agent scan output were identical for this fixture.

## 2. List Output Comparison
- Status: MISMATCH
- Issues:
  - TS omits the top-level `agents` array that Go emits.
- Differences:
  - Go `setup --list` included the reusable agent discovered at `<TARGET_DIR>/.ai/agents/foo`.
  - TS `setup --list` emitted targets/shared paths only and dropped agent inventory entirely.
  - This conflicts with the reusable-agent parity expectations implied by the scan/list contract and is visible in the normalized diff.

## 3. Dry-Run Comparison
- Status: PASS
- Issues:
  - No semantic drift found in the required cases.
- Differences:
  - `setup --dry-run`
  - `setup --dry-run --tool claude-code`
  - `setup --dry-run --global --tool opencode`
  - All three matched after normalization of absolute paths.

## 4. Adopt/Import Comparison
- Status: MISMATCH
- Issues:
  - TS reports `operation.mode` as `adopt+import`; Go reports `adopt-import`.
  - TS adopts and imports additional `workspace` resources that Go does not.
  - TS registry output includes empty `mcpEntries: []` fields that Go omits.
  - TS import destination hashes/IDs differ from Go for several global imports.
- Differences:
  - TS duplicated project-local imports into separate workspace entries for `claude-code`, `codex`, `copilot`, and `opencode`.
  - TS created extra workspace import directories under `<HOME_DIR>/.ai-setup/imports/...` that do not exist in Go output.
  - TS registry content under `setup-scan-registry.json` diverged structurally from Go because of extra workspace resources and empty `mcpEntries` arrays.
  - Per `normalization.md`, empty fields should be omitted; TS currently emits empty `mcpEntries` in persisted registry content.

## 5. Wizard Step Sequence
- Status: MISMATCH
- Issues:
  - Step order matches the contract.
  - Defaults do not fully match.
  - Available tool set does not fully match.
- Differences:
  - Step order matched the conformance contract: Scope → Tool targets → Skills → Agents → MCP preset → MCP servers → Project name → CLI tools → Project identity.
  - **Project name default mismatch**:
    - Go: `defaultPhase1ProjectName()` returns `"my-project"` (`tui/wizard/phase1.go:553-555`).
    - TS interactive flow defaults project name to `path.basename(opts.targetDir)` for non-global scope (`src/wizard/phase1-context.ts:709-727`).
  - **Pi tool mismatch**:
    - TS includes `pi` in phase-1 tool options and filters it by scope (`src/wizard/phase1-context.ts:59-66`, `118-126`).
    - Go phase-1 tool picker omits `pi` entirely (`tui/wizard/phase1.go:303-309`), despite parity rules saying Pi is supported for project/workspace scope.
  - Non-interactive error messaging also reflects the same drift: TS mentions `pi` in required `--tools` guidance; Go does not.

## 6. MCP Compilation
- Status: MISMATCH
- Issues:
  - TS JSON merge behavior is shallower and less conservative than Go for persisted MCP configs.
  - TS lacks Go’s Claude CLI MCP reconciliation path.
  - Orchestrator local-build path resolution is close but not identical.
- Differences:
  - **Claude local secrets / Copilot CLI merge behavior**:
    - Go uses `configmerge.MergeJSONFile(...)` for Claude local secrets and Copilot CLI config, preserving existing nested keys and creating backups on first touch.
    - TS uses a shallow `{ ...existing, ...patch }` merge in `mergeJsonFile(...)`, which can overwrite nested structures and does not create the same backup behavior.
  - **Claude MCP CLI reconciliation**:
    - Go attempts `claude mcp add-json` reconciliation before writing `.mcp.json` (`useCliForMCP` in `internal/adapter/mcp_compiler.go`).
    - TS writes `.mcp.json` directly and has no corresponding CLI registration path.
  - **Orchestrator managed server preparation**:
    - Both implementations auto-build `orchestrator/dist/index.js` if missing and can smoke-test it.
    - Go resolves `node` and `npm` via PATH lookups; TS uses `process.execPath` for node and raw `npm` for build steps. This is a low/medium environment-sensitive drift.
  - **OpenCode / Codex / Copilot / Claude overall shape**:
    - Core emitted shapes are broadly aligned by source review, but the merge/reconciliation differences above mean persisted behavior can still diverge.

## Gap Summary
| Area | Severity | Gap |
| --- | --- | --- |
| Packaging / invocation | high | Supplied `/tmp/ai-setup-go` artifact was not executable; it was an `ar` archive. |
| CLI interface | medium | Neither CLI currently supports the requested `setup --target-dir <dir>` invocation shape. |
| List output | high | TS omits reusable `agents` from `setup --list`. |
| Adopt/import mode | medium | TS emits `adopt+import` instead of Go’s canonical `adopt-import`. |
| Adopt/import scope handling | high | TS adopts/imports extra workspace resources not present in Go output. |
| Registry persistence | medium | TS persists empty `mcpEntries: []` fields and extra workspace records in registry output. |
| Wizard defaults | high | TS project-name default differs from Go (`basename(targetDir)` vs `my-project`). |
| Wizard tool availability | high | Go phase-1 wizard omits `pi`; TS includes it. |
| MCP merge behavior | high | TS shallow-merges persisted MCP JSON where Go deep-merges/preserves with backup behavior. |
| Claude MCP reconciliation | medium | Go uses Claude CLI reconciliation; TS does not. |
| Orchestrator path resolution | low | Go resolves `node` via PATH; TS pins to `process.execPath`. |

## Recommendations
1. Fix TS `setup --list` to emit reusable `agents` exactly like Go.
2. Align TS adopt/import pipeline with Go: canonical `operation.mode`, no extra workspace duplication, same registry omission behavior, same import identity/hash strategy.
3. Align wizard behavior with Go by either:
   - changing TS project-name default to `my-project`, and
   - adding `pi` to Go phase-1 tool options if the contract is correct,
   - or updating the contract only after Go is intentionally changed.
4. Replace TS `mergeJsonFile(...)` MCP persistence with Go-equivalent deep merge / backup semantics.
5. Decide whether TS should implement Go’s Claude CLI MCP reconciliation path or whether Go should stop doing it; today they are not behaviorally equivalent.
6. Fix the Go build packaging issue so `/tmp/ai-setup-go` is an actual executable in parity runs.
7. If `--target-dir` is intended to be supported, add it to both CLIs and the conformance contract; otherwise update test procedures to use working-directory scoping.

## Appendix: Raw Diffs

<details>
<summary><code>setup --list</code> diff</summary>

```diff
--- go
+++ ts
@@ -262,15 +262,5 @@
         }
       ]
     }
-  ],
-  "agents": [
-    {
-      "id": "foo",
-      "directory": "<TARGET_DIR>/.ai/agents/foo",
-      "promptPath": "<TARGET_DIR>/.ai/agents/foo/AGENT.md",
-      "status": "detected",
-      "title": "foo",
-      "description": "Test agent"
-    }
   ]
 }
```

</details>

<details>
<summary><code>setup --scan --adopt --import</code> key diff</summary>

```diff
--- go
+++ ts
@@ -506,7 +506,7 @@
   "operation": {
-    "mode": "adopt-import",
+    "mode": "adopt+import",
@@
+      {
+        "targetId": "claude-code",
+        "scope": "workspace",
+        "origin": "workspace",
+        "rootPath": "<TARGET_DIR>/.claude"
+      },
@@
+      {
+        "targetId": "codex",
+        "scope": "workspace",
+        "origin": "workspace",
+        "rootPath": "<TARGET_DIR>/.codex"
+      },
@@
+      {
+        "targetId": "copilot",
+        "scope": "workspace",
+        "origin": "workspace",
+        "rootPath": "<TARGET_DIR>/.github"
+      },
@@
+      {
+        "targetId": "opencode",
+        "scope": "workspace",
+        "origin": "workspace",
+        "rootPath": "<TARGET_DIR>/.opencode"
+      }
```

</details>

<details>
<summary>Registry / import tree drift</summary>

- TS created extra workspace import directories for Claude, Codex, Copilot, and OpenCode.
- TS registry entries included empty `mcpEntries: []` arrays and additional workspace resources.
- Import directory IDs for multiple global resources differed from Go.

</details>
