#!/usr/bin/env bash
# token-budget.sh — Monitor opencode session token usage with threshold warnings
# Usage: ./token-budget.sh [session-id]

set -euo pipefail

SESSION_ID="${1:-}"

WARNING_THRESHOLD=80000
CRITICAL_THRESHOLD=100000

echo "=== Token Budget Monitor ==="

if [ -z "$SESSION_ID" ]; then
    echo "Checking active session..."
    # Try to get current session stats from opencode
    if opencode stats 2>/dev/null; then
        echo ""
        TOTAL=$(opencode stats 2>/dev/null | grep -i "total\|context" | head -5)
        echo "$TOTAL"
    else
        echo "No session ID provided and no active session detected."
        echo "Usage: ./token-budget.sh [session-id]"
        echo "       opencode stats  # to see current usage"
        exit 0
    fi
else
    echo "Session: $SESSION_ID"
    opencode export "$SESSION_ID" 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    msgs = data.get('messages', [])
    est_tokens = sum(len(m.get('content',''))//4 for m in msgs)
    print(f'Messages: {len(msgs)}')
    print(f'Estimated tokens: {est_tokens:,}')
    if est_tokens > $CRITICAL_THRESHOLD:
        print('🔴 CRITICAL — Context window nearly full. Compress immediately.')
    elif est_tokens > $WARNING_THRESHOLD:
        print('🟡 WARNING — Approaching context limit. Plan compression.')
    else:
        print(f'🟢 OK — {est_tokens:,}/{100000:,} tokens used')
except:
    print('Could not parse session data')
"
fi

echo ""
echo "Thresholds: 🟡 WARNING=$WARNING_THRESHOLD | 🔴 CRITICAL=$CRITICAL_THRESHOLD"
echo "Tip: Use 'compress' tool when approaching critical threshold"
