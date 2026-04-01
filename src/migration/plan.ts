/**
 * Migration Plan Generator
 * 
 * Generates a comprehensive plan for migrating existing AI setups to ai-setup format.
 * Includes conflict detection and resolution strategies.
 */

import { promises as fs } from 'fs';
import path from 'path';
import {
  MigrationContext,
  MigrationPlan,
  MigrationAction,
  MergeConflict,
  DetectionResult,
  ParsedSetup,
} from './types.js';
import { BaseParser } from './parsers/base-parser.js';

export async function generateMigrationPlan(
  context: MigrationContext,
  detections: DetectionResult[],
  parsers: BaseParser[]
): Promise<MigrationPlan> {
  const actions: MigrationAction[] = [];
  const conflicts: MergeConflict[] = [];
  const adapters: string[] = [];

  // Get existing ai-setup files if any
  const existingAiSetupFiles = await getExistingAiSetupFiles(context.targetPath);

  // Process each detected adapter
  for (const detection of detections) {
    adapters.push(detection.adapterId);
    
    // Find matching parser
    const parser = parsers.find(p => p.id === detection.adapterId);
    
    if (!parser) {
      // No parser available - just backup files
      for (const file of detection.files) {
        actions.push({
          type: 'backup',
          sourcePath: file.path,
          targetPath: path.join('.ai-setup-backup', file.path),
          description: `Backup ${file.path} (no parser available)`,
          reason: `No parser found for ${detection.adapterId}`,
        });
      }
      continue;
    }

    // Parse the existing setup
    const parsed = await parser.parse(context);

    // Check if merge is possible
    if (parser.canMerge(parsed)) {
      const mergeResult = await parser.merge(
        parsed,
        context.options.mergeStrategy,
        context.options
      );

      // Add merge result actions
      for (const newFile of mergeResult.newFiles) {
        actions.push({
          type: 'create',
          targetPath: newFile,
          description: `Create ${newFile}`,
          reason: 'New file from merged setup',
        });
      }

      for (const modFile of mergeResult.modifiedFiles) {
        // Check if this file already exists
        if (existingAiSetupFiles.includes(modFile)) {
          actions.push({
            type: 'backup',
            sourcePath: modFile,
            targetPath: path.join('.ai-setup-backup', modFile),
            description: `Backup existing ${modFile}`,
            reason: 'Will be modified by migration',
          });
        }
        
        actions.push({
          type: 'modify',
          targetPath: modFile,
          description: `Update ${modFile}`,
          reason: 'Merged with existing setup',
        });
      }

      // Add conflicts
      conflicts.push(...mergeResult.conflicts);
    } else {
      // Cannot merge - backup and create fresh
      for (const file of detection.files) {
        actions.push({
          type: 'backup',
          sourcePath: file.path,
          targetPath: path.join('.ai-setup-backup', file.path),
          description: `Backup ${file.path}`,
          reason: 'Creating fresh ai-setup structure',
        });
      }
    }
  }

  // Calculate canProceed
  const unresolvedConflicts = conflicts.filter(c => !c.resolved);
  const canProceed = unresolvedConflicts.length === 0 || context.options.mergeStrategy !== 'smart';

  return {
    sourcePath: context.sourcePath,
    targetPath: context.targetPath,
    adapters,
    actions,
    conflicts,
    estimatedFiles: actions.filter(a => a.type !== 'skip').length,
    estimatedConflicts: unresolvedConflicts.length,
    canProceed,
  };
}

async function getExistingAiSetupFiles(targetPath: string): Promise<string[]> {
  const files: string[] = [];
  
  try {
    // Check for .ai-setup.json
    const manifestPath = path.join(targetPath, '.ai-setup.json');
    await fs.access(manifestPath);
    
    // If manifest exists, read it to get tracked files
    const manifest = JSON.parse(await fs.readFile(manifestPath, 'utf-8'));
    if (manifest.files) {
      for (const file of manifest.files) {
        files.push(file.path);
      }
    }
  } catch {
    // No manifest or can't read it
  }

  // Also check for common ai-setup files
  const commonFiles = [
    'AGENTS.md',
    'CLAUDE.md',
    'GEMINI.md',
    '.ai-setup.json',
    'KNOWLEDGE_MAP.md',
  ];

  for (const file of commonFiles) {
    try {
      await fs.access(path.join(targetPath, file));
      if (!files.includes(file)) {
        files.push(file);
      }
    } catch {
      // File doesn't exist
    }
  }

  return files;
}

export function formatPlan(plan: MigrationPlan): string {
  const lines: string[] = [];
  
  lines.push('╔═══════════════════════════════════════════════════════════╗');
  lines.push('║              MIGRATION PLAN                               ║');
  lines.push('╚═══════════════════════════════════════════════════════════╝');
  lines.push('');
  
  lines.push(`Source: ${plan.sourcePath}`);
  lines.push(`Target: ${plan.targetPath}`);
  lines.push(`Adapters: ${plan.adapters.join(', ')}`);
  lines.push('');
  
  // Group actions by type
  const creates = plan.actions.filter(a => a.type === 'create');
  const modifies = plan.actions.filter(a => a.type === 'modify');
  const backups = plan.actions.filter(a => a.type === 'backup');
  const skips = plan.actions.filter(a => a.type === 'skip');
  
  if (creates.length > 0) {
    lines.push(`📁 Create ${creates.length} new file(s):`);
    for (const action of creates) {
      lines.push(`   + ${action.targetPath}`);
    }
    lines.push('');
  }
  
  if (modifies.length > 0) {
    lines.push(`📝 Modify ${modifies.length} file(s):`);
    for (const action of modifies) {
      lines.push(`   ~ ${action.targetPath}`);
    }
    lines.push('');
  }
  
  if (backups.length > 0) {
    lines.push(`💾 Backup ${backups.length} file(s):`);
    for (const action of backups) {
      lines.push(`   ← ${action.sourcePath} → ${action.targetPath}`);
    }
    lines.push('');
  }
  
  if (plan.conflicts.length > 0) {
    const unresolved = plan.conflicts.filter(c => !c.resolved);
    lines.push(`⚠️  ${unresolved.length} unresolved conflict(s):`);
    for (const conflict of unresolved) {
      lines.push(`   ! ${conflict.file} (lines ${conflict.lineStart}-${conflict.lineEnd})`);
    }
    lines.push('');
  }
  
  lines.push(plan.canProceed 
    ? '✅ Migration can proceed'
    : '❌ Migration blocked - resolve conflicts first');
  
  return lines.join('\n');
}
