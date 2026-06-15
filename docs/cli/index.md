# CLI Overview

`lazyai-cli` is a single shipped binary with subcommands grouped by purpose. It is distinct from repository harness scripts such as `bin/doctor`, `bin/inject`, and `bin/startup-self-heal`, which are maintainer tools and not product commands.

- **Setup core**: `init`, `compile`, `update`, `doctor`, `status`, `setup`, `config`, `server`, `workspace`, `sidecar`, `validate`
- **Scaffolding and discovery**: `create`, `add`, `import`, `migrate`, `eject`, `list`, `info`, `build-plugin`
- **Runtime extras**: `session`, `ledger`, `memory`, `message`, `metrics`, `cost`, `backup`, `restore-runtime-db`, `secret`, `auth`, `notify`, `git`
- **Shell and binary lifecycle**: `completion`, `update-self`
- **Maintainer catalog tooling**: `models sync`

Removed command surfaces such as `task`, `workflow`, `orchestration`, `mcp-setup`, and obsolete `eval` are not active CLI commands. The hidden deprecated `completions` alias exists only for compatibility; use `completion`.

See [Product Boundaries](../concepts/product-boundaries.md) for the source-backed category inventory.

Global flags: `--verbose` / `-v`, `--log-level`, and `--log-format`. Command-specific flags such as `--dry-run`, `--force`, `--json`, and `--no-interactive` are documented per command.

## Common workflows

### First-time setup

```bash
lazyai-cli init
lazyai-cli status
lazyai-cli doctor
```

### Add a new tool to an existing setup

```bash
lazyai-cli add claude-code
lazyai-cli compile
```

### Refresh after upgrading the binary

```bash
lazyai-cli update-self
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor
```

### Check drift and health

```bash
lazyai-cli doctor
lazyai-cli doctor --skills-check
lazyai-cli doctor --migration-check
```

## Global flags

| Flag | Description |
|---|---|
| `--verbose`, `-v` | Enable verbose debug output |
| `--log-level` | Set log level (`debug`, `info`, `warn`, `error`) |
| `--log-format` | Set log format (`text`, `json`, `logfmt`) |

Flags such as `--dry-run`, `--force`, `--json`, and `--no-interactive` are command-specific, not global.

## TOML defaults

You can set CLI defaults in `.ai-setup.toml` (project) or `~/.config/lazyai/config.toml` (global):

```toml
default_scope = "project"
default_tools = ["opencode", "claude-code"]
install_mode = "symlink"

[wizard]
preset = "standard"
show_preview = true
```

Precedence: `CLI flags > project TOML > global TOML > built-in defaults`.
