import type { Migration } from '../types.js'

export const migration: Migration = {
  id: '0005_queue',
  sql: `
    CREATE TABLE IF NOT EXISTS queue_jobs (
      id           TEXT    PRIMARY KEY,
      job_type     TEXT    NOT NULL,
      payload_json TEXT    NOT NULL DEFAULT '{}',
      status       TEXT    NOT NULL DEFAULT 'pending'
                           CHECK (status IN ('pending','claimed','completed','failed')),
      priority     INTEGER NOT NULL DEFAULT 0,
      attempts     INTEGER NOT NULL DEFAULT 0,
      max_attempts INTEGER NOT NULL DEFAULT 3,
      error_json   TEXT,
      created_at   TEXT    NOT NULL,
      claimed_at   TEXT,
      completed_at TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_queue_jobs_dequeue
      ON queue_jobs (priority DESC, created_at ASC)
      WHERE status = 'pending';

    CREATE INDEX IF NOT EXISTS idx_queue_jobs_status ON queue_jobs (status);
  `,
}
