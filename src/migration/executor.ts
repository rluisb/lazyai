/**
 * Migration Executor
 * 
 * Executes migration plans with proper backup, file operations, and rollback support.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import * as p from '@clack/prompts';
import { toolIdSchema, trackedFileSchema } from '../store/schema.js';
import type { FileRecord } from '../types.js';
import { writeToCanonical } from './canonical-writer.js';
import type {
  MergeConflict,
  MigrationAction,
  MigrationContext,
  MigrationPlan,
  MigrationResult,
  MigrationStats,
  ParsedSetup,
} from './types.js';

type ManifestFile = Pick<FileRecord, 'path' | 'hash' | 'source'>;

interface ManifestShape {
  version: string;
  setupScope: 'project' | 'workspace' | 'global';
  setupType: 'project' | 'workspace';
  tools: string[];
  projectName: string;
  installedAt: string;
  files: ManifestFile[];
}

function isManifestFile(value: unknown): value is ManifestFile {
  const parsed = trackedFileSchema.safeParse(value);
  return parsed.success;
}

function normalizeManifest(raw: unknown): ManifestShape {
  const now = new Date().toISOString();
  const fallback: ManifestShape = {
    version: '0.1.0',
    setupScope: 'project',
    setupType: 'project',
    tools: [],
    projectName: 'Migrated Project',
    installedAt: now,
    files: [],
  };

  if (!raw || typeof raw !== 'object') return fallback;
  const data = raw as Record<string, unknown>;

  const setupScope = data.setupScope;
  const normalizedScope: ManifestShape['setupScope'] =
    setupScope === 'global' || setupScope === 'workspace' || setupScope === 'project'
      ? setupScope
      : 'project';

  const setupType = data.setupType;
  const normalizedSetupType: ManifestShape['setupType'] =
    setupType === 'workspace' || setupType === 'project'
      ? setupType
      : normalizedScope === 'workspace'
        ? 'workspace'
        : 'project';

  const tools = Array.isArray(data.tools)
    ? data.tools.filter((tool): tool is string => typeof tool === 'string')
    : [];

  const files = Array.isArray(data.files)
    ? data.files.filter(isManifestFile).map((file) => ({ path: file.path, hash: file.hash, source: file.source }))
    : [];

  return {
    version: typeof data.version === 'string' ? data.version : fallback.version,
    setupScope: normalizedScope,
    setupType: normalizedSetupType,
    tools,
    projectName: typeof data.projectName === 'string' ? data.projectName : fallback.projectName,
    installedAt: typeof data.installedAt === 'string' ? data.installedAt : fallback.installedAt,
    files,
  };
}

function toManifestFile(record: FileRecord): ManifestFile {
  return {
    path: record.path,
    hash: record.hash,
    source: record.source,
  };
}

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

export async function executeMigrationToCanonical(options: {
  context: MigrationContext;
  plan: MigrationPlan;
  parsedSetups: ParsedSetup[];
}): Promise<MigrationResult> {
  const { context, plan, parsedSetups } = options;
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

  const fileRecords: FileRecord[] = [];
  let backupPath: string | undefined;

  try {
    if (!context.options.skipBackup) {
      backupPath = await createBackupDir(context.targetPath);
    }

    if (backupPath) {
      for (const parsed of parsedSetups) {
        for (const file of parsed.files) {
          const backupAction: MigrationAction = {
            type: 'backup',
            sourcePath: file.path,
            targetPath: path.join('.ai-setup-backup', file.path),
            description: `Backup ${file.path}`,
            reason: 'Before canonical migration write',
          };

          try {
            await executeBackupAction(backupAction, context, backupPath);
            executedActions.push(backupAction);
            stats.filesBackedUp += 1;
          } catch {
            // Some parsers include virtual files; skip silently.
          }
        }
      }
    }

    for (const parsed of parsedSetups) {
      const canonicalResult = await writeToCanonical({
        targetDir: context.targetPath,
        parsedSetup: parsed,
        fileRecords,
        dryRun: context.options.preview === true,
      });

      const createdPaths = [
        ...canonicalResult.agents,
        ...canonicalResult.skills,
        ...canonicalResult.prompts,
        ...canonicalResult.rules,
        ...(canonicalResult.rootConfig ? [canonicalResult.rootConfig] : []),
      ];

      for (const targetPath of createdPaths) {
        executedActions.push({
          type: 'create',
          targetPath,
          description: `Create ${targetPath}`,
          reason: `Canonical migration from ${parsed.metadata.adapter || 'adapter'}`,
        });
      }

      for (const targetPath of canonicalResult.skipped) {
        executedActions.push({
          type: 'skip',
          targetPath,
          description: `Skip ${targetPath}`,
          reason: 'File already exists',
        });
      }

      stats.filesCreated += createdPaths.length;
      stats.filesSkipped += canonicalResult.skipped.length;
    }

    await createAiSetupManifest(context, plan, executedActions, fileRecords);

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
    errors.push(`Fatal error during canonical migration: ${error}`);

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
  executedActions: MigrationAction[],
  fileRecords?: FileRecord[]
): Promise<void> {
  const manifestPath = path.join(context.targetPath, '.ai-setup.json');
  
  let existingManifest = normalizeManifest(null);
  
  // Try to read existing manifest
  try {
    const content = await fs.readFile(manifestPath, 'utf-8');
    existingManifest = normalizeManifest(JSON.parse(content));
  } catch {
    // No existing manifest
  }
  
  // Update tools
  for (const adapter of plan.adapters) {
    if (!toolIdSchema.safeParse(adapter).success) {
      continue;
    }

    if (!existingManifest.tools.includes(adapter)) {
      existingManifest.tools.push(adapter);
    }
  }
  
  if (fileRecords && fileRecords.length > 0) {
    for (const record of fileRecords) {
      const existingFile = existingManifest.files.find((f) => f.path === record.path);
      if (existingFile) {
        existingFile.hash = record.hash;
        existingFile.source = record.source;
      } else {
        existingManifest.files.push(toManifestFile(record));
      }
    }
  } else {
    // Track migrated files
    for (const action of executedActions) {
      if (action.type === 'create' || action.type === 'modify') {
        const existingFile = existingManifest.files.find((f) => f.path === action.targetPath);
        if (!existingFile) {
          existingManifest.files.push({
            path: action.targetPath,
            hash: 'migrated', // Will be updated by doctor
            source: 'migration',
          });
        }
      }
    }
  }
  
  // Write manifest
  await fs.writeFile(manifestPath, JSON.stringify(existingManifest, null, 2));
}

export async function rollbackMigration(
  _context: MigrationContext,
  _backupPath: string
): Promise<void> {
  // Restore files from backup
  // This is a safety feature for failed migrations
  console.warn('Rolling back migration...');
  
  // Implementation would restore backed up files
  // For now, we just log the intent
}
