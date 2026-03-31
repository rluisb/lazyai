<rule>
  <scope>auto</scope>
  <globs>docs/code-reviews/**</globs>
  <description>Code review workflow — reviewing a teammate's PR using Reviewer in External PR Mode</description>
</rule>

# Code Review Workflow Rules

> Use this flow when you are **reviewing a pull request you did not author.**  
> For reviewing your own work before opening a PR, use the Reviewer agent in standard mode.

## Directory Structure

```
docs/code-reviews/
└── NNN-pr-title-or-number.md    ← review output (one file per PR reviewed)
```

## Flow

### Step 1 — Context Assembly

Before reading any code, load:

1. **PR diff** — the full changeset (paste inline or reference file)
2. **PR description** — stated intent, scope, testing notes
3. **Linked ticket** — original requirement
   - If Jira MCP is available: `fetch ticket [TICKET-NNN]`
   - If not: ask the author for the ticket number before proceeding
4. **Relevant standards** — from `docs/standards/` for the languages/patterns changed
5. **Relevant ADRs** — from `docs/adrs/` for any architectural areas touched

> Context assembly is NOT optional. A review without ticket context misses intent. A review without standards misses conformance.

### Step 2 — Activate Reviewer (External PR Mode)

Invoke the Reviewer agent from `library/agents/reviewer.md`.

Tell Reviewer: *"This is an external PR review. No spec file exists. Infer intent from PR description and ticket."*

The Reviewer will:
1. State inferred intent in one sentence
2. Review diff for correctness, security, edge cases
3. Check test coverage
4. Check standards conformance
5. Issue verdict: APPROVE / REQUEST_CHANGES / COMMENT

### Step 3 — Fill Review Template

Use `docs/templates/code-review-template.md` to structure the output.

Output file: `docs/code-reviews/NNN-pr-title-or-number.md`

### Step 4 — Human Delivers Review

**AI never posts directly to GitHub or Jira.**

The human:
1. Reads the structured review output
2. Decides what to include/trim
3. Posts the review themselves

> The review artifact in `docs/code-reviews/` becomes a reference for future similar issues and standards extraction.

---

## Rules

- **No spec = infer from PR + ticket.** Never skip context assembly to save time.
- **Scope strictly.** Only review what's in the diff. Do not comment on unrelated pre-existing issues — if found, log separately as tech debt.
- **Evidence first.** Every finding references file + line. No vague feedback.
- **Professional tone.** Assume good intent. State findings, not judgments about the author.
- **HUMAN GATE before delivery.** Review output is a draft. Human edits before posting.
- **Missing PR description?** Flag it as a Minor finding, then proceed with best inference.
- **Scope expansion found?** Flag as Major + recommend separate ticket. Do not approve scope surprises.

## Verdict Reference

| Verdict | When to use |
|---|---|
| **APPROVE** | No Critical or Major findings. Minors documented but do not block merge. |
| **REQUEST_CHANGES** | One or more Critical or Major issues that must be addressed. |
| **COMMENT** | Questions or observations only. No changes required. |

## Post-Review (Knowledge Loop)

After each review:
- Did this PR reveal a missing standard? → Note in `docs/memory/` for promotion
- Did this PR reveal a recurring pattern? → Consider extracting to `docs/standards/`
- Did this PR reveal a missing rule? → Propose addition to `docs/rules/`
