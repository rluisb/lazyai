# Design: Go/TS Setup Engine Parity

## Principle

Go is canonical. TypeScript must independently implement the same setup engine behavior and pass the same contract fixtures. TS must not execute or delegate to the Go binary.

## Non-goals

- No TS delegation to Go.
- No runtime execution, ACP sessions, prompt execution, or daemon management.
- No new setup behavior that does not exist in Go.
- No TS-only shortcuts that change output shapes, states, paths, or defaults.

## Parity target

TS must mirror Go for:

- CLI flags and validation for `setup`.
- JSON output shapes for list, dry-run, scan, adopt, and import.
- Tool target IDs, display names, scopes, roots, and expected files.
- MCP preset behavior.
- Reusable agent directory scanning.
- Orchestrator MCP preparation semantics.
- Manifest/store ownership behavior.
- TUI/wizard step logic and defaults.

## Wizard parity

TS must mirror the Go wizard step order:

1. Scope
2. Tool targets
3. Skills
4. Agents
5. MCP preset
6. MCP servers
7. Project name
8. CLI tools
9. Project identity

Terminal rendering can differ, but step order, defaults, choices, labels, and resulting state must match.

## Contract strategy

Contracts live in `specs/contracts/setup-engine/` and define:

- Tool target registry.
- MCP presets.
- Resource states.
- JSON schemas for setup outputs.
- Normalization rules for paths, hashes, and timestamps.
- Parity rules for expected behavior.

## Implementation strategy

1. Contract data and schemas define language-agnostic truth.
2. Go remains the behavioral reference.
3. TS implements the same behavior directly.
4. Both languages run conformance tests against the same fixtures.

## Known mismatches to resolve

- Add TS `setup` command.
- Add TS `pi` target metadata.
- Align TS global scope support with Go.
- Align OpenCode config paths with Go.
- Add TS Codex MCP compilation.
- Align TS orchestrator MCP preparation with Go.
- Add TS reusable agent scan.
- Add TS scan/adopt/import state model.
- Add TS MCP preset parity.
