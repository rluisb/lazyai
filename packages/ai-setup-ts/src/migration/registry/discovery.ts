// Parser Discovery - Loads parsers from multiple sources
// Discovery hierarchy:
// 1. Project local: ./ai-setup/plugins/*/parser.ts
// 2. Global user: ~/.ai-setup/parsers/*/parser.js
// 3. NPM packages: @ai-setup/parsers-*
// 4. Built-in: src/migration/parsers/*.ts (fallback)

import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import type { BaseParser, } from '../parsers/base-parser.js';

export interface ParserSource {
  id: string;
  name: string;
  path: string;
  source: 'local' | 'global' | 'npm' | 'builtin';
  version?: string;
}

export async function discoverParsers(projectPath?: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];

  if (projectPath) {
    const localParsers = await discoverLocalParsers(projectPath);
    sources.push(...localParsers);
  }

  const globalParsers = await discoverGlobalParsers();
  sources.push(...globalParsers);

  const npmParsers = await discoverNpmParsers(projectPath);
  sources.push(...npmParsers);

  return sources;
}

async function discoverLocalParsers(projectPath: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];
  const pluginsDir = path.join(projectPath, 'ai-setup', 'plugins');

  try {
    await fs.access(pluginsDir);
  } catch {
    return sources;
  }

  try {
    const entries = await fs.readdir(pluginsDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const parserPath = path.join(pluginsDir, entry.name, 'parser');
        
        for (const ext of ['.ts', '.js', '.mjs']) {
          const fullPath = parserPath + ext;
          try {
            await fs.access(fullPath);
            sources.push({
              id: entry.name,
              name: entry.name,
              path: fullPath,
              source: 'local',
            });
            break;
          } catch {
            // File doesn't exist with this extension
          }
        }
      }
    }
  } catch (error) {
    console.warn('Failed to discover local parsers:', error);
  }

  return sources;
}

async function discoverGlobalParsers(): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];
  const globalDir = path.join(os.homedir(), '.ai-setup', 'parsers');

  try {
    await fs.access(globalDir);
  } catch {
    return sources;
  }

  try {
    const entries = await fs.readdir(globalDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const parserPath = path.join(globalDir, entry.name, 'parser.js');
        
        try {
          await fs.access(parserPath);
          
          let version: string | undefined;
          const packageJsonPath = path.join(globalDir, entry.name, 'package.json');
          try {
            const packageJson = JSON.parse(await fs.readFile(packageJsonPath, 'utf-8'));
            version = packageJson.version;
          } catch {
            // No package.json or invalid
          }
          
          sources.push({
            id: entry.name,
            name: entry.name,
            path: parserPath,
            source: 'global',
            ...(version ? { version } : {}),
          });
        } catch {
          // Parser doesn't exist
        }
      }
    }
  } catch (error) {
    console.warn('Failed to discover global parsers:', error);
  }

  return sources;
}

async function discoverNpmParsers(projectPath?: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];

  try {
    const searchPaths: string[] = [];
    
    if (projectPath) {
      searchPaths.push(path.join(projectPath, 'node_modules'));
    }
    
    try {
      const { execSync } = await import('node:child_process');
      const globalPath = execSync('npm root -g', { encoding: 'utf-8' }).trim();
      searchPaths.push(globalPath);
    } catch {
      // npm not available
    }

    for (const modulesPath of searchPaths) {
      try {
        await fs.access(modulesPath);
        
        const entries = await fs.readdir(modulesPath, { withFileTypes: true });
        
        for (const entry of entries) {
          if (entry.isDirectory() && entry.name.startsWith('@ai-setup/parsers-')) {
            const parserId = entry.name.replace('@ai-setup/parsers-', '');
            const packagePath = path.join(modulesPath, entry.name);
            const packageJsonPath = path.join(packagePath, 'package.json');
            
            try {
              const packageJson = JSON.parse(await fs.readFile(packageJsonPath, 'utf-8'));
              
              const mainFile = packageJson.main || 'dist/parser.js';
              const parserPath = path.join(packagePath, mainFile);
              
              await fs.access(parserPath);
              
              sources.push({
                id: parserId,
                name: packageJson.name,
                path: parserPath,
                source: 'npm',
                version: packageJson.version,
              });
            } catch {
              // Invalid package
            }
          }
        }
      } catch {
        // modules directory doesn't exist
      }
    }
  } catch (error) {
    console.warn('Failed to discover NPM parsers:', error);
  }

  return sources;
}

export async function loadParser(source: ParserSource): Promise<BaseParser | null> {
  try {
    if (source.source === 'npm' || source.source === 'global') {
      const module = await import(source.path);
      const ParserClass = module.default || module.Parser;
      
      if (ParserClass && typeof ParserClass === 'function') {
        return new ParserClass();
      }
    } else if (source.source === 'local') {
      const module = await import(source.path);
      const ParserClass = module.default || module.Parser;
      
      if (ParserClass && typeof ParserClass === 'function') {
        return new ParserClass();
      }
    }
  } catch (error) {
    console.warn(`Failed to load parser from ${source.path}:`, error);
  }

  return null;
}

export async function getAllParsers(projectPath?: string): Promise<BaseParser[]> {
  const parsers: BaseParser[] = [];
  const loadedIds = new Set<string>();

  const sources = await discoverParsers(projectPath);
  
  for (const source of sources) {
    if (loadedIds.has(source.id)) continue;
    
    const parser = await loadParser(source);
    if (parser) {
      parsers.push(parser);
      loadedIds.add(source.id);
    }
  }

  // Load built-in parsers
  const { OpenCodeParser } = await import('../parsers/opencode-parser.js');
  const { ClaudeCodeParser } = await import('../parsers/claude-parser.js');
  const { GeminiParser } = await import('../parsers/gemini-parser.js');
  const { CopilotParser } = await import('../parsers/copilot-parser.js');

  const builtinParsers = [
    new OpenCodeParser(),
    new ClaudeCodeParser(),
    new GeminiParser(),
    new CopilotParser(),
  ];

  for (const parser of builtinParsers) {
    if (!loadedIds.has(parser.id)) {
      parsers.push(parser);
      loadedIds.add(parser.id);
    }
  }

  return parsers.sort((a, b) => b.getPriority() - a.getPriority());
}

export function createParserTemplate(id: string, name: string): string {
  return `import { BaseParser } from '@ricardoborges-teachable/ai-setup/migration';

export class ${name}Parser extends BaseParser {
  readonly id = '${id}';
  readonly name = '${name}';
  readonly description = 'Parser for ${name} AI setup';
  readonly version = '1.0.0';
  readonly supportedPatterns = [
    '.${id}/**/*',
    '${id}.md',
  ];

  async detect(context) {
    return {
      detected: false,
      confidence: 0,
      adapterId: this.id,
      adapterName: this.name,
      files: [],
    };
  }

  async parse(context) {
    return {
      agents: [],
      rules: [],
      commands: [],
      templates: [],
      customSections: [],
      files: [],
      metadata: {},
    };
  }

  async merge(existing, strategy, options) {
    return {
      success: true,
      merged: true,
      conflicts: [],
      backupPaths: [],
      newFiles: [],
      modifiedFiles: [],
      warnings: [],
    };
  }
}

export default ${name}Parser;
`;
}
