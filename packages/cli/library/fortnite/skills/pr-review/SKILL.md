---
name: pr-review
description: "Review others' pull requests. Uses gh-dash to find PRs, reviews code against Jira tickets/Confluence/specs, rates findings, generates Slack-ready summary, and offers clipboard copy. Use when reviewing team members' PRs."
trigger: /pr-review
---

# /pr-review

Review others' pull requests with structured findings, ratings, and Slack-ready output.

## When to Use

- PRs are assigned to you for review (check `gh dash`)
- Someone asked you to review their PR
- You want to do a thorough review with context from Jira/Confluence/specs
- You need to produce a structured review with ratings and Slack summary

## Workflow

### Phase 1: Discover PRs to Review

```bash
# Use gh-dash to find PRs needing review
gh dash

# Or list PRs requesting your review
gh pr list --review-requested @me --json number,title,author,headRefName,createdAt

# Or check a specific PR
gh pr view <NUMBER> --repo <OWNER/REPO>
```

**gh-dash sections** (from config):
- **Needs My Review** — `is:open is:pr review-requested:@me -author:@me draft:false`
- **Re-review / Follow-up** — `is:open is:pr reviewed-by:@me -author:@me -review:approved`

### Phase 2: Gather Context

```
1. Fetch PR details (title, body, diff, checks, existing comments)
2. Extract Jira ticket reference from PR body, branch name, or labels
3. Fetch Jira ticket (description, acceptance criteria, comments, linked Confluence)
4. Find related Confluence pages (linked from Jira or PR body)
5. Find related specs (bee-gone/specs/ if applicable)
6. Check CI status (all green? any failures?)
```

**Tools:**
- `gh pr view <N> --json ...` — PR details
- `gh pr diff <N>` — Full diff
- `gh pr checks <N>` — CI status
- `gh pr view <N> --comments` — Existing comments
- `atlassian_getJiraIssue` — Jira ticket details
- `atlassian_getConfluencePage` — Confluence pages
- `atlassian_search` — Related specs/docs
- `morph codebase_search` — Understand code patterns

### Phase 3: Review Code

Review the diff against the requirements. For each finding:

**Rating Scale:**

| Rating | Severity | Description | Action |
|--------|----------|-------------|--------|
| 🔴 **Critical** | Blocker | Bug, security issue, data loss, broken functionality | Must fix before merge |
| 🟠 **Major** | Significant | Logic error, missing edge case, performance issue | Should fix before merge |
| 🟡 **Minor** | Improvement | Code quality, readability, missing error handling | Nice to fix, not blocking |
| ⚪ **Nit** | Style | Naming, formatting, comment style | Optional, no discussion needed |

**Review Categories:**

| Category | What to Look For |
|----------|------------------|
| **Correctness** | Does the code do what the ticket asks? Any bugs? |
| **Completeness** | Are all acceptance criteria met? Edge cases handled? |
| **Security** | Input validation, auth checks, SQL injection, XSS, secrets |
| **Performance** | N+1 queries, missing indexes, unnecessary loops, memory leaks |
| **Maintainability** | Code clarity, naming, duplication, test coverage |
| **Consistency** | Follows existing patterns, conventions, style guide |

### Phase 4: Produce Review Output

```markdown
## PR Review — #NNN: <PR Title>

**Author:** @username
**Branch:** feature/branch-name
**Jira:** [PROJ-123](https://jira.url/browse/PROJ-123)
**CI:** ✅ All passing / ❌ <failure details>

### Summary
- 🔴 Critical: N
- 🟠 Major: N
- 🟡 Minor: N
- ⚪ Nit: N

### Findings

#### 🔴 Critical
1. **[File:Line]** — Description of the issue
   - **Why:** Explanation of why this is critical
   - **Evidence:** Reference to spec/ticket/Confluence
   - **Suggestion:** How to fix it

#### 🟠 Major
1. **[File:Line]** — Description
   - **Why:** ...
   - **Evidence:** ...
   - **Suggestion:** ...

#### 🟡 Minor
1. **[File:Line]** — Description
   - **Suggestion:** ...

#### ⚪ Nits
1. **[File:Line]** — Description

### Positive Feedback
- Good use of [pattern/convention]
- Clean implementation of [feature]
- Well-tested [area]

### Verdict
- [ ] ✅ **Approve** — No critical/major findings
- [ ] 🔄 **Request Changes** — Critical/major findings need fixing
- [ ] 💬 **Comment** — Minor findings, author's discretion

### Slack Summary
```
📝 PR Review: #NNN — <Title>
👤 @author | 🔗 <PR URL>
🎫 PROJ-123: <Ticket summary>

Findings: 🔴N 🟠N 🟡N ⚪N

Key points:
• <Critical finding 1>
• <Major finding 1>

Verdict: <Approve/Request Changes/Comment>
```
```

### Phase 5: Generate Review for Human Approval

**⚠️ HARD RULE: This skill NEVER posts to GitHub. It produces output for the human to review and submit.**

The review output is formatted as:
1. A markdown review document (findings, ratings, verdict)
2. A Slack-ready summary
3. A "Suggested comments" section — pre-written inline comments the human can copy/paste

**What the human does:**
- Reviews the findings and adjusts ratings/severity
- Edits the tone of comments to match their voice
- Posts the review themselves via `gh pr review` or the GitHub UI

**What this skill NEVER does:**
- `gh pr review --approve`
- `gh pr review --request-changes`
- `gh pr review --comment`
- `gh pr comment`
- `gh api repos/{owner}/{repo}/pulls/<N>/comments`
- Any command that creates, updates, or deletes GitHub content

### Phase 6: Clipboard (Optional)

Ask the user if they want the Slack summary copied to clipboard:

```bash
# macOS
pbcopy < slack-summary.txt

# Linux
xclip -selection clipboard < slack-summary.txt
```

## Scripts

### `scripts/pr-context.sh`

Gathers all context for a PR: details, diff, Jira ticket, Confluence pages.

```bash
./skills/pr-review/scripts/pr-context.sh <PR_NUMBER> [REPO]
```

Output: `pr-context-<PR_NUMBER>.json` with all gathered context.

## Agent Assignment

- **Primary**: shield-audit (MODE=review, FOCUS=all)
- **Support**: loot-hawk (for codebase research on unfamiliar code)

## Integration with Other Skills

| Skill | When |
|-------|------|
| **feedback-review** | Related — but feedback-review is for reviewing OUR OWN PRs |
| **zero-point** | For deeper security/adversarial review if needed |
| **storm-scout** | If review reveals spec ambiguity |

## Rules

1. **Always check the Jira ticket first** — understand what was requested before reviewing
2. **Reference specs and Confluence** — ground findings in documented requirements
3. **Rate every finding** — use the rating scale consistently
4. **Include positive feedback** — not just criticism
5. **Provide file:line references** — make it easy for the author to find the issue
6. **Generate Slack summary** — make it easy to share findings
7. **NEVER post to GitHub** — produce review output for human approval only. The human submits.
8. **Ask before clipboard** — confirm before copying to clipboard
9. **Use the tone guide** — friendly nerd, not robot. See Comment Tone Guide below.

## Comment Tone Guide

**Voice:** Friendly nerd at a whiteboard. Knows their stuff, doesn't flex. Warm but precise. Uses "we" not "you" — we're on the same squad.

**Do:**
- "Nice approach! One edge case we might want to handle: what happens when the token expires mid-request? A quick `refreshIfNeeded()` before the call would cover it. 🛡️"
- "This is clean. Quick thought — could we extract the retry logic into a shared helper? We've got the same pattern in `payment-service.ts:142` and it'd be nice to DRY it up."
- "Good catch on the null check. While we're here, the linter is also flagging an unused import on line 8 — mind sweeping that up?"

**Don't:**
- "You forgot to handle the edge case." → accusatory
- "This is wrong." → unhelpful, no context
- "Per the spec §3.2, this implementation fails to satisfy requirement AC-4." → robotic, legalese
- "LGTM 👍" → too casual, no substance

**Structure for findings:**
1. Acknowledge what's good (always lead with something positive)
2. State the issue clearly in one sentence
3. Explain why it matters (impact, not just correctness)
4. Suggest a fix (specific, actionable)
5. Close with a light emoji or sign-off

**Emoji palette (use sparingly, 1 per comment max):**
🛡️ security concern | ⚡ performance | 🧹 cleanup/nit | 🐛 bug | 💡 suggestion | 🎯 spec alignment

## gh-dash Integration

Use `gh dash` as the entry point for discovering PRs:

```bash
# Open gh-dash (TUI)
gh dash

# The "Needs My Review" section shows PRs awaiting your review
# Select a PR and press Enter to view details
# Use keybindings:
#   c — checkout PR
#   v — view in browser
#   o — open in browser
```

For programmatic access:

```bash
# List PRs needing review
gh pr list --review-requested @me --json number,title,author,headRefName,createdAt

# View specific PR
gh pr view <N> --json title,body,author,state,headRefName,baseRefName,labels,changedFiles,additions,deletions,files,commits,reviews,latestReviews

# Get diff
gh pr diff <N>

# Get checks
gh pr checks <N>
```
