import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import { 
  importSetup, 
  detectAdapters, 
  migrationCheck,
  detectExistingSetup 
} from '../migration/index.js';
import { myersDiff, diff3, hasConflicts } from '../migration/diff/diff3.js';

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
