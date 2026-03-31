import { Command } from 'commander'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { registerInit } from './commands/init.js'
import { registerAdd } from './commands/add.js'
import { registerUpdate } from './commands/update.js'
import { registerDoctor } from './commands/doctor.js'
import { registerStatus } from './commands/status.js'
import { registerCreate } from './commands/create.js'

const __dirname = dirname(fileURLToPath(import.meta.url))

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

  registerInit(program)
  registerAdd(program)
  registerUpdate(program)
  registerDoctor(program)
  registerStatus(program)
  registerCreate(program)

  return program
}

export async function run(argv?: string[]): Promise<void> {
  const program = createProgram()
  await program.parseAsync(argv ?? process.argv)
}
