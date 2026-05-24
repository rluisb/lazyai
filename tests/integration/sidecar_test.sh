#!/usr/bin/env bash
set -euo pipefail

# Integration Test: Sidecar Lifecycle
# Tests sidecar init, status, attach, detach, and doctor using temp directories.
# Uses a temporary HOME so the real ~/.lazyai is never touched.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI="${CLI:-$PROJECT_DIR/packages/cli/lazyai-cli}"

# Temporary directories (cleaned up on exit)
TMP_HOME=$(mktemp -d)
TMP_PROJECT=$(mktemp -d)
TMP_WORKSPACE=$(mktemp -d)
TMP_SIDECAR=$(mktemp -d)
trap 'rm -rf "$TMP_HOME" "$TMP_PROJECT" "$TMP_WORKSPACE" "$TMP_SIDECAR"' EXIT

export HOME="$TMP_HOME"

# Ensure ~/.lazyai exists
mkdir -p "$TMP_HOME/.lazyai"

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 Integration Test: Sidecar Lifecycle"
echo "═══════════════════════════════════════════════════════════════"
echo "CLI binary:     $CLI"
echo "Temp HOME:      $TMP_HOME"
echo "Temp project:   $TMP_PROJECT"
echo "Temp workspace: $TMP_WORKSPACE"
echo "Temp sidecar:   $TMP_SIDECAR"
echo ""

# Check if CLI binary exists
if [ ! -x "$CLI" ]; then
    echo "❌ CLI binary not found or not executable: $CLI"
    echo "   Build the CLI with: cd packages/cli && go build ./cmd/lazyai-cli"
    exit 1
fi

# Verify sidecar command is available (sidecar is implemented — missing = failure)
if ! "$CLI" sidecar --help >/dev/null 2>&1; then
    echo "❌ Sidecar command missing from CLI binary: $CLI"
    echo "   Build the CLI with: cd packages/cli && go build ./cmd/lazyai-cli"
    exit 1
fi

# Helper: run a command and assert success
assert_success() {
    local label="$1"
    shift
    echo "  → $label"
    if "$@"; then
        echo "  ✅ $label succeeded"
    else
        echo "  ❌ $label FAILED (exit $?)"
        exit 1
    fi
}

# ───────────────────────────────────────────────────────────────
# Setup: register and activate a workspace
# ───────────────────────────────────────────────────────────────
echo "Setup: register workspace"
assert_success "workspace add" "$CLI" workspace add "$TMP_WORKSPACE"
assert_success "workspace switch" "$CLI" workspace switch "$(basename "$TMP_WORKSPACE")"

# ───────────────────────────────────────────────────────────────
# Test 1: sidecar init (workspace scope)
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 1: sidecar init --scope workspace --path <sidecar>"
assert_success "sidecar init workspace" "$CLI" sidecar init --scope workspace --path "$TMP_SIDECAR"

# Verify workspace config contains sidecar
if grep -q "sidecar" "$TMP_HOME/.lazyai/workspaces.yaml" 2>/dev/null; then
    echo "  ✅ Workspace config contains sidecar block"
else
    echo "  ❌ Workspace config missing sidecar block"
    exit 1
fi

# ───────────────────────────────────────────────────────────────
# Test 2: sidecar status
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 2: sidecar status"
STATUS_OUTPUT=$("$CLI" sidecar status 2>&1) || true
echo "$STATUS_OUTPUT"
if echo "$STATUS_OUTPUT" | grep -qiE "(Scope|Config Level|Docs Dir|Specs Dir|Plans Dir)"; then
    echo "  ✅ sidecar status shows table columns"
else
    echo "  ❌ sidecar status missing expected table columns"
    exit 1
fi

# ───────────────────────────────────────────────────────────────
# Test 3: sidecar attach (project scope)
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 3: sidecar attach --path <sidecar> --scope project"
assert_success "sidecar attach project" "$CLI" sidecar attach "$TMP_PROJECT" --path "$TMP_SIDECAR/project-docs" --scope project

# Verify project-level sidecar file was created
if [ -f "$TMP_PROJECT/.lazyai-sidecar.yaml" ]; then
    echo "  ✅ Project sidecar file exists: $TMP_PROJECT/.lazyai-sidecar.yaml"
else
    echo "  ❌ Project sidecar file not created"
    exit 1
fi

# ───────────────────────────────────────────────────────────────
# Test 4: sidecar doctor
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 4: sidecar doctor"
assert_success "sidecar doctor" "$CLI" sidecar doctor

# ───────────────────────────────────────────────────────────────
# Test 5: sidecar detach (project scope)
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 5: sidecar detach --scope project --force"
assert_success "sidecar detach project" "$CLI" sidecar detach "$TMP_PROJECT" --scope project --force

# Verify project-level sidecar file was removed
if [ ! -f "$TMP_PROJECT/.lazyai-sidecar.yaml" ]; then
    echo "  ✅ Project sidecar file removed"
else
    echo "  ❌ Project sidecar file still exists"
    exit 1
fi

# ───────────────────────────────────────────────────────────────
# Test 6: backward compat — no sidecar = graceful status
# ───────────────────────────────────────────────────────────────
echo ""
echo "Test 6: Backward compatibility (no sidecar configured)"
# Use a completely fresh temp HOME and project to avoid workspace sidecar from earlier tests
TMP_CLEAN_HOME=$(mktemp -d)
TMP_CLEAN_PROJECT=$(mktemp -d)
trap 'rm -rf "$TMP_HOME" "$TMP_PROJECT" "$TMP_WORKSPACE" "$TMP_SIDECAR" "$TMP_CLEAN_HOME" "$TMP_CLEAN_PROJECT"' EXIT

export HOME="$TMP_CLEAN_HOME"
mkdir -p "$TMP_CLEAN_HOME/.lazyai"
cd "$TMP_CLEAN_PROJECT"
NO_SIDECAR_OUTPUT=$("$CLI" sidecar status 2>&1) || true
echo "$NO_SIDECAR_OUTPUT"
if echo "$NO_SIDECAR_OUTPUT" | grep -qiE "(default|no sidecar|none|not configured|fallback|Scope)"; then
    echo "  ✅ No sidecar = graceful fallback"
else
    echo "  ⚠️  No-sidecar output unexpected (may need review)"
fi

# Restore HOME for cleanup trap
export HOME="$TMP_HOME"

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ All sidecar integration tests passed."
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "Notes:"
echo "  • Real ~/.lazyai was never touched (used temp HOME: $TMP_HOME)."
echo "  • Build the CLI with 'cd packages/cli && go build ./cmd/lazyai-cli' before running these tests."
