# Task 007 — Wizard selection steps + store persistence for opencode commands/modes

**Phase:** 3
**Status:** ✅ complete (2026-04-19)
**Depends on:** 006

## Implementation Notes

- **Types:** added `ALL_OPENCODE_COMMANDS` and `ALL_OPENCODE_MODES` slices; extended `WizardSelections` with `OpenCodeCommands` / `OpenCodeModes` JSON-tagged fields.
- **DB migration 003:** `ALTER TABLE selections` adds `opencode_commands TEXT NOT NULL DEFAULT '[]'` + `opencode_modes TEXT NOT NULL DEFAULT '[]'`. Down migration drops both. Applied additively so existing databases upgrade cleanly.
- **Store I/O** (`internal/db/store.go`): `readSelections` and `writeSelections` marshal/unmarshal the new columns. Roundtrip test extended in `store_test.go` to cover both fields.
- **Wizard** (`tui/wizard/phase2.go`):
  - `Phase2Result` + `phase2InteractiveState` + `buildPhase2Result` signature all gained the two new selection slices.
  - New steps 8 (`OpenCode Commands`) and 9 (`OpenCode Modes`) added to the interactive loop; only reachable under the `custom` preset. `phase2Total` returns 9 for custom, 4 otherwise.
  - `nextPhase2Step` skips from 5 → 10 (exit) for non-custom presets, preserving the existing "short" path.
  - New `askOpenCodeCommands` / `askOpenCodeModes` functions mirror the Gemini/Copilot pattern with back-option support.
- **Scaffold plumbing:** `ScaffoldContext.OpenCodeCommands` / `OpenCodeModes` added; `ScaffoldArtifacts` wires them into `adapter.AdapterSelections` so the adapter receives the user's selection.
- **`cmd/helpers.go`, `cmd/update.go`, `cmd/add.go`:**
  - Init flow defaults opencode selections to `ALL_OPENCODE_*` for non-minimal presets; `custom` preset honors the wizard's explicit list.
  - **Incidental fix:** `writeStoreFromScaffoldResult` also now persists `Selections.Commands` / `Selections.ChatModes` — these were set on `ScaffoldContext` but never written to the store (latent gap; `update.go`/`add.go` read them back but always saw empty). The same assignment now covers all four lists.
- **Tests:**
  - `store_test.go` roundtrip covers new fields.
  - `phase2_test.go` updated for the new 9-step total and adds subtest assertions for step 8 (`OpenCode Commands`) and step 9 (`OpenCode Modes`).
  - `TestBuildPhase2Result_CustomPreservesCommands` extended to assert opencode lists carry through on custom and are dropped on non-custom.

## Verification

- `go test ./... -count=1` — PASS
- `go vet ./...` — clean

## Flagged for Knowledge Map

- Incidental persistence fix for `Selections.Commands` / `Selections.ChatModes` is worth calling out in the final commit message so the follow-up from 5b5f185 is fully closed.

## Scope

Add two opencode-specific multi-select steps in the wizard (custom preset only, matching spec 010 pattern). Persist selections in the SQLite store.

## Changes

- `tui/wizard/`:
  - New step: "Select OpenCode commands to install" — multi-select, populated from embedded `library/opencode/commands/*.md`, shown only when opencode is selected AND preset == custom.
  - New step: "Select OpenCode modes to install" — same pattern.
- `internal/db/`:
  - Add `OpenCodeCommands []string` and `OpenCodeModes []string` to the selection schema (match existing `Commands`/`ChatModes` fields from spec 009 follow-up).
  - Store migration: additive; no schema break. Follow existing migration helper.
- `internal/scaffold/`:
  - Pipe new selections into adapter context so `CopyLibraryDirectory` receives them via `SelectionKey`.

## Tests

- `internal/db/store_test.go`:
  - Roundtrip: persist → reload → assert fields preserved.
- `tui/wizard/*_test.go`:
  - Selection persists across "custom" preset navigation.
  - Steps are skipped when preset != custom or opencode not selected.

## Definition of Done

- Custom preset offers both selection steps.
- Non-custom presets install the full bundle (no selection step).
- Store roundtrip green.
