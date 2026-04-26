# Setup Engine Conformance Normalization

These rules make Go and TypeScript setup-engine outputs comparable without hiding real behavioral drift.

## Canonical implementation

Go is canonical. Normalization is only for environment-dependent values such as absolute paths, timestamps, hashes, and platform path separators.

## Path tokens

Before comparing fixtures, replace absolute paths with stable tokens:

| Runtime value | Token |
| --- | --- |
| selected target/project/workspace directory | `<TARGET_DIR>` |
| selected home directory | `<HOME_DIR>` |
| repository library directory, when emitted | `<LIBRARY_DIR>` |

Rules:

1. Convert all path separators to `/` before token replacement.
2. Replace longer prefixes before shorter prefixes.
3. Do not normalize relative paths such as `settings.json`, `agents`, or `commands`.
4. Do not sort arrays unless a parity rule explicitly says the Go output is sorted.

## Time and hash tokens

Replace runtime-generated values only where the Go implementation emits them:

| Runtime value | Token |
| --- | --- |
| RFC3339 registry `updatedAt` values | `<TIMESTAMP>` |
| sha256 fingerprints in setup registry internals | `<HASH>` |
| backup path timestamp suffixes | `<TIMESTAMP>` within the filename |

The public `setup --scan`, `--list`, and `--dry-run` outputs should not contain fingerprints. If TS emits fingerprints in those outputs, that is drift.

## JSON formatting

Go writes setup JSON with two-space indentation through `encoding/json.Encoder`.
Conformance comparisons should parse JSON and compare semantic values after normalization, not raw whitespace.

## Ordering rules

The following ordering is part of the contract and must be preserved:

- Tool targets: lexicographic by tool ID.
- Target detections: lexicographic by `origin`, then `scope`, then `rootPath`.
- Desired `expectedFiles`: lexicographic.
- Observed files: lexicographic.
- Observed reusable agents: lexicographic by agent ID.
- Reusable agent tools and reasons: lexicographic.
- MCP preset server IDs: lexicographic by server ID.

## Omitted fields

Go uses `omitempty` on selected fields. TS must omit the same fields when empty instead of emitting empty arrays, empty strings, or nulls.

Key omitted fields:

- `scopeFilter` on unfiltered `setup --list`.
- `agents` when no reusable agents are detected.
- `observedFiles` and `existingState` in dry-run entries when absent.
- `reasons` and `mcpEntries` when absent.
- operation arrays such as `backups`, `adopted`, `imported`, and `skipped` when empty.

## Invalid output

Any of the following should fail parity:

- TS delegates to the Go binary.
- TS emits a tool, scope, state, action, or preset not in the contract.
- TS changes Go field names or casing.
- TS normalizes away real state differences such as `adoptable` vs `managed`.
- TS treats global Pi as supported.
