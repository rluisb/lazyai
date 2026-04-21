# Task 003 — Embed new dirs + schema unit test

**Phase:** 1 (library wiring)
**Estimated LOC:** ~80

## Goal

Include the new `library/copilot/{agents,instructions}/` directories in the embedded library FS and assert their schema shape in a unit test.

## Files to touch

| File | Change |
|---|---|
| `internal/library/library.go` | Add `//go:embed` entries for `copilot/agents/*.agent.yaml` and `copilot/instructions/*.instructions.md` |
| `internal/library/library_test.go` | New test: `TestCopilotAgentsAndInstructionsEmbedded` — walks embed FS, asserts file count matches library dir |
| `internal/library/copilot_schema_test.go` (new) | Parses every `.agent.yaml` and `.instructions.md`, asserts required fields |

## Schema assertions

**`*.agent.yaml`:** `name` (non-empty, lowercase), `prompt` (non-empty), `description` (non-empty). `model` and `tools` optional.

**`*.instructions.md`:** `applyTo` frontmatter present + non-empty. Body non-empty.

## Acceptance criteria

- [ ] `internal/library` tests enumerate the new files from the embed FS
- [ ] Schema test fails loudly if a library file misses required keys
- [ ] `go test ./internal/library/... -count=1` green

## Test plan

- Add YAML parsing via `gopkg.in/yaml.v3` (already in go.mod if OpenCode frontmatter uses it; otherwise add).
- For instructions, reuse `internal/frontmatter.SplitYamlFrontmatter` + YAML unmarshal.

## Notes

- Keep embed pattern consistent with existing `library/agents/`, `library/skills/`, `library/opencode/` patterns.
- If `yaml.v3` isn't already a dep, use `sigs.k8s.io/yaml` or confirm presence before adding.
