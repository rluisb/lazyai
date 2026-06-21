> **SUPERSEDED — Go/TS parity is moot; the TS CLI was removed. Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Research: Go/TS Setup Engine Parity

## Canonical decision

Go is the source of truth for setup-engine behavior and setup wizard logic. TypeScript must independently mirror Go behavior; TS delegation to the Go binary is explicitly out of scope.

If Go and TS disagree, Go wins and TS must be corrected.

## Current Go source of truth

Go currently owns the setup-engine surface:

- `setup --scan`
- `setup --list`
- `setup --dry-run`
- `setup --adopt`
- `setup --import`
- `setup --tool <name>`
- `setup --all`
- `setup --global`

Key Go files:

- `cmd/setup.go`
- `internal/setupscan/setupscan.go`
- `internal/setupscan/absorb.go`
- `internal/setupscan/agents.go`
- `internal/scaffold/mcp.go`
- `internal/adapter/scope.go`
- `internal/types/types.go`
- `tui/wizard/phase1.go`
- `tui/wizard/phase1_cli_mcp.go`

## Current TS gaps

- No TS `setup` command equivalent.
- TS does not support `pi` as a setup target.
- TS global target support is stale relative to Go.
- OpenCode config path expectations differ in TS areas.
- TS does not compile Codex MCP config to TOML.
- TS does not prepare orchestrator MCP the same way Go does.
- TS does not scan `.ai/agents/<id>/AGENT.md` reusable agents.
- TS lacks Go's setup scan/adopt/import inventory model.
- TS lacks Go's MCP preset semantics.

## Shared assets already available

- `library/mcp/catalog.json`
- `library/agents/`
- `library/skills/`
- `library/orchestration/`
- `specs/019-setup-engine-hardening/`

## Research conclusion

To keep Go and TS aligned while allowing both to be first-class implementations, shared contracts must be created first. TS parity work should then implement those contracts directly, with Go behavior used as the canonical reference.
