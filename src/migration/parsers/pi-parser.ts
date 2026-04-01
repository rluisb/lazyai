/**
 * Pi Parser
 * 
 * Detects, parses, and merges Pi AI setups.
 * Pi is a Claude Code wrapper that uses similar structure.
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

export class PiParser extends BaseParser {
  readonly id = 'pi';
  readonly name = 'Pi';
  readonly description = 'Parser for Pi AI assistant configuration (Claude Code wrapper)';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.pi/**/*',
    'CLAUDE.md',
    '.pi/agents/*.md',
    '.pi/skills/*.md',
    '.pi/templates/*.md',
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

    // Check for CLAUDE.md (Pi uses this as root)
    if (detected) {
      try {
        const claudePath = path.join(context.sourcePath, 'CLAUDE.md');
        await fs.access(claudePath);
        files.push({
          path: 'CLAUDE.md',
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
    const customSections: any[] = [];
    const files: ParsedFile[] = [];
    let projectName: string | undefined;
    let description: string | undefined;

    // Parse CLAUDE.md
    try {
      const claudePath = path.join(context.sourcePath, 'CLAUDE.md');
      const content = await fs.readFile(claudePath, 'utf-8');
      
      const nameMatch = content.match(/#\s+(.+?)\s*$/m);
      if (nameMatch) {
        projectName = nameMatch[1].trim();
      }

      const descMatch = content.match(/##\s+Overview\s*\n\n(.+?)(?=\n\n|#{1,6}\s|$)/s);
      if (descMatch) {
        description = descMatch[1].trim();
      }

      files.push({ path: 'CLAUDE.md', content, type: 'config' });
    } catch {
      // Not found
    }

    // Parse .pi/agents/*.md
    try {
      const agentFiles = await glob('.pi/agents/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of agentFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const agentId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : agentId;

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

    // Parse .pi/skills/*.md (Pi calls them skills, we map to commands)
    try {
      const skillFiles = await glob('.pi/skills/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of skillFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const skillId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : skillId;

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

    // Parse .pi/templates/*.md
    try {
      const templateFiles = await glob('.pi/templates/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of templateFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const templateId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : templateId;

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

    return {
      projectName,
      description,
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
    };
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

    // Merge CLAUDE.md (shared with Claude Code)
    try {
      const templatePath = path.join(templateDir, 'root', 'CLAUDE.template.md');
      const templateContent = await fs.readFile(templatePath, 'utf-8');
      
      if (existing.files.find(f => f.path === 'CLAUDE.md')) {
        const existingContent = existing.files.find(f => f.path === 'CLAUDE.md')!.content;
        const lines = templateContent.split('\n');
        const existingLines = existingContent.split('\n');
        
        const diff3Result = diff3(lines, existingLines, lines);
        
        if (diff3Result.hasConflicts) {
          if (strategy === 'smart') {
            conflicts.push(...diff3Result.conflicts);
            modifiedFiles.push('CLAUDE.md');
          } else if (strategy === 'preserve') {
            warnings.push('CLAUDE.md: keeping existing version');
          } else if (strategy === 'replace') {
            newFiles.push('CLAUDE.md');
          }
        } else {
          modifiedFiles.push('CLAUDE.md');
        }
      } else {
        newFiles.push('CLAUDE.md');
      }
    } catch (error) {
      warnings.push(`Could not merge CLAUDE.md: ${error}`);
    }

    // Migrate to .pi structure
    for (const agent of existing.agents) {
      const targetPath = `.pi/agents/${agent.id}.md`;
      newFiles.push(targetPath);
    }

    for (const command of existing.commands) {
      const targetPath = `.pi/skills/${command.id}.md`;
      newFiles.push(targetPath);
    }

    for (const template of existing.templates) {
      const targetPath = `.pi/templates/${template.id}.md`;
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
    
    if (filePath === 'CLAUDE.md') priority += 100;
    if (normalized.includes('/agents/')) priority += 50;
    if (normalized.includes('/skills/')) priority += 40;
    if (normalized.includes('/templates/')) priority += 30;
    
    return priority;
  }

  private extractDescription(content: string): string {
    const match = content.match(/^#\s+.+?\n\n(.+?)(?=\n\n|#{1,6}\s|```|$)/s);
    if (match) {
      return match[1].trim().substring(0, 200);
    }
    return '';
  }
}

export default PiParser;
