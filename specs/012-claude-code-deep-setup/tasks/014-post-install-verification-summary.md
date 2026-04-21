# Task 014 — Post-install verification summary

**Phase:** 4 (verification)
**Estimated LOC:** ~50

## Goal

After a successful Claude Code install, when `claude` is on PATH, run `claude mcp list` and `claude agents --setting-sources user` (or `project` as appropriate), and render a compact summary block in the install output. Failure of this step is non-fatal — it's informational, not blocking.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode.go` | At the end of `Install()`, if CLI is present, call `mcp list` + `agents`; capture output; write to the install summary via whatever channel other adapters use (e.g. `ctx.Log` or the return value). |

## Output shape (target)

```
Claude Code install summary (scope: user)
  • 14 MCP servers registered (via CLI)
  • 7 agents available
  • 10 skills available
  • 3 commands available
  • 2 output styles available
```

## Acceptance criteria

- [ ] Summary prints after successful install when CLI is present
- [ ] Summary is skipped silently when CLI is absent (no error, no warning — the direct-write warning from Task 009 already covered the user)
- [ ] Counts are accurate (compare against filesystem walk)
- [ ] CLI failure during summary (e.g. MCP health-check timeout) does not fail the install

## Test plan

- Fake runner returns canned `mcp list` + `agents` output; assert summary content.
- Fake runner returns error for `mcp list`; assert install still succeeds, summary shows best-effort ("N/A" or similar).

## Dependencies

- Task 008 (probe + runner).
