# RPI Cycle 2 — Plan Phase

After research, produce a concise implementation plan before editing.

## Plan requirements

Include:

```text
1. Current skill/agent validation state
2. Proposed semantic validation checks
3. Which checks will be warnings vs errors
4. Files to change
5. Tests to add/update
6. Asset/doc updates needed
7. Compatibility risks
8. Validation commands to run
```

## Planning constraints

Keep the plan small enough to complete in this cycle.

Prefer:

```text
- validation infrastructure
- tests for representative good/bad cases
- docs/templates for authors
- minimal asset changes required to pass validation
```

Avoid:

```text
- huge rewrites of all skills
- strict failures for subjective quality
- public format changes unless already supported
- runtime behavior
- orchestration behavior
```

## Priority order

```text
A. Skill semantic validation
B. Agent role contract validation
C. Actionable validation output
D. Docs/templates for future authors
```

## Plan output template

```markdown
# RPI Cycle 2 Plan

## 1. Current state

## 2. Implementation scope

## 3. Warning rules

## 4. Error rules

## 5. Files to change

## 6. Tests to add/update

## 7. Docs/templates

## 8. Risks

## 9. Validation commands
```
