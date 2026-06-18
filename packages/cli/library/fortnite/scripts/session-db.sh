#!/usr/bin/env bash
# session-db.sh — SQLite-based session memory for multi-agent projects
# Database: .specify/session.db
# Usage: ./scripts/session-db.sh [global-flags] <command> [args...]
# Global flags: --db <path>, --json, --dry-run

set -euo pipefail

# --- GLOBAL FLAG PARSING ---
DB_PATH="${OPENCODE_WORKSPACE:-.}/.specify/session.db"
JSON_OUT=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --db)
            DB_PATH="$2"
            shift 2
            ;;
        --json)
            JSON_OUT=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --)
            shift
            break
            ;;
        -*)
            echo "Unknown flag: $1" >&2
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

LEDGER_SCRIPT="${OPENCODE_WORKSPACE:-.}/skills/truth-chain/scripts/ledger.sh"
CMD="${1:-help}"
shift 2>/dev/null || true

# --- LEDGER MIRROR (best-effort, non-blocking) ---

ledger_append() {
    local entry_type="$1"
    local data_json="$2"
    local session_id="${3:-unknown}"
    local db_row_id="${4:-}"
    if [[ -x "$LEDGER_SCRIPT" ]]; then
        if [[ -n "$db_row_id" ]]; then
            "$LEDGER_SCRIPT" append --type "$entry_type" --data "$data_json" --session "$session_id" --db-row-id "$db_row_id" >/dev/null 2>&1 || true
        else
            "$LEDGER_SCRIPT" append "$entry_type" "$data_json" "$session_id" >/dev/null 2>&1 || true
        fi
    fi
}

# --- SQL HELPERS (v2: PRAGMAs on every connection) ---

run_sql() {
    local sql="$1"
    sqlite3 -cmd ".output /dev/null" -cmd "PRAGMA busy_timeout=5000;" -cmd ".output stdout" "$DB_PATH" "PRAGMA foreign_keys=ON; $sql"
}

run_sql_table() {
    local sql="$1"
    sqlite3 -cmd ".output /dev/null" -cmd "PRAGMA busy_timeout=5000;" -cmd ".output stdout" -header -column "$DB_PATH" "PRAGMA foreign_keys=ON; $sql"
}

run_sql_json() {
    local sql="$1"
    sqlite3 -cmd ".output /dev/null" -cmd "PRAGMA busy_timeout=5000;" -cmd ".output stdout" -json "$DB_PATH" "PRAGMA foreign_keys=ON; $sql"
}

init_db() {
    sqlite3 -cmd ".output /dev/null" -cmd "PRAGMA busy_timeout=5000;" -cmd ".output stdout" "$DB_PATH" "
        PRAGMA foreign_keys=ON;

        CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            started_at TEXT NOT NULL,
            ended_at TEXT,
            agent TEXT NOT NULL DEFAULT 'loop-driver',
            model TEXT NOT NULL,
            goal TEXT,
            repo TEXT,
            worktree TEXT,
            status TEXT DEFAULT 'active',
            token_total INTEGER DEFAULT 0,
            summary TEXT,
            tags TEXT
        );

        CREATE TABLE IF NOT EXISTS dispatches (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT NOT NULL REFERENCES sessions(id),
            seq INTEGER NOT NULL DEFAULT 0,
            parent_id INTEGER REFERENCES dispatches(id),
            agent TEXT NOT NULL,
            model TEXT NOT NULL,
            task TEXT NOT NULL,
            phase TEXT,
            workflow TEXT,
            mode TEXT,
            started_at TEXT NOT NULL,
            ended_at TEXT,
            result TEXT DEFAULT 'pending',
            token_used INTEGER DEFAULT 0,
            error_message TEXT,
            summary TEXT,
            files_touched TEXT
        );

        CREATE TABLE IF NOT EXISTS decisions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            dispatch_id INTEGER REFERENCES dispatches(id),
            title TEXT NOT NULL,
            context TEXT,
            decision TEXT NOT NULL,
            rationale TEXT,
            alternatives TEXT,
            created_at TEXT NOT NULL,
            tags TEXT
        );

        CREATE TABLE IF NOT EXISTS artifacts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            dispatch_id INTEGER REFERENCES dispatches(id),
            path TEXT NOT NULL,
            action TEXT NOT NULL,
            created_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS memories (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            tags TEXT,
            importance TEXT DEFAULT 'normal',
            verified INTEGER DEFAULT 0,
            created_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS token_log (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            dispatch_id INTEGER REFERENCES dispatches(id),
            token_count INTEGER NOT NULL,
            context_pct REAL,
            recorded_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS parallel_tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            parent_dispatch_id INTEGER REFERENCES dispatches(id),
            wave_id TEXT NOT NULL DEFAULT 'wave_1',
            agent TEXT NOT NULL,
            task TEXT NOT NULL,
            status TEXT DEFAULT 'pending',
            started_at TEXT,
            completed_at TEXT,
            result TEXT,
            output_path TEXT,
            error_message TEXT,
            created_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            from_agent TEXT NOT NULL,
            to_agent TEXT NOT NULL,
            subject TEXT NOT NULL,
            body TEXT NOT NULL,
            priority TEXT DEFAULT 'normal',
            status TEXT DEFAULT 'unread',
            created_at TEXT NOT NULL,
            read_at TEXT
        );

        CREATE TABLE IF NOT EXISTS barriers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            barrier_id TEXT NOT NULL,
            expected_count INTEGER NOT NULL,
            arrived_count INTEGER DEFAULT 0,
            status TEXT DEFAULT 'waiting',
            created_at TEXT NOT NULL,
            resolved_at TEXT
        );

        CREATE TABLE IF NOT EXISTS locks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            lock_name TEXT NOT NULL,
            held_by TEXT NOT NULL,
            acquired_at TEXT NOT NULL,
            released_at TEXT,
            status TEXT DEFAULT 'active'
        );

        CREATE INDEX IF NOT EXISTS idx_token_log_session ON token_log(session_id);
        CREATE INDEX IF NOT EXISTS idx_parallel_tasks_session ON parallel_tasks(session_id);
        CREATE INDEX IF NOT EXISTS idx_parallel_tasks_wave ON parallel_tasks(wave_id);
        CREATE INDEX IF NOT EXISTS idx_parallel_tasks_status ON parallel_tasks(status);
        CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
        CREATE INDEX IF NOT EXISTS idx_messages_to_agent ON messages(to_agent, status);
        CREATE INDEX IF NOT EXISTS idx_messages_from_agent ON messages(from_agent);
        CREATE INDEX IF NOT EXISTS idx_barriers_session ON barriers(session_id);
        CREATE INDEX IF NOT EXISTS idx_barriers_status ON barriers(status);
        CREATE INDEX IF NOT EXISTS idx_locks_session ON locks(session_id);
        CREATE INDEX IF NOT EXISTS idx_locks_name ON locks(lock_name, status);

        CREATE TABLE IF NOT EXISTS teams (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            description TEXT,
            agents TEXT NOT NULL,
            created_at TEXT NOT NULL,
            updated_at TEXT
        );

        CREATE TABLE IF NOT EXISTS workflows (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL UNIQUE,
            description TEXT,
            trigger_cmd TEXT,
            config_json TEXT,
            version INTEGER DEFAULT 1,
            created_at TEXT NOT NULL,
            updated_at TEXT,
            team TEXT,
            steps TEXT
        );

        CREATE TABLE IF NOT EXISTS workflow_instances (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            workflow_name TEXT NOT NULL,
            session_id TEXT REFERENCES sessions(id),
            status TEXT DEFAULT 'pending',
            current_step INTEGER DEFAULT 0,
            total_steps INTEGER NOT NULL,
            result TEXT,
            on_failure TEXT DEFAULT 'stop',
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

        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            applied_at TEXT NOT NULL DEFAULT (datetime('now')),
            checksum TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS workflow_runs (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            workflow_name TEXT NOT NULL,
            mode TEXT NOT NULL,
            status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed', 'cancelled')),
            started_at TEXT NOT NULL DEFAULT (datetime('now')),
            ended_at TEXT,
            latency_ms INTEGER,
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS workflow_phases (
            id TEXT PRIMARY KEY,
            workflow_run_id TEXT NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
            phase_name TEXT NOT NULL,
            agent TEXT,
            status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped')),
            started_at TEXT,
            ended_at TEXT,
            latency_ms INTEGER,
            input_hash TEXT,
            output_hash TEXT,
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS agent_dispatches (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            from_agent TEXT NOT NULL,
            to_agent TEXT NOT NULL,
            task_type TEXT NOT NULL,
            mode TEXT NOT NULL,
            risk TEXT NOT NULL,
            reason TEXT NOT NULL,
            dispatch_depth INTEGER NOT NULL DEFAULT 0,
            fallback_used INTEGER NOT NULL DEFAULT 0,
            expected_output_schema TEXT,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS model_calls (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            agent TEXT,
            provider TEXT NOT NULL,
            model TEXT NOT NULL,
            purpose TEXT NOT NULL,
            tokens_in INTEGER,
            tokens_out INTEGER,
            cached_tokens INTEGER,
            latency_ms INTEGER,
            cost_usd REAL,
            success INTEGER NOT NULL,
            error TEXT,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS tool_calls (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            agent TEXT,
            tool_name TEXT NOT NULL,
            command TEXT,
            status TEXT NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'blocked')),
            exit_code INTEGER,
            latency_ms INTEGER,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS gate_results (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            gate_name TEXT NOT NULL,
            gate_type TEXT NOT NULL,
            status TEXT NOT NULL CHECK (status IN ('pass', 'fail', 'warning', 'skipped')),
            human_required INTEGER NOT NULL DEFAULT 0,
            human_decision TEXT CHECK (human_decision IN ('approved', 'rejected', 'deferred') OR human_decision IS NULL),
            reason TEXT,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS ledger_refs (
            seq INTEGER PRIMARY KEY,
            session_id TEXT REFERENCES sessions(id) ON DELETE SET NULL,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            event_type TEXT NOT NULL,
            event_hash TEXT NOT NULL UNIQUE,
            prev_hash TEXT,
            created_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS eval_runs (
            id TEXT PRIMARY KEY,
            suite_name TEXT NOT NULL,
            dataset_name TEXT NOT NULL,
            model TEXT,
            provider TEXT,
            status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
            started_at TEXT NOT NULL DEFAULT (datetime('now')),
            ended_at TEXT,
            score REAL,
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS eval_results (
            id TEXT PRIMARY KEY,
            eval_run_id TEXT NOT NULL REFERENCES eval_runs(id) ON DELETE CASCADE,
            case_id TEXT NOT NULL,
            status TEXT NOT NULL CHECK (status IN ('pass', 'fail', 'warning')),
            score REAL,
            expected_hash TEXT,
            actual_hash TEXT,
            judge_model TEXT,
            reason TEXT,
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS dataset_examples (
            id TEXT PRIMARY KEY,
            dataset_name TEXT NOT NULL,
            source_type TEXT NOT NULL,
            source_ref TEXT NOT NULL,
            input_hash TEXT NOT NULL,
            output_hash TEXT NOT NULL,
            label TEXT,
            accepted INTEGER,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS provider_health (
            id TEXT PRIMARY KEY,
            provider TEXT NOT NULL,
            model TEXT,
            status TEXT NOT NULL CHECK (status IN ('ok', 'degraded', 'down')),
            latency_ms INTEGER,
            error TEXT,
            checked_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS cost_snapshots (
            id TEXT PRIMARY KEY,
            session_id TEXT REFERENCES sessions(id) ON DELETE SET NULL,
            workflow_run_id TEXT REFERENCES workflow_runs(id) ON DELETE SET NULL,
            provider TEXT NOT NULL,
            model TEXT NOT NULL,
            tokens_in INTEGER NOT NULL DEFAULT 0,
            tokens_out INTEGER NOT NULL DEFAULT 0,
            cached_tokens INTEGER NOT NULL DEFAULT 0,
            estimated_cost_usd REAL NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL DEFAULT (datetime('now'))
        );

        CREATE TABLE IF NOT EXISTS checkpoints (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
            checkpoint_type TEXT NOT NULL CHECK (checkpoint_type IN ('start', 'transition', 'end')),
            summary TEXT NOT NULL,
            state_hash TEXT,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS lessons (
            id TEXT PRIMARY KEY,
            session_id TEXT REFERENCES sessions(id) ON DELETE SET NULL,
            source TEXT NOT NULL,
            lesson_text TEXT NOT NULL,
            confidence REAL,
            accepted INTEGER,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            metadata_json TEXT NOT NULL DEFAULT '{}'
        );

        CREATE TABLE IF NOT EXISTS quality_metrics (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT REFERENCES sessions(id),
            repo TEXT NOT NULL,
            gate_type TEXT NOT NULL CHECK(gate_type IN ('lint','test','typecheck','build')),
            passed INTEGER NOT NULL CHECK(passed IN (0,1)),
            duration_ms INTEGER,
            error_count INTEGER DEFAULT 0,
            warning_count INTEGER DEFAULT 0,
            timestamp TEXT NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_quality_metrics_session ON quality_metrics(session_id);
        CREATE INDEX IF NOT EXISTS idx_quality_metrics_repo ON quality_metrics(repo, gate_type);

        CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);
        CREATE INDEX IF NOT EXISTS idx_workflows_name ON workflows(name);
        CREATE INDEX IF NOT EXISTS idx_workflow_instances_session ON workflow_instances(session_id);
        CREATE INDEX IF NOT EXISTS idx_workflow_instances_status ON workflow_instances(status);
        CREATE INDEX IF NOT EXISTS idx_workflow_steps_instance ON workflow_steps(instance_id);
        CREATE INDEX IF NOT EXISTS idx_workflow_runs_session ON workflow_runs(session_id);
        CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs(status);
        CREATE INDEX IF NOT EXISTS idx_workflow_phases_run ON workflow_phases(workflow_run_id);
        CREATE INDEX IF NOT EXISTS idx_agent_dispatches_session ON agent_dispatches(session_id);
        CREATE INDEX IF NOT EXISTS idx_agent_dispatches_to ON agent_dispatches(to_agent);
        CREATE INDEX IF NOT EXISTS idx_model_calls_session ON model_calls(session_id);
        CREATE INDEX IF NOT EXISTS idx_model_calls_provider_model ON model_calls(provider, model);
        CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id);
        CREATE INDEX IF NOT EXISTS idx_gate_results_session ON gate_results(session_id);
        CREATE INDEX IF NOT EXISTS idx_ledger_refs_hash ON ledger_refs(event_hash);
        CREATE INDEX IF NOT EXISTS idx_eval_runs_suite ON eval_runs(suite_name);
        CREATE INDEX IF NOT EXISTS idx_eval_results_run ON eval_results(eval_run_id);
        CREATE INDEX IF NOT EXISTS idx_dataset_examples_dataset ON dataset_examples(dataset_name);
        CREATE INDEX IF NOT EXISTS idx_provider_health_provider ON provider_health(provider);
        CREATE INDEX IF NOT EXISTS idx_cost_snapshots_session ON cost_snapshots(session_id);
        CREATE INDEX IF NOT EXISTS idx_checkpoints_session ON checkpoints(session_id);
        CREATE INDEX IF NOT EXISTS idx_lessons_session ON lessons(session_id);
    "
    echo "✅ Database initialized: $DB_PATH" >&2
}

now() { date -u '+%Y-%m-%dT%H:%M:%SZ'; }

# --- SQL VALUE HELPER ---
# Returns 'value' with single quotes escaped, or NULL if empty
sql_val() {
    local val="$1"
    if [[ -n "$val" ]]; then
        printf "'%s'" "${val//\'/\'\'}"
    else
        printf "NULL"
    fi
}

case "$CMD" in
    init)
        mkdir -p "$(dirname "$DB_PATH")"
        init_db
        ;;

    # --- V2 LIFECYCLE COMMANDS ---

    migrate)
        mkdir -p "$(dirname "$DB_PATH")"
        init_db 2>/dev/null || true
        MIGRATIONS_DIR="$(dirname "$DB_PATH")/migrations"
        if [[ -d "$MIGRATIONS_DIR" ]]; then
            for mig in $(ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort); do
                MIG_NAME=$(basename "$mig")
                MIG_VERSION=$(echo "$MIG_NAME" | sed -E 's/^([0-9]+)_.*/\1/')
                ALREADY=$(run_sql "SELECT 1 FROM schema_migrations WHERE version=$MIG_VERSION;" 2>/dev/null || echo "")
                if [[ -z "$ALREADY" ]]; then
                    if [[ "$DRY_RUN" == true ]]; then
                        echo "[dry-run] Would apply $MIG_NAME"
                    else
                        sqlite3 -cmd ".output /dev/null" -cmd "PRAGMA busy_timeout=5000;" -cmd ".output stdout" "$DB_PATH" "PRAGMA foreign_keys=ON; $(cat "$mig")"
                        CHECKSUM=$(md5 -q "$mig" 2>/dev/null || md5sum "$mig" | awk '{print $1}')
                        run_sql "INSERT INTO schema_migrations (version, name, applied_at, checksum) VALUES ($MIG_VERSION, '${MIG_NAME//\'/\'\'}', '$(now)', '$CHECKSUM');"
                        echo "✅ Applied migration $MIG_NAME"
                    fi
                else
                    echo "⏭️  Migration $MIG_NAME already applied"
                fi
            done
        fi
        echo "✅ Migrations complete"
        ;;

    backup)
        BACKUP_DIR="$(dirname "$DB_PATH")/backups"
        mkdir -p "$BACKUP_DIR"
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        BACKUP_FILE="$BACKUP_DIR/session-${TIMESTAMP}.db"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would backup to $BACKUP_FILE"
        else
            cp "$DB_PATH" "$BACKUP_FILE"
            echo "✅ Backup created: $BACKUP_FILE"
        fi
        ;;

    verify)
        MISSING_TABLES=()
        for tbl in schema_migrations sessions workflow_runs workflow_phases agent_dispatches model_calls tool_calls gate_results ledger_refs eval_runs eval_results dataset_examples provider_health cost_snapshots checkpoints lessons; do
            EXISTS=$(run_sql "SELECT 1 FROM sqlite_master WHERE type='table' AND name='$tbl';" 2>/dev/null || echo "")
            if [[ -z "$EXISTS" ]]; then
                MISSING_TABLES+=("$tbl")
            fi
        done

        WF_RUNS_COLS=$(run_sql "SELECT name FROM pragma_table_info('workflow_runs');" 2>/dev/null || echo "")
        MISSING_COLS=()
        for col in workflow_name status latency_ms metadata_json; do
            if [[ "$WF_RUNS_COLS" != *"$col"* ]]; then
                MISSING_COLS+=("workflow_runs.$col")
            fi
        done

        WF_PHASES_COLS=$(run_sql "SELECT name FROM pragma_table_info('workflow_phases');" 2>/dev/null || echo "")
        for col in workflow_run_id phase_name status latency_ms metadata_json; do
            if [[ "$WF_PHASES_COLS" != *"$col"* ]]; then
                MISSING_COLS+=("workflow_phases.$col")
            fi
        done

        OK=$([[ ${#MISSING_TABLES[@]} -eq 0 && ${#MISSING_COLS[@]} -eq 0 ]] && echo "true" || echo "false")

        if [[ "$JSON_OUT" == true ]]; then
            # Build JSON arrays safely
            TABLES_JSON="["
            for i in "${!MISSING_TABLES[@]}"; do
                [[ $i -gt 0 ]] && TABLES_JSON+=","
                TABLES_JSON+="\"${MISSING_TABLES[$i]}\""
            done
            TABLES_JSON+="]"

            COLS_JSON="["
            for i in "${!MISSING_COLS[@]}"; do
                [[ $i -gt 0 ]] && COLS_JSON+=","
                COLS_JSON+="\"${MISSING_COLS[$i]}\""
            done
            COLS_JSON+="]"

            echo "{\"ok\":$OK,\"missing_tables\":$TABLES_JSON,\"missing_columns\":$COLS_JSON}"
        else
            if [[ "$OK" == "true" ]]; then
                echo "✅ Schema verification passed"
            else
                echo "❌ Schema verification failed"
                if [[ ${#MISSING_TABLES[@]} -gt 0 ]]; then
                    echo "   Missing tables: ${MISSING_TABLES[*]}"
                fi
                if [[ ${#MISSING_COLS[@]} -gt 0 ]]; then
                    echo "   Missing columns: ${MISSING_COLS[*]}"
                fi
                exit 1
            fi
        fi
        ;;

    start-session)
        GOAL="${1:-unknown}"
        REPO="${2:-$(basename "$(pwd)")}"
        SID="ses_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        MODEL="ollama-cloud/minimax-m2.7"
        init_db 2>/dev/null || true
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would start session $SID"
        else
            ROW_ID=$(run_sql "
                INSERT INTO sessions(id, started_at, agent, model, goal, repo, status)
                VALUES ('$SID', '$(now)', 'loop-driver', '$MODEL', '${GOAL//\'/\'\'}', '$REPO', 'active');
                SELECT last_insert_rowid();
            " 2>/dev/null)
            ledger_append session_start "{\"goal\":\"${GOAL//\'/\'\'}\",\"repo\":\"$REPO\",\"model\":\"$MODEL\"}" "$SID" "$ROW_ID"
        fi
        echo "$SID"
        ;;

    end-session)
        SID="${1:-}"
        STATUS="${2:-completed}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh end-session <session-id> [status]"; exit 1; }
        TOKENS=$(run_sql "SELECT COALESCE(SUM(token_used),0) FROM dispatches WHERE session_id='$SID';")
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would end session $SID"
        else
            run_sql "
                UPDATE sessions SET ended_at='$(now)', status='$STATUS', token_total=$TOKENS WHERE id='$SID';
            "
            ledger_append session_end "{\"status\":\"$STATUS\",\"token_total\":$TOKENS}" "$SID"
        fi
        echo "✅ Session $SID ended ($STATUS, $TOKENS tokens)"
        ;;

    record-dispatch)
        SID="${1:-}"
        TO_AGENT="${2:-}"
        TASK_TYPE="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-dispatch <session-id> <to_agent> <task_type> [from_agent] [mode] [risk] [reason] [workflow_run_id]"; exit 1; }
        [ -z "$TO_AGENT" ] && { echo "Usage: session-db.sh record-dispatch <session-id> <to_agent> <task_type> [from_agent] [mode] [risk] [reason] [workflow_run_id]"; exit 1; }
        [ -z "$TASK_TYPE" ] && { echo "Usage: session-db.sh record-dispatch <session-id> <to_agent> <task_type> [from_agent] [mode] [risk] [reason] [workflow_run_id]"; exit 1; }
        FROM_AGENT="${4:-loop-driver}"
        MODE="${5:-standard}"
        RISK="${6:-low}"
        REASON="${7:-dispatch}"
        WORKFLOW_RUN_ID="${8:-}"
        DISPATCH_ID="disp_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record v2 dispatch $DISPATCH_ID for $SID"
        else
            ROW_ID=$(run_sql "
                INSERT INTO agent_dispatches(id, session_id, workflow_run_id, from_agent, to_agent, task_type, mode, risk, reason, created_at)
                VALUES ('$DISPATCH_ID', '$SID', $(sql_val "$WORKFLOW_RUN_ID"), '$FROM_AGENT', '$TO_AGENT', '${TASK_TYPE//\'/\'\'}', '$MODE', '$RISK', '${REASON//\'/\'\'}', '$(now)');
                SELECT last_insert_rowid();
            ")
            ledger_append dispatch "{\"to_agent\":\"$TO_AGENT\",\"task_type\":\"${TASK_TYPE//\'/\'\'}\",\"mode\":\"$MODE\",\"risk\":\"$RISK\",\"workflow_run_id\":\"${WORKFLOW_RUN_ID}\"}" "$SID" "$ROW_ID"
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"dispatch_id\":\"$DISPATCH_ID\",\"session_id\":\"$SID\",\"to_agent\":\"$TO_AGENT\",\"task_type\":\"${TASK_TYPE//\"/\\\"}\",\"mode\":\"$MODE\",\"risk\":\"$RISK\"}"
        else
            echo "📤 v2 $FROM_AGENT → $TO_AGENT: $TASK_TYPE (dispatch $DISPATCH_ID, session $SID)"
        fi
        ;;

    start-workflow)
        SID="${1:-}"
        WORKFLOW_NAME="${2:-}"
        MODE="${3:-standard}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh start-workflow <session-id> <workflow_name> [mode]"; exit 1; }
        [ -z "$WORKFLOW_NAME" ] && { echo "Usage: session-db.sh start-workflow <session-id> <workflow_name> [mode]"; exit 1; }
        WID="wf_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would start workflow $WID for $SID"
        else
            run_sql "
                INSERT INTO workflow_runs(id, session_id, workflow_name, mode, status, started_at)
                VALUES ('$WID', '$SID', '${WORKFLOW_NAME//\'/\'\'}', '$MODE', 'running', '$(now)');
            "
            ledger_append workflow_start "{\"workflow_run_id\":\"$WID\",\"workflow_name\":\"${WORKFLOW_NAME//\'/\'\'}\",\"mode\":\"$MODE\"}" "$SID"
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"workflow_run_id\":\"$WID\",\"session_id\":\"$SID\",\"workflow_name\":\"${WORKFLOW_NAME//\"/\\\"}\",\"mode\":\"$MODE\",\"status\":\"running\"}"
        else
            echo "🔄 Workflow $WID started: $WORKFLOW_NAME (session $SID)"
        fi
        ;;

    end-workflow)
        WID="${1:-}"
        STATUS="${2:-completed}"
        LATENCY_MS="${3:-}"
        [ -z "$WID" ] && { echo "Usage: session-db.sh end-workflow <workflow_run_id> [status] [latency_ms]"; exit 1; }
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would end workflow $WID"
        else
            run_sql "
                UPDATE workflow_runs SET ended_at='$(now)', status='$STATUS'${LATENCY_MS:+, latency_ms=$LATENCY_MS} WHERE id='$WID';
            "
            ledger_append workflow_end "{\"workflow_run_id\":\"$WID\",\"status\":\"$STATUS\"${LATENCY_MS:+,\"latency_ms\":$LATENCY_MS}}" ""
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"workflow_run_id\":\"$WID\",\"status\":\"$STATUS\"${LATENCY_MS:+,\"latency_ms\":$LATENCY_MS}}"
        else
            echo "✅ Workflow $WID ended ($STATUS)"
        fi
        ;;

    start-phase)
        WID="${1:-}"
        PHASE_NAME="${2:-}"
        AGENT="${3:-}"
        [ -z "$WID" ] && { echo "Usage: session-db.sh start-phase <workflow_run_id> <phase_name> [agent]"; exit 1; }
        [ -z "$PHASE_NAME" ] && { echo "Usage: session-db.sh start-phase <workflow_run_id> <phase_name> [agent]"; exit 1; }
        PID="ph_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would start phase $PID for workflow $WID"
        else
            run_sql "
                INSERT INTO workflow_phases(id, workflow_run_id, phase_name, agent, status, started_at)
                VALUES ('$PID', '$WID', '${PHASE_NAME//\'/\'\'}', '${AGENT}', 'running', '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"phase_id\":\"$PID\",\"workflow_run_id\":\"$WID\",\"phase_name\":\"${PHASE_NAME//\"/\\\"}\",\"agent\":\"${AGENT}\",\"status\":\"running\"}"
        else
            echo "▶️  Phase $PID started: $PHASE_NAME (workflow $WID)"
        fi
        ;;

    end-phase)
        PID="${1:-}"
        STATUS="${2:-completed}"
        LATENCY_MS="${3:-}"
        [ -z "$PID" ] && { echo "Usage: session-db.sh end-phase <phase_id> [status] [latency_ms]"; exit 1; }
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would end phase $PID"
        else
            run_sql "
                UPDATE workflow_phases SET ended_at='$(now)', status='$STATUS'${LATENCY_MS:+, latency_ms=$LATENCY_MS} WHERE id='$PID';
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"phase_id\":\"$PID\",\"status\":\"$STATUS\"${LATENCY_MS:+,\"latency_ms\":$LATENCY_MS}}"
        else
            echo "✅ Phase $PID ended ($STATUS)"
        fi
        ;;

    record-model-call)
        SID="${1:-}"
        PROVIDER="${2:-}"
        MODEL="${3:-}"
        PURPOSE="${4:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-model-call <session-id> <provider> <model> <purpose> [workflow_run_id] [agent] [tokens_in] [tokens_out] [latency_ms] [cost_usd] [success]"; exit 1; }
        [ -z "$PROVIDER" ] && { echo "Usage: session-db.sh record-model-call <session-id> <provider> <model> <purpose> [workflow_run_id] [agent] [tokens_in] [tokens_out] [latency_ms] [cost_usd] [success]"; exit 1; }
        [ -z "$MODEL" ] && { echo "Usage: session-db.sh record-model-call <session-id> <provider> <model> <purpose> [workflow_run_id] [agent] [tokens_in] [tokens_out] [latency_ms] [cost_usd] [success]"; exit 1; }
        [ -z "$PURPOSE" ] && { echo "Usage: session-db.sh record-model-call <session-id> <provider> <model> <purpose> [workflow_run_id] [agent] [tokens_in] [tokens_out] [latency_ms] [cost_usd] [success]"; exit 1; }
        WORKFLOW_RUN_ID="${5:-}"
        AGENT="${6:-}"
        TOKENS_IN="${7:-}"
        TOKENS_OUT="${8:-}"
        LATENCY_MS="${9:-}"
        COST_USD="${10:-}"
        SUCCESS="${11:-1}"
        MCID="mc_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record model call $MCID"
        else
            run_sql "
                INSERT INTO model_calls(id, session_id, workflow_run_id, agent, provider, model, purpose, tokens_in, tokens_out, latency_ms, cost_usd, success, created_at)
                VALUES ('$MCID', '$SID', $(sql_val "$WORKFLOW_RUN_ID"), $(sql_val "$AGENT"), '$PROVIDER', '$MODEL', '${PURPOSE//\'/\'\'}', ${TOKENS_IN:-NULL}, ${TOKENS_OUT:-NULL}, ${LATENCY_MS:-NULL}, ${COST_USD:-NULL}, $SUCCESS, '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"model_call_id\":\"$MCID\",\"session_id\":\"$SID\",\"provider\":\"$PROVIDER\",\"model\":\"$MODEL\",\"purpose\":\"${PURPOSE//\"/\\\"}\",\"success\":$SUCCESS}"
        else
            echo "🤖 Model call $MCID: $PROVIDER/$MODEL ($PURPOSE, success=$SUCCESS)"
        fi
        ;;

    record-tool-call)
        SID="${1:-}"
        TOOL_NAME="${2:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-tool-call <session-id> <tool_name> [command] [workflow_run_id] [agent] [status] [exit_code] [latency_ms]"; exit 1; }
        [ -z "$TOOL_NAME" ] && { echo "Usage: session-db.sh record-tool-call <session-id> <tool_name> [command] [workflow_run_id] [agent] [status] [exit_code] [latency_ms]"; exit 1; }
        COMMAND="${3:-}"
        WORKFLOW_RUN_ID="${4:-}"
        AGENT="${5:-}"
        STATUS="${6:-completed}"
        EXIT_CODE="${7:-}"
        LATENCY_MS="${8:-}"
        TCID="tc_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record tool call $TCID"
        else
            run_sql "
                INSERT INTO tool_calls(id, session_id, workflow_run_id, agent, tool_name, command, status, exit_code, latency_ms, created_at)
                VALUES ('$TCID', '$SID', $(sql_val "$WORKFLOW_RUN_ID"), $(sql_val "$AGENT"), '${TOOL_NAME//\'/\'\'}', $(sql_val "$COMMAND"), '$STATUS', ${EXIT_CODE:-NULL}, ${LATENCY_MS:-NULL}, '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"tool_call_id\":\"$TCID\",\"session_id\":\"$SID\",\"tool_name\":\"${TOOL_NAME//\"/\\\"}\",\"status\":\"$STATUS\"}"
        else
            echo "🔧 Tool call $TCID: $TOOL_NAME (status=$STATUS)"
        fi
        ;;

    record-gate)
        SID="${1:-}"
        GATE_NAME="${2:-}"
        GATE_TYPE="${3:-}"
        STATUS="${4:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-gate <session-id> <gate_name> <gate_type> <status> [workflow_run_id] [human_required] [reason]"; exit 1; }
        [ -z "$GATE_NAME" ] && { echo "Usage: session-db.sh record-gate <session-id> <gate_name> <gate_type> <status> [workflow_run_id] [human_required] [reason]"; exit 1; }
        [ -z "$GATE_TYPE" ] && { echo "Usage: session-db.sh record-gate <session-id> <gate_name> <gate_type> <status> [workflow_run_id] [human_required] [reason]"; exit 1; }
        [ -z "$STATUS" ] && { echo "Usage: session-db.sh record-gate <session-id> <gate_name> <gate_type> <status> [workflow_run_id] [human_required] [reason]"; exit 1; }
        WORKFLOW_RUN_ID="${5:-}"
        HUMAN_REQUIRED="${6:-0}"
        REASON="${7:-}"
        GID="gate_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record gate $GID"
        else
            run_sql "
                INSERT INTO gate_results(id, session_id, workflow_run_id, gate_name, gate_type, status, human_required, reason, created_at)
                VALUES ('$GID', '$SID', $(sql_val "$WORKFLOW_RUN_ID"), '${GATE_NAME//\'/\'\'}', '${GATE_TYPE//\'/\'\'}', '$STATUS', $HUMAN_REQUIRED, $(sql_val "$REASON"), '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"gate_id\":\"$GID\",\"session_id\":\"$SID\",\"gate_name\":\"${GATE_NAME//\"/\\\"}\",\"gate_type\":\"${GATE_TYPE//\"/\\\"}\",\"status\":\"$STATUS\"}"
        else
            echo "🛡️  Gate $GID: $GATE_NAME ($STATUS)"
        fi
        ;;

    record-ledger-ref)
        SEQ="${1:-}"
        EVENT_TYPE="${2:-}"
        EVENT_HASH="${3:-}"
        [ -z "$SEQ" ] && { echo "Usage: session-db.sh record-ledger-ref <seq> <event_type> <event_hash> [session_id] [workflow_run_id] [prev_hash]"; exit 1; }
        [ -z "$EVENT_TYPE" ] && { echo "Usage: session-db.sh record-ledger-ref <seq> <event_type> <event_hash> [session_id] [workflow_run_id] [prev_hash]"; exit 1; }
        [ -z "$EVENT_HASH" ] && { echo "Usage: session-db.sh record-ledger-ref <seq> <event_type> <event_hash> [session_id] [workflow_run_id] [prev_hash]"; exit 1; }
        SESSION_ID="${4:-}"
        WORKFLOW_RUN_ID="${5:-}"
        PREV_HASH="${6:-}"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record ledger ref seq $SEQ"
        else
            run_sql "
                INSERT OR REPLACE INTO ledger_refs(seq, session_id, workflow_run_id, event_type, event_hash, prev_hash, created_at)
                VALUES ($SEQ, $(sql_val "$SESSION_ID"), $(sql_val "$WORKFLOW_RUN_ID"), '${EVENT_TYPE//\'/\'\'}', '$EVENT_HASH', $(sql_val "$PREV_HASH"), '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"seq\":$SEQ,\"event_type\":\"${EVENT_TYPE//\"/\\\"}\",\"event_hash\":\"$EVENT_HASH\"}"
        else
            echo "📜 Ledger ref seq $SEQ: $EVENT_TYPE ($EVENT_HASH)"
        fi
        ;;

    record-eval-run)
        SUITE_NAME="${1:-}"
        DATASET_NAME="${2:-}"
        [ -z "$SUITE_NAME" ] && { echo "Usage: session-db.sh record-eval-run <suite_name> <dataset_name> [model] [provider] [status]"; exit 1; }
        [ -z "$DATASET_NAME" ] && { echo "Usage: session-db.sh record-eval-run <suite_name> <dataset_name> [model] [provider] [status]"; exit 1; }
        MODEL="${3:-}"
        PROVIDER="${4:-}"
        STATUS="${5:-running}"
        ERID="er_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record eval run $ERID"
        else
            run_sql "
                INSERT INTO eval_runs(id, suite_name, dataset_name, model, provider, status, started_at)
                VALUES ('$ERID', '${SUITE_NAME//\'/\'\'}', '${DATASET_NAME//\'/\'\'}', $(sql_val "$MODEL"), $(sql_val "$PROVIDER"), '$STATUS', '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"eval_run_id\":\"$ERID\",\"suite_name\":\"${SUITE_NAME//\"/\\\"}\",\"dataset_name\":\"${DATASET_NAME//\"/\\\"}\",\"status\":\"$STATUS\"}"
        else
            echo "📊 Eval run $ERID: $SUITE_NAME / $DATASET_NAME ($STATUS)"
        fi
        ;;

    record-eval-result)
        EVAL_RUN_ID="${1:-}"
        CASE_ID="${2:-}"
        STATUS="${3:-}"
        [ -z "$EVAL_RUN_ID" ] && { echo "Usage: session-db.sh record-eval-result <eval_run_id> <case_id> <status> [score] [expected_hash] [actual_hash] [judge_model] [reason]"; exit 1; }
        [ -z "$CASE_ID" ] && { echo "Usage: session-db.sh record-eval-result <eval_run_id> <case_id> <status> [score] [expected_hash] [actual_hash] [judge_model] [reason]"; exit 1; }
        [ -z "$STATUS" ] && { echo "Usage: session-db.sh record-eval-result <eval_run_id> <case_id> <status> [score] [expected_hash] [actual_hash] [judge_model] [reason]"; exit 1; }
        SCORE="${4:-}"
        EXPECTED_HASH="${5:-}"
        ACTUAL_HASH="${6:-}"
        JUDGE_MODEL="${7:-}"
        REASON="${8:-}"
        RESID="res_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record eval result $RESID"
        else
            run_sql "
                INSERT INTO eval_results(id, eval_run_id, case_id, status, score, expected_hash, actual_hash, judge_model, reason)
                VALUES ('$RESID', '$EVAL_RUN_ID', '${CASE_ID//\'/\'\'}', '$STATUS', ${SCORE:-NULL}, $(sql_val "$EXPECTED_HASH"), $(sql_val "$ACTUAL_HASH"), $(sql_val "$JUDGE_MODEL"), $(sql_val "$REASON"));
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"eval_result_id\":\"$RESID\",\"eval_run_id\":\"$EVAL_RUN_ID\",\"case_id\":\"${CASE_ID//\"/\\\"}\",\"status\":\"$STATUS\"}"
        else
            echo "📊 Eval result $RESID: $CASE_ID → $STATUS"
        fi
        ;;

    record-dataset-example)
        DATASET_NAME="${1:-}"
        SOURCE_TYPE="${2:-}"
        SOURCE_REF="${3:-}"
        INPUT_HASH="${4:-}"
        OUTPUT_HASH="${5:-}"
        [ -z "$DATASET_NAME" ] && { echo "Usage: session-db.sh record-dataset-example <dataset_name> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]"; exit 1; }
        [ -z "$SOURCE_TYPE" ] && { echo "Usage: session-db.sh record-dataset-example <dataset_name> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]"; exit 1; }
        [ -z "$SOURCE_REF" ] && { echo "Usage: session-db.sh record-dataset-example <dataset_name> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]"; exit 1; }
        [ -z "$INPUT_HASH" ] && { echo "Usage: session-db.sh record-dataset-example <dataset_name> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]"; exit 1; }
        [ -z "$OUTPUT_HASH" ] && { echo "Usage: session-db.sh record-dataset-example <dataset_name> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]"; exit 1; }
        LABEL="${6:-}"
        ACCEPTED="${7:-}"
        DEID="de_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record dataset example $DEID"
        else
            run_sql "
                INSERT INTO dataset_examples(id, dataset_name, source_type, source_ref, input_hash, output_hash, label, accepted, created_at)
                VALUES ('$DEID', '${DATASET_NAME//\'/\'\'}', '${SOURCE_TYPE//\'/\'\'}', '${SOURCE_REF//\'/\'\'}', '$INPUT_HASH', '$OUTPUT_HASH', $(sql_val "$LABEL"), ${ACCEPTED:-NULL}, '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"dataset_example_id\":\"$DEID\",\"dataset_name\":\"${DATASET_NAME//\"/\\\"}\",\"source_type\":\"${SOURCE_TYPE//\"/\\\"}\",\"input_hash\":\"$INPUT_HASH\",\"output_hash\":\"$OUTPUT_HASH\"}"
        else
            echo "📦 Dataset example $DEID: $DATASET_NAME / $SOURCE_TYPE"
        fi
        ;;

    record-provider-health)
        PROVIDER="${1:-}"
        [ -z "$PROVIDER" ] && { echo "Usage: session-db.sh record-provider-health <provider> [model] <status> [latency_ms] [error]"; exit 1; }
        MODEL="${2:-}"
        STATUS="${3:-ok}"
        LATENCY_MS="${4:-}"
        ERROR="${5:-}"
        PHID="ph_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record provider health $PHID"
        else
            run_sql "
                INSERT INTO provider_health(id, provider, model, status, latency_ms, error, checked_at)
                VALUES ('$PHID', '${PROVIDER//\'/\'\'}', $(sql_val "$MODEL"), '$STATUS', ${LATENCY_MS:-NULL}, $(sql_val "$ERROR"), '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"provider_health_id\":\"$PHID\",\"provider\":\"${PROVIDER//\"/\\\"}\",\"model\":\"${MODEL}\",\"status\":\"$STATUS\"}"
        else
            echo "🏥 Provider health $PHID: $PROVIDER ($STATUS)"
        fi
        ;;

    record-cost)
        SID="${1:-}"
        PROVIDER="${2:-}"
        MODEL="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-cost <session-id> <provider> <model> [workflow_run_id] [tokens_in] [tokens_out] [cached_tokens] [estimated_cost_usd]"; exit 1; }
        [ -z "$PROVIDER" ] && { echo "Usage: session-db.sh record-cost <session-id> <provider> <model> [workflow_run_id] [tokens_in] [tokens_out] [cached_tokens] [estimated_cost_usd]"; exit 1; }
        [ -z "$MODEL" ] && { echo "Usage: session-db.sh record-cost <session-id> <provider> <model> [workflow_run_id] [tokens_in] [tokens_out] [cached_tokens] [estimated_cost_usd]"; exit 1; }
        WORKFLOW_RUN_ID="${4:-}"
        TOKENS_IN="${5:-0}"
        TOKENS_OUT="${6:-0}"
        CACHED_TOKENS="${7:-0}"
        ESTIMATED_COST="${8:-0}"
        CSID="cs_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record cost snapshot $CSID"
        else
            run_sql "
                INSERT INTO cost_snapshots(id, session_id, workflow_run_id, provider, model, tokens_in, tokens_out, cached_tokens, estimated_cost_usd, created_at)
                VALUES ('$CSID', '$SID', $(sql_val "$WORKFLOW_RUN_ID"), '$PROVIDER', '$MODEL', ${TOKENS_IN:-0}, ${TOKENS_OUT:-0}, ${CACHED_TOKENS:-0}, ${ESTIMATED_COST:-0}, '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"cost_snapshot_id\":\"$CSID\",\"session_id\":\"$SID\",\"provider\":\"$PROVIDER\",\"model\":\"$MODEL\",\"estimated_cost_usd\":${ESTIMATED_COST:-0}}"
        else
            echo "💰 Cost snapshot $CSID: $PROVIDER/$MODEL ($ESTIMATED_COST)"
        fi
        ;;

    record-checkpoint)
        SID="${1:-}"
        CHECKPOINT_TYPE="${2:-}"
        SUMMARY="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh record-checkpoint <session-id> <checkpoint_type> <summary> [state_hash]"; exit 1; }
        [ -z "$CHECKPOINT_TYPE" ] && { echo "Usage: session-db.sh record-checkpoint <session-id> <checkpoint_type> <summary> [state_hash]"; exit 1; }
        [ -z "$SUMMARY" ] && { echo "Usage: session-db.sh record-checkpoint <session-id> <checkpoint_type> <summary> [state_hash]"; exit 1; }
        STATE_HASH="${4:-}"
        CPID="cp_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record checkpoint $CPID"
        else
            run_sql "
                INSERT INTO checkpoints(id, session_id, checkpoint_type, summary, state_hash, created_at)
                VALUES ('$CPID', '$SID', '$CHECKPOINT_TYPE', '${SUMMARY//\'/\'\'}', $(sql_val "$STATE_HASH"), '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"checkpoint_id\":\"$CPID\",\"session_id\":\"$SID\",\"checkpoint_type\":\"$CHECKPOINT_TYPE\",\"summary\":\"${SUMMARY//\"/\\\"}\"}"
        else
            echo "📌 Checkpoint $CPID: $CHECKPOINT_TYPE — $SUMMARY"
        fi
        ;;

    record-lesson)
        SOURCE="${1:-}"
        LESSON_TEXT="${2:-}"
        [ -z "$SOURCE" ] && { echo "Usage: session-db.sh record-lesson <source> <lesson_text> [session_id] [confidence] [accepted]"; exit 1; }
        [ -z "$LESSON_TEXT" ] && { echo "Usage: session-db.sh record-lesson <source> <lesson_text> [session_id] [confidence] [accepted]"; exit 1; }
        SESSION_ID="${3:-}"
        CONFIDENCE="${4:-}"
        ACCEPTED="${5:-}"
        LID="ls_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        if [[ "$DRY_RUN" == true ]]; then
            echo "[dry-run] Would record lesson $LID"
        else
            run_sql "
                INSERT INTO lessons(id, session_id, source, lesson_text, confidence, accepted, created_at)
                VALUES ('$LID', $(sql_val "$SESSION_ID"), '${SOURCE//\'/\'\'}', '${LESSON_TEXT//\'/\'\'}', ${CONFIDENCE:-NULL}, ${ACCEPTED:-NULL}, '$(now)');
            "
        fi
        if [[ "$JSON_OUT" == true ]]; then
            echo "{\"lesson_id\":\"$LID\",\"source\":\"${SOURCE//\"/\\\"}\",\"lesson_text\":\"${LESSON_TEXT//\"/\\\"}\"}"
        else
            echo "🎓 Lesson $LID: $SOURCE"
        fi
        ;;

    export-json)
        TABLE="${1:-sessions}"
        LIMIT="${2:-100}"
        run_sql_json "SELECT * FROM $TABLE LIMIT $LIMIT;"
        ;;

    # --- SESSIONS (v1 backward compatible) ---

    start)
        GOAL="${1:-unknown}"
        REPO="${2:-$(basename "$(pwd)")}"
        SID="ses_$(date +%s | tail -c 8)$(head -c 4 /dev/urandom | xxd -p | tr -d '\n')"
        MODEL="ollama-cloud/minimax-m2.7"
        init_db 2>/dev/null || true
        ROW_ID=$(run_sql "
            INSERT INTO sessions(id, started_at, agent, model, goal, repo, status)
            VALUES ('$SID', '$(now)', 'loop-driver', '$MODEL', '${GOAL//\'/\'\'}', '$REPO', 'active');
            SELECT last_insert_rowid();
        " 2>/dev/null)
        ledger_append session_start "{\"goal\":\"${GOAL//\'/\'\'}\",\"repo\":\"$REPO\",\"model\":\"$MODEL\"}" "$SID" "$ROW_ID"
        echo "$SID"
        ;;

    end)
        SID="${1:-}"
        STATUS="${2:-completed}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh end <session-id> [status]"; exit 1; }
        TOKENS=$(run_sql "SELECT COALESCE(SUM(token_used),0) FROM dispatches WHERE session_id='$SID';")
        run_sql "
            UPDATE sessions SET ended_at='$(now)', status='$STATUS', token_total=$TOKENS WHERE id='$SID';
        "
        ledger_append session_end "{\"status\":\"$STATUS\",\"token_total\":$TOKENS}" "$SID"
        echo "✅ Session $SID ended ($STATUS, $TOKENS tokens)"
        ;;

    recent)
        LIMIT="${1:-10}"
        run_sql_table "
            SELECT started_at, status, goal, repo, token_total, id
            FROM sessions ORDER BY started_at DESC LIMIT $LIMIT;
        "
        ;;

    # --- DISPATCHES ---

    dispatch)
        SID="${1:-}"
        AGENT="${2:-}"
        TASK="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh dispatch <session-id> <agent> <task> [phase] [workflow] [mode]"; exit 1; }
        PHASE="${4:-}"
        WORKFLOW="${5:-}"
        MODE="${6:-}"
        SEQ=$(run_sql "SELECT COALESCE(MAX(seq),0)+1 FROM dispatches WHERE session_id='$SID';")
        MODEL=$(run_sql "SELECT model FROM sessions WHERE id='$SID' LIMIT 1;" || echo "unknown")
        ROW_ID=$(run_sql "
            INSERT INTO dispatches(session_id, seq, agent, model, task, phase, workflow, mode, started_at)
            VALUES ('$SID', $SEQ, '$AGENT', '$MODEL', '${TASK//\'/\'\'}', '${PHASE}', '${WORKFLOW}', '${MODE}', '$(now)');
            SELECT last_insert_rowid();
        ")
        ledger_append dispatch "{\"agent\":\"$AGENT\",\"task\":\"${TASK//\'/\'\'}\",\"phase\":\"${PHASE}\",\"mode\":\"${MODE}\",\"workflow\":\"${WORKFLOW}\"}" "$SID" "$ROW_ID"
        echo "📤 $AGENT → $TASK (seq $SEQ, session $SID)"
        ;;

    complete)
        SID="${1:-}"
        SEQ="${2:-}"
        RESULT="${3:-pass}"
        SUMMARY="${4:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh complete <session-id> <seq> <result> [summary] [files_touched]"; exit 1; }
        FILES="${5:-}"
        run_sql "
            UPDATE dispatches
            SET ended_at='$(now)', result='$RESULT',
                summary='${SUMMARY//\'/\'\'}',
                files_touched='${FILES}'
            WHERE session_id='$SID' AND seq=$SEQ;
        "
        ledger_append dispatch_complete "{\"seq\":$SEQ,\"result\":\"$RESULT\",\"summary\":\"${SUMMARY//\'/\'\'}\"}" "$SID"
        echo "✅ Dispatch $SID seq $SEQ → $RESULT"
        ;;

    error)
        SID="${1:-}"
        SEQ="${2:-}"
        ERR="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh error <session-id> <seq> <error-message>"; exit 1; }
        run_sql "
            UPDATE dispatches SET result='fail', error_message='${ERR//\'/\'\'}' WHERE session_id='$SID' AND seq=$SEQ;
        "
        ledger_append dispatch_fail "{\"seq\":$SEQ,\"error_message\":\"${ERR//\'/\'\'}\"}" "$SID"
        echo "❌ Error recorded: $SID seq $SEQ"
        ;;

    # --- DECISIONS ---

    decide)
        SID="${1:-}"
        TITLE="${2:-}"
        DECISION="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh decide <session-id> <title> <decision> [rationale] [alternatives] [tags]"; exit 1; }
        RATIONALE="${4:-}"
        ALTS="${5:-}"
        TAGS="${6:-}"
        ROW_ID=$(run_sql "
            INSERT INTO decisions(session_id, title, decision, rationale, alternatives, created_at, tags)
            VALUES ('$SID', '${TITLE//\'/\'\'}', '${DECISION//\'/\'\'}', '${RATIONALE//\'/\'\'}', '${ALTS//\'/\'\'}', '$(now)', '${TAGS}');
            SELECT last_insert_rowid();
        ")
        ledger_append decision "{\"title\":\"${TITLE//\'/\'\'}\",\"decision\":\"${DECISION//\'/\'\'}\",\"rationale\":\"${RATIONALE//\'/\'\'}\",\"alternatives\":\"${ALTS//\'/\'\'}\",\"tags\":\"${TAGS}\"}" "$SID" "$ROW_ID"
        echo "📝 Decision recorded: $TITLE"
        ;;

    decisions)
        TAG="${1:-}"
        if [ -n "$TAG" ]; then
            run_sql_table "SELECT title, decision, rationale, created_at FROM decisions WHERE tags LIKE '%${TAG}%' ORDER BY created_at DESC;"
        else
            run_sql_table "SELECT title, decision, created_at FROM decisions ORDER BY created_at DESC LIMIT 20;"
        fi
        ;;

    # --- MEMORIES ---

    memory)
        SID="${1:-}"
        TITLE="${2:-}"
        CONTENT="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh memory <session-id> <title> <content> [tags] [importance]"; exit 1; }
        TAGS="${4:-}"
        IMPORTANCE="${5:-normal}"
        ROW_ID=$(run_sql "
            INSERT INTO memories(session_id, title, content, tags, importance, created_at)
            VALUES ('$SID', '${TITLE//\'/\'\'}', '${CONTENT//\'/\'\'}', '${TAGS}', '${IMPORTANCE}', '$(now)');
            SELECT last_insert_rowid();
        ")
        ledger_append memory "{\"title\":\"${TITLE//\'/\'\'}\",\"content\":\"${CONTENT//\'/\'\'}\",\"tags\":\"${TAGS}\",\"importance\":\"${IMPORTANCE}\"}" "$SID" "$ROW_ID"
        echo "🧠 Memory saved: $TITLE"
        ;;

    memories)
        TAG="${1:-}"
        if [ -n "$TAG" ]; then
            run_sql_table "SELECT title, content, importance, created_at FROM memories WHERE tags LIKE '%${TAG}%' OR title LIKE '%${TAG}%' ORDER BY created_at DESC LIMIT 20;"
        else
            run_sql_table "SELECT title, importance, created_at FROM memories ORDER BY created_at DESC LIMIT 20;"
        fi
        ;;

    # --- ARTIFACTS ---

    artifact)
        SID="${1:-}"
        ARTIFACT_PATH="${2:-}"
        ACTION="${3:-modified}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh artifact <session-id> <path> [action]"; exit 1; }
        ROW_ID=$(run_sql "
            INSERT INTO artifacts(session_id, path, action, created_at)
            VALUES ('$SID', '${ARTIFACT_PATH}', '${ACTION}', '$(now)');
            SELECT last_insert_rowid();
        ")
        ledger_append artifact "{\"path\":\"${ARTIFACT_PATH}\",\"action\":\"${ACTION}\"}" "$SID" "$ROW_ID"
        ;;

    # --- TOKEN LOGGING ---

    token)
        SID="${1:-}"
        COUNT="${2:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh token <session-id> <token-count> [context-pct]"; exit 1; }
        PCT="${3:-}"
        ROW_ID=$(run_sql "
            INSERT INTO token_log(session_id, token_count, context_pct, recorded_at)
            VALUES ('$SID', $COUNT, ${PCT:-NULL}, '$(now)');
            SELECT last_insert_rowid();
        ")
        ledger_append token_log "{\"token_count\":$COUNT,\"context_pct\":${PCT:-null}}" "$SID" "$ROW_ID"
        echo "📊 Token logged: $COUNT"
        ;;

    # --- PARALLEL TASKS ---

    ptask)
        SID="${1:-}"
        AGENT="${2:-}"
        TASK="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh ptask <sid> <agent> <task> [wave_id] [parent_dispatch_id]"; exit 1; }
        WAVE="${4:-wave_1}"
        PARENT="${5:-}"
        run_sql "
            INSERT INTO parallel_tasks(session_id, parent_dispatch_id, wave_id, agent, task, created_at)
            VALUES ('$SID', ${PARENT:-NULL}, '$WAVE', '$AGENT', '${TASK//\'/\'\'}', '$(now)');
        "
        PTID=$(run_sql "SELECT last_insert_rowid();")
        echo "🔄 Parallel task $PTID → $AGENT: $TASK (wave: $WAVE)"
        ;;

    ptask-start)
        PTID="${1:-}"
        [ -z "$PTID" ] && { echo "Usage: session-db.sh ptask-start <ptask-id>"; exit 1; }
        run_sql "UPDATE parallel_tasks SET status='running', started_at='$(now)' WHERE id=$PTID;"
        echo "▶️  Parallel task $PTID started"
        ;;

    ptask-done)
        PTID="${1:-}"
        RESULT="${2:-pass}"
        OUTPUT="${3:-}"
        [ -z "$PTID" ] && { echo "Usage: session-db.sh ptask-done <ptask-id> <result> [output_path]"; exit 1; }
        run_sql "
            UPDATE parallel_tasks SET status='completed', result='$RESULT', output_path='${OUTPUT}', completed_at='$(now)' WHERE id=$PTID;
        "
        echo "✅ Parallel task $PTID → $RESULT"
        ;;

    ptask-fail)
        PTID="${1:-}"
        ERR="${2:-}"
        [ -z "$PTID" ] && { echo "Usage: session-db.sh ptask-fail <ptask-id> <error>"; exit 1; }
        run_sql "
            UPDATE parallel_tasks SET status='failed', result='fail', error_message='${ERR//\'/\'\'}', completed_at='$(now)' WHERE id=$PTID;
        "
        echo "❌ Parallel task $PTID failed"
        ;;

    ptask-wave)
        SID="${1:-}"
        WAVE="${2:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh ptask-wave <sid> [wave_id]"; exit 1; }
        if [ -n "$WAVE" ]; then
            run_sql_table "SELECT id, agent, task, status, started_at, completed_at FROM parallel_tasks WHERE session_id='$SID' AND wave_id='$WAVE' ORDER BY id;"
        else
            run_sql_table "SELECT wave_id, COUNT(*) AS total, COUNT(CASE WHEN status='completed' THEN 1 END) AS done, COUNT(CASE WHEN status='failed' THEN 1 END) AS failed, COUNT(CASE WHEN status='pending' THEN 1 END) AS pending FROM parallel_tasks WHERE session_id='$SID' GROUP BY wave_id;"
        fi
        ;;

    # --- MESSAGES ---

    msg-send)
        SID="${1:-}"
        FROM="${2:-}"
        TO="${3:-}"
        SUBJECT="${4:-}"
        BODY="${5:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh msg-send <sid> <from> <to> <subject> <body> [priority]"; exit 1; }
        PRIORITY="${6:-normal}"
        run_sql "
            INSERT INTO messages(session_id, from_agent, to_agent, subject, body, priority, created_at)
            VALUES ('$SID', '$FROM', '$TO', '${SUBJECT//\'/\'\'}', '${BODY//\'/\'\'}', '$PRIORITY', '$(now)');
        "
        echo "📨 Message from $FROM → $TO: $SUBJECT"
        ;;

    msg-recv)
        AGENT="${1:-}"
        SID="${2:-}"
        [ -z "$AGENT" ] && { echo "Usage: session-db.sh msg-recv <agent> [session-id]"; exit 1; }
        if [ -n "$SID" ]; then
            run_sql_table "SELECT id, from_agent, subject, body, priority, created_at FROM messages WHERE to_agent='$AGENT' AND session_id='$SID' AND status='unread' ORDER BY CASE priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 ELSE 3 END, created_at;"
        else
            run_sql_table "SELECT id, session_id, from_agent, subject, body, priority, created_at FROM messages WHERE to_agent='$AGENT' AND status='unread' ORDER BY CASE priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 ELSE 3 END, created_at;"
        fi
        ;;

    msg-read)
        MSGID="${1:-}"
        [ -z "$MSGID" ] && { echo "Usage: session-db.sh msg-read <msg-id>"; exit 1; }
        run_sql "UPDATE messages SET status='read', read_at='$(now)' WHERE id=$MSGID;"
        echo "📖 Message $MSGID marked as read"
        ;;

    msg-history)
        SID="${1:-}"
        AGENT="${2:-}"
        if [ -n "$SID" ] && [ -n "$AGENT" ]; then
            run_sql_table "SELECT id, from_agent, to_agent, subject, status, priority, created_at FROM messages WHERE session_id='$SID' AND (from_agent='$AGENT' OR to_agent='$AGENT') ORDER BY created_at DESC;"
        elif [ -n "$AGENT" ]; then
            run_sql_table "SELECT id, session_id, from_agent, to_agent, subject, status, priority, created_at FROM messages WHERE from_agent='$AGENT' OR to_agent='$AGENT' ORDER BY created_at DESC LIMIT 20;"
        else
            run_sql_table "SELECT id, session_id, from_agent, to_agent, subject, status, priority, created_at FROM messages ORDER BY created_at DESC LIMIT 20;"
        fi
        ;;

    # --- BARRIERS ---

    barrier-create)
        SID="${1:-}"
        BID="${2:-}"
        COUNT="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh barrier-create <sid> <barrier-id> <expected-count>"; exit 1; }
        run_sql "
            INSERT INTO barriers(session_id, barrier_id, expected_count, created_at)
            VALUES ('$SID', '$BID', $COUNT, '$(now)');
        "
        echo "🚧 Barrier $BID created (expecting $COUNT arrivals)"
        ;;

    barrier-arrive)
        SID="${1:-}"
        BID="${2:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh barrier-arrive <sid> <barrier-id>"; exit 1; }
        run_sql "
            UPDATE barriers SET arrived_count = arrived_count + 1 WHERE session_id='$SID' AND barrier_id='$BID';
        "
        ARRIVED=$(run_sql "SELECT arrived_count FROM barriers WHERE session_id='$SID' AND barrier_id='$BID';")
        EXPECTED=$(run_sql "SELECT expected_count FROM barriers WHERE session_id='$SID' AND barrier_id='$BID';")
        if [ "$ARRIVED" -ge "$EXPECTED" ]; then
            run_sql "UPDATE barriers SET status='resolved', resolved_at='$(now)' WHERE session_id='$SID' AND barrier_id='$BID';"
            echo "🚩 Barrier $BID resolved ($ARRIVED/$EXPECTED)"
        else
            echo "⏳ Barrier $BID: $ARRIVED/$EXPECTED arrived"
        fi
        ;;

    barrier-status)
        SID="${1:-}"
        BID="${2:-}"
        if [ -n "$SID" ] && [ -n "$BID" ]; then
            run_sql_table "SELECT barrier_id, expected_count, arrived_count, status, created_at, resolved_at FROM barriers WHERE session_id='$SID' AND barrier_id='$BID';"
        elif [ -n "$SID" ]; then
            run_sql_table "SELECT barrier_id, expected_count, arrived_count, status FROM barriers WHERE session_id='$SID' ORDER BY created_at DESC;"
        fi
        ;;

    # --- LOCKS ---

    lock-acquire)
        SID="${1:-}"
        LNAME="${2:-}"
        HOLDER="${3:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh lock-acquire <sid> <lock-name> <holder>"; exit 1; }
        EXISTING=$(run_sql "SELECT id FROM locks WHERE lock_name='$LNAME' AND status='active';")
        if [ -n "$EXISTING" ]; then
            echo "🔒 Lock '$LNAME' already held (id: $EXISTING)"
            exit 1
        fi
        run_sql "
            INSERT INTO locks(session_id, lock_name, held_by, acquired_at)
            VALUES ('$SID', '$LNAME', '$HOLDER', '$(now)');
        "
        echo "🔓 Lock '$LNAME' acquired by $HOLDER"
        ;;

    lock-release)
        SID="${1:-}"
        LNAME="${2:-}"
        [ -z "$SID" ] && { echo "Usage: session-db.sh lock-release <sid> <lock-name>"; exit 1; }
        run_sql "UPDATE locks SET status='released', released_at='$(now)' WHERE session_id='$SID' AND lock_name='$LNAME' AND status='active';"
        echo "🔓 Lock '$LNAME' released"
        ;;

    lock-status)
        SID="${1:-}"
        if [ -n "$SID" ]; then
            run_sql_table "SELECT lock_name, held_by, status, acquired_at, released_at FROM locks WHERE session_id='$SID' AND status='active' ORDER BY acquired_at;"
        else
            run_sql_table "SELECT session_id, lock_name, held_by, status, acquired_at FROM locks WHERE status='active' ORDER BY acquired_at;"
        fi
        ;;

    # --- WORKFLOW ENGINE ---

    team-create)
        TNAME="${1:-}"
        AGENTS="${2:-}"
        DESC="${3:-}"
        [ -z "$TNAME" ] || [ -z "$AGENTS" ] && { echo "Usage: session-db.sh team-create <name> <agents_csv> [description]"; exit 1; }
        init_db 2>/dev/null || true
        run_sql "
            INSERT OR REPLACE INTO teams(name, agents, description, created_at)
            VALUES ('$TNAME', '$AGENTS', '${DESC//\'/\'\'}', '$(now)');
        "
        echo "👥 Team '$TNAME' created: $AGENTS"
        ;;

    team-list)
        run_sql_table "SELECT name, agents, description, created_at FROM teams ORDER BY created_at;"
        ;;

    team-delete)
        TNAME="${1:-}"
        [ -z "$TNAME" ] && { echo "Usage: session-db.sh team-delete <name>"; exit 1; }
        run_sql "DELETE FROM teams WHERE name='$TNAME';"
        echo "👥 Team '$TNAME' deleted"
        ;;

    workflow-create)
        WNAME="${1:-}"
        TEAM="${2:-}"
        STEPS="${3:-}"
        DESC="${4:-}"
        [ -z "$WNAME" ] || [ -z "$TEAM" ] || [ -z "$STEPS" ] && { echo "Usage: session-db.sh workflow-create <name> <team> <steps_json> [description]"; exit 1; }
        echo "⚠️  workflow-create is deprecated. Use workflow-run.sh sync for YAML-backed workflows." >&2
        init_db 2>/dev/null || true
        run_sql "
            INSERT OR REPLACE INTO workflows(id, name, team, steps, description, created_at)
            VALUES ('$WNAME', '$WNAME', '$TEAM', '${STEPS//\'/\'\'}', '${DESC//\'/\'\'}', '$(now)');
        "
        echo "🔄 Workflow '$WNAME' created (team: $TEAM, steps: $STEPS)"
        ;;

    workflow-list)
        run_sql_table "SELECT name, team, description, steps, created_at FROM workflows ORDER BY created_at;"
        ;;

    workflow-delete)
        WNAME="${1:-}"
        [ -z "$WNAME" ] && { echo "Usage: session-db.sh workflow-delete <name>"; exit 1; }
        run_sql "DELETE FROM workflows WHERE name='$WNAME';"
        echo "🔄 Workflow '$WNAME' deleted"
        ;;

    workflow-start)
        WNAME="${1:-}"
        SID="${2:-}"
        ON_FAILURE="${3:-stop}"
        [ -z "$WNAME" ] && { echo "Usage: session-db.sh workflow-start <workflow-name> <session-id> [on_failure]"; exit 1; }
        init_db 2>/dev/null || true
        CONFIG=$(run_sql "SELECT config_json FROM workflows WHERE name='$WNAME' LIMIT 1;")
        [ -z "$CONFIG" ] && { echo "Workflow '$WNAME' not found"; exit 1; }
        STEP_COUNT=$(echo "$CONFIG" | python3 -c "import json,sys; print(len(json.load(sys.stdin).get('phases',[])))" 2>/dev/null || echo "0")
        [ "$STEP_COUNT" -eq 0 ] && { echo "Workflow '$WNAME' has no phases"; exit 1; }
        # Insert instance
        WIID=$(run_sql "
            INSERT INTO workflow_instances(workflow_name, session_id, total_steps, on_failure, started_at)
            VALUES ('$WNAME', '${SID}', $STEP_COUNT, '${ON_FAILURE}', '$(now)');
            SELECT last_insert_rowid();
        ")
        # Parse config_json phases and insert each step via Python
        WIID="$WIID" python3 -c '
import json, sys, os
config = json.load(open("/dev/stdin"))
phases = config.get("phases", [])
wiid = os.environ["WIID"]
for i, p in enumerate(phases):
    agent = p.get("agent","")
    mode = p.get("mode","")
    task = p.get("name","") + ": " + p.get("feedforward","")
    # Escape single quotes for SQL
    agent = agent.replace(chr(39), chr(39)+chr(39))
    task = task.replace(chr(39), chr(39)+chr(39))
    mode = mode.replace(chr(39), chr(39)+chr(39))
    print(f"INSERT INTO workflow_steps(instance_id, step_order, agent, task, mode) VALUES({wiid}, {i+1}, '\''{agent}'\'', '\''{task}'\'', '\''{mode}'\'');")
' <<< "$CONFIG" | sqlite3 "$DB_PATH" > /dev/null
        echo "$WIID"
        ;;
    workflow-status)
        WIID="${1:-}"
        [ -z "$WIID" ] && { echo "Usage: session-db.sh workflow-status <instance-id>"; exit 1; }
        run_sql_table "SELECT id, workflow_name, status, current_step, total_steps, result, started_at, completed_at FROM workflow_instances WHERE id=$WIID;"
        echo "--- Steps ---"
        run_sql_table "SELECT step_order, agent, task, mode, status, result, error_message FROM workflow_steps WHERE instance_id=$WIID ORDER BY step_order;"
        ;;

    workflow-step-start)
        WIID="${1:-}"
        [ -z "$WIID" ] && { echo "Usage: session-db.sh workflow-step-start <instance-id>"; exit 1; }
        CURRENT=$(run_sql "SELECT current_step FROM workflow_instances WHERE id=$WIID;")
        NEXT=$((CURRENT + 1))
        run_sql "
            UPDATE workflow_steps SET status='running', started_at='$(now)' WHERE instance_id=$WIID AND step_order=$NEXT;
            UPDATE workflow_instances SET status='running', current_step=$NEXT WHERE id=$WIID;
        "
        STEP=$(run_sql "SELECT agent, task, mode FROM workflow_steps WHERE instance_id=$WIID AND step_order=$NEXT;")
        echo "▶️  Step $NEXT started for instance $WIID: $STEP"
        ;;

    workflow-step-done)
        WIID="${1:-}"
        RESULT="${2:-pass}"
        OUTPUT="${3:-}"
        [ -z "$WIID" ] && { echo "Usage: session-db.sh workflow-step-done <instance-id> <result> [output_path]"; exit 1; }
        CURRENT=$(run_sql "SELECT current_step FROM workflow_instances WHERE id=$WIID;")
        run_sql "
            UPDATE workflow_steps SET status='completed', result='$RESULT', output_path='${OUTPUT}', completed_at='$(now)' WHERE instance_id=$WIID AND step_order=$CURRENT;
        "
        TOTAL=$(run_sql "SELECT total_steps FROM workflow_instances WHERE id=$WIID;")
        if [ "$CURRENT" -ge "$TOTAL" ]; then
            run_sql "
                UPDATE workflow_instances SET status='completed', result='$RESULT', completed_at='$(now)' WHERE id=$WIID;
            "
            echo "✅ Workflow instance $WIID completed"
        else
            echo "✅ Step $CURRENT done, $((TOTAL - CURRENT)) steps remaining"
        fi
        ;;

    workflow-fail)
        WIID="${1:-}"
        ERR="${2:-}"
        [ -z "$WIID" ] && { echo "Usage: session-db.sh workflow-fail <instance-id> <error>"; exit 1; }
        CURRENT=$(run_sql "SELECT current_step FROM workflow_instances WHERE id=$WIID;")
        run_sql "
            UPDATE workflow_steps SET status='failed', error_message='${ERR//\'/\'\'}' WHERE instance_id=$WIID AND step_order=$CURRENT;
            UPDATE workflow_instances SET status='failed', error_message='${ERR//\'/\'\'}', completed_at='$(now)' WHERE id=$WIID;
        "
        echo "❌ Workflow instance $WIID failed at step $CURRENT"
        ;;

    # --- ANALYTICS ---

    metric-record)
        SID="${1:-}"
        REPO="${2:-}"
        GATE="${3:-}"
        PASSED="${4:-}"
        [ -z "$SID" ] || [ -z "$REPO" ] || [ -z "$GATE" ] || [ -z "$PASSED" ] && { echo "Usage: session-db.sh metric-record <sid> <repo> <gate_type> <passed> [duration_ms] [errors] [warnings]"; exit 1; }
        DUR="${5:-}"
        ERRS="${6:-0}"
        WARNS="${7:-0}"
        run_sql "
            INSERT INTO quality_metrics(session_id, repo, gate_type, passed, duration_ms, error_count, warning_count, timestamp)
            VALUES ('$SID', '$REPO', '$GATE', $PASSED, ${DUR:-NULL}, $ERRS, $WARNS, '$(now)');
        "
        echo "📊 Metric recorded: $REPO/$GATE → passed=$PASSED"
        ;;

    metric-history)
        REPO="${1:-}"
        GATE="${2:-}"
        LIMIT="${3:-20}"
        [ -z "$REPO" ] && { echo "Usage: session-db.sh metric-history <repo> [gate_type] [limit]"; exit 1; }
        if [ -n "$GATE" ]; then
            run_sql_table "SELECT session_id, gate_type, passed, duration_ms, error_count, warning_count, timestamp FROM quality_metrics WHERE repo='$REPO' AND gate_type='$GATE' ORDER BY timestamp DESC LIMIT $LIMIT;"
        else
            run_sql_table "SELECT session_id, gate_type, passed, duration_ms, error_count, warning_count, timestamp FROM quality_metrics WHERE repo='$REPO' ORDER BY timestamp DESC LIMIT $LIMIT;"
        fi
        ;;

    metric-trend)
        REPO="${1:-}"
        DAYS="${2:-30}"
        [ -z "$REPO" ] && { echo "Usage: session-db.sh metric-trend <repo> [days]"; exit 1; }
        run_sql_table "
            SELECT date(timestamp) AS date, gate_type,
                   COUNT(*) AS total,
                   COUNT(CASE WHEN passed=1 THEN 1 END) AS passed,
                   COUNT(CASE WHEN passed=0 THEN 1 END) AS failed
            FROM quality_metrics
            WHERE repo='$REPO' AND timestamp >= date('now', '-$DAYS days')
            GROUP BY date(timestamp), gate_type
            ORDER BY date(timestamp) DESC;
        "
        ;;

    stats)
        echo "=== Session Stats ==="
        run_sql "
            SELECT
                COUNT(*) AS total_sessions,
                COUNT(CASE WHEN status='completed' THEN 1 END) AS completed,
                COUNT(CASE WHEN status='active' THEN 1 END) AS active,
                COALESCE(SUM(token_total),0) AS total_tokens
            FROM sessions;
        "
        echo ""
        echo "=== Agent Stats ==="
        run_sql_table "
            SELECT agent, COUNT(*) AS dispatches,
                   COUNT(CASE WHEN result='pass' THEN 1 END) AS passes,
                   COUNT(CASE WHEN result='fail' THEN 1 END) AS failures,
                   COALESCE(SUM(token_used),0) AS tokens
            FROM dispatches GROUP BY agent ORDER BY dispatches DESC;
        "
        echo ""
        echo "=== Failure Analysis ==="
        run_sql_table "
            SELECT agent, task, error_message, started_at
            FROM dispatches WHERE result='fail' ORDER BY started_at DESC LIMIT 10;
        "
        ;;

    query)
        SQL="${*:-}"
        [ -z "$SQL" ] && { echo "Usage: session-db.sh query '<sql>'"; exit 1; }
        run_sql_table "$SQL"
        ;;

    # --- HELP ---

    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║           🧠 SESSION DATABASE — SQLite Memory            ║
╚══════════════════════════════════════════════════════════╝

Database: .specify/session.db
Global flags: --db <path>  --json  --dry-run

V2 COMMANDS:
  session-db.sh migrate                        Run pending migrations
  session-db.sh backup                       Create timestamped backup
  session-db.sh verify                       Check v2 schema completeness
  session-db.sh start-session <goal> [repo]  Start session → returns ID
  session-db.sh end-session <id> [status]      End session
  session-db.sh start-workflow <sid> <name> [mode]   Start workflow run
  session-db.sh end-workflow <wid> [status] [latency_ms]  End workflow run
  session-db.sh start-phase <wid> <name> [agent]       Start phase
  session-db.sh end-phase <pid> [status] [latency_ms]   End phase
  session-db.sh record-dispatch <sid> <to_agent> <task_type> [from] [mode] [risk] [reason] [wid]
  session-db.sh record-model-call <sid> <provider> <model> <purpose> [wid] [agent] [tokens] [cost] [success]
  session-db.sh record-tool-call <sid> <tool_name> [command] [wid] [agent] [status] [exit] [latency]
  session-db.sh record-gate <sid> <name> <type> <status> [wid] [human_req] [reason]
  session-db.sh record-ledger-ref <seq> <type> <hash> [sid] [wid] [prev_hash]
  session-db.sh record-eval-run <suite> <dataset> [model] [provider] [status]
  session-db.sh record-eval-result <erid> <case_id> <status> [score] [expected] [actual] [judge] [reason]
  session-db.sh record-dataset-example <dataset> <source_type> <source_ref> <input_hash> <output_hash> [label] [accepted]
  session-db.sh record-provider-health <provider> [model] [status] [latency_ms] [error]
  session-db.sh record-cost <sid> <provider> <model> [wid] [tokens_in] [tokens_out] [cached] [cost_usd]
  session-db.sh record-checkpoint <sid> <type> <summary> [state_hash]
  session-db.sh record-lesson <source> <text> [sid] [confidence] [accepted]
  session-db.sh export-json <table> [limit]  Export table as JSON

V1 BACKWARD COMPATIBLE:
  session-db.sh init                          Create database
  session-db.sh start <goal> [repo]          Start session → returns ID
  session-db.sh end <id> [status]            End session
  session-db.sh dispatch <sid> <agent> <task> [phase] [wf] [mode]
  session-db.sh complete <sid> <seq> <result> [summary] [files]
  session-db.sh error <sid> <seq> <error>
  session-db.sh recent [N]                    Show recent sessions

DISPATCHES:
  session-db.sh dispatch <sid> <agent> <task> [phase] [wf] [mode]
  session-db.sh complete <sid> <seq> <result> [summary] [files]
  session-db.sh error <sid> <seq> <error>

KNOWLEDGE:
  session-db.sh decide <sid> <title> <decision> [rationale] [alts] [tags]
  session-db.sh decisions [tag]               Search decisions
  session-db.sh memory <sid> <title> <content> [tags] [importance]
  session-db.sh memories [tag]                Search memories
  session-db.sh artifact <sid> <path> [action]    Track file changes

MONITORING:
  session-db.sh token <sid> <count> [pct]     Log token usage

PARALLEL EXECUTION:
  session-db.sh ptask <sid> <agent> <task> [wave] [parent]  Register parallel task
  session-db.sh ptask-start <ptask-id>        Mark parallel task running
  session-db.sh ptask-done <ptask-id> <result> [output]     Mark parallel task complete
  session-db.sh ptask-fail <ptask-id> <error> Mark parallel task failed
  session-db.sh ptask-wave <sid> [wave_id]    Show wave status

MESSAGES:
  session-db.sh msg-send <sid> <from> <to> <subject> <body> [priority]
  session-db.sh msg-recv <agent> [sid]        Check unread messages
  session-db.sh msg-read <msg-id>             Mark message as read
  session-db.sh msg-history [sid] [agent]     Message history

COORDINATION:
  session-db.sh barrier-create <sid> <id> <count>  Create sync barrier
  session-db.sh barrier-arrive <sid> <id>          Arrive at barrier
  session-db.sh barrier-status <sid> [id]          Check barrier status
  session-db.sh lock-acquire <sid> <name> <holder> Acquire exclusive lock
  session-db.sh lock-release <sid> <name>          Release lock
  session-db.sh lock-status [sid]                  Show active locks

WORKFLOW ENGINE:
  session-db.sh team-create <name> <agents_csv> [desc]      Create a team
  session-db.sh team-list                                     List all teams
  session-db.sh team-delete <name>                            Delete a team
  session-db.sh workflow-create <name> <team> <steps_json> [desc]  Define a workflow
  session-db.sh workflow-list                                 List all workflows
  session-db.sh workflow-delete <name>                        Delete a workflow
  session-db.sh workflow-start <name> <sid>                   Start a workflow instance
  session-db.sh workflow-status <iid>                        Show instance + steps status
  session-db.sh workflow-step-start <iid>                     Mark current step running
  session-db.sh workflow-step-done <iid> <result> [output]   Mark step done / advance
  session-db.sh workflow-fail <iid> <error>                    Mark instance failed

ANALYTICS:
  session-db.sh stats                         Agent + failure + token stats
  session-db.sh metric-record <sid> <repo> <gate_type> <passed> [duration_ms] [errors] [warnings]
  session-db.sh metric-history <repo> [gate_type] [limit]    Show recent quality gate runs
  session-db.sh metric-trend <repo> [days]                    Show pass/fail trend over N days
  session-db.sh query '<sql>'                 Run any SQL query

EXAMPLES:
  SID=$(session-db.sh start "Add auth middleware" fedora)
  session-db.sh dispatch $SID turbo-crank "Clarify auth requirements" clarify rpi
  session-db.sh complete $SID 1 pass "Confirmed 3 auth requirements"
  session-db.sh decide $SID "Use JWT" "Bearer tokens" "Stateless" "Session cookies" auth
  session-db.sh token $SID 15000 60
  session-db.sh end $SID completed

FUTURE AGENT QUERIES:
  "Show all decisions tagged 'auth'"  → decisions auth
  "Show wall-builder failures"        → query "SELECT * FROM dispatchers WHERE agent='wall-builder' AND result='fail'"
  "What did we do Tuesday?"           → recent 20
HELP
        ;;
esac
