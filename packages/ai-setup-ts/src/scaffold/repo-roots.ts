import path from 'node:path'
import {
  type ConflictStrategy,
  DEFAULT_REPO_PERMISSIONS,
  type FileRecord,
  type RepoPermissions,
  type ToolId,
} from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileExists, fileHash, isDirectory, writeFile } from '../utils/files.js'
import { detectProjectStack, type ProjectStack } from '../utils/repo-detection.js'
import { ROOT_FILE_BY_TOOL } from './root-file-map.js'

export interface RepoRootOptions {
  repos: Array<{ name: string; path: string; type?: string; description?: string }>
  planningRepoPath: string
  tools: ToolId[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Generate lightweight root files in each referenced repo.
 * Each repo gets a root file per tool, pre-filled with detected stack info
 * and a pointer to the planning repo.
 */
export async function scaffoldRepoRoots(opts: RepoRootOptions): Promise<Map<string, FileRecord[]>> {
  const results = new Map<string, FileRecord[]>()

  for (const repo of opts.repos) {
    const repoAbsPath = path.resolve(opts.planningRepoPath, repo.path)
    const records: FileRecord[] = []

    if (!fileExists(repoAbsPath) || !isDirectory(repoAbsPath)) {
      results.set(repo.name, records)
      continue
    }

    const stack = detectProjectStack(repoAbsPath)
    const content = generateRepoRootContent({
      repoName: repo.name,
      stack,
      planningRepoPath: opts.planningRepoPath,
    })

    const writtenFiles = new Set<string>()
    for (const tool of opts.tools) {
      const outputFile = ROOT_FILE_BY_TOOL[tool]
      if (writtenFiles.has(outputFile)) continue
      writtenFiles.add(outputFile)

      const destPath = path.join(repoAbsPath, outputFile)
      if (outputFile.includes('/')) {
        ensureDir(path.dirname(destPath))
      }

      const action = applyStrategy(destPath, opts.strategy, opts.perFileOverrides, repoAbsPath)
      if (action === 'skip') continue

      writeFile(destPath, content)
      records.push({
        path: `${repo.name}/${outputFile}`,
        hash: fileHash(destPath),
        source: 'workspace:repo-root',
        owner: 'library',
      })
    }

    if (opts.tools.includes('claude-code')) {
      const settings = generateClaudeSettings(DEFAULT_REPO_PERMISSIONS, stack)
      const claudeDir = path.join(repoAbsPath, '.claude')
      ensureDir(claudeDir)

      const settingsPath = path.join(claudeDir, 'settings.json')
      const action = applyStrategy(settingsPath, opts.strategy, opts.perFileOverrides, repoAbsPath)
      if (action !== 'skip') {
        writeFile(settingsPath, JSON.stringify(settings, null, 2))
        records.push({
          path: `${repo.name}/.claude/settings.json`,
          hash: fileHash(settingsPath),
          source: 'workspace:permissions',
          owner: 'library',
        })
      }
    }

    results.set(repo.name, records)
  }

  return results
}

function generateRepoRootContent(opts: {
  repoName: string
  stack: ProjectStack
  planningRepoPath: string
}): string {
  const { repoName, stack, planningRepoPath } = opts

  const lines: string[] = [
    `# ${repoName}`,
    '',
    '## Project Stack',
    '',
  ]

  if (stack.language && stack.language !== 'Unknown') {
    lines.push(`- **Language**: ${stack.language}`)
  }
  if (stack.framework) {
    lines.push(`- **Framework**: ${stack.framework}`)
  }
  if (stack.testFramework) {
    lines.push(`- **Testing**: ${stack.testFramework}`)
  }
  if (stack.packageManager) {
    lines.push(`- **Package Manager**: ${stack.packageManager}`)
  }

  if (stack.description) {
    lines.push('', `> ${stack.description}`)
  }

  const cmds = stack.commands
  if (cmds.install || cmds.test || cmds.lint || cmds.build || cmds.dev) {
    lines.push('', '## Commands', '')
    lines.push('```bash')
    if (cmds.install) lines.push(`${cmds.install}     # Install dependencies`)
    if (cmds.test) lines.push(`${cmds.test}        # Run tests`)
    if (cmds.lint) lines.push(`${cmds.lint}        # Run linter`)
    if (cmds.build) lines.push(`${cmds.build}       # Build`)
    if (cmds.dev) lines.push(`${cmds.dev}         # Start dev server`)
    lines.push('```')
  }

  lines.push(
    '',
    '## Workspace',
    '',
    'This repo is part of a workspace. Plans, standards, and coordination live in the planning repo.',
    '',
    `- **Planning repo**: ${planningRepoPath}`,
    '- For feature plans, see: specs/features/ in the planning repo',
    '- For coding standards, see: specs/standards/ in the planning repo',
    '',
    '## Claude Code Permissions',
    '',
    '- Default Claude Code permissions: read, write, and safe project commands are allowed',
    '- Destructive commands and git push operations are denied by default',
    '- If this repo needs different access, customize `.claude/settings.json` manually',
    '',
    '## Before Making Changes',
    '',
    '1. Pull latest — other team members may have pushed',
    '2. Check the plan is still current — read the task file before implementing',
    '3. After completing work — update the ledger in the planning repo',
    '',
  )

  return lines.join('\n')
}

export function generateClaudeSettings(permissions: RepoPermissions, stack: ProjectStack): object {
  const allow: string[] = ['Read']

  if (permissions.write) allow.push('Edit')
  if (permissions.runCommands) {
    if (stack.commands.test) allow.push(`Bash(${stack.commands.test})`)
    if (stack.commands.lint) allow.push(`Bash(${stack.commands.lint})`)
    if (stack.commands.build) allow.push(`Bash(${stack.commands.build})`)
  }

  const deny: string[] = []
  if (!permissions.runDestructive) {
    deny.push('Bash(rm -rf *)')
    if (stack.language === 'Ruby') deny.push('Bash(rails db:drop*)', 'Bash(rails db:reset*)')
    if (stack.packageManager) deny.push(`Bash(${stack.packageManager} publish*)`)
  }
  if (!permissions.gitOperations) {
    deny.push('Bash(git push*)', 'Bash(git push --force*)')
  }

  return { permissions: { allow, deny } }
}

export async function scaffoldRepoLedgers(opts: {
  planningRepoPath: string
  repos: Array<{ name: string; path: string; type?: string; description?: string }>
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}): Promise<void> {
  // Detect memory path: .specify/memory/repos/ (speckit) or specs/memory/repos/ (legacy)
  const specifyMemoryRepos = path.join(opts.planningRepoPath, '.specify', 'memory', 'repos')
  const legacyMemoryRepos = path.join(opts.planningRepoPath, 'specs', 'memory', 'repos')
  const baseMemoryRepos = isDirectory(specifyMemoryRepos) ? specifyMemoryRepos : legacyMemoryRepos

  for (const repo of opts.repos) {
    const repoMemoryDir = path.join(baseMemoryRepos, repo.name)
    ensureDir(repoMemoryDir)

    const ledgerPath = path.join(repoMemoryDir, 'ledger.md')
    const ledgerContent = [
      `# ${repo.name} — Activity Ledger`,
      '',
      '| Date | Who | What | Plan ref | Verified |',
      '|------|-----|------|----------|----------|',
      '',
      '<!-- AI: append a new row after every task completed in this repo -->',
      '',
    ].join('\n')

    const action1 = applyStrategy(ledgerPath, opts.strategy, opts.perFileOverrides, opts.planningRepoPath)
    if (action1 !== 'skip') {
      writeFile(ledgerPath, ledgerContent)
      opts.fileRecords.push({
        path: path.relative(opts.planningRepoPath, ledgerPath).replaceAll('\\', '/'),
        hash: fileHash(ledgerPath),
        source: 'workspace:ledger',
        owner: 'library',
      })
    }

    const repoAbsPath = path.resolve(opts.planningRepoPath, repo.path)
    const stack = detectProjectStack(repoAbsPath)
    const statePath = path.join(repoMemoryDir, 'last-known-state.md')
    const stateLines = [
      `# ${repo.name} — Last Known State`,
      '',
      `- **Type**: ${repo.type ?? 'unknown'}`,
      `- **Language**: ${stack.language}`,
    ]
    if (stack.framework) stateLines.push(`- **Framework**: ${stack.framework}`)
    if (stack.testFramework) stateLines.push(`- **Test Framework**: ${stack.testFramework}`)
    if (stack.packageManager) stateLines.push(`- **Package Manager**: ${stack.packageManager}`)
    if (repo.description) stateLines.push(`- **Description**: ${repo.description}`)
    stateLines.push('', `*Generated: ${new Date().toISOString().split('T')[0]}*`, '')

    const action2 = applyStrategy(statePath, opts.strategy, opts.perFileOverrides, opts.planningRepoPath)
    if (action2 !== 'skip') {
      writeFile(statePath, stateLines.join('\n'))
      opts.fileRecords.push({
        path: path.relative(opts.planningRepoPath, statePath).replaceAll('\\', '/'),
        hash: fileHash(statePath),
        source: 'workspace:state',
        owner: 'library',
      })
    }
  }
}
