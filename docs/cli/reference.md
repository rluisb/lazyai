# Command Reference

All command examples use the LazyAI CLI binary, `lazyai-cli`.

## `init`

Initialize a new managed AI setup.

```bash
lazyai-cli init [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--scope` | `project \| global \| workspace` | prompt | Setup scope |
| `--planning-repo` | path | — | Planning repo for workspace scope |
| `--repos` | comma-separated | — | Workspace repo references |
| `--tools` | comma-separated | prompt | `opencode,claude-code,copilot` |
| `--cli-tools` | comma-separated | — | Locally installed CLI tools |
| `--name` | string | dir-derived | Project or workspace name |
| `--force` | boolean | `false` | Overwrite managed files |
| `--no-interactive` | boolean | `false` | Disable prompts |
| `--migrate` | boolean | `false` | Detect and import existing setup |
| `--from` | path | current | Source path for migration |
| `--absorb` | boolean | prompt/`false` | Absorb detected config into `.ai/` |
| `--dry-run` | boolean | `false` | Preview without writing |
| `--preset` | `minimal \| standard \| full \| custom` | wizard | Feature preset |
| `--features` | comma-separated | — | Explicitly enable features |
| `--disable-features` | comma-separated | — | Disable features; `all` to start from nothing |
| `--branch-pattern` | string | `{type}/{ticket}-{description}` | Branch naming pattern |
| `--commit-pattern` | string | `{type}({scope}): {description}` | Commit message pattern |
| `--enable-servers` | comma-separated | — | Enable optional MCP servers |
| `--install-mode` | `copy \| symlink` | `copy` | How library files are installed |

**Examples**

```bash
lazyai-cli init
lazyai-cli init --scope project --tools opencode,claude-code --name my-repo --no-interactive
lazyai-cli init --scope workspace --planning-repo ./planning --repos ../api,../web --no-interactive
lazyai-cli init --migrate --from ../legacy-project
```

---

## `compile`

Recompile canonical content into tool-native directories.

```bash
lazyai-cli compile [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--scope` | `project \| global \| workspace` | manifest scope | Override scope |
| `--tools` | comma-separated | manifest tools | Compile only selected tools |
| `--force` | boolean | `false` | Overwrite existing files |
| `--dry-run` | boolean | `false` | Preview without writing |
| `--planning-repo` | path | manifest | Workspace planning repo |

**Examples**

```bash
lazyai-cli compile
lazyai-cli compile --tools opencode,claude-code
lazyai-cli compile --scope global
```

---

## `add`

Add another tool adapter to an existing setup.

```bash
lazyai-cli add <tool>
```

| Argument | Description |
|---|---|
| `tool` | `opencode`, `claude-code`, or `copilot` |

**Example**

```bash
lazyai-cli add copilot
```

---

## `update`

Refresh tracked files from the bundled library.

```bash
lazyai-cli update [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--force` | boolean | `false` | Overwrite with backup |
| `--check` | boolean | `false` | Preview which skills would be updated |

**Examples**

```bash
lazyai-cli update
lazyai-cli update --force
lazyai-cli update --check
```

---

## `doctor`

Verify setup health and detect drift.

```bash
lazyai-cli doctor [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--migration-check` | boolean | `false` | Compare to a clean LazyAI state |
| `--verbose` | boolean | `false` | Detailed output |
| `--json` | boolean | `false` | Emit JSON |
| `--skills-check` | boolean | `false` | Compare installed skills to library source |

**Examples**

```bash
lazyai-cli doctor
lazyai-cli doctor --verbose
lazyai-cli doctor --skills-check --json
```

---

## `status`

Print setup summary: scope, tools, features, git conventions, file health.

```bash
lazyai-cli status [--json]
```

---

## `create`

Scaffold a new artifact: agent, skill, command, prompt, template, workflow, domain, or mode.

```bash
lazyai-cli create <type> [name] [options]
```

| Flag | Default | Description |
|---|---|---|
| `--name` | prompt | Artifact name |
| `--description` | — | Artifact description |
| `--force` | `false` | Overwrite existing files |
| `--no-interactive` | `false` | Disable prompts |

**Subcommand flags**

- `create agent`: `--model`, `--mode`, `--tools`
- `create skill`: `--command`, `--steps`
- `create command`: `--arguments`, `--flags-description`
- `create prompt`: `--task-context`, `--output-format`
- `create template`: `--sections`, `--fields`
- `create workflow`: `--chain`, `--team`, `--steps`, `--step`
- `create domain`: no extra flags
- `create mode`: no extra flags

**Examples**

```bash
lazyai-cli create agent --name release-manager
lazyai-cli create skill deploy --command /deploy --steps "validate\nbuild\nship"
lazyai-cli create workflow release --chain feature --team review-team --no-interactive
```

---

## `import` / `migrate`

Detect and import an existing AI setup into LazyAI canonical format.

```bash
lazyai-cli import [path] [options]
lazyai-cli migrate [path] [options]
```

| Flag | Default | Description |
|---|---|---|
| `--preview`, `-p` | `false` | Show migration plan without applying |
| `--strategy`, `-s` | `smart` | `smart`, `preserve`, `replace`, `append` |
| `--verbose`, `-v` | `false` | Detailed output |
| `--interactive`, `-i` | `false` | Resolve conflicts interactively |
| `--skip-backup` | `false` | Skip backup creation |
| `--yes`, `-y` | `false` | Auto-confirm |
| `--no-canonical` | `false` | Use legacy output (migrate only) |

**Examples**

```bash
lazyai-cli import --preview
lazyai-cli import ../legacy-project --strategy preserve --yes
lazyai-cli migrate ../legacy-project --no-canonical
```

---

## `eject`

Stop managing the current setup. Removes `.ai-setup.json` while leaving generated files in place.

```bash
lazyai-cli eject
```

---

## `list`

List bundled library content.

```bash
lazyai-cli list [category] [--json] [--enabled]
```

**Categories:** `agents`, `skills`, `templates`, `rules`, `servers`/`mcp`, `tools`/`cli`, `workflows`, `chains`, `teams`, `domains`, `modes`, `orchestration`, `all`

**Examples**

```bash
lazyai-cli list agents
lazyai-cli list servers --enabled
lazyai-cli list orchestration --json
```

---

## `info`

Show detailed information about a library item.

```bash
lazyai-cli info <item> [--json]
```

**Examples**

```bash
lazyai-cli info builder
lazyai-cli info code-style --json
lazyai-cli info review-team
```

---

## `orchestration`

Orchestration-focused commands.

```bash
lazyai-cli orchestration list [kind] [--json]
lazyai-cli orchestration create <type> <name> [options]
lazyai-cli orchestration status [--json]
```

**Kinds:** `workflows`, `chains`, `teams`, `domains`, `modes`

**Examples**

```bash
lazyai-cli orchestration list workflows --json
lazyai-cli orchestration create domain payments --description "Payments domain" --no-interactive
lazyai-cli orchestration status
```

---

## `completions`

Print a shell completion script.

```bash
lazyai-cli completions [bash|zsh|fish]
```

**Example**

```bash
lazyai-cli completions bash
lazyai-cli completions zsh > ~/.config/fish/completions/lazyai-cli.fish
```

---

## `extensions` / `ext`

List discovered LazyAI extensions.

```bash
lazyai-cli extensions [--json]
```

---

## `update-self`

Download the latest `lazyai-cli` binary from GitHub Releases and replace the running binary.

```bash
lazyai-cli update-self [--check] [--dry-run] [--force]
```

| Flag | Description |
|---|---|
| `--check` | See if a newer release exists |
| `--dry-run` | Preview without applying |
| `--force` | Upgrade even if already on latest |

**Example**

```bash
lazyai-cli update-self --check
lazyai-cli update-self --dry-run
lazyai-cli update-self
```
