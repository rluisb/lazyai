#!/usr/bin/env bash
# health-check.sh — System health check for Fortnite multi-agent system
# Usage: ./health-check.sh [--json] [--quick]

set -euo pipefail

WORKSPACE="${OPENCODE_WORKSPACE:-.}"
DB_PATH="$WORKSPACE/.specify/session.db"
LEDGER_PATH="$WORKSPACE/.specify/ledger.jsonl"
LEDGER_SCRIPT="$WORKSPACE/skills/truth-chain/scripts/ledger.sh"
SESSION_SCRIPT="$WORKSPACE/scripts/session-db.sh"
WORKFLOW_SCRIPT="$WORKSPACE/scripts/workflow-run.sh"
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
    echo "🏥 Fortnite System Health Check"
    echo "   $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo ""
fi

# --- DEPENDENCY CHECKS ---
if command -v sqlite3 &>/dev/null; then
    SQLITE_VERSION=$(sqlite3 --version 2>/dev/null | awk '{print $1}')
    check "Dependency: sqlite3" PASS "$SQLITE_VERSION"
else
    check "Dependency: sqlite3" FAIL "not found"
fi

if command -v git &>/dev/null; then
    GIT_VERSION=$(git --version 2>/dev/null | sed 's/git version //')
    check "Dependency: git" PASS "$GIT_VERSION"
else
    check "Dependency: git" FAIL "not found"
fi

if command -v jq &>/dev/null; then
    JQ_VERSION=$(jq --version 2>/dev/null | sed 's/jq-//')
    check "Dependency: jq" PASS "$JQ_VERSION"
else
    check "Dependency: jq" FAIL "not found"
fi

BASH_MAJOR=$(bash --version 2>/dev/null | head -1 | sed -E 's/.*version ([0-9]+).*/\1/')
if [[ -n "$BASH_MAJOR" ]] && [[ "$BASH_MAJOR" -ge 4 ]]; then
    check "Dependency: bash" PASS "version $BASH_MAJOR (≥4)"
else
    check "Dependency: bash" FAIL "version < 4 or not found"
fi

if command -v bats &>/dev/null; then
    check "Dependency: bats" PASS "available"
elif command -v shellspec &>/dev/null; then
    check "Dependency: shellspec" PASS "available"
else
    check "Dependency: test runner" WARN "bats/shellspec not found — using POSIX smoke tests"
fi

if command -v opencode &>/dev/null || command -v opencode-cli &>/dev/null; then
    check "Dependency: opencode CLI" PASS "available"
else
    check "Dependency: opencode CLI" WARN "not found in PATH"
fi

# --- LEDGER INTEGRITY ---
if [[ -f "$LEDGER_PATH" ]]; then
    if [[ -x "$LEDGER_SCRIPT" ]]; then
        RESULT=$("$LEDGER_SCRIPT" verify 2>&1)
        if echo "$RESULT" | grep -q "PASS"; then
            ENTRIES=$(wc -l < "$LEDGER_PATH" | tr -d ' ')
            check "Ledger integrity" PASS "$ENTRIES entries, chain intact"
        else
            check "Ledger integrity" FAIL "Chain broken — run ledger.sh verify"
        fi
    else
        check "Ledger integrity" FAIL "ledger.sh not executable"
    fi
else
    check "Ledger integrity" WARN "No ledger found — run ledger.sh init"
fi

# --- SQLITE ACCESSIBILITY ---
if [[ -f "$DB_PATH" ]]; then
    if command -v sqlite3 &>/dev/null; then
        TABLES=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sqlite_master WHERE type='table';" 2>/dev/null || echo "0")
        SESSIONS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sessions;" 2>/dev/null || echo "0")
        DISPATCHES=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM dispatches;" 2>/dev/null || echo "0")
        WORKFLOWS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflows;" 2>/dev/null || echo "0")
        RUNS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflow_runs;" 2>/dev/null || echo "0")
        DB_SIZE=$(du -sh "$DB_PATH" 2>/dev/null | awk '{print $1}' || echo "?")
        check "SQLite database" PASS "$TABLES tables, $SESSIONS sessions, $DISPATCHES dispatches, $WORKFLOWS workflows, $RUNS runs ($DB_SIZE)"
    else
        check "SQLite database" FAIL "sqlite3 not found"
    fi
else
    check "SQLite database" WARN "No session.db found"
fi

# --- AGENT MODELS ---
if [[ "$QUICK" == "false" ]]; then
    MODELS=("ollama-cloud/deepseek-v4-pro" "ollama-cloud/kimi-k2.6:cloud" "openai/gpt-5.5" "openai/gpt-5.5-fast" "ollama-cloud/deepseek-v4-flash" "ollama-cloud/nemotron-3-super" "ollama-cloud/glm-5.1" "ollama-cloud/gemma4")
    AVAILABLE=0
    UNAVAILABLE=0
    for model in "${MODELS[@]}"; do
        if curl -s --max-time 3 "https://api.openai.com/v1/models" >/dev/null 2>&1; then
            AVAILABLE=$((AVAILABLE + 1))
        else
            UNAVAILABLE=$((UNAVAILABLE + 1))
        fi
    done
    if [[ $UNAVAILABLE -eq 0 ]]; then
        check "Agent models" PASS "$AVAILABLE models available"
    elif [[ $AVAILABLE -gt 0 ]]; then
        check "Agent models" WARN "$AVAILABLE available, $UNAVAILABLE unreachable"
    else
        check "Agent models" FAIL "No models reachable"
    fi
else
    check "Agent models" PASS "Skipped (quick mode)"
fi

# --- SCRIPT SYNTAX ---
SCRIPTS=(
    "scripts/session-db.sh"
    "scripts/workflow-run.sh"
    "skills/truth-chain/scripts/ledger.sh"
    "skills/battle-bus/scripts/agent-msg.sh"
    "skills/battle-bus/scripts/task-barrier.sh"
    "skills/battle-bus/scripts/task-lock.sh"
)
SYNTAX_OK=0
SYNTAX_ERR=0
for script in "${SCRIPTS[@]}"; do
    if [[ -f "$WORKSPACE/$script" ]]; then
        if bash -n "$WORKSPACE/$script" 2>/dev/null; then
            SYNTAX_OK=$((SYNTAX_OK + 1))
        else
            SYNTAX_ERR=$((SYNTAX_ERR + 1))
        fi
    fi
done
if [[ $SYNTAX_ERR -eq 0 ]]; then
    check "Script syntax" PASS "$SYNTAX_OK scripts valid"
else
    check "Script syntax" FAIL "$SYNTAX_ERR scripts have syntax errors"
fi

# --- DISK SPACE ---
DISK_USAGE=$(df -h "$WORKSPACE" 2>/dev/null | tail -1 | awk '{print $5}' | tr -d '%' || echo "?")
if [[ "$DISK_USAGE" != "?" ]]; then
    if [[ $DISK_USAGE -lt 80 ]]; then
        check "Disk space" PASS "${DISK_USAGE}% used"
    elif [[ $DISK_USAGE -lt 95 ]]; then
        check "Disk space" WARN "${DISK_USAGE}% used — approaching limit"
    else
        check "Disk space" FAIL "${DISK_USAGE}% used — critical"
    fi
else
    check "Disk space" WARN "Could not determine"
fi

# --- WORKFLOW CONFIGS ---
if [[ -d "$WORKSPACE/.opencode/workflows" ]]; then
    WF_COUNT=$(ls "$WORKSPACE/.opencode/workflows"/*.yaml 2>/dev/null | wc -l | tr -d ' ')
    if [[ $WF_COUNT -gt 0 ]]; then
        check "Workflow configs" PASS "$WF_COUNT YAML configs"
    else
        check "Workflow configs" WARN "No YAML configs found"
    fi
else
    check "Workflow configs" WARN "No .opencode/workflows/ directory"
fi

# --- SCHEDULED JOBS ---
if command -v launchctl &>/dev/null; then
    JOB_COUNT=$(launchctl list 2>/dev/null | grep -c "opencode" || echo "0")
    if [[ $JOB_COUNT -gt 0 ]]; then
        check "Scheduled jobs" PASS "$JOB_COUNT jobs active"
    else
        check "Scheduled jobs" WARN "No scheduled jobs found"
    fi
else
    check "Scheduled jobs" WARN "launchctl not available"
fi

# --- PROVIDER HEALTH ---
check_provider() {
    local provider="$1"
    local url="$2"
    local headers="${3:-}"
    local model="${4:-}"
    local tmpfile exit_code http_code latency_sec latency_ms status detail error_msg

    tmpfile=$(mktemp 2>/dev/null || echo "/tmp/health-check-$$.tmp")

    # Build curl command
    local curl_cmd=("curl" "-s" "--max-time" "6" "-o" "/dev/null" "-w" "%{http_code},%{time_total}")
    if [[ -n "$headers" ]]; then
        curl_cmd+=("-H" "$headers")
    fi
    curl_cmd+=("$url")

    # Run curl and capture output
    local curl_output
    curl_output=$("${curl_cmd[@]}" 2>"$tmpfile") || exit_code=$?
    exit_code=${exit_code:-0}

    http_code=$(echo "$curl_output" | cut -d, -f1)
    latency_sec=$(echo "$curl_output" | cut -d, -f2)

    # Convert to milliseconds (portable)
    if [[ "$latency_sec" =~ ^[0-9]+\.[0-9]+$ ]]; then
        latency_ms=$(echo "$latency_sec" | awk '{printf "%.0f", $1 * 1000}')
    else
        latency_ms=""
    fi

    # Read stderr
    error_msg=$(cat "$tmpfile" 2>/dev/null | head -c 200)
    rm -f "$tmpfile"

    if [[ "$exit_code" -ne 0 ]] || [[ "$http_code" == "000" ]]; then
        status="down"
        detail="No response"
    elif [[ "$http_code" -ge 200 && "$http_code" -lt 300 ]]; then
        if [[ -n "$latency_ms" && "$latency_ms" -gt 5000 ]]; then
            status="degraded"
            detail="Slow response (${latency_ms}ms)"
            error_msg=""
        else
            status="ok"
            detail="Responding (${latency_ms}ms)"
            error_msg=""
        fi
    else
        status="down"
        detail="HTTP $http_code"
    fi

    # Record to session DB (best effort, don't fail health check)
    if [[ -x "$SESSION_SCRIPT" ]]; then
        "$SESSION_SCRIPT" record-provider-health "$provider" "$model" "$status" "$latency_ms" "$error_msg" >/dev/null 2>&1 || true
    fi

    case "$status" in
        ok) check "Provider: $provider" PASS "$detail" ;;
        degraded) check "Provider: $provider" WARN "$detail" ;;
        down) check "Provider: $provider" WARN "$detail${error_msg:+ — $error_msg}" ;;
    esac
}

if command -v curl &>/dev/null; then
    # Check ollama-cloud (local Ollama)
    check_provider "ollama-cloud" "http://localhost:11434/api/tags" "" ""

    # Check openai (API key + lightweight ping)
    if [[ -n "${OPENAI_API_KEY:-}" ]]; then
        check_provider "openai" "https://api.openai.com/v1/models" "Authorization: Bearer $OPENAI_API_KEY" ""
    else
        if [[ -x "$SESSION_SCRIPT" ]]; then
            "$SESSION_SCRIPT" record-provider-health "openai" "" "down" "" "API key not set" >/dev/null 2>&1 || true
        fi
        check "Provider: openai" WARN "API key not set"
    fi
else
    check "Provider health" WARN "curl not found — cannot check providers"
fi

# --- EVAL METRICS ---
EVAL_DIR="$WORKSPACE/.specify/evals/results"
if [[ -d "$EVAL_DIR" ]]; then
    EVAL_COUNT=$(ls "$EVAL_DIR"/*.json 2>/dev/null | wc -l | tr -d ' ')
    if [[ $EVAL_COUNT -gt 0 ]]; then
        LATEST=$(ls -t "$EVAL_DIR"/*.json 2>/dev/null | head -1)
        if [[ -f "$LATEST" ]]; then
            AVG_SCORE=$(python3 -c "import json; d=json.load(open('$LATEST')); scores=[s for s in d.get('scores',{}).values()]; print(f'{sum(scores)/len(scores):.2f}')" 2>/dev/null || echo "?")
            check "Eval metrics" PASS "$EVAL_COUNT runs, latest avg: $AVG_SCORE"
        else
            check "Eval metrics" WARN "$EVAL_COUNT eval runs found"
        fi
    else
        check "Eval metrics" WARN "No eval results yet — run storm-eye"
    fi
else
    check "Eval metrics" WARN "No eval directory — run storm-eye define"
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
