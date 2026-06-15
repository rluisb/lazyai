#!/usr/bin/env bash
# workflow-create.sh — Define teams and workflows at runtime
# Usage: ./workflow-create.sh <command> [args...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Resolve to workspace root (3 levels up from skills/<name>/scripts/)
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../../../../" && pwd)"
SESSION_DB="${SESSION_DB:-$WORKSPACE_ROOT/scripts/session-db.sh}"

usage() {
    cat << 'USAGE'
workflow-create.sh — Define teams and workflows at runtime

TEAMS:
  workflow-create.sh team-create <name> <agents_csv> [description]
    Examples:
      workflow-create.sh team-create "spec-team" "loot-hawk,turbo-crank,wall-builder"
      workflow-create.sh team-create "deploy-team" "wall-builder,shield-audit,rift-deploy"
      workflow-create.sh team-create "full-stack" "loot-hawk,turbo-crank,wall-builder,shield-audit"

  workflow-create.sh team-list
  workflow-create.sh team-delete <name>

WORKFLOWS:
  workflow-create.sh workflow-create <name> <team> <steps_json> [description]
    Steps JSON format:
      [{"agent":"turbo-crank","task":"Clarify requirements","mode":"clarify"},
       {"agent":"wall-builder","task":"Implement feature","mode":"standard"},
       {"agent":"shield-audit","task":"Review implementation","mode":"review"}]

    Examples:
      workflow-create.sh workflow-create "spec-flow" "spec-team" \
        '[{"agent":"turbo-crank","task":"Clarify requirements","mode":"clarify"},
         {"agent":"turbo-crank","task":"Plan architecture","mode":"plan"},
         {"agent":"wall-builder","task":"Implement","mode":"standard"},
         {"agent":"shield-audit","task":"Verify","mode":"review"}]'

      workflow-create.sh workflow-create "feature-flow" "full-stack" \
        '[{"agent":"loot-hawk","task":"Research codebase","mode":"deep"},
         {"agent":"turbo-crank","task":"Create spec","mode":"full"},
         {"agent":"wall-builder","task":"Implement","mode":"standard"},
         {"agent":"shield-audit","task":"Review","mode":"review"}]'

  workflow-create.sh workflow-list
  workflow-create.sh workflow-delete <name>

SHOW:
  workflow-create.sh show-team <name>
  workflow-create.sh show-workflow <name>
USAGE
}

cmd="${1:-help}"
shift 2>/dev/null || true

case "$cmd" in
    team-create)
        TNAME="${1:-}"; AGENTS="${2:-}"; DESC="${3:-}"
        [ -z "$TNAME" ] || [ -z "$AGENTS" ] && { echo "ERROR: name and agents required"; usage; exit 1; }
        "$SESSION_DB" team-create "$TNAME" "$AGENTS" "$DESC"
        ;;
    team-list)
        "$SESSION_DB" team-list
        ;;
    team-delete)
        TNAME="${1:-}"; [ -z "$TNAME" ] && { echo "ERROR: name required"; usage; exit 1; }
        "$SESSION_DB" team-delete "$TNAME"
        ;;
    show-team)
        TNAME="${1:-}"; [ -z "$TNAME" ] && { echo "ERROR: name required"; usage; exit 1; }
        "$SESSION_DB" query "SELECT * FROM teams WHERE name='$TNAME';"
        ;;
    workflow-create)
        WNAME="${1:-}"; TEAM="${2:-}"; STEPS="${3:-}"; DESC="${4:-}"
        [ -z "$WNAME" ] || [ -z "$TEAM" ] || [ -z "$STEPS" ] && { echo "ERROR: name, team, steps required"; usage; exit 1; }
        "$SESSION_DB" workflow-create "$WNAME" "$TEAM" "$STEPS" "$DESC"
        ;;
    workflow-list)
        "$SESSION_DB" workflow-list
        ;;
    workflow-delete)
        WNAME="${1:-}"; [ -z "$WNAME" ] && { echo "ERROR: name required"; usage; exit 1; }
        "$SESSION_DB" workflow-delete "$WNAME"
        ;;
    show-workflow)
        WNAME="${1:-}"; [ -z "$WNAME" ] && { echo "ERROR: name required"; usage; exit 1; }
        "$SESSION_DB" query "SELECT * FROM workflows WHERE name='$WNAME';"
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        echo "ERROR: unknown command '$cmd'"
        usage
        exit 1
        ;;
esac
