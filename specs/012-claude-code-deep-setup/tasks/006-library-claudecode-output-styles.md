# Task 006 — Starter output-styles for Claude

**Phase:** 2 (starter library assets)
**Estimated LOC:** ~40 (mostly markdown content)

## Goal

Ship two starter output styles under `library/claudecode/output-styles/` with Claude-conformant frontmatter. Gives users a drop-in way to switch Claude's tone for different tasks.

## Files to create

| File | Content |
|---|---|
| `library/claudecode/output-styles/terse.md` | `name: Terse`, `description`, `keep-coding-instructions: true`, body: "Keep responses under 3 sentences unless the user asks for detail. No trailing summaries." |
| `library/claudecode/output-styles/explanatory.md` | `name: Explanatory`, `description`, `keep-coding-instructions: true`, body: "Before every non-trivial tool call, briefly state what you're about to do and why. Surface relevant trade-offs inline." |

## Frontmatter contract (Claude output styles)

```yaml
---
name: Human-readable name
description: What this style is for (shown in /config picker)
keep-coding-instructions: true   # keep default coding guidance
---
```

## Acceptance criteria

- [ ] Two markdown files exist with valid frontmatter
- [ ] Each file parses cleanly as YAML frontmatter + markdown body
- [ ] `name` fields are unique and human-readable
- [ ] Files are embedded in the library FS

## Test plan

- Parser test: load each file, verify frontmatter has `name` and `description`.
- Smoke: after install, `claude /config` picker shows both styles.

## Dependencies

- Library FS embed (same as Task 005).
