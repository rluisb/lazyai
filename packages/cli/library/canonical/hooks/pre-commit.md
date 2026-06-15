# Pre-Commit Hook

Run before local commits.

## Responsibilities

- Fail when the canonical library exceeds the token-rent budget without a valid override.
- Reuse the same budget check command that CI runs.
- Print the exact override instruction so the exception path stays auditable.
