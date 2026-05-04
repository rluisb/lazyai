# CLI Overview

`lazyai-cli` is a single binary with subcommands grouped by purpose:

- **Lifecycle**: `init`, `compile`, `update`, `doctor`, `status`
- **Scaffolding**: `create`, `add`, `import`, `migrate`, `eject`
- **Discovery**: `list`, `info`, `extensions`
- **Orchestration**: `orchestration list`, `orchestration create`, `orchestration status`
- **Utilities**: `completions`, `update-self`

All commands support `--verbose` / `-v`.

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
| `--verbose`, `-v` | Detailed output |
| `--dry-run` | Preview changes without writing |
| `--force` | Overwrite files and create backups |
| `--json` | Emit JSON instead of formatted output |
| `--no-interactive` | Disable prompts; all required flags must be provided |

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
