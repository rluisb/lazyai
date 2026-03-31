import type { Command } from 'commander'
import type { SetupType, ToolId } from '../types.js'
import { runWizard } from '../wizard/index.js'

interface InitOptions {
  type?: SetupType
  tools?: string
  name?: string
  force?: boolean
  interactive: boolean
}

export function registerInit(program: Command): void {
  program
    .command('init')
    .description('Scaffold AI development environment in the current directory')
    .option('--type <type>', 'Setup type: project | workspace')
    .option('--tools <tools>', 'Comma-separated tool list: pi,opencode')
    .option('--name <name>', 'Project name (defaults to directory name)')
    .option('--force', 'Overwrite all existing managed files (creates backups)')
    .option('--no-interactive', 'Non-interactive mode — requires all flags')
    .action(async (opts: InitOptions) => {
      const tools = opts.tools
        ? (opts.tools.split(',').map((t) => t.trim()).filter(Boolean) as ToolId[])
        : undefined

      const cliOverrides: {
        type?: SetupType
        tools?: ToolId[]
        name?: string
      } = {}

      if (opts.type) cliOverrides.type = opts.type
      if (tools) cliOverrides.tools = tools
      if (opts.name) cliOverrides.name = opts.name

      const wizardOpts = {
        interactive: opts.interactive !== false,
        cliOverrides,
        targetDir: process.cwd(),
        ...(opts.force !== undefined ? { force: opts.force } : {}),
      }

      await runWizard(wizardOpts)
    })
}
