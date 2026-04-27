/**
 * Migration Plan Generator
 * 
 * Generates a comprehensive plan for migrating existing AI setups to ai-setup format.
 * Includes conflict detection and resolution strategies.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import type { BaseParser } from './parsers/base-parser.js';
import type {
  AdapterConflict,
  DetectionResult,
  MergeConflict,
  MigrationAction,
  MigrationContext,
  MigrationPlan,
} from './types.js';

interface PendingAction {
  adapterId: string;
  action: MigrationAction;
}

export async function generateMigrationPlan(
  context: MigrationContext,
  detections: DetectionResult[],
  parsers: BaseParser[]
): Promise<MigrationPlan> {
  const allActions: PendingAction[] = [];
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
        allActions.push({
          adapterId: detection.adapterId,
          action: {
            type: 'backup',
            sourcePath: file.path,
            targetPath: path.join('.ai-setup-backup', file.path),
            description: `Backup ${file.path} (no parser available)`,
            reason: `No parser found for ${detection.adapterId}`,
          },
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
        allActions.push({
          adapterId: detection.adapterId,
          action: {
            type: 'create',
            targetPath: newFile,
            description: `Create ${newFile}`,
            reason: 'New file from merged setup',
          },
        });
      }

      for (const modFile of mergeResult.modifiedFiles) {
        // Check if this file already exists
        if (existingAiSetupFiles.includes(modFile)) {
          allActions.push({
            adapterId: detection.adapterId,
            action: {
              type: 'backup',
              sourcePath: modFile,
              targetPath: path.join('.ai-setup-backup', modFile),
              description: `Backup existing ${modFile}`,
              reason: 'Will be modified by migration',
            },
          });
        }
        
        allActions.push({
          adapterId: detection.adapterId,
          action: {
            type: 'modify',
            targetPath: modFile,
            description: `Update ${modFile}`,
            reason: 'Merged with existing setup',
          },
        });
      }

      // Add conflicts
      conflicts.push(...mergeResult.conflicts);
    } else {
      // Cannot merge - backup and create fresh
      for (const file of detection.files) {
        allActions.push({
          adapterId: detection.adapterId,
          action: {
            type: 'backup',
            sourcePath: file.path,
            targetPath: path.join('.ai-setup-backup', file.path),
            description: `Backup ${file.path}`,
            reason: 'Creating fresh ai-setup structure',
          },
        });
      }
    }
  }

  // Deduplicate actions by targetPath, tracking adapter conflicts
  const { dedupedActions, adapterConflicts } = deduplicateActions(allActions);

  // Calculate canProceed
  const unresolvedConflicts = conflicts.filter(c => !c.resolved)
  const canProceed = unresolvedConflicts.length === 0 || context.options.mergeStrategy !== 'smart'

  return {
    sourcePath: context.sourcePath,
    targetPath: context.targetPath,
    adapters,
    actions: dedupedActions,
    conflicts,
    adapterConflicts,
    estimatedFiles: dedupedActions.filter(a => a.type !== 'skip').length,
    estimatedConflicts: unresolvedConflicts.length,
    canProceed,
  };
}

function deduplicateActions(allActions: PendingAction[]): {
  dedupedActions: MigrationAction[];
  adapterConflicts: AdapterConflict[];
} {
  // Group by targetPath
  const byTargetPath = new Map<string, PendingAction[]>();
  for (const pa of allActions) {
    const key = pa.action.targetPath;
    const existing = byTargetPath.get(key);
    if (existing) {
      existing.push(pa);
    } else {
      byTargetPath.set(key, [pa]);
    }
  }

  const dedupedActions: MigrationAction[] = [];
  const adapterConflicts: AdapterConflict[] = [];

  for (const [targetPath, entries] of byTargetPath) {
    const first = entries[0];
    if (!first) {
      continue;
    }

    if (entries.length === 1) {
      // Single adapter — no conflict, just tag and keep
      dedupedActions.push({ ...first.action, adapterId: first.adapterId });
      continue;
    }

    // Multiple entries for the same targetPath — detect if from different adapters
    const uniqueAdapters = new Set(entries.map(e => e.adapterId));

    if (uniqueAdapters.size === 1) {
      // Same adapter generated multiple actions for the same path — keep first
      dedupedActions.push({ ...first.action, adapterId: first.adapterId });
      continue;
    }

    // Conflict: different adapters target the same file
    adapterConflicts.push({
      targetPath,
      adapters: entries.map(e => e.adapterId),
      actions: entries.map(e => ({
        type: e.action.type,
        description: e.action.description,
        reason: e.action.reason,
        adapterId: e.adapterId,
      })),
    });

    // Keep first adapter's action (first wins)
    dedupedActions.push({ ...first.action, adapterId: first.adapterId });
  }

  return { dedupedActions, adapterConflicts };
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

function adapterLabel(adapterId: string): string {
  const labels: Record<string, string> = {
    opencode: 'OpenCode',
    'claude-code': 'Claude Code',
    gemini: 'Gemini CLI',
    copilot: 'GitHub Copilot',
  };
  return labels[adapterId] || adapterId;
}

export function formatPlan(plan: MigrationPlan): string {
  const lines: string[] = [];
  
  lines.push('Migration plan');
  lines.push('==============');
  lines.push('');
  
  lines.push(`Source: ${plan.sourcePath}`);
  lines.push(`Target: ${plan.targetPath}`);
  lines.push(`Detected adapters: ${plan.adapters.map(adapterLabel).join(', ')}`);
  lines.push('');
  
  // Group actions by type
  const creates = plan.actions.filter(a => a.type === 'create');
  const modifies = plan.actions.filter(a => a.type === 'modify');
  const backups = plan.actions.filter(a => a.type === 'backup');
  const skips = plan.actions.filter(a => a.type === 'skip');
  const unresolved = plan.conflicts.filter(c => !c.resolved);

  lines.push('Summary:');
  lines.push(`  Create: ${creates.length}`);
  lines.push(`  Modify: ${modifies.length}`);
  lines.push(`  Backup: ${backups.length}`);
  lines.push(`  Skip: ${skips.length}`);
  lines.push(`  Unresolved conflicts: ${unresolved.length}`);
  lines.push('');
  
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
  
  // Show adapter-level conflicts (multi-adapter targeting same file)
  if (plan.adapterConflicts && plan.adapterConflicts.length > 0) {
    lines.push(`⚠️  ${plan.adapterConflicts.length} multi-adapter conflict(s):`);
    for (const ac of plan.adapterConflicts) {
      const actionDescs = ac.actions.map(a => {
        const label = adapterLabel(a.adapterId || '');
        return `${label} wants to ${a.type}`;
      });
      lines.push(`   ! ${ac.targetPath} — ${actionDescs.join(', ')}`);
    }
    lines.push('');
  }

  if (unresolved.length > 0) {
    lines.push(`⚠️  ${unresolved.length} unresolved merge conflict(s):`);
    for (const conflict of unresolved) {
      lines.push(`   ! ${conflict.file} (lines ${conflict.lineStart}-${conflict.lineEnd})`);
    }
    lines.push('');
    lines.push('Suggested next steps:');
    lines.push('  • Re-run with --interactive to resolve conflicts one by one');
    lines.push('  • Or choose --strategy preserve/replace for a simpler merge path');
    lines.push('');
  } else if ((!plan.adapterConflicts || plan.adapterConflicts.length === 0) && unresolved.length === 0) {
    lines.push('No conflicts detected.');
    lines.push('');
  }
  
  lines.push(plan.canProceed 
    ? 'Status: ready to apply ✅'
    : 'Status: blocked until conflicts are resolved ❌');
  
  return lines.join('\n');
}
