#!/usr/bin/env bash
set -euo pipefail

# Integration Test: Ledger Integrity
# Tests ledger initialization, append, and verification in an isolated temp project.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI="$PROJECT_DIR/packages/cli/lazyai-cli"

# Use an isolated temp project so we don't depend on the repo's own ledger format.
TMP_DIR=$(mktemp -d)
TMP_HOME=$(mktemp -d)
export HOME="$TMP_HOME"
export LAZYAI_HOME="$TMP_HOME/.lazyai"
LEDGER_FILE="$TMP_DIR/.specify/ledger.jsonl"

cleanup() {
    rm -rf "$TMP_DIR" "$TMP_HOME"
}
trap cleanup EXIT

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 Integration Test: Ledger Integrity"
echo "═══════════════════════════════════════════════════════════════"

cd "$TMP_DIR"

# Test 1: Initialize ledger
echo ""
echo "Test 1: Ledger initialization"
INIT_OUTPUT=$($CLI ledger init 2>&1)
if echo "$INIT_OUTPUT" | grep -q "Ledger initialized"; then
    echo "✅ Ledger initialized"
else
    echo "❌ FAIL: Ledger initialization failed"
    echo "$INIT_OUTPUT"
    exit 1
fi

# Test 2: Verify ledger file exists
echo ""
echo "Test 2: Ledger file exists"
if [ -f "$LEDGER_FILE" ]; then
    echo "✅ Ledger file exists"
else
    echo "❌ FAIL: Ledger file not found at $LEDGER_FILE"
    exit 1
fi

# Test 3: Verify ledger has entries
echo ""
echo "Test 3: Ledger has entries"
ENTRY_COUNT=$(wc -l < "$LEDGER_FILE" | tr -d ' ')
if [ "$ENTRY_COUNT" -gt 0 ]; then
    echo "✅ Ledger has $ENTRY_COUNT entries"
else
    echo "❌ FAIL: Ledger is empty"
    exit 1
fi

# Test 4: Verify genesis entry exists
echo ""
echo "Test 4: Genesis entry exists"
if head -1 "$LEDGER_FILE" | grep -q "genesis"; then
    echo "✅ Genesis entry found"
else
    echo "❌ FAIL: Genesis entry not found"
    exit 1
fi

# Test 5: Append and verify hash chain integrity
echo ""
echo "Test 5: Append entry and verify hash chain integrity"
APPEND_OUTPUT=$($CLI ledger append test_event "integration test data" 2>&1)
if echo "$APPEND_OUTPUT" | grep -q "Entry appended"; then
    echo "✅ Entry appended"
else
    echo "❌ FAIL: Ledger append failed"
    echo "$APPEND_OUTPUT"
    exit 1
fi

VERIFY_OUTPUT=$($CLI ledger verify 2>&1)
echo "$VERIFY_OUTPUT"
if echo "$VERIFY_OUTPUT" | grep -q "intact"; then
    echo "✅ Hash chain is intact"
else
    echo "❌ FAIL: Hash chain is broken"
    exit 1
fi

# Test 6: Show ledger
echo ""
echo "Test 6: Show ledger"
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
