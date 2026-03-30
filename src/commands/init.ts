import type { Command } from 'commander'
import { runPrompts, outroSuccess } from '../prompts.js'
import type { ToolId, SetupType } from '../types.js'

interface InitOptions {
  type?: string
  tools?: string
  name?: string
  interactive: boolean
}

export function registerInit(program: Command): void {
  program
    .command('init')
    .description('Scaffold AI development environment in the current directory')
    .option('--type <type>', 'Setup type: project | workspace')
    .option('--tools <tools>', 'Comma-separated tool list: pi,opencode')
    .option('--name <name>', 'Project name (defaults to directory name)')
    .option('--no-interactive', 'Non-interactive mode — requires all flags')
    .action(async (opts: InitOptions) => {
      const tools = opts.tools
        ? (opts.tools.split(',').map((t) => t.trim()) as ToolId[])
        : undefined

      const promptOpts: any = {
        interactive: opts.interactive,
      }
      
      if (opts.type) promptOpts.type = opts.type as SetupType
      if (tools) promptOpts.tools = tools
      if (opts.name) promptOpts.name = opts.name

      const config = await runPrompts(promptOpts)

      // Dynamic import to avoid circular deps — scaffold wired in T009
      const { runScaffold } = await import('../scaffold.js')
      await runScaffold(config)

      outroSuccess(config)
    })
}
