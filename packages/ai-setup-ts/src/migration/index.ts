/**
 * Migration Engine - Main Entry Point
 * 
 * Provides functionality to import existing AI setups into ai-setup format
 * with intelligent merging and conflict resolution.
 */

export * from './detector.js';
export * from './parsers/base-parser.js';
export * from './registry/discovery.js';
export * from './types.js';

import { detectExistingSetup, hasAdapter } from './detector.js';
import { checkDrift } from './doctor.js';
import { executeMigrationPlan, executeMigrationToCanonical } from './executor.js';
import { generateMigrationPlan } from './plan.js';
import { getAllParsers } from './registry/discovery.js';
import type {
  DriftCheckResult,
  MergeStrategy,
  MigrationContext,
  MigrationPlan,
  MigrationResult,
  ParsedSetup,
} from './types.js';

export interface ImportOptions {
  path?: string;
  preview?: boolean;
  mergeStrategy?: MergeStrategy;
  verbose?: boolean;
  skipBackup?: boolean;
  interactive?: boolean;
  canonicalOutput?: boolean;
}

/**
 * Import an existing AI setup into ai-setup format
 */
export async function importSetup(options: ImportOptions = {}): Promise<MigrationResult> {
  const sourcePath = options.path || process.cwd();
  const mergeStrategy = options.mergeStrategy || 'smart';
  const canonicalOutput = options.canonicalOutput !== false;
  
  const context: MigrationContext = {
    sourcePath,
    targetPath: process.cwd(),
    options: {
      preview: options.preview || false,
      mergeStrategy,
      verbose: options.verbose || false,
      skipBackup: options.skipBackup || false,
      interactive: options.interactive || false,
    },
  };

  try {
    // Step 1: Detect existing setups
    const detections = await detectExistingSetup(context);
    
    if (detections.length === 0) {
      return {
        success: false,
        plan: null,
        executedActions: [],
        errors: [`No existing AI setup detected in ${sourcePath}`],
        warnings: [],
        stats: {
          filesCreated: 0,
          filesModified: 0,
          filesBackedUp: 0,
          filesSkipped: 0,
          conflictsResolved: 0,
          conflictsUnresolved: 0,
        },
      };
    }

    // Step 2: Load all parsers
    const parsers = await getAllParsers(context.targetPath);

    // Step 3: Generate migration plan
    const plan = await generateMigrationPlan(context, detections, parsers);

    if (options.preview) {
      return {
        success: true,
        plan,
        executedActions: [],
        errors: [],
        warnings: plan.conflicts.map(c => `Conflict in ${c.file || 'unknown file'}`),
        stats: {
          filesCreated: plan.actions.filter(a => a.type === 'create').length,
          filesModified: plan.actions.filter(a => a.type === 'modify').length,
          filesBackedUp: 0,
          filesSkipped: plan.actions.filter(a => a.type === 'skip').length,
          conflictsResolved: plan.conflicts.filter(c => c.resolved).length,
          conflictsUnresolved: plan.conflicts.filter(c => !c.resolved).length,
        },
      };
    }

    if (!plan.canProceed && !context.options.interactive) {
      return {
        success: false,
        plan,
        executedActions: [],
        errors: ['Migration cannot proceed due to unresolved conflicts'],
        warnings: plan.conflicts.map(c => `Conflict in ${c.file || 'unknown file'}`),
        stats: {
          filesCreated: 0,
          filesModified: 0,
          filesBackedUp: 0,
          filesSkipped: 0,
          conflictsResolved: 0,
          conflictsUnresolved: plan.conflicts.length,
        },
      };
    }

    // Step 5: Execute migration
    if (canonicalOutput) {
      const parsedSetups: ParsedSetup[] = [];
      for (const detection of detections) {
        const parser = parsers.find((candidate) => candidate.id === detection.adapterId);
        if (!parser) continue;
        parsedSetups.push(await parser.parse(context));
      }

      const result = await executeMigrationToCanonical({
        context,
        plan,
        parsedSetups,
      });

      return result;
    }

    const result = await executeMigrationPlan(context, plan);

    return result;

  } catch (error) {
    return {
      success: false,
      plan: null,
      executedActions: [],
      errors: [error instanceof Error ? error.message : String(error)],
      warnings: [],
      stats: {
        filesCreated: 0,
        filesModified: 0,
        filesBackedUp: 0,
        filesSkipped: 0,
        conflictsResolved: 0,
        conflictsUnresolved: 0,
      },
    };
  }
}

/**
 * Migrate an existing setup (alias for import)
 */
export async function migrate(options: ImportOptions = {}): Promise<MigrationResult> {
  return importSetup(options);
}

/**
 * Check for drift between current setup and clean ai-setup state
 */
export async function migrationCheck(
  options: { verbose?: boolean; path?: string } = {}
): Promise<DriftCheckResult> {
  const targetPath = options.path || process.cwd();
  
  return checkDrift(targetPath, options.verbose);
}

/**
 * Detect which adapters exist in a directory
 */
export async function detectAdapters(path?: string): Promise<string[]> {
  const sourcePath = path || process.cwd();
  const adapters: string[] = [];

  const adapterIds = ['opencode', 'claude-code', 'gemini', 'copilot'];
  
  for (const adapterId of adapterIds) {
    if (await hasAdapter(sourcePath, adapterId)) {
      adapters.push(adapterId);
    }
  }

  return adapters;
}

export { formatPlan } from "./plan.js";
export type { DriftCheckResult, MigrationPlan, MigrationResult };
export { detectExistingSetup, generateMigrationPlan, getAllParsers };
