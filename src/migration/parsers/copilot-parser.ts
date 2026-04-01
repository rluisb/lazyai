/**
 * GitHub Copilot Parser
 * 
 * Detects, parses, and merges GitHub Copilot AI setups.
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

export class CopilotParser extends BaseParser {
  readonly id = 'copilot';
  readonly name = 'GitHub Copilot';
  readonly description = 'Parser for GitHub Copilot AI assistant configuration';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.github/copilot-instructions.md',
    '.github/prompts/*.md',
    '.github/instructions/*.md',
    '.github/copilot/*.md',
  ];

  async detect(context: MigrationContext): Promise<DetectionResult> {
    const files: DetectedFile[] = [];
    let detected = false;
    let confidence = 0;

    // Check for copilot-instructions.md (main file)
    try {
      const instructionsPath = path.join(context.sourcePath, '.github', 'copilot-instructions.md');
      await fs.access(instructionsPath);
      files.push({
        path: '.github/copilot-instructions.md',
        type: 'config',
        priority: 100,
      });
      detected = true;
      confidence += 0.5;
    } catch {
      // Not found
    }

    // Check for prompts directory
    try {
      const promptsDir = path.join(context.sourcePath, '.github', 'prompts');
      await fs.access(promptsDir);
      
      const promptFiles = await glob('.github/prompts/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of promptFiles) {
        files.push({
          path: file,
          type: 'command',
          priority: 60,
        });
        detected = true;
        confidence += 0.1;
      }
    } catch {
      // Directory doesn't exist
    }

    // Check for instructions directory
    try {
      const instructionsDir = path.join(context.sourcePath, '.github', 'instructions');
      await fs.access(instructionsDir);
      
      const instructionFiles = await glob('.github/instructions/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of instructionFiles) {
        files.push({
          path: file,
          type: 'rule',
          priority: 50,
        });
        detected = true;
        confidence += 0.1;
      }
    } catch {
      // Directory doesn't exist
    }

    // Check for copilot directory
    try {
      const copilotDir = path.join(context.sourcePath, '.github', 'copilot');
      await fs.access(copilotDir);
      
      const copilotFiles = await glob('.github/copilot/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of copilotFiles) {
        files.push({
          path: file,
          type: this.categorizeFile(file),
          priority: 40,
        });
        detected = true;
        confidence += 0.05;
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

    // Parse copilot-instructions.md
    try {
      const instructionsPath = path.join(context.sourcePath, '.github', 'copilot-instructions.md');
      const content = await fs.readFile(instructionsPath, 'utf-8');
      
      const nameMatch = content.match(/^#\s+(.+?)$/m);
      if (nameMatch) {
        projectName = nameMatch[1].trim();
      }

      const descMatch = content.match(/##\s+(?:Overview|Description)\s*\n\n(.+?)(?=\n\n|#{1,6}\s|$)/is);
      if (descMatch) {
        description = descMatch[1].trim();
      }

      files.push({ path: '.github/copilot-instructions.md', content, type: 'config' });
    } catch {
      // Not found
    }

    // Parse .github/prompts/*.md (Copilot calls them prompts, we map to commands)
    try {
      const promptFiles = await glob('.github/prompts/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of promptFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const promptId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : promptId;

          commands.push({
            id: promptId,
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
      // No prompts
    }

    // Parse .github/instructions/*.md (Copilot instructions, we map to rules)
    try {
      const instructionFiles = await glob('.github/instructions/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of instructionFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const instructionId = path.basename(file, '.md');

          rules.push({
            id: instructionId,
            category: 'copilot-instruction',
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
      // No instructions
    }

    // Parse .github/copilot/*.md
    try {
      const copilotFiles = await glob('.github/copilot/*.md', {
        cwd: context.sourcePath,
        absolute: false,
      });

      for (const file of copilotFiles) {
        try {
          const fullPath = path.join(context.sourcePath, file);
          const content = await fs.readFile(fullPath, 'utf-8');
          
          const fileId = path.basename(file, '.md');
          const nameMatch = content.match(/^#\s+(.+?)$/m);
          const name = nameMatch ? nameMatch[1].trim() : fileId;

          // Categorize based on content
          if (content.toLowerCase().includes('agent') || content.toLowerCase().includes('role')) {
            agents.push({
              id: fileId,
              name,
              description: this.extractDescription(content),
              role: fileId,
              content,
              sourcePath: file,
              custom: true,
            });
            files.push({ path: file, content, type: 'agent' });
          } else {
            templates.push({
              id: fileId,
              name,
              description: this.extractDescription(content),
              content,
              sourcePath: file,
            });
            files.push({ path: file, content, type: 'template' });
          }
        } catch {
          // Skip
        }
      }
    } catch {
      // No copilot directory files
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

    // Merge copilot-instructions.md
    try {
      const templatePath = path.join(templateDir, 'root', 'copilot-instructions.template.md');
      const templateContent = await fs.readFile(templatePath, 'utf-8');
      
      if (existing.files.find(f => f.path === '.github/copilot-instructions.md')) {
        const existingContent = existing.files.find(f => f.path === '.github/copilot-instructions.md')!.content;
        const lines = templateContent.split('\n');
        const existingLines = existingContent.split('\n');
        
        const diff3Result = diff3(lines, existingLines, lines);
        
        if (diff3Result.hasConflicts) {
          if (strategy === 'smart') {
            conflicts.push(...diff3Result.conflicts);
            modifiedFiles.push('.github/copilot-instructions.md');
          } else if (strategy === 'preserve') {
            warnings.push('copilot-instructions.md: keeping existing version');
          } else if (strategy === 'replace') {
            newFiles.push('.github/copilot-instructions.md');
          }
        } else {
          modifiedFiles.push('.github/copilot-instructions.md');
        }
      } else {
        newFiles.push('.github/copilot-instructions.md');
      }
    } catch (error) {
      warnings.push(`Could not merge copilot-instructions.md: ${error}`);
    }

    // Migrate prompts
    for (const command of existing.commands) {
      const targetPath = `.github/prompts/${command.id}.md`;
      newFiles.push(targetPath);
    }

    // Migrate instructions
    for (const rule of existing.rules) {
      const targetPath = `.github/instructions/${rule.id}.md`;
      newFiles.push(targetPath);
    }

    // Migrate agents/templates
    for (const agent of existing.agents) {
      const targetPath = `.github/copilot/${agent.id}.md`;
      newFiles.push(targetPath);
    }

    for (const template of existing.templates) {
      const targetPath = `.github/copilot/${template.id}.md`;
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
    
    if (normalized.includes('/prompts/')) return 'command';
    if (normalized.includes('/instructions/')) return 'rule';
    if (normalized.includes('/copilot/')) return 'agent';
    if (normalized.includes('copilot-instructions')) return 'config';
    
    return 'other';
  }

  private calculatePriority(filePath: string): number {
    let priority = 0;
    const normalized = filePath.toLowerCase();
    
    if (normalized.includes('copilot-instructions.md')) priority += 100;
    if (normalized.includes('/prompts/')) priority += 50;
    if (normalized.includes('/instructions/')) priority += 40;
    if (normalized.includes('/copilot/')) priority += 30;
    
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

export default CopilotParser;
