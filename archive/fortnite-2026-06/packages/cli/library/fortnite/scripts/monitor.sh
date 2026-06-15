#!/usr/bin/env bash
# monitor.sh — OpenCode monitoring dashboard
# Usage: ./monitor.sh [--days N] [--json] [--live]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPENCODE_DIR="$(dirname "$SCRIPT_DIR")"
DB="$OPENCODE_DIR/.specify/session.db"
DAYS="${DAYS:-7}"
JSON_OUTPUT=false
LIVE_MODE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --days) DAYS="$2"; shift 2 ;;
        --json) JSON_OUTPUT=true; shift ;;
        --live) LIVE_MODE=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Colors
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Ensure DB exists
if [[ ! -f "$DB" ]]; then
    echo "Error: Session database not found at $DB"
    exit 1
fi

DATE_CUTOFF=$(date -v-${DAYS}d "+%Y-%m-%d" 2>/dev/null || date -d "-${DAYS} days" "+%Y-%m-%d" 2>/dev/null || echo "2020-01-01")

# ============================================================
# 1. Session Overview
# ============================================================
session_overview() {
    local total_sessions active_sessions completed_sessions failed_sessions

    total_sessions=$(sqlite3 "$DB" "SELECT COUNT(*) FROM sessions WHERE started_at >= '$DATE_CUTOFF';")
    active_sessions=$(sqlite3 "$DB" "SELECT COUNT(*) FROM sessions WHERE started_at >= '$DATE_CUTOFF' AND status = 'active';")
    completed_sessions=$(sqlite3 "$DB" "SELECT COUNT(*) FROM sessions WHERE started_at >= '$DATE_CUTOFF' AND status = 'completed';")
    failed_sessions=$(sqlite3 "$DB" "SELECT COUNT(*) FROM sessions WHERE started_at >= '$DATE_CUTOFF' AND status = 'failed';")

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"sessions\": {\"total\": $total_sessions, \"active\": $active_sessions, \"completed\": $completed_sessions, \"failed\": $failed_sessions},"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  SESSION OVERVIEW (last %d days)${NC}\n" "$DAYS"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  Total sessions:     %d\n" "$total_sessions"
        printf "  ${GREEN}Active:            %d${NC}\n" "$active_sessions"
        printf "  ${GREEN}Completed:         %d${NC}\n" "$completed_sessions"
        printf "  ${RED}Failed:            %d${NC}\n" "$failed_sessions"
        echo ""
    fi
}

# ============================================================
# 2. Dispatch Statistics
# ============================================================
dispatch_stats() {
    local total_dispatches success_count failed_count pending_count success_rate

    total_dispatches=$(sqlite3 "$DB" "SELECT COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF';")
    success_count=$(sqlite3 "$DB" "SELECT COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'success';")
    failed_count=$(sqlite3 "$DB" "SELECT COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'failed';")
    pending_count=$(sqlite3 "$DB" "SELECT COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'pending';")

    if [[ $total_dispatches -gt 0 ]]; then
        success_rate=$(echo "scale=1; $success_count * 100 / $total_dispatches" | bc 2>/dev/null || echo "N/A")
    else
        success_rate="N/A"
    fi

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"dispatches\": {\"total\": $total_dispatches, \"success\": $success_count, \"failed\": $failed_count, \"pending\": $pending_count, \"success_rate\": \"$success_rate\"},"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  DISPATCH STATISTICS${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  Total dispatches:   %d\n" "$total_dispatches"
        printf "  ${GREEN}Success:           %d${NC}\n" "$success_count"
        printf "  ${RED}Failed:            %d${NC}\n" "$failed_count"
        printf "  ${YELLOW}Pending:           %d${NC}\n" "$pending_count"
        printf "  Success rate:       %s%%\n" "$success_rate"
        echo ""
    fi
}

# ============================================================
# 3. Token Usage
# ============================================================
token_usage() {
    local total_tokens avg_tokens_per_session avg_tokens_per_dispatch max_session_tokens

    total_tokens=$(sqlite3 "$DB" "SELECT COALESCE(SUM(token_total), 0) FROM sessions WHERE started_at >= '$DATE_CUTOFF';")
    avg_tokens_per_session=$(sqlite3 "$DB" "SELECT COALESCE(AVG(token_total), 0) FROM sessions WHERE started_at >= '$DATE_CUTOFF';")
    avg_tokens_per_dispatch=$(sqlite3 "$DB" "SELECT COALESCE(AVG(token_used), 0) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF';")
    max_session_tokens=$(sqlite3 "$DB" "SELECT COALESCE(MAX(token_total), 0) FROM sessions WHERE started_at >= '$DATE_CUTOFF';")

    # Format large numbers
    format_num() {
        local num=$1
        if [[ $num -ge 1000000 ]]; then
            echo "$(echo "scale=1; $num / 1000000" | bc)M"
        elif [[ $num -ge 1000 ]]; then
            echo "$(echo "scale=1; $num / 1000" | bc)K"
        else
            echo "$num"
        fi
    }

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"tokens\": {\"total\": $total_tokens, \"avg_per_session\": $avg_tokens_per_session, \"avg_per_dispatch\": $avg_tokens_per_dispatch, \"max_session\": $max_session_tokens},"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  TOKEN USAGE${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  Total tokens:       %s\n" "$(format_num $total_tokens)"
        printf "  Avg per session:    %s\n" "$(format_num $avg_tokens_per_session)"
        printf "  Avg per dispatch:   %s\n" "$(format_num $avg_tokens_per_dispatch)"
        printf "  Max session:        %s\n" "$(format_num $max_session_tokens)"
        echo ""
    fi
}

# ============================================================
# 4. Model Usage
# ============================================================
model_usage() {
    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"models\": ["
        local first=true
        while IFS='|' read -r model count; do
            [[ "$first" == true ]] && first=false || echo ","
            printf '    {"model": "%s", "dispatches": %s}' "$model" "$count"
        done < <(sqlite3 "$DB" "SELECT d.model, COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' GROUP BY d.model ORDER BY COUNT(*) DESC;")
        echo ""
        echo "  ],"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  MODEL USAGE${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        while IFS='|' read -r model count; do
            printf "  %-30s %d dispatches\n" "$model" "$count"
        done < <(sqlite3 "$DB" "SELECT d.model, COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' GROUP BY d.model ORDER BY COUNT(*) DESC;")
        echo ""
    fi
}

# ============================================================
# 5. Agent Performance
# ============================================================
agent_performance() {
    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"agents\": ["
        local first=true
        while IFS='|' read -r agent total success failed; do
            [[ "$first" == true ]] && first=false || echo ","
            printf '    {"agent": "%s", "total": %s, "success": %s, "failed": %s}' "$agent" "$total" "$success" "$failed"
        done < <(sqlite3 "$DB" "
            SELECT d.agent,
                   COUNT(*) as total,
                   SUM(CASE WHEN d.result = 'success' THEN 1 ELSE 0 END) as success,
                   SUM(CASE WHEN d.result = 'failed' THEN 1 ELSE 0 END) as failed
            FROM dispatches d
            JOIN sessions s ON d.session_id = s.id
            WHERE s.started_at >= '$DATE_CUTOFF'
            GROUP BY d.agent
            ORDER BY total DESC;
        ")
        echo ""
        echo "  ],"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  AGENT PERFORMANCE${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  %-20s %8s %8s %8s %10s\n" "Agent" "Total" "Success" "Failed" "Rate"
        printf "  %-20s %8s %8s %8s %10s\n" "--------------------" "--------" "--------" "--------" "----------"
        while IFS='|' read -r agent total success failed; do
            if [[ $total -gt 0 ]]; then
                rate=$(echo "scale=0; $success * 100 / $total" | bc 2>/dev/null || echo "?")
            else
                rate="?"
            fi
            printf "  %-20s %8d %8d %8d %9d%%\n" "$agent" "$total" "$success" "$failed" "$rate"
        done < <(sqlite3 "$DB" "
            SELECT d.agent,
                   COUNT(*) as total,
                   SUM(CASE WHEN d.result = 'success' THEN 1 ELSE 0 END) as success,
                   SUM(CASE WHEN d.result = 'failed' THEN 1 ELSE 0 END) as failed
            FROM dispatches d
            JOIN sessions s ON d.session_id = s.id
            WHERE s.started_at >= '$DATE_CUTOFF'
            GROUP BY d.agent
            ORDER BY total DESC;
        ")
        echo ""
    fi
}

# ============================================================
# 6. Recent Errors
# ============================================================
recent_errors() {
    local error_count
    error_count=$(sqlite3 "$DB" "SELECT COUNT(*) FROM dispatches d JOIN sessions s ON d.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'failed';")

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"recent_errors\": ["
        local first=true
        while IFS='|' read -r session agent task error; do
            [[ "$first" == true ]] && first=false || echo ","
            # Escape quotes for JSON
            task=$(echo "$task" | sed 's/"/\\"/g' | head -c 100)
            error=$(echo "$error" | sed 's/"/\\"/g' | head -c 200)
            printf '    {"session": "%s", "agent": "%s", "task": "%s", "error": "%s"}' "$session" "$agent" "$task" "$error"
        done < <(sqlite3 "$DB" "
            SELECT s.id, d.agent, d.task, d.error_message
            FROM dispatches d
            JOIN sessions s ON d.session_id = s.id
            WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'failed'
            ORDER BY d.started_at DESC
            LIMIT 10;
        ")
        echo ""
        echo "  ]"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  RECENT ERRORS (%d total)${NC}\n" "$error_count"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        if [[ $error_count -gt 0 ]]; then
            while IFS='|' read -r session agent task error; do
                printf "${RED}✗${NC} [%s] %s\n" "$agent" "${task:0:60}"
                printf "  Error: %s\n\n" "${error:0:120}"
            done < <(sqlite3 "$DB" "
                SELECT s.id, d.agent, d.task, d.error_message
                FROM dispatches d
                JOIN sessions s ON d.session_id = s.id
                WHERE s.started_at >= '$DATE_CUTOFF' AND d.result = 'failed'
                ORDER BY d.started_at DESC
                LIMIT 10;
            ")
        else
            printf "${GREEN}✓ No errors in the last %d days${NC}\n" "$DAYS"
        fi
        echo ""
    fi
}

# ============================================================
# 7. Live Mode — Real-time MCP health
# ============================================================
live_mcp_health() {
    if [[ "$LIVE_MODE" == false ]]; then
        return
    fi

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"mcp_health\": ["
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  LIVE MCP SERVER HEALTH${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
    fi

    declare -a MCP_SERVERS=("morph-mcp" "codegraph" "qmd" "memory" "filesystem" "playwright")
    declare -a MCP_PROCS=("morphmcp" "codegraph" "qmd" "mcp-server-memory" "mcp-server-filesystem" "playwright-mcp")

    local first=true
    for i in "${!MCP_SERVERS[@]}"; do
        local name="${MCP_SERVERS[$i]}"
        local proc="${MCP_PROCS[$i]}"
        local running=false

        if pgrep -f "$proc" > /dev/null 2>&1; then
            running=true
        fi

        if [[ "$JSON_OUTPUT" == true ]]; then
            [[ "$first" == true ]] && first=false || echo ","
            printf '    {"server": "%s", "running": %s}' "$name" "$running"
        else
            if [[ "$running" == true ]]; then
                printf "${GREEN}●${NC} %-20s running\n" "$name"
            else
                printf "${RED}○${NC} %-20s stopped\n" "$name"
            fi
        fi
    done

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo ""
        echo "  ]"
    else
        echo ""
    fi
}

# ============================================================
# 8. Parallel Task Stats
# ============================================================
parallel_stats() {
    local total_waves completed_waves active_waves

    total_waves=$(sqlite3 "$DB" "SELECT COUNT(DISTINCT wave_id) FROM parallel_tasks pt JOIN sessions s ON pt.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF';")
    completed_waves=$(sqlite3 "$DB" "SELECT COUNT(DISTINCT pt.wave_id) FROM parallel_tasks pt JOIN sessions s ON pt.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND pt.status = 'completed';")
    active_waves=$(sqlite3 "$DB" "SELECT COUNT(DISTINCT pt.wave_id) FROM parallel_tasks pt JOIN sessions s ON pt.session_id = s.id WHERE s.started_at >= '$DATE_CUTOFF' AND pt.status IN ('pending', 'running');")

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "\"parallel\": {\"total_waves\": $total_waves, \"completed\": $completed_waves, \"active\": $active_waves}"
    else
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  PARALLEL EXECUTION${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  Total waves:        %d\n" "$total_waves"
        printf "  ${GREEN}Completed:         %d${NC}\n" "$completed_waves"
        printf "  ${YELLOW}Active:            %d${NC}\n" "$active_waves"
        echo ""
    fi
}

# ============================================================
# Main
# ============================================================
main() {
    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "{"
        echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
        echo "  \"period_days\": $DAYS,"
    fi

    session_overview
    dispatch_stats
    token_usage
    model_usage
    agent_performance
    parallel_stats
    recent_errors
    live_mcp_health

    if [[ "$JSON_OUTPUT" == true ]]; then
        echo "}"
    fi
}

main "$@"
