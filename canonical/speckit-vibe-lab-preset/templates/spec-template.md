# Spec: {{spec_name}}

## WHAT

Describe the problem or goal in one sentence.

## PURPOSE

State why this run exists and what decision or shipped outcome it serves.

## WHY

Explain the user value or constraint that makes this worth doing now.

## HOW

- Keep implementation details out until requirements are clear.
- Do not guess tech-stack/API/file-structure decisions unless the user supplied them.
- Mark any material unknown with `[NEEDS CLARIFICATION: specific question]`.

## BEHAVIOR SCENARIO

- Given <initial state>
- When <action>
- Then <observable outcome>

Use `n/a` only when the work is deliberately docs-only or read-only.

## DON'T WANT

- Premature implementation details or tech-stack/API/file-structure decisions before requirements are clear.
- Unsupported runtime claims (e.g., "this works with X" without verification).
- Scope creep: anything not in WHAT is out of scope.

## EXIT GATE

- [ ] Purpose is fulfilled or explicitly blocked.
- [ ] Validation evidence is observed.
- [ ] Out-of-scope discoveries are reported, not silently absorbed.

## VALIDATE

- [ ] The acceptance check is observable (test, command, or diff).
- [ ] Ambiguity is explicitly marked, not hidden.
- [ ] No claim is made about unverified integrations.
