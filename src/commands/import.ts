/**
 * Import Command
 * 
 * Imports existing AI setups into ai-setup format.
 * Primary command for the migration engine.
 */

import { Command } from 'commander';
import * as p from '@clack/prompts';
import pc from 'picocolors';
import path from 'path';
import { importSetup, detectAdapters, formatPlan } from '../migration/index.js';
import { MergeStrategy } from '../migration/types.js';

interface ImportOptions {
  path?: string;
  preview?: boolean;
  strategy?: string;
  verbose?: boolean;
  skipBackup?: boolean;
  yes?: boolean;
}

export function createImportCommand(): Command {
  const command = new Command('import')
    .description('Import an existing AI setup into ai-setup format')
    .argument('[path]', 'Path to existing setup (defaults to current directory)')
    .option('-p, --preview', 'Preview changes without executing')
    .option('-s, --strategy <strategy>', 'Merge strategy: smart, preserve, replace, append', 'smart')
    .option('-v, --verbose', 'Show detailed output')
    .option('--skip-backup', 'Skip creating backup')
    .option('-y, --yes', 'Auto-confirm without prompts')
    .action(async (sourcePath: string | undefined, options: ImportOptions) => {
      await handleImport(sourcePath, options);
    });

  return command;
}

async function handleImport(
  sourcePath: string | undefined,
  options: ImportOptions
): Promise<void> {
  const targetPath = process.cwd();
  const resolvedSource = sourcePath ? path.resolve(sourcePath) : targetPath;

  p.intro(pc.blue('ai-setup import'));

  // Validate merge strategy
  const validStrategies: MergeStrategy[] = ['smart', 'preserve', 'replace', 'append'];
  const strategy = (options.strategy?.toLowerCase() || 'smart') as MergeStrategy;
  
  if (!validStrategies.includes(strategy)) {
    p.cancel(`Invalid merge strategy: ${options.strategy}`);
    console.log(pc.yellow(`Valid strategies: ${validStrategies.join(', ')}`));
    process.exit(1);
  }

  // Detect existing adapters
  const spinner = p.spinner();
  spinner.start('Scanning for existing AI setups...');
  
  const adapters = await detectAdapters(resolvedSource);
  
  if (adapters.length === 0) {
    spinner.stop(pc.yellow('No existing AI setup detected'));
    console.log(pc.gray(`Searched in: ${resolvedSource}`));
    console.log(pc.gray('Run with a specific path: ai-setup import /path/to/project'));
    process.exit(1);
  }

  spinner.stop(pc.green(`Found ${adapters.length} adapter(s): ${adapters.join(', ')}`));

  // Run import
  spinner.start('Analyzing existing setup...');
  
  const result = await importSetup({
    path: resolvedSource,
    preview: options.preview,
    mergeStrategy: strategy,
    verbose: options.verbose,
    skipBackup: options.skipBackup,
  });

  spinner.stop('Analysis complete');

  // Show preview
  if (options.preview) {
    console.log('\n' + formatPlan(result.plan));
    console.log(pc.gray('\nThis was a preview. Run without --preview to execute.'));
    return;
  }

  // Show plan
  console.log('\n' + formatPlan(result.plan));

  // Check if migration can proceed
  if (!result.plan.canProceed) {
    console.log(pc.red('\n❌ Migration blocked due to unresolved conflicts'));
    console.log(pc.yellow('Run with --strategy preserve or --strategy replace to override'));
    process.exit(1);
  }

  // Confirm execution
  if (!options.yes) {
    const confirm = await p.confirm({
      message: 'Proceed with migration?',
      initialValue: false,
    });

    if (!confirm) {
      p.cancel('Migration cancelled');
      process.exit(0);
    }
  }

  // Execute migration
  spinner.start('Migrating...');
  
  // Re-run without preview to execute
  const executeResult = await importSetup({
    path: resolvedSource,
    preview: false,
    mergeStrategy: strategy,
    verbose: options.verbose,
    skipBackup: options.skipBackup,
  });

  spinner.stop('Migration complete');

  // Show results
  if (executeResult.success) {
    console.log(pc.green('\n✅ Migration successful!'));
    console.log(pc.gray(`\nStatistics:`));
    console.log(`  Files created: ${executeResult.stats.filesCreated}`);
    console.log(`  Files modified: ${executeResult.stats.filesModified}`);
    console.log(`  Files backed up: ${executeResult.stats.filesBackedUp}`);
    
    if (executeResult.stats.conflictsUnresolved > 0) {
      console.log(pc.yellow(`  Conflicts unresolved: ${executeResult.stats.conflictsUnresolved}`));
    }

    if (executeResult.backupPath) {
      console.log(pc.gray(`\nBackup created at: ${executeResult.backupPath}`));
    }

    console.log(pc.blue('\nNext steps:'));
    console.log('  1. Review the migrated files');
    console.log('  2. Run ai-setup doctor to verify integrity');
    console.log('  3. Commit the changes');
  } else {
    console.log(pc.red('\n❌ Migration failed'));
    
    if (executeResult.errors.length > 0) {
      console.log(pc.red('\nErrors:'));
      for (const error of executeResult.errors) {
        console.log(`  • ${error}`);
      }
    }

    if (executeResult.backupPath) {
      console.log(pc.yellow(`\nBackup available at: ${executeResult.backupPath}`));
    }

    process.exit(1);
  }

  p.outro('Done!');
}
