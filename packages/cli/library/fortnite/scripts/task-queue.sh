#!/usr/bin/env bash
# task-queue.sh — SQLite-backed task queue for multi-agent parallel execution
# Database: .specify/session.db (shared with session-db.sh)
# Usage: ./task-queue.sh <command> [args...]

set -euo pipefail

DB_PATH="${DB_PATH:-${OPENCODE_WORKSPACE:-.}/.specify/session.db}"
CMD="${1:-help}"
shift 2>/dev/null || true

now() { date -u '+%Y-%m-%dT%H:%M:%SZ'; }

sqlite_queue() {
    sqlite3 -cmd ".timeout 5000" "$@"
}

init_db() {
    sqlite_queue "$DB_PATH" "
        CREATE TABLE IF NOT EXISTS task_queue (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT NOT NULL,
            topic TEXT NOT NULL,
            task TEXT NOT NULL,
            status TEXT DEFAULT 'open',
            max_agents INTEGER NOT NULL DEFAULT 1,
            dedupe_key TEXT,
            created_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS task_claims (
            task_id INTEGER NOT NULL REFERENCES task_queue(id),
            agent TEXT NOT NULL,
            claimed_at TEXT NOT NULL,
            PRIMARY KEY (task_id, agent)
        );

        CREATE TABLE IF NOT EXISTS task_dlq (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER REFERENCES task_queue(id),
            failed_agent TEXT,
            error_message TEXT,
            context_dump TEXT,
            failed_at TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS task_messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER NOT NULL REFERENCES task_queue(id),
            from_agent TEXT NOT NULL,
            to_agent TEXT,
            body TEXT NOT NULL,
            created_at TEXT NOT NULL
        );

        CREATE INDEX IF NOT EXISTS idx_task_queue_session ON task_queue(session_id);
        CREATE INDEX IF NOT EXISTS idx_task_queue_topic ON task_queue(topic);
        CREATE INDEX IF NOT EXISTS idx_task_queue_status ON task_queue(status);
        CREATE INDEX IF NOT EXISTS idx_task_claims_task ON task_claims(task_id);
        CREATE INDEX IF NOT EXISTS idx_task_claims_agent ON task_claims(agent);
        CREATE INDEX IF NOT EXISTS idx_task_dlq_task ON task_dlq(task_id);
        CREATE INDEX IF NOT EXISTS idx_task_messages_task ON task_messages(task_id);
    "

    # Add dedupe_key column if it doesn't exist (backward compatibility)
    if ! sqlite_queue "$DB_PATH" "SELECT 1 FROM pragma_table_info('task_queue') WHERE name='dedupe_key';" | grep -q 1; then
        sqlite_queue "$DB_PATH" "ALTER TABLE task_queue ADD COLUMN dedupe_key TEXT;"
    fi

    # Create unique index for active dedupe keys
    sqlite_queue "$DB_PATH" "
        CREATE UNIQUE INDEX IF NOT EXISTS idx_task_queue_active_dedupe
        ON task_queue(dedupe_key)
        WHERE dedupe_key IS NOT NULL AND status IN ('open','claimed');
    "

    echo "✅ Task queue tables initialized: $DB_PATH" >&2
}

case "$CMD" in
    init)
        mkdir -p "$(dirname "$DB_PATH")"
        init_db
        ;;

     add)
         SID="${1:-}"
         TOPIC="${2:-}"
         TASK="${3:-}"
         MAX_AGENTS="${4:-1}"
         DEDUPE_KEY="${5:-}"
          [ -z "$SID" ] || [ -z "$TOPIC" ] || [ -z "$TASK" ] && { echo "Usage: task-queue.sh add <sid> <topic> <task> [max] [dedupe_key]"; exit 1; }
         # Validate max_agents is a positive integer
         if [ -n "$MAX_AGENTS" ] && ! [[ "$MAX_AGENTS" =~ ^[1-9][0-9]*$ ]]; then
             echo "ERROR: max_agents must be a positive integer, got '$MAX_AGENTS'" >&2
             exit 1
         fi
          init_db 2>/dev/null || true

          # Precompute SQL-safe values
          SID_SQL="${SID//\'/\'\'}"
          TOPIC_SQL="${TOPIC//\'/\'\'}"
          TASK_SQL="${TASK//\'/\'\'}"
          DEDUPE_SQL="${DEDUPE_KEY//\'/\'\'}"

          # Handle dedupe_key for idempotency
          if [ -n "$DEDUPE_KEY" ]; then
              # Check if active task with this dedupe_key already exists
              EXISTING_ID=$(sqlite_queue "$DB_PATH" "
                  SELECT id FROM task_queue
                  WHERE dedupe_key = '$DEDUPE_SQL'
                  AND status IN ('open', 'claimed')
                  LIMIT 1;
              ")
              if [ -n "$EXISTING_ID" ]; then
                  echo "⚠️  Task already queued: id=$EXISTING_ID dedupe_key=$DEDUPE_KEY"
                  exit 0
              fi
          fi

          # Build SQL value for dedupe_key (NULL if empty)
          DEDUPE_VALUE="NULL"
          if [ -n "$DEDUPE_KEY" ]; then
              DEDUPE_VALUE="'$DEDUPE_SQL'"
          fi

           ROW_ID=$(sqlite_queue "$DB_PATH" "
               INSERT INTO task_queue(session_id, topic, task, max_agents, dedupe_key, created_at)
               VALUES ('$SID_SQL', '$TOPIC_SQL', '$TASK_SQL', $MAX_AGENTS, $DEDUPE_VALUE, '$(now)');
               SELECT last_insert_rowid();
           ")
         echo "📝 Task queued: id=$ROW_ID topic=$TOPIC max_agents=$MAX_AGENTS${DEDUPE_KEY:+ dedupe_key=$DEDUPE_KEY}"
        ;;

    claim)
        TOPIC="${1:-}"
        AGENT="${2:-}"
        [ -z "$TOPIC" ] || [ -z "$AGENT" ] && { echo "Usage: task-queue.sh claim <topic> <agent>"; exit 1; }
        init_db 2>/dev/null || true

        TOPIC_SQL="${TOPIC//\'/\'\'}"
        AGENT_SQL="${AGENT//\'/\'\'}"
        CLAIM_ERR=$(mktemp)
        TASK_ID=""
        for ATTEMPT in 1 2 3; do
            set +e
            # Atomic claim using CTE and RETURNING
            TASK_ID=$(sqlite_queue "$DB_PATH" "
                WITH available_task AS (
                    SELECT q.id
                    FROM task_queue q
                    LEFT JOIN task_claims c ON q.id = c.task_id
                    WHERE q.topic = '$TOPIC_SQL' AND q.status = 'open'
                      AND q.id NOT IN (SELECT task_id FROM task_claims WHERE agent = '$AGENT_SQL')
                    GROUP BY q.id
                    HAVING COUNT(c.agent) < q.max_agents
                    ORDER BY q.created_at ASC
                    LIMIT 1
                )
                INSERT INTO task_claims (task_id, agent, claimed_at)
                SELECT id, '$AGENT_SQL', '$(now)' FROM available_task
                RETURNING task_id;
            " 2>"$CLAIM_ERR")
            CLAIM_RC=$?
            set -e

            if [ "$CLAIM_RC" -eq 0 ]; then
                break
            fi

            CLAIM_MSG=$(<"$CLAIM_ERR")
            if [ "$CLAIM_RC" -eq 5 ] || [[ "$CLAIM_MSG" == *"database is locked"* ]]; then
                TASK_ID=$(sqlite_queue "$DB_PATH" "
                    SELECT q.id
                    FROM task_queue q
                    JOIN task_claims c ON c.task_id = q.id
                    WHERE q.topic = '$TOPIC_SQL'
                      AND q.status IN ('open', 'claimed')
                      AND c.agent = '$AGENT_SQL'
                    ORDER BY c.claimed_at DESC
                    LIMIT 1;
                ")
                [ -n "$TASK_ID" ] && break
                [ "$ATTEMPT" -lt 3 ] && sleep 1 && continue
            fi

            rm -f "$CLAIM_ERR"
            printf '%s\n' "$CLAIM_MSG" >&2
            exit "$CLAIM_RC"
        done
        rm -f "$CLAIM_ERR"

        if [ -z "$TASK_ID" ]; then
            echo "⏳ No open tasks for topic '$TOPIC'"
            exit 1
        fi

        TASK_PAYLOAD=$(sqlite_queue "$DB_PATH" "SELECT task FROM task_queue WHERE id = $TASK_ID;")
        echo "✅ Claimed task $TASK_ID for $AGENT: $TASK_PAYLOAD"
        ;;

    list)
        TOPIC="${1:-}"
        if [ -n "$TOPIC" ]; then
            sqlite_queue -header -column "$DB_PATH" "
                SELECT tq.id, tq.session_id, tq.topic, tq.status, tq.max_agents,
                       COUNT(tc.task_id) AS claims, tq.created_at
                FROM task_queue tq
                LEFT JOIN task_claims tc ON tc.task_id = tq.id
                WHERE tq.topic = '${TOPIC//\'/\'\'}'
                GROUP BY tq.id
                ORDER BY tq.id;
            "
        else
            sqlite_queue -header -column "$DB_PATH" "
                SELECT tq.id, tq.session_id, tq.topic, tq.status, tq.max_agents,
                       COUNT(tc.task_id) AS claims, tq.created_at
                FROM task_queue tq
                LEFT JOIN task_claims tc ON tc.task_id = tq.id
                GROUP BY tq.id
                ORDER BY tq.id;
            "
        fi
        ;;

     claims)
         TASK_ID="${1:-}"
         # Validate task_id is numeric if provided
         if [ -n "$TASK_ID" ] && ! [[ "$TASK_ID" =~ ^[0-9]+$ ]]; then
             echo "ERROR: task_id must be a numeric ID, got '$TASK_ID'" >&2
             exit 1
         fi
         if [ -n "$TASK_ID" ]; then
            sqlite_queue -header -column "$DB_PATH" "
                SELECT task_id, agent, claimed_at
                FROM task_claims
                WHERE task_id = $TASK_ID
                ORDER BY claimed_at;
            "
        else
            sqlite_queue -header -column "$DB_PATH" "
                SELECT task_id, agent, claimed_at
                FROM task_claims
                ORDER BY claimed_at;
            "
        fi
        ;;

     complete)
         TASK_ID="${1:-}"
         [ -z "$TASK_ID" ] && { echo "Usage: task-queue.sh complete <task_id>"; exit 1; }
         # Validate task_id is numeric
         if ! [[ "$TASK_ID" =~ ^[0-9]+$ ]]; then
             echo "ERROR: task_id must be a numeric ID, got '$TASK_ID'" >&2
             exit 1
         fi
        sqlite_queue "$DB_PATH" "
            UPDATE task_queue SET status = 'completed' WHERE id = $TASK_ID;
        "
        echo "✅ Task $TASK_ID marked completed"
        ;;

     fail)
         TASK_ID="${1:-}"
         AGENT="${2:-}"
         ERR="${3:-}"
         CTX="${4:-}"
         [ -z "$TASK_ID" ] || [ -z "$AGENT" ] && { echo "Usage: task-queue.sh fail <task_id> <agent> [error_message] [context_dump]"; exit 1; }
         # Validate task_id is numeric
         if ! [[ "$TASK_ID" =~ ^[0-9]+$ ]]; then
             echo "ERROR: task_id must be a numeric ID, got '$TASK_ID'" >&2
             exit 1
         fi
        sqlite_queue "$DB_PATH" "
            UPDATE task_queue SET status = 'failed' WHERE id = $TASK_ID;
            DELETE FROM task_claims WHERE task_id = $TASK_ID AND agent = '${AGENT//\'/\'\'}';
            INSERT INTO task_dlq(task_id, failed_agent, error_message, context_dump, failed_at)
            VALUES ($TASK_ID, '${AGENT//\'/\'\'}', '${ERR//\'/\'\'}', '${CTX//\'/\'\'}', '$(now)');
        "
        echo "❌ Task $TASK_ID marked failed by $AGENT"
        ;;

     sweep)
         TIMEOUT="${1:-}"
         [ -z "$TIMEOUT" ] && { echo "Usage: task-queue.sh sweep <timeout_seconds>"; exit 1; }
         # Validate timeout_seconds is a positive integer
         if ! [[ "$TIMEOUT" =~ ^[1-9][0-9]*$ ]]; then
             echo "ERROR: timeout_seconds must be a positive integer, got '$TIMEOUT'" >&2
             exit 1
         fi
        SWEPT=$(sqlite_queue "$DB_PATH" "
            SELECT COUNT(*) FROM task_claims
            WHERE claimed_at < strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '-$TIMEOUT seconds')
              AND task_id IN (SELECT id FROM task_queue WHERE status = 'open');
        ")
        if [ "$SWEPT" -gt 0 ]; then
            sqlite_queue "$DB_PATH" "
                DELETE FROM task_claims
                WHERE claimed_at < strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '-$TIMEOUT seconds')
                  AND task_id IN (SELECT id FROM task_queue WHERE status = 'open');
            "
            echo "🧹 Swept $SWEPT stale claim(s)"
        else
            echo "🧹 No stale claims found"
        fi
        ;;

    dlq)
        sqlite_queue -header -column "$DB_PATH" "
            SELECT d.id, d.task_id, q.task, d.failed_agent, d.error_message, d.failed_at
            FROM task_dlq d
            JOIN task_queue q ON d.task_id = q.id
            ORDER BY d.failed_at DESC;
        "
        ;;

    msg-send)
        TASK_ID="${1:-}"
        FROM_AGENT="${2:-}"
        BODY="${3:-}"
        TO_AGENT="${4:-}"
        [ -z "$TASK_ID" ] || [ -z "$FROM_AGENT" ] || [ -z "$BODY" ] && { echo "Usage: task-queue.sh msg-send <task_id> <from_agent> <body> [to_agent]"; exit 1; }

        TO_VAL="NULL"
        if [ -n "$TO_AGENT" ]; then
            TO_VAL="'${TO_AGENT//\'/\'\'}'"
        fi

        MSG_ID=$(sqlite_queue "$DB_PATH" "
            INSERT INTO task_messages(task_id, from_agent, to_agent, body, created_at)
            VALUES ($TASK_ID, '${FROM_AGENT//\'/\'\'}', $TO_VAL, '${BODY//\'/\'\'}', '$(now)');
            SELECT last_insert_rowid();
        ")
        echo "✉️ Message $MSG_ID sent on task $TASK_ID"
        ;;

    msg-poll)
        TASK_ID="${1:-}"
        LAST_SEEN_ID="${2:-0}"
        [ -z "$TASK_ID" ] && { echo "Usage: task-queue.sh msg-poll <task_id> [last_seen_id]"; exit 1; }

        sqlite_queue -header -column "$DB_PATH" "
            SELECT id, from_agent, to_agent, body, created_at
            FROM task_messages
            WHERE task_id = $TASK_ID AND id > $LAST_SEEN_ID
            ORDER BY id ASC;
        "
        ;;

    help|*)
        cat << 'HELP'
╔══════════════════════════════════════════════════════════╗
║              📋 TASK QUEUE — SQLite-backed               ║
╚══════════════════════════════════════════════════════════╝

Database: .specify/session.db (shared with session-db.sh)

COMMANDS:
  task-queue.sh init                          Initialize task queue tables
  task-queue.sh add <sid> <topic> <task> [max] [dedupe_key]  Enqueue a task
  task-queue.sh claim <topic> <agent>          Atomically claim an open task
  task-queue.sh list [topic]                   List tasks (optionally filter by topic)
  task-queue.sh claims [task_id]               List claims (optionally filter by task)
  task-queue.sh complete <task_id>             Mark task as completed
  task-queue.sh fail <task_id> <agent> [err] [ctx] Mark task as failed and send to DLQ
  task-queue.sh sweep <timeout_seconds>        Remove stale claims (zombie sweep)
  task-queue.sh dlq                            List dead-letter queue entries
  task-queue.sh msg-send <task_id> <from> <body> [to] Send a message on a task
  task-queue.sh msg-poll <task_id> [last_id]   Poll messages for a task

EXAMPLES:
  task-queue.sh add ses_12345 "auth" "Implement login middleware" 2
  task-queue.sh claim "auth" "wall-builder"
  task-queue.sh msg-send 42 "wall-builder" "I'm starting on the HTML"
  task-queue.sh msg-poll 42 0
  task-queue.sh complete 42
HELP
        ;;
esac
