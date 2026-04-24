import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { scaffoldHousekeeping } from '../scaffold/housekeeping.js'

describe('scaffoldHousekeeping', () => {
  const tempDirs: string[] = []

  const makeTempDir = (): string => {
    const dir = mkdtempSync(path.join(tmpdir(), 'ai-setup-housekeeping-'))
    tempDirs.push(dir)
    return dir
  }

  afterEach(() => {
    for (const dir of tempDirs) {
      rmSync(dir, { recursive: true, force: true })
    }
    tempDirs.length = 0
  })

  it('is a no-op when config is null', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string }> = []

    scaffoldHousekeeping({ targetDir: dir, config: null, fileRecords })

    expect(existsSync(path.join(dir, '.ai', 'housekeeping'))).toBe(false)
    expect(fileRecords).toHaveLength(0)
  })

  it('is a no-op when config is undefined', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string }> = []

    scaffoldHousekeeping({ targetDir: dir, config: undefined, fileRecords })

    expect(existsSync(path.join(dir, '.ai', 'housekeeping'))).toBe(false)
  })

  it('writes sync-state.json with v1 schema when config is provided', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string; owner?: 'library' | 'user' | 'migrated' }> = []

    scaffoldHousekeeping({
      targetDir: dir,
      config: { enableQmd: true, qmdIndexPath: '~/qmd.idx', enableCodegraph: false },
      fileRecords,
    })

    const statePath = path.join(dir, '.ai', 'housekeeping', 'sync-state.json')
    expect(existsSync(statePath)).toBe(true)

    const state = JSON.parse(readFileSync(statePath, 'utf-8'))
    expect(state.schemaVersion).toBe(1)
    expect(state.qmd).toEqual({ enabled: true, indexPath: '~/qmd.idx', driftStatus: 'unknown' })
    expect(state.codegraph).toEqual({ enabled: false, dataPath: '', driftStatus: 'unknown' })
    expect(state.staleAcked).toEqual({ qmd: [], codegraph: [] })
    expect(state.repairProposals).toEqual([])

    const record = fileRecords.find((r) => r.path === '.ai/housekeeping/sync-state.json')
    expect(record).toBeDefined()
    expect(record?.source).toBe('scaffold:housekeeping')
    expect(record?.owner).toBe('library')
  })

  it('creates the memory directory using the configured path', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string }> = []

    scaffoldHousekeeping({
      targetDir: dir,
      config: { memoryPath: 'my-memory-dir', enableQmd: true },
      fileRecords,
    })

    expect(existsSync(path.join(dir, 'my-memory-dir'))).toBe(true)
  })

  it('defaults memoryPath to specs/memory when not provided', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string }> = []

    scaffoldHousekeeping({
      targetDir: dir,
      config: { enableCodegraph: true },
      fileRecords,
    })

    expect(existsSync(path.join(dir, 'specs', 'memory'))).toBe(true)
  })

  it('produces deterministic sorted-key output (byte-identical on re-run)', () => {
    const dir = makeTempDir()
    const fileRecords: Array<{ path: string; hash: string; source: string }> = []
    const config = { enableQmd: true, qmdIndexPath: '/idx', enableCodegraph: true, codegraphDataPath: '/cg' }

    scaffoldHousekeeping({ targetDir: dir, config, fileRecords })
    const firstContent = readFileSync(path.join(dir, '.ai', 'housekeeping', 'sync-state.json'), 'utf-8')

    scaffoldHousekeeping({ targetDir: dir, config, fileRecords: [] })
    const secondContent = readFileSync(path.join(dir, '.ai', 'housekeeping', 'sync-state.json'), 'utf-8')

    expect(firstContent).toBe(secondContent)
  })
})
