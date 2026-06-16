---
name: fast-feedback
description: Use during implementation to run the smallest meaningful verification command after each focused change.
---

# Fast Feedback

## When to Use

Use this skill while editing code, docs generation, scripts, or skills where a quick local check can prove the last change behaved as intended.

## Rule

After each focused change, run the smallest command that can catch the likely mistake before continuing.

Prefer narrow checks over broad suites:

1. A test you added or changed.
2. A targeted typecheck, lint, or script check for the touched area.
3. The project-specific verification command from docs or package metadata.
4. The smallest generic command that exercises the changed behavior.

For vibe-lab skill work, the default loop is:

```bash
lazyai-cli compile
lazyai-cli doctor
```

## Choosing the Check

- Skill content or symlink changed: run `lazyai-cli compile`, then `lazyai-cli doctor`.
- Generated context changed: inspect the generated catalog entry for the changed skill.
- Script logic changed: run the script on the normal project path.
- Documentation-only change: run the generator or checker that consumes the doc, if one exists.

## Constraints

- Do not batch unrelated edits before the first check.
- Do not skip a failing check because a broader check might pass.
- Do not claim integration coverage from a syntax-only check.
- Keep the command headless and repeatable.
