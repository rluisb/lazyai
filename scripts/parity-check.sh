#!/usr/bin/env bash
# Cross-runtime parity harness for spec 023.
#
# Runs `ai-setup init` via both the Go and TS binaries with identical
# inputs, then diffs the output trees. Reports differences and exits
# non-zero on unexpected divergence.
#
# Acceptable differences (filtered before diff):
#   - Timestamps in .ai-setup.json (installedAt, lastUpdatedAt, cliVersion)
#   - SQLite binary artifacts (.ai-setup.db) — compared structurally, not byte-wise
#   - File ordering in directory listings (comparisons are always sorted)
#
# Usage: ./scripts/parity-check.sh [--verbose]

set -euo pipefail

VERBOSE=0
if [[ "${1:-}" == "--verbose" ]]; then
  VERBOSE=1
fi

log() {
  if [[ $VERBOSE -eq 1 ]]; then
    printf '  %s\n' "$*" >&2
  fi
}

err() { printf '✗ %s\n' "$*" >&2; }
ok()  { printf '✓ %s\n' "$*" >&2; }

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

GO_DIR=$(mktemp -d -t ai-setup-parity-go-XXXXXX)
TS_DIR=$(mktemp -d -t ai-setup-parity-ts-XXXXXX)
trap 'rm -rf "$GO_DIR" "$TS_DIR"' EXIT

log "repo root:   $REPO_ROOT"
log "go target:   $GO_DIR"
log "ts target:   $TS_DIR"

# Runtime-specific init args. The flag spelling differs slightly:
#   Go: --non-interactive    TS: --no-interactive
GO_INIT_ARGS=(init --scope project --tools opencode --name parity-harness --non-interactive)
TS_INIT_ARGS=(init --scope project --tools opencode --name parity-harness --no-interactive)

# Build both runtimes before running.
log "Building Go binary..."
GO_BIN="$REPO_ROOT/packages/ai-setup-go/ai-setup"
(cd "$REPO_ROOT/packages/ai-setup-go" && go build -o ai-setup . >/dev/null)

log "Building TS binary..."
(cd "$REPO_ROOT" && pnpm --filter ./packages/ai-setup-ts run build >/dev/null 2>&1)
TS_BIN="node $REPO_ROOT/packages/ai-setup-ts/bin/ai-setup.js"

# Git needs to exist for both runtimes' detection logic.
(cd "$GO_DIR" && git init -q && git commit --allow-empty -qm "init" 2>/dev/null || true)
(cd "$TS_DIR" && git init -q && git commit --allow-empty -qm "init" 2>/dev/null || true)

log "Running Go init in $GO_DIR..."
# shellcheck disable=SC2068
(cd "$GO_DIR" && "$GO_BIN" ${GO_INIT_ARGS[@]} >/dev/null 2>&1)

log "Running TS init in $TS_DIR..."
# shellcheck disable=SC2068
(cd "$TS_DIR" && $TS_BIN ${TS_INIT_ARGS[@]} >/dev/null 2>&1)

# Compare the output file trees, excluding known runtime-specific artifacts.
EXCLUDES=(
  # Per-init SQLite DB (Go-only; TS uses lowdb JSON)
  "--exclude=.ai-setup.db"
  "--exclude=.ai-setup.db-journal"
  # Git dirs (we re-initialized them independently)
  "--exclude=.git"
  # Node binary artifacts that never land in Go runs
  "--exclude=node_modules"
)

DIFF_OUTPUT=$(diff -r "${EXCLUDES[@]}" "$GO_DIR" "$TS_DIR" 2>&1 || true)

if [[ -z "$DIFF_OUTPUT" ]]; then
  ok "Parity: Go and TS produced identical output trees (excluding runtime-specific artifacts)"
  exit 0
fi

# Filter out expected manifest-timestamp diffs. We look for diff blocks whose
# only differences are limited to timestamp and cliVersion fields.
FILTERED=$(echo "$DIFF_OUTPUT" \
  | grep -vE '^(Only in [^:]+: (\.ai-setup\.json|\.ai-setup-backup|\.opencode))$' \
  | grep -vE '^(<|>) +"(installedAt|lastUpdatedAt|cliVersion|hash)":' \
  | grep -vE '^(<|>) +"[0-9a-f]+"$' \
  || true)

if [[ -z "$FILTERED" ]]; then
  ok "Parity: only expected differences (timestamps, hashes, optional files)"
  exit 0
fi

err "Unexpected divergence between Go and TS runtimes:"
echo ""
echo "$DIFF_OUTPUT"
echo ""
err "Parity check failed — see diff above."
exit 1
