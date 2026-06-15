#!/usr/bin/env bash
# skill-freshness.sh — Detect stale skills and cross-reference with vault research
# Usage: ./skill-freshness.sh [--vault PATH] [--stale-days N] [--json]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPENCODE_DIR="$(dirname "$SCRIPT_DIR")"
SKILLS_DIR="$OPENCODE_DIR/skills"
VAULT_DIR="${VAULT_DIR:-$HOME/Documents/second_brain}"
STALE_DAYS="${STALE_DAYS:-30}"
JSON_OUTPUT=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --vault) VAULT_DIR="$2"; shift 2 ;;
        --stale-days) STALE_DAYS="$2"; shift 2 ;;
        --json) JSON_OUTPUT=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

NOW=$(date +%s)
STALE_THRESHOLD=$((STALE_DAYS * 86400))

# Colors
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

declare -a STALE_SKILLS=()
declare -a FRESH_SKILLS=()
declare -a VAULT_GAPS=()
declare -a ORPHANED_SCRIPTS=()

# ============================================================
# 1. Check skill freshness by last-modified date
# ============================================================
check_skill_freshness() {
    local skill_name skill_path skill_file mod_date mod_ts age_days status

    for skill_path in "$SKILLS_DIR"/*/; do
        skill_name=$(basename "$skill_path")
        [[ "$skill_name" == _* ]] && continue  # Skip _INDEX.md, _archived, etc.

        skill_file="$skill_path/SKILL.md"
        [[ ! -f "$skill_file" ]] && continue

        mod_ts=$(stat -f %m "$skill_file" 2>/dev/null || stat -c %Y "$skill_file" 2>/dev/null || echo 0)
        mod_date=$(date -r "$mod_ts" "+%Y-%m-%d" 2>/dev/null || echo "unknown")
        age_days=$(( (NOW - mod_ts) / 86400 ))

        if [[ $age_days -gt $STALE_DAYS ]]; then
            STALE_SKILLS+=("$skill_name|$mod_date|$age_days")
            status="STALE"
        else
            FRESH_SKILLS+=("$skill_name|$mod_date|$age_days")
            status="FRESH"
        fi

        if [[ "$JSON_OUTPUT" == false ]]; then
            if [[ "$status" == "STALE" ]]; then
                printf "${RED}[STALE]${NC} %-25s last updated: %s (%d days ago)\n" "$skill_name" "$mod_date" "$age_days"
            else
                printf "${GREEN}[FRESH]${NC} %-25s last updated: %s (%d days ago)\n" "$skill_name" "$mod_date" "$age_days"
            fi
        fi
    done
}

# ============================================================
# 2. Cross-reference with vault research dates
# ============================================================
check_vault_coverage() {
    # Map of skill topics to vault research files
    declare -A TOPIC_MAP=(
        ["storm-scout"]="Harness-Engineering|Spec-Driven|Orchestration"
        ["build-mode"]="Harness-Engineering|GSD|Test-Driven"
        ["zero-point"]="Quality-Gate|Design-by-Contract|Verification"
        ["battle-bus"]="Multi-Agent|Orchestration|Parallel"
        ["workflow-engine"]="Workflow|Orchestration|Composable"
        ["slurp-juice"]="Context-Engineering|Memory|Checkpoint"
        ["reboot-van"]="Debugging|Root-Cause|Diagnose"
        ["feedback-review"]="Code-Review|Feedback"
        ["pr-review"]="PR-Review|Code-Review"
        ["dev-health"]="Health-Check|Environment"
        ["worktree-manager"]="Worktree|Container|Dev-Environment"
        ["refresh-dev-containers"]="Dev-Container|Refresh"
        ["colima"]="Docker|Container|Colima"
        ["dev-cli"]="Dev-CLI|Teachable"
        ["hotctl"]="Hotmart|AWS|Infrastructure"
        ["graphify"]="Knowledge-Graph|Graphify"
        ["qmd"]="RAG|Vector-Search|BM25"
        ["the-vault"]="Memory|Knowledge-Base"
        ["supply-llama"]="Spreadsheet|CSV|Excel"
    )

    for skill_name in "${!TOPIC_MAP[@]}"; do
        local topics="${TOPIC_MAP[$skill_name]}"
        local skill_path="$SKILLS_DIR/$skill_name"
        [[ ! -d "$skill_path" ]] && continue

        local skill_mod_ts
        skill_mod_ts=$(stat -f %m "$skill_path/SKILL.md" 2>/dev/null || stat -c %Y "$skill_path/SKILL.md" 2>/dev/null || echo 0)

        # Find newer vault research on the same topics
        local newer_research=""
        for topic in ${topics//|/ }; do
            while IFS= read -r vault_file; do
                [[ -z "$vault_file" ]] && continue
                local vault_mod_ts
                vault_mod_ts=$(stat -f %m "$vault_file" 2>/dev/null || stat -c %Y "$vault_file" 2>/dev/null || echo 0)

                if [[ $vault_mod_ts -gt $skill_mod_ts ]]; then
                    local vault_rel
                    vault_rel=$(realpath --relative-to="$VAULT_DIR" "$vault_file" 2>/dev/null || echo "$vault_file")
                    newer_research+="$vault_rel,"
                fi
            done < <(find "$VAULT_DIR" -name "*.md" -type f 2>/dev/null | grep -i "$topic" | head -5)
        done

        if [[ -n "$newer_research" ]]; then
            VAULT_GAPS+=("$skill_name|${newer_research%,}")
            if [[ "$JSON_OUTPUT" == false ]]; then
                printf "${YELLOW}[GAP]${NC}   %-25s newer vault research: %s\n" "$skill_name" "${newer_research%,}"
            fi
        fi
    done
}

# ============================================================
# 3. Check for orphaned scripts (scripts without SKILL.md reference)
# ============================================================
check_orphaned_scripts() {
    for skill_path in "$SKILLS_DIR"/*/; do
        skill_name=$(basename "$skill_path")
        [[ "$skill_name" == _* ]] && continue

        local skill_file="$skill_path/SKILL.md"
        [[ ! -f "$skill_file" ]] && continue

        # Find all scripts in the skill
        for script in "$skill_path"scripts/*.sh; do
            [[ ! -f "$script" ]] && continue
            script_name=$(basename "$script")

            # Check if script is referenced in SKILL.md
            if ! grep -q "$script_name" "$skill_file" 2>/dev/null; then
                ORPHANED_SCRIPTS+=("$skill_name/$script_name")
                if [[ "$JSON_OUTPUT" == false ]]; then
                    printf "${YELLOW}[ORPHAN]${NC}  %-25s script not referenced in SKILL.md: %s\n" "$skill_name" "$script_name"
                fi
            fi
        done
    done
}

# ============================================================
# 4. Check archived skills for potential resurrection
# ============================================================
check_archived_skills() {
    local archived_dir="$SKILLS_DIR/_archived"
    [[ ! -d "$archived_dir" ]] && return

    if [[ "$JSON_OUTPUT" == false ]]; then
        printf "\n${BLUE}[ARCHIVED]${NC} Skills in archive (consider resurrection if relevant):\n"
    fi

    for archived in "$archived_dir"/*/; do
        [[ ! -d "$archived" ]] && continue
        local name=$(basename "$archived")
        local skill_file="$archived/SKILL.md"

        if [[ -f "$skill_file" ]]; then
            local mod_ts
            mod_ts=$(stat -f %m "$skill_file" 2>/dev/null || stat -c %Y "$skill_file" 2>/dev/null || echo 0)
            local mod_date
            mod_date=$(date -r "$mod_ts" "+%Y-%m-%d" 2>/dev/null || echo "unknown")

            # Check if functionality is covered by active skills
            local covered="unknown"
            case "$name" in
                verify) covered="zero-point" ;;
                plan) covered="storm-scout" ;;
                research) covered="storm-scout" ;;
                clarify) covered="storm-scout" ;;
                implement) covered="build-mode" ;;
                iterate) covered="reboot-van" ;;
                diagnose) covered="reboot-van" ;;
                memory) covered="the-vault" ;;
                memory-write) covered="the-vault" ;;
                rtk-guard) covered="slurp-juice" ;;
                anti-speculation) covered="zero-point" ;;
                parallel-execution) covered="battle-bus" ;;
                workflow) covered="workflow-engine" ;;
                tdd-loop) covered="build-mode" ;;
                extract-standards) covered="zero-point" ;;
                orchestrate) covered="engine-control" ;;
                xlsx) covered="supply-llama" ;;
                vision-processing) covered="N/A - unique" ;;
            esac

            if [[ "$JSON_OUTPUT" == false ]]; then
                printf "  %-25s archived: %s → covered by: %s\n" "$name" "$mod_date" "$covered"
            fi
        fi
    done
}

# ============================================================
# 5. Summary
# ============================================================
print_summary() {
    local total_skills=$(find "$SKILLS_DIR" -maxdepth 1 -mindepth 1 -type d ! -name "_*" | wc -l)
    local stale_count=${#STALE_SKILLS[@]}
    local fresh_count=${#FRESH_SKILLS[@]}
    local gap_count=${#VAULT_GAPS[@]}
    local orphan_count=${#ORPHANED_SCRIPTS[@]}

    if [[ "$JSON_OUTPUT" == true ]]; then
        # JSON output
        echo "{"
        echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
        echo "  \"total_skills\": $total_skills,"
        echo "  \"stale_threshold_days\": $STALE_DAYS,"
        echo "  \"stale_skills\": ["
        local first=true
        for entry in "${STALE_SKILLS[@]}"; do
            IFS='|' read -r name date days <<< "$entry"
            [[ "$first" == true ]] && first=false || echo ","
            printf '    {"name": "%s", "last_updated": "%s", "days_old": %s}' "$name" "$date" "$days"
        done
        echo ""
        echo "  ],"
        echo "  \"vault_gaps\": ["
        first=true
        for entry in "${VAULT_GAPS[@]}"; do
            IFS='|' read -r name files <<< "$entry"
            [[ "$first" == true ]] && first=false || echo ","
            printf '    {"skill": "%s", "newer_research": "%s"}' "$name" "$files"
        done
        echo ""
        echo "  ],"
        echo "  \"orphaned_scripts\": ["
        first=true
        for entry in "${ORPHANED_SCRIPTS[@]}"; do
            [[ "$first" == true ]] && first=false || echo ","
            printf '    "%s"' "$entry"
        done
        echo ""
        echo "  ]"
        echo "}"
    else
        # Text summary
        printf "\n${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "${BLUE}  SKILL FRESHNESS SUMMARY${NC}\n"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"
        printf "  Total active skills:    %d\n" "$total_skills"
        printf "  ${GREEN}Fresh (≤%d days):%4d${NC}\n" "$STALE_DAYS" "$fresh_count"
        printf "  ${RED}Stale (>%d days):%4d${NC}\n" "$STALE_DAYS" "$stale_count"
        printf "  ${YELLOW}Vault gaps:       %4d${NC}\n" "$gap_count"
        printf "  ${YELLOW}Orphaned scripts: %4d${NC}\n" "$orphan_count"
        printf "${BLUE}═══════════════════════════════════════════════════════════${NC}\n"

        if [[ $stale_count -gt 0 ]]; then
            printf "\n${RED}⚠ Action needed:${NC} Update stale skills or reduce threshold\n"
        fi
        if [[ $gap_count -gt 0 ]]; then
            printf "${YELLOW}ℹ Review vault gaps:${NC} Skills may be missing newer research\n"
        fi
        if [[ $orphan_count -gt 0 ]]; then
            printf "${YELLOW}ℹ Orphaned scripts:${NC} Reference scripts in SKILL.md or remove\n"
        fi
        if [[ $stale_count -eq 0 && $gap_count -eq 0 && $orphan_count -eq 0 ]]; then
            printf "${GREEN}✓ All skills are fresh and well-maintained!${NC}\n"
        fi
    fi
}

# ============================================================
# Main
# ============================================================
main() {
    if [[ "$JSON_OUTPUT" == false ]]; then
        printf "${BLUE}Skill Freshness Checker${NC}\n"
        printf "Vault: %s\n" "$VAULT_DIR"
        printf "Stale threshold: %d days\n\n" "$STALE_DAYS"
    fi

    check_skill_freshness
    echo ""
    check_vault_coverage
    echo ""
    check_orphaned_scripts
    echo ""
    check_archived_skills
    echo ""
    print_summary
}

main "$@"
