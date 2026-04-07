/**
 * Detector Module - Scans directories for existing AI setups
 * 
 * Detects presence of various AI coding assistant configurations
 * and determines which parser(s) should handle them.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { glob } from 'glob';
import type {
  DetectedFile,
  DetectionResult,
  MigrationContext,
} from './types.js';

/**
 * Detection patterns for different AI tools
 */
export const DETECTION_PATTERNS: Record<string, string[]> = {
  opencode: [
    '.opencode/**/*',
    'AGENTS.md',
    '.opencode/agents/*.md',
    '.opencode/commands/*.md',
    '.opencode/templates/*.md',
  ],
  'claude-code': [
    '.claude/**/*',
    'CLAUDE.md',
    '.claude/*.md',
    '.claude/commands/*.md',
    '.claude/rules/*.md',
  ],
  pi: [
    '.pi/**/*',
    'AGENTS.md',
    '.pi/settings.json',
    '.pi/skills/*/SKILL.md',
    '.pi/prompts/*.md',
  ],
  gemini: [
    '.gemini/**/*',
    'GEMINI.md',
    '.gemini/agents/*.md',
    '.gemini/skills/*.md',
  ],
  copilot: [
    '.github/copilot-instructions.md',
    '.github/prompts/*.md',
    '.github/instructions/*.md',
  ],
};

/**
 * File type patterns for categorization
 */
const FILE_TYPE_PATTERNS: Record<string, RegExp> = {
  config: /\.(json|yaml|yml|toml)$/,
  agent: /agent|\.ai-|copilot-?instruction/i,
  rule: /rule|constraint|guideline/i,
  template: /template|prompt/i,
  command: /command|skill/i,
};

/**
 * Detect AI setups in a directory
 */
export async function detectExistingSetup(
  context: MigrationContext
): Promise<DetectionResult[]> {
  const results: DetectionResult[] = [];
  const sourcePath = context.sourcePath;

  // Check if source exists
  try {
    await fs.access(sourcePath);
  } catch {
    return [];
  }

  // Check each adapter pattern
  for (const [adapterId, patterns] of Object.entries(DETECTION_PATTERNS)) {
    const result = await detectAdapter(sourcePath, adapterId, patterns);
    if (result.detected) {
      results.push(result);
    }
  }

  // Sort by confidence (highest first)
  return results.sort((a, b) => b.confidence - a.confidence);
}

/**
 * Detect a specific adapter
 */
async function detectAdapter(
  sourcePath: string,
  adapterId: string,
  patterns: string[]
): Promise<DetectionResult> {
  const detectedFiles: DetectedFile[] = [];
  let totalFiles = 0;
  let matchedFiles = 0;

  const adapterNames: Record<string, string> = {
    opencode: 'OpenCode',
    'claude-code': 'Claude Code',
    pi: 'Pi',
    gemini: 'Gemini CLI',
    copilot: 'GitHub Copilot',
  };

  for (const pattern of patterns) {
    try {
      const files = await glob(pattern, {
        cwd: sourcePath,
        absolute: false,
        dot: true,
      });

      totalFiles += files.length;

      for (const file of files) {
        const fullPath = path.join(sourcePath, file);
        
        try {
          const stat = await fs.stat(fullPath);
          if (stat.isFile()) {
            matchedFiles++;
            detectedFiles.push({
              path: file,
              type: categorizeFile(file),
              priority: calculatePriority(file, adapterId),
            });
          }
        } catch {
          // File might not exist or be accessible
        }
      }
    } catch {
      // Pattern might not match anything
    }
  }

  // Calculate confidence based on file matches
  const confidence = calculateConfidence(detectedFiles, adapterId);

  return {
    detected: detectedFiles.length > 0,
    confidence,
    adapterId,
    adapterName: adapterNames[adapterId] || adapterId,
    files: detectedFiles.sort((a, b) => b.priority - a.priority),
    metadata: {
      totalFiles,
      matchedFiles,
      uniquePatterns: patterns.length,
    },
  };
}

/**
 * Categorize a file based on its path and name
 */
function categorizeFile(filePath: string): DetectedFile['type'] {
  const normalized = filePath.toLowerCase();

  for (const [type, pattern] of Object.entries(FILE_TYPE_PATTERNS)) {
    if (pattern.test(normalized)) {
      return type as DetectedFile['type'];
    }
  }

  return 'other';
}

/**
 * Calculate priority for a detected file
 */
function calculatePriority(filePath: string, _adapterId: string): number {
  let priority = 0;
  const normalized = filePath.toLowerCase();

  // Root config files are highest priority
  if (/^AGENTS\.md$|^CLAUDE\.md$|^GEMINI\.md$/i.test(filePath)) {
    priority += 100;
  }

  // Main config files
  if (/copilot-instructions\.md$/i.test(filePath)) {
    priority += 90;
  }

  // Agent definitions
  if (normalized.includes('agent')) {
    priority += 50;
  }

  // Rules and standards
  if (normalized.includes('rule') || normalized.includes('standard')) {
    priority += 40;
  }

  // Commands/skills
  if (normalized.includes('command') || normalized.includes('skill')) {
    priority += 30;
  }

  // Templates
  if (normalized.includes('template')) {
    priority += 20;
  }

  return priority;
}

/**
 * Calculate detection confidence
 */
function calculateConfidence(files: DetectedFile[], _adapterId: string): number {
  if (files.length === 0) return 0;

  // Base confidence on number of files
  let confidence = Math.min(files.length / 5, 0.8);

  // Boost confidence if we have high-priority files
  const hasHighPriority = files.some(f => f.priority >= 50);
  if (hasHighPriority) {
    confidence += 0.15;
  }

  // Boost if we have root config
  const hasRootConfig = files.some(f => 
    f.path === 'AGENTS.md' || 
    f.path === 'CLAUDE.md' ||
    f.path === 'GEMINI.md' ||
    f.path.includes('copilot-instructions')
  );
  if (hasRootConfig) {
    confidence += 0.1;
  }

  return Math.min(confidence, 1.0);
}

/**
 * Check if a specific adapter exists in a directory
 */
export async function hasAdapter(
  sourcePath: string,
  adapterId: string
): Promise<boolean> {
  const patterns = DETECTION_PATTERNS[adapterId];
  if (!patterns) return false;

  const result = await detectAdapter(sourcePath, adapterId, patterns);
  return result.detected && result.confidence > 0.3;
}

/**
 * Get all detected adapters in a directory
 */
export async function getDetectedAdapters(
  sourcePath: string
): Promise<string[]> {
  const results: string[] = [];

  for (const adapterId of Object.keys(DETECTION_PATTERNS)) {
    if (await hasAdapter(sourcePath, adapterId)) {
      results.push(adapterId);
    }
  }

  return results;
}
