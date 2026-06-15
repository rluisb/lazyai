#!/usr/bin/env bash
# extract-dispatch-pairs.sh — Extract dispatch pairs from session DB into JSONL
# Usage: ./scripts/extract-dispatch-pairs.sh [--output PATH] [--limit N] [--db PATH]
# Default DB: .specify/session.db
# Default output: stdout

set -euo pipefail

# --- DEFAULTS ---
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"
OUTPUT_PATH=""
LIMIT=""

# --- STATIC AGENT LIST (from AGENTS.md) ---
AVAILABLE_AGENTS='["loop-driver","engine-control","loot-hawk","turbo-crank","wall-builder","shield-audit","rift-deploy","respawn-crew"]'

# --- PARSE ARGS ---
while [[ $# -gt 0 ]]; do
    case "$1" in
        --output)
            OUTPUT_PATH="$2"
            shift 2
            ;;
        --limit)
            LIMIT="$2"
            shift 2
            ;;
        --db)
            DB_PATH="$2"
            shift 2
            ;;
        --help|-h)
            cat <<'EOF'
Usage: extract-dispatch-pairs.sh [OPTIONS]

Extract dispatch pairs from the session database into JSONL format for training/eval.

Options:
  --output PATH   Write JSONL to PATH instead of stdout
  --limit N       Limit output to N rows
  --db PATH       Use PATH as the SQLite database (default: .specify/session.db)
  --help, -h      Show this help message

Each output line is a JSON object matching the dispatch pair schema:
  {
    "input": { "user_goal": "...", "context": "...", "available_agents": [...] },
    "output": { "selected_agent": "...", "confidence": float|null, "reasoning": "..."|null },
    "metadata": { "session_id": "...", "timestamp": "...", "human_accepted": bool|null }
  }
EOF
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            echo "Use --help for usage information." >&2
            exit 1
            ;;
    esac
done

# --- VALIDATE DB ---
if [[ ! -f "$DB_PATH" ]]; then
    echo "Error: Database not found: $DB_PATH" >&2
    exit 1
fi

# --- BUILD SQL ---
SQL="
SELECT
    d.id AS dispatch_id,
    d.session_id,
    d.from_agent,
    d.to_agent AS selected_agent,
    d.task_type,
    d.mode,
    d.risk,
    d.reason,
    d.created_at AS timestamp,
    COALESCE(s.goal, '') AS user_goal,
    COALESCE(s.repo, '') AS repo,
    COALESCE(s.model, '') AS model,
    json_extract(d.metadata_json, '$.confidence') AS confidence,
    json_extract(d.metadata_json, '$.reasoning') AS reasoning,
    json_extract(d.metadata_json, '$.human_accepted') AS human_accepted
FROM agent_dispatches d
LEFT JOIN sessions s ON d.session_id = s.id
"

if [[ -n "$LIMIT" ]]; then
    SQL="$SQL LIMIT $LIMIT"
fi

# --- JSON FORMATTER ---
json_fmt() {
    if command -v jq &>/dev/null; then
        jq -c .
    elif command -v python3 &>/dev/null; then
        python3 -c 'import sys, json; [print(json.dumps(json.loads(line), ensure_ascii=False, separators=(",", ":"))) for line in sys.stdin]'
    else
        cat
    fi
}

# --- OUTPUT HANDLER ---
output() {
    if [[ -n "$OUTPUT_PATH" ]]; then
        cat > "$OUTPUT_PATH"
    else
        cat
    fi
}

# --- MAIN QUERY & TRANSFORM ---
RAW_JSON=$(sqlite3 -json "$DB_PATH" "$SQL" 2>&1) || {
    echo "Error: Failed to query database: $DB_PATH" >&2
    exit 1
}

if [[ -z "$RAW_JSON" ]]; then
    echo "✅ Extracted 0 dispatch pair(s)" >&2
    exit 0
fi

# Count rows for the summary
ROW_COUNT=$(echo "$RAW_JSON" | python3 -c 'import sys, json; rows=json.load(sys.stdin); print(len(rows) if isinstance(rows, list) else 1)')

# Transform JSON array into dispatch pair JSONL
echo "$RAW_JSON" | python3 -c "
import sys, json

rows = json.load(sys.stdin)
if not isinstance(rows, list):
    rows = [rows]

for row in rows:
    # Build context string from available fields
    ctx_parts = []
    if row.get('repo'):
        ctx_parts.append(f\"repo={row['repo']}\")
    if row.get('model'):
        ctx_parts.append(f\"model={row['model']}\")
    if row.get('task_type'):
        ctx_parts.append(f\"task={row['task_type']}\")
    if row.get('mode'):
        ctx_parts.append(f\"mode={row['mode']}\")
    if row.get('risk'):
        ctx_parts.append(f\"risk={row['risk']}\")
    if row.get('from_agent'):
        ctx_parts.append(f\"from={row['from_agent']}\")
    context = '; '.join(ctx_parts)

    # Handle confidence
    confidence = row.get('confidence')
    if confidence is not None:
        try:
            confidence = float(confidence)
        except (ValueError, TypeError):
            confidence = None

    # Handle reasoning
    reasoning = row.get('reasoning')
    if reasoning is not None and str(reasoning).strip() == '':
        reasoning = None

    # Handle human_accepted
    human_accepted = row.get('human_accepted')
    if human_accepted is not None:
        if isinstance(human_accepted, bool):
            pass
        elif str(human_accepted).lower() in ('1', 'true', 'yes'):
            human_accepted = True
        elif str(human_accepted).lower() in ('0', 'false', 'no'):
            human_accepted = False
        else:
            human_accepted = None

    record = {
        'input': {
            'user_goal': row.get('user_goal') or '',
            'context': context,
            'available_agents': json.loads('$AVAILABLE_AGENTS')
        },
        'output': {
            'selected_agent': row.get('selected_agent') or '',
            'confidence': confidence,
            'reasoning': reasoning
        },
        'metadata': {
            'session_id': row.get('session_id') or '',
            'timestamp': row.get('timestamp') or '',
            'human_accepted': human_accepted
        }
    }
    print(json.dumps(record, ensure_ascii=False, separators=(',', ':')))
" | json_fmt | output

echo "✅ Extracted ${ROW_COUNT} dispatch pair(s)" >&2
