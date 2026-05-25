#!/usr/bin/env bash
# collect-feedback.sh — Collect PR comments and review comments into structured JSON
# Usage: ./collect-feedback.sh [PR_NUMBER] [REPO]
# Output: feedback-<PR_NUMBER>.json

set -euo pipefail

PR_NUMBER="${1:-}"
REPO="${2:-}"

if [ -z "$PR_NUMBER" ]; then
  # Try to get current branch PR
  PR_NUMBER=$(gh pr view --json number -q '.number' 2>/dev/null || echo "")
  if [ -z "$PR_NUMBER" ]; then
    echo "Error: No PR number provided and no PR found for current branch"
    echo "Usage: $0 <PR_NUMBER> [REPO]"
    exit 1
  fi
fi

REPO_FLAG=""
if [ -n "$REPO" ]; then
  REPO_FLAG="-R $REPO"
fi

echo "Collecting feedback for PR #$PR_NUMBER..."

# Fetch PR details
PR_JSON=$(gh pr view $PR_NUMBER $REPO_FLAG \
  --json title,body,author,state,headRefName,baseRefName,labels,createdAt,updatedAt,changedFiles,additions,deletions \
  2>/dev/null)

# Fetch PR comments
COMMENTS_JSON=$(gh pr view $PR_NUMBER $REPO_FLAG --comments \
  --json comments \
  -q '.comments[] | {author: .author.login, body: .body, createdAt: .createdAt}' 2>/dev/null || echo "[]")

# Fetch review comments (inline comments on code)
REVIEW_COMMENTS=$(gh api repos/{owner}/{repo}/pulls/$PR_NUMBER/comments \
  --jq '.[] | {author: .user.login, body: .body, path: .path, line: .line, side: .side, commitId: .commit_id, createdAt: .created_at}' 2>/dev/null || echo "[]")

# Extract Jira ticket reference from PR body or branch name
JIRA_TICKET=$(echo "$PR_JSON" | jq -r '.body // ""' | grep -oE '[A-Z]+-[0-9]+' | head -1 || echo "")
if [ -z "$JIRA_TICKET" ]; then
  JIRA_TICKET=$(echo "$PR_JSON" | jq -r '.headRefName // ""' | grep -oE '[A-Z]+-[0-9]+' | head -1 || echo "")
fi

# Build output
OUTPUT=$(jq -n \
  --arg prNumber "$PR_NUMBER" \
  --argjson pr "$PR_JSON" \
  --argjson comments "$COMMENTS_JSON" \
  --argjson reviewComments "$REVIEW_COMMENTS" \
  --arg jiraTicket "$JIRA_TICKET" \
  '{
    prNumber: ($prNumber | tonumber),
    pr: $pr,
    jiraTicket: (if $jiraTicket != "" then $jiraTicket else null end),
    comments: $comments,
    reviewComments: $reviewComments,
    collectedAt: (now | todate),
    totalComments: (($comments | length) + ($reviewComments | length))
  }')

OUTPUT_FILE="feedback-${PR_NUMBER}.json"
echo "$OUTPUT" | jq '.' > "$OUTPUT_FILE"

echo "✅ Feedback collected: $OUTPUT_FILE"
echo "   PR: #$PR_NUMBER"
echo "   Jira: ${JIRA_TICKET:-not found}"
echo "   Comments: $(echo "$COMMENTS_JSON" | jq 'length')"
echo "   Review comments: $(echo "$REVIEW_COMMENTS" | jq 'length')"
echo "   Total: $(echo "$OUTPUT" | jq '.totalComments')"
