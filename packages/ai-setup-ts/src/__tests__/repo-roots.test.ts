import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import {
  generateClaudeSettings,
  scaffoldRepoLedgers,
  scaffoldRepoRoots,
} from '../scaffold/repo-roots.js'
import { DEFAULT_REPO_PERMISSIONS } from '../types.js'

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

  it('generateClaudeSettings produces correct allow and deny lists', () => {
    const settings = generateClaudeSettings(
      DEFAULT_REPO_PERMISSIONS,
      {
        language: 'Ruby',
        framework: 'Rails',
        packageManager: 'bundle',
        commands: {
          test: 'bundle exec rspec',
          lint: 'bundle exec rubocop',
          build: 'bundle exec rails assets:precompile',
        },
      },
    )

    expect(settings).toEqual({
      permissions: {
        allow: [
          'Read',
          'Edit',
          'Bash(bundle exec rspec)',
          'Bash(bundle exec rubocop)',
          'Bash(bundle exec rails assets:precompile)',
        ],
        deny: [
          'Bash(rm -rf *)',
          'Bash(rails db:drop*)',
          'Bash(rails db:reset*)',
          'Bash(bundle publish*)',
          'Bash(git push*)',
          'Bash(git push --force*)',
        ],
      },
    })
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
        tools: ['opencode', 'claude-code'],
        strategy: 'skip',
        perFileOverrides: new Map(),
      }),
    ).resolves.toEqual(new Map([['missing-repo', []]]))

    expect(existsSync(path.join(workspaceRoot, 'missing-repo'))).toBe(false)
  })
})
