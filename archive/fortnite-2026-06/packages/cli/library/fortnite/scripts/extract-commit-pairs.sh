#!/usr/bin/env bash
# scripts/extract-commit-pairs.sh — Extract commit pairs from git history into JSONL
# Usage: ./scripts/extract-commit-pairs.sh [--repo PATH] [--output PATH] [--limit N]
#
# Outputs one JSON object per line (JSONL) matching the commit pair schema:
# {
#   "input": { "diff_summary": "...", "files_changed": [...], "lines_added": N, "lines_removed": N },
#   "output": { "commit_message": "...", "commit_type": "...", "scope": "...", "breaking": false },
#   "metadata": { "sha": "...", "author": "...", "timestamp": "...", "conventional": true }
# }

set -euo pipefail
IFS=$'\n\t'

REPO_ROOT="."
OUTPUT_PATH=""
LIMIT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_ROOT="$2"; shift 2 ;;
        --output) OUTPUT_PATH="$2"; shift 2 ;;
        --limit) LIMIT="$2"; shift 2 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Resolve absolute path
REPO_ROOT="$(cd "$REPO_ROOT" && pwd 2>/dev/null)" || {
    echo "Error: repo path does not exist or is not accessible: $REPO_ROOT" >&2
    exit 1
}

# Validate git repository
if [[ ! -d "$REPO_ROOT/.git" ]]; then
    echo "Error: not a git repository: $REPO_ROOT" >&2
    exit 1
fi

# Build git log options
GIT_LOG_OPTS=()
if [[ -n "$LIMIT" ]]; then
    GIT_LOG_OPTS+=("--max-count=$LIMIT")
fi

# Escape a string for JSON
json_escape() {
    local str="$1"
    str="${str//\\/\\\\}"
    str="${str//\"/\\\"}"
    str="${str//$'\n'/\\n}"
    str="${str//$'\r'/}"
    str="${str//$'\t'/\\t}"
    printf '%s' "$str"
}

# Parse conventional commit
# Format: type(scope)!: message
# Or: type!: message
# Or: type(scope): message
# Or: type: message
# BREAKING CHANGE: in body also marks breaking
parse_conventional() {
    local subject="$1"
    local body="$2"
    local type=""
    local scope=""
    local breaking="false"
    local conventional="false"

    # Check for BREAKING CHANGE in body
    if [[ "$body" == *"BREAKING CHANGE:"* ]] || [[ "$body" == *"BREAKING-CHANGE:"* ]]; then
        breaking="true"
    fi

    # Match conventional commit pattern
    if [[ "$subject" =~ ^([a-zA-Z]+)(\(([a-zA-Z0-9_-]+)\))?(!)?:\ (.+)$ ]]; then
        conventional="true"
        type="${BASH_REMATCH[1]}"
        scope="${BASH_REMATCH[3]:-}"
        local bang="${BASH_REMATCH[4]:-}"
        if [[ "$bang" == "!" ]]; then
            breaking="true"
        fi
    fi

    printf '%s\t%s\t%s\t%s' "$conventional" "$type" "$scope" "$breaking"
}

# Build files_changed JSON array from summary string
build_files_array() {
    local summary="$1"
    local result=""
    local first=true

    if [[ -z "$summary" ]]; then
        echo ""
        return
    fi

    # Split by ", " and build array
    local IFS_OLD="$IFS"
    IFS=','
    for f in $summary; do
        # Trim whitespace
        f="$(echo "$f" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        [[ -z "$f" ]] && continue

        if [[ "$first" == "true" ]]; then
            result="\"$(json_escape "$f")\""
            first=false
        else
            result="$result,\"$(json_escape "$f")\""
        fi
    done
    IFS="$IFS_OLD"

    echo "$result"
}

# Collect stats for a single commit
collect_stats() {
    local repo="$1"
    local sha="$2"
    local stats
    stats=$(git -C "$repo" show --numstat --format="" "$sha" 2>/dev/null || true)

    local files=0
    local added=0
    local removed=0
    local summary=""

    while IFS=$'\t' read -r a r f; do
        [[ -z "$f" ]] && continue
        [[ "$f" =~ ^[[:space:]]*$ ]] && continue

        files=$((files + 1))
        if [[ "$a" != "-" ]]; then
            added=$((added + a))
        fi
        if [[ "$r" != "-" ]]; then
            removed=$((removed + r))
        fi
        if [[ -z "$summary" ]]; then
            summary="$f"
        else
            summary="$summary, $f"
        fi
    done <<< "$stats"

    echo "$files"
    echo "$added"
    echo "$removed"
    echo "$summary"
}

# Output a single commit as JSON
output_commit_json() {
    local sha="$1"
    local author="$2"
    local timestamp="$3"
    local subject="$4"
    local body="$5"
    local files_changed="$6"
    local lines_added="$7"
    local lines_removed="$8"
    local diff_summary="$9"

    local conv_data
    conv_data=$(parse_conventional "$subject" "$body")
    local conventional="$(printf '%s' "$conv_data" | cut -f1)"
    local commit_type="$(printf '%s' "$conv_data" | cut -f2)"
    local scope="$(printf '%s' "$conv_data" | cut -f3)"
    local breaking="$(printf '%s' "$conv_data" | cut -f4)"

    # Build files_changed array
    local files_array
    files_array=$(build_files_array "$diff_summary")

    # Handle nulls for non-conventional commits
    local json_type="null"
    local json_scope="null"
    if [[ "$conventional" == "true" ]]; then
        json_type="\"$(json_escape "$commit_type")\""
        if [[ -n "$scope" ]]; then
            json_scope="\"$(json_escape "$scope")\""
        fi
    fi

    # Build JSON
    printf '{"input":{"diff_summary":"%s","files_changed":[%s],"lines_added":%s,"lines_removed":%s},"output":{"commit_message":"%s","commit_type":%s,"scope":%s,"breaking":%s},"metadata":{"sha":"%s","author":"%s","timestamp":"%s","conventional":%s}}\n' \
        "$(json_escape "${diff_summary:-}")" \
        "$files_array" \
        "${lines_added:-0}" \
        "${lines_removed:-0}" \
        "$(json_escape "$subject")" \
        "$json_type" \
        "$json_scope" \
        "$breaking" \
        "$(json_escape "$sha")" \
        "$(json_escape "$author")" \
        "$(json_escape "$timestamp")" \
        "$conventional"
}

# Main processing
main() {
    # Get list of SHAs
    local shas
    shas=$(git -C "$REPO_ROOT" log --format="%H" "${GIT_LOG_OPTS[@]}" 2>/dev/null || true)

    if [[ -z "$shas" ]]; then
        echo "No commits found in repository." >&2
        exit 0
    fi

    # Process each commit
    while IFS= read -r sha; do
        [[ -z "$sha" ]] && continue

        # Get metadata
        local author timestamp subject body
        author=$(git -C "$REPO_ROOT" show --format="%an" --no-patch "$sha" 2>/dev/null || true)
        timestamp=$(git -C "$REPO_ROOT" show --format="%ai" --no-patch "$sha" 2>/dev/null || true)
        subject=$(git -C "$REPO_ROOT" show --format="%s" --no-patch "$sha" 2>/dev/null || true)
        body=$(git -C "$REPO_ROOT" show --format="%b" --no-patch "$sha" 2>/dev/null || true)

        # Get stats
        local stats
        stats=$(collect_stats "$REPO_ROOT" "$sha")
        local files_changed="$(echo "$stats" | sed -n '1p')"
        local lines_added="$(echo "$stats" | sed -n '2p')"
        local lines_removed="$(echo "$stats" | sed -n '3p')"
        local diff_summary="$(echo "$stats" | sed -n '4p')"

        output_commit_json "$sha" "$author" "$timestamp" "$subject" "$body" \
            "$files_changed" "$lines_added" "$lines_removed" "$diff_summary"
    done <<< "$shas"
}

# Main output
if [[ -n "$OUTPUT_PATH" ]]; then
    main > "$OUTPUT_PATH"
else
    main
fi

exit 0
