/**
 * Scaffolds `.ai/housekeeping/sync-state.json` — an initial state file for
 * the QMD / CodeGraph drift-tracking integration.
 *
 * Mirrors Go's `internal/scaffold/housekeeping.go#ScaffoldHousekeeping`.
 * Runs only when a HousekeepingConfig is provided; no-op otherwise.
 *
 * Neither runtime currently *consumes* this file — a future spec will add
 * the drift-detection consumer that reads/updates it. For now, parity
 * means both runtimes emit the same v1 schema on init.
 */

import path from 'node:path'
import type { HousekeepingConfig } from '../types.js'
import { ensureDir, fileHash, writeFile } from '../utils/files.js'
import { marshalSortedJson } from '../utils/configmerge.js'

export interface ScaffoldHousekeepingOptions {
  targetDir: string
  config: HousekeepingConfig | null | undefined
  fileRecords: Array<{ path: string; hash: string; source: string; owner?: 'library' | 'user' | 'migrated' }>
}

export function scaffoldHousekeeping(opts: ScaffoldHousekeepingOptions): void {
  if (!opts.config) return

  const memoryPath = opts.config.memoryPath ?? path.join('specs', 'memory')
  ensureDir(path.join(opts.targetDir, memoryPath))

  const housekeepingDir = path.join(opts.targetDir, '.ai', 'housekeeping')
  ensureDir(housekeepingDir)

  const initialState = {
    schemaVersion: 1,
    updatedAt: '',
    qmd: {
      enabled: opts.config.enableQmd ?? false,
      indexPath: opts.config.qmdIndexPath ?? '',
      driftStatus: 'unknown',
    },
    codegraph: {
      enabled: opts.config.enableCodegraph ?? false,
      dataPath: opts.config.codegraphDataPath ?? '',
      driftStatus: 'unknown',
    },
    staleAcked: {
      qmd: [],
      codegraph: [],
    },
    repairProposals: [],
  }

  const outputPath = path.join(housekeepingDir, 'sync-state.json')
  writeFile(outputPath, marshalSortedJson(initialState))
  opts.fileRecords.push({
    path: '.ai/housekeeping/sync-state.json',
    hash: fileHash(outputPath),
    source: 'scaffold:housekeeping',
    owner: 'library',
  })
}
