---
name: issue-triage
description: Use when a bug report, error message, alert, or issue needs classification, deduplication, severity, ownership, refinement, and reusable triage learning before implementation.
---

# Issue Triage

## When to Use

Use this skill when:
- A new bug report or error arrives without clear categorization.
- The user asks whether an issue is known or where it belongs.
- An unlabeled backlog needs sorting.
- A crash report or alert needs classification.

Do not use when the issue is already labeled, assigned, deduplicated, and actionable. Do not use to fix the bug.

## Rule

Triage is classification, not solution. Answer: Is this real? Is this known? Who should fix it?

## Workflow

1. Read the issue and capture symptom, environment, version, frequency, and component.
2. Search for duplicates in the tracker and recent fixes when available.
3. Classify severity by user/data impact.
4. Apply labels/tags and ownership.
5. Request missing reproduction or environment details with a focused ask.
6. Document the triage decision and next action.
7. Capture reusable triage learning when classification revealed a recurring trap or rule.

## Triage Template

```markdown
Duplicate: <none | issue link>
Severity: <P0 | P1 | P2 | P3> because <impact>
Component: <area>
Owner: <team/person | unknown>
Missing Info: <none | requested detail>
Next Action: <diagnose | close | wait for info | route>
```

## Learning Capture

Use `canonical/learning-template.md` when triage reveals a reusable classification rule, ownership rule, severity trap, or routing pattern.

Raw triage learnings go to `specs/memory/sessions/learning-YYYY-MM-DD-<slug>.md`. Durable promotion requires `memory-promotion` approval.

## Constraints

- Do not debug during triage unless the fix is an obvious one-liner and user asked for it.
- Do not close without a reason.
- Do not leave issues ownerless; use a queue when ownership is unknown.
- Do not guess severity; justify it by impact.

## Verification Checklist

- [ ] Duplicate search performed or marked unavailable.
- [ ] Severity justified by impact.
- [ ] Component and owner recorded.
- [ ] Missing info requested if needed.
- [ ] Triage decision documented.
- [ ] Reusable learning captured or intentionally skipped.

## Related Skills

- `diagnose` — debug after triage.
- `task-to-issues` — create issues from unstructured text.
- `memory-promotion` — promote reusable triage knowledge.
