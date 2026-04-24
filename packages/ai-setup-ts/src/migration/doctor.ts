/**
 * Migration Doctor
 *
 * Checks for drift between current setup and clean ai-setup state.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { glob } from 'glob';
import { storeDataSchema, type TrackedFile } from '../store/schema.js';
import { fileHash } from '../utils/files.js';
import type { DriftCheckResult, DriftItem, ModifiedFile } from './types.js';

interface ManifestLike {
  files?: TrackedFile[];
}

function parseManifest(content: string): ManifestLike {
  const raw = JSON.parse(content) as unknown;
  const parsedStore = storeDataSchema.safeParse(raw);
  if (parsedStore.success) {
    return { files: parsedStore.data.files };
  }

  if (raw && typeof raw === 'object') {
    const record = raw as Record<string, unknown>;
    const files = Array.isArray(record.files) ? record.files : [];
    const normalized = files
      .filter((item): item is { path: string; hash: string; source: string } => {
        return Boolean(
          item &&
            typeof item === 'object' &&
            typeof (item as Record<string, unknown>).path === 'string' &&
            typeof (item as Record<string, unknown>).hash === 'string' &&
            typeof (item as Record<string, unknown>).source === 'string',
        );
      })
      .map((item) => ({
        path: item.path,
        hash: item.hash,
        source: item.source,
        owner: 'library' as const,
      }));

    return { files: normalized };
  }

  return { files: [] };
}

export async function checkDrift(
  targetPath: string,
  _verbose?: boolean
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
    'specs/**/*',
    '.opencode/**/*',
    '.claude/**/*',
    '.gemini/**/*',
    '.github/copilot-instructions.md',
    '.github/prompts/**/*',
    '.github/instructions/**/*',
  ];

  // Load current .ai-setup.json
  let currentConfig: ManifestLike | null = null;
  try {
    const configPath = path.join(targetPath, '.ai-setup.json');
    const configContent = await fs.readFile(configPath, 'utf-8');
    currentConfig = parseManifest(configContent);
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
        const currentHash = fileHash(filePath);

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
        const isTracked = currentConfig.files?.some((f) => f.path === file);
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
