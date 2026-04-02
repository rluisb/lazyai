import * as p from '@clack/prompts'
import path from 'node:path'
import { homedir } from 'node:os'
import { readManifest } from '../utils/manifest.js'
import { validateFilesystemSafeName } from '../utils/validation.js'
import { fileExists } from '../utils/files.js'
import type { SetupScope, SetupType, ToolId, WizardSelections } from '../types.js'

export interface Phase1Result {
  setupScope: SetupScope
  tools: ToolId[]
  projectName: string
  planningRepoPath?: string
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
    planningRepoPath?: string
  }
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    name?: string
    planningRepo?: string
  }
  targetDir: string
}): Promise<Phase1Result> {
  // Non-interactive mode: use cliOverrides or throw
  if (!opts.interactive) {
    const setupScope = opts.cliOverrides.scope ?? opts.cliOverrides.type
    const tools = opts.cliOverrides.tools
    const projectName = opts.cliOverrides.name ?? (setupScope === 'global' ? 'global' : undefined)
    const planningRepoPath = opts.cliOverrides.planningRepo

    if (!setupScope) {
      throw new Error('--scope is required in non-interactive mode (global | workspace | project)')
    }
    if (!tools || tools.length === 0) {
      throw new Error('--tools is required in non-interactive mode (pi, opencode, claude-code, gemini, copilot)')
    }
    if (!projectName) {
      throw new Error('Project name is required in non-interactive mode (use --name or provide via config)')
    }

    return {
      setupScope,
      tools,
      projectName,
      ...(setupScope === 'workspace' && planningRepoPath ? { planningRepoPath } : {}),
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
  let planningRepoPath: string | undefined

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

  if (setupScope === 'workspace') {
    const defaultPlanningRepoPath = opts.cliOverrides.planningRepo || opts.prior.planningRepoPath || opts.targetDir
    const planningRepoPathResult =
      opts.cliOverrides.planningRepo ||
      opts.prior.planningRepoPath ||
      (await p.text({
        message: 'Planning repo location?',
        placeholder: defaultPlanningRepoPath,
        defaultValue: defaultPlanningRepoPath,
      }))

    if (p.isCancel(planningRepoPathResult)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }

    planningRepoPath = planningRepoPathResult
  }

  return {
    setupScope,
    tools: tools as ToolId[],
    projectName,
    ...(planningRepoPath ? { planningRepoPath } : {}),
  }
}
