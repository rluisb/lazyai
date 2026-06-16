// Package runtime provides the Go-native runtime for LazyAI multi-agent execution.
package runtime

// SchemaV2 is the reduced runtime schema introduced by Spec 025 Phase 3.
const SchemaV2 = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    agent TEXT NOT NULL,
    model TEXT,
    goal TEXT,
    repo TEXT,
    worktree TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    token_total INTEGER NOT NULL DEFAULT 0,
    summary TEXT,
    tags TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);

CREATE TABLE IF NOT EXISTS dispatches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    seq INTEGER NOT NULL,
    parent_id INTEGER,
    agent TEXT NOT NULL,
    model TEXT,
    task TEXT,
    phase TEXT,
    workflow TEXT,
    mode TEXT,
    started_at TEXT,
    ended_at TEXT,
    result TEXT,
    token_used INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    summary TEXT,
    files_touched TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_dispatches_session ON dispatches(session_id);
CREATE INDEX IF NOT EXISTS idx_dispatches_parent ON dispatches(parent_id);

CREATE TABLE IF NOT EXISTS handoff (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    path TEXT NOT NULL,
    goal TEXT,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_handoff_session ON handoff(session_id);

CREATE TABLE IF NOT EXISTS agent_defaults (
    tool_id TEXT PRIMARY KEY,
    default_agent TEXT NOT NULL,
    instructions TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ledger_refs (
    id INTEGER PRIMARY KEY,
    session_id TEXT,
    event_type TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_ledger_refs_session ON ledger_refs(session_id);

INSERT OR IGNORE INTO schema_migrations (version, applied_at, name)
VALUES (2, CURRENT_TIMESTAMP, 'runtime_schema_v2');

INSERT INTO agent_defaults (tool_id, default_agent, instructions)
VALUES
    ('opencode', 'guide', '.opencode/AGENTS.md'),
    ('claude-code', 'guide', 'CLAUDE.md'),
    ('copilot', 'guide', '.github/copilot-instructions.md')
ON CONFLICT(tool_id) DO UPDATE SET
    default_agent = excluded.default_agent,
    instructions = excluded.instructions;
`

// SchemaCurrent is the schema applied to fresh and migrated runtime databases.
const SchemaCurrent = SchemaV2
