import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, assert, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@clack/prompts', () => ({
  select: vi.fn(),
  text: vi.fn(),
  note: vi.fn(),
  isCancel: vi.fn(() => false),
}));

import * as p from '@clack/prompts';
import { executeMigrationPlan } from '../executor.js';
import type { MigrationContext, MigrationPlan } from '../types.js';

describe('executeMigrationPlan interactive conflicts', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'ai-setup-executor-'));
    vi.clearAllMocks();
    vi.mocked(p.isCancel).mockReturnValue(false);
  });

  afterEach(async () => {
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  it('resolves conflicts interactively by keeping the existing content', async () => {
    vi.mocked(p.select).mockResolvedValue('keep-existing');

    const result = await executeMigrationPlan(createContext(tempDir, true), createPlan(tempDir));

    assert(result.plan !== null);
    expect(p.note).toHaveBeenCalledWith(
      expect.stringContaining('--- Existing\ncurrent guidance'),
      'Conflict in AGENTS.md:10-12',
    );
    expect(result.plan.conflicts[0]?.resolved).toBe(true);
    expect(result.plan.conflicts[0]?.resolution).toBe('keep-existing');
    expect(result.plan.conflicts[0]?.resolvedContent).toBe('current guidance');
    expect(result.stats.conflictsResolved).toBe(1);
    expect(result.stats.conflictsUnresolved).toBe(0);
    expect(result.plan.canProceed).toBe(true);
  });

  it('supports manual edits during conflict resolution', async () => {
    vi.mocked(p.select).mockResolvedValue('manual-edit');
    vi.mocked(p.text).mockResolvedValue('merged guidance');

    const result = await executeMigrationPlan(createContext(tempDir, true), createPlan(tempDir));

    assert(result.plan !== null);
    expect(p.text).toHaveBeenCalled();
    expect(result.plan.conflicts[0]?.resolution).toBe('manual-edit');
    expect(result.plan.conflicts[0]?.resolvedContent).toBe('merged guidance');
  });
});

function createContext(targetPath: string, interactive: boolean): MigrationContext {
  return {
    sourcePath: targetPath,
    targetPath,
    options: {
      mergeStrategy: 'smart',
      skipBackup: true,
      interactive,
    },
  };
}

function createPlan(targetPath: string): MigrationPlan {
  return {
    sourcePath: targetPath,
    targetPath,
    adapters: ['opencode'],
    actions: [],
    conflicts: [
      {
        file: 'AGENTS.md',
        lineStart: 10,
        lineEnd: 12,
        baseContent: 'base guidance',
        oursContent: 'current guidance',
        theirsContent: 'incoming guidance',
      },
    ],
    adapterConflicts: [],
    estimatedFiles: 0,
    estimatedConflicts: 1,
    canProceed: true,
  };
}
