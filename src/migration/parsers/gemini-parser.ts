/**
 * Gemini Parser
 * 
 * Detects, parses, and merges Gemini CLI AI setups.
 */

import { BaseParser } from './base-parser.js';
import {
  MigrationContext,
  DetectionResult,
  ParsedSetup,
  MergeResult,
  MergeStrategy,
  MigrationOptions,
  DetectedFile,
  AgentDefinition,
  RuleDefinition,
  CommandDefinition,
  TemplateDefinition,
  ParsedFile,
} from '../types.js';
import { promises as fs } from 'fs';
import path from 'path';
import { glob } from 'glob';
import { diff3 } from '../diff/diff3.js';

export class GeminiParser extends BaseParser {
  readonly id = 'gemini';
  readonly name = 'Gemini CLI';
  readonly description = 'Parser for Gemini CLI AI assistant configuration';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.gemini/**/*',
    'GEMINI.md',
    '.gemini/agents/*.md',
    '.gemini/skills/*.md',
    '.gemini/templates/*.md',
  ];

  async detect(context: MigrationContext): Promise<DetectionResult> {
    const files: DetectedFile[] = [];
    let detected = false;
    let confidence = 0;

    // Check for GEMINI.md
    try {
      const geminiPath = path.join(context.sourcePath, 'GEMINI.md');
      await fs.access(geminiPath);
      files.push({
        path: 'GEMINI.md',
        type: 'config',
        priority: 100,
      });
      detected = true;
      confidence += 0.4;
    } catch {
      // Not found
    }

    // Check for .gemini directory
    try {
      const geminiDir = path.join(context.sourcePath, '.gemini');
      await fs.access(geminiDir);
      
      const geminiFiles = await glob('.gemini/**/*', {
        cwd: context.sourcePath,
        absolute: false,
        dot: true,
      });

      for (const file of geminiFiles) {
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
    const customSections: any[] = [];
    const files: ParsedFile[] = [];
    let projectName: string | undefined;
    let description: string | undefined;

    // Parse GEMINI.md
    try {
      const geminiPath = path.join(context.sourcePath, 'GEMINI.md');
      const content = await fs.readFile(geminiPath, 'utf-8');
      
      projectName = this.getFirstMatch(content, /^#\s+(.+?)$/m)

      description = this.getFirstMatch(content, /##\s+(?:Overview|Description)\s*\n\n(.+?)(?=\n\n|#{1,6}\s|$)/is)

      files.push({ path: 'GEMINI.md', content, type: 'config' });
    } catch {
      // Not found
    }

    // Parse .gemini/agents/*.md
    try {
      const agentFiles = await glob('.gemini/agents/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of agentFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const agentId = path.basename(file, '.md');
          const name = this.getFirstMatch(content, /^#\s+(.+?)$/m) || agentId;

          agents.push({
            id: agentId,
            name,
            description: this.extractDescription(content),
            role: agentId,
            content,
            sourcePath: file,
            custom: true,
          });

          files.push({ path: file, content, type: 'agent' });
        } catch {
          // Skip
        }
      }
    } catch {
      // No agents
    }

    // Parse .gemini/skills/*.md
    try {
      const skillFiles = await glob('.gemini/skills/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of skillFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const skillId = path.basename(file, '.md');
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

    // Parse .gemini/templates/*.md
    try {
      const templateFiles = await glob('.gemini/templates/*.md', {
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
    options: MigrationOptions
  ): Promise<MergeResult> {
    const newFiles: string[] = [];
    const modifiedFiles: string[] = [];
    const backupPaths: string[] = [];
    const warnings: string[] = [];
    const conflicts: any[] = [];

    const templateDir = new URL('../../../library', import.meta.url).pathname;

    // Merge GEMINI.md
    try {
      const templatePath = path.join(templateDir, 'root', 'GEMINI.template.md');
      const templateContent = await fs.readFile(templatePath, 'utf-8');
      
      if (existing.files.find(f => f.path === 'GEMINI.md')) {
        const existingContent = existing.files.find(f => f.path === 'GEMINI.md')!.content;
        const lines = templateContent.split('\n');
        const existingLines = existingContent.split('\n');
        
        const diff3Result = diff3(lines, existingLines, lines);
        
        if (diff3Result.hasConflicts) {
          if (strategy === 'smart') {
            conflicts.push(...diff3Result.conflicts.map(conflict => ({
              file: 'GEMINI.md',
              lineStart: conflict.lineStart,
              lineEnd: conflict.lineEnd,
              baseContent: conflict.base.join('\n'),
              oursContent: conflict.ours.join('\n'),
              theirsContent: conflict.theirs.join('\n'),
            })));
            modifiedFiles.push('GEMINI.md');
          } else if (strategy === 'preserve') {
            warnings.push('GEMINI.md: keeping existing version');
          } else if (strategy === 'replace') {
            newFiles.push('GEMINI.md');
          }
        } else {
          modifiedFiles.push('GEMINI.md');
        }
      } else {
        newFiles.push('GEMINI.md');
      }
    } catch (error) {
      warnings.push(`Could not merge GEMINI.md: ${error}`);
    }

    // Migrate to .gemini structure
    for (const agent of existing.agents) {
      const targetPath = `.gemini/agents/${agent.id}.md`;
      newFiles.push(targetPath);
    }

    for (const command of existing.commands) {
      const targetPath = `.gemini/skills/${command.id}.md`;
      newFiles.push(targetPath);
    }

    for (const template of existing.templates) {
      const targetPath = `.gemini/templates/${template.id}.md`;
      newFiles.push(targetPath);
    }

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
    
    if (normalized.includes('/agents/')) return 'agent';
    if (normalized.includes('/skills/')) return 'command';
    if (normalized.includes('/templates/')) return 'template';
    if (normalized.endsWith('.md')) return 'config';
    
    return 'other';
  }

  private calculatePriority(filePath: string): number {
    let priority = 0;
    const normalized = filePath.toLowerCase();
    
    if (filePath === 'GEMINI.md') priority += 100;
    if (normalized.includes('/agents/')) priority += 50;
    if (normalized.includes('/skills/')) priority += 40;
    if (normalized.includes('/templates/')) priority += 30;
    
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

export default GeminiParser;
