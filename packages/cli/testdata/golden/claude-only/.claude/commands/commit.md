---
description: Draft a Conventional Commits message for the staged diff
argument-hint: ""
allowed-tools: Bash Read
---

Draft a commit message for currently staged changes.

Context:
- Staged diff: `git diff --cached`
- Recent commit style: `git log --oneline -8`

Rules:
1. **Format**: Conventional Commits — `type(scope): subject`. Type ∈ {feat, fix, docs, refactor, test, chore, perf, build, ci}. Scope is the package or logical area.
2. **Subject**: imperative mood, lower-case, no trailing period, ≤ 72 chars.
3. **Body** (optional): wrap ~72 chars. Explain the *why*, not the *what*. Include only when subject is not self-evident.
4. **No co-author trailers** — the user adds those separately.

Output only the commit message (no explanation, no code fences). The user will paste it into `git commit -m`.
