# Task 003 — OpenCode-specific agent frontmatter emitter

**Phase:** 2
**Status:** pending
**Depends on:** 002

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
