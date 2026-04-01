import * as p from '@clack/prompts'
import path from 'node:path'
import type { SetupConfig, SetupType, ToolId } from './types.js'
import { validateFilesystemSafeName } from './utils/validation.js'

export interface PromptOptions {
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

  // Setup type
  if (opts.type) {
    config.setupType = opts.type
  } else if (!opts.interactive) {
    throw new Error('--type is required in non-interactive mode (project | workspace)')
  } else {
    const setupType = await p.select({
      message: 'What are you setting up?',
      options: [
        { value: 'project', label: 'Project', hint: 'For a single repository' },
        { value: 'workspace', label: 'Workspace', hint: 'For a multi-repo organization' },
      ],
    })
    if (p.isCancel(setupType)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    config.setupType = setupType as SetupType
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
      p.cancel('Setup cancelled.')
      process.exit(0)
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
      p.cancel('Setup cancelled.')
      process.exit(0)
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
