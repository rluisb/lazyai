---
name: worktree-workflow
description: Use when starting, working in, merging, or coordinating git worktrees with the `wt` (Worktrunk) CLI — including branch naming, lifecycle, parallel-agent isolation, and which wt commands are safe vs. approval-gated.
---

# Worktree Workflow (`wt`)

## When to Use

Use this skill when:
- Starting a feature, bugfix, refactor, or spike that will touch code.
- Deciding whether work needs a worktree at all.
- Merging finished worktree work back to the default branch.
- Running multiple agents/tasks in parallel on one repo.

Cleanup/pruning of stale worktrees is `git-hygiene`; this skill covers creation through merge.

## Rule

Never implement or commit directly on `main`. Every non-trivial task starts with `wt switch -c <branch>` and returns to main only via `wt merge` (or a reviewed PR merge). One worktree branch per task.

## When a Worktree Is Required

Required for: features, multi-file refactors, bug fixes with review/repair loops, anything beyond a quick one-file edit.

Direct branchless work is acceptable only for: read-only investigation, local documentation notes, tiny non-risky tweaks with explicit operator intent.

## Branch Naming

`{type}/{descriptive-name}`, ticket-linked where possible:

```bash
wt switch -c feat/user-auth
wt switch -c fix/PROJ-123-login-bug
wt switch -c chore/update-deps
wt switch -c spike/streaming-poc   # exploration, may be abandoned
```

## Lifecycle

```bash
# 1. CREATE
wt switch -c feat/PROJ-123

# 2. SYNC before starting (and periodically during long tasks)
git fetch origin && git rebase origin/main

# 3. WORK — atomic conventional commits (see git-operations)

# 4. FINISH — either:
wt merge                     # squash + rebase + fast-forward + remove worktree
# or the PR route:
git push -u origin <branch> && gh pr create --fill && gh pr checks --watch
gh pr merge --squash --delete-branch
```

## No `wt` Available?

Fall back to plain git — the policy (isolated worktree per task, never work on main) still applies:

```bash
git worktree add ../<repo>-<branch> -b <type>/<name>   # create + branch
# work there; merge via the normal PR flow or `git merge --squash`
git worktree remove ../<repo>-<branch>                  # confirm-first cleanup
```

This is the secondary path; prefer the `wt` flow whenever the CLI is installed.

## Command Safety Tiers

Per the approved damage-control policy:

| Tier | Commands | Behavior |
|------|----------|----------|
| Safe — run freely | `wt list`, `wt status`, `wt config`, `wt switch`/`wt switch -c`, `wt step diff`, any `--dry-run` step | Read-only or safe creation |
| Allowed workflow | `wt merge`, `wt step commit`, `wt step push` | Normal approved flow; proceed |
| Ask first | plain `wt remove`, `wt step promote`, `wt step prune`/`relocate` (non-dry-run), `wt switch --clobber` | Confirm with the user |
| Ask + explicit warning | `wt remove --force` / `-D` | Discards uncommitted work — never without explicit confirmation |

## Parallel-Agent Patterns

- Each concurrent task/agent gets its own worktree — isolation prevents cross-task conflicts and accidental commits to main.
- Dispatch pattern: `wt switch -c feat/<name> -x "<command>"` creates the worktree and launches work inside it in one step.
- Review pattern: `wt switch pr:<number>` checks out a GitHub PR into an isolated worktree.
- `wt switch ^` resolves the default-branch worktree.
- Audit health across all parallel work with `wt list --full` plus dry-run prune/relocate (see `git-hygiene`).

## Constraints

- No `git push --force`; use `--force-with-lease` only if truly necessary.
- Keep the worktree rebased on `origin/main` before merging.
- Abandoning a worktree (forced remove) is an operator decision, never an agent default — prefer committing WIP or stashing.
- One task = one worktree = one branch; don't reuse a worktree for unrelated work.

## Verification Checklist

- [ ] Work is happening in a worktree, not on main.
- [ ] Branch name is typed and descriptive/ticket-linked.
- [ ] Branch synced with origin/main before merge.
- [ ] Merge path was `wt merge` or a green-CI PR merge.
- [ ] No forced/destructive wt command ran without explicit confirmation.

## Related Skills

- `git-operations` — commit hygiene and PR/CI workflow inside the worktree.
- `git-hygiene` — pruning merged worktrees and stale branches afterwards.
- `shell-strategy` — all wt/git invocations must stay non-interactive.
