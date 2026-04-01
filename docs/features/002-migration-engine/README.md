# Migration Engine Documentation

## Overview

The Migration Engine allows you to import existing AI setups from various tools into ai-setup format, preserving your customizations while gaining the benefits of ai-setup's structured approach.

## Supported Adapters

| Adapter | Detection | Migration | Status |
|---------|-----------|-----------|--------|
| OpenCode | ✅ | ✅ | Complete |
| Claude Code | ✅ | ✅ | Complete |
| Pi | ✅ | ✅ | Complete |
| Gemini CLI | ✅ | ✅ | Complete |
| GitHub Copilot | ✅ | ✅ | Complete |

## Quick Start

### Import an Existing Setup

```bash
# Detect and migrate automatically
ai-setup import

# Preview changes without executing
ai-setup import --preview

# Use a specific merge strategy
ai-setup import --strategy preserve

# Migrate from specific path
ai-setup import /path/to/project
```

### Init with Migration

```bash
# Initialize new project, migrating existing setup
ai-setup init --migrate

# Specify source path
ai-setup init --migrate --from /path/to/existing
```

### Check for Drift

```bash
# Compare current setup to clean ai-setup state
ai-setup doctor --migration-check

# Show detailed differences
ai-setup doctor --migration-check --verbose
```

## Merge Strategies

- **smart** (default): Intelligent merge with conflict markers for manual resolution
- **preserve**: Keep existing files, add new ai-setup files
- **replace**: Replace with ai-setup templates (creates backup)
- **append**: Combine existing and ai-setup content

## Architecture

The migration system uses a 3-way merge algorithm:
- **Base**: ai-setup template
- **Ours**: Your existing content
- **Theirs**: User preferences

Conflicts are marked with `<<<<< OURS`, `=====`, `>>>>> THEIRS` markers.

## Community Extensions

Create custom parsers for other AI tools:

```typescript
import { BaseParser } from '@ricardoborges-teachable/ai-setup/migration';

export class MyCustomParser extends BaseParser {
  readonly id = 'my-custom';
  readonly name = 'My Custom Tool';
  readonly version = '1.0.0';
  readonly supportedPatterns = ['.my-tool/**/*'];

  async detect(context) {
    // Detect existing setup
  }

  async parse(context) {
    // Parse into structured format
  }

  async merge(existing, strategy, options) {
    // Merge with ai-setup templates
  }
}
```

### Parser Discovery

Parsers are discovered from:
1. Project local: `./ai-setup/plugins/*/parser.ts`
2. Global user: `~/.ai-setup/parsers/*/parser.js`
3. NPM packages: `@ai-setup/parsers-*`
4. Built-in: Core library parsers

## CLI Reference

### import
```
ai-setup import [path] [options]

Options:
  -p, --preview          Preview changes without executing
  -s, --strategy        Merge strategy: smart, preserve, replace, append
  -v, --verbose         Show detailed output
  --skip-backup         Skip creating backup
  -y, --yes            Auto-confirm without prompts
```

### migrate
Alias for `import`.

### init --migrate
```
ai-setup init --migrate [--from path]
```

### doctor --migration-check
```
ai-setup doctor --migration-check [--verbose]
```
