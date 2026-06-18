#!/usr/bin/env bash
# knowledge-inject.sh — Proactive vault context injection at session start
# Usage: knowledge-inject.sh [light|task-aware|deep] [task-description]
#
# Levels:
#   light      — Load current goals + learning context (~500 tokens)
#   task-aware — Add methodology research relevant to task (~2,000 tokens)
#   deep       — Full methodology + orchestration research (~5,000 tokens)

set -euo pipefail

LEVEL="${1:-light}"
TASK="${2:-}"

case "$LEVEL" in
  light)
    echo "## Knowledge Injection — Level 1: Lightweight"
    echo ""
    echo "### Current Context"
    qmd search "current goals learning focus ideas" -c second-brain -l 5 2>/dev/null || echo "(qmd not available — vault context skipped)"
    ;;

  task-aware)
    echo "## Knowledge Injection — Level 2: Task-Aware"
    echo ""
    echo "### Current Context"
    qmd search "current goals learning focus" -c second-brain -l 3 2>/dev/null || echo "(qmd not available)"
    echo ""
    echo "### Relevant Methodologies"
    if [ -n "$TASK" ]; then
      qmd query "intent: $TASK
lex: \"harness engineering\" \"spec-driven\" \"quality gate\" \"design-by-contract\"
vec: how to structure specifications with quality gates and verification checkpoints" -c second-brain -l 5 2>/dev/null || echo "(no methodology matches)"
    else
      qmd query 'lex: "harness engineering" "spec-driven" "quality gate"
vec: specification structure with verification checkpoints' -c second-brain -l 5 2>/dev/null || echo "(no methodology matches)"
    fi
    ;;

  deep)
    echo "## Knowledge Injection — Level 3: Deep Research"
    echo ""
    echo "### Current Context"
    qmd search "current goals learning focus" -c second-brain -l 3 2>/dev/null || echo "(qmd not available)"
    echo ""
    echo "### AI Methodologies"
    qmd search "harness engineering spec-driven development GSD quality gate design by contract" -c second-brain -l 10 2>/dev/null || echo "(no methodology matches)"
    echo ""
    echo "### Agent Orchestration Research"
    if [ -n "$TASK" ]; then
      qmd query "intent: $TASK
lex: \"orchestrator\" \"multi-agent\" \"chain\" \"team\" \"composable\"
vec: agent orchestration architecture and workflow patterns" -c second-brain -l 10 2>/dev/null || echo "(no orchestration matches)"
    else
      qmd query 'lex: "orchestrator" "multi-agent" "chain" "team"
vec: composable agent orchestration patterns' -c second-brain -l 10 2>/dev/null || echo "(no orchestration matches)"
    fi
    ;;

  *)
    echo "Usage: knowledge-inject.sh [light|task-aware|deep] [task-description]"
    echo ""
    echo "Levels:"
    echo "  light      — Current context only (~500 tokens)"
    echo "  task-aware — + Methodology research (~2,000 tokens)"
    echo "  deep       — + Full orchestration research (~5,000 tokens)"
    exit 1
    ;;
esac
