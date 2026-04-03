import { dirname } from 'node:path'
import type { AiSetupConfig } from '../types.js'
import { extractSelections } from '../utils/manifest.js'
import { Errors } from '../errors/index.js'
import { CURRENT_SCHEMA_VERSION, type StoreData, storeDataSchema } from './schema.js'

function asRecord(data: unknown): Record<string, unknown> | null {
  return typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : null
}

function asString(value: unknown, fallback: string): string {
  return typeof value === 'string' && value.length > 0 ? value : fallback
}

function asStringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : []
}

function ensureError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value))
}

function isLegacyFileRecord(value: unknown): value is { path: string; hash: string; source: string } {
  const record = asRecord(value)
  return Boolean(
    record && typeof record.path === 'string' && typeof record.hash === 'string' && typeof record.source === 'string',
  )
}

function normalizeSelections(
  selections: Partial<StoreData['selections']> | undefined,
): StoreData['selections'] {
  return {
    templates: selections?.templates ?? [],
    rules: selections?.rules ?? [],
    agents: selections?.agents ?? [],
    skills: selections?.skills ?? [],
    prompts: selections?.prompts ?? [],
    infra: selections?.infra ?? [],
    constitution: selections?.constitution ?? [],
  }
}

export function isLegacyFormat(data: unknown): boolean {
  const parsed = asRecord(data)
  if (!parsed) return false

  const hasVersion = typeof parsed.version === 'string'
  const hasLegacyFiles = Array.isArray(parsed.files) && parsed.files.every(isLegacyFileRecord)

  const meta = asRecord(parsed.meta)
  const hasV1SchemaVersion = typeof meta?.schemaVersion === 'number'

  return hasVersion && hasLegacyFiles && !hasV1SchemaVersion
}

export function migrateV0toV1(data: Record<string, unknown>, targetDir: string = process.cwd()): StoreData {
  const now = new Date().toISOString()
  const installedAt = asString(data.installedAt, now)

  const legacyFiles = Array.isArray(data.files) ? data.files.filter(isLegacyFileRecord) : []

  const rawSetupType = asString(data.setupType, 'project')
  const setupScope = rawSetupType === 'workspace' ? 'workspace' : 'project'

  const config = {
    setupScope,
    setupType: rawSetupType,
    tools: asStringArray(data.tools),
    projectName: asString(data.projectName, ''),
    targetDir,
  }

  const legacyManifest = {
    version: asString(data.version, '0.0.0'),
    setupScope: config.setupScope,
    setupType: config.setupType,
    tools: config.tools,
    projectName: config.projectName,
    installedAt,
    files: legacyFiles,
    ...(asRecord(data.selections) ? { selections: data.selections } : {}),
  } as unknown as AiSetupConfig

  const inferredSelections = extractSelections(legacyManifest)

  const selections = normalizeSelections(
    (asRecord(data.selections) ? (data.selections as Partial<StoreData['selections']>) : undefined) ?? inferredSelections,
  )

  return {
    meta: {
      schemaVersion: CURRENT_SCHEMA_VERSION,
      cliVersion: asString(data.version, '0.0.0'),
      installedAt,
      lastUpdatedAt: now,
    },
    config: {
      setupScope: config.setupScope as StoreData['config']['setupScope'],
      setupType: config.setupType as StoreData['config']['setupType'],
      tools: config.tools as StoreData['config']['tools'],
      projectName: config.projectName,
      targetDir,
    },
    selections,
    files: legacyFiles.map((file) => ({
      path: file.path,
      hash: file.hash,
      source: file.source,
      status: 'installed',
      installedAt,
      lastCheckedAt: now,
    })),
    sync: {
      lastSyncAt: now,
      dirty: false,
    },
    operations: [],
  }
}

export function migrate(targetDir: string, data: unknown): StoreData {
  try {
    const migrated = isLegacyFormat(data)
      ? migrateV0toV1(data as Record<string, unknown>, targetDir)
      : (() => {
          const parsed = asRecord(data)
          const schemaVersion = asRecord(parsed?.meta)?.schemaVersion
          if (typeof schemaVersion === 'number') {
            return data as StoreData
          }

          throw Errors.manifestCorrupt(targetDir, new Error('Unknown store format'))
        })()

    return storeDataSchema.parse(migrated)
  } catch (error) {
    if (error instanceof Error && 'code' in error) {
      throw error
    }

    throw Errors.migrationFailed('0', '1', ensureError(error))
  }
}

export function inferTargetDirFromManifestPath(manifestPath: string): string {
  return dirname(manifestPath)
}
