---
name: issue-triage
description: Classify bug reports and error messages by severity, deduplication, ownership, and next action before implementation.
trigger: /issue-triage
phase: triage
techniques: [chain-of-thought]
output_schema:
  sections:
    - Duplicate Check
    - Severity Classification
    - Triage Decision
consumes:
  - bug report or error message
  - issue tracker access (when available)
produces_for:
  - diagnose (confirmed bugs)
---

# Issue Triage

## When to Use

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
