/**
 * Claude Code Parser
 * 
 * Detects, parses, and merges Claude Code AI setups.
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

export class ClaudeCodeParser extends BaseParser {
  readonly id = 'claude-code';
  readonly name = 'Claude Code';
  readonly description = 'Parser for Claude Code AI assistant configuration';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.claude/**/*',
    'CLAUDE.md',
    '.claude/*.md',
    '.claude/commands/*.md',
    '.claude/rules/*.md',
    '.claude/templates/*.md',
  ];

  async detect(context: MigrationContext): Promise<DetectionResult> {
    const files: DetectedFile[] = [];
    let detected = false;
    let confidence = 0;

    // Check for CLAUDE.md
    try {
      const claudePath = path.join(context.sourcePath, 'CLAUDE.md');
      await fs.access(claudePath);
      files.push({
        path: 'CLAUDE.md',
        type: 'config',
        priority: 100,
      });
      detected = true;
      confidence += 0.4;
    } catch {
      // Not found
    }

    // Check for .claude directory
    try {
      const claudeDir = path.join(context.sourcePath, '.claude');
      await fs.access(claudeDir);
      
      const claudeFiles = await glob('.claude/**/*', {
        cwd: context.sourcePath,
        absolute: false,
        dot: true,
      });

      for (const file of claudeFiles) {
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
            confidence += 0.05;
          }
        } catch {
          // Skip inaccessible files
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

    // Parse CLAUDE.md
    try {
      const claudePath = path.join(context.sourcePath, 'CLAUDE.md');
      const content = await fs.readFile(claudePath, 'utf-8');
      
      const nameMatch = content.match(/#\s+(.+?)\s*$/m);
      if (nameMatch) {
        projectName = nameMatch[1].trim();
      }

      const descMatch = content.match(/##\s+Overview\s*\n\n(.+?)(?=\n\n|#{1,6}\s|```|$)/s);
      if (descMatch) {
        description = descMatch[1].trim();
      }

      files.push({ path: 'CLAUDE.md', content, type: 'config' });
    } catch {
      // CLAUDE.md doesn't exist
    }

    // Parse .claude/*.md (agent definitions)
    try {
      const agentFiles = await glob('.claude/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of agentFiles) {
        if (path.basename(file) === 'CLAUDE.md') continue;
        
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

    // Parse .claude/commands/*.md
    try {
      const commandFiles = await glob('.claude/commands/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of commandFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const commandId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : commandId;

          commands.push({
            id: commandId,
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
      // No commands
    }

    // Parse .claude/rules/*.md
    try {
      const ruleFiles = await glob('.claude/rules/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of ruleFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const ruleId = path.basename(file, '.md');

          rules.push({
            id: ruleId,
            category: 'claude-rule',
            content,
            sourcePath: file,
            priority: 50,
          });

          files.push({ path: file, content, type: 'rule' });
        } catch {
          // Skip
        }
      }
    } catch {
      // No rules
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

    // Load ai-setup templates
    const templateDir = new URL('../../../library', import.meta.url).pathname;

    // Merge CLAUDE.md
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
          } else if (strategy === 'append') {
            modifiedFiles.push('CLAUDE.md');
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

    // Migrate agents
    for (const agent of existing.agents) {
      const targetPath = `.claude/${agent.id}.md`;
      newFiles.push(targetPath);
    }

    // Migrate commands
    for (const command of existing.commands) {
      const targetPath = `.claude/commands/${command.id}.md`;
      newFiles.push(targetPath);
    }

    // Migrate rules
    for (const rule of existing.rules) {
      const targetPath = `.claude/rules/${rule.id}.md`;
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
    
    if (normalized.includes('/commands/')) return 'command';
    if (normalized.includes('/rules/')) return 'rule';
    if (normalized.includes('/templates/')) return 'template';
    if (normalized.endsWith('.md') && !normalized.includes('/')) return 'agent';
    
    return 'config';
  }

  private calculatePriority(filePath: string): number {
    let priority = 0;
    const normalized = filePath.toLowerCase();
    
    if (filePath === 'CLAUDE.md') priority += 100;
    if (!normalized.includes('/')) priority += 50; // Root level agents
    if (normalized.includes('/commands/')) priority += 40;
    if (normalized.includes('/rules/')) priority += 30;
    if (normalized.includes('/templates/')) priority += 20;
    
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

export default ClaudeCodeParser;
