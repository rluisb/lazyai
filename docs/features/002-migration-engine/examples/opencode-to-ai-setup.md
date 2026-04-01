# Example Migration: OpenCode → ai-setup

This example shows a conservative migration from an existing OpenCode setup into `ai-setup`.

## Starting point

Assume the source project looks like this:

```text
legacy-project/
├── AGENTS.md
└── .opencode/
    ├── agents/
    │   └── builder.md
    ├── commands/
    │   └── research.md
    └── templates/
        └── task.md
```

## Step 1: Preview the plan

```bash
ai-setup import ./legacy-project --preview
```

Typical outcome:

- OpenCode is detected
- the CLI prints a migration plan
- you see counts for files to create, modify, and back up
- the CLI tells you whether unresolved conflicts would block execution

Example plan shape:

```text
Migration plan
==============

Source: /path/to/legacy-project
Target: /path/to/current-working-tree
Detected adapters: OpenCode

Summary:
  Create: 2
  Modify: 1
  Backup: 1
  Skip: 0
  Unresolved conflicts: 0

Status: ready to apply ✅
```

## Step 2: Choose a strategy

Good defaults for OpenCode:

- `smart` if you want conflict detection and review
- `preserve` if the existing `AGENTS.md` is heavily customized

Example:

```bash
ai-setup import ./legacy-project --strategy preserve --yes
```

## Step 3: Apply the migration

```bash
ai-setup import ./legacy-project --yes
```

Expected UX:

- the CLI confirms OpenCode detection
- a backup directory is created unless `--skip-backup` is used
- a summary lists files created, modified, backed up, and skipped
- the CLI recommends running `ai-setup doctor --migration-check`

## Step 4: Review the result

Typical review checklist:

1. Open the migrated root instruction file.
2. Check `.opencode/agents/`, `.opencode/commands/`, and `.opencode/templates/` for expected content.
3. Inspect `.ai-setup-backup/` if any overlapping files were backed up.
4. Confirm `.ai-setup.json` was written or updated.

## Step 5: Verify drift

```bash
ai-setup doctor --migration-check --verbose
```

This helps you answer:

- which files are now managed
- whether extra files still exist outside the manifest
- whether anything looks unexpectedly modified

## When to use `--interactive`

If the preview reports unresolved conflicts, rerun with:

```bash
ai-setup import ./legacy-project --interactive
```

This is a good fit when your existing `AGENTS.md` includes custom sections you want to resolve manually rather than preserve or replace wholesale.
