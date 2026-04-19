# Task 007 — Wizard selection steps + store persistence for opencode commands/modes

**Phase:** 3
**Status:** pending
**Depends on:** 006

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
