#!/usr/bin/env bash
set -euo pipefail

# Integration Test: Session Lifecycle
# Tests full session creation, listing, and ending

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI="$PROJECT_DIR/packages/cli/lazyai-cli"

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 Integration Test: Session Lifecycle"
echo "═══════════════════════════════════════════════════════════════"

# Test 1: Create session
echo ""
echo "Test 1: Create session"
SESSION_OUTPUT=$($CLI session start "Integration test session" 2>&1)
echo "$SESSION_OUTPUT"

# Extract session ID - look for "ses_" prefix
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -o 'ses_[0-9]*' | head -1)
if [ -z "$SESSION_ID" ]; then
    echo "❌ FAIL: Could not extract session ID"
    exit 1
fi
echo "✅ Session created: $SESSION_ID"

# Test 2: List sessions
echo ""
echo "Test 2: List sessions"
LIST_OUTPUT=$($CLI session list 2>&1)
if echo "$LIST_OUTPUT" | grep -q "$SESSION_ID"; then
    echo "✅ Session found in list"
else
    echo "❌ FAIL: Session not found in list"
    exit 1
fi

# Test 3: Show session details
echo ""
echo "Test 3: Show session details"
SHOW_OUTPUT=$($CLI session show "$SESSION_ID" 2>&1)
if echo "$SHOW_OUTPUT" | grep -q "Integration test session"; then
    echo "✅ Session details correct"
else
    echo "❌ FAIL: Session details incorrect"
    exit 1
fi

# Test 4: End session
echo ""
echo "Test 4: End session"
END_OUTPUT=$($CLI session end "$SESSION_ID" 2>&1)
echo "$END_OUTPUT"
if echo "$END_OUTPUT" | grep -q "ended"; then
    echo "✅ Session ended successfully"
else
    echo "❌ FAIL: Could not end session"
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ All session integration tests passed!"
echo "═══════════════════════════════════════════════════════════════"
