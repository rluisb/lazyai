#!/usr/bin/env bash
# agent-dispatch.sh — Record agent dispatch history with results
# Produces GSD .planning/STATE.md compatible output
# Usage: ./agent-dispatch.sh record <agent> <task> <result>

set -euo pipefail

DISPATCH_LOG="${OPENCODE_WORKSPACE:-.}/.planning/dispatch-log.md"
CMD="${1:-help}"

case "$CMD" in
    record)
        AGENT="${2:-unknown}"
        TASK="${3:-unknown}"
        RESULT="${4:-pending}"
        TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

        mkdir -p "$(dirname "$DISPATCH_LOG")"

        if [ ! -f "$DISPATCH_LOG" ]; then
            cat > "$DISPATCH_LOG" << 'EOH'
# Agent Dispatch Log

| Time | Agent | Task | Result |
|------|-------|------|--------|
EOH
        fi

        echo "| $TIMESTAMP | \`$AGENT\` | $TASK | $RESULT |" >> "$DISPATCH_LOG"
        echo "Recorded: $AGENT → $TASK ($RESULT)"
        ;;

    list)
        if [ -f "$DISPATCH_LOG" ]; then
            tail -n +5 "$DISPATCH_LOG"
        else
            echo "No dispatch log found. Run 'record' first."
        fi
        ;;

    stats)
        if [ -f "$DISPATCH_LOG" ]; then
            TOTAL=$(tail -n +5 "$DISPATCH_LOG" | wc -l | tr -d ' ')
            echo "Total dispatches: $TOTAL"
            echo ""
            echo "By agent:"
            tail -n +5 "$DISPATCH_LOG" | awk -F'|' '{gsub(/[ `]/, "", $2); print $2}' | sort | uniq -c | sort -rn
            echo ""
            echo "By result:"
            tail -n +5 "$DISPATCH_LOG" | awk -F'|' '{print $4}' | sort | uniq -c | sort -rn
        else
            echo "No dispatch log found."
        fi
        ;;

    state)
        # Export current state in GSD .planning/STATE.md format
        mkdir -p "$(dirname "$DISPATCH_LOG")"
        STATE_FILE="$(dirname "$DISPATCH_LOG")/STATE.md"
        TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

        cat > "$STATE_FILE" << EOS
# State — $TIMESTAMP

## Last Dispatch
$(tail -1 "$DISPATCH_LOG" 2>/dev/null || echo "No dispatches yet")

## Active Worktrees
$(ls .worktrees/ 2>/dev/null || echo "none")

## Current Phase
$(tail -5 "$DISPATCH_LOG" 2>/dev/null | grep -oP '(?<=`)[^`]+(?=`)' | tail -1 || echo "unknown")

## Next Action
[human must define]
EOS
        echo "State exported to $STATE_FILE"
        ;;

    help|*)
        echo "Agent Dispatch Recorder"
        echo ""
        echo "Usage:"
        echo "  ./agent-dispatch.sh record <agent> <task> <result>    Record a dispatch"
        echo "  ./agent-dispatch.sh list                                Show dispatch history"
        echo "  ./agent-dispatch.sh stats                               Show dispatch statistics"
        echo "  ./agent-dispatch.sh state                               Export GSD STATE.md"
        echo ""
        echo "Log file: $DISPATCH_LOG"
        ;;
esac
