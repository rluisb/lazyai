#!/usr/bin/env bash
# tests/scripts/smoke-test.sh — Smoke tests for lazyai CLI

set -uo pipefail

PASS=0
FAIL=0

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local msg="${3:-}"
    if [[ "$haystack" == *"$needle"* ]]; then
        PASS=$((PASS + 1))
        echo "  ✅ PASS: $msg"
    else
        FAIL=$((FAIL + 1))
        echo "  ❌ FAIL: $msg"
    fi
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI_DIR="$REPO_ROOT/packages/cli"

echo "═══════════════════════════════════════════════════════════════"
echo "  LazyAI Smoke Tests"
echo "  $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# ─── Test: Build ───

echo "🧪 Build"
if (cd "$CLI_DIR" && go build ./cmd/... 2>/dev/null); then
    PASS=$((PASS + 1))
    echo "  ✅ PASS: go build succeeds"
else
    FAIL=$((FAIL + 1))
    echo "  ❌ FAIL: go build fails"
fi

# ─── Test: Doctor Command ───

echo ""
echo "🧪 Doctor Command"
output=$(cd "$REPO_ROOT" && go run ./packages/cli/cmd/lazyai-cli doctor --help 2>&1) || true
assert_contains "$output" "doctor" "doctor command is available"
assert_contains "$output" "--json" "doctor exposes json output"
assert_contains "$output" "--fix" "doctor exposes fix option"

# ─── Test: Session Command ───

echo ""
echo "🧪 Session Command"
output=$(cd "$REPO_ROOT" && go run ./packages/cli/cmd/lazyai-cli session start "smoke-test" 2>&1) || true
assert_contains "$output" "Session started" "session start works"

# Extract session ID and test other commands
session_id=$(echo "$output" | grep "Session started:" | sed 's/.*Session started: //' | tr -d ' ')
if [[ -n "$session_id" ]]; then
    list_output=$(cd "$REPO_ROOT" && go run ./packages/cli/cmd/lazyai-cli session list 2>&1) || true
    assert_contains "$list_output" "$session_id" "session list shows session"
    
    end_output=$(cd "$REPO_ROOT" && go run ./packages/cli/cmd/lazyai-cli session end "$session_id" 2>&1) || true
    assert_contains "$end_output" "Session ended" "session end works"
else
    FAIL=$((FAIL + 2))
    echo "  ❌ FAIL: could not extract session ID"
fi

# ─── Test: Ledger Command ───

echo ""
echo "🧪 Ledger Command"

# Build binary first
bin_path="$REPO_ROOT/packages/cli/lazyai-cli-test"
(cd "$CLI_DIR" && go build -o "$bin_path" ./cmd/lazyai-cli) || true

tmp_dir=$(mktemp -d)

# Test ledger init
init_output=$(cd "$tmp_dir" && "$bin_path" ledger init 2>&1) || true
assert_contains "$init_output" "Ledger initialized" "ledger init works"

# Test ledger append
append_output=$(cd "$tmp_dir" && "$bin_path" ledger append test "entry" 2>&1) || true
assert_contains "$append_output" "Entry appended" "ledger append works"

# Test ledger verify
verify_output=$(cd "$tmp_dir" && "$bin_path" ledger verify 2>&1) || true
assert_contains "$verify_output" "Chain intact" "ledger verify works"

rm -rf "$tmp_dir"
rm -f "$bin_path"

# ─── Test: Validate Command ───

echo ""
echo "🧪 Validate Command"
validate_output=$(cd "$REPO_ROOT" && go run ./packages/cli/cmd/lazyai-cli validate agents 2>&1) || true
assert_contains "$validate_output" "Agent Validation Results" "validate agents works"

# ─── Summary ───

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  Results"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  ✅ $PASS passed"
echo "  ❌ $FAIL failed"
echo ""

if [[ $FAIL -eq 0 ]]; then
    echo "  🎉 All tests passed!"
    exit 0
else
    echo "  ⚠️  $FAIL test(s) failed"
    exit 1
fi
