#!/usr/bin/env bash
# tests/scripts/drift-scan.sh — Banned-token drift prevention scan.
#
# Fails if any non-historical, non-generated file under packages/cli/library/
# or the maintainer-mirror .agents/ contains a banned token. The banned
# tokens are references to retired runtime surfaces that the alignment
# cleanup explicitly removed (Phase 1+2 of the 026-vibe-lab-alignment refactor).
#
# Usage:
#   tests/scripts/drift-scan.sh
#
# Exit code:
#   0  — no banned tokens found
#   1  — at least one banned token found; offending files printed to stderr
#
# Scope (text-bearing files only):
#   packages/cli/library/**
#   .agents/**
#   canonical/**
#   docs/concepts/**
#   packages/cli/internal/**
#   root + per-package KNOWLEDGE_MAP.md
#
# Excluded (justified one-liners inline in EXCLUDE_DIRS / EXCLUDE_FILES):
#   bin/doctor, bin/inject       — legacy bash scripts that still mention
#                                  vibe-lab (a separate cleanup follow-up)
#   curation.yaml                 — historical provenance comment
#   .archive/**/recommendation-summary.md  — archived historical planning document
#   specs/issues                  — historical parity audit notes
#   docs/AI-Agentic-Setup-Templates/ — historical template library

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

BANNED_TOKENS=(
  '\.vibe-lab/'
  '\.gemini/hooks/vibe-lab'
  'hooks/vibe-lab'
  'vibe-lab generated'
)

PLACEHOLDER_TOKENS=(
  '\[YOUR_APPROVED_LANGUAGES\]'
  '\[YOUR_APPROVED_FRAMEWORKS\]'
  '\[YOUR_APPROVED_RUNTIMES\]'
  '\[YOUR_APPROVED_TEST_STACKS\]'
  '\[YOUR_LINE_COVERAGE_TARGET\]'
  '\[YOUR_LINE_COVERAGE_MINIMUM\]'
  '\[YOUR_BRANCH_COVERAGE_TARGET\]'
  '\[YOUR_BRANCH_COVERAGE_MINIMUM\]'
  '\[YOUR_BUILD_TIME_TARGET\]'
  '\[YOUR_BUILD_TIME_MAXIMUM\]'
)

EXCLUDE_DIRS=(
  --exclude-dir=.git
  --exclude-dir=node_modules
  --exclude-dir=archive
  --exclude-dir=specs
  --exclude-dir=AI-Agentic-Setup-Templates
)

EXCLUDE_FILES=(
  --exclude=curation.yaml
  --exclude=doctor
  --exclude=.archive/**/recommendation-summary.md
  --exclude=drift-canary-test.md
)

INCLUDES=(
  --include=*.md
  --include=*.js
  --include=*.ts
  --include=*.tsx
  --include=*.json
  --include=*.yaml
  --include=*.yml
  --include=*.sh
  --include=*.go
)

SCOPES=(
  packages/cli/library
  .agents
  canonical
  docs/concepts
  packages/cli/internal
  AGENTS.md
  KNOWLEDGE_MAP.md
  packages/cli/KNOWLEDGE_MAP.md
  specs/KNOWLEDGE_MAP.md
)

FAIL=0
for token in "${BANNED_TOKENS[@]}"; do
  for scope in "${SCOPES[@]}"; do
    [ -e "$scope" ] || continue
    if grep -rEn \
        "${INCLUDES[@]}" \
        "${EXCLUDE_DIRS[@]}" \
        "${EXCLUDE_FILES[@]}" \
        -- "$token" "$scope" 2>/dev/null; then
      echo "drift-scan: banned token '${token}' found in scope '$scope'" >&2
      FAIL=1
    fi
  done
done

for token in "${PLACEHOLDER_TOKENS[@]}"; do
  for scope in "${SCOPES[@]}"; do
    [ -e "$scope" ] || continue
    if grep -rEn \
        "${INCLUDES[@]}" \
        "${EXCLUDE_DIRS[@]}" \
        "${EXCLUDE_FILES[@]}" \
        -- "$token" "$scope" 2>/dev/null; then
      echo "drift-scan: unresolved template placeholder '${token}' found in scope '$scope'" >&2
      FAIL=1
    fi
  done
done

if [ "$FAIL" -eq 0 ]; then
  echo "drift-scan: clean (no banned tokens in active scope)"
fi

exit "$FAIL"
