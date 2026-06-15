#!/usr/bin/env bash
# pr-context.sh — Gather all context for a PR review
# Usage: ./pr-context.sh <PR_NUMBER> [REPO]
# Output: pr-context-<PR_NUMBER>.json

set -euo pipefail

PR_NUMBER="${1:-}"
REPO="${2:-}"

if [ -z "$PR_NUMBER" ]; then
  echo "Error: PR number required"
  echo "Usage: $0 <PR_NUMBER> [REPO]"
  exit 1
fi

REPO_FLAG=""
if [ -n "$REPO" ]; then
  REPO_FLAG="-R $REPO"
fi

echo "Gathering context for PR #$PR_NUMBER..."

# Fetch PR details
echo "  → Fetching PR details..."
PR_JSON=$(gh pr view $PR_NUMBER $REPO_FLAG \
  --json title,body,author,state,headRefName,baseRefName,labels,createdAt,updatedAt,changedFiles,additions,deletions,files,commits,reviews,latestReviews,mergeable,mergeStateStatus 2>/dev/null)

# Fetch PR comments
echo "  → Fetching comments..."
COMMENTS_JSON=$(gh pr view $PR_NUMBER $REPO_FLAG --comments \
  --json comments -q '.comments' 2>/dev/null || echo "[]")

# Fetch review comments (inline)
echo "  → Fetching review comments..."
REVIEW_COMMENTS=$(gh api repos/{owner}/{repo}/pulls/$PR_NUMBER/comments 2>/dev/null || echo "[]")

# Fetch diff summary
echo "  → Fetching diff summary..."
DIFF_STATS=$(gh pr diff $PR_NUMBER $REPO_FLAG --name-only 2>/dev/null || echo "")

# Fetch CI checks
echo "  → Fetching CI checks..."
CHECKS_JSON=$(gh pr checks $PR_NUMBER $REPO_FLAG --json name,status,conclusion,startedAt,completedAt 2>/dev/null || echo "[]")

# Extract Jira ticket reference
JIRA_TICKET=$(echo "$PR_JSON" | jq -r '.body // ""' | grep -oE '[A-Z]+-[0-9]+' | head -1 || echo "")
if [ -z "$JIRA_TICKET" ]; then
  JIRA_TICKET=$(echo "$PR_JSON" | jq -r '.headRefName // ""' | grep -oE '[A-Z]+-[0-9]+' | head -1 || echo "")
fi
if [ -z "$JIRA_TICKET" ]; then
  JIRA_TICKET=$(echo "$PR_JSON" | jq -r '.labels[].name // ""' | grep -E '[A-Z]+-[0-9]+' | head -1 || echo "")
fi

# Extract Confluence links from PR body
CONFLUENCE_LINKS=$(echo "$PR_JSON" | jq -r '.body // ""' | grep -oE 'https://[^ ]+atlassian\.net/wiki[^ ]*' | sort -u || echo "")

# Build output
OUTPUT=$(jq -n \
  --arg prNumber "$PR_NUMBER" \
  --argjson pr "$PR_JSON" \
  --argjson comments "$COMMENTS_JSON" \
  --argjson reviewComments "$REVIEW_COMMENTS" \
  --arg diffStats "$DIFF_STATS" \
  --argjson checks "$CHECKS_JSON" \
  --arg jiraTicket "$JIRA_TICKET" \
  --arg confluenceLinks "$CONFLUENCE_LINKS" \
  '{
    prNumber: ($prNumber | tonumber),
    pr: $pr,
    jiraTicket: (if $jiraTicket != "" then $jiraTicket else null end),
    confluenceLinks: ($confluenceLinks | split("\n") | map(select(. != ""))),
    comments: $comments,
    reviewComments: $reviewComments,
    diffFiles: ($diffStats | split("\n") | map(select(. != ""))),
    checks: $checks,
    collectedAt: (now | todate),
    totalComments: (($comments | length) + ($reviewComments | length))
  }')

OUTPUT_FILE="pr-context-${PR_NUMBER}.json"
echo "$OUTPUT" | jq '.' > "$OUTPUT_FILE"

echo "✅ Context collected: $OUTPUT_FILE"
echo "   PR: #$PR_NUMBER"
echo "   Jira: ${JIRA_TICKET:-not found}"
echo "   Confluence links: $(echo "$CONFLUENCE_LINKS" | wc -l | tr -d ' ')"
echo "   Comments: $(echo "$COMMENTS_JSON" | jq 'length')"
echo "   Review comments: $(echo "$REVIEW_COMMENTS" | jq 'length')"
echo "   Changed files: $(echo "$DIFF_STATS" | wc -l | tr -d ' ')"
echo "   CI checks: $(echo "$CHECKS_JSON" | jq 'length')"
