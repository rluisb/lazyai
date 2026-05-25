#!/usr/bin/env bash
# wave-summary.sh — Aggregate results from all parallel tasks in a wave
# Usage: ./wave-summary.sh <wave-id> [--json] [--output file]
#
# Collects results from all parallel tasks in a wave and produces a summary:
# - Task completion status
# - Output file locations
# - Errors (if any)
# - Overall wave status

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"

WAVE_ID="${1:-}"
JSON_OUTPUT=false
OUTPUT_FILE=""

shift 2>/dev/null || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --json) JSON_OUTPUT=true; shift ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$WAVE_ID" ]]; then
    echo "Usage: $0 <wave-id> [--json] [--output file]"
    exit 1
fi

if [[ ! -f "$DB_PATH" ]]; then
    echo "❌ Session database not found: $DB_PATH"
    exit 1
fi

# Get wave tasks
TASKS=$(sqlite3 "$DB_PATH" "SELECT id, agent, task, status, result, output_path, started_at, completed_at FROM parallel_tasks WHERE wave_id='$WAVE_ID' ORDER BY id;" 2>/dev/null)

if [[ -z "$TASKS" ]]; then
    echo "❌ No tasks found for wave: $WAVE_ID"
    exit 1
fi

# Count statuses
TOTAL=0
COMPLETED=0
FAILED=0
PENDING=0
RUNNING=0

while IFS='|' read -r id agent task status result output_path started completed; do
    TOTAL=$((TOTAL + 1))
    case "$status" in
        completed) COMPLETED=$((COMPLETED + 1)) ;;
        failed) FAILED=$((FAILED + 1)) ;;
        pending) PENDING=$((PENDING + 1)) ;;
        running) RUNNING=$((RUNNING + 1)) ;;
    esac
done <<< "$TASKS"

# Determine overall status
if [[ $PENDING -gt 0 || $RUNNING -gt 0 ]]; then
    OVERALL="in_progress"
elif [[ $FAILED -gt 0 ]]; then
    OVERALL="partial_failure"
else
    OVERALL="complete"
fi

# Generate summary
generate_summary() {
    echo "## Wave Summary: $WAVE_ID"
    echo ""
    echo "**Status:** $OVERALL"
    echo "**Total Tasks:** $TOTAL"
    echo "**Completed:** $COMPLETED"
    echo "**Failed:** $FAILED"
    echo "**Pending:** $PENDING"
    echo "**Running:** $RUNNING"
    echo ""

    echo "### Task Details"
    echo ""
    echo "| # | Agent | Task | Status | Output |"
    echo "|---|-------|------|--------|--------|"

    while IFS='|' read -r id agent task status result output_path started completed; do
        status_icon=""
        case "$status" in
            completed) status_icon="✅" ;;
            failed) status_icon="❌" ;;
            pending) status_icon="⏸" ;;
            running) status_icon="🔄" ;;
        esac

        output_display="${output_path:--}"
        echo "| $id | $agent | $task | $status_icon $status | $output_display |"
    done <<< "$TASKS"

    echo ""

    # Show errors if any
    if [[ $FAILED -gt 0 ]]; then
        echo "### Errors"
        echo ""
        while IFS='|' read -r id agent task status result output_path started completed; do
            if [[ "$status" == "failed" && -n "$result" ]]; then
                echo "**Task $id ($agent):** $result"
                echo ""
            fi
        done <<< "$TASKS"
    fi

    # Next actions
    echo "### Next Actions"
    echo ""
    if [[ "$OVERALL" == "complete" ]]; then
        echo "✅ All tasks completed. Proceed to next wave or merge results."
    elif [[ "$OVERALL" == "partial_failure" ]]; then
        echo "⚠️  Some tasks failed. Review errors above and decide:"
        echo "- Retry failed tasks"
        echo "- Proceed with partial results"
        echo "- Abort wave"
    else
        echo "⏳ Wave still in progress. Wait for completion:"
        echo "\`\`\`bash"
        echo "./scripts/wait-barrier.sh barrier-$WAVE_ID --timeout 300"
        echo "\`\`\`"
    fi
}

generate_json() {
    echo "{"
    echo "  \"wave_id\": \"$WAVE_ID\","
    echo "  \"status\": \"$OVERALL\","
    echo "  \"total\": $TOTAL,"
    echo "  \"completed\": $COMPLETED,"
    echo "  \"failed\": $FAILED,"
    echo "  \"pending\": $PENDING,"
    echo "  \"running\": $RUNNING,"
    echo "  \"tasks\": ["

    FIRST=true
    while IFS='|' read -r id agent task status result output_path started completed; do
        if [[ "$FIRST" != true ]]; then
            echo ","
        fi
        FIRST=false

        echo "    {"
        echo "      \"id\": $id,"
        echo "      \"agent\": \"$agent\","
        echo "      \"task\": \"${task//\"/\\\"}\","
        echo "      \"status\": \"$status\","
        echo "      \"result\": \"${result//\"/\\\"}\","
        echo "      \"output_path\": \"${output_path:--}\","
        echo "      \"started_at\": \"$started\","
        echo "      \"completed_at\": \"$completed\""
        echo -n "    }"
    done <<< "$TASKS"

    echo ""
    echo "  ]"
    echo "}"
}

# Output
if [[ "$JSON_OUTPUT" == true ]]; then
    SUMMARY=$(generate_json)
else
    SUMMARY=$(generate_summary)
fi

if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$SUMMARY" > "$OUTPUT_FILE"
    echo "✅ Summary saved to: $OUTPUT_FILE"
else
    echo "$SUMMARY"
fi
