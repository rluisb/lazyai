# CLI Overview

`ai-setup` is a single binary with subcommands grouped by purpose:

- **Lifecycle**: `init`, `compile`, `update`, `doctor`, `status`
- **Scaffolding**: `create`, `add`, `import`, `migrate`, `eject`
- **Discovery**: `list`, `info`, `extensions`
- **Orchestration**: `orchestration list`, `orchestration create`, `orchestration status`
- **Utilities**: `completions`, `update-self`

All commands support `--verbose` / `-v`.

## Common workflows

### First-time setup

```bash
ai-setup init
ai-setup status
ai-setup doctor
```

### Add a new tool to an existing setup

```bash
ai-setup add claude-code
ai-setup compile
```

### Refresh after upgrading the binary

```bash
ai-setup update-self
ai-setup update --check
ai-setup update
ai-setup doctor
```

### Check drift and health

```bash
ai-setup doctor
ai-setup doctor --skills-check
ai-setup doctor --migration-check
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

You can set CLI defaults in `.ai-setup.toml` (project) or `~/.config/ai-setup/config.toml` (global):

```toml
default_scope = "project"
default_tools = ["opencode", "claude-code"]
install_mode = "symlink"

[wizard]
preset = "standard"
show_preview = true
```

Precedence: `CLI flags > project TOML > global TOML > built-in defaults`.
