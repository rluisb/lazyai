#!/usr/bin/env bash
# generate-standup.sh — Generate formatted Slack standup message
# Usage: ./generate-standup.sh [--yesterday <date>] [--output file]
# Default: uses yesterday's date for "yesterday" section

set -euo pipefail

YESTERDAY=$(date -v-1d +"%Y-%m-%d" 2>/dev/null || date -d "yesterday" +"%Y-%m-%d" 2>/dev/null || echo "")
OUTPUT_FILE=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --yesterday) YESTERDAY="$2"; shift 2 ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        --dry-run) DRY_RUN=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

echo "Gathering standup data..."
echo ""

# Initialize sections
YESTERDAY_ITEMS=""
TODAY_ITEMS=""
BLOCKERS=""

# Check git commits from yesterday
if [[ -n "$YESTERDAY" ]]; then
    COMMITS=$(git log --since="$YESTERDAY" --until="$YESTERDAY 23:59:59" --oneline --no-merges 2>/dev/null || echo "")
    if [[ -n "$COMMITS" ]]; then
        YESTERDAY_ITEMS+="*Commits:*\n"
        while IFS= read -r line; do
            hash=$(echo "$line" | cut -d' ' -f1)
            msg=$(echo "$line" | cut -d' ' -f2-)
            YESTERDAY_ITEMS+="• \`$hash\` $msg\n"
        done <<< "$COMMITS"
        YESTERDAY_ITEMS+="\n"
    fi
fi

# Check recent PRs (last 2 days)
echo "Checking PR activity..."
PR_INFO=""
if command -v gh &>/dev/null; then
    PR_INFO=$(gh pr list --state all --limit 10 --json number,title,state,updatedAt --jq '.[] | select(.updatedAt >= "'"$YESTERDAY"'") | "• #\(.number) \(.title) (\(.state))"' 2>/dev/null || echo "")
fi

if [[ -n "$PR_INFO" ]]; then
    YESTERDAY_ITEMS+="*PR Activity:*\n"
    YESTERDAY_ITEMS+="$PR_INFO\n"
fi

# Check Jira issues (if atlassian MCP is available)
echo "Checking Jira status..."
JIRA_INFO=""
# This would use atlassian MCP tools - placeholder for now
JIRA_INFO="_Check Jira for issues updated recently_"

# Check for blockers
echo "Checking for blockers..."
BLOCKER_INFO=""
if command -v gh &>/dev/null; then
    BLOCKER_INFO=$(gh pr list --state open --limit 20 --json number,title,reviewRequests --jq '.[] | select(.reviewRequests != null and (.reviewRequests | length) > 0) | "• #\(.number) \(.title) - awaiting review"' 2>/dev/null || echo "")
fi

# Generate output
{
    echo "*Yesterday:*"
    if [[ -n "$YESTERDAY_ITEMS" ]]; then
        echo -e "$YESTERDAY_ITEMS"
    else
        echo "_No commits or PR activity found for yesterday_"
    fi
    echo ""
    echo "*Today:*"
    echo "_Check Jira for issues in progress_"
    echo ""
    echo "*Blockers:*"
    if [[ -n "$BLOCKER_INFO" ]]; then
        echo "$BLOCKER_INFO"
    else
        echo "None"
    fi
} > /tmp/standup-output.txt

if [[ -n "$OUTPUT_FILE" ]]; then
    cp /tmp/standup-output.txt "$OUTPUT_FILE"
    echo "✅ Standup saved to: $OUTPUT_FILE"
fi

echo ""
echo "=== SLACK STANDUP ==="
cat /tmp/standup-output.txt
echo "====================="
echo ""
echo "Copy the above and paste into Slack standup channel."
echo "Edit as needed before posting."
