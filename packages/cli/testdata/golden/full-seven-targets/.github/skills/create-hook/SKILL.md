---
name: create-hook
description: Use when asked to create or revise a LazyAI hook policy so runtime enforcement stays objective, generated, and tool-specific.
---

# Create Hook

## When to Use

Use this skill when the user asks to:
- block or warn on an observable action
- add a reusable hook policy to a managed source library
- map one objective policy across Claude Code, OpenCode, or Antigravity surfaces
- revise an existing generated runtime hook contract

Do not use when a normal markdown rule is enough and no runtime enforcement is needed.

## Rule

A shipped LazyAI hook starts with one policy source. In this repository that source is `packages/cli/library/hooks/<name>.md`. Generated shell scripts, JSON descriptors, and OpenCode plugins are outputs, not authoring locations.

## Workflow

1. Confirm whether the policy must block, warn, or only document behavior.
2. Choose a kebab-case hook name.
3. Write or update the policy source with purpose, trigger, decision rule, and fail-closed behavior.
4. Map the policy only to tool events the runtime can actually observe.
5. Keep runtime scripts non-interactive and deterministic.
6. Recompile affected tool surfaces and inspect the generated hook/plugin files.
7. Run `lazyai-cli compile` plus `lazyai-cli doctor` in a consuming project, or `go test ./packages/cli/...` plus `go build ./packages/cli/...` when changing LazyAI's embedded hook library or adapters.

## Event Mapping Guardrails

- Claude `PreToolUse` / `PostToolUse` map to direct hook events.
- OpenCode uses the generated plugin runtime under `.opencode/plugins/`.
- Antigravity is minimal `.gemini/settings.json` plus shell hooks only.
- Pi has no runtime hook surface.
- If no equivalent event exists, document the gap instead of faking parity.

## Constraints

- Hooks must never silently drop data.
- Safety hooks deny when required parsing/runtime fails.
- Do not add a workflow engine, queue, or hidden state machine.
- Generated runtime files are outputs; do not edit them directly.

## Verification Checklist

- [ ] Policy source is canonical and tool-agnostic.
- [ ] Supported tool mappings are explicit; unsupported ones are called out.
- [ ] Runtime scripts or plugins have clear inputs, outputs, and failure behavior.
- [ ] Recompiled generated hook assets match the intended policy.
