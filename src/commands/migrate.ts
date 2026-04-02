import { Command } from 'commander';
import { runMigrationCommand, type MigrationCommandOptions } from './migration-shared.js';

type MigrateOptions = MigrationCommandOptions

export function createMigrateCommand(): Command {
  const command = new Command('migrate')
    .description('Migrate an existing AI setup into ai-setup format (alias for import)')
    .argument('[path]', 'Path to existing setup (defaults to current directory)')
    .option('-p, --preview', 'Preview changes without executing')
    .option('-s, --strategy <strategy>', 'Merge strategy: smart, preserve, replace, append', 'smart')
    .option('-v, --verbose', 'Show detailed output')
    .option('-i, --interactive', 'Resolve merge conflicts interactively')
    .option('--skip-backup', 'Skip creating backup')
    .option('-y, --yes', 'Auto-confirm without prompts')
    .action(async (sourcePath: string | undefined, options: MigrateOptions) => {
      await runMigrationCommand('migrate', sourcePath, options);
    });

  return command;
}
