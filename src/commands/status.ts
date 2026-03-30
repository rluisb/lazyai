import type { Command } from 'commander'

export function registerStatus(program: Command): void {
  program
    .command('status')
    .description('Show current setup status')
    .action(() => {
      console.log('⏳  status — coming soon')
    })
}
