/**
 * 3-Way Diff Algorithm
 * 
 * Implements Myers diff algorithm for line-level comparison
 * and 3-way merge with conflict detection.
 */

export interface DiffLine {
  line: string;
  index: number;
}

export interface DiffResult {
  added: DiffLine[];
  removed: DiffLine[];
  unchanged: DiffLine[];
}

export interface Diff3Result {
  merged: string[];
  conflicts: Array<{
    lineStart: number;
    lineEnd: number;
    base: string[];
    ours: string[];
    theirs: string[];
  }>;
  hasConflicts: boolean;
}

/**
 * Myers diff algorithm for finding shortest edit script
 */
export function myersDiff(oldLines: string[], newLines: string[]): DiffResult {
  const added: DiffLine[] = [];
  const removed: DiffLine[] = [];
  const unchanged: DiffLine[] = [];

  // Handle edge cases
  if (oldLines.length === 0 && newLines.length === 0) {
    return { added, removed, unchanged };
  }

  if (oldLines.length === 0) {
    newLines.forEach((line, i) => added.push({ line, index: i }));
    return { added, removed, unchanged };
  }

  if (newLines.length === 0) {
    oldLines.forEach((line, i) => removed.push({ line, index: i }));
    return { added, removed, unchanged };
  }

  // Build edit graph and find shortest path
  const { edits } = computeEdits(oldLines, newLines);

  // Process edits to build diff result
  let oldIndex = 0;
  let newIndex = 0;

  for (const edit of edits) {
    switch (edit.type) {
      case 'equal':
        unchanged.push({ line: oldLines[oldIndex], index: oldIndex });
        oldIndex++;
        newIndex++;
        break;
      case 'delete':
        removed.push({ line: oldLines[oldIndex], index: oldIndex });
        oldIndex++;
        break;
      case 'insert':
        added.push({ line: newLines[newIndex], index: newIndex });
        newIndex++;
        break;
    }
  }

  return { added, removed, unchanged };
}

interface Edit {
  type: 'equal' | 'insert' | 'delete';
  oldIndex?: number;
  newIndex?: number;
}

function computeEdits(oldLines: string[], newLines: string[]): { edits: Edit[] } {
  const edits: Edit[] = [];
  
  // Simple LCS implementation
  const oldLen = oldLines.length;
  const newLen = newLines.length;
  
  // Build LCS table
  const table: number[][] = Array(oldLen + 1).fill(null).map(() => Array(newLen + 1).fill(0));
  
  for (let i = 1; i <= oldLen; i++) {
    for (let j = 1; j <= newLen; j++) {
      if (oldLines[i - 1] === newLines[j - 1]) {
        table[i][j] = table[i - 1][j - 1] + 1;
      } else {
        table[i][j] = Math.max(table[i - 1][j], table[i][j - 1]);
      }
    }
  }
  
  // Backtrack to find edits
  let i = oldLen;
  let j = newLen;
  const tempEdits: Edit[] = [];
  
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      tempEdits.unshift({ type: 'equal', oldIndex: i - 1, newIndex: j - 1 });
      i--;
      j--;
    } else if (j > 0 && (i === 0 || table[i][j - 1] >= table[i - 1][j])) {
      tempEdits.unshift({ type: 'insert', newIndex: j - 1 });
      j--;
    } else {
      tempEdits.unshift({ type: 'delete', oldIndex: i - 1 });
      i--;
    }
  }
  
  return { edits: tempEdits };
}

/**
 * 3-way merge: base + ours + theirs
 */
export function diff3(
  base: string[],
  ours: string[],
  theirs: string[]
): Diff3Result {
  // Diff base vs ours
  const diffOurs = myersDiff(base, ours);
  
  // Diff base vs theirs
  const diffTheirs = myersDiff(base, theirs);
  
  // Apply both diffs with conflict detection
  const merged: string[] = [];
  const conflicts: Diff3Result['conflicts'] = [];
  
  let baseIndex = 0;
  let oursIndex = 0;
  let theirsIndex = 0;
  
  while (baseIndex < base.length || oursIndex < ours.length || theirsIndex < theirs.length) {
    // Check what changes are happening at this position
    const oursChanged = isChangedAt(diffOurs, baseIndex);
    const theirsChanged = isChangedAt(diffTheirs, baseIndex);
    
    if (!oursChanged && !theirsChanged) {
      // No changes - use base
      if (baseIndex < base.length) {
        merged.push(base[baseIndex]);
        baseIndex++;
      }
      oursIndex++;
      theirsIndex++;
    } else if (oursChanged && !theirsChanged) {
      // Only ours changed - use ours
      if (oursIndex < ours.length) {
        merged.push(ours[oursIndex]);
      }
      baseIndex++;
      oursIndex++;
      theirsIndex++;
    } else if (!oursChanged && theirsChanged) {
      // Only theirs changed - use theirs
      if (theirsIndex < theirs.length) {
        merged.push(theirs[theirsIndex]);
      }
      baseIndex++;
      oursIndex++;
      theirsIndex++;
    } else {
      // Both changed - conflict!
      const conflictStart = merged.length;
      
      // Add conflict markers
      merged.push('\u003c\u003c\u003c\u003c\u003c OURS');
      
      // Add our changes
      while (oursIndex < ours.length && isChangedAt(diffOurs, baseIndex)) {
        merged.push(ours[oursIndex]);
        oursIndex++;
      }
      
      merged.push('=====');
      
      // Add their changes
      while (theirsIndex < theirs.length && isChangedAt(diffTheirs, baseIndex)) {
        merged.push(theirs[theirsIndex]);
        theirsIndex++;
      }
      
      merged.push('\u003e\u003e\u003e\u003e\u003e THEIRS');
      
      conflicts.push({
        lineStart: conflictStart + 1,
        lineEnd: merged.length,
        base: base.slice(baseIndex, baseIndex + 1),
        ours: ours.slice(oursIndex - 1, oursIndex),
        theirs: theirs.slice(theirsIndex - 1, theirsIndex),
      });
      
      baseIndex++;
    }
  }
  
  return {
    merged,
    conflicts,
    hasConflicts: conflicts.length > 0,
  };
}

function isChangedAt(diff: DiffResult, index: number): boolean {
  return diff.added.some(l => l.index === index) || 
         diff.removed.some(l => l.index === index);
}

/**
 * Simple 2-way merge: base + changes
 */
export function merge2Way(base: string[], changes: string[]): string[] {
  const result: string[] = [];
  const diff = myersDiff(base, changes);
  
  let baseIndex = 0;
  let changesIndex = 0;
  
  while (baseIndex < base.length || changesIndex < changes.length) {
    const isDeleted = diff.removed.some(l => l.index === baseIndex);
    const isAdded = diff.added.some(l => l.index === changesIndex);
    
    if (isAdded) {
      result.push(changes[changesIndex]);
      changesIndex++;
    } else if (!isDeleted) {
      result.push(base[baseIndex]);
      baseIndex++;
      changesIndex++;
    } else {
      baseIndex++;
    }
  }
  
  return result;
}

/**
 * Check if a merge has conflicts
 */
export function hasConflicts(merged: string[]): boolean {
  return merged.some(line => 
    line.includes('\u003c\u003c\u003c\u003c\u003c') || 
    line.includes('=====') || 
    line.includes('\u003e\u003e\u003e\u003e\u003e')
  );
}

/**
 * Resolve conflicts by choosing a side
 */
export function resolveConflicts(
  merged: string[],
  resolution: 'ours' | 'theirs' | 'base'
): string[] {
  const result: string[] = [];
  let inConflict = false;
  let conflictSide: 'ours' | 'theirs' | 'base' | null = null;
  
  for (const line of merged) {
    if (line.startsWith('\u003c\u003c\u003c\u003c\u003c')) {
      inConflict = true;
      conflictSide = 'ours';
    } else if (line.startsWith('=====')) {
      conflictSide = 'theirs';
    } else if (line.startsWith('\u003e\u003e\u003e\u003e\u003e')) {
      inConflict = false;
      conflictSide = null;
    } else if (inConflict) {
      if (conflictSide === resolution) {
        result.push(line);
      }
      // Otherwise skip
    } else {
      result.push(line);
    }
  }
  
  return result;
}
