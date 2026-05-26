#!/usr/bin/env bash
set -euo pipefail

# Integration Test: Ledger Integrity
# Tests ledger initialization, append, and verification

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI="$PROJECT_DIR/packages/cli/lazyai-cli"
LEDGER_FILE="$PROJECT_DIR/.specify/ledger.jsonl"

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 Integration Test: Ledger Integrity"
echo "═══════════════════════════════════════════════════════════════"

# Test 1: Verify ledger exists
echo ""
echo "Test 1: Ledger file exists"
if [ -f "$LEDGER_FILE" ]; then
    echo "✅ Ledger file exists"
else
    echo "❌ FAIL: Ledger file not found at $LEDGER_FILE"
    exit 1
fi

# Test 2: Verify ledger has entries
echo ""
echo "Test 2: Ledger has entries"
ENTRY_COUNT=$(wc -l < "$LEDGER_FILE" | tr -d ' ')
if [ "$ENTRY_COUNT" -gt 0 ]; then
    echo "✅ Ledger has $ENTRY_COUNT entries"
else
    echo "❌ FAIL: Ledger is empty"
    exit 1
fi

# Test 3: Verify genesis entry exists
echo ""
echo "Test 3: Genesis entry exists"
if head -1 "$LEDGER_FILE" | grep -q "genesis"; then
    echo "✅ Genesis entry found"
else
    echo "❌ FAIL: Genesis entry not found"
    exit 1
fi

# Test 4: Verify hash chain integrity
echo ""
echo "Test 4: Hash chain integrity"
VERIFY_OUTPUT=$($CLI ledger verify 2>&1)
echo "$VERIFY_OUTPUT"
if echo "$VERIFY_OUTPUT" | grep -q "intact"; then
    echo "✅ Hash chain is intact"
else
    echo "❌ FAIL: Hash chain is broken"
    exit 1
fi

# Test 5: Show ledger
echo ""
echo "Test 5: Show ledger"
SHOW_OUTPUT=$($CLI ledger show 2>&1)
if echo "$SHOW_OUTPUT" | grep -q "entries"; then
    echo "✅ Ledger show works"
else
    echo "❌ FAIL: Ledger show failed"
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ All ledger integration tests passed!"
echo "═══════════════════════════════════════════════════════════════"
