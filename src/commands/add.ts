import type { Command } from 'commander'

export function registerAdd(program: Command): void {
  program
    .command('add')
    .description('Add a tool to existing setup')
    .argument('<tool>', 'Tool to add: pi | opencode')
    .action(() => {
      console.log('⏳  add — coming soon')
    })
}
