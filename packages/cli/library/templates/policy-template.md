# Policy Template

Use for markdown-only policies and rules.

```markdown
# <Policy Name>

## Purpose

<What risk or drift this policy prevents.>

## Applies To

- <Artifact, phase, or event.>

## Rule

<One sentence with MUST / MUST NOT language.>

## Allowed

- <Allowed behavior.>

## Denied

- <Denied behavior.>

## Exception

<Who can approve, what must be recorded, and what removes the exception.>

## Verification

- [ ] <Check.>
```

Prefer one policy file over a script. Add runtime hooks only when the rule must block or warn automatically.
