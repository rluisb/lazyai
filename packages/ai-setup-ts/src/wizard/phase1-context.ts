import { execSync } from 'node:child_process'
import { readdirSync } from 'node:fs'
import { homedir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import * as p from '@clack/prompts'
import type { AgentId, SetupScope, SetupType, SkillId, ToolId, WizardSelections } from '../types.js'
import { ALL_AGENTS, ALL_SKILLS } from '../types.js'
import { fileExists, isDirectory, readFile, resolveLibraryDir } from '../utils/files.js'
import { readManifest } from '../utils/manifest.js'
import { detectRepoInfo, scanWorkspaceRepos } from '../utils/repo-detection.js'
import { GO_BACK, isGoBack } from '../utils/ui.js'
import { validateFilesystemSafeName } from '../utils/validation.js'

interface McpServerConfig {
  enabled?: boolean
  requiresInstall?: boolean
  installHint?: string
}

interface McpCliToolConfig {
  installHint?: string
  enabled?: boolean
  description?: string
}

interface CliToolOption {
  value: string
  label: string
  hint: string
  isInstalled: boolean
}

/**
 * Check if a CLI tool is installed by running `which` command
 */
function checkToolInstalled(toolName: string): boolean {
  try {
    execSync(`which ${toolName}`, { stdio: 'pipe' })
    return true
  } catch {
    return false
  }
}

interface McpCatalog {
  servers: Record<string, McpServerConfig>
  cliTools?: Record<string, McpCliToolConfig>
}

export type McpWizardPreset = 'minimal' | 'recommended' | 'full'

const MCP_PRESET_OPTIONS: Array<{ value: McpWizardPreset; label: string; hint: string }> = [
  { value: 'minimal', label: 'Minimal', hint: 'Core local setup tools only.' },
  { value: 'recommended', label: 'Recommended', hint: 'Balanced default from enabled catalog servers.' },
  { value: 'full', label: 'Full', hint: 'All catalog server IDs.' },
]

const TOOL_OPTIONS: Array<{ value: ToolId; label: string; hint: string }> = [
  { value: 'opencode', label: 'OpenCode', hint: 'Uses opencode.json + .opencode/ directory + AGENTS.md' },
  { value: 'claude-code', label: 'Claude Code', hint: 'Uses .claude/ with rules, skills, agents + CLAUDE.md' },
  { value: 'gemini', label: 'Gemini CLI', hint: 'Uses .gemini/ with settings.json + GEMINI.md' },
  { value: 'copilot', label: 'GitHub Copilot', hint: 'Uses .github/ + root AGENTS.md' },
  { value: 'codex', label: 'Codex (OpenAI)', hint: 'Uses .agents/skills/ + AGENTS.md' },
  { value: 'pi', label: 'Pi', hint: 'Uses .pi/ with settings.json, skills, and prompts' },
]

function loadMcpCatalog(_targetDir: string): McpCatalog | null {
  const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)))
  const catalogPath = path.join(libraryDir, 'mcp', 'catalog.json')
  if (!fileExists(catalogPath)) return null

  try {
    return JSON.parse(readFile(catalogPath)) as McpCatalog
  } catch {
    return null
  }
}

export function normalizeMcpPreset(preset: string | undefined): McpWizardPreset {
  switch (preset) {
    case 'minimal':
    case 'full':
    case 'recommended':
      return preset
    default:
      return 'recommended'
  }
}

function sortedCatalogServerIDs(catalog: McpCatalog): string[] {
  return Object.keys(catalog.servers).sort((a, b) => a.localeCompare(b))
}

export function defaultMcpServersForPreset(preset: McpWizardPreset, targetDir: string): string[] {
  const catalog = loadMcpCatalog(targetDir)
  if (!catalog) return []

  const serverIds = sortedCatalogServerIDs(catalog)
  switch (normalizeMcpPreset(preset)) {
    case 'minimal':
      return serverIds.filter((serverId) => serverId === 'filesystem' || serverId === 'ripgrep')
    case 'full':
      return serverIds
    default:
      return serverIds.filter((serverId) => catalog.servers[serverId]?.enabled)
  }
}

function defaultMcpSelection(current: string[] | undefined, preset: McpWizardPreset, targetDir: string): string[] {
  if (current && current.length > 0) {
    return [...current].sort((a, b) => a.localeCompare(b))
  }

  return defaultMcpServersForPreset(preset, targetDir)
}

export function filterToolsByScope(tools: ToolId[], scope: SetupScope): ToolId[] {
  return tools.filter((tool) => tool !== 'pi' || scope === 'project' || scope === 'workspace')
}

export function toolOptionsForScope(scope: SetupScope): Array<{ value: ToolId; label: string; hint: string }> {
  return filterToolsByScope(TOOL_OPTIONS.map((option) => option.value), scope)
    .map((toolId) => TOOL_OPTIONS.find((option) => option.value === toolId))
    .filter((option): option is { value: ToolId; label: string; hint: string } => option != null)
}

export function detectInstalledCliToolsFromCatalog(targetDir: string): string[] {
  const catalog = loadMcpCatalog(targetDir)
  if (!catalog?.cliTools) return []

  return Object.keys(catalog.cliTools)
    .filter((toolName) => checkToolInstalled(toolName))
    .sort((a, b) => a.localeCompare(b))
}

export interface Phase1Result {
  setupScope: SetupScope
  tools: ToolId[]
  skills: SkillId[]
  agents: AgentId[]
  mcpPreset: McpWizardPreset
  projectName: string
  workspaceName?: string
  workspaceRoot?: string
  planningRepoPath?: string
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
  cliTools?: string[]
  /** MCP server names explicitly enabled by user (e.g., ['atlassian']) */
  enableServers?: string[]
  organization?: string
  team?: string
}

function getCliToolOptions(_targetDir: string): CliToolOption[] {
  const catalog = loadMcpCatalog(_targetDir)
  if (!catalog) return []

  return Object.entries(catalog.cliTools ?? {})
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([name, tool]) => {
      const isInstalled = checkToolInstalled(name)
      return {
        value: name,
        label: name.toUpperCase(),
        hint: isInstalled
          ? '✓ Already installed'
          : tool.installHint
            ? `Not installed (${tool.installHint})`
            : tool.description ?? 'CLI tool requiring local install',
        isInstalled,
      }
    })
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

function parseWorkspaceRepos(raw: string | undefined, planningRepoPath: string): Array<{ name: string; path: string; type?: string; description?: string }> {
  if (!raw) return []

  const planningRepoAbsolute = path.resolve(planningRepoPath)
  const trimmed = raw.trim()
  if (!trimmed) return []

  const toRepoRef = (repoPathInput: string): { name: string; path: string; type?: string; description?: string } => {
    const resolved = path.resolve(planningRepoAbsolute, repoPathInput)
    const detected = detectRepoInfo(resolved, planningRepoAbsolute)
    return {
      name: detected.name,
      path: detected.path || '.',
      type: detected.type,
      ...(detected.description ? { description: detected.description } : {}),
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
        ...(repo.description ? { description: repo.description } : {}),
      }))
      .filter((repo) => repo.path !== '.')
  } catch {
    return [toRepoRef(trimmed)]
  }
}

/**
 * Run Phase 1 of the interactive wizard: gather setupScope, tools, and projectName.
 * Behavior depends on interactive mode and CLI overrides.
 *
 * When re-running, prior values are pre-filled as defaults so the user can
 * just press Enter to keep them or change what they need.
 *
 * Returns GO_BACK sentinel if user selects "Back" (only for phases 2+).
 */
export async function runPhase1(opts: {
  interactive: boolean
    prior: Partial<WizardSelections> & {
      setupScope?: SetupScope
      setupType?: SetupType
      tools?: ToolId[]
      skills?: SkillId[]
      agents?: AgentId[]
      mcpPreset?: McpWizardPreset
      projectName?: string
      workspaceName?: string
      workspaceRoot?: string
      planningRepoPath?: string
      enableServers?: string[]
      organization?: string
      team?: string
    }
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    cliTools?: string[]
    name?: string
    planningRepo?: string
    workspaceRoot?: string
    repos?: string[]
    enableServers?: string[]
  }
  targetDir: string
  canGoBack?: boolean
}): Promise<Phase1Result | typeof GO_BACK> {
  // Non-interactive mode: use cliOverrides or throw
  if (!opts.interactive) {
    const setupScope = opts.cliOverrides.scope ?? opts.cliOverrides.type
    const tools = opts.cliOverrides.tools
    const mcpPreset = normalizeMcpPreset(opts.prior.mcpPreset)
    const skills = opts.prior.skills ?? [...ALL_SKILLS]
    const agents = opts.prior.agents ?? [...ALL_AGENTS]
    const projectName =
      setupScope === 'workspace'
        ? path.basename(path.resolve(opts.cliOverrides.planningRepo ?? opts.targetDir))
        : opts.cliOverrides.name ?? (setupScope === 'global' ? 'global' : 'my-project')
    const workspaceName = setupScope === 'workspace' ? opts.cliOverrides.name : undefined
    const planningRepoPath = opts.cliOverrides.planningRepo
      ? path.resolve(opts.cliOverrides.planningRepo)
      : undefined
    const workspaceRoot =
      setupScope === 'workspace'
        ? path.resolve(opts.cliOverrides.workspaceRoot ?? path.dirname(planningRepoPath ?? opts.targetDir))
        : undefined
    const parsedRepos =
      setupScope === 'workspace'
        ? parseWorkspaceRepos((opts.cliOverrides.repos ?? []).join(','), planningRepoPath ?? opts.targetDir)
        : []
    const cliTools = opts.cliOverrides.cliTools ?? detectInstalledCliToolsFromCatalog(opts.targetDir)
    const enableServers = defaultMcpSelection(opts.cliOverrides.enableServers ?? opts.prior.enableServers, mcpPreset, opts.targetDir)

    if (!setupScope) {
      throw new Error('--scope is required in non-interactive mode (global | workspace | project)')
    }
    if (!tools || tools.length === 0) {
      throw new Error('--tools is required in non-interactive mode (opencode, claude-code, gemini, copilot, codex, pi)')
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
      skills,
      agents,
      mcpPreset,
      projectName: setupScope === 'global' ? 'global' : projectName,
      ...(workspaceName ? { workspaceName } : {}),
      ...(setupScope === 'workspace' && workspaceRoot ? { workspaceRoot } : {}),
      ...(setupScope === 'workspace' && planningRepoPath ? { planningRepoPath } : {}),
      ...(setupScope === 'workspace' && parsedRepos.length > 0 ? { repos: parsedRepos } : {}),
      ...(cliTools && cliTools.length > 0 ? { cliTools } : {}),
      ...(enableServers && enableServers.length > 0 ? { enableServers } : {}),
    }
  }

  // Interactive mode
  p.intro('🤖  ai-setup — AI development environment scaffold')

  // Check for existing manifest and show note if found
  const existingManifest = await readManifest(opts.targetDir)
  if (existingManifest) {
    p.note('Re-running setup — previous selections will be pre-filled as defaults')
  }

  let setupScope: SetupScope
  // biome-ignore lint/style/useConst: assigned after async prompt below
  let tools: ToolId[]
  let projectName: string
  let workspaceName: string | undefined
  let workspaceRoot: string | undefined
  let planningRepoPath: string | undefined
  let repos: Array<{ name: string; path: string; type?: string; description?: string }> = []
  let cliTools: string[] = opts.cliOverrides.cliTools ?? []
  let organization = opts.prior.organization ?? ''
  let team = opts.prior.team ?? ''

  // Resolve default values from CLI overrides or prior selections
  const priorScope = opts.cliOverrides.scope ?? opts.cliOverrides.type ?? opts.prior.setupScope ?? opts.prior.setupType ?? 'project'
  const priorTools = filterToolsByScope(opts.cliOverrides.tools ?? opts.prior.tools ?? TOOL_OPTIONS.map((option) => option.value), priorScope)
  const priorSkills = opts.prior.skills ?? [...ALL_SKILLS]
  const priorAgents = opts.prior.agents ?? [...ALL_AGENTS]
  const priorMcpPreset = normalizeMcpPreset(opts.prior.mcpPreset)
  const priorEnableServers = opts.prior.enableServers && opts.prior.enableServers.length > 0
    ? [...opts.prior.enableServers].sort((a, b) => a.localeCompare(b))
    : undefined
  const priorProjectName = opts.cliOverrides.name ?? opts.prior.projectName

  // --- Prompt 1: Setup scope ---
  // Always show the prompt (even on re-run), but pre-fill with prior value
  const scopeOptions: Array<{ value: string; label: string; hint: string }> = [
    { value: 'global', label: 'Global', hint: 'Install to ~/.ai/ + native tool global paths' },
    { value: 'workspace', label: 'Workspace', hint: 'Planning repo with multi-project management' },
    { value: 'project', label: 'Project', hint: 'Self-contained single repository' },
  ]

  if (opts.canGoBack) {
    scopeOptions.push({ value: GO_BACK, label: '↩ Back', hint: 'Go back to previous step' })
  }

  const priorLabel = priorScope === 'global' ? 'Global' : priorScope === 'workspace' ? 'Workspace' : 'Project'
  const setupScopeResult = await p.select({
    message: `Setup scope: (previous: ${priorLabel})`,
    options: scopeOptions,
    initialValue: priorScope,
  })

  if (p.isCancel(setupScopeResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  if (isGoBack(setupScopeResult)) {
    return GO_BACK
  }

  setupScope = setupScopeResult as SetupScope

  if (setupScope === 'project' || setupScope === 'workspace') {
    const globalAiExists = fileExists(path.join(homedir(), '.ai'))
    if (globalAiExists) {
      p.note('Global AI setup detected at ~/.ai/. Project-level artifacts will layer on top of global ones.')
    }
  }

  // --- Prompt 2: Tools selection ---
  // Always show the prompt, pre-fill with prior tools if available
  const toolOptions = toolOptionsForScope(setupScope)

  const scopedPriorTools = filterToolsByScope(priorTools, setupScope)
  const toolsMessage = scopedPriorTools.length > 0
    ? `Which AI tools are you using? (previous: ${priorTools.join(', ')})`
    : 'Which AI tools are you using?'

  const toolsResult = await p.multiselect({
    message: toolsMessage,
    options: toolOptions,
    initialValues: scopedPriorTools,
    required: true,
  })

  if (p.isCancel(toolsResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  tools = toolsResult as ToolId[]

  const skillsResult = (await p.multiselect({
    message: `Which skills should be installed? (previous: ${priorSkills.join(', ')})`,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- @clack infers narrow types from options
    options: ALL_SKILLS.map((skill) => ({ value: skill as string, label: skill, hint: 'Skill' })) as any,
    initialValues: priorSkills,
    required: true,
  })) as SkillId[]

  if (p.isCancel(skillsResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  const skills = skillsResult as SkillId[]

  const agentsResult = await p.multiselect({
    message: `Which agents should be installed? (previous: ${priorAgents.join(', ')})`,
    options: ALL_AGENTS.map((agent) => ({ value: agent, label: agent, hint: 'Agent' })),
    initialValues: priorAgents,
    required: true,
  })

  if (p.isCancel(agentsResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  const agents = agentsResult as AgentId[]

  const mcpPresetResult = await p.select({
    message: `Which MCP preset should be enabled? (previous: ${priorMcpPreset})`,
    options: MCP_PRESET_OPTIONS,
    initialValue: priorMcpPreset,
  })

  if (p.isCancel(mcpPresetResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  const mcpPreset = normalizeMcpPreset(mcpPresetResult as string)

  const initialMcpServers = defaultMcpSelection(priorEnableServers, mcpPreset, opts.targetDir)
  const serverOptions = [
    ...defaultMcpServersForPreset('full', opts.targetDir).map((serverId) => ({ value: serverId, label: serverId, hint: 'MCP server' })),
  ]

  const mcpServersResult = await p.multiselect({
    message: 'Which MCP servers would you like to enable?',
    options: serverOptions,
    initialValues: initialMcpServers,
    required: false,
  })

  if (p.isCancel(mcpServersResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const enableServers = ((mcpServersResult as string[]) || []).sort((a, b) => a.localeCompare(b))

  if (setupScope === 'workspace') {
    const defaultPlanningRepoPath = opts.cliOverrides.planningRepo || opts.prior.planningRepoPath || opts.targetDir
    const priorWorkspaceName = opts.cliOverrides.name ?? opts.prior.workspaceName

    // Always show workspace name prompt, pre-fill with prior if available
    const workspaceDefault = priorWorkspaceName ?? path.basename(opts.targetDir)
    const workspaceMessage = priorWorkspaceName
      ? `Workspace name? (previous: ${priorWorkspaceName})`
      : 'Workspace name?'
    const workspaceNameResult = await p.text({
      message: workspaceMessage,
      placeholder: workspaceDefault,
      defaultValue: workspaceDefault,
      validate: (value) => validateFilesystemSafeName(value, 'Workspace name'),
    })

    if (p.isCancel(workspaceNameResult)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    workspaceName = workspaceNameResult

    // Workspace root: where AI tool configs live (defaults to parent of planning repo)
    const priorWorkspaceRoot = opts.cliOverrides.workspaceRoot ?? opts.prior.workspaceRoot
    const defaultWorkspaceRoot = priorWorkspaceRoot ?? path.resolve(opts.targetDir, '..')
    if (priorWorkspaceRoot) {
      p.note(`Using workspace root from previous run: ${priorWorkspaceRoot}`)
      workspaceRoot = priorWorkspaceRoot
    } else {
      const workspaceRootResult = await p.text({
        message: 'Workspace root directory? (AI tool configs .claude/ .opencode/ etc. live here)',
        placeholder: defaultWorkspaceRoot,
        defaultValue: defaultWorkspaceRoot,
        validate: (value) => {
          if (!value?.trim()) return 'Workspace root is required'
          const resolved = path.resolve(value)
          if (!fileExists(resolved)) return `Directory not found: ${resolved}`
          return undefined
        },
      })

      if (p.isCancel(workspaceRootResult)) {
        p.cancel('Setup cancelled.')
        process.exit(0)
      }
      workspaceRoot = path.resolve(workspaceRootResult)
    }

    let planningRepoPathResolved: string

    // Planning repo path: always show prompt but pre-fill with prior if available
    const priorPlanningRepo = opts.cliOverrides.planningRepo ?? opts.prior.planningRepoPath
    if (priorPlanningRepo) {
      planningRepoPathResolved = path.resolve(priorPlanningRepo)
      // Show a note about reusing the prior path
      p.note(`Using planning repo from previous run: ${planningRepoPathResolved}`)
    } else {
      const subdirs = listSubdirectories(opts.targetDir)
      const dirOptions: Array<{ value: string; label: string; hint?: string }> = [
        { value: opts.targetDir, label: `${path.basename(opts.targetDir)} (current directory)`, hint: opts.targetDir },
      ]

      for (const dir of subdirs) {
        const name = path.basename(dir)
        const hasGit = fileExists(path.join(dir, '.git'))
        const info = detectRepoInfo(dir, opts.targetDir)
        const parts: string[] = []
        if (hasGit) parts.push('git repo')
        if (info.type !== 'unknown') parts.push(info.type)
        if (info.description) parts.push(info.description)
        dirOptions.push({
          value: dir,
          label: name,
          hint: parts.join(' · ') || 'directory',
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

      const repoOptions: Array<{ value: string; label: string; hint?: string }> = []

      for (const dir of siblingDirs) {
        const info = detectRepoInfo(dir, planningRepoPathResolved)
        if (!info.isGitRepo) continue

        const parts: string[] = []
        if (info.type !== 'unknown') parts.push(info.type)
        if (info.description) parts.push(info.description)
        repoOptions.push({
          value: dir,
          label: info.name,
          hint: parts.join(' · ') || 'git repo',
        })
      }

      if (repoOptions.length > 0) {
        repoOptions.push({ value: '__custom__', label: 'Add custom path…', hint: 'enter a repo path not listed above' })

        const reposResult = await p.multiselect({
          message: 'Which repos should this workspace reference?',
          options: repoOptions,
          required: false,
        })

        if (p.isCancel(reposResult)) {
          p.cancel('Setup cancelled.')
          process.exit(0)
        }

        const selectedDirs = (reposResult as string[]).filter((v) => v !== '__custom__')
        const wantsCustom = (reposResult as string[]).includes('__custom__')

        repos = selectedDirs
          // biome-ignore lint/style/noNonNullAssertion: planningRepoPath is always set in workspace scope before this point
          .map((dir) => detectRepoInfo(dir, planningRepoPath!))
          .map((info) => ({
            name: info.name,
            path: info.path,
            type: info.type,
            ...(info.description ? { description: info.description } : {}),
          }))

        if (wantsCustom) {
          let addMore = true
          while (addMore) {
            const customPath = await p.text({
              message: 'Custom repo path (absolute or relative to planning repo):',
              placeholder: '../my-other-repo',
            })

            if (p.isCancel(customPath)) {
              p.cancel('Setup cancelled.')
              process.exit(0)
            }

            if (customPath) {
              // biome-ignore lint/style/noNonNullAssertion: planningRepoPath is always set in workspace scope before this point
              const resolved = path.resolve(planningRepoPath!, customPath)
              // biome-ignore lint/style/noNonNullAssertion: planningRepoPath is always set in workspace scope before this point
              const info = detectRepoInfo(resolved, planningRepoPath!)
              repos.push({
                name: info.name,
                path: info.path,
                type: info.type,
                ...(info.description ? { description: info.description } : {}),
              })
              p.log.success(`Added: ${info.name} (${info.type}${info.description ? ` · ${info.description}` : ''})`)
            }

            const another = await p.confirm({
              message: 'Add another custom repo?',
              initialValue: false,
            })

            if (p.isCancel(another)) {
              p.cancel('Setup cancelled.')
              process.exit(0)
            }

            addMore = another
          }
        }
      } else {
        const addManual = await p.confirm({
          message: 'No git repos found nearby. Add a repo path manually?',
          initialValue: false,
        })

        if (p.isCancel(addManual)) {
          p.cancel('Setup cancelled.')
          process.exit(0)
        }

        if (addManual) {
          let addMore = true
          while (addMore) {
            const customPath = await p.text({
              message: 'Repo path (absolute or relative to planning repo):',
              placeholder: '../my-repo',
            })

            if (p.isCancel(customPath)) {
              p.cancel('Setup cancelled.')
              process.exit(0)
            }

            if (customPath) {
              // biome-ignore lint/style/noNonNullAssertion: planningRepoPath is always set in workspace scope before this point
              const resolved = path.resolve(planningRepoPath!, customPath)
              // biome-ignore lint/style/noNonNullAssertion: planningRepoPath is always set in workspace scope before this point
              const info = detectRepoInfo(resolved, planningRepoPath!)
              repos.push({
                name: info.name,
                path: info.path,
                type: info.type,
                ...(info.description ? { description: info.description } : {}),
              })
              p.log.success(`Added: ${info.name} (${info.type}${info.description ? ` · ${info.description}` : ''})`)
            }

            const another = await p.confirm({
              message: 'Add another repo?',
              initialValue: false,
            })

            if (p.isCancel(another)) {
              p.cancel('Setup cancelled.')
              process.exit(0)
            }

            addMore = another
          }
        } else {
          p.note('You can add repos later in .ai/config.yml.')
        }
      }
    }
  } else {
    // Prompt 7: Project name — always show, pre-fill with prior if available
    const defaultName = setupScope === 'global' ? 'global' : 'my-project'
    const nameDefault = priorProjectName ?? defaultName
    const projectNameMessage = priorProjectName
      ? `Project name? (previous: ${priorProjectName})`
      : 'Project name?'
    const projectNameResult = await p.text({
      message: projectNameMessage,
      placeholder: nameDefault,
      defaultValue: nameDefault,
      validate: (value) => validateFilesystemSafeName(value, 'Project name'),
    })

    if (p.isCancel(projectNameResult)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    projectName = setupScope === 'global' ? 'global' : projectNameResult
  }

  if (!opts.cliOverrides.cliTools) {
    const cliToolOptions = getCliToolOptions(opts.targetDir)
    if (cliToolOptions.length > 0) {
      const initialCliTools = detectInstalledCliToolsFromCatalog(opts.targetDir)

      const cliToolsResult = await p.multiselect({
        message: 'Which CLI tools do you have installed? (press space to select, enter to confirm or skip)',
        options: cliToolOptions,
        initialValues: initialCliTools,
        required: false,
      })

      if (p.isCancel(cliToolsResult)) {
        p.cancel('Setup cancelled.')
        process.exit(0)
      }

      cliTools = (cliToolsResult as string[]) || []
    }
  }

  const orgResult = await p.text({
    message: `Organization? (optional${organization ? `, previous: ${organization}` : ''})`,
    placeholder: organization || 'Acme',
    defaultValue: organization,
  })

  if (p.isCancel(orgResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  organization = orgResult

  const teamResult = await p.text({
    message: `Team? (optional${team ? `, previous: ${team}` : ''})`,
    placeholder: team || 'Platform',
    defaultValue: team,
  })

  if (p.isCancel(teamResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  team = teamResult

  return {
    setupScope,
    tools: tools as ToolId[],
    skills,
    agents,
    mcpPreset,
    projectName,
    ...(workspaceName ? { workspaceName } : {}),
    ...(workspaceRoot ? { workspaceRoot } : {}),
    ...(planningRepoPath ? { planningRepoPath } : {}),
    ...(repos.length > 0 ? { repos } : {}),
    ...(cliTools.length > 0 ? { cliTools } : {}),
    ...(enableServers.length > 0 ? { enableServers } : {}),
    ...(organization ? { organization } : {}),
    ...(team ? { team } : {}),
  }
}
