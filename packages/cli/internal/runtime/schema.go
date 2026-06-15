// Package runtime provides the Go-native runtime for LazyAI multi-agent execution.
package runtime

// SchemaV1 is the initial schema definition for the Go runtime.
// This replaces the bash V2 schema with a unified, versioned Go implementation.
const SchemaV1 = `
-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL,
    name TEXT NOT NULL
);

-- Core Session Tables

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    agent TEXT NOT NULL DEFAULT 'primary-agent',
    model TEXT,
    goal TEXT,
    repo TEXT,
    worktree TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    token_total INTEGER DEFAULT 0,
    summary TEXT,
    tags TEXT
);

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
    token_used INTEGER DEFAULT 0,
    error_message TEXT,
    summary TEXT,
    files_touched TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_dispatches_session ON dispatches(session_id);
CREATE INDEX IF NOT EXISTS idx_dispatches_parent ON dispatches(parent_id);

CREATE TABLE IF NOT EXISTS decisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    dispatch_id INTEGER,
    title TEXT,
    context TEXT,
    decision TEXT,
    rationale TEXT,
    alternatives TEXT,
    created_at TEXT NOT NULL,
    tags TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_decisions_session ON decisions(session_id);

CREATE TABLE IF NOT EXISTS artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    dispatch_id INTEGER,
    path TEXT NOT NULL,
    action TEXT NOT NULL DEFAULT 'modified',
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_artifacts_session ON artifacts(session_id);

CREATE TABLE IF NOT EXISTS memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    title TEXT,
    content TEXT,
    tags TEXT,
    importance TEXT NOT NULL DEFAULT 'normal',
    verified INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id);

CREATE TABLE IF NOT EXISTS token_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    dispatch_id INTEGER,
    token_count INTEGER NOT NULL,
    context_pct REAL,
    recorded_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_token_log_session ON token_log(session_id);

CREATE TABLE IF NOT EXISTS parallel_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    parent_dispatch_id INTEGER,
    wave_id TEXT NOT NULL DEFAULT 'wave_1',
    agent TEXT NOT NULL,
    task TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    started_at TEXT,
    completed_at TEXT,
    result TEXT,
    output_path TEXT,
    error_message TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_parallel_tasks_session ON parallel_tasks(session_id);
CREATE INDEX IF NOT EXISTS idx_parallel_tasks_wave ON parallel_tasks(wave_id);
CREATE INDEX IF NOT EXISTS idx_parallel_tasks_status ON parallel_tasks(status);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    from_agent TEXT NOT NULL,
    to_agent TEXT,
    subject TEXT,
    body TEXT NOT NULL,
    priority TEXT NOT NULL DEFAULT 'normal',
    status TEXT NOT NULL DEFAULT 'unread',
    created_at TEXT NOT NULL,
    read_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_to_agent ON messages(to_agent, status);
CREATE INDEX IF NOT EXISTS idx_messages_from_agent ON messages(from_agent);

CREATE TABLE IF NOT EXISTS barriers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    barrier_id TEXT NOT NULL,
    expected_count INTEGER NOT NULL,
    arrived_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'waiting',
    created_at TEXT NOT NULL,
    resolved_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_barriers_session ON barriers(session_id);
CREATE INDEX IF NOT EXISTS idx_barriers_status ON barriers(status);

CREATE TABLE IF NOT EXISTS locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    lock_name TEXT NOT NULL,
    held_by TEXT,
    acquired_at TEXT,
    released_at TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_locks_session ON locks(session_id);
CREATE INDEX IF NOT EXISTS idx_locks_name ON locks(lock_name, status);

-- Workflow Tables

CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    agents TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);

CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    trigger_cmd TEXT,
    config_json TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT,
    team TEXT,
    steps TEXT
);

CREATE INDEX IF NOT EXISTS idx_workflows_name ON workflows(name);

CREATE TABLE IF NOT EXISTS workflow_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_name TEXT NOT NULL,
    session_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    current_step INTEGER NOT NULL DEFAULT 0,
    total_steps INTEGER NOT NULL DEFAULT 0,
    result TEXT,
    on_failure TEXT NOT NULL DEFAULT 'stop',
    started_at TEXT,
    completed_at TEXT,
    error_message TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_workflow_instances_session ON workflow_instances(session_id);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_status ON workflow_instances(status);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id INTEGER NOT NULL,
    step_order INTEGER NOT NULL,
    agent TEXT,
    task TEXT,
    mode TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    started_at TEXT,
    completed_at TEXT,
    result TEXT,
    output_path TEXT,
    error_message TEXT,
    FOREIGN KEY (instance_id) REFERENCES workflow_instances(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workflow_steps_instance ON workflow_steps(instance_id);

-- V2 Observability Tables

CREATE TABLE IF NOT EXISTS model_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    dispatch_id INTEGER,
    provider TEXT,
    model TEXT,
    tokens_input INTEGER,
    tokens_output INTEGER,
    cost REAL,
    latency_ms INTEGER,
    called_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_model_calls_session ON model_calls(session_id);

CREATE TABLE IF NOT EXISTS tool_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    dispatch_id INTEGER,
    tool_name TEXT,
    command TEXT,
    exit_code INTEGER,
    output TEXT,
    called_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (dispatch_id) REFERENCES dispatches(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_tool_calls_session ON tool_calls(session_id);

CREATE TABLE IF NOT EXISTS gate_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    workflow_instance_id INTEGER,
    phase_name TEXT,
    gate_type TEXT,
    status TEXT,
    human_required INTEGER,
    human_decision TEXT,
    evaluated_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_gate_results_session ON gate_results(session_id);

CREATE TABLE IF NOT EXISTS ledger_refs (
    seq INTEGER PRIMARY KEY,
    session_id TEXT,
    workflow_run_id TEXT,
    event_type TEXT NOT NULL,
    event_hash TEXT NOT NULL,
    prev_hash TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ledger_refs_session ON ledger_refs(session_id);

CREATE TABLE IF NOT EXISTS cost_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    workflow_instance_id INTEGER,
    total_cost REAL,
    token_count INTEGER,
    recorded_at TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_cost_snapshots_session ON cost_snapshots(session_id);

CREATE TABLE IF NOT EXISTS checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    checkpoint_name TEXT,
    state_json TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_session ON checkpoints(session_id);

CREATE TABLE IF NOT EXISTS eval_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    suite_name TEXT,
    dataset_path TEXT,
    started_at TEXT,
    completed_at TEXT,
    status TEXT DEFAULT 'pending',
    score REAL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_eval_runs_session ON eval_runs(session_id);

CREATE TABLE IF NOT EXISTS eval_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    eval_run_id INTEGER NOT NULL,
    assertion_name TEXT,
    passed INTEGER,
    expected TEXT,
    actual TEXT,
    latency_ms INTEGER,
    recorded_at TEXT,
    FOREIGN KEY (eval_run_id) REFERENCES eval_runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_eval_results_run ON eval_results(eval_run_id);

-- Task Queue Tables

CREATE TABLE IF NOT EXISTS task_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    max_agents INTEGER NOT NULL DEFAULT 1,
    dedupe_key TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_queue_session ON task_queue(session_id);
CREATE INDEX IF NOT EXISTS idx_task_queue_topic ON task_queue(topic);
CREATE INDEX IF NOT EXISTS idx_task_queue_status ON task_queue(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_task_queue_active_dedupe ON task_queue(dedupe_key) 
    WHERE dedupe_key IS NOT NULL AND status IN ('open', 'claimed');

CREATE TABLE IF NOT EXISTS task_claims (
    task_id INTEGER NOT NULL,
    agent TEXT NOT NULL,
    claimed_at TEXT NOT NULL,
    PRIMARY KEY (task_id, agent),
    FOREIGN KEY (task_id) REFERENCES task_queue(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_claims_task ON task_claims(task_id);
CREATE INDEX IF NOT EXISTS idx_task_claims_agent ON task_claims(agent);

CREATE TABLE IF NOT EXISTS task_dlq (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER,
    failed_agent TEXT,
    error_message TEXT,
    context_dump TEXT,
    failed_at TEXT NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task_queue(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_task_dlq_task ON task_dlq(task_id);

CREATE TABLE IF NOT EXISTS task_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    from_agent TEXT NOT NULL,
    to_agent TEXT,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task_queue(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_messages_task ON task_messages(task_id);
`

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
    ('opencode', 'primary-agent', '.opencode/AGENTS.md'),
    ('claude-code', 'primary-agent', 'CLAUDE.md'),
    ('copilot', 'primary-agent', '.github/copilot-instructions.md')
ON CONFLICT(tool_id) DO UPDATE SET
    default_agent = excluded.default_agent,
    instructions = excluded.instructions;
`

// SchemaCurrent is the schema applied to fresh and migrated runtime databases.
const SchemaCurrent = SchemaV2
