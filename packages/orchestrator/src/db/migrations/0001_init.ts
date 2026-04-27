import type { Migration } from '../types.js'

export const migration: Migration = {
  id: '0001_init',
  sql: '-- bootstrap: schema_migrations is created by the runner',
}
