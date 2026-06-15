---
description: Draft a Conventional-Commits message for the staged diff
---

Draft a commit message for the currently staged changes.

Context:

- Staged diff: !`git diff --cached`
- Recent commit style (for tone): !`git log --oneline -8`

Rules:

1. **Format**: Conventional Commits — `type(scope): subject`. Type ∈
   { feat, fix, docs, refactor, test, chore, perf, build, ci }. Scope is
   the package or logical area being changed.
2. **Subject**: imperative mood, lower-case, no trailing period, ≤ 72 chars.
3. **Body (optional)**: wrap at ~72 chars. Explain the *why*, not the
   *what*. Only include a body when the subject is not self-evident.
4. **No co-author trailers.** The caller adds those.

Output only the commit message (no explanation, no code fences). The user
will paste it into `git commit -m`.
