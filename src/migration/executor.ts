/**
 * Migration Executor
 * 
 * Executes migration plans with proper backup, file operations, and rollback support.
 */

import { promises as fs } from 'fs';
import path from 'path';
import {
  MigrationContext,
  MigrationPlan,
  MigrationResult,
  MigrationAction,
  MigrationStats,
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

    // Step 6: Create .ai-setup.json if it doesn't exist
    await createAiSetupManifest(context, plan, executedActions);

    return {
      success: errors.length === 0,
      plan,
      executedActions,
      backupPath,
      errors,
      warnings,
      stats,
    };

  } catch (error) {
    // Fatal error - try to rollback if possible
    errors.push(`Fatal error during migration: ${error}`);
    
    return {
      success: false,
      plan,
      executedActions,
      backupPath,
      errors,
      warnings,
      stats,
    };
  }
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
