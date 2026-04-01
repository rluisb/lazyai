/**
 * Parser Discovery - Loads parsers from multiple sources
 * 
 * Discovery hierarchy:
 * 1. Project local: ./ai-setup/plugins/*/parser.ts
 * 2. Global user: ~/.ai-setup/parsers/*/parser.js
 * 3. NPM packages: @ai-setup/parsers-*
 * 4. Built-in: src/migration/parsers/*.ts (fallback)
 */

import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import { createRequire } from 'module';
import { glob } from 'glob';
import { BaseParser, ParserFactory } from '../parsers/base-parser.js';

const require = createRequire(import.meta.url);

export interface ParserSource {
  id: string;
  name: string;
  path: string;
  source: 'local' | 'global' | 'npm' | 'builtin';
  version?: string;
}

/**
 * Discover all available parsers from all sources
 */
export async function discoverParsers(projectPath?: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];

  // 1. Discover project local parsers
  if (projectPath) {
    const localParsers = await discoverLocalParsers(projectPath);
    sources.push(...localParsers);
  }

  // 2. Discover global user parsers
  const globalParsers = await discoverGlobalParsers();
  sources.push(...globalParsers);

  // 3. Discover NPM package parsers
  const npmParsers = await discoverNpmParsers(projectPath);
  sources.push(...npmParsers);

  // 4. Built-in parsers are loaded separately

  return sources;
}

/**
 * Discover project-local parsers
 * Location: ./ai-setup/plugins/*/parser.ts (or .js)
 */
async function discoverLocalParsers(projectPath: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];
  const pluginsDir = path.join(projectPath, 'ai-setup', 'plugins');

  try {
    await fs.access(pluginsDir);
  } catch {
    return sources; // Directory doesn't exist
  }

  try {
    const entries = await fs.readdir(pluginsDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const parserPath = path.join(pluginsDir, entry.name, 'parser');
        
        // Check for .ts or .js
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

/**
 * Discover global user parsers
 * Location: ~/.ai-setup/parsers/*/parser.js
 */
async function discoverGlobalParsers(): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];
  const globalDir = path.join(os.homedir(), '.ai-setup', 'parsers');

  try {
    await fs.access(globalDir);
  } catch {
    return sources; // Directory doesn't exist
  }

  try {
    const entries = await fs.readdir(globalDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const parserPath = path.join(globalDir, entry.name, 'parser.js');
        
        try {
          await fs.access(parserPath);
          
          // Try to read version from package.json if exists
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
            version,
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

/**
 * Discover NPM package parsers
 * Pattern: @ai-setup/parsers-*
 */
async function discoverNpmParsers(projectPath?: string): Promise<ParserSource[]> {
  const sources: ParserSource[] = [];

  try {
    // Try to find ai-setup in node_modules
    const searchPaths: string[] = [];
    
    if (projectPath) {
      searchPaths.push(path.join(projectPath, 'node_modules'));
    }
    
    // Also check global node_modules
    try {
      const { execSync } = await import('child_process');
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
              
              // Look for main entry point
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

/**
 * Load a parser from a source
 */
export async function loadParser(source: ParserSource): Promise<BaseParser | null> {
  try {
    if (source.source === 'npm' || source.source === 'global') {
      // Dynamic import for JS modules
      const module = await import(source.path);
      const ParserClass = module.default || module.Parser;
      
      if (ParserClass && typeof ParserClass === 'function') {
        return new ParserClass();
      }
    } else if (source.source === 'local') {
      // For local TypeScript files, we might need to use ts-node or similar
      // For now, treat as dynamic import
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

/**
 * Get all parsers including built-in
 */
export async function getAllParsers(projectPath?: string): Promise<BaseParser[]> {
  const parsers: BaseParser[] = [];
  const loadedIds = new Set<string>();

  // 1. Load discovered parsers (project-local, global, npm)
  const sources = await discoverParsers(projectPath);
  
  for (const source of sources) {
    if (loadedIds.has(source.id)) continue;
    
    const parser = await loadParser(source);
    if (parser) {
      parsers.push(parser);
      loadedIds.add(source.id);
    }
  }

  // 2. Load built-in parsers (fallback)
  // These are imported statically to ensure they're always available
  const { OpenCodeParser } = await import('../parsers/opencode-parser.js');
  const { ClaudeCodeParser } = await import('../parsers/claude-parser.js');
  const { PiParser } = await import('../parsers/pi-parser.js');
  const { GeminiParser } = await import('../parsers/gemini-parser.js');
  const { CopilotParser } = await import('../parsers/copilot-parser.js');

  const builtinParsers = [
    new OpenCodeParser(),
    new ClaudeCodeParser(),
    new PiParser(),
    new GeminiParser(),
    new CopilotParser(),
  ];

  for (const parser of builtinParsers) {
    if (!loadedIds.has(parser.id)) {
      parsers.push(parser);
      loadedIds.add(parser.id);
    }
  }

  // Sort by priority
  return parsers.sort((a, b) => b.getPriority() - a.getPriority());
}

/**
 * Create a parser template for community extensions
 */
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
    // TODO: Implement detection logic
    return {
      detected: false,
      confidence: 0,
      adapterId: this.id,
      adapterName: this.name,
      files: [],
    };
  }

  async parse(context) {
    // TODO: Implement parsing logic
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
    // TODO: Implement merge logic
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
