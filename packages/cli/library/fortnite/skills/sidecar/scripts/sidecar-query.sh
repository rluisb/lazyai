#!/usr/bin/env bash
# sidecar-query.sh — Query SQLite index for specs relevant to a repo
# Usage: sidecar-query.sh <repo> [--json] [--min-confidence <float>]
#
# Spec reference: R3, R6

set -euo pipefail

# ---------------------------------------------------------------------------
# Source common helpers
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
OUTPUT_JSON=false
MIN_CONFIDENCE=0.0
REPO=""

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
    cat <<EOF
Usage: sidecar-query.sh <repo> [--json] [--min-confidence <float>] [--help]

Query the sidecar index for specs relevant to a repo.

Arguments:
  <repo>                     Repo name to query (required)

Options:
  --json                     Output JSON instead of text table
  --min-confidence <float>   Filter results below threshold (default: 0.0)
  --help                     Show this help message

Examples:
  sidecar-query.sh fedora
  sidecar-query.sh fedora --json
  sidecar-query.sh fedora --min-confidence 0.5
EOF
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
parse_args() {
    if [[ $# -eq 0 ]]; then
        usage >&2
        exit 1
    fi

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --json)
                OUTPUT_JSON=true
                shift
                ;;
            --min-confidence)
                if [[ $# -lt 2 ]]; then
                    echo "Error: --min-confidence requires a value" >&2
                    exit 1
                fi
                MIN_CONFIDENCE="$2"
                validate_min_confidence "$MIN_CONFIDENCE"
                shift 2
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            --*)
                echo "Error: Unknown option: $1" >&2
                usage >&2
                exit 1
                ;;
            *)
                if [[ -n "$REPO" ]]; then
                    echo "Error: Multiple repo arguments provided" >&2
                    usage >&2
                    exit 1
                fi
                REPO="$1"
                shift
                ;;
        esac
    done

    if [[ -z "$REPO" ]]; then
        echo "Error: <repo> argument is required" >&2
        usage >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# SQL string sanitization
# ---------------------------------------------------------------------------
_sql_escape() {
    printf '%s' "${1//\'/\'\'}"
}

# ---------------------------------------------------------------------------
# JSON string escaping (fallback when jq is unavailable)
# ---------------------------------------------------------------------------
_json_escape() {
    local str="$1"
    str="${str//\\/\\\\}"
    str="${str//\"/\\\"}"
    str="${str//$'\n'/\\n}"
    printf '%s' "$str"
}

# ---------------------------------------------------------------------------
# Validate min_confidence is a valid float
# ---------------------------------------------------------------------------
validate_min_confidence() {
    local value="$1"
    if [[ ! "$value" =~ ^[0-9]*\.?[0-9]+$ ]]; then
        echo "Error: --min-confidence must be a valid float (got: '$value')" >&2
        exit 1
    fi
    if awk "BEGIN {exit !($value < 0 || $value > 1)}"; then
        echo "Error: --min-confidence must be between 0.0 and 1.0 (got: '$value')" >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Validate repo is a workspace member
# ---------------------------------------------------------------------------
validate_repo() {
    local repo="$1"
    local found=false

    while IFS= read -r r; do
        if [[ "$r" == "$repo" ]]; then
            found=true
            break
        fi
    done < <(get_repos)

    if [[ "$found" != true ]]; then
        echo "${repo} is not a workspace member. Run 'sidecar add ${repo}' first." >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Ensure index exists (with auto-index support)
# ---------------------------------------------------------------------------
ensure_index() {
    local index_db="$1"

    local needs_index=false

    if [[ ! -f "$index_db" ]]; then
        needs_index=true
    else
        # Check if specs table exists (DB may be empty/corrupt)
        local table_exists
        table_exists="$(sqlite3 "$index_db" "SELECT 1 FROM sqlite_master WHERE type='table' AND name='specs' LIMIT 1;" 2>/dev/null || true)"
        if [[ "$table_exists" != "1" ]]; then
            needs_index=true
        else
            # Check if .sidecar.yml is newer than the last index
            local sidecar_mtime
            if [[ "$(uname)" == "Darwin" ]]; then
                sidecar_mtime="$(stat -f %m "$SIDECAR_YML")"
            else
                sidecar_mtime="$(stat -c %Y "$SIDECAR_YML")"
            fi

            local last_indexed_at
            last_indexed_at="$(sqlite3 "$index_db" "SELECT COALESCE(MAX(indexed_at), '1970-01-01T00:00:00Z') FROM specs;")"

            # Convert ISO timestamp to epoch seconds
            local last_indexed_epoch
            if [[ "$(uname)" == "Darwin" ]]; then
                last_indexed_epoch="$(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$last_indexed_at" "+%s" 2>/dev/null || echo 0)"
            else
                last_indexed_epoch="$(date -d "$last_indexed_at" "+%s" 2>/dev/null || echo 0)"
            fi

            if [[ "$sidecar_mtime" -gt "$last_indexed_epoch" ]]; then
                needs_index=true
            fi
        fi
    fi

    if [[ "$needs_index" != true ]]; then
        return 0
    fi

    local auto_index
    auto_index="$(yq eval '.settings.auto_index // true' "$SIDECAR_YML")"

    if [[ "$auto_index" == "true" ]]; then
        echo "Index stale or missing. Auto-indexing..." >&2
        local index_script="$SCRIPT_DIR/sidecar-index.sh"
        if [[ ! -x "$index_script" ]]; then
            echo "Error: index script is not executable: $index_script" >&2
            exit 1
        fi
        "$index_script" --force >&2
    else
        echo "Index stale or missing. Run 'sidecar index' first." >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Query SQLite for specs linked to repo
# ---------------------------------------------------------------------------
query_specs() {
    local index_db="$1"
    local repo="$2"
    local min_confidence="$3"

    local safe_repo
    safe_repo="$(_sql_escape "$repo")"

    sqlite3 "$index_db" <<EOF
SELECT s.slug, s.title, l.confidence, l.match_source, s.path
FROM spec_repo_links l
JOIN specs s ON l.spec_id = s.id
JOIN repos r ON l.repo_id = r.id
WHERE r.name = '${safe_repo}'
  AND l.confidence >= ${min_confidence}
ORDER BY l.confidence DESC;
EOF
}

# ---------------------------------------------------------------------------
# Output results as text table
# ---------------------------------------------------------------------------
output_text() {
    local repo="$1"
    local count="$2"
    shift 2

    if [[ "$count" -eq 0 ]]; then
        echo "No specs found for '${repo}'"
        return 0
    fi

    echo "Specs relevant to '${repo}' (${count} found):"
    echo ""
    printf "  %-40s %-12s %s\n" "SLUG" "CONFIDENCE" "MATCH"

    while IFS='|' read -r slug title confidence match_source path; do
        printf "  %-40s %-12s %s\n" "$slug" "$confidence" "$match_source"
    done <<< "$@"
}

# ---------------------------------------------------------------------------
# Output results as JSON
# ---------------------------------------------------------------------------
output_json() {
    local repo="$1"
    local count="$2"
    shift 2

    if [[ "$count" -eq 0 ]]; then
        if command -v jq >/dev/null 2>&1; then
            jq -n --arg repo "$repo" '{repo: $repo, specs: []}'
        else
            printf '{"repo":"%s","specs":[]}\n' "$(_json_escape "$repo")"
        fi
        return 0
    fi

    if command -v jq >/dev/null 2>&1; then
        # Build JSON array using jq
        local json_array="[]"
        while IFS='|' read -r slug title confidence match_source path; do
            json_array=$(jq --arg slug "$slug" \
                          --arg title "$title" \
                          --argjson confidence "$confidence" \
                          --arg match_source "$match_source" \
                          --arg path "$path" \
                          '. + [{slug: $slug, title: $title, confidence: $confidence, match_source: $match_source, path: $path}]' \
                          <<< "$json_array")
        done <<< "$@"
        jq -n --arg repo "$repo" --argjson specs "$json_array" '{repo: $repo, specs: $specs}'
    else
        # Fallback: manual JSON construction with printf
        printf '{"repo":"%s","specs":[' "$(_json_escape "$repo")"
        local first=true
        while IFS='|' read -r slug title confidence match_source path; do
            if [[ "$first" == true ]]; then
                first=false
            else
                printf ","
            fi
            printf '{"slug":"%s","title":"%s","confidence":%s,"match_source":"%s","path":"%s"}' \
                "$(_json_escape "$slug")" \
                "$(_json_escape "$title")" \
                "$confidence" \
                "$(_json_escape "$match_source")" \
                "$(_json_escape "$path")"
        done <<< "$@"
        printf ']}\n'
    fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    parse_args "$@"

    # Discover workspace
    discover_sidecar
    resolve_sidecar_dir

    # Validate repo membership
    validate_repo "$REPO"

    # Construct index path manually (read-only: no mkdir)
    local index_db="$SIDECAR_DIR/.sidecar/index.db"

    # Check if index exists and is valid
    local needs_index=false
    if [[ ! -f "$index_db" ]]; then
        needs_index=true
    else
        local table_exists
        table_exists="$(sqlite3 "$index_db" "SELECT 1 FROM sqlite_master WHERE type='table' AND name='specs' LIMIT 1;" 2>/dev/null || true)"
        if [[ "$table_exists" != "1" ]]; then
            needs_index=true
        fi
    fi

    if [[ "$needs_index" == true ]]; then
        local auto_index
        auto_index="$(yq eval '.settings.auto_index // true' "$SIDECAR_YML")"
        if [[ "$auto_index" == "true" ]]; then
            echo "Index missing. Running sidecar-index.sh..." >&2
            local index_script="$SCRIPT_DIR/sidecar-index.sh"
            if [[ ! -x "$index_script" ]]; then
                echo "Error: index script is not executable: $index_script" >&2
                exit 1
            fi
            "$index_script" >&2
        else
            echo "Error: index missing. Run 'sidecar index' first." >&2
            exit 1
        fi
    fi

    # Check staleness (table exists, check mtime)
    ensure_index "$index_db"

    # Query specs
    local results
    results="$(query_specs "$index_db" "$REPO" "$MIN_CONFIDENCE")"

    # Count results
    local count=0
    if [[ -n "$results" ]]; then
        count="$(wc -l <<< "$results" | tr -d ' ')"
    fi

    # Output
    if [[ "$OUTPUT_JSON" == true ]]; then
        output_json "$REPO" "$count" "$results"
    else
        output_text "$REPO" "$count" "$results"
    fi

    exit 0
}

main "$@"
