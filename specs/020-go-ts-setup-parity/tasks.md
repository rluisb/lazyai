# Tasks: Go/TS Setup Engine Parity

## Phase 1 — Contracts

- [ ] Add schemas for list, dry-run, and inventory outputs.
- [ ] Add tool target registry from Go source of truth.
- [ ] Add MCP preset and resource state data.
- [ ] Add normalization and parity rules.

## Phase 2 — TS setup list/dry-run

- [ ] Register TS `setup` command.
- [ ] Implement `setup --list`.
- [ ] Implement `setup --dry-run`.
- [ ] Implement `--tool`, `--all`, and `--global` filtering to match Go.
- [ ] Add TS conformance tests against contract expected output.

## Phase 3 — TS scan

- [ ] Implement TS target scan inventory.
- [ ] Add TS reusable agent scan for `.ai/agents/<id>/AGENT.md`.
- [ ] Validate optional `mcp.json` with top-level `mcpServers`.
- [ ] Add TS scan fixtures matching Go.

## Phase 4 — TS adopt/import

- [ ] Implement shared registry semantics.
- [ ] Implement adopt behavior for adoptable resources.
- [ ] Implement import behavior with backups.
- [ ] Add state transition tests.

## Phase 5 — MCP and orchestrator parity

- [ ] Align OpenCode config paths with Go.
- [ ] Add TS Codex MCP compilation.
- [ ] Align Copilot global behavior.
- [ ] Align Claude global/local secrets behavior.
- [ ] Align orchestrator MCP local build/path/smoke behavior.

## Phase 6 — Wizard parity

- [ ] Align TS wizard step order with Go.
- [ ] Align defaults for tools, skills, agents, MCP preset, and servers.
- [ ] Add wizard state parity tests.

## Dependency order

1. Contracts
2. TS list/dry-run
3. TS scan
4. TS adopt/import
5. TS MCP/orchestrator parity
6. TS wizard parity
7. Full validation

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
