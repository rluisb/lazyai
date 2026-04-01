import { describe, it, expect } from 'vitest';
import { myersDiff, diff3, merge2Way, hasConflicts, resolveConflicts } from '../diff/diff3.js';

describe('Myers Diff', () => {
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
    expect(result.added[0].line).toBe('line3');
    expect(result.removed).toHaveLength(0);
  });

  it('should detect removed lines', () => {
    const base = ['line1', 'line2', 'line3'];
    const changed = ['line1', 'line3'];
    const result = myersDiff(base, changed);
    
    expect(result.removed).toHaveLength(1);
    expect(result.removed[0].line).toBe('line2');
    expect(result.added).toHaveLength(0);
  });
});

describe('3-Way Diff', () => {
  it('should merge cleanly when only base changes', () => {
    const base = ['line1', 'line2'];
    const ours = ['line1', 'line2', 'line3'];
    const theirs = ['line1', 'line2', 'line3'];
    
    const result = diff3(base, ours, theirs);
    
    expect(result.hasConflicts).toBe(false);
    expect(result.conflicts).toHaveLength(0);
  });

  it('should handle conflict when both sides change same line', () => {
    const base = ['line1', 'line2'];
    const ours = ['line1', 'ours-change'];
    const theirs = ['line1', 'theirs-change'];
    
    const result = diff3(base, ours, theirs);
    
    expect(result.hasConflicts).toBe(true);
  });
});

describe('Conflict Resolution', () => {
  it('should detect conflict markers', () => {
    const merged = [
      'line1',
      '<<<<< OURS',
      'ours-line',
      '=====',
      'theirs-line',
      '>>>>> THEIRS',
      'line2'
    ];
    
    expect(hasConflicts(merged)).toBe(true);
  });
});
