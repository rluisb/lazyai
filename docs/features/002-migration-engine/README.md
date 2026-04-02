# Migration Engine User Guide

The migration engine helps you move an existing AI-assistant setup into `ai-setup` with a preview-first workflow, merge strategies, backups, and drift checks.

If you already have files like `AGENTS.md`, `CLAUDE.md`, `.opencode/`, `.claude/`, `.pi/`, `.gemini/`, or GitHub Copilot instruction files, this is the feature that lets you adopt `ai-setup` without starting from scratch.

## What it does

The migration engine can:

- detect supported AI setup formats in an existing project
- build a migration plan before changing files
- preview the plan without writing anything
- back up overlapping files before applying changes
- stop and explain when manual conflict resolution is needed
- record migration-managed files in `.ai-setup.json`
- check post-migration drift with `ai-setup doctor --migration-check`

## Supported adapters

| Adapter | Typical markers |
|---|---|
| OpenCode | `AGENTS.md`, `.opencode/` |
| Claude Code | `CLAUDE.md`, `.claude/` |
| Pi | `CLAUDE.md`, `.pi/` |
| Gemini CLI | `GEMINI.md`, `.gemini/` |
| GitHub Copilot | `.github/copilot-instructions.md`, `.github/prompts/`, `.github/instructions/` |

## When to use which command

### `ai-setup import`

Use this as the primary migration command.

```bash
ai-setup import [path] [options]
```

Choose `import` when you want to:

- migrate the current repository into `ai-setup`
- migrate another repo into the current working tree
- preview and inspect a migration plan before applying it

### `ai-setup migrate`

`migrate` is an alias for `import`.

```bash
ai-setup migrate [path] [options]
```

Use it if `migrate` reads better in your workflow; behavior is the same.

### `ai-setup init --migrate`

Use this when you want to run a normal init flow, but first import an existing AI setup.

```bash
ai-setup init --migrate
ai-setup init --migrate --from ../existing-project
```

This flow:

1. detects the existing setup
2. tries to migrate it
3. continues into the regular init flow for any extra configuration

### `ai-setup doctor --migration-check`

Use this after migration to see how your working tree differs from a clean `ai-setup`-managed state.

```bash
ai-setup doctor --migration-check
ai-setup doctor --migration-check --verbose
```

## Quick start

### Preview before writing files

```bash
ai-setup import --preview
```

This is the safest starting point. It shows:

- which adapters were detected
- which files would be created, modified, or backed up
- whether unresolved conflicts would block execution

### Import the current repo

```bash
ai-setup import --yes
```

Use `--yes` in non-interactive workflows to skip the confirmation prompt.

### Import another repo

```bash
ai-setup import ../legacy-ai-config --preview
ai-setup import ../legacy-ai-config --strategy preserve --yes
```

### Check drift afterwards

```bash
ai-setup doctor --migration-check --verbose
```

## End-to-end workflow

Recommended sequence:

1. **Preview**
   ```bash
   ai-setup import --preview
   ```
2. **Pick a merge strategy** if the default is not what you want.
3. **Apply**
   ```bash
   ai-setup import --yes
   ```
4. **Review files and backups**
5. **Verify drift**
   ```bash
   ai-setup doctor --migration-check --verbose
   ```
6. **Commit once satisfied**

## Merge strategies

| Strategy | Best for | Behavior |
|---|---|---|
| `smart` | Careful migrations where you want conflict visibility | Builds a plan, attempts a 3-way merge, and blocks when manual review is still required |
| `preserve` | Protecting existing files | Keeps current content where files overlap and adds new `ai-setup` files around it |
| `replace` | Resetting to a cleaner baseline | Replaces overlapping files with `ai-setup`-managed versions and creates backups first |
| `append` | Combining content when supported | Appends or combines file content for parsers that support that mode |

### Strategy selection tips

- Start with **`smart`** if you want the safest default.
- Use **`preserve`** when your current instructions are highly customized.
- Use **`replace`** when the existing setup is outdated and you mainly want the new baseline.
- Use **`append`** only when you explicitly want combined content and are prepared to review the results.

## CLI reference

### `ai-setup import`

```text
ai-setup import [path] [options]

Options:
  -p, --preview              Preview changes without executing
  -s, --strategy <strategy>  Merge strategy: smart, preserve, replace, append
  -v, --verbose              Show detailed output
  -i, --interactive          Resolve merge conflicts interactively
      --skip-backup          Skip creating backup
  -y, --yes                  Auto-confirm without prompts
```

### `ai-setup migrate`

Same flags and behavior as `import`.

### `ai-setup init --migrate`

```text
ai-setup init --migrate [--from path]
```

### `ai-setup doctor --migration-check`

```text
ai-setup doctor --migration-check [--verbose]
```

## Reading the migration plan

The plan shows:

- **Source** and **Target** paths
- detected adapters
- counts for files to **create**, **modify**, **backup**, or **skip**
- unresolved conflicts that need attention
- a final status line showing whether the migration is ready to apply

Typical sections include:

- `Create` for new files that will be added
- `Modify` for files that will be updated
- `Backup` for files that will be copied into `.ai-setup-backup/`
- `Suggested next steps` when conflicts block execution

## Backups

By default, the engine creates backups before applying overlapping changes.

Backups are written under:

```text
.ai-setup-backup/migration-<timestamp>/
```

Use `--skip-backup` only if you are sure you do not need rollback material.

## Example migrations

Detailed walkthroughs:

- [OpenCode → ai-setup](./examples/opencode-to-ai-setup.md)
- [Claude Code → ai-setup](./examples/claude-code-to-ai-setup.md)

These examples show:

- a representative source layout
- the recommended commands to run
- what the plan/output should look like
- what to review after execution

## Troubleshooting

### “No supported AI setup detected”

The CLI could not find any recognized markers in the scanned path.

Try:

- running the command from the correct repository root
- passing an explicit path
  ```bash
  ai-setup import /path/to/project --preview
  ```
- checking whether the project has files such as `AGENTS.md`, `CLAUDE.md`, `.opencode/`, `.claude/`, `.pi/`, `.gemini/`, or `.github/copilot-instructions.md`

### “Unknown merge strategy”

Use one of:

- `smart`
- `preserve`
- `replace`
- `append`

Example:

```bash
ai-setup import --strategy preserve
```

### “Migration needs attention before it can continue”

The plan found unresolved conflicts.

Try one of these:

```bash
ai-setup import --interactive
ai-setup import --strategy preserve
ai-setup import --strategy replace
```

### Drift is reported after migration

That usually means one of these is true:

- files were edited after migration
- extra AI-tooling files exist outside the tracked manifest
- some expected files are missing

Use:

```bash
ai-setup doctor --migration-check --verbose
```

Then review the missing, extra, and modified file lists before deciding whether to update, restore, or keep local changes.

## Community extensions

You can create parsers for additional tools.

- Parser guide: [COMMUNITY-PARSERS.md](./COMMUNITY-PARSERS.md)
- Reference implementation: `src/migration/parsers/opencode-parser.ts`

## Related documents

- [PRD](./prd.md)
- [Tech spec](./techspec.md)
- [Implementation plan](./implementation-plan.md)
