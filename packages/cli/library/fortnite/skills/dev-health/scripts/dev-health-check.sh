#!/usr/bin/env bash
# dev-health-check.sh — Full environment health assessment
# Usage: ./dev-health-check.sh [quick|full|deep]
# Default: full

set -euo pipefail

LEVEL="${1:-full}"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
WARNINGS=0
ERRORS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ok() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; ((WARNINGS++)); }
err() { echo -e "${RED}❌ $1${NC}"; ((ERRORS++)); }
info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
section() { echo ""; echo "=== $1 ==="; }

echo "## Dev Environment Health Report"
echo "Generated: $TIMESTAMP"
echo "Level: $LEVEL"

# ─── Runtime Check ───
section "Runtime"

if command -v colima &>/dev/null; then
    if colima status 2>/dev/null | grep -q "Running"; then
        CPU=$(colima list 2>/dev/null | grep -oP '\d+(?= CPUs)' || echo "?")
        RAM=$(colima list 2>/dev/null | grep -oP '\d+GiB' || echo "?")
        DISK=$(colima list 2>/dev/null | grep -oP '\d+GiB.*disk' || echo "?")
        ok "Colima running ($CPU, $RAM, $DISK)"
    else
        err "Colima is not running — run 'colima start'"
    fi
else
    warn "Colima not installed"
fi

if command -v docker &>/dev/null; then
    if docker info &>/dev/null; then
        ok "Docker responsive"
    else
        err "Docker daemon not responding"
    fi
else
    warn "Docker not installed"
fi

# ─── Services Check ───
if [ "$LEVEL" != "quick" ] || true; then
    section "Services"

    if command -v dev &>/dev/null; then
        echo "| Service | Status | Container | Worktree |"
        echo "|---------|--------|-----------|----------|"

        # Try to get service list
        if dev list --json &>/dev/null; then
            dev list --json 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    services = data if isinstance(data, list) else data.get('services', [])
    for s in services:
        name = s.get('name', '?')
        status = s.get('status', '?')
        container = s.get('container', '—')
        worktree = s.get('worktree', '—')
        icon = '✅' if status == 'running' else '⚠️'
        print(f'| {icon} {name} | {status} | {container} | {worktree} |')
except:
    print('| (could not parse service list) |')
" 2>/dev/null || warn "Could not parse dev list output"
        else
            # Fallback: try plain dev list
            dev list 2>/dev/null | tail -n +2 | while read -r line; do
                echo "| $line | — | — |"
            done || warn "dev list failed"
        fi
    else
        warn "dev CLI not installed"
    fi
fi

# ─── Worktrees ───
if [ "$LEVEL" != "quick" ]; then
    section "Worktrees"

    # Dev worktrees (fedora-worktrees/)
    if [ -d "fedora-worktrees" ]; then
        echo "**Dev worktrees (fedora-worktrees/):**"
        for wt in fedora-worktrees/*/; do
            [ -d "$wt" ] || continue
            name=$(basename "$wt")
            echo "  📁 $name"
        done
    else
        info "No fedora-worktrees/ directory"
    fi

    # Agent worktrees (.worktrees/)
    if [ -d ".worktrees" ]; then
        echo "**Agent worktrees (.worktrees/):**"
        for wt in .worktrees/*/; do
            [ -d "$wt" ] || continue
            name=$(basename "$wt")
            age=$(find "$wt" -maxdepth 0 -mtime +7 -print 2>/dev/null | wc -l | tr -d ' ')
            if [ "$age" -gt 0 ]; then
                warn "$name (stale, >7 days)"
            else
                ok "$name (active)"
            fi
        done
    else
        info "No .worktrees/ directory"
    fi

    # Manual git worktrees
    WT_COUNT=$(git worktree list 2>/dev/null | wc -l | tr -d ' ')
    if [ "$WT_COUNT" -gt 1 ]; then
        echo "**Manual git worktrees:**"
        git worktree list 2>/dev/null | tail -n +2 | while read -r line; do
            echo "  📁 $line"
        done
    fi
fi

# ─── Containers ───
if [ "$LEVEL" != "quick" ]; then
    section "Containers"

    if docker info &>/dev/null; then
        # Running containers
        RUNNING=$(docker ps --format '{{.Names}} ({{.Status}})' 2>/dev/null | head -20)
        if [ -n "$RUNNING" ]; then
            echo "**Running:**"
            echo "$RUNNING" | while read -r line; do
                echo "  ✅ $line"
            done
        else
            info "No running containers"
        fi

        # Stopped containers (potential cleanup)
        STOPPED=$(docker ps -a --filter "status=exited" --format '{{.Names}} ({{.Status}})' 2>/dev/null | head -20)
        if [ -n "$STOPPED" ]; then
            echo "**Stopped (cleanup candidates):**"
            echo "$STOPPED" | while read -r line; do
                warn "$line"
            done
        fi
    fi
fi

# ─── Disk Usage ───
if [ "$LEVEL" = "deep" ] || [ "$LEVEL" = "full" ]; then
    section "Disk Usage"

    # Colima disk
    if command -v colima &>/dev/null; then
        colima list 2>/dev/null | grep -v "NAME" | while read -r line; do
            echo "  Colima: $line"
        done
    fi

    # Docker system info
    if docker info &>/dev/null; then
        DOCKER_DISK=$(docker system df --format '{{.Type}}: {{.Size}}' 2>/dev/null | head -5)
        if [ -n "$DOCKER_DISK" ]; then
            echo "**Docker:**"
            echo "$DOCKER_DISK" | while read -r line; do
                echo "  📦 $line"
            done
        fi
    fi

    # Worktree disk usage
    if [ -d ".worktrees" ]; then
        WT_SIZE=$(du -sh .worktrees 2>/dev/null | cut -f1)
        echo "  Agent worktrees: $WT_SIZE"
    fi
    if [ -d "fedora-worktrees" ]; then
        FW_SIZE=$(du -sh fedora-worktrees 2>/dev/null | cut -f1)
        echo "  Dev worktrees: $FW_SIZE"
    fi
fi

# ─── Deep Checks ───
if [ "$LEVEL" = "deep" ]; then
    section "Deep Checks"

    # Container logs (last 5 lines for errors)
    if docker info &>/dev/null; then
        docker ps --format '{{.Names}}' 2>/dev/null | while read -r container; do
            ERRORS=$(docker logs --tail 100 "$container" 2>&1 | grep -ci "error\|fatal\|panic" || true)
            if [ "$ERRORS" -gt 0 ]; then
                warn "$container: $ERRORS error(s) in recent logs"
            else
                ok "$container: no recent errors"
            fi
        done
    fi

    # Service health endpoints (if configured)
    if command -v dev &>/dev/null; then
        dev list --json 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    services = data if isinstance(data, list) else data.get('services', [])
    for s in services:
        name = s.get('name', '?')
        health = s.get('health_url', '')
        if health and s.get('status') == 'running':
            print(f'  Health check: {name} → {health}')
except:
    pass
" 2>/dev/null || true
    fi
fi

# ─── Summary ───
section "Summary"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ "$ERRORS" -gt 0 ]; then
    echo -e "${RED}❌ Environment has issues that need attention${NC}"
    exit 1
elif [ "$WARNINGS" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Environment is functional but has warnings${NC}"
    exit 0
else
    echo -e "${GREEN}✅ Environment is healthy${NC}"
    exit 0
fi
