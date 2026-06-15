#!/usr/bin/env bash
# quality-gate.sh — Repo-aware quality gate runner
# Detects repo profile and runs the right checks.
# Usage: ./quality-gate.sh [repo-path]

set -euo pipefail

REPO_PATH="${1:-.}"
cd "$REPO_PATH" || exit 1

detect_profile() {
    if [ -f "Gemfile" ] && grep -q "rails" Gemfile 2>/dev/null; then
        if grep -q "creator-checkout\|creator_experience" Gemfile 2>/dev/null; then
            echo "creator-checkout"
        else
            echo "fedora"
        fi
    elif [ -f "go.mod" ]; then
        echo "school-plan-service"
    elif [ -f "package.json" ]; then
        if grep -q '"next"' package.json 2>/dev/null; then
            echo "creator-checkout"
        elif grep -q '"react"' package.json 2>/dev/null; then
            echo "mono-frontend"
        else
            echo "unknown-node"
        fi
    else
        echo "unknown"
    fi
}

run_gates() {
    local profile="$1"
    local failed=0

    echo "=== Quality Gate Runner ==="
    echo "Profile: $profile"
    echo "Repo: $REPO_PATH"
    echo ""

    case "$profile" in
        fedora)
            echo "--- RuboCop ---"
            bundle exec rubocop || failed=1
            echo ""
            echo "--- RSpec ---"
            bundle exec rspec || failed=1
            ;;
        creator-checkout)
            echo "--- Quality ---"
            npm run quality || failed=1
            echo "--- Build ---"
            npm run build || failed=1
            ;;
        "mono-frontend")
            echo "--- Lint ---"
            yarn lint || failed=1
            echo "--- TypeCheck ---"
            yarn typecheck || failed=1
            echo "--- Test ---"
            yarn test || failed=1
            echo "--- Build ---"
            yarn build || failed=1
            ;;
        "school-plan-service")
            echo "--- Go Test ---"
            go test ./... || failed=1
            echo "--- Go Vet ---"
            go vet ./... || failed=1
            ;;
        *)
            echo "ERROR: Unknown profile. Cannot determine quality gates."
            echo "Supported: fedora, creator-checkout, mono-frontend, school-plan-service"
            exit 1
            ;;
    esac

    echo ""
    if [ "$failed" -eq 0 ]; then
        echo "✅ ALL GATES PASSED"
        exit 0
    else
        echo "❌ SOME GATES FAILED"
        exit 1
    fi
}

PROFILE=$(detect_profile)
run_gates "$PROFILE"
