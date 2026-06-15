#!/usr/bin/env bash
# worktree-manager.sh — Create/list/clean isolated worktrees for parallel execution
# Usage: ./worktree-manager.sh {create|list|clean|merge} [name]

set -euo pipefail

WORKTREE_ROOT="${OPENCODE_WORKSPACE:-.}/.worktrees"
CMD="${1:-help}"
NAME="${2:-}"

mkdir -p "$WORKTREE_ROOT"

case "$CMD" in
    create)
        if [ -z "$NAME" ]; then
            NAME="task-$(date +%s | tail -c 7)"
        fi
        WT_PATH="$WORKTREE_ROOT/$NAME"

        if [ -d "$WT_PATH" ]; then
            echo "Worktree $NAME already exists at $WT_PATH"
            exit 1
        fi

        BRANCH="wt/$NAME"
        BASE_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

        git worktree add -b "$BRANCH" "$WT_PATH" "$BASE_BRANCH"
        echo "✅ Created worktree: $NAME"
        echo "   Path: $WT_PATH"
        echo "   Branch: $BRANCH"
        echo ""
        echo "Dispatch agents to: $WT_PATH"
        ;;

    list)
        echo "=== Active Worktrees ==="
        if [ -d "$WORKTREE_ROOT" ] && [ "$(ls -A "$WORKTREE_ROOT" 2>/dev/null)" ]; then
            for wt in "$WORKTREE_ROOT"/*/; do
                WT_NAME=$(basename "$wt")
                BRANCH=$(cd "$wt" 2>/dev/null && git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
                echo "  $WT_NAME → branch: $BRANCH → $wt"
            done
        else
            echo "  No active worktrees"
        fi
        ;;

    clean)
        if [ -z "$NAME" ]; then
            echo "Cleaning ALL worktrees..."
            for wt in "$WORKTREE_ROOT"/*/; do
                WT_NAME=$(basename "$wt")
                echo "  Removing $WT_NAME..."
                git worktree remove "$wt" --force 2>/dev/null || rm -rf "$wt"
            done
            git branch | grep 'wt/' | xargs -r git branch -D 2>/dev/null || true
            echo "✅ All worktrees cleaned"
        else
            WT_PATH="$WORKTREE_ROOT/$NAME"
            if [ -d "$WT_PATH" ]; then
                git worktree remove "$WT_PATH" --force 2>/dev/null || rm -rf "$WT_PATH"
                git branch -D "wt/$NAME" 2>/dev/null || true
                echo "✅ Removed: $NAME"
            else
                echo "Worktree $NAME not found"
            fi
        fi
        ;;

    merge)
        if [ -z "$NAME" ]; then
            echo "Usage: ./worktree-manager.sh merge <name>"
            echo "Merges the worktree branch back to current branch."
            exit 1
        fi
        BRANCH="wt/$NAME"
        if git branch | grep -q "$BRANCH"; then
            CURRENT=$(git rev-parse --abbrev-ref HEAD)
            echo "Merging $BRANCH → $CURRENT..."
            git merge "$BRANCH" --no-ff -m "Merge worktree: $NAME"
            echo "✅ Merged $NAME"
        else
            echo "Branch $BRANCH not found"
        fi
        ;;

    help|*)
        echo "Worktree Manager — Isolated parallel execution"
        echo ""
        echo "Usage:"
        echo "  ./worktree-manager.sh create [name]    Create isolated worktree"
        echo "  ./worktree-manager.sh list              List all active worktrees"
        echo "  ./worktree-manager.sh clean [name]      Remove worktree(s)"
        echo "  ./worktree-manager.sh merge <name>      Merge worktree branch"
        echo ""
        echo "Worktrees live in: $WORKTREE_ROOT/"
        ;;
esac
