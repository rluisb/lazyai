---
name: caveman
description: Use when a specification, plan, or assistant message is too verbose and needs a compact working summary without losing links to source context or replacing durable ai-memory.
---

# Caveman

## When to Use

Use this skill when:
- A spec is over 200 lines and requirements are buried in prose.
- Stakeholders need quick agreement on what is being built.
- A feature request is vague and needs direct requirements.
- You need to compare two approaches without essays.
- The user says "keep it simple" or "just the facts".
- Assistant/system context is verbose and needs a working compression layer.

Do not use when the spec is already concise, when ambiguity is intentional, or when creating durable memory.

## Rule

Caveman compresses working context; it never replaces source docs or ai-memory. Always link to the full source.

## Format

```markdown
# Caveman: <Feature Name>

## Goal
<One sentence. What problem is solved?>

## Must
- <Hard requirement>

## Must Not
- <Safety boundary or rejected path>

## Can
- <Optional capability>

## Cannot
- <Known limitation or out-of-scope item>
```

Each section is five lines or fewer. Total output should be 25 lines or fewer unless the source has multiple independent features.

## Workflow

1. Read the source material and preserve its link/path.
2. Separate actual requirements from background and rationale.
3. Strip adjectives, adverbs, and speculative language.
4. Fill Goal, Must, Must Not, Can, Cannot.
5. Verify that Must items are non-negotiable and Can items are optional.
6. Link to the full source.
7. If the summary reveals reusable knowledge, use `memory-promotion`; do not store the caveman summary directly.

## ai-memory Boundary

Use `canonical/caveman-ai-memory.md` as the source of truth.

- Caveman is disposable working compression.
- ai-memory is durable shared memory.
- Memory entries need context: source, evidence, decision, scope, and removal/expiry condition when relevant.
- Never promote a bare caveman bullet. Promote the stable insight with enough context to reuse safely.
- A hook/plugin may propose promotion at `SessionEnd` or `PreCompact`, but it must not write memory automatically.

## When NOT to Caveman

- Capturing durable learning — use `memory-promotion`.
- Writing rules or templates — use canonical templates.
- TDD planning artifacts — use `tdd-planning`.
- Replacing a full spec — link the full spec instead.

## Verification Checklist

- [ ] Goal is exactly one sentence.
- [ ] Each section is five lines or fewer.
- [ ] No weasel words: should, might, consider.
- [ ] Must items are truly required.
- [ ] Must Not items are explicit safety boundaries.
- [ ] Full source is linked.
- [ ] Durable memory candidates are routed through `memory-promotion`.

## Related Skills

- `memory-promotion` — promote reusable insights after approval.
- `doc-backed-clarify` — resolve disagreements about Must vs Can.
- `tdd-planning` — plan test-first implementation work.
