import type { Migration } from '../types.js'

export const migration: Migration = {
  id: '0002_run_state',
  sql: `
CREATE TABLE IF NOT EXISTS execution_plans (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('chain','team','workflow')),
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  project_root TEXT NOT NULL,
  plan_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chain_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  current_step_id TEXT,
  project_root TEXT NOT NULL,
  state_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS team_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  project_root TEXT NOT NULL,
  state_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_runs (
  id TEXT PRIMARY KEY,
  definition_name TEXT NOT NULL,
  definition_version TEXT,
  state TEXT NOT NULL,
  current_phase_id TEXT,
  project_root TEXT NOT NULL,
  state_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS handoffs (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  run_kind TEXT NOT NULL,
  doc_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS error_journal (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  run_kind TEXT,
  definition_name TEXT NOT NULL,
  step_id TEXT,
  category TEXT NOT NULL,
  code TEXT NOT NULL,
  message TEXT NOT NULL,
  entry_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chain_runs_state ON chain_runs(state);
CREATE INDEX IF NOT EXISTS idx_chain_runs_project ON chain_runs(project_root, updated_at);
CREATE INDEX IF NOT EXISTS idx_team_runs_project ON team_runs(project_root, updated_at);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_project ON workflow_runs(project_root, updated_at);
CREATE INDEX IF NOT EXISTS idx_error_journal_run ON error_journal(run_id);
CREATE INDEX IF NOT EXISTS idx_error_journal_project ON error_journal(definition_name);
  `,
}
