<template>
  <name>Task — Individual Task File</name>
  <output>docs/features/NNN-feature-name/tasks/NNN-task-name.md</output>
  <input>tasks.md + techspec.md</input>
  <phase>Plan — Step 4 (breakdown)</phase>
</template>

# Task NNN: [Task Title]

**Feature:** NNN-feature-name
**Phase:** [N]
**User Story:** [US-N]
**Status:** TODO | IN_PROGRESS | DONE | BLOCKED
**Depends on:** [TNNN, TNNN — or "None"]

---

## Objective

<!-- One sentence. What this task accomplishes. Do one thing well. -->

[OBJECTIVE]

## Context

<!-- What the Builder needs to know. Reference files, not content.
     Keep under 10 lines. Builder reads the files directly. -->

- Read: [path/to/relevant/file]
- Read: [path/to/relevant/file]
- TechSpec section: [which section is relevant]

## Patterns to Follow

<!-- Files the Builder MUST read before writing. These are the style guide.
     Match their structure, naming, error handling, and conventions. -->

- Follow: [docs/standards/relevant-pattern.md]
- Mirror: [path/to/reference/file] — [what to mirror from it]

## Subtasks

<!-- Ordered steps. Builder checks these off one by one.
     Each step = one action. Keep it atomic. -->

- [ ] [step 1 — specific action]
- [ ] [step 2]
- [ ] [step 3]
- [ ] [step 4]
- [ ] Run: `[test command]`
- [ ] Verify: [what to check]

## Files to Touch

| File | Action |
|------|--------|
| [path] | Create | Modify | Delete |
| [path] | Create | Modify |

## Done When

<!-- Verifiable. Not "it works" — specific checks. -->

- [ ] [test passes: specific test name or command]
- [ ] [behavior: specific scenario verified]
- [ ] [no regressions: existing tests still pass]
- [ ] progress.md updated with this task's completion entry

## Notes

<!-- Optional. Gotchas, edge cases, or warnings. Delete if empty. -->

- [note]
