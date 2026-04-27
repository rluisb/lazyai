import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { Command } from 'commander'
import { registerAdd } from './commands/add.js'
import { registerCompile } from './commands/compile.js'
import { registerCompletions } from './commands/completions.js'
import { registerCreate } from './commands/create.js'
import { registerDoctor } from './commands/doctor.js'
import { registerEject } from './commands/eject.js'
import { createImportCommand } from './commands/import.js'
import { registerInfo } from './commands/info.js'
import { registerInit } from './commands/init.js'
import { registerList } from './commands/list.js'
import { createMigrateCommand } from './commands/migrate.js'
import { registerOrchestration } from './commands/orchestration.js'
import { registerServer } from './commands/server.js'
import { registerSetup } from './commands/setup.js'
import { registerStatus } from './commands/status.js'
import { registerUpdate } from './commands/update.js'
import type { TomlConfig } from './utils/toml.js'
import { loadConfig } from './utils/toml.js'

const __dirname = dirname(fileURLToPath(import.meta.url))

let cachedTomlConfig: TomlConfig | null = null

/** Get the TOML config loaded at startup. */
export function getTomlConfig(): TomlConfig {
  return cachedTomlConfig ?? {}
}

function getVersion(): string {
  try {
    const pkg = JSON.parse(readFileSync(join(__dirname, '../package.json'), 'utf-8')) as { version: string }
    return pkg.version
  } catch {
    return '0.0.0'
  }
}

export function createProgram(): Command {
  const program = new Command()

  program
    .name('ai-setup')
    .description('AI development environment scaffold — one command to set up your AI tools')
    .version(getVersion())
    .option('-v, --verbose', 'Enable verbose debug output')
    .addHelpText('after', '\nGlobal flags:\n  -v, --verbose           Enable verbose debug output')
    .hook('preAction', (thisCommand) => {
      if (thisCommand.opts().verbose) {
        process.env.AI_SETUP_DEBUG = '1'
      }
      // Load TOML config once at startup
      if (!cachedTomlConfig) {
        cachedTomlConfig = loadConfig(process.cwd())
      }
    })

  registerInit(program)
  registerAdd(program)
  registerUpdate(program)
  registerDoctor(program)
  registerStatus(program)
  registerCreate(program)
  registerEject(program)
  registerCompile(program)
  registerList(program)
  registerInfo(program)
  registerSetup(program)
  registerOrchestration(program)
  registerServer(program)
  registerCompletions(program)

  // Add migration commands
  program.addCommand(createImportCommand())
  program.addCommand(createMigrateCommand())

  return program
}

export async function run(argv?: string[]): Promise<void> {
  const program = createProgram()
  await program.parseAsync(argv ?? process.argv)
}
