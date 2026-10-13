---
name: git-operations
description: Use for commit hygiene and the PR/CI loop — atomic conventional commits, remote alignment, PR creation with `gh`, and watching CI until green before merging.
---

# Git Operations

## When to Use

Use this skill when:
- Committing work (message format, granularity).
- Preparing and pushing a branch for review.
- Creating a PR and monitoring CI.
- CI fails after a push and needs a fix-and-rerun loop.

Worktree creation/lifecycle is `worktree-workflow`; cleanup is `git-hygiene`.

## Rule

Commit after every green state — one logical change per commit, in Conventional Commits format — and never consider a push done until CI is green.

## Atomic Conventional Commits

Format: `type(scope): imperative description`

| Type | Use | Type | Use |
|------|-----|------|-----|
| `feat` | new feature | `refactor` | restructure, no behavior change |
| `fix` | bug fix | `perf` | performance |
| `docs` | documentation | `build` / `ci` | build system / CI config |
| `test` | adding tests | `chore` | maintenance |
| `style` | formatting only | `revert` | revert a commit |

```bash
git commit -m "feat(auth): add JWT validation middleware"
git commit -m "fix(api): handle null response in user endpoint"
git commit -m "chore(deps): update tokio to 1.35"
```

Anti-patterns: `WIP`, `fix`, `updates`, or accumulating many unrelated changes into one commit.

## Remote Alignment

```bash
# before starting, periodically during long tasks, and before creating a PR:
git fetch origin
git rebase origin/main
# on conflict: resolve → git add . → git rebase --continue (or --abort to retry)
```

## PR & CI Loop

```bash
git push -u origin <branch>
gh pr create --fill              # --draft if not review-ready
gh pr checks --watch             # REQUIRED: watch until green

# if CI fails:
gh pr checks && gh run view <run-id>   # inspect
# fix locally, then:
git add -A && git commit -m "fix(ci): resolve linting errors"
git push && gh pr checks --watch

# after green:
gh pr merge --squash --delete-branch   # or `wt merge`
```

## Daily Reference

```bash
git status && git diff / git diff --staged / git diff main..HEAD
git log --oneline -20 / --graph / -p <file> / git blame <file>
git stash push -m "WIP: description" / git stash pop
```

## Safety Checklist (before push)

- [ ] Tests pass; format/lint clean.
- [ ] Commits atomic and conventional.
- [ ] Branch rebased on origin/main.
- [ ] Not pushing to a shared/protected branch directly.

## Constraints

- Never `git push --force`; `--force-with-lease` only when strictly necessary.
- Never merge with red or pending CI.
- Squash merge is the default merge strategy.
- All commands must stay non-interactive (`-m`/`-F` messages, no `rebase -i`) — see `shell-strategy`.

## Verification Checklist

- [ ] Every commit message is conventional and descriptive.
- [ ] PR created with an accurate title/body.
- [ ] `gh pr checks --watch` observed until green.
- [ ] Branch deleted (or worktree merged) after the PR merged.

## Related Skills

- `worktree-workflow` — where the branch lives and how it merges.
- `git-hygiene` — post-merge cleanup.
- `shell-strategy` — non-interactive git invocation patterns.
