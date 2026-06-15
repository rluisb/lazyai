---
name: research
description: Delegated to research-rpi — RPI-bounded research with trace protocol and structured output. Preserved here for backward compatibility with /research trigger.
argument-hint: "[topic-or-question]"
trigger: /research
phase: research
deprecated: true
alias_of: research-rpi
---

# Research Skill (Deprecated — Use research-rpi)

**This skill is a thin wrapper delegating to `research-rpi`.**

The canonical research skill for RPI workflows is **`research-rpi`** which provides:
- Bounded scope with feed-forward from spec.md
- Trace protocol (Thought/Action/Observation/Decision)
- Structured output schema
- Anti-slope guards

## Delegation

This skill exists for backward compatibility with the `/research` trigger.
All new work should use the `research-rpi` skill directly.

## Integration
- Agent: Scout
- Delegates to: `research-rpi` skill
- Preserves: `/research` trigger for backward compatibility
