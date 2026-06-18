#!/usr/bin/env bash
# wait-barrier.sh — Non-blocking barrier wait with timeout and status polling
# Usage: ./wait-barrier.sh <barrier-id> [--timeout N] [--poll-interval N] [--json]
#
# This script provides a non-blocking way to wait for parallel tasks to complete.
# It polls the barrier status and returns when all tasks have arrived or timeout is reached.
#
# Exit codes:
#   0 — All tasks arrived (barrier resolved)
#   1 — Timeout reached
#   2 — Barrier not found

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BARRIER_SCRIPT="$SCRIPT_DIR/../skills/battle-bus/scripts/task-barrier.sh"
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"

BARRIER_ID="${1:-}"
TIMEOUT=300
POLL_INTERVAL=2
JSON_OUTPUT=false

shift 2>/dev/null || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --timeout) TIMEOUT="$2"; shift 2 ;;
        --poll-interval) POLL_INTERVAL="$2"; shift 2 ;;
        --json) JSON_OUTPUT=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$BARRIER_ID" ]]; then
    echo "Usage: $0 <barrier-id> [--timeout N] [--poll-interval N] [--json]"
    exit 1
fi

# Check if barrier exists
check_barrier() {
    if [[ ! -f "$DB_PATH" ]]; then
        return 1
    fi

    sqlite3 "$DB_PATH" "SELECT status, arrived_count, expected_count FROM barriers WHERE id='$BARRIER_ID' LIMIT 1;" 2>/dev/null || echo ""
}

# Get barrier status
get_barrier_status() {
    local result
    result=$(check_barrier)

    if [[ -z "$result" ]]; then
        echo "not_found"
        return
    fi

    IFS='|' read -r status arrived expected <<< "$result"
    echo "$status|$arrived|$expected"
}

# Main wait loop
echo "⏳ Waiting for barrier: $BARRIER_ID"
echo "   Timeout: ${TIMEOUT}s, Poll interval: ${POLL_INTERVAL}s"
echo ""

ELAPSED=0
while [[ $ELAPSED -lt $TIMEOUT ]]; do
    STATUS=$(get_barrier_status)

    if [[ "$STATUS" == "not_found" ]]; then
        if [[ "$JSON_OUTPUT" == true ]]; then
            echo '{"status":"not_found","barrier":"'"$BARRIER_ID"'"}'
        else
            echo "❌ Barrier not found: $BARRIER_ID"
        fi
        exit 2
    fi

    IFS='|' read -r state arrived expected <<< "$STATUS"

    if [[ "$state" == "resolved" ]]; then
        if [[ "$JSON_OUTPUT" == true ]]; then
            echo '{"status":"resolved","barrier":"'"$BARRIER_ID"'","arrived":'"$arrived"',"expected":'"$expected"',"elapsed":'"$ELAPSED"'}'
        else
            echo "✅ Barrier resolved: $arrived/$expected tasks arrived (${ELAPSED}s)"
        fi
        exit 0
    fi

    # Show progress
    if [[ "$JSON_OUTPUT" != true ]]; then
        echo "   ⏸  $arrived/$expected arrived (${ELAPSED}s elapsed)"
    fi

    sleep "$POLL_INTERVAL"
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

# Timeout reached
if [[ "$JSON_OUTPUT" == true ]]; then
    echo '{"status":"timeout","barrier":"'"$BARRIER_ID"'","elapsed":'"$ELAPSED"'}'
else
    echo "⏰ Timeout reached after ${ELAPSED}s"

    # Show current status
    STATUS=$(get_barrier_status)
    IFS='|' read -r state arrived expected <<< "$STATUS"
    echo "   Current: $arrived/$expected arrived"
    echo ""
    echo "   Pending tasks:"
    sqlite3 "$DB_PATH" "SELECT agent, task FROM parallel_tasks WHERE session_id=(SELECT session_id FROM barriers WHERE id='$BARRIER_ID') AND status='pending';" 2>/dev/null | while IFS='|' read -r agent task; do
        echo "   - $agent: $task"
    done
fi

exit 1
