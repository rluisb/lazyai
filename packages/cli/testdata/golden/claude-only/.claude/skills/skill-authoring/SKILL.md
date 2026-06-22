---
name: skill-authoring
description: Use when creating, renaming, or modifying LazyAI skill source files so the canonical source and generated tool-native outputs stay aligned.
---

# Skill Authoring

## When to Use

Use this skill when creating, renaming, or modifying a shipped skill source. In this repository the active embedded skill library is `packages/cli/library/skills/*.md`.

For agents, hooks, and workflow-catalog artifacts, use the corresponding authoring skill first.

## Rule

The current LazyAI embedded skill library is flat markdown source. Treat `packages/cli/library/skills/<name>.md` as the canonical file and treat generated tool-native skill surfaces as outputs.

## Verification Loop

After changing a skill source:

1. Refresh the affected generated surfaces.
2. Verify the changed guidance still reads correctly in the compiled target format.
3. Run `lazyai-cli compile` plus `lazyai-cli doctor` in a consuming project, or `go test ./packages/cli/...` plus `go build ./packages/cli/...` when changing LazyAI's embedded library or adapters.

## Constraints

- Do not duplicate the same guidance across multiple source files.
- Do not hand-edit generated tool-native skill files.
- Do not add helper scripts, workflow runtimes, or parallel source systems without explicit approval.
- Keep trigger text specific; broaden only with evidence.
