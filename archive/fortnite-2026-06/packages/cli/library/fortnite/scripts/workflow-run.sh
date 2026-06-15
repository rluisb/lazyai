#!/usr/bin/env bash
# workflow-run.sh — Workflow orchestrator for Fortnite multi-agent system
# Reads YAML configs from .opencode/workflows/, syncs to SQLite, executes phases
# Usage: ./workflow-run.sh <command> [args...]

set -euo pipefail

WORKSPACE="${OPENCODE_WORKSPACE:-.}"
WORKFLOW_DIR="$WORKSPACE/.opencode/workflows"
DB_PATH="$WORKSPACE/.specify/session.db"
LEDGER_SCRIPT="$WORKSPACE/skills/truth-chain/scripts/ledger.sh"
CMD="${1:-help}"
shift 2>/dev/null || true

# --- HELPERS ---

now() { date -u '+%Y-%m-%dT%H:%M:%SZ'; }
uuid() { python3 -c "import uuid; print(str(uuid.uuid4())[:12])"; }

init_db() {
    # Ensure shared session tables exist (including workflow_instances / workflow_steps)
    local SESSION_DB_SCRIPT="${OPENCODE_WORKSPACE:-.}/scripts/session-db.sh"
    if [[ -x "$SESSION_DB_SCRIPT" ]]; then
        "$SESSION_DB_SCRIPT" init 2>/dev/null || true
    fi

    sqlite3 "$DB_PATH" "
        CREATE TABLE IF NOT EXISTS workflows (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            trigger_cmd TEXT NOT NULL,
            config_json TEXT NOT NULL,
            version INTEGER DEFAULT 1,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS workflow_runs (
            id TEXT PRIMARY KEY,
            workflow_id TEXT REFERENCES workflows(id),
            session_id TEXT,
            mode TEXT NOT NULL,
            status TEXT NOT NULL,
            current_phase TEXT,
            phases_completed INTEGER DEFAULT 0,
            phases_total INTEGER,
            started_at TEXT NOT NULL,
            completed_at TEXT,
            feedforward_context TEXT
        );

        CREATE TABLE IF NOT EXISTS workflow_phases (
            id TEXT PRIMARY KEY,
            run_id TEXT REFERENCES workflow_runs(id),
            phase_name TEXT NOT NULL,
            agent TEXT NOT NULL,
            skill TEXT,
            mode TEXT,
            status TEXT NOT NULL,
            input_context TEXT,
            output TEXT,
            gate_passed BOOLEAN,
            started_at TEXT,
            completed_at TEXT,
            ledger_seq INTEGER
        );

        CREATE TABLE IF NOT EXISTS workflow_instances (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            workflow_name TEXT NOT NULL,
            session_id TEXT REFERENCES sessions(id),
            status TEXT DEFAULT 'pending',
            current_step INTEGER DEFAULT 0,
            total_steps INTEGER NOT NULL,
            result TEXT,
            started_at TEXT NOT NULL,
            completed_at TEXT,
            error_message TEXT
        );

        CREATE TABLE IF NOT EXISTS workflow_steps (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            instance_id INTEGER REFERENCES workflow_instances(id),
            step_order INTEGER NOT NULL,
            agent TEXT NOT NULL,
            task TEXT NOT NULL,
            mode TEXT,
            status TEXT DEFAULT 'pending',
            started_at TEXT,
            completed_at TEXT,
            result TEXT,
            output_path TEXT,
            error_message TEXT
        );

        CREATE INDEX IF NOT EXISTS idx_workflow_instances_session ON workflow_instances(session_id);
        CREATE INDEX IF NOT EXISTS idx_workflow_instances_status ON workflow_instances(status);
        CREATE INDEX IF NOT EXISTS idx_workflow_steps_instance ON workflow_steps(instance_id);
    "
}

# Simple YAML-to-JSON parser (handles our workflow config format)
yaml_to_json() {
    local file="$1"
    python3 - "$file" << 'PYEOF'
import json, sys, re

def parse_yaml(filepath):
    """Simple YAML parser for workflow configs (no pyyaml needed)."""
    with open(filepath, 'r') as f:
        lines = f.readlines()

    result = {}
    current_key = None
    current_dict = None
    current_list = None
    current_list_item = None
    in_modes = False
    in_phases = False
    mode_key = None
    phase_idx = -1

    for line in lines:
        stripped = line.rstrip()
        if not stripped or stripped.startswith('#'):
            continue

        indent = len(line) - len(line.lstrip())

        # Top-level key
        if indent == 0 and ':' in stripped:
            key, _, val = stripped.partition(':')
            key = key.strip()
            val = val.strip()
            if val:
                result[key] = parse_value(val)
                current_key = None
                current_dict = None
                current_list = None
                in_modes = False
                in_phases = False
            elif key == 'modes':
                result['modes'] = {}
                in_modes = True
                in_phases = False
                current_key = 'modes'
            elif key == 'phases':
                result['phases'] = []
                in_phases = True
                in_modes = False
                current_key = 'phases'
            else:
                result[key] = {}
                current_key = key
                current_dict = result[key]
                in_modes = False
                in_phases = False
            continue

        # Modes section
        if in_modes:
            if indent == 2 and ':' in stripped:
                mode_key, _, mode_val = stripped.partition(':')
                mode_key = mode_key.strip()
                mode_val = mode_val.strip()
                if mode_val:
                    result['modes'][mode_key] = parse_value(mode_val)
                else:
                    result['modes'][mode_key] = {}
                current_dict = result['modes'][mode_key]
            elif indent == 4 and ':' in stripped and current_dict is not None:
                key, _, val = stripped.partition(':')
                key = key.strip()
                val = val.strip()
                if val.startswith('[') and val.endswith(']'):
                    current_dict[key] = [v.strip().strip('"').strip("'") for v in val[1:-1].split(',') if v.strip()]
                else:
                    current_dict[key] = parse_value(val)
            continue

        # Phases section
        if in_phases:
            content = stripped.lstrip()
            if indent == 2 and content.startswith('- '):
                phase_idx += 1
                result['phases'].append({})
                current_dict = result['phases'][phase_idx]
                # Handle inline key-value after -
                item = content[2:]
                if ':' in item:
                    key, _, val = item.partition(':')
                    current_dict[key.strip()] = parse_value(val.strip())
            elif indent == 4 and ':' in stripped and current_dict is not None:
                key, _, val = stripped.partition(':')
                key = key.strip()
                val = val.strip()
                if val.startswith('[') and val.endswith(']'):
                    current_dict[key] = [v.strip().strip('"').strip("'") for v in val[1:-1].split(',') if v.strip()]
                else:
                    current_dict[key] = parse_value(val)
            continue

    return result

def parse_value(val):
    if val == 'null' or val == 'None':
        return None
    if val == 'true':
        return True
    if val == 'false':
        return False
    if val.startswith('"') and val.endswith('"'):
        return val[1:-1]
    if val.startswith("'") and val.endswith("'"):
        return val[1:-1]
    try:
        return int(val)
    except ValueError:
        pass
    try:
        return float(val)
    except ValueError:
        pass
    return val

config = parse_yaml(sys.argv[1])
print(json.dumps(config, indent=2))
PYEOF
}

# --- COMMANDS ---

sync() {
    init_db
    if [[ ! -d "$WORKFLOW_DIR" ]]; then
        echo "❌ Workflow directory not found: $WORKFLOW_DIR"
        exit 1
    fi

    local count=0
    for yaml_file in "$WORKFLOW_DIR"/*.yaml; do
        [[ -f "$yaml_file" ]] || continue

        local json
        json=$(yaml_to_json "$yaml_file")
        local wf_id
        wf_id=$(echo "$json" | python3 -c "import sys,json; print(json.load(sys.stdin)['name'])")
        local trigger
        trigger=$(echo "$json" | python3 -c "import sys,json; print(json.load(sys.stdin)['trigger'])")
        local name
        name=$(echo "$json" | python3 -c "import sys,json; print(json.load(sys.stdin).get('description',''))")
        local escaped_json
        escaped_json=$(echo "$json" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin)))")

        local existing
        existing=$(sqlite3 "$DB_PATH" "SELECT id FROM workflows WHERE id='$wf_id';")

        if [[ -z "$existing" ]]; then
            sqlite3 "$DB_PATH" "
                INSERT INTO workflows(id, name, description, trigger_cmd, config_json, version, created_at, updated_at)
                VALUES ('$wf_id', '$wf_id', '$name', '$trigger', '$(echo "$escaped_json" | sed "s/'/''/g")', 1, '$(now)', '$(now)');
            "
            echo "✅ Registered workflow: $wf_id ($trigger)"
        else
            sqlite3 "$DB_PATH" "
                UPDATE workflows SET config_json='$(echo "$escaped_json" | sed "s/'/''/g")', updated_at='$(now)', version=version+1
                WHERE id='$wf_id';
            "
            echo "🔄 Updated workflow: $wf_id ($trigger)"
        fi
        count=$((count + 1))
    done

    echo ""
    echo "📊 Synced $count workflows to SQLite"
}

list() {
    init_db
    local count
    count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflows;")
    if [[ "$count" -eq 0 ]]; then
        echo "⚠️  No workflows registered. Run 'workflow-run.sh sync' first."
        exit 0
    fi

    echo "📋 Available Workflows ($count):"
    echo ""
    sqlite3 -header -column "$DB_PATH" "
        SELECT id as workflow, trigger_cmd as trigger, description, version, updated_at
        FROM workflows ORDER BY id;
    "
    echo ""
    echo "Run: workflow-run.sh run <workflow> --mode <mode> \"<goal>\""
}

run() {
    init_db
    local WORKFLOW_ID="" MODE="" GOAL="" SKIP_GATES=false DRY_RUN=false SESSION_ID=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --mode) MODE="$2"; shift 2 ;;
            --skip-gates) SKIP_GATES=true; shift ;;
            --dry-run) DRY_RUN=true; shift ;;
            --session) SESSION_ID="$2"; shift 2 ;;
            --) shift; GOAL="$*"; break ;;
            -*) echo "Unknown flag: $1"; exit 1 ;;
            *)
                if [[ -z "$WORKFLOW_ID" ]]; then
                    WORKFLOW_ID="$1"
                else
                    GOAL="${GOAL:+$GOAL }$1"
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$WORKFLOW_ID" ]]; then
        echo "Usage: workflow-run.sh run <workflow> [--mode <mode>] [--skip-gates] [--dry-run] \"<goal>\""
        echo ""
        echo "Examples:"
        echo "  workflow-run.sh run rpi --mode complex \"Add OAuth support\""
        echo "  workflow-run.sh run bugfix --mode known \"Fix null pointer in auth\""
        echo "  workflow-run.sh run hotfix \"Production timeout on /api/users\""
        echo "  workflow-run.sh run rpi --dry-run --mode simple \"Test workflow\""
        exit 1
    fi

    # Load workflow config
    local config
    config=$(sqlite3 "$DB_PATH" "SELECT config_json FROM workflows WHERE id='$WORKFLOW_ID';")
    if [[ -z "$config" ]]; then
        echo "❌ Workflow not found: $WORKFLOW_ID"
        echo "   Available: $(sqlite3 "$DB_PATH" "SELECT GROUP_CONCAT(id, ', ') FROM workflows;")"
        exit 1
    fi

    # Parse config
    local default_mode phases_json
    default_mode=$(echo "$config" | python3 -c "import sys,json; c=json.load(sys.stdin); print(c.get('default_mode','standard'))")
    MODE="${MODE:-$default_mode}"

    # Get mode config
    local mode_phases skip_gates_mode
    mode_phases=$(echo "$config" | python3 -c "
import sys, json
c = json.load(sys.stdin)
modes = c.get('modes', {})
mode = modes.get('$MODE', {})
phases = mode.get('phases', [p['name'] for p in c.get('phases', [])])
skip = mode.get('skip_gates', False)
print(json.dumps({'phases': phases, 'skip_gates': skip}))
")
    skip_gates_mode=$(echo "$mode_phases" | python3 -c "import sys,json; print(json.load(sys.stdin)['skip_gates'])")
    if [[ "$SKIP_GATES" == "true" ]]; then
        skip_gates_mode="true"
    fi

    # Get phase definitions
    phases_json=$(echo "$config" | python3 -c "
import sys, json
c = json.load(sys.stdin)
mode_phases = json.loads('$mode_phases')['phases']
all_phases = {p['name']: p for p in c.get('phases', [])}
ordered = [all_phases[p] for p in mode_phases if p in all_phases]
print(json.dumps(ordered))
")

    local phase_count
    phase_count=$(echo "$phases_json" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")

    # Create workflow run record
    local RUN_ID="wfr_$(uuid)"
    local TS=$(now)

    if [[ "$DRY_RUN" == "false" ]]; then
        sqlite3 "$DB_PATH" "
            INSERT INTO workflow_runs(id, workflow_id, session_id, mode, status, phases_total, started_at, feedforward_context)
            VALUES ('$RUN_ID', '$WORKFLOW_ID', '${SESSION_ID:-}', '$MODE', 'running', $phase_count, '$TS', '{\"goal\":\"${GOAL:-}\"}');
        "
    fi

    echo "🚀 Workflow: $WORKFLOW_ID (mode: $MODE)"
    echo "   Run ID: $RUN_ID"
    echo "   Goal: ${GOAL:-N/A}"
    echo "   Phases: $phase_count"
    echo "   Skip gates: $skip_gates_mode"
    echo "   Dry run: $DRY_RUN"
    echo ""

    # Execute phases
    local feedforward_context="{\"goal\":\"${GOAL:-}\"}"
    local phase_num=0

    echo "$phases_json" | python3 -c "
import sys, json

def interpolate(val, context):
    if not val or not isinstance(val, str): return val
    result = val
    for key, value in context.items():
        result = result.replace('{'+key+'}', str(value))
    # Handle ternary: \${COMPLEXITY == 'complex' ? 'senior' : 'standard'}
    import re
    def ternary(m):
        expr = m.group(1)
        parts = expr.split('?')
        if len(parts) == 2:
            condition = parts[0].strip()
            then_else = parts[1].split(':')
            if len(then_else) == 2:
                then_val = then_else[0].strip().strip(\"'\").strip('\"')
                else_val = then_else[1].strip().strip(\"'\").strip('\"')
                # Simple equality check
                if '==' in condition:
                    left, right = condition.split('==')
                    left = left.strip()
                    right = right.strip().strip(\"'\").strip('\"')
                    ctx_val = context.get(left, '')
                    return then_val if ctx_val == right else else_val
        return m.group(0)
    result = re.sub(r'\\\$\\{([^}]+)\\}', ternary, result)
    return result

phases = json.load(sys.stdin)
context = {'COMPLEXITY': '$MODE', 'GOAL': '$GOAL'}
for i, p in enumerate(phases):
    mode = interpolate(p.get('mode','?'), context)
    ff = interpolate(p.get('feedforward',''), context)
    print(f'{i+1}|{p.get(\"name\",\"?\")}|{p.get(\"agent\",\"?\")}|{p.get(\"skill\",\"?\")}|{mode}|{ff}|{p.get(\"gate\",\"null\")}|{p.get(\"gate_prompt\",\"\")}')
" | while IFS='|' read -r num name agent skill mode feedforward gate gate_prompt; do
        phase_num=$((phase_num + 1))
        echo "━━━ Phase $num/$phase_count: $name ━━━"
        echo "   Agent: $agent"
        echo "   Skill: $skill"
        echo "   Mode: $mode"
        echo "   Gate: $gate"
        echo ""

        if [[ "$DRY_RUN" == "true" ]]; then
            echo "   [DRY RUN] Would dispatch to $agent with skill=$skill mode=$mode"
            echo "   [DRY RUN] Feedforward: $feedforward"
        else
            # Log to ledger
            if [[ -x "$LEDGER_SCRIPT" ]]; then
                "$LEDGER_SCRIPT" append --type workflow_step_start \
                    --data "{\"workflow\":\"$WORKFLOW_ID\",\"run_id\":\"$RUN_ID\",\"phase\":\"$name\",\"agent\":\"$agent\"}" \
                    --session "${SESSION_ID:-workflow}" >/dev/null 2>&1 || true
            fi

            # Record phase in SQLite
            local PHASE_ID="wfp_$(uuid)"
            sqlite3 "$DB_PATH" "
                INSERT INTO workflow_phases(id, run_id, phase_name, agent, skill, mode, status, input_context, started_at)
                VALUES ('$PHASE_ID', '$RUN_ID', '$name', '$agent', '${skill:-}', '${mode:-}', 'running', '$(echo "$feedforward_context" | sed "s/'/''/g")', '$(now)');
            "

            # TODO: Actual dispatch via task tool would go here
            # For now, simulate completion
            echo "   ✅ Phase $name completed (simulated)"

            # Update phase status
            sqlite3 "$DB_PATH" "
                UPDATE workflow_phases SET status='completed', completed_at='$(now)' WHERE id='$PHASE_ID';
            "

            # Log to ledger
            if [[ -x "$LEDGER_SCRIPT" ]]; then
                "$LEDGER_SCRIPT" append --type workflow_step_done \
                    --data "{\"workflow\":\"$WORKFLOW_ID\",\"run_id\":\"$RUN_ID\",\"phase\":\"$name\"}" \
                    --session "${SESSION_ID:-workflow}" >/dev/null 2>&1 || true
            fi
        fi

        # Human gate
        if [[ "$gate" != "null" && "$skip_gates_mode" != "true" && "$DRY_RUN" == "false" ]]; then
            echo ""
            echo "   ⏸️  GATE: ${gate_prompt:-Proceed?}"
            echo "   (In production, this would pause for human approval)"
            echo ""
        fi

        echo ""
    done

    # Complete workflow run
    if [[ "$DRY_RUN" == "false" ]]; then
        sqlite3 "$DB_PATH" "
            UPDATE workflow_runs SET status='completed', completed_at='$(now)', phases_completed=$phase_count
            WHERE id='$RUN_ID';
        "

        # Log to ledger
        if [[ -x "$LEDGER_SCRIPT" ]]; then
            "$LEDGER_SCRIPT" append --type workflow_complete \
                --data "{\"workflow\":\"$WORKFLOW_ID\",\"run_id\":\"$RUN_ID\",\"phases_completed\":$phase_count}" \
                --session "${SESSION_ID:-workflow}" >/dev/null 2>&1 || true
        fi
    fi

    echo "✅ Workflow $WORKFLOW_ID completed ($phase_count phases)"
}

status() {
    init_db
    local RUN_ID="${1:-}"

    if [[ -n "$RUN_ID" ]]; then
        # Show specific run
        sqlite3 -header -column "$DB_PATH" "
            SELECT id, workflow_id, mode, status, current_phase, phases_completed, phases_total, started_at, completed_at
            FROM workflow_runs WHERE id='$RUN_ID';
        "
        echo ""
        echo "Phases:"
        sqlite3 -header -column "$DB_PATH" "
            SELECT phase_name, agent, skill, mode, status, gate_passed, started_at, completed_at
            FROM workflow_phases WHERE run_id='$RUN_ID' ORDER BY phase_name;
        "
    else
        # Show recent runs
        local count
        count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflow_runs;")
        if [[ "$count" -eq 0 ]]; then
            echo "⚠️  No workflow runs recorded."
            exit 0
        fi
        echo "📊 Recent Workflow Runs ($count):"
        echo ""
        sqlite3 -header -column "$DB_PATH" "
            SELECT id, workflow_id, mode, status, phases_completed || '/' || phases_total as progress, started_at
            FROM workflow_runs ORDER BY started_at DESC LIMIT 20;
        "
    fi
}

# --- MAIN ---

case "$CMD" in
    sync) sync ;;
    list) list ;;
    run) run "$@" ;;
    status) status "$@" ;;
    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║     🔗 WORKFLOW RUNNER — Multi-Agent Workflow Orchestrator║
╚══════════════════════════════════════════════════════════╝

COMMANDS:
  workflow-run.sh sync                          Sync YAML configs to SQLite
  workflow-run.sh list                          List available workflows
  workflow-run.sh run <wf> [opts] "<goal>"      Execute a workflow
  workflow-run.sh status [run-id]               Show workflow run status

RUN OPTIONS:
  --mode <mode>         Workflow mode (simple, complex, investigate, etc.)
  --skip-gates          Skip human approval gates
  --dry-run             Show what would happen without executing
  --session <id>        Link to existing session

EXAMPLES:
  workflow-run.sh sync
  workflow-run.sh list
  workflow-run.sh run rpi --mode complex "Add OAuth support"
  workflow-run.sh run bugfix --mode known "Fix null pointer in auth"
  workflow-run.sh run hotfix "Production timeout on /api/users"
  workflow-run.sh run refactor --dry-run --mode standard "Clean up auth module"
  workflow-run.sh status wfr_abc123
HELP
        ;;
esac
