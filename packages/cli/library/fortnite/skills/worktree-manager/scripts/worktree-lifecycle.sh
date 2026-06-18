#!/usr/bin/env bash
# worktree-lifecycle.sh — Full worktree lifecycle management
# Usage: ./worktree-lifecycle.sh <create|provision|start|stop|merge|cleanup> <service> <ticket>

set -euo pipefail

ACTION="${1:-}"
SERVICE="${2:-}"
TICKET="${3:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }
err() { echo -e "${RED}❌ $1${NC}"; exit 1; }
info() { echo "ℹ️  $1"; }

if [ -z "$ACTION" ]; then
    echo "Usage: $0 <create|provision|start|stop|merge|cleanup> <service> <ticket>"
    echo ""
    echo "Actions:"
    echo "  create     Create worktree + container for a ticket"
    echo "  provision  Update container dependencies"
    echo "  start      Start the service in worktree container"
    echo "  stop       Stop the worktree container"
    echo "  merge      Merge worktree changes back to main"
    echo "  cleanup    Remove worktree and container"
    exit 1
fi

if [ -z "$SERVICE" ] || [ -z "$TICKET" ]; then
    err "Service and ticket are required for this action"
fi

WT_PATH="fedora-worktrees/$TICKET"
CONTAINER_NAME="${SERVICE}-wt-$(echo "$TICKET" | tr '[:upper:]' '[:lower:]' | tr '_' '-')"

case "$ACTION" in
    create)
        echo "## Creating worktree for $SERVICE/$TICKET"

        # Check if worktree already exists
        if [ -d "$WT_PATH" ]; then
            warn "Worktree already exists at $WT_PATH"
            exit 0
        fi

        # Create worktree + container
        info "Running: dev worktree $SERVICE $TICKET"
        dev worktree "$SERVICE" "$TICKET" --non-interactive 2>&1 || {
            err "Failed to create worktree"
        }

        ok "Worktree created at $WT_PATH"
        info "Container: $CONTAINER_NAME"
        info "Next: provision with './worktree-lifecycle.sh provision $SERVICE $TICKET'"
        ;;

    provision)
        echo "## Provisioning worktree for $SERVICE/$TICKET"

        if [ ! -d "$WT_PATH" ]; then
            err "Worktree not found at $WT_PATH — create first"
        fi

        info "Running: dev update $SERVICE --worktree-path $WT_PATH"
        dev update "$SERVICE" --worktree-path "$WT_PATH" --non-interactive 2>&1 || {
            err "Failed to provision worktree"
        }

        ok "Worktree provisioned"
        info "Next: start with './worktree-lifecycle.sh start $SERVICE $TICKET'"
        ;;

    start)
        echo "## Starting worktree for $SERVICE/$TICKET"

        if [ ! -d "$WT_PATH" ]; then
            err "Worktree not found at $WT_PATH — create first"
        fi

        info "Running: dev start $SERVICE --worktree-path $WT_PATH"
        dev start "$SERVICE" --worktree-path "$WT_PATH" --non-interactive 2>&1 || {
            err "Failed to start worktree"
        }

        ok "Worktree started"
        info "Container: $CONTAINER_NAME"
        info "Run commands: dev exec $CONTAINER_NAME -- <command>"
        ;;

    stop)
        echo "## Stopping worktree for $SERVICE/$TICKET"

        if [ ! -d "$WT_PATH" ]; then
            warn "Worktree not found at $WT_PATH — may already be cleaned up"
            exit 0
        fi

        info "Running: dev stop $SERVICE --worktree-path $WT_PATH"
        dev stop "$SERVICE" --worktree-path "$WT_PATH" --non-interactive 2>&1 || {
            warn "Failed to stop worktree (may already be stopped)"
        }

        ok "Worktree stopped"
        ;;

    merge)
        echo "## Merging worktree for $SERVICE/$TICKET"

        if [ ! -d "$WT_PATH" ]; then
            err "Worktree not found at $WT_PATH"
        fi

        # Check for uncommitted changes
        cd "$WT_PATH"
        if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
            err "Uncommitted changes in worktree — commit or stash first"
        fi

        # Check if branch is pushed
        BRANCH=$(git branch --show-current 2>/dev/null)
        info "Branch: $BRANCH"
        info "Worktree path: $WT_PATH"
        echo ""
        warn "Merge requires user approval"
        info "Manual steps:"
        info "  1. cd $WT_PATH"
        info "  2. git push origin $BRANCH (if not already pushed)"
        info "  3. Create PR or merge manually"
        info "  4. Run cleanup after merge"
        ;;

    cleanup)
        echo "## Cleaning up worktree for $SERVICE/$TICKET"

        # Stop container if running
        if [ -d "$WT_PATH" ]; then
            info "Stopping worktree container..."
            dev stop "$SERVICE" --worktree-path "$WT_PATH" --non-interactive 2>/dev/null || true
        fi

        # Remove container
        if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
            info "Removing container: $CONTAINER_NAME"
            docker rm "$CONTAINER_NAME" 2>/dev/null || true
            ok "Container removed"
        else
            info "Container $CONTAINER_NAME not found (already removed?)"
        fi

        # Remove worktree directory
        if [ -d "$WT_PATH" ]; then
            info "Removing worktree directory: $WT_PATH"
            rm -rf "$WT_PATH"
            ok "Worktree directory removed"
        fi

        # Remove branch
        BRANCH_NAME="feature/$(echo "$TICKET" | tr '[:upper:]' '[:lower:]')"
        if git rev-parse --verify "$BRANCH_NAME" &>/dev/null; then
            if git branch --merged main 2>/dev/null | grep -q "$BRANCH_NAME"; then
                info "Removing merged branch: $BRANCH_NAME"
                git branch -d "$BRANCH_NAME" 2>/dev/null || true
                ok "Branch removed"
            else
                warn "Branch $BRANCH_NAME not merged — keeping"
                info "  Manual removal: git branch -D $BRANCH_NAME"
            fi
        fi

        ok "Cleanup complete for $SERVICE/$TICKET"
        ;;

    *)
        err "Unknown action: $ACTION"
        ;;
esac
