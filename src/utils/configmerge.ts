/**
 * Deep-merge read-modify-write helpers for shared config files
 * (currently `opencode.jsonc`). A `.bak` sidecar is created the first
 * time an existing file is touched so users can restore their original
 * config; subsequent runs never overwrite the existing `.bak`.
 *
 * Semantics (mirrors Go's `internal/configmerge`):
 *   - Maps are recursively merged; the patch wins on leaf collisions.
 *   - Arrays are replaced wholesale (not concatenated), because list
 *     concatenation creates duplicates users cannot easily remove.
 *   - Idempotent: re-running with the same patch produces identical
 *     bytes (keys sorted alphabetically at every depth).
 *
 * Ported from `internal/configmerge/configmerge.go`. JSON-only — Go's
 * TOML branch is unused after the OpenCode-only consolidation (spec 019).
 */

import { copyFile, fileExists, readFile, writeFile } from './files.js'
import { parseJsonc } from './jsonc.js'

export interface MergeResult {
  backupPath: string | null
}

/**
 * Deep-merge `patch` into the existing JSON/JSONC file at `path`, writing a
 * `.bak` sidecar on first touch. If the file does not exist, `patch` is
 * written directly and `backupPath` is null.
 */
export function mergeJsonFile(path: string, patch: Record<string, unknown>): MergeResult {
  let existing: Record<string, unknown> = {}
  let backupPath: string | null = null

  if (fileExists(path)) {
    existing = parseJsonc(readFile(path))
    backupPath = ensureBackup(path)
  }

  const merged = deepMerge(existing, patch)
  writeFile(path, marshalSortedJson(merged))
  return { backupPath }
}

/**
 * Recursively overlay `patch` onto `base`. Plain-object values recurse;
 * everything else (primitives, arrays) takes the patch value wholesale.
 */
export function deepMerge(
  base: Record<string, unknown>,
  patch: Record<string, unknown>,
): Record<string, unknown> {
  const out: Record<string, unknown> = { ...base }
  for (const [key, patchVal] of Object.entries(patch)) {
    const baseVal = out[key]
    if (isPlainObject(baseVal) && isPlainObject(patchVal)) {
      out[key] = deepMerge(baseVal, patchVal)
    } else {
      out[key] = patchVal
    }
  }
  return out
}

function ensureBackup(path: string): string {
  const bak = `${path}.bak`
  if (fileExists(bak)) return bak
  copyFile(path, bak)
  return bak
}

/**
 * Emit a deterministic JSON representation: keys sorted at every depth,
 * 2-space indentation, trailing newline. Matches Go's `marshalSortedJSON`
 * byte-for-byte on equivalent input.
 */
export function marshalSortedJson(value: unknown): string {
  return `${JSON.stringify(value, sortedKeyReplacer(), 2)}\n`
}

function sortedKeyReplacer(): (this: unknown, key: string, value: unknown) => unknown {
  return function (_key, value) {
    if (isPlainObject(value)) {
      const sorted: Record<string, unknown> = {}
      for (const k of Object.keys(value).sort()) {
        sorted[k] = value[k]
      }
      return sorted
    }
    return value
  }
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (value === null || typeof value !== 'object') return false
  if (Array.isArray(value)) return false
  const proto = Object.getPrototypeOf(value)
  return proto === Object.prototype || proto === null
}
