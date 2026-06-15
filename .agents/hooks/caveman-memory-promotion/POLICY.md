# Caveman Memory Promotion Policy

## Purpose

Detect when a caveman summary may contain reusable knowledge and route it to `memory-promotion` without writing memory automatically.

## Events

- Claude Code: `PreCompact`, `SessionEnd`
- OpenCode plugin: `experimental.session.compacting`, `session.idle`
- OMP/Pi: markdown-only advisory unless runtime hook support is verified

## Decision

- Allow: normal compaction/session end when no caveman summary is present.
- Warn: caveman summary contains decision, trap, convention, root cause, or reusable template.
- Deny: never. This is advisory only.

## Proposal Template

```markdown
Should I promote this to memory?
Target: <ai-memory entry | docs/path.md | canonical/path.md>
Classification: <rule | template | trap | pattern>
Source: <caveman source link/path>
Evidence: <observed fact>
Draft:
- <context-rich reusable fact>
```

## Safety Guardrails

- Must not write ai-memory automatically.
- Must not promote bare caveman bullets.
- Must include source, evidence, decision, scope, and expiry/removal condition when relevant.
- Must ignore secrets, credentials, and machine-local paths unless the user explicitly asks for private memory.
