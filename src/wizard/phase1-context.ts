import * as p from '@clack/prompts'
import path from 'node:path'
import { homedir } from 'node:os'
import { readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { readManifest } from '../utils/manifest.js'
import { validateFilesystemSafeName } from '../utils/validation.js'
import { fileExists, isDirectory, readFile, resolveLibraryDir } from '../utils/files.js'
import { detectRepoInfo, scanWorkspaceRepos } from '../utils/repo-detection.js'
import type { SetupScope, SetupType, ToolId, WizardSelections } from '../types.js'

interface McpServerConfig {
  requiresInstall?: boolean
  installHint?: string
}

interface McpCliToolConfig {
  installHint?: string
}

interface McpCatalog {
  servers: Record<string, McpServerConfig>
  cliTools?: Record<string, McpCliToolConfig>
}

export interface Phase1Result {
  setupScope: SetupScope
  tools: ToolId[]
  projectName: string
  workspaceName?: string
  planningRepoPath?: string
  repos?: Array<{ name: string; path: string; type?: string }>
  cliTools?: string[]
}

function getCliToolOptions(targetDir: string): Array<{ value: string; label: string; hint: string }> {
  const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)))
  const catalogPath = path.join(libraryDir, 'mcp', 'catalog.json')
  if (!fileExists(catalogPath)) return []

  try {
    const catalog = JSON.parse(readFile(catalogPath)) as McpCatalog
    const options: Array<{ value: string; label: string; hint: string }> = []

    for (const [name, server] of Object.entries(catalog.servers)) {
      if (!server.requiresInstall) continue
      const label = name.charAt(0).toUpperCase() + name.slice(1)
      const hint = server.installHint ? `MCP server (${server.installHint})` : 'MCP server requiring local install'
      options.push({ value: name, label, hint })
    }

    for (const [name, tool] of Object.entries(catalog.cliTools ?? {})) {
      const label = name.toUpperCase()
      const hint = tool.installHint ? `CLI tool (${tool.installHint})` : 'CLI tool requiring local install'
      options.push({ value: name, label, hint })
    }

    return options
  } catch {
    return []
  }
}

function listSubdirectories(parentDir: string): string[] {
  try {
    return readdirSync(parentDir, { withFileTypes: true })
      .filter((entry) => entry.isDirectory() && !entry.name.startsWith('.'))
      .map((entry) => path.join(parentDir, entry.name))
      .sort()
  } catch {
    return []
  }
}

function parseWorkspaceRepos(raw: string | undefined, planningRepoPath: string): Array<{ name: string; path: string; type?: string }> {
  if (!raw) return []

  const planningRepoAbsolute = path.resolve(planningRepoPath)
  const trimmed = raw.trim()
  if (!trimmed) return []

  const toRepoRef = (repoPathInput: string): { name: string; path: string; type?: string } => {
    const resolved = path.resolve(planningRepoAbsolute, repoPathInput)
    const detected = detectRepoInfo(resolved, planningRepoAbsolute)
    return {
      name: detected.name,
      path: detected.path || '.',
      type: detected.type,
    }
  }

  if (trimmed.includes(',')) {
    return trimmed
      .split(',')
      .map(entry => entry.trim())
      .filter(Boolean)
      .map(toRepoRef)
  }

  const parentDirCandidate = path.resolve(planningRepoAbsolute, trimmed)
  if (!fileExists(parentDirCandidate) || !isDirectory(parentDirCandidate)) {
    return [toRepoRef(trimmed)]
  }

  try {
    const scannedRepos = scanWorkspaceRepos(parentDirCandidate, planningRepoAbsolute)
    const parentEntries = readdirSync(parentDirCandidate, { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      .map((entry) => path.join(parentDirCandidate, entry.name))
      .filter((entryPath) => path.resolve(entryPath) !== planningRepoAbsolute)
      .map((entryPath) => detectRepoInfo(entryPath, planningRepoAbsolute))
      .filter((repo) => repo.path !== '.')
    const scannedRepoPaths = new Set(scannedRepos.map((repo) => repo.path))

    const nonGitRepos = parentEntries.filter((repo) => !scannedRepoPaths.has(repo.path))
    if (nonGitRepos.length > 0) {
      p.note(`Skipped non-git directories: ${nonGitRepos.map((repo) => repo.name).join(', ')}`)
    }

    return scannedRepos
      .map((repo) => ({
        name: repo.name,
        path: repo.path,
        type: repo.type,
      }))
      .filter((repo) => repo.path !== '.')
  } catch {
    return [toRepoRef(trimmed)]
  }
}

/**
 * Run Phase 1 of the interactive wizard: gather setupScope, tools, and projectName.
 * Behavior depends on interactive mode and CLI overrides.
 */
export async function runPhase1(opts: {
  interactive: boolean
  prior: Partial<WizardSelections> & {
    setupScope?: SetupScope
    setupType?: SetupType
    tools?: ToolId[]
    projectName?: string
    workspaceName?: string
    planningRepoPath?: string
  }
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    cliTools?: string[]
    name?: string
    planningRepo?: string
    repos?: string[]
  }
  targetDir: string
}): Promise<Phase1Result> {
  // Non-interactive mode: use cliOverrides or throw
  if (!opts.interactive) {
    const setupScope = opts.cliOverrides.scope ?? opts.cliOverrides.type
    const tools = opts.cliOverrides.tools
    const projectName =
      setupScope === 'workspace'
        ? path.basename(path.resolve(opts.cliOverrides.planningRepo ?? opts.targetDir))
        : opts.cliOverrides.name ?? (setupScope === 'global' ? 'global' : undefined)
    const workspaceName = setupScope === 'workspace' ? opts.cliOverrides.name : undefined
    const planningRepoPath = opts.cliOverrides.planningRepo
      ? path.resolve(opts.cliOverrides.planningRepo)
      : undefined
    const parsedRepos =
      setupScope === 'workspace'
        ? parseWorkspaceRepos((opts.cliOverrides.repos ?? []).join(','), planningRepoPath ?? opts.targetDir)
        : []
    const cliTools = opts.cliOverrides.cliTools

    if (!setupScope) {
      throw new Error('--scope is required in non-interactive mode (global | workspace | project)')
    }
    if (!tools || tools.length === 0) {
      throw new Error('--tools is required in non-interactive mode (pi, opencode, claude-code, gemini, copilot)')
    }
    if (!projectName) {
      throw new Error('Project name is required in non-interactive mode (use --name or provide via config)')
    }
    if (setupScope === 'workspace' && !planningRepoPath) {
      throw new Error('--planning-repo is required in non-interactive mode when --scope=workspace')
    }

    return {
      setupScope,
      tools,
      projectName,
      ...(workspaceName ? { workspaceName } : {}),
      ...(setupScope === 'workspace' && planningRepoPath ? { planningRepoPath } : {}),
      ...(setupScope === 'workspace' && parsedRepos.length > 0 ? { repos: parsedRepos } : {}),
      ...(cliTools && cliTools.length > 0 ? { cliTools } : {}),
    }
  }

  // Interactive mode
  p.intro('🤖  ai-setup — AI development environment scaffold')

  // Check for existing manifest and show note if found
  const existingManifest = await readManifest(opts.targetDir)
  if (existingManifest) {
    p.note('Re-running setup — previous selections will be pre-filled')
  }

  let setupScope: SetupScope
  let tools: ToolId[]
  let projectName: string
  let workspaceName: string | undefined
  let planningRepoPath: string | undefined
  let repos: Array<{ name: string; path: string; type?: string }> = []
  let cliTools: string[] = opts.cliOverrides.cliTools ?? []

  // Prompt 1: Setup scope
  const setupScopeResult =
    opts.cliOverrides.scope ||
    opts.cliOverrides.type ||
    opts.prior.setupScope ||
    opts.prior.setupType ||
    (await p.select({
      message: 'Setup scope:',
      options: [
        { value: 'global', label: 'Global', hint: 'Install to ~/.ai/ + native tool global paths' },
        { value: 'workspace', label: 'Workspace', hint: 'Planning repo with multi-project management' },
        { value: 'project', label: 'Project', hint: 'Self-contained single repository' },
      ],
    }))

  if (p.isCancel(setupScopeResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  setupScope = setupScopeResult as SetupScope

  if (setupScope === 'project' || setupScope === 'workspace') {
    const globalAiExists = fileExists(path.join(homedir(), '.ai'))
    if (globalAiExists) {
      p.note('Global AI setup detected at ~/.ai/. Project-level artifacts will layer on top of global ones.')
    }
  }

  // Prompt 2: Tools selection
  const toolsResult =
    opts.cliOverrides.tools ||
    opts.prior.tools ||
    (await p.multiselect({
      message: 'Which AI tools are you using?',
      options: [
        { value: 'pi', label: 'Pi (Claude Code)', hint: 'Uses .pi/ directory + CLAUDE.md' },
        { value: 'opencode', label: 'OpenCode', hint: 'Uses .opencode/ directory + AGENTS.md' },
        { value: 'claude-code', label: 'Claude Code', hint: 'Uses .claude/ directory + CLAUDE.md' },
        { value: 'gemini', label: 'Gemini CLI', hint: 'Uses .gemini/ directory + GEMINI.md' },
        { value: 'copilot', label: 'GitHub Copilot', hint: 'Uses .github/ + copilot-instructions.md' },
      ],
      required: true,
    }))

  if (p.isCancel(toolsResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  tools = toolsResult as ToolId[]

  if (!opts.cliOverrides.cliTools) {
    const cliToolOptions = getCliToolOptions(opts.targetDir)
    if (cliToolOptions.length > 0) {
      const cliToolsResult = await p.multiselect({
        message: 'Which CLI tools do you have installed? (optional, press Enter to skip)',
        options: cliToolOptions,
        required: false,
      })

      if (p.isCancel(cliToolsResult)) {
        p.cancel('Setup cancelled.')
        process.exit(0)
      }

      cliTools = (cliToolsResult as string[]) || []
    }
  }

  if (setupScope === 'workspace') {
    const defaultPlanningRepoPath = opts.cliOverrides.planningRepo || opts.prior.planningRepoPath || opts.targetDir

    const workspaceNameResult =
      opts.cliOverrides.name ||
      opts.prior.workspaceName ||
      (await p.text({
        message: 'Workspace name?',
        placeholder: path.basename(opts.targetDir),
        defaultValue: path.basename(opts.targetDir),
        validate: (value) => validateFilesystemSafeName(value, 'Workspace name'),
      }))

    if (p.isCancel(workspaceNameResult)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    workspaceName = workspaceNameResult

    let planningRepoPathResolved: string

    if (opts.cliOverrides.planningRepo || opts.prior.planningRepoPath) {
      planningRepoPathResolved = path.resolve(opts.cliOverrides.planningRepo || opts.prior.planningRepoPath || opts.targetDir)
    } else {
      const subdirs = listSubdirectories(opts.targetDir)
      const dirOptions: Array<{ value: string; label: string; hint?: string }> = [
        { value: opts.targetDir, label: path.basename(opts.targetDir) + ' (current directory)', hint: opts.targetDir },
      ]

      for (const dir of subdirs) {
        const name = path.basename(dir)
        const hasGit = fileExists(path.join(dir, '.git'))
        dirOptions.push({
          value: dir,
          label: name,
          hint: hasGit ? 'git repo' : 'directory',
        })
      }

      dirOptions.push({ value: '__manual__', label: 'Enter path manually…' })

      const planningRepoPick = await p.select({
        message: 'Which directory is the planning repo?',
        options: dirOptions,
      })

      if (p.isCancel(planningRepoPick)) {
        p.cancel('Setup cancelled.')
        process.exit(0)
      }

      if (planningRepoPick === '__manual__') {
        const manualPath = await p.text({
          message: 'Planning repo path:',
          placeholder: defaultPlanningRepoPath,
          defaultValue: defaultPlanningRepoPath,
        })

        if (p.isCancel(manualPath)) {
          p.cancel('Setup cancelled.')
          process.exit(0)
        }
        planningRepoPathResolved = path.resolve(manualPath)
      } else {
        planningRepoPathResolved = path.resolve(planningRepoPick as string)
      }
    }

    planningRepoPath = planningRepoPathResolved
    projectName = path.basename(planningRepoPath)

    // Referenced repos: scan sibling directories and show as multiselect
    if (opts.cliOverrides.repos) {
      repos = parseWorkspaceRepos(opts.cliOverrides.repos.join(','), planningRepoPath)
    } else {
      const parentDir = path.dirname(planningRepoPath)
      const siblingDirs = listSubdirectories(parentDir)
        .filter((dir) => path.resolve(dir) !== path.resolve(planningRepoPathResolved))

      if (siblingDirs.length > 0) {
        const repoOptions: Array<{ value: string; label: string; hint?: string }> = []

        for (const dir of siblingDirs) {
          const info = detectRepoInfo(dir, planningRepoPathResolved)
          if (!info.isGitRepo) continue

          const typeLabel = info.type !== 'unknown' ? info.type : 'unknown stack'
          repoOptions.push({
            value: dir,
            label: info.name,
            hint: typeLabel,
          })
        }

        if (repoOptions.length > 0) {
          const reposResult = await p.multiselect({
            message: 'Which repos should this workspace reference?',
            options: repoOptions,
            required: false,
          })

          if (p.isCancel(reposResult)) {
            p.cancel('Setup cancelled.')
            process.exit(0)
          }

          repos = (reposResult as string[])
            .map((dir) => detectRepoInfo(dir, planningRepoPath!))
            .map((info) => ({ name: info.name, path: info.path, type: info.type }))
        } else {
          p.note('No git repos found in sibling directories. You can add repos later in .ai/config.yml.')
        }
      } else {
        p.note('No sibling directories found. You can add repos later in .ai/config.yml.')
      }
    }
  } else {
    // Prompt 3: Project name
    const defaultName = setupScope === 'global' ? 'global' : path.basename(opts.targetDir)
    const projectNameResult =
      opts.cliOverrides.name ||
      opts.prior.projectName ||
      (await p.text({
        message: 'Project name?',
        placeholder: defaultName,
        defaultValue: defaultName,
        validate: (value) => validateFilesystemSafeName(value, 'Project name'),
      }))

    if (p.isCancel(projectNameResult)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    projectName = projectNameResult
  }

  return {
    setupScope,
    tools: tools as ToolId[],
    projectName,
    ...(workspaceName ? { workspaceName } : {}),
    ...(planningRepoPath ? { planningRepoPath } : {}),
    ...(repos.length > 0 ? { repos } : {}),
    ...(cliTools.length > 0 ? { cliTools } : {}),
  }
}
