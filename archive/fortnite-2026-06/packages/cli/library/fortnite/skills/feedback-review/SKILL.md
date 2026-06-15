---
name: feedback-review
description: "Process feedback on our own work-in-progress. Classifies PR comments as questions or change requests, evaluates against specs/docs/codebase, and produces an action plan. Use when reviewing PR comments, code review feedback, or deciding what to fix."
trigger: /feedback-review
---

# /feedback-review

Process feedback on our own work-in-progress. Classifies comments, evaluates against specs, and produces a fix plan.

## When to Use

- Someone left comments on your PR and you need to decide what to fix
- Code review feedback is ambiguous (question vs change request)
- You need to understand if a comment is valid based on specs and requirements
- You need a structured action plan for addressing feedback

## Workflow

### Phase 1: Gather Context

```
1. Identify the PR (current branch or specified)
2. Fetch PR comments and review comments
3. Extract Jira ticket reference from PR body, branch name, or labels
4. Fetch Jira ticket details (description, acceptance criteria, comments)
5. Find related Confluence pages (linked from Jira ticket or PR body)
6. Find related specs (bee-gone/specs/ if applicable)
```

**Tools:**
- `gh pr view --comments --json ...` — Get PR details and comments
- `gh pr review --list` or `gh api` — Get review comments with file/line context
- `atlassian_getJiraIssue` — Fetch Jira ticket
- `atlassian_getConfluencePage` — Fetch Confluence pages
- `atlassian_search` — Search for related specs/docs

### Phase 2: Classify Comments

For each comment, classify into one of these categories:

| Category | Description | Action |
|----------|-------------|--------|
| **Question** | Asks for clarification, understanding, or rationale | Answer based on specs/docs/codebase |
| **Change Request** | Asks for a specific code change, fix, or improvement | Evaluate against specs → decide if/how to fix |
| **Nit/Suggestion** | Style preference, optional improvement | Note but don't block |
| **Already Fixed** | Issue was already addressed | Point to the fix |
| **Out of Scope** | Request is outside the ticket/spec scope | Document and defer |

**Classification Rules:**
- If the comment asks "why" or "how" → **Question**
- If the comment says "should", "needs to", "please change" → **Change Request**
- If the comment is about formatting, naming, or style → **Nit/Suggestion**
- If the code already does what's requested → **Already Fixed**
- If the request is not in the acceptance criteria → **Out of Scope**

### Phase 3: Evaluate Change Requests

For each **Change Request**, evaluate against the spec:

```
1. Does the spec/acceptance criteria require this behavior?
   - YES → Must fix (spec-aligned)
   - NO → Is it a valid improvement?
     - YES → Should fix (quality improvement)
     - NO → Is it a valid concern?
       - YES → Discuss with reviewer (trade-off)
       - NO → Decline with justification
```

**Evidence Sources (in priority order):**
1. **Spec/acceptance criteria** — definitive source of truth
2. **Jira ticket description** — requirements and context
3. **Confluence docs** — architecture decisions, patterns
4. **Codebase patterns** — existing conventions
5. **Reviewer expertise** — trust domain knowledge

### Phase 4: Produce Action Plan

Output a structured action plan:

```markdown
## Feedback Review — PR #NNN

### Summary
- Total comments: N
- Questions: N (answered)
- Change requests: N (N must fix, N should fix, N discuss, N decline)
- Nits: N
- Already fixed: N
- Out of scope: N

### Questions (Answered)
| # | Comment | Answer | Evidence |
|---|---------|--------|----------|
| 1 | "Why use X instead of Y?" | "Because spec requires X per AC-3" | SPEC.md §3.2 |

### Must Fix (Spec-Aligned)
| # | Comment | File:Line | Fix Plan | Priority |
|---|---------|-----------|----------|----------|
| 1 | "Handle null case" | auth.ts:42 | Add null check before access | P0 |

### Should Fix (Quality)
| # | Comment | File:Line | Fix Plan | Priority |
|---|---------|-----------|----------|----------|
| 1 | "Add error message" | api.ts:15 | Return 400 with message | P1 |

### Discuss (Trade-offs)
| # | Comment | Trade-off | Recommendation |
|---|---------|-----------|----------------|
| 1 | "Use cache here" | Adds complexity vs perf gain | Defer to next iteration |

### Decline (With Justification)
| # | Comment | Reason | Evidence |
|---|---------|--------|----------|
| 1 | "Refactor entire module" | Out of scope for this ticket | JIRA-123 AC doesn't include |

### Next Steps
1. [ ] Fix #1: auth.ts:42 — add null check
2. [ ] Fix #2: api.ts:15 — add error message
3. [ ] Reply to reviewer on #3 (trade-off discussion)
```

## Scripts

### `scripts/collect-feedback.sh`

Collects all PR comments and review comments into a structured JSON file.

```bash
./skills/feedback-review/scripts/collect-feedback.sh [PR_NUMBER]
```

Output: `feedback-<PR_NUMBER>.json` with classified comments.

## Agent Assignment

- **Primary**: shield-audit (MODE=review, FOCUS=spec)
- **Secondary**: wall-builder (for implementing fixes after review)
- **Support**: loot-hawk (for codebase research on ambiguous comments)

## Integration with Other Skills

| Skill | When |
|-------|------|
| **build-mode** | After feedback review, implement the fixes |
| **zero-point** | Verify fixes address the feedback |
| **storm-scout** | If feedback reveals spec ambiguity, clarify requirements |
| **pr-review** | Related — but pr-review is for reviewing OTHERS' PRs |

## Rules

1. **Never dismiss a comment without evidence** — always cite spec, doc, or code
2. **Questions get answers, not deflections** — if you don't know, research it
3. **Change requests get evaluated, not auto-accepted** — check against spec first
4. **Out of scope gets documented, not ignored** — create a follow-up ticket
5. **Nits get acknowledged, not debated** — "Good catch, will fix" or "Noted, deferring"
6. **Always produce an action plan** — no vague "I'll look into it"
7. **NEVER post to GitHub** — this skill reads PR comments and produces an action plan. The human decides what to fix and how to respond. All GitHub mutations are human-only.
