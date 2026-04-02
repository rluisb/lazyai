import path from 'node:path';
import { ensureDir, fileExists, fileHash, writeFile } from '../utils/files.js';
import type { FileRecord } from '../types.js';
import type { ParsedSetup } from './types.js';

export interface CanonicalWriterOptions {
  targetDir: string;
  parsedSetup: ParsedSetup;
  fileRecords: FileRecord[];
  dryRun?: boolean;
}

export interface CanonicalWriteResult {
  agents: string[];
  skills: string[];
  prompts: string[];
  rules: string[];
  rootConfig: string | null;
  skipped: string[];
}

function normalizeName(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'item';
}

function writeCanonicalFile(args: {
  destination: string;
  content: string;
  targetDir: string;
  source: string;
  fileRecords: FileRecord[];
  dryRun: boolean;
  skipped: string[];
}): boolean {
  if (fileExists(args.destination)) {
    args.skipped.push(path.relative(args.targetDir, args.destination));
    return false;
  }

  if (args.dryRun) {
    return true;
  }

  writeFile(args.destination, args.content);
  args.fileRecords.push({
    path: path.relative(args.targetDir, args.destination),
    hash: fileHash(args.destination),
    source: args.source,
  });

  return true;
}

export async function writeToCanonical(opts: CanonicalWriterOptions): Promise<CanonicalWriteResult> {
  const aiDir = path.join(opts.targetDir, '.ai');
  const dryRun = opts.dryRun === true;
  const result: CanonicalWriteResult = {
    agents: [],
    skills: [],
    prompts: [],
    rules: [],
    rootConfig: null,
    skipped: [],
  };

  if (!dryRun) {
    ensureDir(aiDir);
  }

  const agentsDir = path.join(aiDir, 'agents');
  if (!dryRun) ensureDir(agentsDir);
  for (const agent of opts.parsedSetup.agents ?? []) {
    const fileName = `${normalizeName(agent.id || agent.name)}.md`;
    const destination = path.join(agentsDir, fileName);
    const wrote = writeCanonicalFile({
      destination,
      content: agent.content,
      targetDir: opts.targetDir,
      source: `migrated:${agent.sourcePath}`,
      fileRecords: opts.fileRecords,
      dryRun,
      skipped: result.skipped,
    });
    if (wrote) {
      result.agents.push(path.relative(opts.targetDir, destination));
    }
  }

  const skillsDir = path.join(aiDir, 'skills');
  if (!dryRun) ensureDir(skillsDir);
  for (const command of opts.parsedSetup.commands ?? []) {
    const fileName = `${normalizeName(command.id || command.name)}.md`;
    const destination = path.join(skillsDir, fileName);
    const wrote = writeCanonicalFile({
      destination,
      content: command.content,
      targetDir: opts.targetDir,
      source: `migrated:${command.sourcePath}`,
      fileRecords: opts.fileRecords,
      dryRun,
      skipped: result.skipped,
    });
    if (wrote) {
      result.skills.push(path.relative(opts.targetDir, destination));
    }
  }

  const promptsDir = path.join(aiDir, 'prompts');
  if (!dryRun) ensureDir(promptsDir);
  for (const template of opts.parsedSetup.templates ?? []) {
    const fileName = `${normalizeName(template.id || template.name)}.md`;
    const destination = path.join(promptsDir, fileName);
    const wrote = writeCanonicalFile({
      destination,
      content: template.content,
      targetDir: opts.targetDir,
      source: `migrated:${template.sourcePath}`,
      fileRecords: opts.fileRecords,
      dryRun,
      skipped: result.skipped,
    });
    if (wrote) {
      result.prompts.push(path.relative(opts.targetDir, destination));
    }
  }

  const rulesDir = path.join(aiDir, 'rules');
  if (!dryRun) ensureDir(rulesDir);
  for (const rule of opts.parsedSetup.rules ?? []) {
    const fileName = `${normalizeName(rule.id || rule.category || 'rule')}.md`;
    const destination = path.join(rulesDir, fileName);
    const wrote = writeCanonicalFile({
      destination,
      content: rule.content,
      targetDir: opts.targetDir,
      source: `migrated:${rule.sourcePath}`,
      fileRecords: opts.fileRecords,
      dryRun,
      skipped: result.skipped,
    });
    if (wrote) {
      result.rules.push(path.relative(opts.targetDir, destination));
    }
  }

  if ((opts.parsedSetup.customSections?.length ?? 0) > 0) {
    const destination = path.join(aiDir, 'constitution', `${normalizeName(opts.parsedSetup.metadata?.adapter as string || opts.parsedSetup.projectName || 'migration')}.md`);
    const content = opts.parsedSetup.customSections
      .map((section) => `## ${section.title}\n\n${section.content}`)
      .join('\n\n');
    const wrote = writeCanonicalFile({
      destination,
      content,
      targetDir: opts.targetDir,
      source: `migrated:${opts.parsedSetup.metadata?.adapter ?? 'unknown'}`,
      fileRecords: opts.fileRecords,
      dryRun,
      skipped: result.skipped,
    });
    if (wrote) {
      result.rootConfig = path.relative(opts.targetDir, destination);
    }
  }

  return result;
}
