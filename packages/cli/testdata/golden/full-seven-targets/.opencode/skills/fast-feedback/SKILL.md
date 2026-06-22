---
name: fast-feedback
description: Use during implementation to run the smallest meaningful verification command after each focused change.
---

# Fast Feedback

## When to Use

Use this skill while editing code, docs generation, scripts, or library content where a quick local check can prove the last change behaved as intended.

## Rule

After each focused change, run the smallest command that can catch the likely mistake before continuing.

Prefer narrow checks over broad suites:

1. A test you added or changed.
2. A focused package test, typecheck, or lint step for the touched area.
3. The smallest compile/regeneration step that exercises the changed surface.
4. A broader package build/test only after the seam is stable.

## Choosing the Check

- Embedded library or adapter change: run the smallest relevant Go test first; fall back to `go test ./packages/cli/...` plus `go build ./packages/cli/...` when the change spans multiple setup surfaces.
- Consuming-project setup change: run `lazyai-cli compile`, then `lazyai-cli doctor` or a focused smoke check for the touched tool.
- Generated output change: inspect the generated file that should carry the new behavior.
- Documentation-only change: run the generator or checker that consumes the doc, if one exists.

## Constraints

- Do not batch unrelated edits before the first check.
- Do not skip a failing check because a broader check might pass.
- Do not claim integration coverage from a syntax-only check.
- Keep the command headless and repeatable.
