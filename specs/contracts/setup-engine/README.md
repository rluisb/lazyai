# Setup Engine Contracts

These contracts define language-agnostic setup-engine behavior for both Go and TypeScript.

## Canonical rule

Go is the source of truth. TypeScript must conform to Go behavior. If outputs differ, fix TS unless the Go source of truth is intentionally changed first.

## Delegation rule

TS must not delegate setup behavior to the Go binary. TS must implement the same contracts independently.

## Scope

Contracts cover:

- setup target registry
- MCP preset expansion
- resource states
- setup list output
- setup dry-run output
- setup scan inventory output
- normalization for path/hash/timestamp parity

## Files

- `data/tool-target-registry.json` — canonical tool IDs, names, scopes, roots, and expected files.
- `data/mcp-presets.json` — canonical MCP preset expansion and catalog-derived defaults.
- `data/resource-states.json` — canonical setup scan states and adopt/import state transitions.
- `schemas/setup-list-result.schema.json` — JSON schema for `setup --list`.
- `schemas/setup-dry-run-result.schema.json` — JSON schema for `setup --dry-run`.
- `schemas/setup-inventory.schema.json` — JSON schema for `setup --scan`, including adopt/import operation results.
- `conformance/normalization.md` — fixture normalization rules for paths, timestamps, hashes, ordering, and omitted fields.
- `conformance/parity-rules.md` — behavior rules TS must mirror from Go.

## Out of scope

- runtime execution
- ACP sessions
- daemon control
- prompt execution

`ai-setup` remains setup/bootstrap only.
