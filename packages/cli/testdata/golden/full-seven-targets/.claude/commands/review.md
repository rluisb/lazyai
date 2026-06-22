---
description: Review the current branch's changes against main
argument-hint: "[pr-number-or-path]"
allowed-tools: Bash Read
---

Review the current branch's changes or a named PR/path against main.

If $ARGUMENTS is provided:
- If it's a number: review that PR
- If it's a path: review changes to that file/directory

Context to gather:
1. Current branch diff: `git diff main...HEAD`
2. Commit messages: `git log --oneline main..HEAD`

Review dimensions:
1. **Correctness** — does the change do what its messages claim?
2. **Scope** — anything outside the stated scope?
3. **Tests** — are code paths covered? Edge cases missing?
4. **Risks** — migrations, destructive ops, security/perf implications?
5. **Follow-ups** — anything worth a separate task?

Report: short punch list (done / concerns / follow-ups), not prose.
