#!/usr/bin/env bash
# scripts/extract-lesson-pairs.sh — Extract lesson pairs from .specify/memory/ into JSONL
# Usage: ./scripts/extract-lesson-pairs.sh [--output PATH] [--limit N]
#
# Schema:
# {
#   "input": {"correction": "string", "context": "string"},
#   "output": {"lesson": "string", "category": "string"},
#   "metadata": {"session_id": "string", "agent": "string"}
# }

set -euo pipefail
IFS=$'\n\t'

MEMORY_DIR=".specify/memory"
OUTPUT_PATH=""
LIMIT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --output) OUTPUT_PATH="$2"; shift 2 ;;
        --limit) LIMIT="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

json_escape() {
    local str="$1"
    str="${str//\\/\\\\}"
    str="${str//\"/\\\"}"
    str="${str//$'\n'/\\n}"
    str="${str//$'\r'/}"
    str="${str//$'\t'/\\t}"
    printf '%s' "$str"
}

emit_record() {
    local correction="$1"
    local context="$2"
    local lesson="$3"
    local category="$4"
    local session_id="$5"
    local agent="$6"
    printf '{"input":{"correction":"%s","context":"%s"},"output":{"lesson":"%s","category":"%s"},"metadata":{"session_id":"%s","agent":"%s"}}\n' \
        "$(json_escape "$correction")" \
        "$(json_escape "$context")" \
        "$(json_escape "$lesson")" \
        "$(json_escape "$category")" \
        "$(json_escape "$session_id")" \
        "$(json_escape "$agent")"
}

main() {
    local count=0

    if [[ ! -d "$MEMORY_DIR" ]]; then
        echo "⚠️  Memory directory not found: $MEMORY_DIR. Emitting stub record." >&2
        emit_record "" "" "No lesson data available" "stub" "N/A" "N/A"
        echo "✅ Extracted 0 lesson pair(s) (stub emitted)" >&2
        return
    fi

    local files
    files=$(find "$MEMORY_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null || true)

    if [[ -z "$files" ]]; then
        echo "⚠️  No lesson files found in $MEMORY_DIR. Emitting stub record." >&2
        emit_record "" "" "No lesson data available" "stub" "N/A" "N/A"
        echo "✅ Extracted 0 lesson pair(s) (stub emitted)" >&2
        return
    fi

    while IFS= read -r f; do
        [[ -f "$f" ]] || continue

        local correction context lesson category session_id agent
        correction="$(grep -m1 '^\*\*Correction:\*\*' "$f" 2>/dev/null | sed 's/^\*\*Correction:\*\* //' || true)"
        context="$(grep -m1 '^\*\*Context:\*\*' "$f" 2>/dev/null | sed 's/^\*\*Context:\*\* //' || true)"
        lesson="$(grep -m1 '^\*\*Mistake:\*\*' "$f" 2>/dev/null | sed 's/^\*\*Mistake:\*\* //' || true)"
        category="$(grep -m1 '^\*\*Tags:\*\*' "$f" 2>/dev/null | sed 's/^\*\*Tags:\*\* //' | awk -F, '{print $1}' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' || true)"
        session_id="$(basename "$f" .md)"
        agent="lesson-loot"

        emit_record "$correction" "$context" "$lesson" "$category" "$session_id" "$agent"

        count=$((count + 1))
        if [[ -n "$LIMIT" && "$count" -ge "$LIMIT" ]]; then
            break
        fi
    done <<< "$files"

    echo "✅ Extracted ${count} lesson pair(s)" >&2
}

if [[ -n "$OUTPUT_PATH" ]]; then
    main > "$OUTPUT_PATH"
else
    main
fi

exit 0
