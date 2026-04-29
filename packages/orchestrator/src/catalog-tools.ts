import fs from 'node:fs'
import path from 'node:path'
import type { Db } from './db/index.js'
import type { CatalogKindExtended } from './catalog/schemas.js'
import { CatalogStore } from './catalog/store.js'
import { importFromHostFiles, importFromLibrary } from './catalog/importer.js'

export interface CatalogListInput {
  kind?: CatalogKindExtended
}

export interface CatalogCreateVersionInput {
  kind: CatalogKindExtended
  name: string
  frontmatter: Record<string, unknown>
  body: string
  createdBy?: string
}

export interface CatalogSetActiveInput {
  kind: CatalogKindExtended
  name: string
  version: number
}

export interface CatalogDefinitionInput {
  kind: CatalogKindExtended
  name: string
}

export type CatalogGetVersionInput =
  | { kind: CatalogKindExtended; name: string; version: number }
  | { kind: CatalogKindExtended; name: string }

export interface CatalogListVersionsInput {
  kind: CatalogKindExtended
  name: string
}

export interface CatalogDiffInput {
  kind: CatalogKindExtended
  name: string
  fromVersion: number
  toVersion: number
}

export interface CatalogImportInput {
  hosts?: Array<'opencode' | 'claude-code'>
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
  projectRoot?: string
}

export interface CatalogExportVersionInput {
  kind: CatalogKindExtended
  name: string
  version?: number
  targetPath: string
}

export class CatalogToolHandlers {
  private readonly store: CatalogStore

  constructor(private readonly db: Db) {
    this.store = new CatalogStore(db)
  }

  catalogList(input: CatalogListInput = {}) {
    return { definitions: this.store.listDefinitions(input.kind) }
  }

  catalogListVersions(input: CatalogListVersionsInput) {
    return { versions: this.store.listVersions(input.kind, input.name) }
  }

  catalogGetVersion(input: CatalogGetVersionInput) {
    const row = 'version' in input && input.version !== undefined
      ? this.store.getVersion(input.kind, input.name, input.version)
      : this.store.getActiveVersion(input.kind, input.name)
    const versionNum = 'version' in input ? input.version : undefined
    if (!row) {
      throw new Error(
        versionNum !== undefined
          ? `Version ${versionNum} of ${input.kind}/${input.name} not found`
          : `No active version for ${input.kind}/${input.name}`,
      )
    }
    return {
      kind: input.kind,
      name: input.name,
      version: row.version,
      frontmatter: JSON.parse(row.frontmatterJson) as Record<string, unknown>,
      body: row.body,
      checksum: row.checksum,
      createdAt: row.createdAt,
      ...(row.createdBy !== undefined ? { createdBy: row.createdBy } : {}),
    }
  }

  catalogCreateVersion(input: CatalogCreateVersionInput) {
    const result = this.store.createVersion({
      kind: input.kind,
      name: input.name,
      frontmatter: input.frontmatter,
      body: input.body,
      ...(input.createdBy !== undefined ? { createdBy: input.createdBy } : {}),
    })
    return result
  }

  catalogSetActive(input: CatalogSetActiveInput) {
    this.store.setActiveVersion(input.kind, input.name, input.version)
    return { kind: input.kind, name: input.name, activeVersion: input.version }
  }

  catalogDeactivate(input: CatalogDefinitionInput) {
    this.store.deactivateDefinition(input.kind, input.name)
    return { kind: input.kind, name: input.name, activeVersion: null, deactivated: true }
  }

  catalogRemove(input: CatalogDefinitionInput) {
    const result = this.store.removeDefinition(input.kind, input.name)
    return { kind: input.kind, name: input.name, removed: true, versionsRemoved: result.versionsRemoved }
  }

  catalogDiff(input: CatalogDiffInput) {
    return this.store.diffVersions(input.kind, input.name, input.fromVersion, input.toVersion)
  }

  catalogImport(input: CatalogImportInput) {
    const results = []
    if (input.libraryOrchestrationRoot) {
      results.push({
        source: 'library',
        ...importFromLibrary(this.db, input.libraryOrchestrationRoot, input.libraryAgentsRoot),
      })
    }
    if (input.hosts?.length) {
      results.push({
        source: 'host-files',
        ...importFromHostFiles(this.db, input.hosts, input.projectRoot),
      })
    }
    return { results }
  }

  catalogExportVersion(input: CatalogExportVersionInput): { targetPath: string; kind: CatalogKindExtended; name: string; version: number } {
    const row = input.version !== undefined
      ? this.store.getVersion(input.kind, input.name, input.version)
      : this.store.getActiveVersion(input.kind, input.name)

    if (!row) {
      throw new Error(
        input.version !== undefined
          ? `Version ${input.version} of ${input.kind}/${input.name} not found`
          : `No active version for ${input.kind}/${input.name}`,
      )
    }

    const dir = path.dirname(input.targetPath)
    if (dir !== '.') fs.mkdirSync(dir, { recursive: true })
    fs.writeFileSync(input.targetPath, row.body, 'utf-8')

    return { targetPath: input.targetPath, kind: input.kind, name: input.name, version: row.version }
  }
}
