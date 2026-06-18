#!/usr/bin/env bash
# gh-dash-team-reviews.sh — Query gh-dash config sections for team member review requests
# Usage: ./gh-dash-team-reviews.sh [--notify] [--json] [--reviewer <username>]
#
# --notify: Output in Slack-notification format
# --json: Output as JSON
# --reviewer: Check specific reviewer only

set -euo pipefail

NOTIFY=false
JSON_OUTPUT=false
SPECIFIC_REVIEWER=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --notify) NOTIFY=true; shift ;;
        --json) JSON_OUTPUT=true; shift ;;
        --reviewer) SPECIFIC_REVIEWER="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Check if gh CLI is available
if ! command -v gh &>/dev/null; then
    echo "❌ Error: GitHub CLI (gh) is required but not installed."
    echo "   Install: brew install gh"
    echo "   Auth: gh auth login"
    exit 1
fi

# Check if authenticated
if ! gh auth status &>/dev/null; then
    echo "❌ Error: Not authenticated with GitHub CLI."
    echo "   Run: gh auth login"
    exit 1
fi

# Team reviewers (matches gh-dash config sections)
REVIEWERS=(
    "KimGonzales"
    "nicholashoyte-teachable"
    "alisonbuki"
    "daviddelossantos-teachable"
    "ronaldopassos-teachable"
)

# Filter to specific reviewer if requested
if [[ -n "$SPECIFIC_REVIEWER" ]]; then
    REVIEWERS=("$SPECIFIC_REVIEWER")
fi

echo "🔍 Checking gh-dash team review sections..."
echo ""

RESULTS=()

for reviewer in "${REVIEWERS[@]}"; do
    # Same filter as gh-dash config: is:open is:pr review-requested:<user> -author:<user> draft:false
    prs=$(gh pr list --search "review-requested:$reviewer -author:$reviewer draft:false" --state open --limit 50 --json number,title,author,repository,createdAt,url --jq ".[] | \"[\(.repository.nameWithOwner)] #\(.number) \(.title) by \(.author.login)\"" 2>/dev/null || echo "")

    if [[ -n "$prs" ]]; then
        while IFS= read -r pr; do
            RESULTS+=("📋 *Review requested from $reviewer:* $pr")
        done <<< "$prs"
    fi
done

# Output results
if [[ ${#RESULTS[@]} -eq 0 ]]; then
    echo "✅ No pending team review requests found."
else
    echo "📊 Found ${#RESULTS[@]} PR(s) with team review requests:"
    echo ""

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "["
        for i in "${!RESULTS[@]}"; do
            echo "  \"${RESULTS[$i]}\"$([ $i -lt $((${#RESULTS[@]}-1)) ] && echo "," || echo "")"
        done
        echo "]"
    elif [[ "$NOTIFY" == true ]]; then
        # Slack notification format
        echo "🔔 *PR Review Alerts — Team Watch*"
        echo ""
        for result in "${RESULTS[@]}"; do
            echo "$result"
        done
        echo ""
        echo "_Run gh-dash-team-reviews.sh for details_"
    else
        for result in "${RESULTS[@]}"; do
            echo "$result"
        done
    fi
fi

echo ""
echo "---"
echo "Checked at: $(date +"%Y-%m-%d %H:%M %Z")"
echo "Reviewers: ${REVIEWERS[*]}"
echo ""
echo "💡 Tip: Run 'gh dash' for interactive PR management with team review sections"
