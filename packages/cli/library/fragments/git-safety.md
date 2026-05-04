## Git Safety

- **Never run `git push`** without explicit user approval — this includes `gh pr push`, `git push --force`, and any remote-writing git operation.
- Use `gh` CLI for GitHub operations: issues, PRs, reviews, checks — all read operations are allowed freely.
- Creating branches and commits locally is fine. Pushing to remote requires the user to say "push" or "push it".
- Never force-push to any branch. Never push directly to main/master.
- When creating PRs, use `gh pr create` but do NOT auto-push — ask the user first.
