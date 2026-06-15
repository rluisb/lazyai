#!/usr/bin/env bash
# pr-status-summary.sh — Generate summary of all PR status across watched repos
# Usage: ./pr-status-summary.sh [--repos <list>] [--days <N>]
#
# --repos: comma-separated org/repo pairs (default: all accessible)
# --days: lookback period in days (default: 7)

set -euo pipefail

REPOS="${PR_WATCH_REPOS:-}"
DAYS="${PR_WATCH_DAYS:-7}"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repos) REPOS="$2"; shift 2 ;;
        --days) DAYS="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Check if gh CLI is available
if ! command -v gh &>/dev/null; then
    echo "❌ Error: GitHub CLI (gh) is required but not installed."
    exit 1
fi

SINCE=$(date -v-${DAYS}d +"%Y-%m-%d" 2>/dev/null || date -d "$DAYS days ago" +"%Y-%m-%d" 2>/dev/null || echo "")

echo "📊 PR Status Summary — Last $DAYS days"
echo "Generated: $(date +"%Y-%m-%d %H:%M %Z")"
echo ""

# Function to get PR stats for a repo
get_repo_stats() {
    local repo="$1"
    echo "### Repository: $repo"
    echo ""

    # Open PRs
    open_count=$(gh pr list --repo "$repo" --state open --limit 100 --json number 2>/dev/null | jq 'length' || echo "0")
    echo "**Open PRs:** $open_count"

    # PRs awaiting my review
    my_review=$(gh pr list --repo "$repo" --search "review-requested:@me state:open" --limit 100 --json number 2>/dev/null | jq 'length' || echo "0")
    echo "**Awaiting my review:** $my_review"

    # My PRs open
    my_open=$(gh pr list --repo "$repo" --author "@me" --state open --limit 100 --json number 2>/dev/null | jq 'length' || echo "0")
    echo "**My open PRs:** $my_open"

    # Merged in last N days
    if [[ -n "$SINCE" ]]; then
        merged=$(gh pr list --repo "$repo" --state merged --search "merged:>$SINCE" --limit 100 --json number 2>/dev/null | jq 'length' || echo "0")
        echo "**Merged (last $DAYS days):** $merged"
    fi

    echo ""
}

if [[ -n "$REPOS" ]]; then
    IFS=',' read -ra REPO_ARRAY <<< "$REPOS"
    for repo in "${REPO_ARRAY[@]}"; do
        repo=$(echo "$repo" | xargs)
        get_repo_stats "$repo"
    done
else
    echo "⚠️  No repos configured. Set PR_WATCH_REPOS or use --repos flag."
    echo ""
    echo "Example:"
    echo "  export PR_WATCH_REPOS='org/repo1,org/repo2'"
    echo "  ./pr-status-summary.sh"
    echo ""
    echo "Or:"
    echo "  ./pr-status-summary.sh --repos org/repo1,org/repo2"
fi

echo "---"
echo "Tip: Run check-pr-reviews.sh to see specific review requests"
