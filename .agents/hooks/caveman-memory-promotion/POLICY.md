# Caveman Memory Promotion Policy

## Purpose

Detect when a caveman summary may contain reusable knowledge and route it to `memory-promotion` review without writing memory automatically.

## Events

- OpenCode plugin: `experimental.session.compacting`, `session.idle`
- Manual workflow: after caveman summaries, diagnosis, triage, or handoff work
- Claude Code/Copilot/Antigravity: policy retained; no runtime hook is emitted by LazyAI today
- OMP/Pi: markdown-only advisory unless runtime hook support is explicitly implemented later

## Decision

- Allow: normal compaction/session end when no promotion review is needed.
- Warn: a caveman summary contains reusable knowledge and should be reviewed through the `memory-promotion` skill.
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

## Runtime

- OpenCode: generated plugin prints a non-blocking advisory for caveman summaries that mention reusable decisions, traps, conventions, root causes, templates, patterns, rules, or lessons.
- Other generated adapters: no caveman-memory-promotion runtime hook is emitted today.

## Safety Guardrails

- Must not write ai-memory automatically.
- Must not promote bare caveman bullets.
- Must include source, evidence, decision, scope, and expiry/removal condition when relevant.
- Must ignore secrets, credentials, and machine-local paths unless the user explicitly asks for private memory.
