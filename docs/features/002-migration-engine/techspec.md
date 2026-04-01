---
title: Migration Engine Technical Specification
status: draft
date: 2026-04-01
---

# Migration Engine - Technical Specification

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    MIGRATION ENGINE                         │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   DETECTOR   │  │    PARSER    │  │   MERGER     │      │
│  │              │  │              │  │              │      │
│  │ Scan existing│  │ Extract data │  │ 3-way merge  │      │
│  │ files        │  │ from files   │  │              │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │               │
│         └──────────────────┼──────────────────┘             │
│                            ▼                                │
│              ┌─────────────────────────────┐                 │
│              │    MIGRATION PLAN           │                 │
│              │  (what to do, conflicts)    │                 │
│              └──────────────┬────────────────┘                 │
│                            │                                 │
│         ┌──────────────────┼──────────────────┐              │
│         ▼                  ▼                  ▼              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    EXECUTE   │  │   GENERATE   │  │   VALIDATE   │      │
│  │              │  │              │  │              │      │
│  │ Backup       │  │ Create merged│  │ Verify       │      │
│  │ Transform    │  │ files        │  │ integrity    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │  REGISTRY SYSTEM │
              │                 │
              │  parsers/        │
              │    ├── opencode.ts│
              │    ├── claude.ts │
              │    ├── custom/   │
              │    └── ...       │
              │                 │
              └─────────────────┘
```

## Directory Structure

```
src/
└── migration/
    ├── index.ts                    # Main migration orchestrator
    ├── detector.ts                 # File scanning and detection
    ├── merger.ts                   # 3-way merge logic
    ├── plan.ts                     # Migration plan generation
    ├── executor.ts                 # Execute the migration
    ├── registry/
    │   ├── index.ts                # Registry and loader
    │   ├── discovery.ts            # Parser discovery from npm/global/local
    │   └── types.ts                # Registry interfaces
    ├── parsers/                    # Parser implementations
    │   ├── base-parser.ts          # Abstract base class
    │   ├── opencode-parser.ts      # OpenCode specific
    │   ├── claude-parser.ts        # Claude Code specific
    │   ├── pi-parser.ts            # Pi specific
    │   ├── gemini-parser.ts        # Gemini CLI specific
    │   └── copilot-parser.ts       # GitHub Copilot specific
    └── diff/                       # Diff utilities
        └── diff3.ts                # 3-way diff algorithm
```

## Core Interfaces

### Parser Interface
```typescript
export interface MigrationParser {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly supportedFormats: string[];
  
  detect(context: MigrationContext): DetectionResult;
  parse(context: MigrationContext): Promise<ParsedSetup>;
  canMerge(existing: ParsedSetup, target: AiSetupTemplate): boolean;
  merge(existing: ParsedSetup, target: AiSetupTemplate): MergeResult;
}
```

### Parsed Setup
```typescript
export interface ParsedSetup {
  projectName?: string;
  techStack?: TechStack;
  agents: AgentDefinition[];
  rules: RuleDefinition[];
  commands: CommandDefinition[];
  templates: TemplateDefinition[];
  customSections: CustomSection[];
  metadata: Record<string, unknown>;
}
```

### Merge Result
```typescript
export interface MergeResult {
  merged: boolean;
  conflicts: MergeConflict[];
  backupPaths: string[];
  newFiles: string[];
  modifiedFiles: string[];
  warnings: string[];
}
```

## Parser Discovery Algorithm

1. **Scan project local**: `./ai-setup/plugins/*/parser.ts`
2. **Load global user**: `~/.ai-setup/parsers/*/parser.js`
3. **Resolve npm packages**: `@ai-setup/parsers-*`
4. **Fallback to built-in**: `src/migration/parsers/*.ts`

## 3-Way Merge Algorithm

Uses Myers diff algorithm for line-level comparison:
- Input: Base (template), Ours (existing), Theirs (user prefs)
- Output: Merged file with conflict markers where necessary
- Conflict markers: `<<<<<`, `=====`, `>>>>>`

## CLI Commands

### import
```bash
ai-setup import [path] [--preview] [--merge-strategy=smart|preserve|replace|append]
```

### init --migrate
```bash
ai-setup init --migrate [--from=/path/to/existing]
```

### doctor --migration-check
```bash
ai-setup doctor --migration-check [--verbose]
```

## Testing Strategy

1. **Unit Tests**: Each parser, merger, detector
2. **Integration Tests**: Full migration scenarios
3. **Conflict Tests**: Complex merge scenarios
4. **Sample Setups**: Real-world existing setups to migrate

## Performance Considerations

- Lazy parser loading
- Caching of discovered parsers
- Incremental diff computation
- Parallel file processing

## Security

- Validate parser signatures (if from npm)
- Sandboxed parser execution
- No arbitrary code execution from parsers
