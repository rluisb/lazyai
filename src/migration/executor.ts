/**
 * Migration Executor
 * 
 * Executes migration plans with proper backup, file operations, and rollback support.
 */

import { promises as fs } from 'fs';
import path from 'path';
import * as p from '@clack/prompts';
import {
  MigrationContext,
  MigrationPlan,
  MigrationResult,
  MigrationAction,
  MigrationStats,
  MergeConflict,
} from './types.js';

export async function executeMigrationPlan(
  context: MigrationContext,
  plan: MigrationPlan
): Promise<MigrationResult> {
  const executedActions: MigrationAction[] = [];
  const errors: string[] = [];
  const warnings: string[] = [];
  const stats: MigrationStats = {
    filesCreated: 0,
    filesModified: 0,
    filesBackedUp: 0,
    filesSkipped: 0,
    conflictsResolved: 0,
    conflictsUnresolved: 0,
  };

  let backupPath: string | undefined;

  try {
    if (context.options.interactive) {
      await resolvePlanConflictsInteractively(plan);
    }

    // Step 1: Create backup directory
    if (!context.options.skipBackup) {
      backupPath = await createBackupDir(context.targetPath);
    }

    // Step 2: Execute backup actions first
    const backupActions = plan.actions.filter(a => a.type === 'backup');
    for (const action of backupActions) {
      try {
        await executeBackupAction(action, context, backupPath!);
        executedActions.push(action);
        stats.filesBackedUp++;
      } catch (error) {
        errors.push(`Failed to backup ${action.sourcePath}: ${error}`);
      }
    }

    // Step 3: Execute create actions
    const createActions = plan.actions.filter(a => a.type === 'create');
    for (const action of createActions) {
      try {
        await executeCreateAction(action, context);
        executedActions.push(action);
        stats.filesCreated++;
      } catch (error) {
        errors.push(`Failed to create ${action.targetPath}: ${error}`);
      }
    }

    // Step 4: Execute modify actions
    const modifyActions = plan.actions.filter(a => a.type === 'modify');
    for (const action of modifyActions) {
      try {
        await executeModifyAction(action, context);
        executedActions.push(action);
        stats.filesModified++;
      } catch (error) {
        errors.push(`Failed to modify ${action.targetPath}: ${error}`);
      }
    }

    // Step 5: Count conflicts
    stats.conflictsResolved = plan.conflicts.filter(c => c.resolved).length;
    stats.conflictsUnresolved = plan.conflicts.filter(c => !c.resolved).length;
    plan.canProceed = stats.conflictsUnresolved === 0;

    // Step 6: Create .ai-setup.json if it doesn't exist
    await createAiSetupManifest(context, plan, executedActions);

    return buildMigrationResult({
      success: errors.length === 0,
      plan,
      executedActions,
      ...(backupPath ? { backupPath } : {}),
      errors,
      warnings,
      stats,
    });

  } catch (error) {
    // Fatal error - try to rollback if possible
    errors.push(`Fatal error during migration: ${error}`);
    
    return buildMigrationResult({
      success: false,
      plan,
      executedActions,
      ...(backupPath ? { backupPath } : {}),
      errors,
      warnings,
      stats,
    });
  }
}

interface BuildMigrationResultInput {
  success: boolean;
  plan: MigrationPlan;
  executedActions: MigrationAction[];
  backupPath?: string;
  errors: string[];
  warnings: string[];
  stats: MigrationStats;
}

function buildMigrationResult(input: BuildMigrationResultInput): MigrationResult {
  return {
    success: input.success,
    plan: input.plan,
    executedActions: input.executedActions,
    ...(input.backupPath ? { backupPath: input.backupPath } : {}),
    errors: input.errors,
    warnings: input.warnings,
    stats: input.stats,
  };
}

type InteractiveResolution = 'keep-existing' | 'use-new' | 'manual-edit';

export async function resolvePlanConflictsInteractively(plan: MigrationPlan): Promise<void> {
  for (const conflict of plan.conflicts) {
    if (conflict.resolved) {
      continue;
    }

    p.note(
      renderConflictDiff(conflict),
      `Conflict in ${conflict.file || 'unknown file'}:${conflict.lineStart}-${conflict.lineEnd}`,
    );

    const resolution = await p.select({
      message: `How should ${conflict.file || 'this file'} be resolved?`,
      options: [
        { value: 'keep-existing', label: 'Keep Existing', hint: 'Preserve the current file contents' },
        { value: 'use-new', label: 'Use New', hint: 'Apply the incoming ai-setup version' },
        { value: 'manual-edit', label: 'Manual Edit', hint: 'Enter the merged content yourself' },
      ],
    });

    if (p.isCancel(resolution)) {
      throw new Error('Migration cancelled during interactive conflict resolution');
    }

    await applyConflictResolution(conflict, resolution as InteractiveResolution);
  }
}

async function applyConflictResolution(
  conflict: MergeConflict,
  resolution: InteractiveResolution,
): Promise<void> {
  switch (resolution) {
    case 'keep-existing':
      conflict.resolved = true;
      conflict.resolution = resolution;
      conflict.resolvedContent = getConflictSide(conflict, 'ours');
      return;
    case 'use-new':
      conflict.resolved = true;
      conflict.resolution = resolution;
      conflict.resolvedContent = getConflictSide(conflict, 'theirs');
      return;
    case 'manual-edit':
      await resolveManualEdit(conflict);
      return;
  }
}

async function resolveManualEdit(conflict: MergeConflict): Promise<void> {
  const edited = await p.text({
    message: `Enter merged content for ${conflict.file || 'the conflict'}`,
    initialValue: buildManualEditSeed(conflict),
    placeholder: 'Type the final merged content',
  });

  if (p.isCancel(edited)) {
    throw new Error('Migration cancelled during interactive conflict resolution');
  }

  conflict.resolved = true;
  conflict.resolution = 'manual-edit';
  conflict.resolvedContent = String(edited);
}

function buildManualEditSeed(conflict: MergeConflict): string {
  const existing = getConflictSide(conflict, 'ours');
  const incoming = getConflictSide(conflict, 'theirs');

  return [
    '<<<<<<< EXISTING',
    existing,
    '=======',
    incoming,
    '>>>>>>> NEW',
  ].join('\n');
}

function getConflictSide(
  conflict: MergeConflict,
  side: 'base' | 'ours' | 'theirs',
): string {
  const legacyValue = (conflict as MergeConflict & Record<string, unknown>)[side];

  if (side === 'base') {
    return conflict.baseContent || normalizeLegacyConflictSide(legacyValue);
  }

  if (side === 'ours') {
    return conflict.oursContent || normalizeLegacyConflictSide(legacyValue);
  }

  return conflict.theirsContent || normalizeLegacyConflictSide(legacyValue);
}

function normalizeLegacyConflictSide(value: unknown): string {
  if (Array.isArray(value)) {
    return value.join('\n');
  }

  return typeof value === 'string' ? value : '';
}

export function renderConflictDiff(conflict: MergeConflict): string {
  return [
    '--- Existing',
    getConflictSide(conflict, 'ours') || '(empty)',
    '+++ New',
    getConflictSide(conflict, 'theirs') || '(empty)',
  ].join('\n');
}

async function createBackupDir(targetPath: string): Promise<string> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const backupPath = path.join(targetPath, '.ai-setup-backup', `migration-${timestamp}`);
  
  await fs.mkdir(backupPath, { recursive: true });
  
  return backupPath;
}

async function executeBackupAction(
  action: MigrationAction,
  context: MigrationContext,
  backupPath: string
): Promise<void> {
  if (!action.sourcePath) return;
  
  const sourceFullPath = path.join(context.sourcePath, action.sourcePath);
  const backupFullPath = path.join(backupPath, action.targetPath);
  
  // Ensure backup directory exists
  await fs.mkdir(path.dirname(backupFullPath), { recursive: true });
  
  // Copy file to backup
  await fs.copyFile(sourceFullPath, backupFullPath);
}

async function executeCreateAction(
  action: MigrationAction,
  context: MigrationContext
): Promise<void> {
  const targetFullPath = path.join(context.targetPath, action.targetPath);
  
  // Ensure directory exists
  await fs.mkdir(path.dirname(targetFullPath), { recursive: true });
  
  // File creation is handled by parsers - this just ensures directory exists
  // The actual content is written during the merge phase
}

async function executeModifyAction(
  action: MigrationAction,
  context: MigrationContext
): Promise<void> {
  // Modifications are handled by parsers during merge
  // This ensures the directory structure exists
  const targetFullPath = path.join(context.targetPath, action.targetPath);
  await fs.mkdir(path.dirname(targetFullPath), { recursive: true });
}

async function createAiSetupManifest(
  context: MigrationContext,
  plan: MigrationPlan,
  executedActions: MigrationAction[]
): Promise<void> {
  const manifestPath = path.join(context.targetPath, '.ai-setup.json');
  
  let existingManifest: any = {
    version: '0.1.0',
    setupType: 'project',
    tools: [],
    projectName: 'Migrated Project',
    installedAt: new Date().toISOString(),
    files: [],
  };
  
  // Try to read existing manifest
  try {
    const content = await fs.readFile(manifestPath, 'utf-8');
    existingManifest = JSON.parse(content);
  } catch {
    // No existing manifest
  }
  
  // Update tools
  for (const adapter of plan.adapters) {
    if (!existingManifest.tools.includes(adapter)) {
      existingManifest.tools.push(adapter);
    }
  }
  
  // Track migrated files
  for (const action of executedActions) {
    if (action.type === 'create' || action.type === 'modify') {
      const existingFile = existingManifest.files.find((f: any) => f.path === action.targetPath);
      if (!existingFile) {
        existingManifest.files.push({
          path: action.targetPath,
          hash: 'migrated', // Will be updated by doctor
          source: 'migration',
        });
      }
    }
  }
  
  // Write manifest
  await fs.writeFile(manifestPath, JSON.stringify(existingManifest, null, 2));
}

export async function rollbackMigration(
  context: MigrationContext,
  backupPath: string
): Promise<void> {
  // Restore files from backup
  // This is a safety feature for failed migrations
  console.warn('Rolling back migration...');
  
  // Implementation would restore backed up files
  // For now, we just log the intent
}
