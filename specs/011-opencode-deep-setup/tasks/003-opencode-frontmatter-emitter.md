# Task 003 — OpenCode-specific agent frontmatter emitter

**Phase:** 2
**Status:** ✅ complete (2026-04-19)
**Depends on:** 002

## Implementation Notes

- New file `internal/adapter/opencode_frontmatter.go` provides `BuildOpenCodeAgentFrontmatter(source, opts)` plus the `OpenCodeAgentOpts` struct (`Description`, `Mode`, `Tools map[string]bool`, `Model`, `Permission map[string]string`).
- Emitter behavior: strips existing frontmatter, inherits `description` from source `name` (fallback: source `description`, then "Agent"), defaults `mode: all`, omits optional keys when nil/empty, escapes the description as a double-quoted YAML scalar, emits map keys in sorted order for byte-stable output.
- Dropped source keys explicitly (verified by `TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys`): `name` (opencode derives name from filename), `tools` (source holds MCP-server names — a different keyspace than opencode tool names), `model` (source values like "sonnet" aren't in opencode's provider/model format).
- Orchestrator branch in `opencode.go` now emits `mode: primary`; other agents use the default (`mode: all`). This aligns with opencode's `default_agent: "orchestrator"` config.
- Wired the emitter into `OpenCodeAdapter.Install` — replaced the `StripFrontmatterAndInjectModel` transform (which emitted HTML comments) with a call to `BuildOpenCodeAgentFrontmatter`. Other adapters still use the shared stripper (unchanged).
- `TestOpenCodeAdapter_Install_FromFS` now asserts every installed agent parses as valid YAML frontmatter with `description` and `mode` keys populated.

## Verification

- `go test ./... -count=1` — PASS (6 new frontmatter unit tests + extended install assertion)
- `go vet ./...` — clean

## Scope

Write a dedicated opencode frontmatter builder used only by `OpenCodeAdapter.Install`. Replace the shared `StripFrontmatterAndInjectModel` call for opencode.

## Changes

- `internal/adapter/opencode_frontmatter.go` (new):
  - `OpenCodeAgentOpts` struct: `Description`, `Mode`, `Tools`, `Model`, `Permission`.
  - `BuildOpenCodeAgentFrontmatter(source []byte, opts OpenCodeAgentOpts) []byte`:
    - Strips existing frontmatter.
    - Inherits missing fields from source frontmatter where present (e.g., `description`).
    - Emits:
      ```yaml
      ---
      description: <string>
      mode: primary | subagent | all
      tools: { bash: true, read: true, write: true, ... }
      model: <provider/model>
      permission:
        edit: ask
        bash: ask
      ---
      <body>
      ```
    - `tools` shape: follow whatever opencode v1.4.9 accepts (confirm during implementation — map vs. comma string).
- `internal/adapter/opencode.go`:
  - Replace the `Transform: StripFrontmatterAndInjectModel` arg in the agents `CopyLibraryDirectory` call with a closure that calls `BuildOpenCodeAgentFrontmatter` per agent, pulling per-agent defaults from a small table.
  - Same treatment for the orchestrator branch.

## Tests

- `internal/adapter/opencode_frontmatter_test.go` (new):
  - Table test: various `OpenCodeAgentOpts` combinations → assert YAML output matches expected string.
  - YAML parse roundtrip: the output must parse back to the same struct.
  - Edge cases: source has extra frontmatter keys → dropped; source has matching `description` → inherited.
- `internal/adapter/opencode_test.go`:
  - After install, each agent file frontmatter parses without error and has the required keys.

## Definition of Done

- All installed opencode agents have schema-valid frontmatter.
- New emitter unit tests cover happy path + inheritance + edge cases.
- No regressions in other adapters (they still use `StripFrontmatterAndInjectModel`).
