# Hook Template

Use for `.agents/hooks/<name>/POLICY.md` plus optional runtime scripts.

```markdown
# <Hook Name> Policy

## Purpose

<What the hook prevents or automates.>

## Events

- Claude Code: `<EventName>`
- OpenCode plugin: `<event.name>`
- OMP/Pi: <supported | unsupported | markdown-only>

## Decision

- Allow when: <condition>.
- Warn when: <condition>.
- Deny when: <condition>.

## Runtime

- Claude Code: generated command hook reads JSON from stdin and returns JSON or exit code.
- OpenCode: generated TypeScript plugin implements the mapped event.
- OMP/Pi: markdown-only or unsupported unless `bin/doctor` verifies a runtime adapter.

## Fail-Closed Semantics

- Safety hooks deny when they cannot parse input or load required runtime.
- Advisory hooks warn and continue when unavailable.
```

Claude Code hooks are lifecycle handlers configured in JSON. OpenCode hooks are TypeScript/JavaScript plugins. A hook is portable only when the policy names the event mapping and `bin/doctor` can validate both adapters.
