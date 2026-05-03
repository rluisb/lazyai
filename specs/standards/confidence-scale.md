# Standard: Reversa Confidence Scale

**Category:** Process
**Date:** 2026-05-03
**Owner:** ai-setup
**Status:** Active

---

## Rule

Every value filled by `/populate` or any AI-inferred content in `AGENTS.md` MUST carry a 🟢🟡🔴 confidence tag, and the tag dictates whether the value may be written, reviewed, or left as a marker.

## Confidence Levels

| Level | Meaning | Required Evidence | Write Behavior |
|---|---|---|---|
| 🟢 **CONFIRMED** | Direct, verifiable evidence exists. | Config file, linter rule, `package.json` entry, or code at a specific `file:line`. | **Write immediately.** Include evidence in parentheses. |
| 🟡 **INFERRED** | Observed pattern or framework convention, but not enforced. | Consistent across multiple files, directory structure, or git log patterns. | **Write after dry-run approval.** Explain the observed pattern. |
| 🔴 **GAP** | No evidence, ambiguous evidence, or contradictory evidence. | None. | **Leave untouched.** Keep the `<!-- fill-in: hint -->` marker. |

## Write Protocol

1. **Dry-run first.** Present all 🟢 and 🟡 findings with evidence before writing anything.
2. **Never write 🔴 gaps.** An honest 🔴 is better than a convincing-sounding 🟡 guess.
3. **Never overwrite human-authored values.** Only replace `<!-- fill-in: ... -->` markers.
4. **When evidence is ambiguous**, present both interpretations with 🟡 and note the ambiguity.

## When to Ask the Human

- **Before writing** any 🟡 values that affect core conventions (e.g., error handling, architecture, naming).
- **When 🔴 gaps exceed 20%** of all placeholders — shallow analysis may indicate missing tooling or an unusual project structure.
- **When 🟢 and 🟡 findings contradict each other** — signal data-quality issues rather than picking a winner.
- **When all three indexes (codegraph, qmd, graphify) are missing** — fall back to mechanical placeholders only and warn the user.

## Examples

**🟢 CONFIRMED (good):**
```markdown
Language: TypeScript 🟢 (tsconfig.json + 142 .ts files)
Test framework: Jest 🟢 (package.json devDependencies + jest.config.ts)
```

**🟡 INFERRED (acceptable with explanation):**
```markdown
Architecture: Layered with services/ and controllers/ 🟡
(observed: src/services/, src/controllers/, src/middleware/)
```

**🔴 GAP (must leave untouched):**
```markdown
<!-- fill-in: Git workflow branch naming convention -->
```

## Golden Rule

> When in doubt, use the **lower** level. Never invent evidence.

## Enforcement

- **Skill-level:** `/populate` hard rules (see `packages/ai-setup-go/library/skills/populate/SKILL.md`).
- **Review-level:** Any PR that introduces untagged inferred values into `AGENTS.md` should be flagged during code review.
- **CI-level:** Not automated; relies on agent self-reporting in `.ai/populate-report.md`.
