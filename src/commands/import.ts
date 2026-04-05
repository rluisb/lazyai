import { Command } from 'commander';
import { type MigrationCommandOptions, runMigrationCommand } from './migration-shared.js';

type ImportOptions = MigrationCommandOptions

export function createImportCommand(): Command {
  const command = new Command('import')
    .description('Import an existing AI setup into ai-setup format')
    .argument('[path]', 'Path to existing setup (defaults to current directory)')
    .option('-p, --preview', 'Preview changes without executing')
    .option('-s, --strategy <strategy>', 'Merge strategy: smart, preserve, replace, append', 'smart')
    .option('-v, --verbose', 'Show detailed output')
    .option('-i, --interactive', 'Resolve merge conflicts interactively')
    .option('--skip-backup', 'Skip creating backup')
    .option('-y, --yes', 'Auto-confirm without prompts')
    .action(async (sourcePath: string | undefined, options: ImportOptions) => {
      await runMigrationCommand('import', sourcePath, options);
    });

  return command;
}
