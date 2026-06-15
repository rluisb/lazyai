#!/usr/bin/env bash
# dev-rebase-safe.sh — Safe rebase with stash protection and conflict handling
# Usage: ./dev-rebase-safe.sh <service> [--reset] [--soft] [--target-branch BRANCH]
set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <service> [--reset] [--soft] [--target-branch BRANCH]"
    exit 1
fi

SERVICE="$1"
shift

FLAGS=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --reset) FLAGS="$FLAGS --reset"; shift ;;
        --soft) FLAGS="$FLAGS --soft"; shift ;;
        --target-branch) FLAGS="$FLAGS --target-branch $2"; shift 2 ;;
        *) echo "Unknown flag: $1"; exit 1 ;;
    esac
done

echo "=== Safe Rebase: $SERVICE ==="
echo "Flags: ${FLAGS:-none}"
echo ""

# Check service exists
if ! dev list --json 2>/dev/null | grep -q "$SERVICE"; then
    echo "ERROR: Service '$SERVICE' not found"
    dev list 2>&1
    exit 1
fi

# Run rebase with non-interactive flag
dev rebase $SERVICE $FLAGS --non-interactive 2>&1
EXIT_CODE=$?

echo ""
if [[ $EXIT_CODE -eq 0 ]]; then
    echo "=== Rebase SUCCESS ==="
else
    echo "=== Rebase FAILED (exit code: $EXIT_CODE) ==="
    echo "Stash preserved for manual resolution"
    echo "Use 'git stash pop' to restore changes"
fi

exit $EXIT_CODE
