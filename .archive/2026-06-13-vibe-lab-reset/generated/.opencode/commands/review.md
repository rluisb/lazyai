---
description: Review the current branch's changes against main
---

Review the changes on the current branch relative to `main`.

Context to use:

- Diff: !`git diff main...HEAD`
- Commit log: !`git log --oneline main..HEAD`

Your review should cover:

1. **Correctness** — does the change do what its commit messages claim?
2. **Scope** — anything outside the stated scope of the change?
3. **Tests** — are the new/changed code paths covered? Are edge cases missing?
4. **Risks** — migrations, destructive operations, security or perf implications?
5. **Follow-ups** — anything worth a separate task, not a blocker here?

Return a short punch list (done / concerns / follow-ups), not a prose essay.
