#!/usr/bin/env bash
# task-lock.sh — Exclusive locks for parallel agent tasks
# Prevents race conditions when multiple agents write to shared resources
# Usage: ./task-lock.sh <command> [args...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SESSION_DB="$SCRIPT_DIR/session-db.sh"
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
    acquire)
        SID="${1:-$(detect_session)}"
        LNAME="${2:-}"
        HOLDER="${3:-}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$LNAME" ] && { echo "Usage: task-lock.sh acquire <lock-name> <holder> [session-id]"; exit 1; }
        "$SESSION_DB" lock-acquire "$SID" "$LNAME" "$HOLDER"
        ;;

    release)
        SID="${1:-$(detect_session)}"
        LNAME="${2:-}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$LNAME" ] && { echo "Usage: task-lock.sh release <lock-name> [session-id]"; exit 1; }
        "$SESSION_DB" lock-release "$SID" "$LNAME"
        ;;

    try)
        SID="${1:-$(detect_session)}"
        LNAME="${2:-}"
        HOLDER="${3:-}"
        TIMEOUT="${4:-10}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$LNAME" ] && { echo "Usage: task-lock.sh try <lock-name> <holder> [timeout] [session-id]"; exit 1; }
        ELAPSED=0
        while [ $ELAPSED -lt $TIMEOUT ]; do
            if "$SESSION_DB" lock-acquire "$SID" "$LNAME" "$HOLDER" 2>/dev/null; then
                exit 0
            fi
            sleep 1
            ELAPSED=$((ELAPSED + 1))
        done
        echo "⏰ Lock '$LNAME' timeout after ${TIMEOUT}s"
        exit 1
        ;;

    status)
        SID="${1:-$(detect_session)}"
        "$SESSION_DB" lock-status "$SID"
        ;;

    list)
        "$SESSION_DB" lock-status
        ;;

    stale-cleanup)
        SID="${1:-$(detect_session)}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM locks WHERE session_id='$SID' AND status='active' AND acquired_at < datetime('now', '-30 minutes');")
        if [ "$COUNT" -gt 0 ]; then
            sqlite3 "$DB_PATH" "UPDATE locks SET status='released', released_at=datetime('now') WHERE session_id='$SID' AND status='active' AND acquired_at < datetime('now', '-30 minutes');"
            echo "🧹 Cleaned up $COUNT stale locks"
        else
            echo "✅ No stale locks found"
        fi
        ;;

    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║          🔒 TASK LOCK — Exclusive Resource Locks         ║
╚══════════════════════════════════════════════════════════╝

Commands:
  task-lock.sh acquire <lock-name> <holder> [sid]    Acquire exclusive lock
  task-lock.sh release <lock-name> [sid]             Release lock
  task-lock.sh try <lock-name> <holder> [timeout]    Try with timeout
  task-lock.sh status [sid]                          Show active locks
  task-lock.sh list                                  Show all active locks
  task-lock.sh stale-cleanup [sid]                   Release locks > 30min old

Examples:
  task-lock.sh acquire "spec-write" "turbo-crank"
  task-lock.sh release "spec-write"
  task-lock.sh try "deploy-lock" "rift-deploy" 30
HELP
        ;;
esac
