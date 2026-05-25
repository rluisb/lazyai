#!/usr/bin/env bash
# scripts/run-tests.sh — Deterministic test runner
# Detects test runner from repo files and executes tests without LLM.
# Usage: ./scripts/run-tests.sh [--repo PATH] [--json] [--dry-run]

set -euo pipefail
IFS=$'\n\t'

REPO_ROOT="."
OUTPUT_JSON=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo) REPO_ROOT="$2"; shift 2 ;;
        --json) OUTPUT_JSON=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        *) shift ;;
    esac
done

# Resolve absolute path
REPO_ROOT="$(cd "$REPO_ROOT" && pwd)"

# Ensure repo exists
if [[ ! -d "$REPO_ROOT" ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.test_result.v1","repo_root":"'"$REPO_ROOT"'","detected_stack":"unknown","runner":"","command":"","status":"unsupported","exit_code":1,"duration_ms":0,"stdout_log":null,"stderr_log":"Repo path does not exist"}'
    else
        echo "Error: repo path does not exist: $REPO_ROOT" >&2
    fi
    exit 1
fi

# Detection order per techspec §8
detect_stack() {
    local root="$1"
    local stack="unknown"
    local runner=""
    local cmd=""

    if [[ -f "$root/Makefile" ]]; then
        if grep -qE '^test:' "$root/Makefile" 2>/dev/null; then
            stack="make"
            runner="make"
            cmd="make test"
        fi
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/package.json" ]]; then
        local scripts
        scripts=$(cat "$root/package.json" 2>/dev/null || echo "{}")
        if echo "$scripts" | grep -q '"test"'; then
            stack="node"
            runner="npm"
            cmd="npm test"
        elif echo "$scripts" | grep -q '"test:unit"'; then
            stack="node"
            runner="npm"
            cmd="npm run test:unit"
        elif echo "$scripts" | grep -q '"vitest"'; then
            stack="node"
            runner="npm"
            cmd="npm run vitest"
        elif echo "$scripts" | grep -q '"jest"'; then
            stack="node"
            runner="npm"
            cmd="npm run jest"
        fi
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/pnpm-lock.yaml" ]]; then
        stack="node"
        runner="pnpm"
        cmd="pnpm test"
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/yarn.lock" ]]; then
        stack="node"
        runner="yarn"
        cmd="yarn test"
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/package-lock.json" ]]; then
        stack="node"
        runner="npm"
        cmd="npm test"
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/Cargo.toml" ]]; then
        stack="rust"
        runner="cargo"
        cmd="cargo test"
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/go.mod" ]]; then
        stack="go"
        runner="go"
        cmd="go test ./..."
    fi

    if [[ -z "$runner" ]] && { [[ -f "$root/build.gradle" ]] || [[ -f "$root/build.gradle.kts" ]] || [[ -f "$root/gradlew" ]]; }; then
        stack="java"
        runner="gradle"
        cmd="gradle test"
    fi

    if [[ -z "$runner" ]] && [[ -f "$root/pom.xml" ]]; then
        stack="java"
        runner="maven"
        cmd="mvn test"
    fi

    echo "$stack|$runner|$cmd"
}

DETECTION=$(detect_stack "$REPO_ROOT")
STACK=$(echo "$DETECTION" | cut -d'|' -f1)
RUNNER=$(echo "$DETECTION" | cut -d'|' -f2)
CMD=$(echo "$DETECTION" | cut -d'|' -f3)

if [[ "$DRY_RUN" == "true" ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.test_result.v1","repo_root":"'"$REPO_ROOT"'","detected_stack":"'"$STACK"'","runner":"'"$RUNNER"'","command":"'"$CMD"'","status":"dry_run","exit_code":0,"duration_ms":0,"stdout_log":null,"stderr_log":null}'
    else
        echo "[dry-run] Detected stack: $STACK"
        echo "[dry-run] Runner: $RUNNER"
        echo "[dry-run] Command: $CMD"
    fi
    exit 0
fi

if [[ -z "$RUNNER" ]]; then
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo '{"schema_version":"command.test_result.v1","repo_root":"'"$REPO_ROOT"'","detected_stack":"unknown","runner":"","command":"","status":"unsupported","exit_code":1,"duration_ms":0,"stdout_log":null,"stderr_log":"No supported test runner detected"}'
    else
        echo "No supported test runner detected in: $REPO_ROOT" >&2
    fi
    exit 1
fi

# Portable millisecond timing
if command -v python3 &>/dev/null; then
    _now_ms() { python3 -c "import time; print(int(time.time()*1000))"; }
elif date +%s%3N &>/dev/null 2>&1; then
    _now_ms() { date +%s%3N; }
else
    _now_ms() { echo "0"; }
fi

# Run tests
START_MS=$(_now_ms)
STDOUT_LOG=""
STDERR_LOG=""
EXIT_CODE=0

# Run in subshell to capture output without failing script
set +e
TMP_OUT=$(mktemp)
TMP_ERR=$(mktemp)
trap 'rm -f "$TMP_OUT" "$TMP_ERR"' EXIT

(cd "$REPO_ROOT" && bash -c "$CMD" >"$TMP_OUT" 2>"$TMP_ERR")
EXIT_CODE=$?

STDOUT_LOG=$(cat "$TMP_OUT" || true)
STDERR_LOG=$(cat "$TMP_ERR" || true)
set -e

END_MS=$(_now_ms)
DURATION_MS=$((END_MS - START_MS))

if [[ "$EXIT_CODE" -eq 0 ]]; then
    STATUS="pass"
else
    STATUS="fail"
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
    JSON_STDOUT="null"
    JSON_STDERR="null"
    if [[ -n "$STDOUT_LOG" ]]; then
        JSON_STDOUT='"'"$(escape_json "$STDOUT_LOG")"'"'
    fi
    if [[ -n "$STDERR_LOG" ]]; then
        JSON_STDERR='"'"$(escape_json "$STDERR_LOG")"'"'
    fi
    echo '{"schema_version":"command.test_result.v1","repo_root":"'"$REPO_ROOT"'","detected_stack":"'"$STACK"'","runner":"'"$RUNNER"'","command":"'"$CMD"'","status":"'"$STATUS"'","exit_code":'"$EXIT_CODE"',"duration_ms":'"$DURATION_MS"',"stdout_log":'"$JSON_STDOUT"',"stderr_log":'"$JSON_STDERR"'}'
else
    echo "Detected stack: $STACK"
    echo "Runner: $RUNNER"
    echo "Command: $CMD"
    echo "Status: $STATUS"
    echo "Exit code: $EXIT_CODE"
    echo "Duration: ${DURATION_MS}ms"
    if [[ -n "$STDOUT_LOG" ]]; then
        echo "--- stdout ---"
        echo "$STDOUT_LOG"
    fi
    if [[ -n "$STDERR_LOG" ]]; then
        echo "--- stderr ---"
        echo "$STDERR_LOG"
    fi
fi

exit "$EXIT_CODE"
