/**
 * Pi Parser
 * 
 * Detects, parses, and merges Pi AI setups.
 * Pi uses AGENTS.md, .pi/settings.json, AgentSkills-style skills,
 * and prompt templates under .pi/prompts/.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { glob } from 'glob';
import { resolveLibraryDir } from '../../utils/files.js';
import { diff3 } from '../diff/diff3.js';
import type { 
  AgentDefinition,
  CommandDefinition,CustomSection, 
  DetectedFile,
  DetectionResult,MergeConflict, 
  MergeResult,
  MergeStrategy,
  MigrationContext,
  MigrationOptions,
  ParsedFile,
  ParsedSetup,
  RuleDefinition,
  TemplateDefinition,} from '../types.js';
import { BaseParser } from './base-parser.js';

const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)));

export class PiParser extends BaseParser {
  readonly id = 'pi';
  readonly name = 'Pi';
  readonly description = 'Parser for Pi AI assistant configuration';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.pi/**/*',
    'AGENTS.md',
    '.pi/settings.json',
    '.pi/skills/*/SKILL.md',
    '.pi/prompts/*.md',
  ];

  async detect(context: MigrationContext): Promise<DetectionResult> {
    const files: DetectedFile[] = [];
    let detected = false;
    let confidence = 0;

    // Check for .pi directory
    try {
      const piDir = path.join(context.sourcePath, '.pi');
      await fs.access(piDir);
      
      const piFiles = await glob('.pi/**/*', {
        cwd: context.sourcePath,
        absolute: false,
        dot: true,
      });

      for (const file of piFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const stat = await fs.stat(fullPath);
          if (stat.isFile()) {
            files.push({
              path: file,
              type: this.categorizeFile(file),
              priority: this.calculatePriority(file),
            });
            detected = true;
            confidence += 0.1;
          }
        } catch {
          // Skip
        }
      }
    } catch {
      // Directory doesn't exist
    }

    // Check for AGENTS.md (Pi uses this as root)
    if (detected) {
      try {
        const agentsPath = path.join(context.sourcePath, 'AGENTS.md');
        await fs.access(agentsPath);
        files.push({
          path: 'AGENTS.md',
          type: 'config',
          priority: 100,
        });
        confidence += 0.3;
      } catch {
        // Not found
      }
    }

    confidence = Math.min(confidence, 1.0);

    return {
      detected,
      confidence,
      adapterId: this.id,
      adapterName: this.name,
      files: files.sort((a, b) => b.priority - a.priority),
    };
  }

  async parse(context: MigrationContext): Promise<ParsedSetup> {
    const agents: AgentDefinition[] = [];
    const rules: RuleDefinition[] = [];
    const commands: CommandDefinition[] = [];
    const templates: TemplateDefinition[] = [];
    const customSections: CustomSection[] = [];
    const files: ParsedFile[] = [];
    let projectName: string | undefined;
    let description: string | undefined;

    // Parse AGENTS.md
    try {
      const agentsPath = path.join(context.sourcePath, 'AGENTS.md');
      const content = await fs.readFile(agentsPath, 'utf-8');
      
      projectName = this.getFirstMatch(content, /#\s+(.+?)\s*$/m)

      description = this.getFirstMatch(content, /##\s+Overview\s*\n\n(.+?)(?=\n\n|#{1,6}\s|$)/s)

      files.push({ path: 'AGENTS.md', content, type: 'config' });
    } catch {
      // Not found
    }

    // Parse .pi/settings.json
    try {
      const settingsPath = path.join(context.sourcePath, '.pi', 'settings.json');
      const content = await fs.readFile(settingsPath, 'utf-8');
      files.push({ path: '.pi/settings.json', content, type: 'config' });
    } catch {
      // No settings
    }

    // Parse .pi/skills/*/SKILL.md (Pi calls them skills, we map to commands)
    try {
      const skillFiles = await glob('.pi/skills/*/SKILL.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of skillFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const skillId = path.basename(path.dirname(file));
          const name = this.getFirstMatch(content, /^#\s+(.+?)$/m) || skillId;

          commands.push({
            id: skillId,
            name,
            description: this.extractDescription(content),
            content,
            sourcePath: file,
          });

          files.push({ path: file, content, type: 'command' });
        } catch {
          // Skip
        }
      }
    } catch {
      // No skills
    }

    // Parse .pi/prompts/*.md
    try {
      const templateFiles = await glob('.pi/prompts/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of templateFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const templateId = path.basename(file, '.md');
          const name = this.getFirstMatch(content, /^#\s+(.+?)$/m) || templateId;

          templates.push({
            id: templateId,
            name,
            description: this.extractDescription(content),
            content,
            sourcePath: file,
          });

          files.push({ path: file, content, type: 'template' });
        } catch {
          // Skip
        }
      }
    } catch {
      // No templates
    }

    return this.buildParsedSetup({
      ...(projectName ? { projectName } : {}),
      ...(description ? { description } : {}),
      agents,
      rules,
      commands,
      templates,
      customSections,
      files,
      metadata: {
        adapter: this.id,
        parsedAt: new Date().toISOString(),
      },
    });
  }

  async merge(
    existing: ParsedSetup,
    strategy: MergeStrategy,
    _options: MigrationOptions
  ): Promise<MergeResult> {
    const newFiles: string[] = [];
    const modifiedFiles: string[] = [];
    const backupPaths: string[] = [];
    const warnings: string[] = [];
    const conflicts: MergeConflict[] = [];

    const templateDir = libraryDir;

    // Merge AGENTS.md
    try {
      const templatePath = path.join(templateDir, 'root', 'AGENTS.template.md');
      const templateContent = await fs.readFile(templatePath, 'utf-8');
      
      if (existing.files.find(f => f.path === 'AGENTS.md')) {
        const existingContent = existing.files.find(f => f.path === 'AGENTS.md')?.content;
        const lines = templateContent.split('\n');
        const existingLines = (existingContent ?? '').split('\n');
        
        const diff3Result = diff3(lines, existingLines, lines);
        
        if (diff3Result.hasConflicts) {
          if (strategy === 'smart') {
            conflicts.push(...diff3Result.conflicts.map(conflict => ({
              file: 'AGENTS.md',
              lineStart: conflict.lineStart,
              lineEnd: conflict.lineEnd,
              baseContent: conflict.base.join('\n'),
              oursContent: conflict.ours.join('\n'),
              theirsContent: conflict.theirs.join('\n'),
            })));
            modifiedFiles.push('AGENTS.md');
          } else if (strategy === 'preserve') {
            warnings.push('AGENTS.md: keeping existing version');
          } else if (strategy === 'replace') {
            newFiles.push('AGENTS.md');
          }
        } else {
          modifiedFiles.push('AGENTS.md');
        }
      } else {
        newFiles.push('AGENTS.md');
      }
    } catch (error) {
      warnings.push(`Could not merge AGENTS.md: ${error}`);
    }

    if (existing.agents.length > 0) {
      warnings.push('Pi does not support separate agent files; inline agent instructions into AGENTS.md manually.');
    }

    for (const command of existing.commands) {
      const targetPath = `.pi/skills/${command.id}/SKILL.md`;
      newFiles.push(targetPath);
    }

    for (const template of existing.templates) {
      const targetPath = `.pi/prompts/${template.id}.md`;
      newFiles.push(targetPath);
    }

    newFiles.push('.pi/settings.json');

    return {
      success: true,
      merged: true,
      conflicts,
      backupPaths,
      newFiles,
      modifiedFiles,
      warnings,
    };
  }

  private categorizeFile(filePath: string): DetectedFile['type'] {
    const normalized = filePath.toLowerCase();
    
    if (normalized.endsWith('agents.md') || normalized.endsWith('settings.json')) return 'config';
    if (normalized.includes('/skills/')) return 'command';
    if (normalized.includes('/prompts/')) return 'template';
    if (normalized.endsWith('.md')) return 'config';
    
    return 'other';
  }

  private calculatePriority(filePath: string): number {
    let priority = 0;
    const normalized = filePath.toLowerCase();
    
    if (filePath === 'AGENTS.md') priority += 100;
    if (normalized.endsWith('settings.json')) priority += 80;
    if (normalized.includes('/skills/')) priority += 40;
    if (normalized.includes('/prompts/')) priority += 30;
    
    return priority;
  }

  private extractDescription(content: string): string {
    const description = this.getFirstMatch(content, /^#\s+.+?\n\n(.+?)(?=\n\n|#{1,6}\s|```|$)/s);
    if (description) {
      return description.substring(0, 200);
    }
    return '';
  }
}

export default PiParser;
