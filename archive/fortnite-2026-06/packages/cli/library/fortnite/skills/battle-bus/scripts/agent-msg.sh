#!/usr/bin/env bash
# agent-msg.sh — High-level inter-agent message bus
# Wraps session-db.sh msg-* commands with agent-friendly interface
# Usage: ./agent-msg.sh <command> [args...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# Resolve to root session-db.sh
ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SESSION_DB="$ROOT_DIR/scripts/session-db.sh"
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"

# Auto-detect session from environment or use most recent
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
    send)
        SID="${1:-$(detect_session)}"
        FROM="${2:-}"
        TO="${3:-}"
        SUBJECT="${4:-}"
        BODY="${5:-}"
        PRIORITY="${6:-normal}"
        [ -z "$SID" ] && { echo "❌ No active session. Set SESSION_ID or pass <session-id>"; exit 1; }
        [ -z "$FROM" ] && { echo "Usage: agent-msg.sh send <session-id> <from-agent> <to-agent> <subject> <body> [priority]"; exit 1; }
        "$SESSION_DB" msg-send "$SID" "$FROM" "$TO" "$SUBJECT" "$BODY" "$PRIORITY"
        ;;

    recv)
        AGENT="${1:-}"
        SID="${2:-$(detect_session)}"
        [ -z "$AGENT" ] && { echo "Usage: agent-msg.sh recv <agent> [session-id]"; exit 1; }
        echo "📬 Inbox for $AGENT:"
        echo "---"
        "$SESSION_DB" msg-recv "$AGENT" "$SID"
        ;;

    read)
        MSGID="${1:-}"
        [ -z "$MSGID" ] && { echo "Usage: agent-msg.sh read <msg-id>"; exit 1; }
        "$SESSION_DB" msg-read "$MSGID"
        ;;

    history)
        SID="${1:-$(detect_session)}"
        AGENT="${2:-}"
        "$SESSION_DB" msg-history "$SID" "$AGENT"
        ;;

    broadcast)
        SID="${1:-$(detect_session)}"
        FROM="${2:-}"
        SUBJECT="${3:-}"
        BODY="${4:-}"
        PRIORITY="${5:-normal}"
        [ -z "$SID" ] && { echo "❌ No active session."; exit 1; }
        [ -z "$FROM" ] && { echo "Usage: agent-msg.sh broadcast <session-id> <from-agent> <subject> <body> [priority]"; exit 1; }
        for AGENT in loot-hawk turbo-crank wall-builder shield-audit rift-deploy respawn-crew loop-driver; do
            "$SESSION_DB" msg-send "$SID" "$FROM" "$AGENT" "$SUBJECT" "$BODY" "$PRIORITY" 2>/dev/null || true
        done
        echo "📡 Broadcast from $FROM sent to all agents"
        ;;

    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║          📨 AGENT MESSAGE BUS — Inter-Agent Comms        ║
╚══════════════════════════════════════════════════════════╝

Commands:
   agent-msg.sh send <session-id> <from-agent> <to-agent> <subject> <body> [priority]
  agent-msg.sh recv <agent> [session-id]
  agent-msg.sh read <msg-id>
  agent-msg.sh history [session-id] [agent]
   agent-msg.sh broadcast <session-id> <from-agent> <subject> <body> [priority]

Priority levels: critical, high, normal, low

Examples:
  agent-msg.sh send <session-id> wall-builder shield-audit "Review ready" "Auth middleware implemented, ready for review" high
  agent-msg.sh recv shield-audit
  agent-msg.sh broadcast <session-id> respawn-crew "P1 Incident" "API gateway down, investigating" critical
HELP
        ;;
esac
