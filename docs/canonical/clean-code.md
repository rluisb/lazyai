# Clean Code for Agents

> Small, explicit, and greppable. You are the primary reader.

## Functions & Files

- Small functions: 4–20 lines.
- Small files: <500 lines, ideally 200–300.
- One concept per file. If you need an `and` in the filename, split it.

## Naming

- Unique and greppable: `grep <name>` should return <5 hits.
- Avoid: `Manager`, `Service`, `handler`, `data`, `utils`.
- Prefer domain terms over generic ones.

## Comments

- Explain WHY and provenance (links, decisions, trade-offs).
- Keep your own comments on refactor — they are breadcrumbs.
- No "obvious" comments (`// increment i`).

## Types & Contracts

- Explicit types. No `any`.
- DRY — but prefer duplication over the wrong abstraction.
- Early returns. Max 2 levels of nesting.

## Errors

- Include the offending value + expected shape.
- Contextual, not generic.

## Tests

- Must run headless via a single command (see README).
- Behavior changes need a failing test first unless explicitly exempt.

## Structure & Dependencies

- Predictable structure: group by feature, not layer.
- DI / config isolation: no hardcoded env values in logic.
- Shallow call stacks. Prefer flat over deep.

## Formatting

- Use the project's formatter. No manual alignment wars.
