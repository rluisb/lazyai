#!/usr/bin/env bash
# setup-gh-dash-team-sections.sh — Add team review sections to gh-dash config
# Usage: ./setup-gh-dash-team-sections.sh
#
# This script adds team review sections to your gh-dash config
# so you can see PRs waiting for review from specific team members.

set -euo pipefail

CONFIG_FILE="$HOME/.config/gh-dash/config.yml"

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "❌ gh-dash config not found at: $CONFIG_FILE"
    echo "   Run 'gh dash' first to generate the config."
    exit 1
fi

# Check if team sections already exist
if grep -q "Team Reviews · Kim" "$CONFIG_FILE" 2>/dev/null; then
    echo "✅ Team review sections already exist in gh-dash config."
    echo "   No changes needed."
    exit 0
fi

echo "📝 Adding team review sections to gh-dash config..."

# Create backup
cp "$CONFIG_FILE" "${CONFIG_FILE}.bak"
echo "   Backup created: ${CONFIG_FILE}.bak"

# Add team review sections after "My Drafts" section
# Using sed to insert after the "My Drafts" filters line
sed -i '' '/filters: is:open is:pr author:@me draft:true sort:updated-desc/a\
  - title: "Team Reviews · Kim"\
    filters: is:open is:pr review-requested:KimGonzales -author:KimGonzales draft:false\
  - title: "Team Reviews · Nick"\
    filters: is:open is:pr review-requested:nicholashoyte-teachable -author:nicholashoyte-teachable draft:false\
  - title: "Team Reviews · Alison"\
    filters: is:open is:pr review-requested:alisonbuki -author:alisonbuki draft:false\
  - title: "Team Reviews · David"\
    filters: is:open is:pr review-requested:daviddelossantos-teachable -author:daviddelossantos-teachable draft:false\
  - title: "Team Reviews · Ronaldo"\
    filters: is:open is:pr review-requested:ronaldopassos-teachable -author:ronaldopassos-teachable draft:false' "$CONFIG_FILE"

echo "✅ Team review sections added to gh-dash config."
echo ""
echo "New sections:"
echo "  - Team Reviews · Kim"
echo "  - Team Reviews · Nick"
echo "  - Team Reviews · Alison"
echo "  - Team Reviews · David"
echo "  - Team Reviews · Ronaldo"
echo ""
echo "Run 'gh dash' to see the new sections."
echo "If something went wrong, restore backup:"
echo "  cp ${CONFIG_FILE}.bak ${CONFIG_FILE}"
