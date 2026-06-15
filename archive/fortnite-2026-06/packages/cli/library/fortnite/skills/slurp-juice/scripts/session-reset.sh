#!/usr/bin/env bash
# session-reset.sh — Full session health check before starting work
# Usage: ./session-reset.sh [--verbose]

set -euo pipefail

VERBOSE=false
[ "${1:-}" = "--verbose" ] && VERBOSE=true

echo "╔══════════════════════════════════════╗"
echo "║     🏥 SESSION HEALTH CHECK          ║"
echo "╚══════════════════════════════════════╝"
echo ""

# 1. Models
echo "📡 Checking models..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if "$SCRIPT_DIR/check-models.sh" 2>/dev/null; then
    echo "   ✅ All primary models available"
else
    echo "   ⚠️  Some models unavailable — check-models.sh for details"
fi

# 2. MCPs
echo "🔌 Checking MCP servers..."
if "$SCRIPT_DIR/mcp-health.sh" 2>/dev/null; then
    echo "   ✅ MCP servers ready"
else
    echo "   ⚠️  Some MCP servers unavailable — mcp-health.sh for details"
fi

# 3. Git state
echo "📁 Git status..."
if [ -d .git ]; then
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    DIRTY=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')
    echo "   Branch: $BRANCH"
    echo "   Uncommitted files: $DIRTY"
    [ "$DIRTY" -gt 0 ] && echo "   ⚠️  Working directory is dirty"
else
    echo "   Not a git repository"
fi

# 4. Worktrees
echo "🌲 Worktrees..."
if [ -d .worktrees ] && [ "$(ls -A .worktrees 2>/dev/null)" ]; then
    WT_COUNT=$(ls -d .worktrees/*/ 2>/dev/null | wc -l | tr -d ' ')
    echo "   Active: $WT_COUNT"
    if "$VERBOSE"; then
        for wt in .worktrees/*/; do
            echo "      $(basename $wt)"
        done
    fi
else
    echo "   None active"
fi

# 5. Specs
echo "📋 Specs..."
if [ -d .specify ] && [ "$(ls -A .specify 2>/dev/null)" ]; then
    SPEC_COUNT=$(find .specify -maxdepth 2 -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
    SLUGS=$(find .specify -maxdepth 2 -name "spec.md" 2>/dev/null | sed 's|/spec.md||' | sed 's|\.specify/||' | tr '\n' ' ')
    echo "   Artifacts: $SPEC_COUNT files"
    echo "   Active slugs: $SLUGS"
else
    echo "   No .specify/ directory — run 'task-init.sh <slug>' to scaffold"
fi

# 6. Memory
echo "🧠 Project memory..."
if [ -d .specify/memory ] && [ "$(ls -A .specify/memory 2>/dev/null)" ]; then
    MEM_COUNT=$(ls .specify/memory/*.md 2>/dev/null | wc -l | tr -d ' ')
    echo "   $MEM_COUNT memory entries"
    if "$VERBOSE"; then
        for mem in .specify/memory/*.md; do
            echo "      $(basename "$mem")"
        done
    fi
fi

# 7. Token budget
echo "💰 Token budget..."
"$SCRIPT_DIR/token-budget.sh" 2>/dev/null || echo "   Could not read token stats"

echo ""
echo "╔══════════════════════════════════════╗"
echo "║     ✅ READY                         ║"
echo "╠══════════════════════════════════════╣"
echo "║  squad-map.sh — see full squad       ║"
echo "║  task-init.sh <slug> — new task      ║"
echo "╚══════════════════════════════════════╝"
