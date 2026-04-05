/**
 * OpenCode Parser
 * 
 * Detects, parses, and merges OpenCode AI setups.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { glob } from 'glob';
import { resolveLibraryDir } from '../../utils/files.js';
import { diff3, } from '../diff/diff3.js';
import type { 
  AgentDefinition,
  CommandDefinition,
  CustomSection,
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

export class OpenCodeParser extends BaseParser {
  readonly id = 'opencode';
  readonly name = 'OpenCode';
  readonly description = 'Parser for OpenCode AI assistant configuration';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.opencode/**/*',
    'AGENTS.md',
    '.opencode/agents/*.md',
    '.opencode/commands/*.md',
    '.opencode/templates/*.md',
  ];

  async detect(context: MigrationContext): Promise<DetectionResult> {
    const files: DetectedFile[] = [];
    let detected = false;
    let confidence = 0;

    // Check for AGENTS.md
    try {
      const agentsPath = path.join(context.sourcePath, 'AGENTS.md');
      await fs.access(agentsPath);
      files.push({
        path: 'AGENTS.md',
        type: 'config',
        priority: 100,
      });
      detected = true;
      confidence += 0.3;
    } catch {
      // Not found
    }

    // Check for .opencode directory
    try {
      const opencodePath = path.join(context.sourcePath, '.opencode');
      await fs.access(opencodePath);
      
      // Find all files in .opencode
      const opencodeFiles = await glob('.opencode/**/*', {
        cwd: context.sourcePath,
        absolute: false,
        dot: true,
      });

      for (const file of opencodeFiles) {
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
          // Skip inaccessible files
        }
      }
    } catch {
      // Directory doesn't exist
    }

    // Cap confidence at 1.0
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
      
      // Extract project name
      projectName = this.getFirstMatch(content, /#\s+(.+?)\s*-?\s*AI Agent/i) ||
        this.getFirstMatch(content, /#\s+(.+?)\s*$/m)

      // Extract description (first paragraph after title)
      description = this.getFirstMatch(content, /#\s+.+?\n\n(.+?)(?=\n\n|#{1,6}\s)/s)

      files.push({ path: 'AGENTS.md', content, type: 'config' });
    } catch {
      // AGENTS.md doesn't exist
    }

    // Parse .opencode/agents/*.md
    try {
      const agentFiles = await glob('.opencode/agents/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of agentFiles) {
        if (path.basename(file) === 'AGENTS.md') continue;
        
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          // Extract agent name from filename
          const agentId = path.basename(file, '.md');
          
          // Extract agent name from first heading
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
          // Skip unreadable files
        }
      }
    } catch {
      // No agents directory
    }

    // Parse .opencode/commands/*.md
    try {
      const commandFiles = await glob('.opencode/commands/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of commandFiles) {
        if (path.basename(file) === 'AGENTS.md') continue;
        
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const commandId = path.basename(file, '.md');
          const name = this.getFirstMatch(content, /^#\s+(.+?)$/m) || commandId;

          commands.push({
            id: commandId,
            name,
            description: this.extractDescription(content),
            content,
            sourcePath: file,
          });

          files.push({ path: file, content, type: 'command' });
        } catch {
          // Skip unreadable files
        }
      }
    } catch {
      // No commands directory
    }

    // Parse .opencode/templates/*.md
    try {
      const templateFiles = await glob('.opencode/templates/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of templateFiles) {
        if (path.basename(file) === 'AGENTS.md') continue;
        
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
          // Skip unreadable files
        }
      }
    } catch {
      // No templates directory
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

    // Load ai-setup templates
    const templateDir = libraryDir;

    // Merge AGENTS.md
    try {
      const templatePath = path.join(templateDir, 'root', 'AGENTS.template.md');
      const templateContent = await fs.readFile(templatePath, 'utf-8');
      
      if (existing.files.find(f => f.path === 'AGENTS.md')) {
        // File exists - merge it
        const existingContent = existing.files.find(f => f.path === 'AGENTS.md')?.content;
        const lines = templateContent.split('\n');
        const existingLines = (existingContent ?? '').split('\n');
        
        // Use diff3 for merge
        const diff3Result = diff3(lines, existingLines, lines);
        
        if (diff3Result.hasConflicts) {
          if (strategy === 'smart') {
            // Keep conflicts for manual resolution
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
            // Keep existing
            warnings.push('AGENTS.md: keeping existing version');
          } else if (strategy === 'replace') {
            // Use template
            newFiles.push('AGENTS.md');
          } else if (strategy === 'append') {
            // Append custom sections
            modifiedFiles.push('AGENTS.md');
          }
        } else {
          modifiedFiles.push('AGENTS.md');
        }
      } else {
        // New file
        newFiles.push('AGENTS.md');
      }
    } catch (error) {
      warnings.push(`Could not merge AGENTS.md: ${error}`);
    }

    // Migrate agents
    for (const agent of existing.agents) {
      const targetPath = `.opencode/agents/${agent.id}.md`;
      
      try {
        const templatePath = path.join(templateDir, 'agents', `${agent.role}.md`);
        await fs.access(templatePath);
        
        // Template exists - this is a modify
        modifiedFiles.push(targetPath);
      } catch {
        // No template - this is a custom agent, create as new
        newFiles.push(targetPath);
      }
    }

    // Migrate commands
    for (const command of existing.commands) {
      const targetPath = `.opencode/commands/${command.id}.md`;
      newFiles.push(targetPath);
    }

    // Migrate templates
    for (const template of existing.templates) {
      const targetPath = `.opencode/templates/${template.id}.md`;
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
    
    if (normalized.includes('agent')) return 'agent';
    if (normalized.includes('command')) return 'command';
    if (normalized.includes('template')) return 'template';
    if (normalized.includes('rule')) return 'rule';
    if (normalized.endsWith('.md')) return 'config';
    
    return 'other';
  }

  private calculatePriority(filePath: string): number {
    let priority = 0;
    const normalized = filePath.toLowerCase();
    
    // Root files
    if (filePath === 'AGENTS.md') priority += 100;
    
    // Directory roots
    if (normalized.includes('/agents/agents.md')) priority += 90;
    if (normalized.includes('/commands/agents.md')) priority += 80;
    if (normalized.includes('/templates/agents.md')) priority += 70;
    
    // Individual files
    if (normalized.includes('/agents/')) priority += 50;
    if (normalized.includes('/commands/')) priority += 40;
    if (normalized.includes('/templates/')) priority += 30;
    
    return priority;
  }

  private extractDescription(content: string): string {
    // Try to find description after title
    const description = this.getFirstMatch(content, /^#\s+.+?\n\n(.+?)(?=\n\n|#{1,6}\s|```|$)/s);
    if (description) {
      return description.substring(0, 200);
    }
    return '';
  }
}

export default OpenCodeParser;
