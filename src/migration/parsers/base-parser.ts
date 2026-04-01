/**
 * Base Parser - Abstract class for all migration parsers
 * 
 * Parsers detect, parse, and merge existing AI setups from various tools.
 * They follow a consistent interface for extensibility.
 */

import {
  MigrationContext,
  DetectionResult,
  ParsedSetup,
  MergeResult,
  MergeStrategy,
  MigrationOptions,
} from '../types.js';

export abstract class BaseParser {
  /**
   * Unique identifier for this parser (e.g., 'opencode', 'claude-code')
   */
  abstract readonly id: string;

  /**
   * Human-readable name for this parser
   */
  abstract readonly name: string;

  /**
   * Description of what this parser handles
   */
  abstract readonly description: string;

  /**
   * Version of this parser
   */
  abstract readonly version: string;

  /**
   * File patterns that this parser can detect
   * Supports glob patterns
   */
  abstract readonly supportedPatterns: string[];

  /**
   * Detect if this parser can handle the given context
   * @param context Migration context with source path
   * @returns Detection result with confidence score
   */
  abstract detect(context: MigrationContext): Promise<DetectionResult>;

  /**
   * Parse the existing setup into a structured format
   * @param context Migration context
   * @returns Parsed setup with all extracted data
   */
  abstract parse(context: MigrationContext): Promise<ParsedSetup>;

  /**
   * Check if this parser can merge the given parsed setup
   * @param existing The existing parsed setup
   * @returns True if merge is possible
   */
  canMerge(existing: ParsedSetup): boolean {
    // Default implementation - override for specific logic
    return existing.agents.length > 0 || 
           existing.rules.length > 0 || 
           existing.templates.length > 0;
  }

  /**
   * Merge existing setup with ai-setup templates
   * @param existing The existing parsed setup
   * @param strategy Merge strategy to use
   * @param options Migration options
   * @returns Merge result with conflicts and output
   */
  abstract merge(
    existing: ParsedSetup,
    strategy: MergeStrategy,
    options: MigrationOptions
  ): Promise<MergeResult>;

  /**
   * Get priority for this parser (higher = more specific)
   * Parsers with higher priority are checked first
   */
  getPriority(): number {
    return 100;
  }

  /**
   * Validate that the parser is properly configured
   * @returns Validation result
   */
  validate(): { valid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!this.id) errors.push('Parser ID is required');
    if (!this.name) errors.push('Parser name is required');
    if (!this.supportedPatterns?.length) errors.push('At least one supported pattern is required');
    
    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Get supported file types for this parser
   * Override to provide specific file types
   */
  getSupportedFileTypes(): string[] {
    return ['config', 'agent', 'rule', 'template', 'command'];
  }

  /**
   * Extract project metadata from files
   * Override to provide custom extraction
   */
  protected async extractMetadata(files: { path: string; content: string }[]): Promise<Record<string, unknown>> {
    const metadata: Record<string, unknown> = {};
    
    for (const file of files) {
      // Try to extract project name from common patterns
      const projectNameMatch = file.content.match(/#\s+(.+?)\s*-?\s*AI Agent/i) ||
                               file.content.match(/#\s+Project[:\s]+(.+)/i) ||
                               file.content.match(/name:\s*["']?([^"'\n]+)["']?/i);
      if (projectNameMatch && !metadata.projectName) {
        metadata.projectName = projectNameMatch[1].trim();
      }

      // Try to extract tech stack
      const stackMatch = file.content.match(/Stack[:\s]+\n?([\s\S]+?)(?=\n\n|\n#|$)/i);
      if (stackMatch && !metadata.techStack) {
        metadata.techStack = stackMatch[1].trim();
      }
    }

    return metadata;
  }
}

/**
 * Parser factory interface for creating parser instances
 */
export interface ParserFactory {
  create(): BaseParser;
  getMetadata(): {
    id: string;
    name: string;
    description: string;
    version: string;
  };
}

/**
 * Registry of all available parsers
 */
export class ParserRegistry {
  private parsers: Map<string, BaseParser> = new Map();
  private factories: Map<string, ParserFactory> = new Map();

  register(parser: BaseParser): void {
    const validation = parser.validate();
    if (!validation.valid) {
      throw new Error(
        `Parser validation failed: ${validation.errors.join(', ')}`
      );
    }
    this.parsers.set(parser.id, parser);
  }

  registerFactory(factory: ParserFactory): void {
    const metadata = factory.getMetadata();
    this.factories.set(metadata.id, factory);
  }

  get(id: string): BaseParser | undefined {
    return this.parsers.get(id);
  }

  getAll(): BaseParser[] {
    return Array.from(this.parsers.values()).sort((a, b) => b.getPriority() - a.getPriority());
  }

  async detectAll(context: MigrationContext): Promise<DetectionResult[]> {
    const results: DetectionResult[] = [];
    
    for (const parser of this.getAll()) {
      try {
        const result = await parser.detect(context);
        if (result.detected) {
          results.push(result);
        }
      } catch (error) {
        // Log but don't fail - other parsers might work
        console.warn(`Parser ${parser.id} detection failed:`, error);
      }
    }

    // Sort by confidence
    return results.sort((a, b) => b.confidence - a.confidence);
  }
}

// Global registry instance
export const globalParserRegistry = new ParserRegistry();
