import type { Command } from 'commander'

export function registerDoctor(program: Command): void {
  program
    .command('doctor')
    .description('Verify setup integrity against .ai-setup.json')
    .action(() => {
      console.log('⏳  doctor — coming soon')
    })
}
