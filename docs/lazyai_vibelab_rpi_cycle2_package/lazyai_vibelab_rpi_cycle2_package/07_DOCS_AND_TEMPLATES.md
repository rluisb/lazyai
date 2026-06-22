# RPI Cycle 2 — Docs and Templates

## Goal

Add or improve author guidance so future LazyAI/vibe-lab assets are easier to validate, maintain, and compile into host-tool surfaces.

## Suggested docs

```text
docs/concepts/skill-quality.md
docs/concepts/agent-contracts.md
```

If equivalent docs already exist, consolidate instead of duplicating.

## Suggested library templates

```text
packages/cli/library/templates/skill-quality.md
packages/cli/library/templates/agent-contract.md
```

If the project prefers another location, follow existing conventions.

## Skill quality doc should explain

```text
- what a skill is
- how skills differ from agents
- progressive disclosure
- trigger guidance
- non-trigger/misuse guidance
- required evidence
- output/done criteria
- human gates
- examples and anti-examples
- validation rule IDs
```

## Agent contract doc should explain

```text
- what an agent is
- how agents differ from skills
- role/purpose
- when to use
- when not to use
- expected workflow
- referenced skills/tools/fragments
- evidence requirements
- human gates
- output format
- handoff behavior
- safety boundaries
- validation rule IDs
```

## Template expectations

Templates should be concrete and ready to copy.

### Skill template sections

```markdown
# Skill: <name>

## Purpose

## When to use

## When not to use

## Required evidence

## Expected output

## Required tools or dependencies

## Human gates

## Procedure

## Examples

## Anti-examples
```

### Agent template sections

```markdown
# Agent: <name>

## Role

## When to use

## When not to use

## Workflow

## Evidence requirements

## Human gates

## Output format

## Handoff behavior

## Safety boundaries

## Referenced skills/tools/fragments
```

## Acceptance criteria

```text
- Docs are linked from harness principles if appropriate.
- Templates are included in library manifest/curation if required.
- Docs do not claim LazyAI executes agents.
- Docs preserve LazyAI’s boundary as a compiler/asset manager.
```
