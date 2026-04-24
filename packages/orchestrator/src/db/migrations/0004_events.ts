import type { Migration } from '../types.js'

export const migration: Migration = {
  id: '0004_events',
  sql: `
    CREATE TABLE IF NOT EXISTS run_events (
      id           INTEGER PRIMARY KEY AUTOINCREMENT,
      event_type   TEXT    NOT NULL,
      run_id       TEXT    NOT NULL,
      run_kind     TEXT    NOT NULL CHECK (run_kind IN ('chain', 'team', 'workflow')),
      payload_json TEXT    NOT NULL DEFAULT '{}',
      emitted_at   TEXT    NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_run_events_run ON run_events (run_id);
    CREATE INDEX IF NOT EXISTS idx_run_events_kind ON run_events (run_kind, emitted_at);
  `,
}
