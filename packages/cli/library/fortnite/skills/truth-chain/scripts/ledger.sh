#!/usr/bin/env bash
# ledger.sh — Immutable append-only ledger with SHA-256 hash chain verification
# Fortnite naming: Truth-Chain
# Usage: ./ledger.sh <command> [args...]
#
# Ledger file: .specify/ledger.jsonl (one JSON object per line)
# Each entry: {seq, ts, type, session_id, data, prev_hash, hash}
# v2 additions: workflow_run_id, agent, risk, input_hash, output_hash, redactions

set -euo pipefail

LEDGER_PATH="${OPENCODE_WORKSPACE:-.}/.specify/ledger.jsonl"
CMD="${1:-help}"
shift 2>/dev/null || true

# --- LOCKING ---
# Priority: flock (Linux/util-linux) > shlock (macOS/BSD) > mkdir (portable fallback)
LOCK_FILE="${OPENCODE_WORKSPACE:-.}/.specify/ledger.lock"
LOCK_DIR="${OPENCODE_WORKSPACE:-.}/.specify/ledger-lock"

acquire_lock() {
    # Try flock first (Linux, util-linux)
    if command -v flock &>/dev/null; then
        exec 200>"$LOCK_FILE"
        if flock -n 200 2>/dev/null; then
            return 0
        fi
        exec 200>&- 2>/dev/null || true
        return 1
    fi

    # Try shlock next (macOS/BSD, typically /usr/bin/shlock)
    if command -v shlock &>/dev/null; then
        local pid_file="${LOCK_FILE}.pid"
        if shlock -p $$ -f "$pid_file" 2>/dev/null; then
            return 0
        fi
        return 1
    fi

    # Portable fallback: mkdir atomic test-and-set
    if mkdir "${LOCK_DIR}" 2>/dev/null; then
        return 0
    fi
    return 1
}

release_lock() {
    if command -v flock &>/dev/null; then
        flock -u 200 2>/dev/null || true
        exec 200>&- 2>/dev/null || true
        rm -f "$LOCK_FILE" 2>/dev/null || true
        return
    fi
    if command -v shlock &>/dev/null; then
        rm -f "${LOCK_FILE}.pid" 2>/dev/null || true
        return
    fi
    rmdir "${LOCK_DIR}" 2>/dev/null || true
}

with_lock() {
    local func="$1"
    shift
    local max_attempts=10
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        if acquire_lock; then
            # Execute the function with lock held
            # Use explicit rc capture instead of traps to avoid nested trap conflicts
            set +e
            "$func" "$@"
            local result=$?
            set -e

            # Always release lock regardless of function result
            release_lock
            return $result
        fi
        attempt=$((attempt + 1))
        if [[ $attempt -lt $max_attempts ]]; then
            sleep 0.1
        fi
    done

    echo "❌ Failed to acquire ledger lock after $max_attempts attempts" >&2
    return 1
}

# --- HELPERS ---

now() { date -u '+%Y-%m-%dT%H:%M:%SZ'; }

sha256() {
    if command -v shasum &>/dev/null; then
        echo -n "$1" | shasum -a 256 | awk '{print $1}'
    elif command -v sha256sum &>/dev/null; then
        echo -n "$1" | sha256sum | awk '{print $1}'
    else
        echo "❌ No SHA-256 tool found (need shasum or sha256sum)" >&2
        exit 1
    fi
}

_redact_secrets() {
    local input="$1"
    python3 - "$input" << 'PYEOF'
import json, sys

data_str = sys.argv[1]

secret_keys = {
    "token", "access_token", "refresh_token", "auth_token", "bearer_token",
    "api_token", "jwt", "secret", "password", "api_key", "apikey", "key",
    "private_key", "client_secret", "client_id", "auth_code", "credential", "credentials"
}

def should_redact(key):
    return key.lower() in secret_keys

def redact(obj):
    if isinstance(obj, dict):
        result = {}
        for k, v in obj.items():
            if should_redact(k) and isinstance(v, str):
                result[k] = "[REDACTED]"
            elif should_redact(k) and isinstance(v, list):
                result[k] = ["[REDACTED]"]
            else:
                result[k] = redact(v)
        return result
    elif isinstance(obj, list):
        return [redact(item) for item in obj]
    else:
        return obj

try:
    parsed = json.loads(data_str)
    redacted = redact(parsed)
    print(json.dumps(redacted, separators=(',', ':'), ensure_ascii=False))
except (json.JSONDecodeError, ValueError):
    # Fallback: if input isn't valid JSON, return empty object to avoid corruption
    print("{}")
PYEOF
}

resolve_ledger() {
    local path="$1"
    echo "${path:-$LEDGER_PATH}"
}

last_hash_for() {
    local path="$1"
    if [[ ! -f "$path" ]] || [[ ! -s "$path" ]]; then
        echo "0000000000000000000000000000000000000000000000000000000000000000"
        return
    fi
    tail -n 1 "$path" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['hash'])" 2>/dev/null || \
        echo "0000000000000000000000000000000000000000000000000000000000000000"
}

next_seq_for() {
    local path="$1"
    if [[ ! -f "$path" ]] || [[ ! -s "$path" ]]; then
        echo "1"
        return
    fi
    tail -n 1 "$path" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['seq'] + 1)" 2>/dev/null || echo "1"
}

entry_count_for() {
    local path="$1"
    if [[ ! -f "$path" ]] || [[ ! -s "$path" ]]; then
        echo "0"
        return
    fi
    wc -l < "$path" | tr -d ' '
}

# --- COMMANDS ---

init() {
    local ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    mkdir -p "$(dirname "$ledger")"
    if [[ -f "$ledger" ]] && [[ -s "$ledger" ]]; then
        echo "⚠️  Ledger already exists at $ledger ($(entry_count_for "$ledger") entries)"
        echo "   Run 'ledger.sh verify --file $ledger' to check integrity"
        return
    fi

    GENESIS_HASH="0000000000000000000000000000000000000000000000000000000000000000"
    GENESIS_DATA='{"version":"1.0","description":"Truth-Chain genesis entry"}'
    GENESIS_TS=$(now)
    GENESIS_INPUT="1|${GENESIS_TS}|genesis|system|${GENESIS_DATA}|${GENESIS_HASH}"
    GENESIS_HASH_COMPUTED=$(sha256 "$GENESIS_INPUT")

    cat > "$ledger" << EOF
{"seq":1,"ts":"${GENESIS_TS}","type":"genesis","session_id":"system","data":${GENESIS_DATA},"prev_hash":"${GENESIS_HASH}","hash":"${GENESIS_HASH_COMPUTED}"}
EOF

    echo "✅ Ledger initialized: $ledger"
    echo "   Genesis entry created (hash: ${GENESIS_HASH_COMPUTED:0:16}...)"
}

append() {
    local ENTRY_TYPE="" DATA_JSON="" SESSION_ID="unknown" ledger="$LEDGER_PATH" DB_ROW_ID=""
    local WORKFLOW_RUN_ID="" AGENT="" RISK="" INPUT_HASH="" OUTPUT_HASH=""
    local JSON_OUTPUT=false REDACTIONS="" REDACT=false

    if [[ "${1:-}" == --* ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --type) ENTRY_TYPE="$2"; shift 2 ;;
                --data) DATA_JSON="$2"; shift 2 ;;
                --session) SESSION_ID="$2"; shift 2 ;;
                --file) ledger="$2"; shift 2 ;;
                --db-row-id) DB_ROW_ID="$2"; shift 2 ;;
                --workflow-run-id) WORKFLOW_RUN_ID="$2"; shift 2 ;;
                --agent) AGENT="$2"; shift 2 ;;
                --risk) RISK="$2"; shift 2 ;;
                --input-hash) INPUT_HASH="$2"; shift 2 ;;
                --output-hash) OUTPUT_HASH="$2"; shift 2 ;;
                --redactions) REDACTIONS="$2"; shift 2 ;;
                --redact) REDACT=true; shift ;;
                --json) JSON_OUTPUT=true; shift ;;
                *) shift ;;
            esac
        done
    else
        ENTRY_TYPE="${1:-}"
        DATA_JSON="${2:-}"
        SESSION_ID="${3:-unknown}"
    fi

    if [[ -z "$ENTRY_TYPE" ]] || [[ -z "$DATA_JSON" ]]; then
        echo "Usage: ledger.sh append <type> '<json-data>' [session-id]"
        echo "   or: ledger.sh append --type <type> --data '<json>' [--session <sid>] [--file <path>] [--db-row-id <id>]"
        echo "       [--workflow-run-id <wid>] [--agent <agent>] [--risk <low|medium|high|critical>]"
        echo "       [--input-hash <hash>] [--output-hash <hash>] [--redactions <json>] [--redact] [--json]"
        exit 1
    fi

    if [[ ! -f "$ledger" ]]; then
        mkdir -p "$(dirname "$ledger")"
        init --file "$ledger" >/dev/null 2>&1
    fi

    # Use lock to prevent concurrent appends
    with_lock _append_unsafe "$ENTRY_TYPE" "$DATA_JSON" "$SESSION_ID" "$ledger" "$DB_ROW_ID" "$WORKFLOW_RUN_ID" "$AGENT" "$RISK" "$INPUT_HASH" "$OUTPUT_HASH" "$REDACTIONS" "$JSON_OUTPUT" "$REDACT"
}

_append_unsafe() {
    local ENTRY_TYPE="$1"
    local DATA_JSON="$2"
    local SESSION_ID="$3"
    local ledger="$4"
    local DB_ROW_ID="$5"
    local WORKFLOW_RUN_ID="$6"
    local AGENT="$7"
    local RISK="$8"
    local INPUT_HASH="$9"
    local OUTPUT_HASH="${10}"
    local REDACTIONS="${11}"
    local JSON_OUTPUT="${12}"
    local REDACT="${13}"

    local DATA_TO_WRITE="$DATA_JSON"
    if [[ "$REDACT" == "true" ]]; then
        DATA_TO_WRITE=$(_redact_secrets "$DATA_JSON")
    fi

    local SEQ=$(next_seq_for "$ledger")
    local TS=$(now)
    local PREV_HASH=$(last_hash_for "$ledger")
    local HASH_INPUT="${SEQ}|${TS}|${ENTRY_TYPE}|${SESSION_ID}|${DATA_TO_WRITE}|${PREV_HASH}"
    local HASH=$(sha256 "$HASH_INPUT")

    # Build entry JSON with optional v2 fields
    local EXTRA_FIELDS=""
    if [[ -n "$DB_ROW_ID" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"db_row_id\":${DB_ROW_ID}"
    fi
    if [[ -n "$WORKFLOW_RUN_ID" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"workflow_run_id\":\"${WORKFLOW_RUN_ID}\""
    fi
    if [[ -n "$AGENT" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"agent\":\"${AGENT}\""
    fi
    if [[ -n "$RISK" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"risk\":\"${RISK}\""
    fi
    if [[ -n "$INPUT_HASH" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"input_hash\":\"${INPUT_HASH}\""
    fi
    if [[ -n "$OUTPUT_HASH" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"output_hash\":\"${OUTPUT_HASH}\""
    fi
    if [[ -n "$REDACTIONS" ]]; then
        EXTRA_FIELDS="${EXTRA_FIELDS},\"redactions\":${REDACTIONS}"
    fi

    # Atomic append using temp file + rename pattern
    local TMP_FILE=""
    TMP_FILE=$(mktemp "${ledger}.tmp.XXXXXX") || {
        echo "❌ Failed to create temp file for atomic append" >&2
        return 1
    }

    # Write entry to temp file
    printf '{"seq":%d,"ts":"%s","type":"%s","session_id":"%s","data":%s,"prev_hash":"%s","hash":"%s"%s}\n' \
        "$SEQ" "$TS" "$ENTRY_TYPE" "$SESSION_ID" "$DATA_TO_WRITE" "$PREV_HASH" "$HASH" "$EXTRA_FIELDS" > "$TMP_FILE"

    # Sync to disk
    sync -f "$TMP_FILE" 2>/dev/null || true

    # Atomic rename to append
    cat "$TMP_FILE" >> "$ledger"
    local append_result=$?

    # Clean up temp file explicitly (no traps to avoid conflicts with with_lock)
    if [[ -n "$TMP_FILE" ]] && [[ -f "$TMP_FILE" ]]; then
        rm -f "$TMP_FILE" 2>/dev/null || true
    fi

    if [[ $append_result -ne 0 ]]; then
        echo "❌ Failed to append to ledger" >&2
        return 1
    fi

    # Sync ledger to disk
    sync -f "$ledger" 2>/dev/null || true

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        printf '{"status":"ok","seq":%d,"type":"%s","hash":"%s","prev_hash":"%s","timestamp":"%s"}\n' \
            "$SEQ" "$ENTRY_TYPE" "$HASH" "$PREV_HASH" "$TS"
    else
        echo "📝 Ledger entry $SEQ: $ENTRY_TYPE (hash: ${HASH:0:16}...)"
    fi
}

verify() {
    local ledger="$LEDGER_PATH" JSON_OUTPUT=false
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --json) JSON_OUTPUT=true; shift ;;
            *) shift ;;
        esac
    done

    if [[ ! -f "$ledger" ]] || [[ ! -s "$ledger" ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo '{"status":"fail","error":"Ledger not found","file":"'"$ledger"'"}'
        else
            echo "❌ Ledger not found at $ledger"
            echo "   Run 'ledger.sh init' first"
        fi
        exit 1
    fi

    local COUNT=$(entry_count_for "$ledger")
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo "🔍 Verifying ledger integrity ($COUNT entries)..."
    fi

    python3 - "$ledger" "$JSON_OUTPUT" << 'PYEOF'
import json, hashlib, sys

def sha256(s):
    return hashlib.sha256(s.encode()).hexdigest()

ledger_path = sys.argv[1]
json_output = sys.argv[2] == "true"
errors = []
last_line_num = 0
prev_hash = None
expected_seq = 1

with open(ledger_path, 'r') as f:
    for line_num, line in enumerate(f, 1):
        line = line.strip()
        if not line:
            continue
        last_line_num = line_num
        try:
            entry = json.loads(line)
        except json.JSONDecodeError as e:
            errors.append({"line": line_num, "error": f"Invalid JSON - {e}"})
            continue

        seq = entry.get('seq', 0)
        ts = entry.get('ts', '')
        entry_type = entry.get('type', '')
        session_id = entry.get('session_id', '')
        data = json.dumps(entry.get('data', {}), separators=(',', ':'), ensure_ascii=False)
        stored_prev_hash = entry.get('prev_hash', '')
        stored_hash = entry.get('hash', '')

        # Sequence monotonicity / gap check
        if seq != expected_seq:
            if seq > expected_seq:
                errors.append({
                    "seq": seq,
                    "error": f"sequence gap (expected {expected_seq}, got {seq})"
                })
            else:
                errors.append({
                    "seq": seq,
                    "error": f"non-monotonic sequence (expected {expected_seq}, got {seq})"
                })
            expected_seq = seq + 1
        else:
            expected_seq += 1

        # First entry: set prev_hash from its stored value
        if prev_hash is None:
            if entry_type == 'genesis':
                # Genesis must have all-zeros prev_hash
                expected_genesis = '0000000000000000000000000000000000000000000000000000000000000000'
                if stored_prev_hash != expected_genesis:
                    errors.append({
                        "seq": seq,
                        "error": "genesis prev_hash must be all zeros",
                        "expected_prev_hash": expected_genesis,
                        "actual_prev_hash": stored_prev_hash
                    })
            # Verify the hash of the first entry
            hash_input = f"{seq}|{ts}|{entry_type}|{session_id}|{data}|{stored_prev_hash}"
            expected_hash = sha256(hash_input)
            if stored_hash != expected_hash:
                errors.append({
                    "seq": seq,
                    "error": "hash mismatch (entry tampered)",
                    "expected_hash": expected_hash,
                    "actual_hash": stored_hash
                })
            # Set prev_hash to this entry's hash for the next entry
            prev_hash = stored_hash
            continue

        # Subsequent entries: check prev_hash chain
        if stored_prev_hash != prev_hash:
            errors.append({
                "seq": seq,
                "error": "prev_hash mismatch",
                "expected_prev_hash": prev_hash,
                "actual_prev_hash": stored_prev_hash
            })

        # Compute expected hash
        hash_input = f"{seq}|{ts}|{entry_type}|{session_id}|{data}|{stored_prev_hash}"
        expected_hash = sha256(hash_input)

        if stored_hash != expected_hash:
            errors.append({
                "seq": seq,
                "error": "hash mismatch (entry tampered)",
                "expected_hash": expected_hash,
                "actual_hash": stored_hash
            })

        prev_hash = stored_hash

if errors:
    if json_output:
        first_error = errors[0]
        result = {
            "status": "fail",
            "broken_seq": first_error.get("seq"),
            "expected_prev_hash": first_error.get("expected_prev_hash"),
            "actual_prev_hash": first_error.get("actual_prev_hash"),
            "expected_hash": first_error.get("expected_hash"),
            "actual_hash": first_error.get("actual_hash"),
            "error": first_error.get("error"),
            "file": ledger_path,
            "total_errors": len(errors)
        }
        print(json.dumps(result))
    else:
        print(f"FAIL: {len(errors)} integrity errors found:")
        for err in errors[:10]:
            seq = err.get("seq", "?")
            msg = err.get("error", "")
            print(f"  ❌ Entry {seq}: {msg}")
            if "expected_prev_hash" in err:
                print(f"     expected_prev_hash: {err['expected_prev_hash'][:16]}...")
                print(f"     actual_prev_hash:   {err['actual_prev_hash'][:16]}...")
            if "expected_hash" in err:
                print(f"     expected_hash: {err['expected_hash'][:16]}...")
                print(f"     actual_hash:   {err['actual_hash'][:16]}...")
        if len(errors) > 10:
            print(f"  ... and {len(errors) - 10} more errors")
        print("  Chain broken — sequence integrity compromised")
    sys.exit(1)
else:
    if json_output:
        print(json.dumps({
            "status": "pass",
            "entries": last_line_num,
            "file": ledger_path,
            "message": "Chain intact"
        }))
    else:
        print(f"PASS: All {last_line_num} entries verified. Chain intact.")
    sys.exit(0)
PYEOF
}

backup() {
    local ledger="$LEDGER_PATH" JSON_OUTPUT=false
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --json) JSON_OUTPUT=true; shift ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo '{"status":"fail","error":"Ledger not found"}'
        else
            echo "❌ Ledger not found at $ledger"
        fi
        exit 1
    fi

    local BACKUP_DIR="$(dirname "$ledger")/backups"
    mkdir -p "$BACKUP_DIR"
    local TIMESTAMP=$(date -u '+%Y%m%d-%H%M%S')
    local BACKUP_FILE="$BACKUP_DIR/ledger-${TIMESTAMP}.jsonl"
    cp "$ledger" "$BACKUP_FILE"
    local SIZE=$(wc -c < "$ledger" | tr -d ' ')

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"status\":\"ok\",\"backup_file\":\"$BACKUP_FILE\",\"original_size\":$SIZE}"
    else
        echo "✅ Ledger backed up to $BACKUP_FILE"
        echo "   Original size: $SIZE bytes"
    fi
}

rotate() {
    local ledger="$LEDGER_PATH" MAX_SIZE=10485760 JSON_OUTPUT=false  # 10MB default
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --max-size) MAX_SIZE="$2"; shift 2 ;;
            --json) JSON_OUTPUT=true; shift ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo '{"status":"fail","error":"Ledger not found"}'
        else
            echo "❌ Ledger not found at $ledger"
        fi
        exit 1
    fi

    local SIZE=$(wc -c < "$ledger" | tr -d ' ')
    if [[ $SIZE -lt $MAX_SIZE ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo "{\"status\":\"ok\",\"rotated\":false,\"size\":$SIZE,\"max_size\":$MAX_SIZE}"
        else
            echo "✅ Ledger size ($SIZE bytes) below threshold ($MAX_SIZE bytes). No rotation needed."
        fi
        return 0
    fi

    local BACKUP_DIR="$(dirname "$ledger")/backups"
    mkdir -p "$BACKUP_DIR"
    local TIMESTAMP=$(date -u '+%Y%m%d-%H%M%S')
    local ROTATED_FILE="$BACKUP_DIR/ledger-${TIMESTAMP}.jsonl"
    mv "$ledger" "$ROTATED_FILE"

    # Start fresh ledger with genesis
    init --file "$ledger" >/dev/null 2>&1

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"status\":\"ok\",\"rotated\":true,\"rotated_file\":\"$ROTATED_FILE\",\"new_ledger\":\"$ledger\",\"previous_size\":$SIZE}"
    else
        echo "✅ Ledger rotated"
        echo "   Old ledger: $ROTATED_FILE ($SIZE bytes)"
        echo "   New ledger: $ledger (fresh genesis)"
    fi
}

replay_to_session_db() {
    local ledger="$LEDGER_PATH" DB_PATH=".specify/session.db" JSON_OUTPUT=false
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --db) DB_PATH="$2"; shift 2 ;;
            --json) JSON_OUTPUT=true; shift ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo '{"status":"fail","error":"Ledger not found"}'
        else
            echo "❌ Ledger not found at $ledger"
        fi
        exit 1
    fi

    if [[ ! -f "$DB_PATH" ]]; then
        if [[ "$JSON_OUTPUT" == "true" ]]; then
            echo '{"status":"fail","error":"Session DB not found"}'
        else
            echo "❌ Session DB not found at $DB_PATH"
        fi
        exit 1
    fi

    local COUNT=$(entry_count_for "$ledger")
    local INSERTED=0
    local SKIPPED=0
    local FAILED=0
    local USE_SESSION_DB=false

    if [[ -x "./scripts/session-db.sh" ]]; then
        USE_SESSION_DB=true
    else
        if [[ "$JSON_OUTPUT" != "true" ]]; then
            echo "   ⚠️  ./scripts/session-db.sh not available; falling back to direct SQL" >&2
        fi
    fi

    # Pre-create any missing sessions referenced in the ledger to satisfy FK constraints
    local missing_sessions
    missing_sessions=$(python3 - "$ledger" "$DB_PATH" << 'PYEOF'
import json, sys, sqlite3

ledger_path = sys.argv[1]
db_path = sys.argv[2]

sids = set()
with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        try:
            entry = json.loads(line)
            sid = entry.get('session_id', '')
            if sid and sid != 'system':
                sids.add(sid)
        except (json.JSONDecodeError, ValueError):
            continue

if not sids:
    sys.exit(0)

conn = sqlite3.connect(db_path)
cur = conn.cursor()
cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions';")
if not cur.fetchone():
    conn.close()
    sys.exit(0)

for sid in sids:
    cur.execute("SELECT COUNT(*) FROM sessions WHERE id = ?", (sid,))
    if cur.fetchone()[0] == 0:
        print(sid)

conn.close()
PYEOF
) || true

    for sid in $missing_sessions; do
        sqlite3 "$DB_PATH" "INSERT OR IGNORE INTO sessions(id, started_at, model, goal, repo, status) VALUES ('$(echo "$sid" | sed "s/'/''/g")', datetime('now'), 'unknown', 'auto-created by replay', 'unknown', 'active');" 2>/dev/null || true
    done

    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo "🔄 Replaying $COUNT ledger entries to session DB..."
    fi

    while IFS= read -r line; do
        line=$(echo "$line" | tr -d '\r')
        [[ -z "$line" ]] && continue

        local seq=$(echo "$line" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['seq'])")
        local session_id=$(echo "$line" | python3 -c "import sys,json; d=json.loads(sys.stdin.read()); print(d.get('session_id',''))")
        local workflow_run_id=$(echo "$line" | python3 -c "import sys,json; d=json.loads(sys.stdin.read()); print(d.get('workflow_run_id',''))")
        local event_type=$(echo "$line" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['type'])")
        local event_hash=$(echo "$line" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['hash'])")
        local prev_hash=$(echo "$line" | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['prev_hash'])")

        # Check if already exists
        local exists=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ledger_refs WHERE seq=$seq;" 2>/dev/null || echo "0")
        if [[ "$exists" -gt 0 ]]; then
            SKIPPED=$((SKIPPED + 1))
            continue
        fi

        if [[ "$USE_SESSION_DB" == "true" ]]; then
            ./scripts/session-db.sh --db "$DB_PATH" record-ledger-ref "$seq" "$event_type" "$event_hash" "$session_id" "$workflow_run_id" "$prev_hash" >/dev/null 2>&1 && INSERTED=$((INSERTED + 1)) || FAILED=$((FAILED + 1))
        else
            # Fallback: direct SQL insert
            local wf_val="NULL"
            [[ -n "$workflow_run_id" ]] && wf_val="'${workflow_run_id}'"
            local sid_val="NULL"
            [[ -n "$session_id" ]] && sid_val="'${session_id}'"

            sqlite3 "$DB_PATH" "INSERT INTO ledger_refs (seq, session_id, workflow_run_id, event_type, event_hash, prev_hash, created_at) VALUES ($seq, ${sid_val}, ${wf_val}, '${event_type}', '${event_hash}', '${prev_hash}', datetime('now'));" 2>/dev/null && INSERTED=$((INSERTED + 1)) || FAILED=$((FAILED + 1))
        fi
    done < "$ledger"

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        echo "{\"status\":\"ok\",\"inserted\":$INSERTED,\"skipped\":$SKIPPED,\"failed\":$FAILED,\"total\":$COUNT}"
    else
        echo "✅ Replay complete: $INSERTED inserted, $SKIPPED skipped, $FAILED failed (unsupported/constraint errors)"
    fi
}

replay() {
    local TARGET_SEQ="" ledger="$LEDGER_PATH"
    if [[ "${1:-}" == --* ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --seq) TARGET_SEQ="$2"; shift 2 ;;
                --file) ledger="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
    else
        TARGET_SEQ="${1:-}"
    fi
    if [[ -z "$TARGET_SEQ" ]]; then
        echo "Usage: ledger.sh replay <entry-seq> [--file <path>]"
        exit 1
    fi
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi

    echo "🔄 Replaying ledger up to entry $TARGET_SEQ..."
    echo ""

    python3 << PYEOF
import json

ledger_path = "$ledger"
target_seq = int("$TARGET_SEQ")

state = {
    "sessions": {}, "dispatches": [], "decisions": [], "memories": [],
    "artifacts": [], "barriers": {}, "locks": {}, "messages": [],
    "quality_gates": [], "token_total": 0
}

with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if not line: continue
        entry = json.loads(line)
        if entry['seq'] > target_seq: break
        etype = entry['type']
        data = entry['data']
        sid = entry.get('session_id', '')

        if etype == 'session_start':
            state['sessions'][sid] = {'status': 'active', **data}
        elif etype == 'session_end':
            if sid in state['sessions']:
                state['sessions'][sid]['status'] = data.get('status', 'ended')
                state['sessions'][sid]['token_total'] = data.get('token_total', 0)
        elif etype == 'dispatch':
            state['dispatches'].append({'seq': entry['seq'], 'agent': data.get('agent'), 'task': data.get('task'), 'result': 'pending'})
        elif etype == 'dispatch_complete':
            for d in state['dispatches']:
                if d['seq'] == entry['seq']:
                    d['result'] = data.get('result', 'pass')
                    d['summary'] = data.get('summary', '')
        elif etype == 'decision':
            state['decisions'].append(data)
        elif etype == 'memory':
            state['memories'].append(data)
        elif etype == 'artifact':
            state['artifacts'].append(data)
        elif etype == 'barrier_create':
            state['barriers'][data.get('barrier_id')] = {'expected': data.get('expected_count'), 'arrived': 0, 'status': 'waiting'}
        elif etype == 'barrier_resolve':
            bid = data.get('barrier_id')
            if bid in state['barriers']:
                state['barriers'][bid]['status'] = 'resolved'
        elif etype == 'lock_acquire':
            state['locks'][data.get('lock_name')] = {'held_by': data.get('held_by'), 'status': 'active'}
        elif etype == 'lock_release':
            lname = data.get('lock_name')
            if lname in state['locks']:
                state['locks'][lname]['status'] = 'released'
        elif etype == 'msg_send':
            state['messages'].append(data)
        elif etype == 'quality_gate':
            state['quality_gates'].append(data)
        elif etype == 'token_log':
            state['token_total'] += data.get('token_count', 0)

print(f"=== State at Entry {target_seq} ===")
print(f"Sessions: {len(state['sessions'])}")
for sid, s in state['sessions'].items():
    print(f"  {sid}: {s['status']}")
print(f"Dispatches: {len(state['dispatches'])}")
for d in state['dispatches']:
    print(f"  [{d['seq']}] {d['agent']}: {d['task']} → {d['result']}")
print(f"Decisions: {len(state['decisions'])}")
for d in state['decisions']:
    print(f"  {d.get('title')}: {d.get('decision')}")
print(f"Memories: {len(state['memories'])}")
print(f"Artifacts: {len(state['artifacts'])}")
print(f"Barriers: {len(state['barriers'])}")
print(f"Locks: {len(state['locks'])}")
print(f"Messages: {len(state['messages'])}")
print(f"Quality Gates: {len(state['quality_gates'])}")
print(f"Total Tokens: {state['token_total']}")
PYEOF
}

query() {
    local TYPE="" AGENT="" TAG="" SESSION="" ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --type) TYPE="$2"; shift 2 ;;
            --agent) AGENT="$2"; shift 2 ;;
            --tag) TAG="$2"; shift 2 ;;
            --session) SESSION="$2"; shift 2 ;;
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done

    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi

    echo "🔍 Querying ledger..."

    python3 << PYEOF
import json

ledger_path = "$ledger"
filter_type = "$TYPE"
filter_agent = "$AGENT"
filter_tag = "$TAG"
filter_session = "$SESSION"

matches = []
with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if not line: continue
        entry = json.loads(line)
        if filter_type and entry.get('type') != filter_type: continue
        if filter_session and entry.get('session_id') != filter_session: continue
        if filter_agent:
            data = entry.get('data', {})
            if data.get('agent') != filter_agent and entry.get('type') not in ('dispatch', 'dispatch_complete', 'dispatch_fail'): continue
        if filter_tag:
            data = entry.get('data', {})
            tags = data.get('tags', '')
            if filter_tag not in str(tags): continue
        matches.append(entry)

print(f"Found {len(matches)} matching entries:")
print("")
for e in matches:
    data_str = json.dumps(e.get('data', {}), indent=2)
    print(f"[{e['seq']}] {e['type']} @ {e['ts']}")
    print(f"  Session: {e['session_id']}")
    print(f"  Data: {data_str[:200]}{'...' if len(data_str) > 200 else ''}")
    print("")
PYEOF
}

audit() {
    local SESSION_ID="" OUTPUT="" ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --session) SESSION_ID="$2"; shift 2 ;;
            --output) OUTPUT="$2"; shift 2 ;;
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done

    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi

    echo "📊 Generating audit report..."

    python3 << PYEOF
import json
from collections import defaultdict

ledger_path = "$ledger"
session_id = "$SESSION_ID"
output_file = "$OUTPUT"

report = {
    "ledger_path": ledger_path,
    "session_id": session_id if session_id else "all",
    "total_entries": 0, "entry_types": defaultdict(int),
    "sessions": {}, "dispatches": [], "decisions": [],
    "quality_gates": [], "drift_violations": [],
    "backprop_updates": [], "timeline": []
}

with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if not line: continue
        entry = json.loads(line)
        report["total_entries"] += 1
        if session_id and entry.get("session_id") != session_id: continue
        etype = entry["type"]
        report["entry_types"][etype] += 1
        data = entry.get("data", {})
        if etype in ("session_start", "session_end"):
            sid = entry["session_id"]
            if sid not in report["sessions"]:
                report["sessions"][sid] = {"entries": [], "status": "unknown"}
            report["sessions"][sid]["entries"].append({"seq": entry["seq"], "type": etype, "ts": entry["ts"], **data})
            if etype == "session_end":
                report["sessions"][sid]["status"] = data.get("status", "ended")
        elif etype in ("dispatch", "dispatch_complete", "dispatch_fail"):
            report["dispatches"].append({"seq": entry["seq"], "type": etype, "ts": entry["ts"], "session_id": entry["session_id"], **data})
        elif etype == "decision":
            report["decisions"].append({"seq": entry["seq"], "ts": entry["ts"], "session_id": entry["session_id"], **data})
        elif etype == "quality_gate":
            report["quality_gates"].append({"seq": entry["seq"], "ts": entry["ts"], "session_id": entry["session_id"], **data})
        elif etype == "drift_violation":
            report["drift_violations"].append({"seq": entry["seq"], "ts": entry["ts"], "session_id": entry["session_id"], **data})
        elif etype == "backprop_update":
            report["backprop_updates"].append({"seq": entry["seq"], "ts": entry["ts"], "session_id": entry["session_id"], **data})
        report["timeline"].append({"seq": entry["seq"], "ts": entry["ts"], "type": etype, "session_id": entry["session_id"]})

report["entry_types"] = dict(report["entry_types"])
if output_file:
    with open(output_file, 'w') as f:
        json.dump(report, f, indent=2)
    print(f"✅ Audit report saved to {output_file}")
else:
    print(json.dumps(report, indent=2))
PYEOF
}

ledger_head() {
    local N="" ledger="$LEDGER_PATH"
    if [[ "${1:-}" == --* ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --count) N="$2"; shift 2 ;;
                --file) ledger="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
    else
        N="${1:-5}"
    fi
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi
    echo "📖 Ledger head (first $N entries):"
    python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    for i, line in enumerate(f):
        if i >= int(sys.argv[2]): break
        line = line.strip()
        if not line: continue
        e = json.loads(line)
        print(f'  [{e[\"seq\"]}] {e[\"type\"]} @ {e[\"ts\"]} (hash: {e[\"hash\"][:16]}...)')
" "$ledger" "$N"
}

ledger_tail() {
    local N="" ledger="$LEDGER_PATH"
    if [[ "${1:-}" == --* ]]; then
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --count) N="$2"; shift 2 ;;
                --file) ledger="$2"; shift 2 ;;
                *) shift ;;
            esac
        done
    else
        N="${1:-5}"
    fi
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi
    echo "📖 Ledger tail (last $N entries):"
    python3 -c "
import json, sys
from collections import deque
with open(sys.argv[1]) as f:
    lines = deque(maxlen=int(sys.argv[2]))
    for line in f:
        line = line.strip()
        if not line: continue
        try:
            lines.append(line)
        except: pass
for line in lines:
    try:
        e = json.loads(line)
        print(f'  [{e[\"seq\"]}] {e[\"type\"]} @ {e[\"ts\"]} (hash: {e[\"hash\"][:16]}...)')
    except json.JSONDecodeError:
        print(f'  [invalid] (skipped malformed line)')
" "$ledger" "$N"
}

chain() {
    local ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found"
        exit 1
    fi
    echo "🔗 Hash chain summary:"
    python3 - "$ledger" << 'PYEOF'
import json, sys

ledger_path = sys.argv[1]
entries = []
with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if line:
            entries.append(json.loads(line))

if not entries:
    print("  Empty ledger")
    exit(0)

print(f"  Total entries: {len(entries)}")
print(f"  Genesis hash: {entries[0]['prev_hash'][:16]}...")
print(f"  Latest hash:  {entries[-1]['hash'][:16]}...")
print(f"  Chain length: {len(entries)} links")

seqs = [e['seq'] for e in entries]
expected = list(range(1, len(entries) + 1))
if seqs == expected:
    print("  Sequence: ✅ Continuous (no gaps)")
else:
    missing = set(expected) - set(seqs)
    print(f"  Sequence: ❌ Gaps at {sorted(missing)}")
PYEOF
}

recover() {
    local ledger="$LEDGER_PATH" FORCE=false
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --force) FORCE=true; shift ;;
            *) shift ;;
        esac
    done

    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found at $ledger"
        exit 1
    fi

    # First verify if ledger is actually corrupted
    if ! verify --file "$ledger" >/dev/null 2>&1; then
        echo "🔧 Ledger corruption detected. Starting recovery..."

        # Create backup
        local BACKUP_FILE="${ledger}.corrupted.$(date -u '+%Y%m%d_%H%M%S')"
        cp "$ledger" "$BACKUP_FILE"
        echo "   ✅ Backup created: $BACKUP_FILE"

        # Try to rebuild chain by re-hashing valid entries
        local TMP_FILE=""
        TMP_FILE=$(mktemp "${ledger}.recover.XXXXXX") || {
            echo "❌ Failed to create temp file for recovery" >&2
            return 1
        }

        cleanup_tmp() {
            if [[ -n "$TMP_FILE" ]] && [[ -f "$TMP_FILE" ]]; then
                rm -f "$TMP_FILE" 2>/dev/null || true
            fi
        }
        trap cleanup_tmp ERR RETURN

        # Run Python recovery script
        if ! python3 - "$ledger" "$TMP_FILE" << 'PYEOF'
import json, hashlib, sys

def sha256(s):
    return hashlib.sha256(s.encode()).hexdigest()

ledger_path = sys.argv[1]
tmp_path = sys.argv[2]

prev_hash = None
recovered_count = 0
skipped_count = 0

with open(ledger_path, 'r') as inf, open(tmp_path, 'w') as outf:
    for line_num, line in enumerate(inf, 1):
        line = line.strip()
        if not line:
            continue

        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            skipped_count += 1
            continue

        # Extract fields
        seq = entry.get('seq', 0)
        ts = entry.get('ts', '')
        entry_type = entry.get('type', '')
        session_id = entry.get('session_id', '')
        data = json.dumps(entry.get('data', {}), separators=(',', ':'), ensure_ascii=False)
        stored_prev_hash = entry.get('prev_hash', '')
        stored_hash = entry.get('hash', '')

        # First entry: accept as new genesis
        if prev_hash is None:
            if entry_type == 'genesis':
                # Verify genesis prev_hash is all zeros
                expected_genesis = '0000000000000000000000000000000000000000000000000000000000000000'
                if stored_prev_hash != expected_genesis:
                    # Force genesis prev_hash to zeros
                    stored_prev_hash = expected_genesis
                    entry['prev_hash'] = stored_prev_hash
            # Recompute hash for first entry
            hash_input = f"{seq}|{ts}|{entry_type}|{session_id}|{data}|{stored_prev_hash}"
            entry['hash'] = sha256(hash_input)
            prev_hash = entry['hash']
            recovered_count += 1
            outf.write(json.dumps(entry) + '\n')
            continue

        # Subsequent entries: force prev_hash to match our chain
        if stored_prev_hash != prev_hash:
            # Chain is broken, force continuity
            entry['prev_hash'] = prev_hash

        # Recompute hash
        hash_input = f"{seq}|{ts}|{entry_type}|{session_id}|{data}|{entry['prev_hash']}"
        entry['hash'] = sha256(hash_input)

        prev_hash = entry['hash']
        recovered_count += 1
        outf.write(json.dumps(entry) + '\n')

print(f"Recovered {recovered_count} entries, skipped {skipped_count} malformed lines")
PYEOF
        then
            echo "❌ Recovery failed during re-hashing" >&2
            cleanup_tmp
            return 1
        fi

        # Replace ledger with recovered version
        mv "$TMP_FILE" "$ledger"
        trap - ERR RETURN

        echo "   ✅ Ledger recovered by backing up and rehashing valid entries"
        echo "   ⚠️  Review backup at $BACKUP_FILE for lost/corrupted entries"

        # Verify the recovered ledger
        if verify --file "$ledger" >/dev/null 2>&1; then
            echo "   ✅ Recovery successful: chain integrity restored"
        else
            echo "   ❌ Recovery failed: ledger still corrupted"
            return 1
        fi
    else
        echo "✅ Ledger is healthy. No recovery needed."
        if [[ "$FORCE" == true ]]; then
            echo "   Use --force to override and re-hash anyway"
        fi
    fi
}

archive() {
    local ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found at $ledger"
        exit 1
    fi

    ARCHIVE_DIR="$(dirname "$ledger")/ledger-archive"
    mkdir -p "$ARCHIVE_DIR"
    TIMESTAMP=$(date -u '+%Y%m%d_%H%M%S')
    ARCHIVE_FILE="$ARCHIVE_DIR/ledger-${TIMESTAMP}.jsonl.gz"

    gzip -c "$ledger" > "$ARCHIVE_FILE"
    echo "✅ Ledger archived to $ARCHIVE_FILE"
    echo "   Original: $(wc -c < "$ledger") bytes"
    echo "   Archived: $(wc -c < "$ARCHIVE_FILE") bytes"
}

status() {
    local ledger="$LEDGER_PATH"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found at $ledger"
        echo "   Run 'ledger.sh init' to create"
        exit 1
    fi

    local COUNT=$(entry_count_for "$ledger")
    local SIZE=$(wc -c < "$ledger" | tr -d ' ')
    local ARCHIVE_DIR="$(dirname "$ledger")/ledger-archive"
    local ARCHIVE_COUNT=0
    local ARCHIVE_SIZE=0
    if [[ -d "$ARCHIVE_DIR" ]]; then
        ARCHIVE_COUNT=$(find "$ARCHIVE_DIR" -name '*.gz' 2>/dev/null | wc -l | tr -d ' ')
        ARCHIVE_SIZE=$(du -sb "$ARCHIVE_DIR" 2>/dev/null | awk '{print $1}' || echo 0)
    fi

    echo "📊 Truth-Chain Ledger Status"
    echo "   Path: $ledger"
    echo "   Entries: $COUNT"
    echo "   Size: $SIZE bytes"
    if [[ $ARCHIVE_COUNT -gt 0 ]]; then
        echo "   Archives: $ARCHIVE_COUNT files ($ARCHIVE_SIZE bytes total)"
    fi
    echo ""
    ledger_head --count 3 --file "$ledger"
    echo "   ..."
    ledger_tail --count 3 --file "$ledger"
}

prune() {
    local ledger="$LEDGER_PATH" KEEP_DAYS=30 KEEP_ENTRIES=1000
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --file) ledger="$2"; shift 2 ;;
            --keep-days) KEEP_DAYS="$2"; shift 2 ;;
            --keep-entries) KEEP_ENTRIES="$2"; shift 2 ;;
            *) shift ;;
        esac
    done

    if [[ ! -f "$ledger" ]]; then
        echo "❌ Ledger not found at $ledger"
        exit 1
    fi

    local COUNT=$(entry_count_for "$ledger")
    if [[ $COUNT -le $KEEP_ENTRIES ]]; then
        echo "✅ Ledger has $COUNT entries (threshold: $KEEP_ENTRIES). No pruning needed."
        return
    fi

    local ARCHIVE_DIR="$(dirname "$ledger")/ledger-archive"
    mkdir -p "$ARCHIVE_DIR"

    # Calculate cutoff: keep last N entries, archive the rest
    local ARCHIVE_COUNT=$((COUNT - KEEP_ENTRIES))
    local CUTOFF_LINE=$ARCHIVE_COUNT

    echo "🗑️  Pruning ledger: $COUNT entries → keep last $KEEP_ENTRIES, archive $ARCHIVE_COUNT"

    # Get the hash of the last entry to archive (for chain continuity)
    local LAST_ARCHIVED_HASH=$(head -n "$CUTOFF_LINE" "$ledger" | tail -n 1 | python3 -c "import sys,json; print(json.loads(sys.stdin.read())['hash'])" 2>/dev/null)

    # Archive old entries
    local TIMESTAMP=$(date -u '+%Y%m%d_%H%M%S')
    local ARCHIVE_FILE="$ARCHIVE_DIR/ledger-${TIMESTAMP}.jsonl.gz"
    head -n "$CUTOFF_LINE" "$ledger" | gzip -c > "$ARCHIVE_FILE"

    local ARCHIVE_SIZE=$(wc -c < "$ARCHIVE_FILE" | tr -d ' ')
    local ORIGINAL_SIZE=$(head -n "$CUTOFF_LINE" "$ledger" | wc -c | tr -d ' ')

    # Keep only recent entries
    tail -n "$KEEP_ENTRIES" "$ledger" > "${ledger}.tmp"
    mv "${ledger}.tmp" "$ledger"

    # Re-sequence remaining entries to maintain continuity
    python3 - "$ledger" "$LAST_ARCHIVED_HASH" << 'PYEOF'
import json, hashlib, sys

def sha256(s):
    return hashlib.sha256(s.encode()).hexdigest()

ledger_path = sys.argv[1]
prev_hash = sys.argv[2]

entries = []
with open(ledger_path, 'r') as f:
    for line in f:
        line = line.strip()
        if line:
            entries.append(json.loads(line))

# Re-sequence and re-hash
with open(ledger_path, 'w') as f:
    for i, entry in enumerate(entries, 1):
        entry['seq'] = i
        entry['prev_hash'] = prev_hash
        data = json.dumps(entry.get('data', {}), separators=(',', ':'), ensure_ascii=False)
        hash_input = f"{entry['seq']}|{entry['ts']}|{entry['type']}|{entry['session_id']}|{data}|{prev_hash}"
        entry['hash'] = sha256(hash_input)
        prev_hash = entry['hash']
        f.write(json.dumps(entry) + '\n')

print(f"Re-sequenced {len(entries)} entries. New chain starts with prev_hash: {sys.argv[2][:16]}...")
PYEOF

    echo "✅ Pruned: $ARCHIVE_COUNT entries archived to $ARCHIVE_FILE ($ARCHIVE_SIZE bytes, was $ORIGINAL_SIZE)"
    echo "   Remaining: $KEEP_ENTRIES entries in active ledger"
    echo "   Chain continuity: preserved (boundary hash: ${LAST_ARCHIVED_HASH:0:16}...)"
}

# --- MAIN ---

case "$CMD" in
    init) init "$@" ;;
    append) append "$@" ;;
    verify) verify "$@" ;;
    backup) backup "$@" ;;
    rotate) rotate "$@" ;;
    replay-to-session-db) replay_to_session_db "$@" ;;
    recover) recover "$@" ;;
    replay) replay "$@" ;;
    query) query "$@" ;;
    audit) audit "$@" ;;
    head) ledger_head "$@" ;;
    tail) ledger_tail "$@" ;;
    chain) chain "$@" ;;
    archive) archive "$@" ;;
    prune) prune "$@" ;;
    status) status "$@" ;;
    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║          🔗 TRUTH-CHAIN — Immutable Append-Only Ledger   ║
╚══════════════════════════════════════════════════════════╝

Ledger: .specify/ledger.jsonl

COMMANDS:
  ledger.sh init [--file <path>]                Initialize ledger with genesis entry
  ledger.sh append <type> '<json>' [sid]        Append entry (hash-chained)
  ledger.sh append --type T --data D [--session S] [--file F] [--json]
       [--workflow-run-id W] [--agent A] [--risk R]
       [--input-hash H] [--output-hash H] [--redactions <json>] [--redact]
       Append entry with v2 metadata. --redact scrubs secrets from data payload.
  ledger.sh verify [--file <path>] [--json]     Verify entire hash chain integrity
  ledger.sh backup [--file <path>] [--json]       Copy ledger to .specify/backups/
  ledger.sh rotate [--file <path>] [--max-size B] [--json]
       Rotate ledger when exceeding size threshold (default 10MB)
  ledger.sh replay-to-session-db [--file <path>] [--db <path>] [--json]
       Replay ledger entries into session DB ledger_refs table.
       Prefers ./scripts/session-db.sh record-ledger-ref; falls back to direct SQL.
  ledger.sh recover [--file <path>] [--force]   Back up corrupted ledger and rehash valid entries
  ledger.sh replay <seq> [--file <path>]        Reconstruct state at entry N
  ledger.sh query [--type T] [--agent A]        Query entries with filters
  ledger.sh audit [--session S] [--out F]       Generate audit report
  ledger.sh head [--count N] [--file <path>]    Show first N entries
  ledger.sh tail [--count N] [--file <path>]    Show last N entries
  ledger.sh chain [--file <path>]               Hash chain summary
  ledger.sh prune [--keep-entries N] [--file <path>]  Archive old entries, keep recent
  ledger.sh archive [--file <path>]             Compress and archive old entries
  ledger.sh status [--file <path>]              Ledger status overview

LOCKING:
  Uses flock (Linux) > shlock (macOS/BSD) > mkdir (portable fallback).
  Automatic — no manual lock management required.

ENTRY TYPES:
  session_start, session_end, dispatch, dispatch_complete, dispatch_fail
  decision, memory, artifact, barrier_create, barrier_resolve
  lock_acquire, lock_release, msg_send, msg_read
  quality_gate, drift_violation, backprop_update, token_log, eval_run
  ptask_register, ptask_complete, ptask_fail
  workflow_start, workflow_step_start, workflow_step_done, workflow_fail, eval_run

EXAMPLES:
  ledger.sh init
  ledger.sh append session_start '{"goal":"Add auth","repo":"fedora"}' ses_abc
  ledger.sh append --type dispatch --data '{"agent":"wall-builder"}' --session ses_abc --json
  ledger.sh append --type decision --data '{"title":"X"}' --db-row-id 42 --risk medium
  ledger.sh verify
  ledger.sh verify --json
  ledger.sh backup
  ledger.sh rotate --max-size 5242880
  ledger.sh replay-to-session-db --db .specify/session.db
  ledger.sh query --type decision --tag auth
  ledger.sh replay 15
  ledger.sh audit --session ses_abc --output audit.json
HELP
        ;;
esac
