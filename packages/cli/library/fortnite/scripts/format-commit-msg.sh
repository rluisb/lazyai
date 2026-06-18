#!/usr/bin/env bash
# scripts/format-commit-msg.sh — Deterministic commit message formatter
# Reads staged diff and infers conventional commit type deterministically.
# Usage: ./scripts/format-commit-msg.sh [--diff FILE] [--json] [--dry-run]

set -euo pipefail
IFS=$'\n\t'

DIFF_SOURCE=""
OUTPUT_JSON=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --diff) DIFF_SOURCE="$2"; shift 2 ;;
        --json) OUTPUT_JSON=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        *) shift ;;
    esac
done

# Read diff
if [[ -n "$DIFF_SOURCE" ]]; then
    if [[ -f "$DIFF_SOURCE" ]]; then
        DIFF=$(cat "$DIFF_SOURCE")
    else
        DIFF=""
    fi
else
    DIFF=$(git diff --cached 2>/dev/null || true)
fi

if [[ -z "$DIFF" ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.commit_message.v1","status":"no_changes","type":null,"scope":null,"subject":null,"body":null,"requires_llm_fallback":false,"reason":"No staged changes detected"}'
    else
        echo "No staged changes detected."
    fi
    exit 0
fi

# Extract changed files from diff
CHANGED_FILES=$(echo "$DIFF" | grep -E '^diff --git' | sed -E 's|^diff --git a/([^ ]+) b/[^ ]+|\1|' || true)
FILE_COUNT=$(echo "$CHANGED_FILES" | grep -v '^$' | wc -l | tr -d ' ')

if [[ "$FILE_COUNT" -eq 0 ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.commit_message.v1","status":"no_changes","type":null,"scope":null,"subject":null,"body":null,"requires_llm_fallback":false,"reason":"No files in staged diff"}'
    else
        echo "No files in staged diff."
    fi
    exit 0
fi

# Classification helpers
is_only_test_files() {
    local files="$1"
    local non_test=0
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        if [[ ! "$f" =~ ^(test/|tests/|spec/|__tests__/) ]] && [[ ! "$f" =~ \.(test|spec)\.(js|ts|jsx|tsx|py|rb|go|rs|java|kt)$ ]]; then
            non_test=$((non_test + 1))
        fi
    done <<< "$files"
    [[ "$non_test" -eq 0 ]]
}

is_only_docs_files() {
    local files="$1"
    local non_docs=0
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        if [[ ! "$f" =~ ^(docs/|doc/|README|CHANGELOG|LICENSE|CONTRIBUTING) ]] && [[ ! "$f" =~ \.md$ ]]; then
            non_docs=$((non_docs + 1))
        fi
    done <<< "$files"
    [[ "$non_docs" -eq 0 ]]
}

is_only_formatting() {
    local diff="$1"
    # If diff contains only whitespace changes (no +/- lines with non-whitespace)
    local has_real_change
    has_real_change=$(echo "$diff" | grep -E '^[\+\-].*[^[:space:]]' || true)
    [[ -z "$has_real_change" ]]
}

is_only_deletions() {
    local diff="$1"
    local has_real_add
    has_real_add=$(echo "$diff" | grep -E '^\+\+\+ b/' || true)
    [[ -z "$has_real_add" ]]
}

is_only_new_files() {
    local diff="$1"
    local has_existing
    has_existing=$(echo "$diff" | grep -E '^--- a/' || true)
    local has_add
    has_add=$(echo "$diff" | grep -E '^\+\+\+ b/' || true)
    # If there are additions to real files but no existing files (only new files)
    [[ -n "$has_add" ]] && [[ -z "$has_existing" ]]
}

is_ci_or_build_files() {
    local files="$1"
    local match=0
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        if [[ "$f" =~ ^(\.github/|\.ci/|scripts/|\.husky/|Dockerfile|docker-compose|Makefile|\.gitlab-ci) ]]; then
            match=$((match + 1))
        fi
    done <<< "$files"
    [[ "$match" -gt 0 ]]
}

is_only_modifications() {
    local diff="$1"
    local files="$2"
    local file_count
    file_count=$(echo "$files" | grep -v '^$' | wc -l | tr -d ' ')
    local hunk_count
    hunk_count=$(echo "$diff" | grep -cE '^@@ ' || true)
    local has_existing
    has_existing=$(echo "$diff" | grep -E '^--- a/' || true)
    local has_add
    has_add=$(echo "$diff" | grep -E '^\+\+\+ b/' || true)
    local has_new
    has_new=$(echo "$diff" | grep -E '^--- /dev/null' || true)
    local has_del
    has_del=$(echo "$diff" | grep -E '^\+\+\+ /dev/null' || true)
    [[ "$file_count" -eq 1 ]] && [[ "$hunk_count" -eq 1 ]] && [[ -n "$has_existing" ]] && [[ -n "$has_add" ]] && [[ -z "$has_new" ]] && [[ -z "$has_del" ]]
}

# Determine type
TYPE=""
REASON=""
REQUIRES_LLM_FALLBACK=false
STATUS="ok"

if is_only_test_files "$CHANGED_FILES"; then
    TYPE="test"
    REASON="All changed files are test files"
elif is_only_docs_files "$CHANGED_FILES"; then
    TYPE="docs"
    REASON="All changed files are documentation"
elif is_only_formatting "$DIFF"; then
    TYPE="style"
    REASON="Only whitespace/formatting changes"
elif is_only_deletions "$DIFF"; then
    TYPE="chore"
    REASON="Only file deletions"
elif is_only_new_files "$DIFF"; then
    TYPE="feat"
    REASON="Only new files added"
elif is_ci_or_build_files "$CHANGED_FILES"; then
    TYPE="ci"
    REASON="CI/build script changes"
elif is_only_modifications "$DIFF" "$CHANGED_FILES"; then
    TYPE="fix"
    REASON="Modification to existing file"
else
    # Mixed or ambiguous
    TYPE="refactor"
    REASON="Mixed changes — defaulting to refactor"
    REQUIRES_LLM_FALLBACK=true
    STATUS="ambiguous"
fi

# Extract scope from most common directory
SCOPE=""
if [[ -n "$CHANGED_FILES" ]]; then
    SCOPE=$(echo "$CHANGED_FILES" | grep -v '^$' | awk -F'/' '{print $1}' | sort | uniq -c | sort -rn | head -1 | sed 's/^ *[0-9]* *//' || true)
    if [[ "$SCOPE" == "." ]] || [[ -z "$SCOPE" ]]; then
        SCOPE=""
    fi
fi

# Generate subject
FIRST_FILE=$(echo "$CHANGED_FILES" | grep -v '^$' | head -1 || true)
FILE_BASENAME=$(basename "$FIRST_FILE" 2>/dev/null || echo "")
if [[ -n "$FILE_BASENAME" ]]; then
    SUBJECT="${TYPE}: update ${FILE_BASENAME}"
else
    SUBJECT="${TYPE}: update files"
fi

# Body: list changed files
BODY=""
if [[ "$FILE_COUNT" -gt 1 ]]; then
    BODY=$(echo "$CHANGED_FILES" | grep -v '^$' | sed 's/^/- /' | head -20)
fi

if [[ "$DRY_RUN" == "true" ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.commit_message.v1","status":"dry_run","type":"'"$TYPE"'","scope":"'"${SCOPE:-}"'","subject":"'"$SUBJECT"'","body":null,"requires_llm_fallback":'"$REQUIRES_LLM_FALLBACK"',"reason":"'"$REASON"'"}'
    else
        echo "[dry-run] Type: $TYPE"
        echo "[dry-run] Scope: ${SCOPE:-(none)}"
        echo "[dry-run] Subject: $SUBJECT"
        echo "[dry-run] Reason: $REASON"
    fi
    exit 0
fi

# Escape JSON strings
escape_json() {
    local str="$1"
    str="${str//\\/\\\\}"
    str="${str//\"/\\\"}"
    str="${str//$'\n'/\\n}"
    str="${str//$'\r'/}"
    str="${str//$'\t'/\\t}"
    echo "$str"
}

if [[ "$OUTPUT_JSON" == "true" ]]; then
    JSON_SCOPE="null"
    [[ -n "$SCOPE" ]] && JSON_SCOPE='"'"$(escape_json "$SCOPE")"'"'
    JSON_BODY="null"
    [[ -n "$BODY" ]] && JSON_BODY='"'"$(escape_json "$BODY")"'"'
    echo '{"schema_version":"command.commit_message.v1","status":"'"$STATUS"'","type":"'"$TYPE"'","scope":'"$JSON_SCOPE"',"subject":"'"$(escape_json "$SUBJECT")"'","body":'"$JSON_BODY"',"requires_llm_fallback":'"$REQUIRES_LLM_FALLBACK"',"reason":"'"$(escape_json "$REASON")"'"}'
else
    echo "Type: $TYPE"
    echo "Scope: ${SCOPE:-(none)}"
    echo "Subject: $SUBJECT"
    if [[ -n "$BODY" ]]; then
        echo ""
        echo "$BODY"
    fi
    if [[ "$REQUIRES_LLM_FALLBACK" == "true" ]]; then
        echo ""
        echo "Note: Classification is ambiguous — consider LLM fallback."
    fi
fi

exit 0
