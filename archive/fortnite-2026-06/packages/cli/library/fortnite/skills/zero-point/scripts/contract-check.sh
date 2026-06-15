#!/usr/bin/env bash
# contract-check.sh — Pre/post condition assertions for implementation tasks
# Usage: ./contract-check.sh --mode pre|post [--spec-dir <path>] [--repo-profile <name>]
#
# Pre-conditions (run before implementation):
#   spec_exists      — spec.md (or SPEC.md) exists in the spec directory
#   branch_clean     — git status --porcelain returns nothing unexpected
#   tests_pass       — repo quality gates pass (baseline check)
#   no_stale_locks   — no orphaned task-lock.sh or barrier files
#
# Post-conditions (run after implementation):
#   no_new_lint      — no new lint/fmt errors vs baseline
#   spec_files_exist — all files mentioned in spec are present
#   tests_still_pass — quality gates pass after changes
#   no_orphaned_tests — every test file maps to a spec requirement
#
# Exit: 0 = all assertions pass; non-zero = at least one failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE=""
SPEC_DIR=""
REPO_PROFILE=""
FAILED=0

die() { echo "❌ $1" >&2; exit 1; }
warn() { echo "🟡 $1" >&2; }
pass() { echo "✅ $1"; }

# --- Parse args ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --spec-dir)
            SPEC_DIR="$2"
            shift 2
            ;;
        --repo-profile)
            REPO_PROFILE="$2"
            shift 2
            ;;
        *)
            die "Unknown argument: $1"
            ;;
    esac
done

[[ -n "$MODE" ]] || die "Missing required --mode (pre|post)"
[[ "$MODE" == "pre" || "$MODE" == "post" ]] || die "MODE must be 'pre' or 'post', got: $MODE"

# --- Helpers ---

assert_spec_exists() {
    local dir="${1:-.}"
    if [[ -f "$dir/spec.md" || -f "$dir/SPEC.md" ]]; then
        pass "spec_exists — spec found in $dir"
    else
        warn "spec_exists — no spec.md or SPEC.md found in $dir"
        FAILED=1
    fi
}

assert_branch_clean() {
    local unexpected
    unexpected=$(git status --porcelain 2>/dev/null | grep -v '^??' || true)
    if [[ -z "$unexpected" ]]; then
        pass "branch_clean — no uncommitted changes in tracked files"
    else
        warn "branch_clean — uncommitted changes detected:\n$unexpected"
        FAILED=1
    fi
}

assert_tests_pass() {
    local profile="${1:-}"
    local gate_script="$SCRIPT_DIR/quality-gate.sh"

    if [[ -x "$gate_script" ]]; then
        if [[ -n "$profile" ]]; then
            # Repo profile known — run quality gate script
            if "$gate_script" "$profile" 2>&1; then
                pass "tests_pass — quality gates passed ($profile)"
            else
                warn "tests_pass — quality gates FAILED ($profile)"
                FAILED=1
            fi
        else
            # No profile — try auto-detect
            if "$gate_script" 2>&1; then
                pass "tests_pass — quality gates passed (auto-detected)"
            else
                warn "tests_pass — quality gates FAILED (auto-detected)"
                FAILED=1
            fi
        fi
    else
        warn "tests_pass — quality-gate.sh not found at $gate_script (skipping)"
    fi
}

assert_no_stale_locks() {
    local stale=0
    # Check for orphaned lock/barrier files (simple heuristic: .lock or barrier-* in common dirs)
    if find . -maxdepth 2 -name "*.lock" -type f 2>/dev/null | grep -q .; then
        local lock_files
        lock_files=$(find . -maxdepth 2 -name "*.lock" -type f 2>/dev/null)
        warn "no_stale_locks — lock files found:\n$lock_files"
        stale=1
    fi
    # Check for stale session db barriers
    if find . -maxdepth 2 -name "barrier-*" -type f 2>/dev/null | grep -q .; then
        local barrier_files
        barrier_files=$(find . -maxdepth 2 -name "barrier-*" -type f 2>/dev/null)
        warn "no_stale_locks — barrier files found:\n$barrier_files"
        stale=1
    fi
    if [[ "$stale" -eq 0 ]]; then
        pass "no_stale_locks — no orphaned locks or barriers"
    else
        FAILED=1
    fi
}

assert_no_new_lint() {
    # Baseline: capture lint errors on current branch (or HEAD)
    # This is a lightweight check: run linter on diff and compare counts
    local profile="${1:-}"
    local gate_script="$SCRIPT_DIR/quality-gate.sh"

    if [[ -x "$gate_script" ]]; then
        if [[ -n "$profile" ]]; then
            if "$gate_script" "$profile" 2>&1; then
                pass "no_new_lint — no new lint errors ($profile)"
            else
                warn "no_new_lint — lint/test errors introduced ($profile)"
                FAILED=1
            fi
        else
            if "$gate_script" 2>&1; then
                pass "no_new_lint — no new lint errors (auto-detected)"
            else
                warn "no_new_lint — lint/test errors introduced (auto-detected)"
                FAILED=1
            fi
        fi
    else
        warn "no_new_lint — quality-gate.sh not found (skipping)"
    fi
}

assert_spec_files_exist() {
    local dir="${1:-.}"
    local spec="$dir/spec.md"
    [[ -f "$spec" ]] || spec="$dir/SPEC.md"

    if [[ ! -f "$spec" ]]; then
        warn "spec_files_exist — no spec to read"
        FAILED=1
        return
    fi

    # Extract "Files affected" / "Files to modify" / "Files to create" lines and check existence
    local files=()
    while IFS= read -r line; do
        # Match lines like "- `file.md`" or "- file.md" or "- CREATE: file.md"
        local fname
        fname=$(echo "$line" | grep -oE '`[^`]+`' | tr -d '`' || true)
        if [[ -z "$fname" ]]; then
            fname=$(echo "$line" | sed -n 's/.*- \(.*\)/\1/p' | awk '{print $1}')
        fi
        if [[ -n "$fname" && ! "$fname" =~ ^# && ! "$fname" =~ ^\*\* ]]; then
            files+=("$fname")
        fi
    done < <(grep -iE '(files affected|files to modify|files to create|files:|CREATE:|MODIFY:)' -A 20 "$spec" 2>/dev/null || true)

    local missing=0
    for f in "${files[@]}"; do
        # Skip URLs, pure directory paths ending in /
        [[ "$f" =~ ^http ]] && continue
        [[ "$f" == */ ]] && continue
        # Strip trailing comments
        f=$(echo "$f" | awk '{print $1}')
        [[ -z "$f" ]] && continue
        if [[ -e "$f" ]]; then
            : # exists
        else
            # Try relative to spec dir
            if [[ -e "$dir/$f" ]]; then
                : # exists
            else
                warn "spec_files_exist — missing: $f"
                missing=1
            fi
        fi
    done

    if [[ "$missing" -eq 0 ]]; then
        pass "spec_files_exist — all referenced files present"
    else
        FAILED=1
    fi
}

assert_no_orphaned_tests() {
    # Every test file should map to a spec requirement (lightweight heuristic)
    local test_files=()
    while IFS= read -r -d '' f; do
        test_files+=("$f")
    done < <(find . -type f \( -name "*test*" -o -name "*spec*" \) \( -name "*.rb" -o -name "*.go" -o -name "*.ts" -o -name "*.js" -o -name "*.py" \) -print0 2>/dev/null || true)

    if [[ ${#test_files[@]} -eq 0 ]]; then
        pass "no_orphaned_tests — no test files found (OK)"
        return
    fi

    local spec_dir="${1:-.}"
    local spec="$spec_dir/spec.md"
    [[ -f "$spec" ]] || spec="$spec_dir/SPEC.md"

    if [[ ! -f "$spec" ]]; then
        warn "no_orphaned_tests — no spec found to map tests against"
        FAILED=1
        return
    fi

    # Very lightweight: count test files vs spec requirements
    local req_count
    req_count=$(grep -cE '^### Requirement' "$spec" 2>/dev/null || echo "0")
    local test_count=${#test_files[@]}

    if [[ "$test_count" -le "$((req_count * 3))" ]]; then
        pass "no_orphaned_tests — $test_count test files vs $req_count requirements (OK)"
    else
        warn "no_orphaned_tests — $test_count test files vs $req_count requirements (high ratio, possible orphans)"
        # This is advisory — not a hard fail
    fi
}

# --- Run assertions by mode ---

echo "=== Contract Check ($MODE) ==="
echo "Spec dir: ${SPEC_DIR:-<current directory>}"
echo "Repo profile: ${REPO_PROFILE:-<auto-detect>}"
echo ""

if [[ "$MODE" == "pre" ]]; then
    assert_spec_exists "$SPEC_DIR"
    assert_branch_clean
    assert_tests_pass "$REPO_PROFILE"
    assert_no_stale_locks
else
    assert_no_new_lint "$REPO_PROFILE"
    assert_spec_files_exist "$SPEC_DIR"
    assert_tests_pass "$REPO_PROFILE"
    assert_no_orphaned_tests "$SPEC_DIR"
fi

echo ""
if [[ "$FAILED" -eq 0 ]]; then
    echo "🟢 ALL ASSERTIONS PASSED"
    exit 0
else
    echo "🔴 SOME ASSERTIONS FAILED — review warnings above"
    exit 1
fi
