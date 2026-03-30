import * as p from '@clack/prompts'
import path from 'node:path'
import type { SetupConfig, SetupType, ToolId } from './types.js'

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
    throw new Error('--tools is required in non-interactive mode (pi, opencode)')
  } else {
    const tools = await p.multiselect({
      message: 'Which AI tools are you using?',
      options: [
        { value: 'pi', label: 'Pi (Claude Code)', hint: 'Uses .pi/ directory + CLAUDE.md' },
        { value: 'opencode', label: 'OpenCode', hint: 'Uses .opencode/ directory + AGENTS.md' },
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
  p.outro(`✅  Setup complete for ${config.projectName}!

Next steps:
  1. Open AGENTS.md and fill in the [YOUR_*] placeholders
  2. Review .ai-setup.json to see what was installed
  3. Commit the generated files to your repository
  `)
}
