# Command Reference

## `init`

Initialize a new managed AI setup.

```bash
ai-setup init [options]
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
ai-setup init
ai-setup init --scope project --tools opencode,claude-code --name my-repo --no-interactive
ai-setup init --scope workspace --planning-repo ./planning --repos ../api,../web --no-interactive
ai-setup init --migrate --from ../legacy-project
```

---

## `compile`

Recompile canonical content into tool-native directories.

```bash
ai-setup compile [options]
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
ai-setup compile
ai-setup compile --tools opencode,claude-code
ai-setup compile --scope global
```

---

## `add`

Add another tool adapter to an existing setup.

```bash
ai-setup add <tool>
```

| Argument | Description |
|---|---|
| `tool` | `opencode`, `claude-code`, or `copilot` |

**Example**

```bash
ai-setup add copilot
```

---

## `update`

Refresh tracked files from the bundled library.

```bash
ai-setup update [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--force` | boolean | `false` | Overwrite with backup |
| `--check` | boolean | `false` | Preview which skills would be updated |

**Examples**

```bash
ai-setup update
ai-setup update --force
ai-setup update --check
```

---

## `doctor`

Verify setup health and detect drift.

```bash
ai-setup doctor [options]
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `--migration-check` | boolean | `false` | Compare to a clean ai-setup state |
| `--verbose` | boolean | `false` | Detailed output |
| `--json` | boolean | `false` | Emit JSON |
| `--skills-check` | boolean | `false` | Compare installed skills to library source |

**Examples**

```bash
ai-setup doctor
ai-setup doctor --verbose
ai-setup doctor --skills-check --json
```

---

## `status`

Print setup summary: scope, tools, features, git conventions, file health.

```bash
ai-setup status [--json]
```

---

## `create`

Scaffold a new artifact: agent, skill, command, prompt, template, workflow, domain, or mode.

```bash
ai-setup create <type> [name] [options]
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
ai-setup create agent --name release-manager
ai-setup create skill deploy --command /deploy --steps "validate\nbuild\nship"
ai-setup create workflow release --chain feature --team review-team --no-interactive
```

---

## `import` / `migrate`

Detect and import an existing AI setup into `ai-setup` canonical format.

```bash
ai-setup import [path] [options]
ai-setup migrate [path] [options]
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
ai-setup import --preview
ai-setup import ../legacy-project --strategy preserve --yes
ai-setup migrate ../legacy-project --no-canonical
```

---

## `eject`

Stop managing the current setup. Removes `.ai-setup.json` while leaving generated files in place.

```bash
ai-setup eject
```

---

## `list`

List bundled library content.

```bash
ai-setup list [category] [--json] [--enabled]
```

**Categories:** `agents`, `skills`, `templates`, `rules`, `servers`/`mcp`, `tools`/`cli`, `workflows`, `chains`, `teams`, `domains`, `modes`, `orchestration`, `all`

**Examples**

```bash
ai-setup list agents
ai-setup list servers --enabled
ai-setup list orchestration --json
```

---

## `info`

Show detailed information about a library item.

```bash
ai-setup info <item> [--json]
```

**Examples**

```bash
ai-setup info builder
ai-setup info code-style --json
ai-setup info review-team
```

---

## `orchestration`

Orchestration-focused commands.

```bash
ai-setup orchestration list [kind] [--json]
ai-setup orchestration create <type> <name> [options]
ai-setup orchestration status [--json]
```

**Kinds:** `workflows`, `chains`, `teams`, `domains`, `modes`

**Examples**

```bash
ai-setup orchestration list workflows --json
ai-setup orchestration create domain payments --description "Payments domain" --no-interactive
ai-setup orchestration status
```

---

## `completions`

Print a shell completion script.

```bash
ai-setup completions [bash|zsh|fish]
```

**Example**

```bash
ai-setup completions bash
ai-setup completions zsh > ~/.config/fish/completions/ai-setup.fish
```

---

## `extensions` / `ext`

List discovered ai-setup extensions.

```bash
ai-setup extensions [--json]
```

---

## `update-self`

Download the latest `ai-setup` binary from GitHub Releases and replace the running binary.

```bash
ai-setup update-self [--check] [--dry-run] [--force]
```

| Flag | Description |
|---|---|
| `--check` | See if a newer release exists |
| `--dry-run` | Preview without applying |
| `--force` | Upgrade even if already on latest |

**Example**

```bash
ai-setup update-self --check
ai-setup update-self --dry-run
ai-setup update-self
```
