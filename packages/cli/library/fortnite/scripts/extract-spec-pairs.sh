#!/usr/bin/env bash
# scripts/extract-spec-pairs.sh — Extract spec pairs from bee-gone/specs/ into JSONL
# Usage: ./scripts/extract-spec-pairs.sh [--output PATH] [--limit N]
#
# Schema:
# {
#   "input": {"goal": "string", "research": "string"},
#   "output": {"spec": "string", "tasks": "string"},
#   "metadata": {"spec_id": "string", "timestamp": "string"}
# }

set -euo pipefail
IFS=$'\n\t'

SPECS_DIR="bee-gone/specs"
OUTPUT_PATH=""
LIMIT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --output) OUTPUT_PATH="$2"; shift 2 ;;
        --limit) LIMIT="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

if [[ ! -d "$SPECS_DIR" ]]; then
    echo "Error: specs directory not found: $SPECS_DIR" >&2
    exit 1
fi

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
    local spec_id="$1"
    local goal="$2"
    local research="$3"
    local spec="$4"
    local tasks="$5"
    local ts="$6"
    printf '{"input":{"goal":"%s","research":"%s"},"output":{"spec":"%s","tasks":"%s"},"metadata":{"spec_id":"%s","timestamp":"%s"}}\n' \
        "$(json_escape "$goal")" \
        "$(json_escape "$research")" \
        "$(json_escape "$spec")" \
        "$(json_escape "$tasks")" \
        "$(json_escape "$spec_id")" \
        "$(json_escape "$ts")"
}

main() {
    local count=0
    for spec_dir in "$SPECS_DIR"/*; do
        [[ -d "$spec_dir" ]] || continue
        local spec_id
        spec_id="$(basename "$spec_dir")"
        local spec_file="$spec_dir/SPEC.md"
        local tasks_file="$spec_dir/tasks.md"

        [[ -f "$spec_file" ]] || continue
        [[ -f "$tasks_file" ]] || continue

        local goal research spec tasks ts
        goal="$(grep -m1 -E '^\*\*Goal:\*\*|^## Goal' "$spec_file" 2>/dev/null | sed -E 's/^\*\*Goal:\*\* //; s/^## Goal[[:space:]]*//' || true)"
        # Use first paragraph after Goal as research/context
        research="$(awk '/^\*\*Goal:\*\*|^## Goal/{getline; while(NF==0){getline}; print; exit}' "$spec_file" 2>/dev/null || true)"
        spec="$(cat "$spec_file" 2>/dev/null || true)"
        tasks="$(cat "$tasks_file" 2>/dev/null || true)"
        ts="$(date -r "$spec_file" '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date '+%Y-%m-%dT%H:%M:%SZ')"

        emit_record "$spec_id" "$goal" "$research" "$spec" "$tasks" "$ts"

        count=$((count + 1))
        if [[ -n "$LIMIT" && "$count" -ge "$LIMIT" ]]; then
            break
        fi
    done

    echo "✅ Extracted ${count} spec pair(s)" >&2
}

if [[ -n "$OUTPUT_PATH" ]]; then
    main > "$OUTPUT_PATH"
else
    main
fi

exit 0
