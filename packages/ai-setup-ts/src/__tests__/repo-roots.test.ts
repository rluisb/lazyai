import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import {
  scaffoldRepoLedgers,
  scaffoldRepoRoots,
} from '../scaffold/repo-roots.js'

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('workspace repo roots', () => {
  const tempDirs: string[] = []

  afterEach(() => {
    for (const dir of tempDirs) {
      rmSync(dir, { recursive: true, force: true })
    }
    tempDirs.length = 0
  })

  it('scaffoldRepoLedgers creates ledger and state files in the planning repo', async () => {
    const workspaceRoot = makeTempDir('ai-setup-repo-ledgers-')
    tempDirs.push(workspaceRoot)

    const planningRepoPath = path.join(workspaceRoot, 'planning')
    const apiRepoPath = path.join(workspaceRoot, 'api')

    mkdirSync(planningRepoPath, { recursive: true })
    mkdirSync(apiRepoPath, { recursive: true })
    writeFileSync(path.join(apiRepoPath, 'Cargo.toml'), '[package]\nname = "api"\ndescription = "API service"')
    writeFileSync(path.join(apiRepoPath, 'Cargo.lock'), '')

    const fileRecords: Array<{ path: string; hash: string; source: string; owner?: 'library' | 'user' | 'migrated' }> = []

    await scaffoldRepoLedgers({
      planningRepoPath,
      repos: [{ name: 'api', path: '../api', type: 'rust', description: 'Backend API' }],
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const ledgerPath = path.join(planningRepoPath, 'specs', 'memory', 'repos', 'api', 'ledger.md')
    const statePath = path.join(planningRepoPath, 'specs', 'memory', 'repos', 'api', 'last-known-state.md')

    expect(existsSync(ledgerPath)).toBe(true)
    expect(existsSync(statePath)).toBe(true)
    expect(readFileSync(ledgerPath, 'utf-8')).toContain('# api — Activity Ledger')
    expect(readFileSync(statePath, 'utf-8')).toContain('- **Language**: Rust')
    expect(readFileSync(statePath, 'utf-8')).toContain('- **Package Manager**: cargo')
    expect(readFileSync(statePath, 'utf-8')).toContain('- **Description**: Backend API')
    expect(fileRecords).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          path: 'specs/memory/repos/api/ledger.md',
          source: 'workspace:ledger',
        }),
        expect.objectContaining({
          path: 'specs/memory/repos/api/last-known-state.md',
          source: 'workspace:state',
        }),
      ]),
    )
  })

  it('handles missing repo paths gracefully', async () => {
    const workspaceRoot = makeTempDir('ai-setup-missing-repo-')
    tempDirs.push(workspaceRoot)

    const planningRepoPath = path.join(workspaceRoot, 'planning')
    mkdirSync(planningRepoPath, { recursive: true })

    await expect(
      scaffoldRepoRoots({
        repos: [{ name: 'missing-repo', path: '../missing-repo', type: 'unknown' }],
        planningRepoPath,
        tools: ['opencode'],
        strategy: 'skip',
        perFileOverrides: new Map(),
      }),
    ).resolves.toEqual(new Map([['missing-repo', []]]))

    expect(existsSync(path.join(workspaceRoot, 'missing-repo'))).toBe(false)
  })
})
