#!/usr/bin/env bash
# refresh-service.sh — Safe refresh of local branch + dev container
# Usage: ./refresh-service.sh <service> [--force]
# Example: ./refresh-service.sh fedora
#          ./refresh-service.sh mono-frontend
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <service> [--force]"
    echo "  --force: Skip running-service guard"
    exit 1
fi

SERVICE="$1"
FORCE="${2:-}"

echo "=== Refresh Dev Containers ==="
echo "Service: $SERVICE"
echo ""

# Step 0: Running-service guard
if [[ "$FORCE" != "--force" ]]; then
    echo "--- Checking if service is running ---"
    if dev list --json 2>/dev/null | grep -q "\"$SERVICE\".*running"; then
        echo "⚠️  Service '$SERVICE' appears to be running."
        echo "Refresh may interrupt your work. Continue? (y/N)"
        read -r CONFIRM
        if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
            echo "Refresh cancelled by user."
            exit 0
        fi
    fi
fi

# Step 1: Check branch + working tree
echo "--- Git Status ---"
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
echo "Current branch: $BRANCH"

STASH_NEEDED=false
if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
    echo "Working tree is dirty. Stashing changes..."
    git stash push -u -m "refresh-dev-containers: $SERVICE"
    STASH_NEEDED=true
    echo "Changes stashed."
else
    echo "Working tree is clean."
fi

# Step 2: Update from main
echo "--- Pulling from origin/main ---"
if git pull origin main 2>&1; then
    echo "Pull successful."
else
    echo "❌ Pull failed (possible merge conflict)."
    echo "Please resolve conflicts manually, then re-run this script."
    exit 1
fi

# Step 3: Restore stash
if [[ "$STASH_NEEDED" == true ]]; then
    echo "--- Restoring stashed changes ---"
    if git stash pop 2>&1; then
        echo "Stash restored successfully."
    else
        echo "❌ Stash pop failed (possible conflict)."
        echo "Stash preserved. Run 'git stash list' to see it."
        exit 1
    fi
fi

# Step 4: Service-specific refresh
echo "--- Service Refresh ---"

if [[ "$SERVICE" == "mono-frontend" ]]; then
    echo "mono-frontend detected — using yarn install path"

    # Node version check
    NODE_VERSION=$(node --version 2>/dev/null | cut -d'v' -f2 | cut -d'.' -f1)
    if [[ "$NODE_VERSION" -lt 16 ]] || [[ "$NODE_VERSION" -gt 22 ]]; then
        echo "Switching to Node 22..."
        nvm install 22 2>/dev/null && nvm use 22 2>/dev/null || echo "⚠️  nvm not available, continuing with current Node"
    fi

    echo "Running yarn install..."
    yarn install 2>&1
    RESULT=$?
else
    # Containerized service
    echo "Containerized service detected — using dev update path"

    # Verify service exists
    if ! dev list --json 2>/dev/null | grep -q "$SERVICE"; then
        echo "❌ Service '$SERVICE' not found in dev list."
        echo "Available services:"
        dev list 2>&1
        exit 1
    fi

    # Check for Dev CLI update
    echo "Checking for Dev CLI updates..."
    dev --non-interactive list 2>&1 | grep -q "update available" && echo "Dev CLI update available — will update during refresh"

    # Run dev update
    echo "Running dev update $SERVICE..."
    if dev --non-interactive update "$SERVICE" 2>&1; then
        echo "Dev update successful."
        RESULT=0
    else
        echo "❌ Dev update failed."
        RESULT=1
    fi
fi

# Step 5: Report
echo ""
echo "=== Refresh Dev Containers Report ==="
echo "- Service: $SERVICE"
echo "- Branch status: $([ "$STASH_NEEDED" = true ] && echo 'stashed changes' || echo 'clean')"
echo "- Pull from main: success"
echo "- Stash restore: $([ "$STASH_NEEDED" = true ] && echo 'success' || echo 'not needed')"
echo "- Refresh command: $([ "$SERVICE" = "mono-frontend" ] && echo 'yarn install' || echo "dev update $SERVICE")"
echo "- Refresh result: $([ $RESULT -eq 0 ] && echo 'success' || echo 'failed')"
echo "=== Report Complete ==="

exit $RESULT
