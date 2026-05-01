# Setup Engine Parity Rules

## Canonical source

Go is the source of truth for setup-engine and setup-wizard behavior. TypeScript must independently implement the same behavior. TS must not shell out to, import, or wrap the Go binary for setup parity.

## CLI action rules

The `setup` command supports exactly one primary action at a time:

- `--scan`
- `--list`
- `--dry-run`

Rules:

1. No primary action returns an error.
2. More than one primary action returns an error.
3. `--adopt` and `--import` require `--scan`.
4. `--tool`, `--all`, and `--global` are only valid with `--list` or `--dry-run`.
5. `--all` cannot be combined with `--tool`.
6. Unknown tool IDs return an error.
7. Selecting a tool for an unsupported filtered scope returns an error.

## Scope rules

Supported scopes are `global`, `project`, and `workspace`.

- Default `setup --list` includes all candidate roots for all supported scopes.
- `setup --list --global` filters to global roots and emits `scopeFilter: "global"`.
- `setup --dry-run` defaults to project scope.
- `setup --dry-run --global` uses global scope.
Project and workspace layouts are project-shaped and rooted at the selected target directory.

## Shared path rules

Shared desired paths are:

- `global-ai-setup`: `<HOME_DIR>/.ai-setup`
- `project-ai`: `<TARGET_DIR>/.ai`

For filtered outputs:

- global scope includes IDs prefixed by `global-`.
- project scope includes IDs prefixed by `project-`.

## Target detection rules

A target detection has status `detected` when at least one expected or optional path exists under the root. If `countRootOnly` is true, an existing root directory also counts as detected.

Otherwise, status is `missing`.

Detected resources receive a state based on registry comparison:

- no registry record: `adoptable`
- matching managed registry record: `managed`
- changed/missing/unexpected paths or MCP entries: `conflict`
- registry state `user-owned`: `user-owned`
- missing status: `missing`

## Dry-run action rules

Each dry-run target starts as:

- `existingStatus: "missing"`
- `action: "initialize"`

If a detection exists for the target/scope, dry-run copies `observedFiles`, `existingStatus`, and `existingState` from scan output.

If `existingStatus` is `detected`, dry-run action is `preserve-existing`.

## Reusable agent scan rules

Reusable agents are scanned from `<TARGET_DIR>/.ai/agents/<id>/`.

Agent rules:

- `<id>` must match `^[a-z][a-z0-9-]{0,63}$`.
- `AGENT.md` is required.
- `AGENT.md` must have valid YAML frontmatter and non-empty body.
- `title` comes from frontmatter `title`, then `name`, then first `# ` heading.
- `description` comes from frontmatter `description`, then first non-heading paragraph.
- `tools` accepts a string or a list of strings and is sorted.
- optional `mcp.json` must have exactly one top-level key: `mcpServers`.
- `mcpServers` must be an object and server names must be non-empty after trimming.
- agent `mcp.serverNames` are sorted.

Invalid agents are emitted with `status: "invalid"` and sorted reasons.

## MCP preset rules

Preset normalization:

- empty preset -> `recommended`
- unknown preset -> `recommended`

Preset selections:

- `minimal`: `filesystem`, `ripgrep`
- `recommended`: catalog servers where `enabled` is true
- `full`: all catalog server IDs

All preset outputs are sorted lexicographically by server ID.

## Adopt/import rules

Adopt:

- only `adoptable` resources are adopted.
- ineligible resources are skipped with reason `not-adoptable`.
- adopted resources are recorded as `managed`.

Import:

- importable states are `adoptable`, `managed`, and `user-owned`.
- ineligible resources are skipped with reason `not-importable`.
- imported files are copied under the ai-setup import root.
- existing destination files are backed up before replacement.

## Wizard parity rules

TS wizard behavior must match Go wizard state transitions and defaults, even if terminal rendering differs.

Phase 1 step order:

1. Scope
2. Tool targets
3. Skills
4. Agents
5. MCP preset
6. MCP servers
7. Project name
8. CLI tools
9. Project identity

Defaults:

- scope defaults to project.
- skills default to all known skills.
- agents default to all known agents.
- MCP preset defaults to recommended.
- project name is `global` for global scope.
- CLI tool defaults come from installed catalog tools.

## Drift policy

If Go and TS disagree, TS is wrong unless a spec explicitly changes Go first. Any intentional Go behavior change must update these contracts in the same change set.
