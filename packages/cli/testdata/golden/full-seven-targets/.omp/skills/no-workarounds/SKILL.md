---
name: no-workarounds
description: Use during review or debugging to reject temporary patches and require root-cause fixes for workaround-shaped changes.
---

# No Workarounds

## When to Use

Use this skill before committing, during review, or whenever a fix starts to look like a patch over the symptom instead of the cause.

## Workaround Signals

Treat these as review triggers:

```text
TODO|FIXME|HACK|temporary|temp|quick fix|workaround|for now|bypass|skip validation|ignore error
```

Also inspect changes that:

- silence errors without explaining why the error is impossible or acceptable;
- add retries, sleeps, broad catches, or fallback paths not requested by the task;
- special-case one input without explaining the general invariant;
- loosen validation or types to make a check pass.

## Required Response

When a workaround signal appears:

1. Name the symptom being patched.
2. Identify the likely root cause or the missing evidence.
3. Replace the workaround with the smallest root-cause fix.
4. If the root cause is not yet known, stop and investigate before merging.

## Exception

A temporary marker is acceptable only when the user explicitly chooses a time-bounded mitigation and the code records:

- why the mitigation is safe;
- what removes it;
- how it is tracked outside the code.

## Constraints

- Do not add defensive infrastructure unless the task asks for it.
- Do not hide failures to make verification green.
- Do not leave workaround language in committed code without an explicit mitigation decision.
