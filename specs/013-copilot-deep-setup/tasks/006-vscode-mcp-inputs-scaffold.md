# Task 006 — `.vscode/mcp.json` `inputs` scaffolding on env-placeholder detection

**Phase:** 2 (defect G5)
**Estimated LOC:** ~70

## Goal

When any MCP server's `env` block contains `${VAR}`-style placeholders, emit a matching top-level `inputs: [{type:"promptString",id:"VAR",password:true}]` array in `.vscode/mcp.json`. VS Code then prompts the user on first use and caches the value. When no placeholders exist, omit the `inputs` key entirely.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/mcp_compiler.go` | In `toCopilotMcp` (L339-364), scan each server's `env` values with a regex `\${([A-Z0-9_]+)}`. Deduplicate ids. If non-empty, add `inputs: [...]` to the output map. |
| `internal/adapter/mcp_compiler_test.go` | Add `TestToCopilotMcp_InputsScaffold` covering: (a) no placeholders → no `inputs` key; (b) single placeholder → single input; (c) duplicate placeholder across servers → single deduped input; (d) mixed literal + placeholder env values. |

## Input entry shape (VS Code docs)

```json
{
  "type": "promptString",
  "id": "<VAR>",
  "description": "${VAR}",
  "password": true
}
```

## Acceptance criteria

- [ ] Servers with placeholder env vars get matching input entries
- [ ] Servers with literal env values emit no `inputs` key
- [ ] Duplicate `${SAME_KEY}` across multiple servers = one input (dedupe by id)
- [ ] Existing tests for `toCopilotMcp` remain green

## Test plan

Unit tests inline in `mcp_compiler_test.go` covering the four cases above.

## Notes

- Keep the regex restrictive (`[A-Z0-9_]+`) so normal env values like `"$PATH:/extra"` don't accidentally match.
- `password: true` by default since most secrets are API keys; trade-off is that non-secret config values (e.g. a URL) also get hidden input — acceptable for the scaffold, user can edit.
