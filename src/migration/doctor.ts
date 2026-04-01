/**
 * Migration Doctor
 *
 * Checks for drift between current setup and clean ai-setup state.
 */

import { promises as fs } from 'fs';
import path from 'path';
import { glob } from 'glob';
import { DriftCheckResult, DriftItem, ModifiedFile } from './types.js';

export async function checkDrift(
  targetPath: string,
  verbose?: boolean
): Promise<DriftCheckResult> {
  const drifts: DriftItem[] = [];
  const missingFiles: string[] = [];
  const extraFiles: string[] = [];
  const modifiedFiles: ModifiedFile[] = [];

  const aiSetupPatterns = [
    'AGENTS.md',
    'CLAUDE.md',
    'GEMINI.md',
    'KNOWLEDGE_MAP.md',
    'docs/**/*',
    '.opencode/**/*',
    '.claude/**/*',
    '.pi/**/*',
    '.gemini/**/*',
    '.github/copilot-instructions.md',
    '.github/prompts/**/*',
    '.github/instructions/**/*',
  ];

  // Load current .ai-setup.json
  let currentConfig: any = null;
  try {
    const configPath = path.join(targetPath, '.ai-setup.json');
    const configContent = await fs.readFile(configPath, 'utf-8');
    currentConfig = JSON.parse(configContent);
  } catch {
    // No config file - treat discovered AI setup files as extra drift
    for (const pattern of aiSetupPatterns) {
      try {
        const files = await glob(pattern, {
          cwd: targetPath,
          absolute: false,
        });

        for (const file of files) {
          if (!extraFiles.includes(file)) {
            extraFiles.push(file);
          }
        }
      } catch {
        // Pattern might not match anything
      }
    }

    return {
      clean: false,
      drifts: [
        {
          file: '.ai-setup.json',
          type: 'missing',
        },
        ...extraFiles.map((file) => ({ file, type: 'extra' as const })),
      ],
      missingFiles: ['.ai-setup.json'],
      extraFiles,
      modifiedFiles: [],
    };
  }

  // Check expected files from config
  if (currentConfig.files) {
    for (const file of currentConfig.files) {
      const filePath = path.join(targetPath, file.path);
      try {
        await fs.access(filePath);

        // File exists - check if modified
        const content = await fs.readFile(filePath, 'utf-8');
        // Simple hash check (in real implementation, use proper hash)
        const currentHash = Buffer.from(content).toString('base64').slice(0, 16);

        if (currentHash !== file.hash && file.hash !== 'migrated') {
          modifiedFiles.push({
            path: file.path,
            expectedHash: file.hash,
            actualHash: currentHash,
            difference: 'Content differs from expected',
          });
        }
      } catch {
        // File missing
        missingFiles.push(file.path);
      }
    }
  }

  // Find extra files (not tracked in .ai-setup.json)
  for (const pattern of aiSetupPatterns) {
    try {
      const files = await glob(pattern, {
        cwd: targetPath,
        absolute: false,
      });

      for (const file of files) {
        const isTracked = currentConfig.files?.some((f: any) => f.path === file);
        if (!isTracked && !extraFiles.includes(file)) {
          extraFiles.push(file);
        }
      }
    } catch {
      // Pattern might not match anything
    }
  }

  // Build drifts list
  for (const file of missingFiles) {
    drifts.push({ file, type: 'missing' });
  }
  for (const file of extraFiles) {
    drifts.push({ file, type: 'extra' });
  }
  for (const file of modifiedFiles) {
    drifts.push({
      file: file.path,
      type: 'modified',
      expectedHash: file.expectedHash,
      actualHash: file.actualHash,
      diff: file.difference,
    });
  }

  return {
    clean: drifts.length === 0,
    drifts,
    missingFiles,
    extraFiles,
    modifiedFiles,
  };
}
