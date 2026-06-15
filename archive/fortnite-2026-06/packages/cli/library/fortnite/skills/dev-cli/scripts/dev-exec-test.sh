#!/usr/bin/env bash
# dev-exec-test.sh — Run tests in container with optional dependency update
# Usage: ./dev-exec-test.sh <service> <test-command> [--update-deps]
# Example: ./dev-exec-test.sh fedora "./bin/rspec spec/models/user_spec.rb"
set -euo pipefail

if [[ $# -lt 2 ]]; then
    echo "Usage: $0 <service> <test-command> [--update-deps]"
    echo "Example: $0 fedora './bin/rspec spec/models/user_spec.rb'"
    exit 1
fi

SERVICE="$1"
shift
TEST_CMD="$1"
shift

UPDATE_FLAG=""
if [[ "${1:-}" == "--update-deps" ]]; then
    UPDATE_FLAG="--update-dependencies"
fi

echo "=== Running tests in $SERVICE ==="
echo "Command: $TEST_CMD"
if [[ -n "$UPDATE_FLAG" ]]; then
    echo "Dependencies: Will update before running"
fi
echo ""

dev exec $SERVICE $UPDATE_FLAG -- $TEST_CMD --non-interactive 2>&1
EXIT_CODE=$?

echo ""
if [[ $EXIT_CODE -eq 0 ]]; then
    echo "=== Tests PASSED ==="
else
    echo "=== Tests FAILED (exit code: $EXIT_CODE) ==="
fi

exit $EXIT_CODE
