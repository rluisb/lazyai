#!/usr/bin/env bash
# check-file-collision.sh — Pre-flight check for parallel write conflicts
# Usage: ./check-file-collision.sh <wave-file> [--strict]
#
# Checks if any parallel tasks in a wave would write to the same file or
# overlapping directories. Prevents race conditions before dispatch.
#
# Wave file format (one task per line):
#   agent|mode|task-description|output-path
#
# Exit codes:
#   0 — No collisions found
#   1 — Collisions detected
#   2 — Invalid wave file

set -euo pipefail

WAVE_FILE="${1:-}"
STRICT=false

shift 2>/dev/null || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --strict) STRICT=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$WAVE_FILE" ]]; then
    echo "Usage: $0 <wave-file> [--strict]"
    echo ""
    echo "Wave file format (one task per line):"
    echo "  agent|mode|task-description|output-path"
    exit 1
fi

if [[ ! -f "$WAVE_FILE" ]]; then
    echo "❌ Wave file not found: $WAVE_FILE"
    exit 2
fi

# Parse wave file and collect output paths
declare -A PATH_TO_TASKS
COLLISIONS=()
OVERLAPS=()

echo "🔍 Checking for file collisions in: $WAVE_FILE"
echo ""

LINE_NUM=0
while IFS='|' read -r agent mode task output_path; do
    LINE_NUM=$((LINE_NUM + 1))

    # Skip empty lines and comments
    [[ -z "$agent" || "$agent" =~ ^# ]] && continue

    # Validate format
    if [[ -z "$mode" || -z "$task" || -z "$output_path" ]]; then
        echo "❌ Invalid format at line $LINE_NUM: expected agent|mode|task|output-path"
        exit 2
    fi

    # Normalize path (remove trailing slashes, resolve relative)
    normalized_path=$(echo "$output_path" | sed 's:/*$::')

    # Check for exact collision
    if [[ -n "${PATH_TO_TASKS[$normalized_path]:-}" ]]; then
        COLLISIONS+=("$normalized_path (line ${PATH_TO_TASKS[$normalized_path]} and $LINE_NUM)")
    fi
    PATH_TO_TASKS[$normalized_path]=$LINE_NUM

    # Check for directory overlap (strict mode)
    if [[ "$STRICT" == true ]]; then
        for existing_path in "${!PATH_TO_TASKS[@]}"; do
            # Check if one path is a parent of another
            if [[ "$normalized_path" == "$existing_path"/* || "$existing_path" == "$normalized_path"/* ]]; then
                if [[ "$normalized_path" != "$existing_path" ]]; then
                    OVERLAPS+=("$existing_path → $normalized_path (directory overlap)")
                fi
            fi
        done
    fi
done < "$WAVE_FILE"

# Report results
HAS_ISSUES=false

if [[ ${#COLLISIONS[@]} -gt 0 ]]; then
    HAS_ISSUES=true
    echo "❌ Exact file collisions detected:"
    for collision in "${COLLISIONS[@]}"; do
        echo "   - $collision"
    done
    echo ""
fi

if [[ ${#OVERLAPS[@]} -gt 0 ]]; then
    HAS_ISSUES=true
    echo "⚠️  Directory overlaps detected (strict mode):"
    for overlap in "${OVERLAPS[@]}"; do
        echo "   - $overlap"
    done
    echo ""
fi

if [[ "$HAS_ISSUES" == false ]]; then
    echo "✅ No file collisions found"
    echo ""
    echo "   Checked ${#PATH_TO_TASKS[@]} unique output paths"
    if [[ "$STRICT" == true ]]; then
        echo "   Strict mode: directory overlap check enabled"
    fi
    exit 0
else
    echo "💡 Fix suggestions:"
    echo "   - Use unique output paths for each parallel task"
    echo "   - Use wave-specific subdirectories: worktrees/wave-<id>/<task>/"
    echo "   - Use task-specific filenames: <task-slug>-result.md"
    echo ""
    exit 1
fi
