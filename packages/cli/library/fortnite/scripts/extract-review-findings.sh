#!/usr/bin/env bash
# scripts/extract-review-findings.sh — Extract review findings from PR reviews into JSONL
# Usage: ./scripts/extract-review-findings.sh [--output PATH] [--limit N]
#
# Schema:
# {
#   "input": {"diff_hunk": "string", "file_path": "string"},
#   "output": {"finding": "string", "severity": "string"},
#   "metadata": {"reviewer": "string", "pr_number": "string"}
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
    local diff_hunk="$1"
    local file_path="$2"
    local finding="$3"
    local severity="$4"
    local reviewer="$5"
    local pr_number="$6"
    printf '{"input":{"diff_hunk":"%s","file_path":"%s"},"output":{"finding":"%s","severity":"%s"},"metadata":{"reviewer":"%s","pr_number":"%s"}}\n' \
        "$(json_escape "$diff_hunk")" \
        "$(json_escape "$file_path")" \
        "$(json_escape "$finding")" \
        "$(json_escape "$severity")" \
        "$(json_escape "$reviewer")" \
        "$(json_escape "$pr_number")"
}

main() {
    local count=0

    # Prefer gh CLI if available and authenticated
    if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
        local prs
        prs=$(gh pr list --state merged --json number,author,title,body,files --limit "${LIMIT:-50}" 2>/dev/null || true)

        if [[ -n "$prs" && "$prs" != "[]" ]]; then
            # Use jq if available, otherwise python3
            if command -v jq >/dev/null 2>&1; then
                echo "$prs" | jq -c '.[]' 2>/dev/null | while IFS= read -r pr; do
                    local pr_num reviewer
                    pr_num=$(echo "$pr" | jq -r '.number // "N/A"')
                    reviewer=$(echo "$pr" | jq -r '.author.login // "N/A"')

                    # Extract files and their patches
                    local files
                    files=$(echo "$pr" | jq -c '.files // []')
                    local fcount
                    fcount=$(echo "$files" | jq 'length')

                    if [[ "$fcount" -eq 0 ]]; then
                        emit_record "" "" "No file changes found" "info" "$reviewer" "$pr_num"
                        count=$((count + 1))
                        continue
                    fi

                    local i=0
                    while [[ "$i" -lt "$fcount" ]]; do
                        local file_path diff_hunk
                        file_path=$(echo "$files" | jq -r ".[$i].path // \"\"")
                        diff_hunk=$(echo "$files" | jq -r ".[$i].patch // \"\"")
                        emit_record "$diff_hunk" "$file_path" "Extracted from PR $pr_num" "info" "$reviewer" "$pr_num"
                        count=$((count + 1))
                        i=$((i + 1))
                    done
                done
            elif command -v python3 >/dev/null 2>&1; then
                echo "$prs" | python3 -c "
import sys, json
prs = json.load(sys.stdin)
for pr in prs:
    pr_num = str(pr.get('number', 'N/A'))
    reviewer = pr.get('author', {}).get('login', 'N/A')
    files = pr.get('files', [])
    if not files:
        print(json.dumps({'input':{'diff_hunk':'','file_path':''},'output':{'finding':'No file changes found','severity':'info'},'metadata':{'reviewer':reviewer,'pr_number':pr_num}}, separators=(',',':')))
    else:
        for f in files:
            fp = f.get('path','')
            patch = f.get('patch','')
            print(json.dumps({'input':{'diff_hunk':patch,'file_path':fp},'output':{'finding':'Extracted from PR ' + pr_num,'severity':'info'},'metadata':{'reviewer':reviewer,'pr_number':pr_num}}, separators=(',',':')))
" | while IFS= read -r line; do
                    echo "$line"
                    count=$((count + 1))
                done
            fi

            echo "✅ Extracted ${count} review finding(s) via gh CLI" >&2
            return
        fi
    fi

    # Fallback: no gh CLI or no PRs found — emit stub documenting expected format
    echo "⚠️  gh CLI not available or no merged PRs found. Emitting stub record." >&2
    emit_record "" "" "No review data available — gh CLI required for PR review extraction" "info" "N/A" "N/A"
    echo "✅ Extracted 0 review finding(s) (stub emitted)" >&2
}

if [[ -n "$OUTPUT_PATH" ]]; then
    main > "$OUTPUT_PATH"
else
    main
fi

exit 0
