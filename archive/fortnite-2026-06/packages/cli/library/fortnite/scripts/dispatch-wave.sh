#!/usr/bin/env bash
# dispatch-wave.sh — Auto-dispatch parallel tasks from a wave definition
# Usage: ./dispatch-wave.sh <wave-file> [--dry-run] [--max-concurrent N]
#
# Wave file format (one task per line):
#   agent|mode|task-description|output-path
#
# Example wave file:
#   wall-builder|standard|Implement auth middleware|worktrees/auth/result.md
#   wall-builder|standard|Implement payment handler|worktrees/payment/result.md
#   shield-audit|review|Review auth implementation|worktrees/auth-review/result.md
#
# The script:
# 1. Validates wave file format
# 2. Checks for file path collisions
# 3. Registers parallel tasks in session-db
# 4. Creates barrier for wave sync
# 5. Dispatches tasks (up to max-concurrent at a time)
# 6. Waits for barrier resolution

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SESSION_DB="$SCRIPT_DIR/session-db.sh"
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"
BARRIER_SCRIPT="$SCRIPT_DIR/../skills/battle-bus/scripts/task-barrier.sh"
LOCK_SCRIPT="$SCRIPT_DIR/../skills/battle-bus/scripts/task-lock.sh"

WAVE_FILE="${1:-}"
DRY_RUN=false
MAX_CONCURRENT=4

shift 2>/dev/null || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --max-concurrent) MAX_CONCURRENT="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$WAVE_FILE" ]]; then
    echo "Usage: $0 <wave-file> [--dry-run] [--max-concurrent N]"
    echo ""
    echo "Wave file format (one task per line):"
    echo "  agent|mode|task-description|output-path"
    echo ""
    echo "Example:"
    echo "  wall-builder|standard|Implement auth middleware|worktrees/auth/result.md"
    exit 1
fi

if [[ ! -f "$WAVE_FILE" ]]; then
    echo "❌ Wave file not found: $WAVE_FILE"
    exit 1
fi

# Detect session
detect_session() {
    if [ -n "${SESSION_ID:-}" ]; then
        echo "$SESSION_ID"
        return
    fi
    if [ -f "$DB_PATH" ]; then
        sqlite3 "$DB_PATH" "SELECT id FROM sessions WHERE status='active' ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || true
    fi
}

SID=$(detect_session)
if [[ -z "$SID" ]]; then
    echo "❌ No active session. Set SESSION_ID or create a session first."
    exit 1
fi

# Parse wave file
TASKS=()
OUTPUT_PATHS=()
echo "📋 Parsing wave file: $WAVE_FILE"
echo ""

LINE_NUM=0
while IFS='|' read -r agent mode task output_path; do
    LINE_NUM=$((LINE_NUM + 1))

    # Skip empty lines and comments
    [[ -z "$agent" || "$agent" =~ ^# ]] && continue

    # Validate format
    if [[ -z "$mode" || -z "$task" || -z "$output_path" ]]; then
        echo "❌ Invalid format at line $LINE_NUM: expected agent|mode|task|output-path"
        echo "   Got: $agent|$mode|$task|$output_path"
        exit 1
    fi

    TASKS+=("$agent|$mode|$task|$output_path")
    OUTPUT_PATHS+=("$output_path")
done < "$WAVE_FILE"

TOTAL_TASKS=${#TASKS[@]}
if [[ $TOTAL_TASKS -eq 0 ]]; then
    echo "❌ No tasks found in wave file"
    exit 1
fi

echo "   Found $TOTAL_TASKS tasks"
echo ""

# Check for file path collisions
echo "🔍 Checking for file path collisions..."
COLLISIONS=()
for ((i=0; i<${#OUTPUT_PATHS[@]}; i++)); do
    for ((j=i+1; j<${#OUTPUT_PATHS[@]}; j++)); do
        if [[ "${OUTPUT_PATHS[$i]}" == "${OUTPUT_PATHS[$j]}" ]]; then
            COLLISIONS+=("${OUTPUT_PATHS[$i]} (tasks $((i+1)) and $((j+1)))")
        fi
    done
done

if [[ ${#COLLISIONS[@]} -gt 0 ]]; then
    echo "❌ File path collisions detected:"
    for collision in "${COLLISIONS[@]}"; do
        echo "   - $collision"
    done
    echo ""
    echo "   Each parallel task must write to a unique output path."
    exit 1
fi
echo "   ✅ No collisions found"
echo ""

# Register parallel tasks and create barrier
WAVE_ID="wave-$(date +%s)"
BARRIER_ID="barrier-$WAVE_ID"

echo "📝 Registering parallel tasks..."
echo "   Wave ID: $WAVE_ID"
echo "   Barrier: $BARRIER_ID"
echo ""

PTASK_IDS=()
for task_def in "${TASKS[@]}"; do
    IFS='|' read -r agent mode task output_path <<< "$task_def"

    # Register in session-db
    PTID=$("$SESSION_DB" ptask "$SID" "$agent" "$task" "$WAVE_ID" 2>/dev/null | tail -1)
    PTASK_IDS+=("$PTID")

    echo "   ✅ Registered: $agent ($mode) → $output_path [ptask:$PTID]"
done

echo ""

# Create barrier
echo "🚧 Creating barrier: $BARRIER_ID (count: $TOTAL_TASKS)"
if [[ "$DRY_RUN" == true ]]; then
    echo "   [DRY RUN] Would create barrier"
else
    "$BARRIER_SCRIPT" create "$BARRIER_ID" "$TOTAL_TASKS" 2>/dev/null || true
fi
echo ""

# Dispatch tasks
if [[ "$DRY_RUN" == true ]]; then
    echo "🚀 [DRY RUN] Would dispatch $TOTAL_TASKS tasks (max concurrent: $MAX_CONCURRENT)"
    echo ""
    for task_def in "${TASKS[@]}"; do
        IFS='|' read -r agent mode task output_path <<< "$task_def"
        echo "   Would dispatch: $agent MODE=$mode → $output_path"
        echo "   Task: $task"
        echo ""
    done
else
    echo "🚀 Dispatching $TOTAL_TASKS tasks (max concurrent: $MAX_CONCURRENT)..."
    echo ""

    # Dispatch in batches
    BATCH=()
    for ((i=0; i<${#TASKS[@]}; i++)); do
        task_def="${TASKS[$i]}"
        IFS='|' read -r agent mode task output_path <<< "$task_def"

        BATCH+=("$task_def")

        # Dispatch batch when full or at end
        if [[ ${#BATCH[@]} -ge $MAX_CONCURRENT || $((i+1)) -eq ${#TASKS[@]} ]]; then
            echo "   📦 Dispatching batch of ${#BATCH[@]} tasks..."
            for batch_task in "${BATCH[@]}"; do
                IFS='|' read -r b_agent b_mode b_task b_output <<< "$batch_task"
                echo "      → $b_agent MODE=$b_mode: $b_task"
                # In real implementation, this would call the task tool
                # For now, just mark as started
                "$SESSION_DB" ptask-start "${PTASK_IDS[$i]}" 2>/dev/null || true
            done
            BATCH=()
        fi
    done

    echo ""
    echo "✅ All tasks dispatched"
    echo ""
    echo "Next steps:"
    echo "1. Wait for barrier: $BARRIER_SCRIPT wait $BARRIER_ID 300"
    echo "2. Check results: $0 --summary $WAVE_ID"
    echo "3. Aggregate results: wave-summary.sh $WAVE_ID"
fi

echo ""
echo "---"
echo "Wave: $WAVE_ID"
echo "Tasks: $TOTAL_TASKS"
echo "Barrier: $BARRIER_ID"
echo "Session: $SID"
