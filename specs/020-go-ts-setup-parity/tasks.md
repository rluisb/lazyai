> **SUPERSEDED — Go/TS parity is moot; the TS CLI was removed. Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Tasks: Go/TS Setup Engine Parity

## Phase 1 — Contracts

- [x] Add schemas for list, dry-run, and inventory outputs.
- [x] Add tool target registry from Go source of truth.
- [x] Add MCP preset and resource state data.
- [x] Add normalization and parity rules.

## Phase 2 — TS setup list/dry-run

- [x] Register TS `setup` command.
- [x] Implement `setup --list`.
- [x] Implement `setup --dry-run`.
- [x] Implement `--tool`, `--all`, and `--global` filtering to match Go.
- [x] Add TS conformance tests against contract expected output.

## Phase 3 — TS scan

- [x] Implement TS target scan inventory.
- [x] Add TS reusable agent scan for `.ai/agents/<id>/AGENT.md`.
- [x] Validate optional `mcp.json` with top-level `mcpServers`.
- [x] Add TS scan fixtures matching Go.

## Phase 4 — TS adopt/import

- [x] Implement shared registry semantics.
- [x] Implement adopt behavior for adoptable resources.
- [x] Implement import behavior with backups.
- [x] Add state transition tests.

## Phase 5 — MCP and orchestrator parity

- [x] Align OpenCode config paths with Go.
- [x] Add TS Codex MCP compilation.
- [x] Align Copilot global behavior.
- [x] Align Claude global/local secrets behavior.
- [x] Align orchestrator MCP local build/path/smoke behavior.

## Phase 6 — Wizard parity

- [x] Align TS wizard step order with Go.
- [x] Align defaults for tools, skills, agents, MCP preset, and servers.
- [x] Add wizard state parity tests.

## Dependency order

1. Contracts ✅
2. TS list/dry-run ✅
3. TS scan ✅
4. TS adopt/import ✅
5. TS MCP/orchestrator parity ✅
6. TS wizard parity ✅
7. Full validation — see Phase 5 verification below

## Validation

Run after every implementation phase:

- `go test ./... -v -count=1`
- `go vet ./...`
- `go build`
- `npm run lint`
- `npm test`
- `npm run typecheck`
- `npm run build`
- `npm --prefix orchestrator run build`
