#!/usr/bin/env bash
# scripts/extract-pr-description-pairs.sh — Extract PR description pairs from git history into JSONL
# Usage: ./scripts/extract-pr-description-pairs.sh [--output PATH] [--limit N]
#
# Schema:
# {
#   "input": {"diff": "string", "files_changed": ["string"]},
#   "output": {"pr_title": "string", "pr_body": "string"},
#   "metadata": {"pr_number": "string", "author": "string"}
# }

set -euo pipefail
IFS=$'\n\t'

OUTPUT_PATH=""
LIMIT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --output) OUTPUT_PATH="$2"; shift 2 ;;
        --limit) LIMIT="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "Error: not a git repository" >&2
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
    local diff="$1"
    local files="$2"
    local title="$3"
    local body="$4"
    local pr_num="$5"
    local author="$6"
    printf '{"input":{"diff":"%s","files_changed":[%s]},"output":{"pr_title":"%s","pr_body":"%s"},"metadata":{"pr_number":"%s","author":"%s"}}\n' \
        "$(json_escape "$diff")" \
        "$files" \
        "$(json_escape "$title")" \
        "$(json_escape "$body")" \
        "$(json_escape "$pr_num")" \
        "$(json_escape "$author")"
}

main() {
    local count=0
    local log_opts=()
    if [[ -n "$LIMIT" ]]; then
        log_opts+=("--max-count=$LIMIT")
    fi

    # Look for merge commits that reference PRs
    local merges
    merges=$(git log --grep="Merge pull request" --format="%H" "${log_opts[@]}" 2>/dev/null || true)

    if [[ -z "$merges" ]]; then
        echo "⚠️  No merge commits with PR references found. Emitting stub record." >&2
        emit_record "" "" "No PR data available" "This repository has no merge commits referencing pull requests." "N/A" "N/A"
        echo "✅ Extracted 0 PR description pair(s) (stub emitted)" >&2
        return
    fi

    while IFS= read -r sha; do
        [[ -z "$sha" ]] && continue

        local subject body author diff files_json pr_num
        subject=$(git show --format="%s" --no-patch "$sha" 2>/dev/null || true)
        body=$(git show --format="%b" --no-patch "$sha" 2>/dev/null || true)
        author=$(git show --format="%an" --no-patch "$sha" 2>/dev/null || true)
        diff=$(git show --stat --format="" "$sha" 2>/dev/null || true)

        # Extract PR number from subject like "Merge pull request #123 from ..."
        pr_num="N/A"
        if [[ "$subject" =~ Merge[[:space:]]pull[[:space:]]request[[:space:]]#([0-9]+) ]]; then
            pr_num="${BASH_REMATCH[1]}"
        fi

        # Build files_changed array from diff stat
        files_json=""
        local first=true
        while IFS= read -r line; do
            [[ -z "$line" ]] && continue
            local f
            f="$(echo "$line" | awk '{print $NF}')"
            [[ -z "$f" ]] && continue
            if [[ "$first" == "true" ]]; then
                files_json="\"$(json_escape "$f")\""
                first=false
            else
                files_json="$files_json,\"$(json_escape "$f")\""
            fi
        done <<< "$diff"

        emit_record "$diff" "$files_json" "$subject" "$body" "$pr_num" "$author"
        count=$((count + 1))
    done <<< "$merges"

    echo "✅ Extracted ${count} PR description pair(s)" >&2
}

if [[ -n "$OUTPUT_PATH" ]]; then
    main > "$OUTPUT_PATH"
else
    main
fi

exit 0
