import type { Migration } from '../types.js'

export const migration: Migration = {
  id: '0003_catalog',
  sql: `
CREATE TABLE IF NOT EXISTS definitions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL CHECK (kind IN ('agent','skill','chain','team','workflow','mode','command')),
  name TEXT NOT NULL,
  active_version_id INTEGER,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(kind, name)
);

CREATE TABLE IF NOT EXISTS definition_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  definition_id INTEGER NOT NULL REFERENCES definitions(id),
  version INTEGER NOT NULL,
  frontmatter_json TEXT NOT NULL,
  body TEXT NOT NULL,
  checksum TEXT NOT NULL,
  created_at TEXT NOT NULL,
  created_by TEXT,
  UNIQUE(definition_id, version)
);

CREATE INDEX IF NOT EXISTS idx_def_versions_checksum ON definition_versions(checksum);
CREATE INDEX IF NOT EXISTS idx_def_versions_def ON definition_versions(definition_id, version);
  `,
}
