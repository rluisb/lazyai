# Clarification Levels

All clarification modes MUST preserve the four points: WHAT, HOW, DON'T WANT, VALIDATE.

## lightweight

Use when the user supplied at least three points and risk is low.

- Resolve missing facts from repo, docs, and tools first.
- Ask at most one focused question only when the missing point materially changes implementation.
- Output: the four points in 1-2 lines total, then proceed.

## grill-me

Use when requirements are vague, risky, or internally inconsistent.

- Ask targeted questions grouped by the four points only after repo, docs, and tools cannot answer.
- Stop when every material point has an explicit answer or documented assumption.
- Output: four-point summary plus the chosen implementation boundary.

## grill-me-with-docs

Use when local docs likely contain constraints or the change crosses subsystem boundaries.

- Read relevant docs and use available tools before asking.
- For each unresolved material point, cite the doc/tool-backed fact or gap.
- Ask only questions that repo, docs, and tools cannot answer and that materially change implementation.
- Output: four-point summary with cited constraints and validation path.

Never use clarification to expand scope or ask for information available from repo, docs, or tools. It exists to remove material ambiguity before work starts.
