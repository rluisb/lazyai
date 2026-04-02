# Community Parser Guide

## Overview

Create custom parsers to support any AI coding assistant tool with the ai-setup migration engine.

## Parser Interface

```typescript
import { BaseParser } from '@ricardoborges-teachable/ai-setup/migration';
import {
  MigrationContext,
  DetectionResult,
  ParsedSetup,
  MergeResult,
  MergeStrategy,
  MigrationOptions,
} from '@ricardoborges-teachable/ai-setup/migration';

export class MyParser extends BaseParser {
  // Required properties
  readonly id = 'my-tool';                    // Unique identifier
  readonly name = 'My Tool';                  // Display name
  readonly description = 'Parser for My Tool'; // Description
  readonly version = '1.0.0';                 // Parser version
  readonly supportedPatterns = [              // File patterns to detect
    '.my-tool/**/*',
    'MY-TOOL.md',
  ];

  // Detection logic
  async detect(context: MigrationContext): Promise<DetectionResult> {
    return {
      detected: true,                           // Found setup?
      confidence: 0.95,                        // 0-1 confidence score
      adapterId: this.id,
      adapterName: this.name,
      files: [
        { path: 'MY-TOOL.md', type: 'config', priority: 100 },
      ],
      metadata: {},
    };
  }

  // Parse existing setup
  async parse(context: MigrationContext): Promise<ParsedSetup> {
    return {
      projectName: 'My Project',
      description: 'Project description',
      agents: [],       // Agent definitions
      rules: [],        // Rules
      commands: [],     // Commands/skills
      templates: [],    // Templates
      customSections: [], // Custom content
      files: [],        // All parsed files
      metadata: {},     // Additional metadata
    };
  }

  // Merge with ai-setup templates
  async merge(
    existing: ParsedSetup,
    strategy: MergeStrategy,
    options: MigrationOptions
  ): Promise<MergeResult> {
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

export default MyParser;
```

## Distribution Options

### Option 1: Project Local

Place parser in your project:

```
./ai-setup/plugins/my-tool/parser.ts
```

### Option 2: Global User

Install parser globally:

```
~/.ai-setup/parsers/my-tool/parser.js
```

### Option 3: NPM Package

Publish as scoped package:

```json
{
  "name": "@ai-setup/parsers-my-tool",
  "version": "1.0.0",
  "main": "dist/parser.js",
  "peerDependencies": {
    "@ricardoborges-teachable/ai-setup": "^0.1.0"
  }
}
```

Install:
```bash
npm install -g @ai-setup/parsers-my-tool
```

## Testing Your Parser

```typescript
import { describe, it, expect } from 'vitest';
import { MyParser } from './parser.js';

describe('MyParser', () => {
  const parser = new MyParser();

  it('should detect existing setup', async () => {
    const context = {
      sourcePath: '/path/to/project',
      targetPath: '/output',
      options: { mergeStrategy: 'smart' as const },
    };

    const result = await parser.detect(context);
    expect(result.detected).toBe(true);
  });

  it('should parse setup correctly', async () => {
    // ... test parse
  });

  it('should merge successfully', async () => {
    // ... test merge
  });
});
```

## Best Practices

1. **Detection Confidence**: Return realistic confidence scores (0-1)
2. **Error Handling**: Handle missing files gracefully
3. **Backup**: Always backup before modifying
4. **Conflict Markers**: Use standard conflict markers when conflicts occur
5. **Documentation**: Document any custom behavior

## Parser Template

Use the built-in template generator:

```bash
# Coming soon
ai-setup create parser --name my-tool
```

Or copy the reference implementation from:
- `src/migration/parsers/opencode-parser.ts` (most complete)
