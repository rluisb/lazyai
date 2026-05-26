#!/usr/bin/env bash
# check-pr-reviews.sh — Check for PRs awaiting review from specific reviewers
# Usage: ./check-pr-reviews.sh [--reviewers <list>] [--repos <list>] [--notify] [--json]
#
# Reviewers: comma-separated GitHub usernames
# Repos: comma-separated org/repo pairs
# --notify: Output in Slack-notification format
# --json: Output as JSON

set -euo pipefail

# Default configuration
REVIEWERS="${PR_WATCH_REVIEWERS:-KimGonzales,nicholashoyte-teachable,alisonbuki,daviddelossantos-teachable,ronaldopassos-teachable}"
REPOS="${PR_WATCH_REPOS:-}"
NOTIFY=false
JSON_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --reviewers) REVIEWERS="$2"; shift 2 ;;
        --repos) REPOS="$2"; shift 2 ;;
        --notify) NOTIFY=true; shift ;;
        --json) JSON_OUTPUT=true; shift ;;
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

# Parse reviewers into array
IFS=',' read -ra REVIEWER_ARRAY <<< "$REVIEWERS"

# Build search query for PRs with review requests
echo "🔍 Checking for PRs with review requests from: $REVIEWERS"
echo ""

RESULTS=()

for reviewer in "${REVIEWER_ARRAY[@]}"; do
    reviewer=$(echo "$reviewer" | xargs)  # trim whitespace

    if [[ -n "$REPOS" ]]; then
        # Search specific repos
        IFS=',' read -ra REPO_ARRAY <<< "$REPOS"
        for repo in "${REPO_ARRAY[@]}"; do
            repo=$(echo "$repo" | xargs)
            prs=$(gh pr list --repo "$repo" --search "review-requested:$reviewer state:open" --limit 50 --json number,title,author,createdAt,url --jq ".[] | \"[$repo] #\(.number) \(.title) by \(.author.login) (\(.createdAt))\"" 2>/dev/null || echo "")
            if [[ -n "$prs" ]]; then
                while IFS= read -r pr; do
                    RESULTS+=("📋 *Review requested from $reviewer:* $pr")
                done <<< "$prs"
            fi
        done
    else
        # Search all accessible repos (broader search)
        prs=$(gh search prs --review-requested-by "$reviewer" --state open --limit 50 --json number,title,author,repository,createdAt,url --jq ".[] | \"[\(.repository.nameWithOwner)] #\(.number) \(.title) by \(.author.login)\"" 2>/dev/null || echo "")
        if [[ -n "$prs" ]]; then
            while IFS= read -r pr; do
                RESULTS+=("📋 *Review requested from $reviewer:* $pr")
            done <<< "$prs"
        fi
    fi
done

# Output results
if [[ ${#RESULTS[@]} -eq 0 ]]; then
    echo "✅ No pending review requests found for: $REVIEWERS"
else
    echo "📊 Found ${#RESULTS[@]} PR(s) with review requests:"
    echo ""

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "["
        for i in "${!RESULTS[@]}"; do
            echo "  \"${RESULTS[$i]}\"$([ $i -lt $((${#RESULTS[@]}-1)) ] && echo "," || echo "")"
        done
        echo "]"
    elif [[ "$NOTIFY" == true ]]; then
        # Slack notification format
        echo "🔔 *PR Review Alerts*"
        echo ""
        for result in "${RESULTS[@]}"; do
            echo "$result"
        done
        echo ""
        echo "_Run check-pr-reviews.sh for details_"
    else
        for result in "${RESULTS[@]}"; do
            echo "$result"
        done
    fi
fi

echo ""
echo "---"
echo "Checked at: $(date +"%Y-%m-%d %H:%M %Z")"
echo "Reviewers: $REVIEWERS"
if [[ -n "$REPOS" ]]; then
    echo "Repos: $REPOS"
else
    echo "Repos: all accessible"
fi
