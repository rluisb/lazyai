import * as p from '@clack/prompts'
import path from 'node:path'
import { readManifest } from '../utils/manifest.js'
import type { SetupType, ToolId, WizardSelections } from '../types.js'

export interface Phase1Result {
  setupType: SetupType
  tools: ToolId[]
  projectName: string
}

/**
 * Run Phase 1 of the interactive wizard: gather setupType, tools, and projectName.
 * Behavior depends on interactive mode and CLI overrides.
 */
export async function runPhase1(opts: {
  interactive: boolean
  prior: Partial<WizardSelections> & { setupType?: SetupType; tools?: ToolId[]; projectName?: string }
  cliOverrides: { type?: SetupType; tools?: ToolId[]; name?: string }
  targetDir: string
}): Promise<Phase1Result> {
  // Non-interactive mode: use cliOverrides or throw
  if (!opts.interactive) {
    const setupType = opts.cliOverrides.type
    const tools = opts.cliOverrides.tools
    const projectName = opts.cliOverrides.name

    if (!setupType) {
      throw new Error('--type is required in non-interactive mode (project | workspace)')
    }
    if (!tools || tools.length === 0) {
      throw new Error('--tools is required in non-interactive mode (pi, opencode, claude-code, gemini, copilot)')
    }
    if (!projectName) {
      throw new Error('Project name is required in non-interactive mode (use --name or provide via config)')
    }

    return {
      setupType,
      tools,
      projectName,
    }
  }

  // Interactive mode
  p.intro('🤖  ai-setup — AI development environment scaffold')

  // Check for existing manifest and show note if found
  const existingManifest = await readManifest(opts.targetDir)
  if (existingManifest) {
    p.note('Re-running setup — previous selections will be pre-filled')
  }

  let setupType: SetupType
  let tools: ToolId[]
  let projectName: string

  // Prompt 1: Setup type
  const setupTypeResult =
    opts.cliOverrides.type ||
    opts.prior.setupType ||
    (await p.select({
      message: 'What are you setting up?',
      options: [
        { value: 'project', label: 'Project', hint: 'For a single repository' },
        { value: 'workspace', label: 'Workspace', hint: 'For a multi-repo organization' },
      ],
    }))

  if (p.isCancel(setupTypeResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  setupType = setupTypeResult as SetupType

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
  const defaultName = path.basename(opts.targetDir)
  const projectNameResult =
    opts.cliOverrides.name ||
    opts.prior.projectName ||
    (await p.text({
      message: 'Project name?',
      placeholder: defaultName,
      defaultValue: defaultName,
    }))

  if (p.isCancel(projectNameResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }
  projectName = projectNameResult

  return {
    setupType: setupType as SetupType,
    tools: tools as ToolId[],
    projectName,
  }
}
