import path from 'node:path'
import * as p from '@clack/prompts'
import { Errors } from './errors/index.js'
import type { SetupConfig, SetupScope, SetupType, ToolId } from './types.js'
import { validateFilesystemSafeName } from './utils/validation.js'

function cancelPrompt(): never {
  p.cancel('Setup cancelled.')
  throw Errors.userCancelled()
}

export interface PromptOptions {
  scope?: SetupScope
  type?: SetupType
  tools?: ToolId[]
  name?: string
  interactive: boolean
}

export async function runPrompts(opts: PromptOptions): Promise<SetupConfig> {
  p.intro('🤖  ai-setup — AI development environment scaffold')

  const config: Partial<SetupConfig> = {
    targetDir: process.cwd(),
  }

  // Setup scope
  if (opts.scope || opts.type) {
    const setupScope = (opts.scope ?? opts.type) as SetupScope
    config.setupScope = setupScope
    config.setupType = setupScope
  } else if (!opts.interactive) {
    throw new Error('--scope is required in non-interactive mode (global | workspace | project)')
  } else {
    const setupScope = await p.select({
      message: 'Setup scope:',
      options: [
        { value: 'global', label: 'Global', hint: 'Install to ~/.ai/ + native tool global paths' },
        { value: 'workspace', label: 'Workspace', hint: 'Planning repo with multi-project management' },
        { value: 'project', label: 'Project', hint: 'Self-contained single repository' },
      ],
    })
    if (p.isCancel(setupScope)) {
      cancelPrompt()
    }
    config.setupScope = setupScope as SetupScope
    config.setupType = setupScope as SetupScope
  }

  // Tool selection
  if (opts.tools && opts.tools.length > 0) {
    config.tools = opts.tools
  } else if (!opts.interactive) {
    throw new Error('--tools is required in non-interactive mode (pi, opencode, claude-code, gemini, copilot)')
  } else {
    const tools = await p.multiselect({
      message: 'Which AI tools are you using?',
      options: [
        { value: 'pi', label: 'Pi (Claude Code)', hint: 'Uses .pi/ directory + CLAUDE.md' },
        { value: 'opencode', label: 'OpenCode', hint: 'Uses .opencode/ directory + AGENTS.md' },
        { value: 'claude-code', label: 'Claude Code', hint: 'Uses .claude/ directory + CLAUDE.md' },
        { value: 'gemini', label: 'Gemini CLI', hint: 'Uses .gemini/ directory + GEMINI.md' },
        { value: 'copilot', label: 'GitHub Copilot', hint: 'Uses .github/ + copilot-instructions.md' },
      ],
      required: true,
    })
    if (p.isCancel(tools)) {
      cancelPrompt()
    }
    config.tools = tools as ToolId[]
  }

  // Project name
  const defaultName = path.basename(process.cwd())
  if (opts.name) {
    config.projectName = opts.name
  } else if (!opts.interactive) {
    config.projectName = defaultName
  } else {
    const projectName = await p.text({
      message: 'Project name?',
      placeholder: defaultName,
      defaultValue: defaultName,
      validate: (value) => validateFilesystemSafeName(value, 'Project name'),
    })
    if (p.isCancel(projectName)) {
      cancelPrompt()
    }
    config.projectName = projectName
  }

  return config as SetupConfig
}

export function outroSuccess(config: SetupConfig): void {
  const rootFiles: string[] = []
  if (config.tools.includes('opencode')) rootFiles.push('AGENTS.md')
  if (config.tools.includes('pi') || config.tools.includes('claude-code')) rootFiles.push('CLAUDE.md')
  if (config.tools.includes('gemini')) rootFiles.push('GEMINI.md')
  if (config.tools.includes('copilot')) rootFiles.push('.github/copilot-instructions.md')

  const fileList = rootFiles.length > 0 ? rootFiles.join(', ') : 'your config files'

  p.outro(`✅  Setup complete for ${config.projectName}!

Next steps:
  1. Open ${fileList} and fill in the [YOUR_*] placeholders
  2. Review .ai-setup.json to see what was installed
  3. Commit the generated files to your repository
  `)
}
