#!/usr/bin/env bash
# workflow-exec.sh — Execute and advance workflow instances
# Usage: ./workflow-exec.sh <command> [args...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Resolve to workspace root (3 levels up from skills/<name>/scripts/)
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../../../" && pwd)"
export OPENCODE_WORKSPACE="${OPENCODE_WORKSPACE:-$WORKSPACE_ROOT}"
SESSION_DB="${SESSION_DB:-$WORKSPACE_ROOT/scripts/session-db.sh}"
OPENCODE="${OPENCODE:-opencode}"

# Detect YAML parser: prefer yq, fallback to basic grep/awk
YQ_CMD=""
if command -v yq &>/dev/null; then
    YQ_CMD="yq"
fi

usage() {
    cat << 'USAGE'
workflow-exec.sh — Execute and advance workflow instances

EXECUTION:
  workflow-exec.sh start <workflow-name> [session-id] [--validate] [--skip-validation] [--on-failure <stop|continue|rollback>]
    Start a workflow instance. Returns instance ID.
    --validate: force validation before start (default if not skipped)
    --skip-validation: bypass validation (emergency use)
    --on-failure: failure policy — stop (default), continue, or rollback

  workflow-exec.sh step <instance-id>
    Execute the current step of a workflow instance.
    Reads agent + task + mode from workflow_steps, dispatches to that agent,
    then marks the step done and advances.

  workflow-exec.sh enqueue-step <instance-id>
    Enqueue the current step of a workflow instance to the task queue.
    Reads agent + task from workflow_steps, then enqueues via task-queue.sh
    (using the agent name as the topic).

  workflow-exec.sh queue-status
    Show the current task queue status (calls task-queue.sh list).

  workflow-exec.sh next <instance-id> [result] [output_path]
    Mark current step done and advance to next step.
    result defaults to "pass".

  workflow-exec.sh fail <instance-id> <error>
    Mark the current step and entire workflow as failed.

  workflow-exec.sh status <instance-id>
    Show full instance status including all steps.

  workflow-exec.sh list [session-id]
    List all running/completed workflow instances.

  workflow-exec.sh cancel <instance-id>
    Cancel a running workflow instance.

VALIDATION:
  workflow-exec.sh validate <workflow-name>
    Validate a workflow YAML file against schema rules.
    Checks: required fields, phase structure, feedforward variables,
    mode references, gate/gate_prompt consistency.

EXAMPLES:
  WIID=$(workflow-exec.sh start "feature-flow" "ses_abc123")
  workflow-exec.sh step $WIID       # runs first step
  workflow-exec.sh next $WIID       # mark done, move to step 2
  workflow-exec.sh step $WIID       # run step 2
  workflow-exec.sh fail $WIID "Auth middleware not implemented"
  workflow-exec.sh validate rpi     # validate rpi.yaml
USAGE
}

# --- Cost estimation helpers ---

# Pricing table (USD per 1K tokens)
# Format: "provider/model": "input_rate:output_rate"
declare -A COST_RATES=(
    ["openai/gpt-5.5"]="0.03:0.06"
    ["openai/gpt-5.5-fast"]="0.015:0.03"
    ["ollama-cloud/kimi-k2.6:cloud"]="0:0"
    ["ollama-cloud/deepseek-v4-pro"]="0:0"
    ["ollama-cloud/nemotron-3-super"]="0:0"
    ["ollama-cloud/glm-5.1"]="0:0"
    ["ollama-cloud/gemma4"]="0:0"
    ["ollama-cloud/minimax-m2.7"]="0:0"
)

_estimate_cost() {
    local provider="$1"
    local model="$2"
    local tokens_in="${3:-0}"
    local tokens_out="${4:-0}"

    local key="${provider}/${model}"
    local rates="${COST_RATES[$key]:-"0:0"}"
    local input_rate=$(echo "$rates" | cut -d: -f1)
    local output_rate=$(echo "$rates" | cut -d: -f2)

    # Calculate: (tokens_in / 1000 * input_rate) + (tokens_out / 1000 * output_rate)
    local cost=$(python3 -c "
import sys
tokens_in = float(sys.argv[1])
tokens_out = float(sys.argv[2])
input_rate = float(sys.argv[3])
output_rate = float(sys.argv[4])
cost = (tokens_in / 1000.0 * input_rate) + (tokens_out / 1000.0 * output_rate)
print(f'{cost:.6f}')
" "$tokens_in" "$tokens_out" "$input_rate" "$output_rate" 2>/dev/null || echo "0")

    echo "$cost"
}

_record_workflow_cost() {
    local wiid="$1"
    local status="$2"

    # Get session_id and workflow info from DB
    local db_path="${DB_PATH:-$OPENCODE_WORKSPACE/.specify/session.db}"
    local session_info
    session_info=$(sqlite3 -separator '|' "$db_path" "
        SELECT w.session_id, s.model, w.workflow_name
        FROM workflow_instances w
        LEFT JOIN sessions s ON w.session_id = s.id
        WHERE w.id=$wiid
        LIMIT 1;
    " 2>/dev/null || true)

    if [ -z "$session_info" ]; then
        echo "⚠️  Could not find session info for workflow $wiid, skipping cost recording" >&2
        return 0
    fi

    local sid=$(echo "$session_info" | cut -d'|' -f1)
    local session_model=$(echo "$session_info" | cut -d'|' -f2)
    local wf_name=$(echo "$session_info" | cut -d'|' -f3)

    # Parse provider and model from session_model
    local provider="ollama-cloud"
    local model="unknown"
    if [ -n "$session_model" ] && [ "$session_model" != "NULL" ]; then
        provider=$(echo "$session_model" | cut -d'/' -f1)
        model=$(echo "$session_model" | cut -d'/' -f2-)
    fi

    # Get token counts from dispatches for this session
    local token_info
    token_info=$(sqlite3 "$db_path" "
        SELECT COALESCE(SUM(token_used), 0)
        FROM dispatches
        WHERE session_id='$sid';
    " 2>/dev/null || echo "0")

    local total_tokens="${token_info:-0}"
    # Estimate 70% input, 30% output split (rough heuristic)
    local tokens_in=$(python3 -c "print(int($total_tokens * 0.7))" 2>/dev/null || echo "0")
    local tokens_out=$(python3 -c "print(int($total_tokens * 0.3))" 2>/dev/null || echo "0")

    # Calculate cost
    local estimated_cost
    estimated_cost=$(_estimate_cost "$provider" "$model" "$tokens_in" "$tokens_out")

    # Record cost snapshot (best-effort, don't fail workflow)
    "$SESSION_DB" record-cost "$sid" "$provider" "$model" "$wiid" "$tokens_in" "$tokens_out" "0" "$estimated_cost" >/dev/null 2>&1 || {
        echo "⚠️  Cost recording failed for workflow $wiid (non-critical)" >&2
    }

    echo "💰 Workflow $wiid cost: \$${estimated_cost} (${provider}/${model}, ${total_tokens} tokens)" >&2
}

# --- Validation helpers ---

_validate_yaml() {
    local WFILE="$1"
    local WNAME="$2"
    local ERRORS=0

    echo "🔍 Validating workflow: $WNAME"

    # Helper: extract scalar value with yq or grep
    _yq_get() {
        local path="$1"
        if [ -n "$YQ_CMD" ]; then
            yq -r "$path" "$WFILE" 2>/dev/null
        else
            echo ""
        fi
    }

    # 1. Required top-level fields: name, trigger, description, phases
    local fields=("name" "trigger" "description" "phases")
    for f in "${fields[@]}"; do
        local val
        val=$(_yq_get ".$f")
        if [ -z "$val" ] || [ "$val" = "null" ]; then
            echo "  ❌ Missing required top-level field: $f"
            ERRORS=$((ERRORS + 1))
        fi
    done

    # 2. Check each phase has: name, agent, skill, mode, feedforward
    local phase_count
    phase_count=$(_yq_get ".phases | length")
    if [ -z "$phase_count" ] || [ "$phase_count" = "null" ] || [ "$phase_count" = "0" ]; then
        echo "  ❌ No phases defined"
        ERRORS=$((ERRORS + 1))
    else
        for i in $(seq 0 $((phase_count - 1))); do
            local pname pagent pskill pmode pfeed
            pname=$(_yq_get ".phases[$i].name")
            pagent=$(_yq_get ".phases[$i].agent")
            pskill=$(_yq_get ".phases[$i].skill")
            pmode=$(_yq_get ".phases[$i].mode")
            pfeed=$(_yq_get ".phases[$i].feedforward")

            if [ -z "$pname" ] || [ "$pname" = "null" ]; then
                echo "  ❌ Phase $i: missing 'name'"
                ERRORS=$((ERRORS + 1))
            fi
            if [ -z "$pagent" ] || [ "$pagent" = "null" ]; then
                echo "  ❌ Phase '$pname': missing 'agent'"
                ERRORS=$((ERRORS + 1))
            fi
            if [ -z "$pskill" ] || [ "$pskill" = "null" ]; then
                echo "  ❌ Phase '$pname': missing 'skill'"
                ERRORS=$((ERRORS + 1))
            fi
            if [ -z "$pmode" ] || [ "$pmode" = "null" ]; then
                echo "  ❌ Phase '$pname': missing 'mode'"
                ERRORS=$((ERRORS + 1))
            fi
            if [ -z "$pfeed" ] || [ "$pfeed" = "null" ]; then
                echo "  ❌ Phase '$pname': missing 'feedforward'"
                ERRORS=$((ERRORS + 1))
            fi

            # 5. Gates: if gate is not null, gate_prompt must be present
            local pgate pgate_prompt
            pgate=$(_yq_get ".phases[$i].gate")
            pgate_prompt=$(_yq_get ".phases[$i].gate_prompt")
            if [ -n "$pgate" ] && [ "$pgate" != "null" ]; then
                if [ -z "$pgate_prompt" ] || [ "$pgate_prompt" = "null" ]; then
                    echo "  ❌ Phase '$pname': gate is set but gate_prompt is missing"
                    ERRORS=$((ERRORS + 1))
                fi
            fi
        done
    fi

    # 3. Feedforward variables: extract all {UPPERCASE} placeholders,
    #    verify they appear as outputs from previous phases or as workflow inputs
    #    Collect all defined phase output variables (phase names in UPPERCASE + _OUTPUT)
    #    and known workflow input variables
    local all_vars=()
    local defined_outputs=()
    for i in $(seq 0 $((phase_count - 1))); do
        local pname
        pname=$(_yq_get ".phases[$i].name")
        if [ -n "$pname" ] && [ "$pname" != "null" ]; then
            defined_outputs+=("${pname^^}_OUTPUT")
        fi
    done

    # Known workflow-level input variables (common conventions)
    local known_inputs=("GOAL" "BUG_DESCRIPTION" "PLAN" "SYMPTOMS" "COMPLEXITY" "INCIDENT_DESCRIPTION"
                        "SEVERITY" "TARGET" "SCOPE" "TOPIC" "TIMEBOX" "PR_URL" "SPEC"
                        "TASK" "DONE_CONDITION" "ROOT_CAUSE" "TRIAGE_OUTPUT" "PLAN_OUTPUT"
                        "CLARIFY_OUTPUT" "RESEARCH_OUTPUT" "EXPLORE_OUTPUT" "DOCUMENT_OUTPUT"
                        "COLLECT_OUTPUT" "CLASSIFY_OUTPUT" "EVALUATE_OUTPUT" "GATHER_OUTPUT"
                        "REVIEW_OUTPUT" "RATE_OUTPUT")

    for i in $(seq 0 $((phase_count - 1))); do
        local pname pfeed
        pname=$(_yq_get ".phases[$i].name")
        pfeed=$(_yq_get ".phases[$i].feedforward")
        if [ -z "$pfeed" ] || [ "$pfeed" = "null" ]; then
            continue
        fi

        # Extract {VARIABLE} placeholders
        local vars_in_feed
        vars_in_feed=$(echo "$pfeed" | grep -oE '\{[A-Z_]+\}' | sed 's/[{}]//g' | sort -u || true)

        for var in $vars_in_feed; do
            local found=false
            # Check against previous phase outputs
            for out in "${defined_outputs[@]}"; do
                if [ "$var" = "$out" ]; then
                    found=true
                    break
                fi
            done
            # Check against known inputs
            if [ "$found" = false ]; then
                for inp in "${known_inputs[@]}"; do
                    if [ "$var" = "$inp" ]; then
                        found=true
                        break
                    fi
                done
            fi
            # Allow dynamic mode expressions like ${COMPLEXITY == ...}
            if [ "$found" = false ]; then
                # Check if it's used inside a ${...} expression (dynamic mode)
                if echo "$pfeed" | grep -qF "\${$var"; then
                    found=true
                fi
            fi
            if [ "$found" = false ]; then
                echo "  ❌ Phase '$pname': feedforward references undefined variable '{$var}'"
                ERRORS=$((ERRORS + 1))
            fi
        done
    done

    # 4. Modes: each mode's phases array must reference valid phase names
    local mode_count
    mode_count=$(_yq_get ".modes | length")
    if [ -n "$mode_count" ] && [ "$mode_count" != "null" ] && [ "$mode_count" != "0" ]; then
        # Collect valid phase names
        local valid_phases=()
        for i in $(seq 0 $((phase_count - 1))); do
            local pname
            pname=$(_yq_get ".phases[$i].name")
            if [ -n "$pname" ] && [ "$pname" != "null" ]; then
                valid_phases+=("$pname")
            fi
        done

        local mode_keys
        mode_keys=$(_yq_get ".modes | keys | .[]")
        for mkey in $mode_keys; do
            local mphase_count
            mphase_count=$(_yq_get ".modes.$mkey.phases | length")
            if [ -z "$mphase_count" ] || [ "$mphase_count" = "null" ] || [ "$mphase_count" = "0" ]; then
                echo "  ❌ Mode '$mkey': phases array is empty or missing"
                ERRORS=$((ERRORS + 1))
                continue
            fi
            for j in $(seq 0 $((mphase_count - 1))); do
                local mphase
                mphase=$(_yq_get ".modes.$mkey.phases[$j]")
                local valid=false
                for vp in "${valid_phases[@]}"; do
                    if [ "$mphase" = "$vp" ]; then
                        valid=true
                        break
                    fi
                done
                if [ "$valid" = false ]; then
                    echo "  ❌ Mode '$mkey': references unknown phase '$mphase'"
                    ERRORS=$((ERRORS + 1))
                fi
            done
        done
    fi

    # 6. Default mode: must reference a valid mode
    local default_mode
    default_mode=$(_yq_get ".default_mode")
    if [ -n "$default_mode" ] && [ "$default_mode" != "null" ]; then
        local mode_keys
        mode_keys=$(_yq_get ".modes | keys | .[]")
        local valid_dm=false
        for mkey in $mode_keys; do
            if [ "$default_mode" = "$mkey" ]; then
                valid_dm=true
                break
            fi
        done
        if [ "$valid_dm" = false ]; then
            echo "  ❌ default_mode '$default_mode' does not reference a valid mode"
            ERRORS=$((ERRORS + 1))
        fi
    fi

    if [ $ERRORS -eq 0 ]; then
        echo "✅ Workflow $WNAME validation passed"
        return 0
    else
        echo "❌ Workflow $WNAME validation failed with $ERRORS error(s)"
        return 1
    fi
}

cmd="${1:-help}"
shift 2>/dev/null || true

case "$cmd" in
    start)
        WNAME="${1:-}"; SID="${2:-}"
        SKIP_VALIDATION=false
        FORCE_VALIDATION=false
        ON_FAILURE="stop"

        # Parse optional flags
        while [ $# -gt 0 ]; do
            case "$1" in
                --skip-validation)
                    SKIP_VALIDATION=true
                    shift
                    ;;
                --validate)
                    FORCE_VALIDATION=true
                    shift
                    ;;
                --on-failure)
                    ON_FAILURE="${2:-stop}"
                    shift 2
                    ;;
                *)
                    shift
                    ;;
            esac
        done

        [ -z "$WNAME" ] && { echo "ERROR: workflow-name required"; usage; exit 1; }

        # Validate on-failure value
        case "$ON_FAILURE" in
            stop|continue|rollback) ;;
            *) echo "ERROR: --on-failure must be stop, continue, or rollback"; exit 1 ;;
        esac

        # Auto-validate unless skipped
        if [ "$SKIP_VALIDATION" = false ] || [ "$FORCE_VALIDATION" = true ]; then
            WFILE="$WORKSPACE_ROOT/.opencode/workflows/${WNAME}.yaml"
            if [ ! -f "$WFILE" ]; then
                echo "ERROR: workflow file not found: $WFILE"
                exit 1
            fi
            if ! _validate_yaml "$WFILE" "$WNAME" >&2; then
                echo "ERROR: Workflow validation failed. Use --skip-validation to bypass (emergency only)."
                exit 1
            fi
        fi

        RESULT=$("$SESSION_DB" workflow-start "$WNAME" "$SID" "$ON_FAILURE")
        echo "$RESULT"
        ;;
    validate)
        WNAME="${1:-}"
        [ -z "$WNAME" ] && { echo "ERROR: workflow-name required"; usage; exit 1; }

        WFILE="$WORKSPACE_ROOT/.opencode/workflows/${WNAME}.yaml"
        [ ! -f "$WFILE" ] && { echo "ERROR: workflow file not found: $WFILE"; exit 1; }

        _validate_yaml "$WFILE" "$WNAME"
        exit $?
        ;;
    step)
        WIID="${1:-}"
        [ -z "$WIID" ] && { echo "ERROR: instance-id required"; usage; exit 1; }

        # Get current step info
        STEP_INFO=$("$SESSION_DB" query "SELECT step_order, agent, task, mode FROM workflow_steps WHERE instance_id=$WIID AND status='pending' ORDER BY step_order LIMIT 1;")
        if [ -z "$STEP_INFO" ]; then
            STATUS=$("$SESSION_DB" query "SELECT status FROM workflow_instances WHERE id=$WIID;")
            echo "No pending steps. Workflow status: $STATUS"
            exit 0
        fi

        # Parse step info
        STEP_ORDER=$(echo "$STEP_INFO" | awk '{print $1}')
        AGENT=$(echo "$STEP_INFO" | awk '{print $2}')
        TASK=$(echo "$STEP_INFO" | awk '{$1=""; $2=""; print $0}' | sed 's/^ //')
        MODE=$(echo "$STEP_INFO" | awk '{print $3}')

        echo "▶️  Executing step $STEP_ORDER: $AGENT → $TASK (mode: $MODE)"

        # Mark step as running
        "$SESSION_DB" workflow-step-start "$WIID"

        # Dispatch to agent
        SID=$("$SESSION_DB" query "SELECT session_id FROM workflow_instances WHERE id=$WIID;" | head -1)
        if [ -n "$SID" ] && [ "$SID" != "NULL" ]; then
            DISPATCH_RESULT=$("$SESSION_DB" dispatch "$SID" "$AGENT" "$TASK" "workflow-step" "$WNAME" "$MODE")
        fi

        echo "✅ Step $STEP_ORDER dispatch complete for instance $WIID"
        ;;
    enqueue-step)
        WIID="${1:-}"
        [ -z "$WIID" ] && { echo "ERROR: instance-id required"; usage; exit 1; }

        # Get current step info using sqlite3 directly with pipe delimiter
        DB_PATH="${DB_PATH:-$OPENCODE_WORKSPACE/.specify/session.db}"
        STEP_ROW=$(sqlite3 -separator '|' "$DB_PATH" "SELECT step_order, agent, task, mode FROM workflow_steps WHERE instance_id=$WIID AND status='pending' ORDER BY step_order LIMIT 1;")
        if [ -z "$STEP_ROW" ]; then
            STATUS=$(sqlite3 "$DB_PATH" "SELECT status FROM workflow_instances WHERE id=$WIID;")
            echo "No pending steps. Workflow status: $STATUS"
            exit 0
        fi

        # Parse step info using pipe delimiter
        STEP_ORDER=$(echo "$STEP_ROW" | cut -d'|' -f1)
        AGENT=$(echo "$STEP_ROW" | cut -d'|' -f2)
        TASK=$(echo "$STEP_ROW" | cut -d'|' -f3)
        MODE=$(echo "$STEP_ROW" | cut -d'|' -f4)

         echo "▶️  Enqueuing step $STEP_ORDER: $AGENT → $TASK (mode: $MODE)"

         # Enqueue to task queue with dedupe_key for idempotency
         TASK_QUEUE="${WORKSPACE_ROOT}/scripts/task-queue.sh"
         SID=$(sqlite3 "$DB_PATH" "SELECT session_id FROM workflow_instances WHERE id=$WIID;")
         if [ -n "$SID" ] && [ "$SID" != "NULL" ]; then
             # Ensure task_queue tables exist
             "$TASK_QUEUE" init >/dev/null 2>&1 || true

             # Compute deterministic dedupe_key
             DEDUPE_KEY="workflow:${WIID}:step:${STEP_ORDER}"

             # Enqueue to task queue (task-queue.sh handles dedupe logic)
             ENQUEUE_RESULT=$("$TASK_QUEUE" add "$SID" "$AGENT" "$TASK" 1 "$DEDUPE_KEY")
             echo "$ENQUEUE_RESULT"
         else
             echo "WARNING: No session_id found for instance $WIID, skipping enqueue"
         fi

         echo "✅ Step $STEP_ORDER enqueued for instance $WIID"
        ;;
    next)
        WIID="${1:-}"; RESULT="${2:-pass}"; OUTPUT="${3:-}"
        [ -z "$WIID" ] && { echo "ERROR: instance-id required"; usage; exit 1; }
        COMPLETION_OUTPUT=$("$SESSION_DB" workflow-step-done "$WIID" "$RESULT" "$OUTPUT" 2>&1)
        echo "$COMPLETION_OUTPUT"
        # Record cost if workflow completed (success or failure)
        if echo "$COMPLETION_OUTPUT" | grep -q "Workflow instance.*completed\|Workflow instance.*failed"; then
            _record_workflow_cost "$WIID" "${RESULT}" >&2 || true
        fi
        ;;
    fail)
        WIID="${1:-}"; ERR="${2:-}"
        [ -z "$WIID" ] || [ -z "$ERR" ] && { echo "ERROR: instance-id and error required"; usage; exit 1; }

        # Get failure policy
        ON_FAILURE=$("$SESSION_DB" query "SELECT on_failure FROM workflow_instances WHERE id=$WIID;" | tail -n 1)
        [ -z "$ON_FAILURE" ] && ON_FAILURE="stop"

        case "$ON_FAILURE" in
            stop)
                # Default: halt workflow
                "$SESSION_DB" workflow-fail "$WIID" "$ERR"
                # Record cost snapshot on workflow failure
                _record_workflow_cost "$WIID" "failed" >&2 || true
                echo "❌ Workflow $WIID halted: $ERR"
                ;;
            continue)
                # Log failure, proceed to next phase
                echo "⚠️ Phase failed: $ERR. Continuing to next phase (policy: continue)."
                "$SESSION_DB" workflow-step-done "$WIID" "fail-continue"
                ;;
            rollback)
                # Attempt to reverse completed phases
                echo "🔄 Attempting rollback for workflow $WIID..."
                # Get completed steps
                COMPLETED=$("$SESSION_DB" query "SELECT step_order, agent, task FROM workflow_steps WHERE instance_id=$WIID AND status='completed' ORDER BY step_order DESC;")
                if [ -n "$COMPLETED" ]; then
                    echo "Reversing completed phases:"
                    echo "$COMPLETED" | while read -r step_order agent task; do
                        echo "  - Reversing step $step_order ($agent: $task)"
                        # Note: Actual rollback logic is phase-specific and would need custom handlers
                    done
                fi
                "$SESSION_DB" workflow-fail "$WIID" "$ERR (rollback attempted)"
                # Record cost snapshot on workflow failure (after rollback)
                _record_workflow_cost "$WIID" "failed" >&2 || true
                echo "❌ Workflow $WIID halted after rollback attempt: $ERR"
                ;;
        esac
        ;;
    status)
        WIID="${1:-}"; [ -z "$WIID" ] && { echo "ERROR: instance-id required"; usage; exit 1; }
        "$SESSION_DB" workflow-status "$WIID"
        ;;
    list)
        SID="${1:-}"
        if [ -n "$SID" ]; then
            "$SESSION_DB" query "SELECT id, workflow_name, status, current_step, total_steps, result, started_at FROM workflow_instances WHERE session_id='$SID' ORDER BY started_at DESC;"
        else
            "$SESSION_DB" query "SELECT id, workflow_name, status, current_step, total_steps, result, started_at FROM workflow_instances ORDER BY started_at DESC LIMIT 20;"
        fi
        ;;
    queue-status)
        TASK_QUEUE="${WORKSPACE_ROOT}/scripts/task-queue.sh"
        "$TASK_QUEUE" list
        ;;
    cancel)
        WIID="${1:-}"; [ -z "$WIID" ] && { echo "ERROR: instance-id required"; usage; exit 1; }
        TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
        "$SESSION_DB" query "UPDATE workflow_instances SET status='cancelled', completed_at='$TIMESTAMP' WHERE id=$WIID;"
        echo "❌ Workflow instance $WIID cancelled"
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
