<rule>
  <scope>auto</scope>
  <globs>docs/change-requests/**</globs>
  <description>Change request workflow — respond to PR review feedback</description>
</rule>

# Change Request Workflow Rules

## What This Flow Is For

A reviewer has left feedback on your PR. This workflow classifies each piece of feedback,
routes it to the right response path, and keeps a trace log.

---

## Step 1 — Triage: Classify Each Feedback Item

Before acting on anything, classify every feedback item into one of these types:

| Type | Description | Route |
|------|-------------|-------|
| **CI failure** | Pipeline is red, tests failing | Use `/iterate` skill |
| **Bug found** | Reviewer found a correctness issue | Mini-bugfix path ↓ |
| **Style / Trivial** | Naming, formatting, minor clarity | Fix direct |
| **Disagreement** | Valid alternative perspective | Discuss with human first |
| **Scope expansion** | Reviewer requesting new features | Reject — open new ticket |

⛔ **HUMAN GATE:** If any item is classified as **Scope Expansion**, surface it
to the human immediately. Do NOT implement scope expansions silently.

---

## Step 2 — Size Gate

After classifying, apply the size gate to determine planning depth:

| Size | Lines to Change | What to Do |
|------|----------------|------------|
| **Trivial** | < 5 lines | Fix direct. No planning artifact needed. |
| **Small** | 5–20 lines | Fix, then append CR log entry to progress.md |
| **Non-trivial** | > 20 lines | Mini-research note → written plan → fix → CR log |

For **Bug found** items, always use `docs/templates/bugfix-rca-template.md`
regardless of size, even if abbreviated.

---

## Step 3 — Implement

Follow the routed path from Step 1 and the depth from Step 2.

- Fix ONLY what was flagged. No drive-by improvements.
- For CI failures: run tests locally before pushing.
- For disagreements: write your reasoning clearly. The human decides.

---

## Step 4 — Append CR Log

After every round that required code changes, append to the `## Change Requests`
section of the parent feature's `progress.md`.

```
### CR Round [N] — [YYYY-MM-DD]
- **Reviewer:** [name]
- **Feedback type:** [CI failure | Bug found | Style | Disagreement | Scope expansion]
- **Items:** [N total — N critical, N major, N minor]
- **Action taken:** [fixed direct | fix + plan | rejected scope (ticket #NNN) | discussed]
- **Files changed:** [paths — or "N/A"]
- **Status:** ✅ Resolved | ⏳ In Progress | 🚫 Rejected
```

> **Preferred:** append to existing `docs/features/NNN-*/progress.md`.
> Only create a standalone `docs/change-requests/NNN-pr-name/` entry if there
> is no parent feature progress.md (e.g. hotfixes, dependency PRs).

---

## Directory Structure (when standalone needed)

```
docs/change-requests/NNN-pr-name/
└── progress.md          ← CR log only; no techspec unless non-trivial
```

---

## Principles

- **Fix the feedback, nothing else.** No drive-by refactors or improvements.
- **Never silently reject feedback.** Disagree → surface it. Don't just ignore it.
- **Scope expansions become new tickets.** Always. No exceptions.
- Trivial style-only rounds may be batched into one CR log entry.

---

## Self-Improvement — After Every Change Request Round

Before ending the session:

- Feedback revealed a missing rule? → Flag `docs/rules/` update
- CI failure revealed a missing test? → Flag `docs/standards/testing/` improvement
- Feedback pattern repeated across multiple PRs? → Write a memory note to `docs/memory/`
- Scope expansion rejected? → Confirm new ticket was created before closing
