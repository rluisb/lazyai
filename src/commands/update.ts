import type { Command } from 'commander'

export function registerUpdate(program: Command): void {
  program
    .command('update')
    .description('Update ai-setup library files (skips customized files)')
    .action(() => {
      console.log('⏳  update — coming soon')
    })
}
