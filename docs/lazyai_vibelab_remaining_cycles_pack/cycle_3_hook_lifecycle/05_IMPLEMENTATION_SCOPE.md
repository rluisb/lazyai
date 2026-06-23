# Implementation Scope — Cycle 3

## Priority A — Neutral hook lifecycle catalog

Define this lifecycle:

```text
before_agent
before_model
before_tool
after_tool
after_model
after_agent
on_error
on_compaction
on_handoff
on_human_gate
```

Add or update:

```text
packages/cli/library/hooks/catalog.md
docs/concepts/hook-lifecycle.md
```

If code support is appropriate:

```text
packages/cli/internal/hooks/lifecycle.go
packages/cli/internal/hooks/capabilities.go
```

Acceptance:

```text
- Lifecycle is documented.
- It does not imply LazyAI executes hooks itself.
- Host-tool execution boundary is explicit.
```

## Priority B — Hook capability matrix

Map support per adapter:

```text
supported
partial
instruction_only
unsupported
not_applicable
```

For each surface:

```text
opencode
claude
copilot
pi
omp
antigravity
kiro
```

Acceptance:

```text
- Weak surfaces use instruction-only fallbacks instead of fake hooks.
- Capability matrix matches adapter code.
```

## Priority C — Classify existing hooks

Classify:

```text
pre-commit
rpi-gate-check
caveman-memory-promotion
startup-self-heal
block-destructive-shell
objective-workflow-gate
```

For each hook record:

```yaml
name:
lifecycle:
purpose:
blocks_actions:
requires_human_approval:
captures_evidence:
surfaces:
```

Acceptance:

```text
- Every shipped hook has a lifecycle classification.
- Unsupported claims are not made.
```

## Priority D — Validation/tests

If validation exists for hook assets, add checks for:

```text
- unknown lifecycle
- unsupported adapter claim
- missing purpose
- missing safety behavior for blocking hooks
```

If validation does not exist yet, add docs and tests around capability mapping only.

Acceptance:

```text
- Tests pass.
- No runtime hook scheduler is added.
```
