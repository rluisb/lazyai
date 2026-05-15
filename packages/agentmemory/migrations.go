package agentmemory

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a versioned schema change.
type Migration struct {
	ID  int
	SQL string
}

const migration001CreateTables = `
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	namespace TEXT NOT NULL,
	project_root TEXT NOT NULL,
	task_type TEXT NOT NULL,
	state TEXT NOT NULL,
	current_step TEXT,
	state_json TEXT,
	goal TEXT,
	tags TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS task_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	run_id TEXT,
	event_type TEXT NOT NULL,
	payload_json TEXT,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS checkpoints (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	step_id TEXT,
	summary TEXT,
	state_json TEXT,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS artifacts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT NOT NULL,
	namespace TEXT NOT NULL,
	path TEXT NOT NULL,
	content_preview TEXT,
	size_bytes INTEGER,
	content_hash TEXT,
	mime_type TEXT,
	tags TEXT,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS memories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	namespace TEXT NOT NULL,
	content TEXT NOT NULL,
	source_task_id TEXT,
	source_step_id TEXT,
	tags TEXT,
	importance TEXT NOT NULL DEFAULT 'normal',
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tasks_namespace_updated ON tasks(namespace, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_events_task_id ON task_events(task_id, id);
CREATE INDEX IF NOT EXISTS idx_task_events_namespace ON task_events(namespace, id DESC);
CREATE INDEX IF NOT EXISTS idx_checkpoints_task_id ON checkpoints(task_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_artifacts_task_namespace ON artifacts(task_id, namespace);
CREATE INDEX IF NOT EXISTS idx_memories_namespace ON memories(namespace, id DESC);
`

const migration002FTS = `
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
	content,
	namespace,
	source_task_id,
	tags,
	content='memories',
	content_rowid='rowid'
);
CREATE VIRTUAL TABLE IF NOT EXISTS artifacts_fts USING fts5(
	content_preview,
	path,
	namespace,
	tags,
	content='artifacts',
	content_rowid='rowid'
);
CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
	INSERT INTO memories_fts(rowid, content, namespace, source_task_id, tags)
	VALUES (new.rowid, new.content, new.namespace, new.source_task_id, new.tags);
END;
CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
	INSERT INTO memories_fts(memories_fts, rowid, content, namespace, source_task_id, tags)
	VALUES('delete', old.rowid, old.content, old.namespace, old.source_task_id, old.tags);
END;
CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
	INSERT INTO memories_fts(memories_fts, rowid, content, namespace, source_task_id, tags)
	VALUES('delete', old.rowid, old.content, old.namespace, old.source_task_id, old.tags);
	INSERT INTO memories_fts(rowid, content, namespace, source_task_id, tags)
	VALUES (new.rowid, new.content, new.namespace, new.source_task_id, new.tags);
END;
CREATE TRIGGER IF NOT EXISTS artifacts_ai AFTER INSERT ON artifacts BEGIN
	INSERT INTO artifacts_fts(rowid, content_preview, path, namespace, tags)
	VALUES (new.rowid, new.content_preview, new.path, new.namespace, new.tags);
END;
CREATE TRIGGER IF NOT EXISTS artifacts_ad AFTER DELETE ON artifacts BEGIN
	INSERT INTO artifacts_fts(artifacts_fts, rowid, content_preview, path, namespace, tags)
	VALUES('delete', old.rowid, old.content_preview, old.path, old.namespace, old.tags);
END;
CREATE TRIGGER IF NOT EXISTS artifacts_au AFTER UPDATE ON artifacts BEGIN
	INSERT INTO artifacts_fts(artifacts_fts, rowid, content_preview, path, namespace, tags)
	VALUES('delete', old.rowid, old.content_preview, old.path, old.namespace, old.tags);
	INSERT INTO artifacts_fts(rowid, content_preview, path, namespace, tags)
	VALUES (new.rowid, new.content_preview, new.path, new.namespace, new.tags);
END;
`

var migrations = []Migration{
	{ID: 1, SQL: migration001CreateTables},
	{ID: 2, SQL: migration002FTS},
}

// RunMigrations applies pending migrations in ID order. Each migration is transactional.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (id INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })
	for _, migration := range migrations {
		var present int
		err := db.QueryRow("SELECT 1 FROM _migrations WHERE id = ?", migration.ID).Scan(&present)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %d: %w", migration.ID, err)
		}
		if err := runMigration(db, migration); err != nil {
			return err
		}
	}
	return nil
}

func runMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", migration.ID, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("migration %d: %w", migration.ID, err)
	}
	if _, err := tx.Exec("INSERT INTO _migrations (id, applied_at) VALUES (?, ?)", migration.ID, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("record migration %d: %w", migration.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", migration.ID, err)
	}
	return nil
}
