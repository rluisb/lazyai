#!/usr/bin/env bash
# worktree-status.sh — Inventory of all worktrees with status, age, and container mapping
# Usage: ./worktree-status.sh [ticket-name]

set -euo pipefail

TARGET="${1:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ok() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }
err() { echo -e "${RED}❌ $1${NC}"; }
info() { echo -e "${BLUE}ℹ️  $1${NC}"; }

echo "## Worktree Inventory"
echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# ─── Dev Worktrees (fedora-worktrees/) ───
echo "=== Dev Worktrees (fedora-worktrees/) ==="
echo "| Ticket | Path | Container | Container Status | Branch | Age |"
echo "|--------|------|-----------|------------------|--------|-----|"

if [ -d "fedora-worktrees" ]; then
    for wt in fedora-worktrees/*/; do
        [ -d "$wt" ] || continue
        ticket=$(basename "$wt")

        # Skip if filtering by target
        if [ -n "$TARGET" ] && [[ "$ticket" != *"$TARGET"* ]]; then
            continue
        fi

        # Get branch name
        cd "$wt" 2>/dev/null
        branch=$(git branch --show-current 2>/dev/null || echo "?")
        cd - >/dev/null

        # Container name
        container="fedora-wt-$(echo "$ticket" | tr '[:upper:]' '[:lower:]' | tr '_' '-')"

        # Container status
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"; then
            container_status="running"
        elif docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"; then
            container_status="stopped"
        else
            container_status="none"
        fi

        # Age
        age_days=$(( ($(date +%s) - $(stat -f %m "$wt" 2>/dev/null || stat -c %Y "$wt" 2>/dev/null || echo 0)) / 86400 ))

        # Status icon
        if [ "$container_status" = "running" ]; then
            icon="✅"
        elif [ "$container_status" = "stopped" ]; then
            icon="⚠️"
        else
            icon="❓"
        fi

        echo "| $icon $ticket | $wt | $container | $container_status | $branch | ${age_days}d |"
    done
else
    echo "| (none) | — | — | — | — | — |"
fi

echo ""

# ─── Agent Worktrees (.worktrees/) ───
echo "=== Agent Worktrees (.worktrees/) ==="
echo "| Name | Path | Branch | Merged? | Age |"
echo "|------|------|--------|---------|-----|"

if [ -d ".worktrees" ]; then
    for wt in .worktrees/*/; do
        [ -d "$wt" ] || continue
        name=$(basename "$wt")

        if [ -n "$TARGET" ] && [[ "$name" != *"$TARGET"* ]]; then
            continue
        fi

        branch="wt/$name"

        # Check if merged
        if git branch --merged main 2>/dev/null | grep -q "$branch"; then
            merged="yes"
        else
            merged="no"
        fi

        # Age
        age_days=$(( ($(date +%s) - $(stat -f %m "$wt" 2>/dev/null || stat -c %Y "$wt" 2>/dev/null || echo 0)) / 86400 ))

        # Status
        if [ "$merged" = "yes" ] && [ "$age_days" -gt 7 ]; then
            icon="⚠️"
        elif [ "$merged" = "yes" ]; then
            icon="✅"
        else
            icon="🔄"
        fi

        echo "| $icon $name | $wt | $branch | $merged | ${age_days}d |"
    done
else
    echo "| (none) | — | — | — | — |"
fi

echo ""

# ─── Manual Git Worktrees ───
echo "=== Manual Git Worktrees ==="
echo "| Name | Path | Branch | Age |"
echo "|------|------|--------|-----|"

WT_LIST=$(git worktree list 2>/dev/null | grep -v " (bare)" || true)
if [ -n "$WT_LIST" ]; then
    echo "$WT_LIST" | while read -r path rest; do
        name=$(basename "$path")

        # Skip main worktree
        if [[ "$path" == *".config/opencode" ]] && [[ "$rest" == *"main"* ]]; then
            continue
        fi

        # Skip if filtering
        if [ -n "$TARGET" ] && [[ "$name" != *"$TARGET"* ]]; then
            continue
        fi

        branch=$(echo "$rest" | awk '{print $1}')

        # Age
        age_days=$(( ($(date +%s) - $(stat -f %m "$path" 2>/dev/null || stat -c %Y "$path" 2>/dev/null || echo 0)) / 86400 ))

        echo "| 📁 $name | $path | $branch | ${age_days}d |"
    done
else
    echo "| (none) | — | — | — |"
fi

echo ""

# ─── Summary ───
echo "### Summary"
DEV_COUNT=$(ls -d fedora-worktrees/*/ 2>/dev/null | wc -l | tr -d ' ')
AGENT_COUNT=$(ls -d .worktrees/*/ 2>/dev/null | wc -l | tr -d ' ')
MANUAL_COUNT=$(git worktree list 2>/dev/null | grep -v " (bare)" | wc -l | tr -d ' ')

echo "- Dev worktrees: $DEV_COUNT"
echo "- Agent worktrees: $AGENT_COUNT"
echo "- Manual worktrees: $MANUAL_COUNT"
echo "- Total: $((DEV_COUNT + AGENT_COUNT + MANUAL_COUNT))"
