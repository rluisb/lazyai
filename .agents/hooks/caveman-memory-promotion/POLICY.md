# Caveman Memory Promotion Policy

## Purpose

Keep the caveman-to-memory review workflow documented without claiming an emitted runtime hook where none is wired today.

## Events

- Generated adapters: none
- Manual workflow: after caveman summaries, diagnosis, triage, or handoff work
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

- Claude Code: no caveman-memory-promotion runtime hook is emitted.
- OpenCode: no caveman-memory-promotion plugin event is emitted.
- Copilot: no caveman-memory-promotion hook descriptor is emitted.
- Antigravity: no caveman-memory-promotion hook is emitted.
- OMP/Pi: no runtime hook is emitted.

## Safety Guardrails

- Must not write ai-memory automatically.
- Must not promote bare caveman bullets.
- Must include source, evidence, decision, scope, and expiry/removal condition when relevant.
- Must ignore secrets, credentials, and machine-local paths unless the user explicitly asks for private memory.
