#!/usr/bin/env bash
# task-barrier.sh — Synchronization barriers for parallel agent tasks
# Agents arrive at barriers and wait until all expected agents arrive
# Usage: ./task-barrier.sh <command> [args...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SESSION_DB="$SCRIPT_DIR/../../../scripts/session-db.sh"
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"

detect_session() {
    if [ -n "${SESSION_ID:-}" ]; then
        echo "$SESSION_ID"
        return
    fi
    if [ -f "$DB_PATH" ]; then
        sqlite3 "$DB_PATH" "SELECT id FROM sessions WHERE status='active' ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || true
    fi
}

CMD="${1:-help}"
shift 2>/dev/null || true

case "$CMD" in
    create)
        SID="${1:-$(detect_session)}"
        BID="${2:-}"
        COUNT="${3:-}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$BID" ] && { echo "Usage: task-barrier.sh create <barrier-id> <expected-count> [session-id]"; exit 1; }
        "$SESSION_DB" barrier-create "$SID" "$BID" "$COUNT"
        ;;

    arrive)
        SID="${1:-$(detect_session)}"
        BID="${2:-}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$BID" ] && { echo "Usage: task-barrier.sh arrive <barrier-id> [session-id]"; exit 1; }
        "$SESSION_DB" barrier-arrive "$SID" "$BID"
        ;;

    wait)
        SID="${1:-$(detect_session)}"
        BID="${2:-}"
        TIMEOUT="${3:-30}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$BID" ] && { echo "Usage: task-barrier.sh wait <barrier-id> [timeout-seconds] [session-id]"; exit 1; }
        ELAPSED=0
        while [ $ELAPSED -lt $TIMEOUT ]; do
            STATUS=$("$SESSION_DB" barrier-status "$SID" "$BID" 2>/dev/null | grep -o 'status=[a-z]*' | cut -d= -f2 || echo "unknown")
            if [ "$STATUS" = "resolved" ]; then
                echo "🚩 Barrier $BID resolved after ${ELAPSED}s"
                exit 0
            fi
            sleep 1
            ELAPSED=$((ELAPSED + 1))
        done
        echo "⏰ Barrier $BID timeout after ${TIMEOUT}s"
        exit 1
        ;;

    status)
        SID="${1:-$(detect_session)}"
        BID="${2:-}"
        "$SESSION_DB" barrier-status "$SID" "$BID"
        ;;

    list)
        SID="${1:-$(detect_session)}"
        "$SESSION_DB" barrier-status "$SID"
        ;;

    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║          🚧 TASK BARRIER — Parallel Sync Points          ║
╚══════════════════════════════════════════════════════════╝

Commands:
  task-barrier.sh create <barrier-id> <count> [sid]   Create barrier
  task-barrier.sh arrive <barrier-id> [sid]           Arrive at barrier
  task-barrier.sh wait <barrier-id> [timeout] [sid]   Wait for barrier (polls)
  task-barrier.sh status [sid] [barrier-id]           Check barrier status
  task-barrier.sh list [sid]                          List all barriers

Examples:
  task-barrier.sh create "deploy-sync" 3
  task-barrier.sh arrive "deploy-sync"
  task-barrier.sh wait "deploy-sync" 60
HELP
        ;;
esac
