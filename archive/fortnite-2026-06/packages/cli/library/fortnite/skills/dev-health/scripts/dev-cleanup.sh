#!/usr/bin/env bash
# dev-cleanup.sh — Clean up stale worktrees, stopped containers, orphaned branches
# Usage: ./dev-cleanup.sh [containers|worktrees|branches|all] [--dry-run] [--age N]

set -euo pipefail

DRY_RUN=false
AGE_DAYS=7
TARGET="all"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --age) AGE_DAYS="$2"; shift 2 ;;
        containers|worktrees|branches|all) TARGET="$1"; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }
ok() { echo -e "${GREEN}✅ $1${NC}"; }
info() { echo "ℹ️  $1"; }

DRY_PREFIX=""
if [ "$DRY_RUN" = true ]; then
    DRY_PREFIX="[DRY RUN] "
    echo "=== Dry Run Mode — No changes will be made ==="
fi

run_cmd() {
    if [ "$DRY_RUN" = true ]; then
        echo "${DRY_PREFIX}Would run: $*"
    else
        echo "Running: $*"
        eval "$@"
    fi
}

# ─── Container Cleanup ───
cleanup_containers() {
    echo ""
    echo "=== Stopped Containers (older than $AGE_DAYS days) ==="

    if ! docker info &>/dev/null; then
        warn "Docker not available"
        return
    fi

    # Find stopped containers older than N days
    CONTAINERS=$(docker ps -a --filter "status=exited" --filter "status=dead" --format '{{.Names}}|{{.Status}}' 2>/dev/null || true)

    if [ -z "$CONTAINERS" ]; then
        ok "No stopped containers to clean"
        return
    fi

    echo "$CONTAINERS" | while IFS='|' read -r name status; do
        # Check age from status string (e.g., "Exited (0) 5 days ago")
        if echo "$status" | grep -qP "$AGE_DAYS.*days? ago|weeks? ago|months? ago"; then
            warn "Stale: $name ($status)"
            run_cmd "docker rm $name"
        else
            info "Recent: $name ($status) — skipping"
        fi
    done
}

# ─── Worktree Cleanup ───
cleanup_worktrees() {
    echo ""
    echo "=== Stale Agent Worktrees ==="

    if [ ! -d ".worktrees" ]; then
        ok "No .worktrees/ directory"
        return
    fi

    for wt in .worktrees/*/; do
        [ -d "$wt" ] || continue
        name=$(basename "$wt")
        branch="wt/$name"

        # Check if branch exists
        if ! git rev-parse --verify "$branch" &>/dev/null; then
            warn "Orphaned worktree: $name (branch $branch doesn't exist)"
            run_cmd "rm -rf $wt"
            continue
        fi

        # Check if branch is merged
        if git branch --merged main 2>/dev/null | grep -q "$branch"; then
            warn "Merged worktree: $name (branch $branch is merged into main)"
            run_cmd "git worktree remove $wt --force 2>/dev/null || rm -rf $wt"
            run_cmd "git branch -d $branch"
        else
            # Check age
            age=$(find "$wt" -maxdepth 0 -mtime +"$AGE_DAYS" -print 2>/dev/null | wc -l | tr -d ' ')
            if [ "$age" -gt 0 ]; then
                warn "Stale worktree: $name (>$AGE_DAYS days old, not merged)"
                info "  → Manual review needed before removal"
            else
                ok "Active worktree: $name"
            fi
        fi
    done
}

# ─── Branch Cleanup ───
cleanup_branches() {
    echo ""
    echo "=== Orphaned Branches (wt/* without worktrees) ==="

    # Find wt/* branches that don't have a corresponding worktree
    git branch --list 'wt/*' 2>/dev/null | while read -r branch; do
        branch=$(echo "$branch" | tr -d ' *')
        name="${branch#wt/}"

        if [ ! -d ".worktrees/$name" ]; then
            # Check if merged
            if git branch --merged main 2>/dev/null | grep -q "$branch"; then
                warn "Merged orphan: $branch"
                run_cmd "git branch -d $branch"
            else
                warn "Unmerged orphan: $branch"
                info "  → Manual review needed before deletion"
            fi
        fi
    done

    # Also check for stale feature branches (opencode.*)
    echo ""
    echo "=== Stale Feature Worktrees (opencode.*) ==="
    git worktree list 2>/dev/null | grep "opencode\." | while read -r path rest; do
        name=$(basename "$path")
        age=$(find "$path" -maxdepth 0 -mtime +"$AGE_DAYS" -print 2>/dev/null | wc -l | tr -d ' ')
        if [ "$age" -gt 0 ]; then
            warn "Stale: $name at $path (>$AGE_DAYS days)"
            info "  → Manual removal: git worktree remove $path"
        fi
    done
}

# ─── Docker Image Cleanup ───
cleanup_images() {
    echo ""
    echo "=== Dangling Docker Images ==="

    if ! docker info &>/dev/null; then
        warn "Docker not available"
        return
    fi

    DANGLING=$(docker images --filter "dangling=true" -q 2>/dev/null | wc -l | tr -d ' ')
    if [ "$DANGLING" -gt 0 ]; then
        warn "$DANGLING dangling images found"
        run_cmd "docker image prune -f"
    else
        ok "No dangling images"
    fi
}

# ─── Main ───
echo "## Dev Environment Cleanup"
echo "Target: $TARGET | Age threshold: $AGE_DAYS days | Dry run: $DRY_RUN"

case "$TARGET" in
    containers)
        cleanup_containers
        ;;
    worktrees)
        cleanup_worktrees
        cleanup_branches
        ;;
    branches)
        cleanup_branches
        ;;
    all)
        cleanup_containers
        cleanup_worktrees
        cleanup_branches
        cleanup_images
        ;;
    *)
        echo "Unknown target: $TARGET"
        echo "Usage: $0 [containers|worktrees|branches|all] [--dry-run] [--age N]"
        exit 1
        ;;
esac

echo ""
if [ "$DRY_RUN" = true ]; then
    echo "=== Dry run complete. Remove --dry-run to execute cleanup ==="
else
    echo "=== Cleanup complete ==="
fi
