---
name: git-hygiene
description: Use when cleaning up git worktrees and merged branches — after finishing feature work, before starting major new work, or when stale worktrees accumulate. Enforces dry-run-first previews and protected-branch guards.
---

# Git Hygiene

## When to Use

Use this skill when:
- Feature/bugfix work just merged and its worktree/branch should be cleaned up.
- Worktree count is growing and needs an audit.
- Starting major new work and the repo should be tidied first.
- A worktree lives at the wrong path (path mismatch) and needs relocating.

Do not use for creating worktrees or the day-to-day branch workflow — that is `worktree-workflow`.

## Rule

Preview before delete, always. Every prune/relocate runs `--dry-run` first, the candidate list is shown to the user, and only integrated + clean + unprotected items are removed. Deletion of anything dirty, unmerged, or protected halts and escalates.

## Hard Guards

1. **Never delete `main`, `master`, `develop`, `release/*`, `hotfix/*`.** Violation = halt and escalate.
2. **Never remove a worktree with uncommitted changes.** Check `git -C <path> status --porcelain` is empty; otherwise prompt to commit or stash.
3. **Verify integration before deletion** — the branch must be merged into the default branch (`wt step prune --dry-run` confirms; fallback `git branch --merged main`).
4. **List what will be deleted and get confirmation** before executing.
5. **Approval tiers** (per approved damage-control policy): `wt step prune --dry-run` / `wt step relocate --dry-run` / `wt list` are safe to run freely; non-dry-run `wt step prune`, `wt step relocate`, and plain `wt remove` require user confirmation; forced removal (`wt remove --force`/`-D`) discards work and always requires explicit confirmation.

## Workflow

### Quick mode — prune merged worktrees/branches

```bash
wt list --full                          # 1. inventory
wt step prune --dry-run --min-age=2d    # 2. preview candidates
# review the list with the user, then:
wt step prune --min-age=2d --foreground # 3. execute (confirmed)
```

### Full mode — quick + remote prune + path-mismatch audit

```bash
# quick mode, then:
git fetch --prune                       # drop stale remote-tracking refs
wt step relocate --dry-run              # preview path mismatches
wt step relocate                        # only with explicit approval
```

### Plain-git fallback (no `wt` available)

```bash
git worktree list
git -C <path> status --porcelain        # must be empty
git branch --merged main                # must contain the branch
git worktree remove <path> && git worktree prune
git branch -d <branch>                  # -d, never -D
git remote prune origin
```

## Decision Tree

- Worktree is the current one → skip.
- Dirty or branch unmerged → skip with warning.
- Clean and merged/branch-gone → remove candidate.
- Protected branch → skip, always.
- Unclear state → skip and report; never guess.

## Constraints

- Use `git branch -d` (safe delete), never `-D`, in cleanup flows.
- The `--min-age` guard (default 2d) prevents pruning just-merged work someone may still reference.
- Relocation of mismatched worktrees is never automatic — dry-run, report, and wait for approval.

## Verification Checklist

- [ ] Dry-run output reviewed before any destructive command.
- [ ] No protected branch in the deletion list.
- [ ] All removed worktrees were clean and integrated.
- [ ] `git fetch --prune` run in full mode.
- [ ] Skipped items reported with reasons.

## Related Skills

- `worktree-workflow` — creating and merging worktrees (the other half of the lifecycle).
- `git-operations` — commit hygiene and the PR/CI loop that precedes cleanup.
