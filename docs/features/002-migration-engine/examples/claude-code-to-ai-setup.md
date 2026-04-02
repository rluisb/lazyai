# Example Migration: Claude Code → ai-setup

This example shows how to bring a Claude Code setup under `ai-setup` management while keeping review and rollback easy.

## Starting point

Assume the source project includes:

```text
existing-project/
├── CLAUDE.md
└── .claude/
    ├── planner.md
    ├── commands/
    │   └── ship.md
    └── rules/
        └── review.md
```

## Step 1: Preview first

```bash
ai-setup migrate ./existing-project --preview
```

Why preview first:

- you confirm Claude Code was detected
- you see whether `CLAUDE.md` would be modified or backed up
- you learn whether unresolved conflicts will block execution

## Step 2: Pick the right strategy

Recommended decision guide:

- choose `smart` if you want the default review-oriented path
- choose `preserve` if your current `CLAUDE.md` is the source of truth
- choose `replace` if you want to reset to the `ai-setup` baseline quickly

Example:

```bash
ai-setup migrate ./existing-project --strategy smart --yes
```

## Step 3: Apply the migration

```bash
ai-setup migrate ./existing-project --yes
```

Expected summary fields:

- files created
- files modified
- files backed up
- files skipped
- unresolved conflicts, if any remain

If there are unresolved conflicts, use:

```bash
ai-setup migrate ./existing-project --interactive
```

## Step 4: Inspect the migrated tree

After a successful migration, look for:

```text
current-working-tree/
├── CLAUDE.md
├── .claude/
│   ├── planner.md
│   ├── commands/
│   └── rules/
├── .ai-setup.json
└── .ai-setup-backup/
```

Review focus:

1. Confirm `CLAUDE.md` still contains the project-specific context you care about.
2. Check migrated commands and rules for naming/placement.
3. Verify backups exist for overlapping files.
4. Check `.ai-setup.json` to confirm migration-managed files were recorded.

## Step 5: Run drift check

```bash
ai-setup doctor --migration-check --verbose
```

This is especially useful for Claude Code migrations because repositories often contain extra local conventions and untracked prompt files.

## Good follow-up actions

- remove obsolete duplicate files only after confirming the migrated equivalents are correct
- commit the migration in a dedicated changeset
- document any repo-specific follow-up conventions in your project docs
