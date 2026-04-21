-- +migrate Up
CREATE TABLE IF NOT EXISTS meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    schema_version INTEGER NOT NULL DEFAULT 1,
    cli_version TEXT NOT NULL,
    installed_at TEXT NOT NULL,
    last_updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    scope TEXT NOT NULL,
    tools TEXT NOT NULL DEFAULT '[]',
    cli_tools TEXT NOT NULL DEFAULT '[]',
    enable_servers TEXT NOT NULL DEFAULT '[]',
    project_name TEXT NOT NULL DEFAULT '',
    workspace_name TEXT NOT NULL DEFAULT '',
    target_dir TEXT NOT NULL DEFAULT '',
    planning_dir TEXT NOT NULL DEFAULT 'specs',
    planning_repo_path TEXT NOT NULL DEFAULT '',
    repos TEXT NOT NULL DEFAULT '[]',
    global_ref TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS selections (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    templates TEXT NOT NULL DEFAULT '[]',
    rules TEXT NOT NULL DEFAULT '[]',
    agents TEXT NOT NULL DEFAULT '[]',
    skills TEXT NOT NULL DEFAULT '[]',
    prompts TEXT NOT NULL DEFAULT '[]',
    infra TEXT NOT NULL DEFAULT '[]',
    constitution TEXT NOT NULL DEFAULT '[]',
    features TEXT NOT NULL DEFAULT '{}',
    git_conventions TEXT NOT NULL DEFAULT '{}',
    preset TEXT NOT NULL DEFAULT 'standard'
);

CREATE TABLE IF NOT EXISTS tracked_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL,
    source TEXT NOT NULL,
    owner TEXT NOT NULL DEFAULT 'library',
    status TEXT NOT NULL DEFAULT 'installed',
    installed_at TEXT NOT NULL,
    last_checked_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS operations (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    files_affected TEXT NOT NULL DEFAULT '[]',
    result TEXT NOT NULL,
    backup_paths TEXT NOT NULL DEFAULT '[]',
    error TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sync (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_sync_at TEXT NOT NULL,
    dirty INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS feature_flags (
    key TEXT PRIMARY KEY,
    value INTEGER NOT NULL DEFAULT 0
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tracked_files_path ON tracked_files(path);
CREATE INDEX IF NOT EXISTS idx_operations_timestamp ON operations(timestamp);

-- +migrate Down
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS sync;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS tracked_files;
DROP TABLE IF EXISTS selections;
DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS meta;