import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { writeToCanonical } from '../migration/canonical-writer.js';
import { diff3, hasConflicts, myersDiff } from '../migration/diff/diff3.js';
import { detectAdapters, formatPlan, importSetup, migrationCheck } from '../migration/index.js';
import type { BaseParser } from '../migration/parsers/base-parser.js';
import { generateMigrationPlan } from '../migration/plan.js';
import type {
  DetectionResult,
  MigrationContext,
  ParsedSetup,
} from '../migration/types.js';

describe('Migration Engine', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'ai-setup-test-'));
  });

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  describe('Detection', () => {
    it('should detect OpenCode setup', async () => {
      // Create AGENTS.md
      await fs.writeFile(
        path.join(tempDir, 'AGENTS.md'),
        '# Test Project\n\nOverview here.\n'
      );
      
      // Create .opencode directory
      await fs.mkdir(path.join(tempDir, '.opencode'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.opencode', 'agents'), { recursive: true });
      await fs.writeFile(
        path.join(tempDir, '.opencode', 'agents', 'builder.md'),
        '# Builder Agent\n\nBuilds things.\n'
      );

      const adapters = await detectAdapters(tempDir);
      
      expect(adapters).toContain('opencode');
    });

    it('should detect Claude Code setup', async () => {
      await fs.writeFile(
        path.join(tempDir, 'CLAUDE.md'),
        '# Test Project\n\nOverview here.\n'
      );
      
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
      await fs.writeFile(
        path.join(tempDir, '.claude', 'planner.md'),
        '# Planner Agent\n\nPlans things.\n'
      );

      const adapters = await detectAdapters(tempDir);
      
      expect(adapters).toContain('claude-code');
    });

    it('should return empty array for clean directory', async () => {
      const adapters = await detectAdapters(tempDir);
      expect(adapters).toHaveLength(0);
    });
  });

  describe('Diff Algorithm', () => {
    it('should detect no changes when files are identical', () => {
      const base = ['line1', 'line2', 'line3'];
      const result = myersDiff(base, base);
      
      expect(result.added).toHaveLength(0);
      expect(result.removed).toHaveLength(0);
      expect(result.unchanged).toHaveLength(3);
    });

    it('should detect added lines', () => {
      const base = ['line1', 'line2'];
      const changed = ['line1', 'line2', 'line3'];
      const result = myersDiff(base, changed);
      
      expect(result.added).toHaveLength(1);
      expect(result.removed).toHaveLength(0);
    });

    it('should handle 3-way merge conflicts', () => {
      const base = ['line1', 'line2'];
      const ours = ['line1', 'ours-change'];
      const theirs = ['line1', 'theirs-change'];
      
      const result = diff3(base, ours, theirs);
      
      expect(result.hasConflicts).toBe(true);
    });

    it('should merge cleanly when no conflicts', () => {
      const base = ['line1', 'line2'];
      const ours = ['line1', 'line2', 'line3'];
      const theirs = ['line1', 'line2', 'line3'];
      
      const result = diff3(base, ours, theirs);
      
      expect(result.hasConflicts).toBe(false);
    });

    it('should detect conflict markers', () => {
      const merged = [
        'line1',
        '<<<<< OURS',
        'ours-line',
        '=====',
        '>>>>> THEIRS'
      ];
      
      expect(hasConflicts(merged)).toBe(true);
    });
  });

  describe('Import', () => {
    it('should fail when no setup detected', async () => {
      const result = await importSetup({
        path: tempDir,
        preview: true,
        mergeStrategy: 'smart',
      });

      expect(result.success).toBe(false);
      expect(result.errors[0]).toContain('No existing AI setup detected');
    });

    it('should generate preview without executing', async () => {
      // Create OpenCode setup
      await fs.writeFile(
        path.join(tempDir, 'AGENTS.md'),
        '# Test Project\n\nOverview.\n'
      );
      
      const result = await importSetup({
        path: tempDir,
        preview: true,
        mergeStrategy: 'smart',
      });

      // Should succeed in preview mode even if migration can't proceed
      // because it shows the plan
      expect(result.plan).toBeDefined();
      expect(result.stats.filesCreated).toBeGreaterThanOrEqual(0);
    });

    it('formats migration plans with a readable summary', () => {
      const formatted = formatPlan({
        sourcePath: '/source',
        targetPath: '/target',
        adapters: ['opencode'],
        actions: [
          { type: 'create', targetPath: 'AGENTS.md', description: 'Create AGENTS.md', reason: 'new file' },
          { type: 'backup', sourcePath: 'AGENTS.md', targetPath: '.ai-setup-backup/AGENTS.md', description: 'Backup AGENTS.md', reason: 'safe backup' },
        ],
        conflicts: [
          {
            file: 'AGENTS.md',
            lineStart: 10,
            lineEnd: 12,
            baseContent: 'base',
            oursContent: 'ours',
            theirsContent: 'theirs',
          },
        ],
        estimatedFiles: 2,
        estimatedConflicts: 1,
        canProceed: false,
      });

      expect(formatted).toContain('Migration plan');
      expect(formatted).toContain('Detected adapters: OpenCode');
      expect(formatted).toContain('Summary:');
      expect(formatted).toContain('Unresolved conflicts: 1');
      expect(formatted).toContain('Status: blocked until conflicts are resolved');
    });
  });

  describe('Drift Check', () => {
    it('should detect clean state', async () => {
      const result = await migrationCheck({
        path: tempDir,
      });

      // No .ai-setup.json means not managed by ai-setup
      expect(result.clean).toBe(false);
      expect(result.missingFiles).toContain('.ai-setup.json');
    });

    it('should detect extra files', async () => {
      // Create a file that looks like an AI setup file
      await fs.mkdir(path.join(tempDir, 'docs'), { recursive: true });
      await fs.writeFile(
        path.join(tempDir, 'AGENTS.md'),
        '# Test\n'
      );

      const result = await migrationCheck({
        path: tempDir,
      });

      // Should detect AGENTS.md as extra since no .ai-setup.json
      expect(result.extraFiles).toContain('AGENTS.md');
    });
  });
});

describe('Parser Registry', () => {
  it('should detect all 5 adapters', async () => {
    const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'parser-test-'));
    
    // Create a multi-adapter setup
    await fs.writeFile(path.join(tempDir, 'AGENTS.md'), '# Test\n');
    await fs.writeFile(path.join(tempDir, 'CLAUDE.md'), '# Test\n');
    await fs.writeFile(path.join(tempDir, 'GEMINI.md'), '# Test\n');
    
    await fs.mkdir(path.join(tempDir, '.opencode'), { recursive: true });
    await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
    await fs.mkdir(path.join(tempDir, '.gemini'), { recursive: true });
    await fs.mkdir(path.join(tempDir, '.pi'), { recursive: true });
    await fs.mkdir(path.join(tempDir, '.github'), { recursive: true });
    
    await fs.writeFile(
      path.join(tempDir, '.github', 'copilot-instructions.md'),
      '# Instructions\n'
    );

    const adapters = await detectAdapters(tempDir);
    
    expect(adapters).toContain('opencode');
    expect(adapters).toContain('claude-code');
    expect(adapters).toContain('gemini');
    expect(adapters).toContain('pi');
    expect(adapters).toContain('copilot');
    
    await fs.rm(tempDir, { recursive: true, force: true });
  });
});

describe('generateMigrationPlan', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'plan-test-'));
  });

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  it('should return empty plan for no detections', async () => {
    const context: MigrationContext = {
      sourcePath: tempDir,
      targetPath: tempDir,
      options: { mergeStrategy: 'smart' },
    };

    const plan = await generateMigrationPlan(context, [], []);

    expect(plan.adapters).toHaveLength(0);
    expect(plan.actions).toHaveLength(0);
    expect(plan.conflicts).toHaveLength(0);
    expect(plan.canProceed).toBe(true);
  });

  it('should backup files when no parser matches', async () => {
    const context: MigrationContext = {
      sourcePath: tempDir,
      targetPath: tempDir,
      options: { mergeStrategy: 'smart' },
    };

    const detections: DetectionResult[] = [
      {
        detected: true,
        confidence: 0.9,
        adapterId: 'unknown-adapter',
        adapterName: 'Unknown',
        files: [
          { path: 'config.yml', type: 'config', priority: 1 },
        ],
      },
    ];

    const plan = await generateMigrationPlan(context, detections, []);

    expect(plan.adapters).toContain('unknown-adapter');
    expect(plan.actions).toHaveLength(1);
    const action = plan.actions[0];
    expect(action).toBeDefined();
    expect(action!.type).toBe('backup');
    expect(action!.sourcePath).toBe('config.yml');
  });

  it('should set canProceed false when smart strategy has unresolved conflicts', async () => {
    const context: MigrationContext = {
      sourcePath: tempDir,
      targetPath: tempDir,
      options: { mergeStrategy: 'smart' },
    };

    const mockParser = {
      id: 'test-adapter',
      parse: async () => ({
        agents: [],
        rules: [],
        commands: [],
        templates: [],
        customSections: [],
        files: [],
        metadata: {},
      }),
      canMerge: () => true,
      merge: async () => ({
        success: true,
        merged: true,
        conflicts: [
          {
            file: 'AGENTS.md',
            lineStart: 1,
            lineEnd: 5,
            baseContent: 'base',
            oursContent: 'ours',
            theirsContent: 'theirs',
          },
        ],
        backupPaths: [],
        newFiles: [],
        modifiedFiles: [],
        warnings: [],
      }),
    } as unknown as BaseParser;

    const detections: DetectionResult[] = [
      {
        detected: true,
        confidence: 1,
        adapterId: 'test-adapter',
        adapterName: 'Test',
        files: [],
      },
    ];

    const plan = await generateMigrationPlan(context, detections, [mockParser]);

    expect(plan.conflicts).toHaveLength(1);
    expect(plan.estimatedConflicts).toBe(1);
    expect(plan.canProceed).toBe(false);
  });

  it('should allow proceed with replace strategy even with conflicts', async () => {
    const context: MigrationContext = {
      sourcePath: tempDir,
      targetPath: tempDir,
      options: { mergeStrategy: 'replace' },
    };

    const mockParser = {
      id: 'test-adapter',
      parse: async () => ({
        agents: [],
        rules: [],
        commands: [],
        templates: [],
        customSections: [],
        files: [],
        metadata: {},
      }),
      canMerge: () => true,
      merge: async () => ({
        success: true,
        merged: true,
        conflicts: [
          {
            file: 'AGENTS.md',
            lineStart: 1,
            lineEnd: 5,
            baseContent: 'base',
            oursContent: 'ours',
            theirsContent: 'theirs',
          },
        ],
        backupPaths: [],
        newFiles: ['new.md'],
        modifiedFiles: [],
        warnings: [],
      }),
    } as unknown as BaseParser;

    const detections: DetectionResult[] = [
      {
        detected: true,
        confidence: 1,
        adapterId: 'test-adapter',
        adapterName: 'Test',
        files: [],
      },
    ];

    const plan = await generateMigrationPlan(context, detections, [mockParser]);

    expect(plan.canProceed).toBe(true);
    expect(plan.actions.some(a => a.type === 'create')).toBe(true);
  });
});

describe('writeToCanonical', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'canonical-test-'));
  });

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  it('should write agents to .ai/agents/', async () => {
    const parsed: ParsedSetup = {
      agents: [
        { id: 'builder', name: 'Builder', description: 'Builds', role: 'builder', content: '# Builder\nBuilds things.', sourcePath: 'agents/builder.md' },
      ],
      rules: [],
      commands: [],
      templates: [],
      customSections: [],
      files: [],
      metadata: {},
    };

    const fileRecords: { path: string; hash: string; source: string }[] = [];
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: parsed,
      fileRecords,
    });

    expect(result.agents).toHaveLength(1);
    expect(result.agents[0]).toBe(path.join('.ai', 'agents', 'builder.md'));

    const content = await fs.readFile(path.join(tempDir, '.ai', 'agents', 'builder.md'), 'utf-8');
    expect(content).toContain('# Builder');
  });

  it('should skip existing files without overwriting', async () => {
    await fs.mkdir(path.join(tempDir, '.ai', 'agents'), { recursive: true });
    await fs.writeFile(path.join(tempDir, '.ai', 'agents', 'builder.md'), 'existing content');

    const parsed: ParsedSetup = {
      agents: [
        { id: 'builder', name: 'Builder', description: 'Builds', role: 'builder', content: 'new content', sourcePath: 'agents/builder.md' },
      ],
      rules: [],
      commands: [],
      templates: [],
      customSections: [],
      files: [],
      metadata: {},
    };

    const fileRecords: { path: string; hash: string; source: string }[] = [];
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: parsed,
      fileRecords,
    });

    expect(result.agents).toHaveLength(0);
    expect(result.skipped).toContain(path.join('.ai', 'agents', 'builder.md'));

    const content = await fs.readFile(path.join(tempDir, '.ai', 'agents', 'builder.md'), 'utf-8');
    expect(content).toBe('existing content');
  });

  it('should handle dryRun without writing files', async () => {
    const parsed: ParsedSetup = {
      agents: [
        { id: 'test', name: 'Test', description: 'Test', role: 'test', content: '# Test', sourcePath: 'test.md' },
      ],
      rules: [],
      commands: [],
      templates: [],
      customSections: [],
      files: [],
      metadata: {},
    };

    const fileRecords: { path: string; hash: string; source: string }[] = [];
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: parsed,
      fileRecords,
      dryRun: true,
    });

    expect(result.agents).toHaveLength(1);
    expect(fileRecords).toHaveLength(0);

    const dirExists = await fs.access(path.join(tempDir, '.ai')).then(() => true).catch(() => false);
    expect(dirExists).toBe(false);
  });

  it('should write skills, prompts, and rules to correct subdirs', async () => {
    const parsed: ParsedSetup = {
      agents: [],
      rules: [
        { id: 'no-any', category: 'typescript', content: 'No any allowed', sourcePath: 'rules/ts.md', priority: 1 },
      ],
      commands: [
        { id: 'deploy', name: 'Deploy', description: 'Deploy to prod', content: '# Deploy cmd', sourcePath: 'commands/deploy.md' },
      ],
      templates: [
        { id: 'pr', name: 'PR Template', description: 'PR template', content: '## PR', sourcePath: 'templates/pr.md' },
      ],
      customSections: [],
      files: [],
      metadata: {},
    };

    const fileRecords: { path: string; hash: string; source: string }[] = [];
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: parsed,
      fileRecords,
    });

    expect(result.rules).toHaveLength(1);
    expect(result.skills).toHaveLength(1);
    expect(result.prompts).toHaveLength(1);

    expect(result.rules[0]).toContain(path.join('.ai', 'rules'));
    expect(result.skills[0]).toContain(path.join('.ai', 'skills'));
    expect(result.prompts[0]).toContain(path.join('.ai', 'prompts'));
  });

  it('should write customSections as constitution file', async () => {
    const parsed: ParsedSetup = {
      agents: [],
      rules: [],
      commands: [],
      templates: [],
      customSections: [
        { id: 'overview', title: 'Project Overview', content: 'This is the project.', sourcePath: 'AGENTS.md' },
        { id: 'stack', title: 'Tech Stack', content: 'TypeScript + Node', sourcePath: 'AGENTS.md' },
      ],
      files: [],
      metadata: { adapter: 'opencode' },
    };

    const fileRecords: { path: string; hash: string; source: string }[] = [];
    const result = await writeToCanonical({
      targetDir: tempDir,
      parsedSetup: parsed,
      fileRecords,
    });

    expect(result.rootConfig).not.toBeNull();
    expect(result.rootConfig).toContain('opencode.md');

    const content = await fs.readFile(
      path.join(tempDir, result.rootConfig ?? ''),
      'utf-8'
    );
    expect(content).toContain('## Project Overview');
    expect(content).toContain('## Tech Stack');
  });
});
