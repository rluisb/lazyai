#!/usr/bin/env bash
# scripts/preflight.sh — Preflight readiness checks
# Runs health, branch, DB, ledger, and workflow checks deterministically.
# Usage: ./scripts/preflight.sh [--json] [--quick]

set -euo pipefail
IFS=$'\n\t'

WORKSPACE="${OPENCODE_WORKSPACE:-.}"
OUTPUT_JSON=false
QUICK=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --json) OUTPUT_JSON=true; shift ;;
        --quick) QUICK=true; shift ;;
        *) shift ;;
    esac
done

PASS=0
FAIL=0
WARN=0
CHECKS=""

check() {
    local name="$1"
    local status="$2"
    local detail="${3:-}"
    case "$status" in
        PASS) PASS=$((PASS + 1)); emoji="✅" ;;
        FAIL) FAIL=$((FAIL + 1)); emoji="❌" ;;
        WARN) WARN=$((WARN + 1)); emoji="⚠️" ;;
    esac
    CHECKS="${CHECKS}{\"name\":\"$name\",\"status\":\"$status\",\"detail\":\"$detail\"},"
    if [[ "$OUTPUT_JSON" == "false" ]]; then
        printf "  %s %-40s %s\n" "$emoji" "$name" "$detail"
    fi
}

if [[ "$OUTPUT_JSON" == "false" ]]; then
    echo "🛫 Preflight Check"
    echo "   $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo ""
fi

# --- HEALTH CHECK ---
if [[ -x "$WORKSPACE/scripts/health-check.sh" ]]; then
    HEALTH_JSON=$(bash "$WORKSPACE/scripts/health-check.sh" --json 2>/dev/null || true)
    if [[ -n "$HEALTH_JSON" ]]; then
        HEALTH_PASS=$(echo "$HEALTH_JSON" | jq -r '.pass // 0' 2>/dev/null || echo "0")
        HEALTH_FAIL=$(echo "$HEALTH_JSON" | jq -r '.fail // 0' 2>/dev/null || echo "0")
        if [[ "$HEALTH_FAIL" -eq 0 ]]; then
            check "Health check" PASS "$HEALTH_PASS passed"
        else
            check "Health check" WARN "$HEALTH_FAIL failed"
        fi
    else
        check "Health check" WARN "No JSON output"
    fi
else
    check "Health check" FAIL "scripts/health-check.sh not found or not executable"
fi

# --- BRANCH CHECK ---
if command -v git &>/dev/null; then
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    if [[ "$CURRENT_BRANCH" =~ ^(main|master|production|release/.*)$ ]]; then
        check "Branch protection" FAIL "On protected branch: $CURRENT_BRANCH"
    else
        check "Branch protection" PASS "$CURRENT_BRANCH"
    fi
else
    check "Branch protection" WARN "git not available"
fi

# --- SESSION DB CHECK ---
DB_PATH="$WORKSPACE/.specify/session.db"
if [[ -f "$DB_PATH" ]]; then
    if command -v sqlite3 &>/dev/null; then
        TABLES=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sqlite_master WHERE type='table';" 2>/dev/null || echo "0")
        check "Session DB" PASS "$TABLES tables"
    else
        check "Session DB" WARN "sqlite3 not found"
    fi
else
    check "Session DB" WARN "No session.db found"
fi

# --- LEDGER CHECK ---
LEDGER_PATH="$WORKSPACE/.specify/ledger.jsonl"
LEDGER_SCRIPT="$WORKSPACE/skills/truth-chain/scripts/ledger.sh"
if [[ -f "$LEDGER_PATH" ]]; then
    if [[ -x "$LEDGER_SCRIPT" ]]; then
        RESULT=$(bash "$LEDGER_SCRIPT" verify 2>&1 || true)
        if echo "$RESULT" | grep -q "PASS"; then
            ENTRIES=$(wc -l < "$LEDGER_PATH" | tr -d ' ')
            check "Ledger" PASS "$ENTRIES entries, chain intact"
        else
            check "Ledger" FAIL "Chain broken"
        fi
    else
        check "Ledger" FAIL "ledger.sh not executable"
    fi
else
    check "Ledger" WARN "No ledger found"
fi

# --- WORKFLOW CONFIGS ---
if [[ -d "$WORKSPACE/.opencode/workflows" ]]; then
    WF_COUNT=$(ls "$WORKSPACE/.opencode/workflows"/*.yaml 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$WF_COUNT" -gt 0 ]]; then
        check "Workflow configs" PASS "$WF_COUNT YAML configs"
    else
        check "Workflow configs" WARN "No YAML configs found"
    fi
else
    check "Workflow configs" WARN "No .opencode/workflows/ directory"
fi

# --- SUMMARY ---
TOTAL=$((PASS + FAIL + WARN))

if [[ "$OUTPUT_JSON" == "true" ]]; then
    echo "{\"timestamp\":\"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\",\"pass\":$PASS,\"fail\":$FAIL,\"warn\":$WARN,\"checks\":[${CHECKS%,}]}"
else
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "  ✅ %d passed  ❌ %d failed  ⚠️  %d warnings  (%d total)\n" "$PASS" "$FAIL" "$WARN" "$TOTAL"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

if [[ $FAIL -gt 0 ]]; then
    exit 1
else
    exit 0
fi
