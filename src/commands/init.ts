import type { Command } from 'commander'
import type { SetupType, ToolId } from '../types.js'
import { runWizard } from '../wizard/index.js'
import * as p from '@clack/prompts'
import pc from 'picocolors'
import { detectAdapters, importSetup } from '../migration/index.js'

interface InitOptions {
  type?: SetupType
  tools?: string
  name?: string
  force?: boolean
  interactive: boolean
  migrate?: boolean
  from?: string
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
    .option('--migrate', 'Migrate existing AI setup (detects and imports)')
    .option('--from <path>', 'Path to existing setup for migration (defaults to current directory)')
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

      // Check if we should migrate
      if (opts.migrate) {
        const sourcePath = opts.from || process.cwd()
        
        p.intro(pc.blue('ai-setup init --migrate'))
        
        const spinner = p.spinner()
        spinner.start('Detecting existing AI setups...')
        
        const adapters = await detectAdapters(sourcePath)
        
        if (adapters.length === 0) {
          spinner.stop(pc.yellow('No existing AI setup detected'))
          console.log(pc.gray(`Searched in: ${sourcePath}`))
          console.log(pc.gray('Continuing with fresh init...'))
        } else {
          spinner.stop(pc.green(`Found ${adapters.length} adapter(s): ${adapters.join(', ')}`))
          
          // Run migration
          spinner.start('Migrating existing setup...')
          
          const result = await importSetup({
            path: sourcePath,
            preview: false,
            mergeStrategy: 'smart',
          })
          
          spinner.stop('Migration complete')
          
          if (result.success) {
            console.log(pc.green('\n✅ Successfully migrated existing setup!'))
            console.log(pc.gray(`\nMigrated ${result.stats.filesCreated + result.stats.filesModified} files`))
            
            if (result.backupPath) {
              console.log(pc.gray(`Backup created at: ${result.backupPath}`))
            }
            
            // Continue with wizard for any additional configuration
            console.log(pc.blue('\nContinuing with init wizard for additional configuration...'))
          } else {
            console.log(pc.yellow('\n⚠️  Migration had issues, continuing with fresh init...'))
            if (result.errors.length > 0) {
              console.log(pc.gray('Errors: ' + result.errors.join(', ')))
            }
          }
        }
      }

      const wizardOpts = {
        interactive: opts.interactive !== false,
        cliOverrides,
        targetDir: process.cwd(),
        ...(opts.force !== undefined ? { force: opts.force } : {}),
      }

      await runWizard(wizardOpts)
    })
}
