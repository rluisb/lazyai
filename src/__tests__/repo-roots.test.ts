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

  it('scaffoldRepoRoots generates root files in referenced repos', async () => {
    const workspaceRoot = makeTempDir('ai-setup-repo-roots-')
    tempDirs.push(workspaceRoot)

    const planningRepoPath = path.join(workspaceRoot, 'planning')
    const appRepoPath = path.join(workspaceRoot, 'app')

    mkdirSync(planningRepoPath, { recursive: true })
    mkdirSync(appRepoPath, { recursive: true })
    writeFileSync(
      path.join(appRepoPath, 'package.json'),
      JSON.stringify({
        description: 'Workspace app',
        dependencies: { react: '^18.0.0' },
        devDependencies: { vitest: '^2.0.0' },
        scripts: {
          test: 'vitest run',
          lint: 'biome check .',
          build: 'vite build',
          dev: 'vite dev',
        },
      }),
    )
    writeFileSync(path.join(appRepoPath, 'pnpm-lock.yaml'), 'lockfileVersion: 9.0')

    const results = await scaffoldRepoRoots({
      repos: [{ name: 'app', path: '../app', type: 'react-typescript' }],
      planningRepoPath,
      tools: ['opencode', 'claude-code', 'copilot'],
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(appRepoPath, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(appRepoPath, 'CLAUDE.md'))).toBe(true)
    expect(existsSync(path.join(appRepoPath, '.github', 'copilot-instructions.md'))).toBe(true)
    expect(existsSync(path.join(appRepoPath, '.claude', 'settings.json'))).toBe(true)

    expect(results.get('app')).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ path: 'app/AGENTS.md', source: 'workspace:repo-root' }),
        expect.objectContaining({ path: 'app/CLAUDE.md', source: 'workspace:repo-root' }),
        expect.objectContaining({ path: 'app/.github/copilot-instructions.md', source: 'workspace:repo-root' }),
        expect.objectContaining({ path: 'app/.claude/settings.json', source: 'workspace:permissions' }),
      ]),
    )
  })

  it('root file content includes detected stack info', async () => {
    const workspaceRoot = makeTempDir('ai-setup-repo-root-content-')
    tempDirs.push(workspaceRoot)

    const planningRepoPath = path.join(workspaceRoot, 'planning')
    const webRepoPath = path.join(workspaceRoot, 'web')

    mkdirSync(planningRepoPath, { recursive: true })
    mkdirSync(webRepoPath, { recursive: true })
    writeFileSync(
      path.join(webRepoPath, 'package.json'),
      JSON.stringify({
        description: 'Customer-facing app',
        dependencies: { next: '14.0.0' },
        devDependencies: { vitest: '^2.0.0' },
        scripts: {
          test: 'vitest run',
          lint: 'biome check .',
          build: 'next build',
          dev: 'next dev',
        },
      }),
    )
    writeFileSync(path.join(webRepoPath, 'package-lock.json'), '{}')

    await scaffoldRepoRoots({
      repos: [{ name: 'web', path: '../web', type: 'nextjs-typescript' }],
      planningRepoPath,
      tools: ['opencode', 'claude-code'],
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const content = readFileSync(path.join(webRepoPath, 'AGENTS.md'), 'utf-8')

    expect(content).toContain('# web')
    expect(content).toContain('- **Language**: TypeScript')
    expect(content).toContain('- **Framework**: Next.js')
    expect(content).toContain('- **Testing**: Vitest')
    expect(content).toContain('- **Package Manager**: npm')
    expect(content).toContain('npm test        # Run tests')
  })

  it('root file content includes planning repo pointer and permission guidance', async () => {
    const workspaceRoot = makeTempDir('ai-setup-repo-root-pointer-')
    tempDirs.push(workspaceRoot)

    const planningRepoPath = path.join(workspaceRoot, 'planning')
    const webRepoPath = path.join(workspaceRoot, 'web')

    mkdirSync(planningRepoPath, { recursive: true })
    mkdirSync(webRepoPath, { recursive: true })
    writeFileSync(
      path.join(webRepoPath, 'package.json'),
      JSON.stringify({
        dependencies: { next: '14.0.0' },
        scripts: { test: 'vitest run' },
      }),
    )

    await scaffoldRepoRoots({
      repos: [{ name: 'web', path: '../web', type: 'nextjs-typescript' }],
      planningRepoPath,
      tools: ['opencode', 'claude-code'],
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const content = readFileSync(path.join(webRepoPath, 'AGENTS.md'), 'utf-8')

    expect(content).toContain(`- **Planning repo**: ${planningRepoPath}`)
    expect(content).toContain('Default Claude Code permissions')
    expect(content).toContain('customize `.claude/settings.json` manually')
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
