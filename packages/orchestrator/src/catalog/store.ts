import crypto from 'node:crypto'
import type { Db } from '../db/index.js'
import type { CatalogKindExtended } from './schemas.js'
import { validateFrontmatterForKind } from './schemas.js'

export interface DefinitionVersionRow {
  id: number
  definitionId: number
  kind: CatalogKindExtended
  name: string
  version: number
  frontmatterJson: string
  body: string
  checksum: string
  createdAt: string
  createdBy?: string
}

export interface DefinitionSummary {
  kind: CatalogKindExtended
  name: string
  activeVersion: number | null
  totalVersions: number
  createdAt: string
  updatedAt: string
}

export interface CreateVersionInput {
  kind: CatalogKindExtended
  name: string
  frontmatter: Record<string, unknown>
  body: string
  createdBy?: string
  setActive?: boolean
}

export interface CreateVersionResult {
  version: number
  checksum: string
  alreadyExists: boolean
}

function checksum(frontmatter: Record<string, unknown>, body: string): string {
  const canonical = JSON.stringify({ frontmatter, body })
  return crypto.createHash('sha256').update(canonical).digest('hex').slice(0, 16)
}

export class CatalogStore {
  constructor(private readonly db: Db) {}

  private ensureDefinition(kind: CatalogKindExtended, name: string): number {
    const now = new Date().toISOString()
    this.db
      .prepare(`
        INSERT INTO definitions (kind, name, created_at, updated_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(kind, name) DO NOTHING
      `)
      .run(kind, name, now, now)
    const row = this.db
      .prepare<[string, string], { id: number }>('SELECT id FROM definitions WHERE kind = ? AND name = ?')
      .get(kind, name)
    if (!row) throw new Error(`Failed to ensure definition: ${kind}/${name}`)
    return row.id
  }

  private getDefinitionId(kind: CatalogKindExtended, name: string): number | null {
    const row = this.db
      .prepare<[string, string], { id: number }>('SELECT id FROM definitions WHERE kind = ? AND name = ?')
      .get(kind, name)
    return row?.id ?? null
  }

  private nextVersion(definitionId: number): number {
    const row = this.db
      .prepare<[number], { max_v: number | null }>('SELECT MAX(version) AS max_v FROM definition_versions WHERE definition_id = ?')
      .get(definitionId)
    return (row?.max_v ?? 0) + 1
  }

  createVersion(input: CreateVersionInput): CreateVersionResult {
    const validation = validateFrontmatterForKind(input.kind, input.frontmatter)
    if (!validation.valid) {
      throw new Error(`Invalid frontmatter for ${input.kind}: ${validation.issues.join(', ')}`)
    }

    const cs = checksum(input.frontmatter, input.body)

    const defId = this.ensureDefinition(input.kind, input.name)

    const existing = this.db
      .prepare<[number, string], { id: number; version: number }>(
        'SELECT id, version FROM definition_versions WHERE definition_id = ? AND checksum = ?',
      )
      .get(defId, cs)

    if (existing) {
      return { version: existing.version, checksum: cs, alreadyExists: true }
    }

    const version = this.nextVersion(defId)
    const now = new Date().toISOString()
    const result = this.db
      .prepare(`
        INSERT INTO definition_versions (definition_id, version, frontmatter_json, body, checksum, created_at, created_by)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `)
      .run(defId, version, JSON.stringify(input.frontmatter), input.body, cs, now, input.createdBy ?? null)

    const newId = result.lastInsertRowid as number

    if (input.setActive !== false) {
      this.db
        .prepare('UPDATE definitions SET active_version_id = ?, updated_at = ? WHERE id = ?')
        .run(newId, now, defId)
    }

    return { version, checksum: cs, alreadyExists: false }
  }

  setActiveVersion(kind: CatalogKindExtended, name: string, version: number): void {
    const row = this.db
      .prepare<[string, string, number], { dv_id: number }>(`
        SELECT dv.id AS dv_id FROM definition_versions dv
        JOIN definitions d ON d.id = dv.definition_id
        WHERE d.kind = ? AND d.name = ? AND dv.version = ?
      `)
      .get(kind, name, version)
    if (!row) throw new Error(`Version ${version} of ${kind}/${name} not found`)
    const now = new Date().toISOString()
    this.db
      .prepare('UPDATE definitions SET active_version_id = ?, updated_at = ? WHERE kind = ? AND name = ?')
      .run(row.dv_id, now, kind, name)
  }

  deactivateDefinition(kind: CatalogKindExtended, name: string): void {
    const definitionId = this.getDefinitionId(kind, name)
    if (definitionId === null) throw new Error(`Definition ${kind}/${name} not found`)

    const now = new Date().toISOString()
    this.db
      .prepare('UPDATE definitions SET active_version_id = NULL, updated_at = ? WHERE id = ?')
      .run(now, definitionId)
  }

  removeDefinition(kind: CatalogKindExtended, name: string): { versionsRemoved: number } {
    const definitionId = this.getDefinitionId(kind, name)
    if (definitionId === null) throw new Error(`Definition ${kind}/${name} not found`)

    const row = this.db
      .prepare<[number], { total: number }>('SELECT COUNT(*) AS total FROM definition_versions WHERE definition_id = ?')
      .get(definitionId)
    const versionsRemoved = row?.total ?? 0

    const remove = this.db.transaction((id: number) => {
      this.db.prepare('UPDATE definitions SET active_version_id = NULL WHERE id = ?').run(id)
      this.db.prepare('DELETE FROM definition_versions WHERE definition_id = ?').run(id)
      this.db.prepare('DELETE FROM definitions WHERE id = ?').run(id)
    })
    remove(definitionId)

    return { versionsRemoved }
  }

  private readonly versionSelect = `
    dv.id,
    dv.definition_id AS "definitionId",
    d.kind,
    d.name,
    dv.version,
    dv.frontmatter_json AS "frontmatterJson",
    dv.body,
    dv.checksum,
    dv.created_at AS "createdAt",
    dv.created_by AS "createdBy"
  `

  getActiveVersion(kind: CatalogKindExtended, name: string): DefinitionVersionRow | null {
    return (this.db
      .prepare(`
        SELECT ${this.versionSelect}
        FROM definitions d
        JOIN definition_versions dv ON d.active_version_id = dv.id
        WHERE d.kind = ? AND d.name = ?
      `)
      .get(kind, name) as DefinitionVersionRow | undefined) ?? null
  }

  getVersion(kind: CatalogKindExtended, name: string, version: number): DefinitionVersionRow | null {
    return (this.db
      .prepare(`
        SELECT ${this.versionSelect}
        FROM definitions d
        JOIN definition_versions dv ON dv.definition_id = d.id
        WHERE d.kind = ? AND d.name = ? AND dv.version = ?
      `)
      .get(kind, name, version) as DefinitionVersionRow | undefined) ?? null
  }

  listVersions(kind: CatalogKindExtended, name: string): DefinitionVersionRow[] {
    return this.db
      .prepare(`
        SELECT ${this.versionSelect}
        FROM definitions d
        JOIN definition_versions dv ON dv.definition_id = d.id
        WHERE d.kind = ? AND d.name = ?
        ORDER BY dv.version ASC
      `)
      .all(kind, name) as DefinitionVersionRow[]
  }

  listDefinitions(kind?: CatalogKindExtended): DefinitionSummary[] {
    const where = kind ? 'WHERE d.kind = ?' : ''
    const params: unknown[] = kind ? [kind] : []
    const rows = this.db
      .prepare<unknown[], {
        kind: string; name: string; active_version: number | null;
        total_versions: number; created_at: string; updated_at: string
      }>(`
        SELECT d.kind, d.name,
               (SELECT dv.version FROM definition_versions dv WHERE dv.id = d.active_version_id) AS active_version,
               (SELECT COUNT(*) FROM definition_versions dv WHERE dv.definition_id = d.id) AS total_versions,
               d.created_at, d.updated_at
        FROM definitions d
        ${where}
        ORDER BY d.kind, d.name
      `)
      .all(...params)
    return rows.map((r) => ({
      kind: r.kind as CatalogKindExtended,
      name: r.name,
      activeVersion: r.active_version,
      totalVersions: r.total_versions,
      createdAt: r.created_at,
      updatedAt: r.updated_at,
    }))
  }

  diffVersions(
    kind: CatalogKindExtended,
    name: string,
    fromVersion: number,
    toVersion: number,
  ): { from: DefinitionVersionRow | null; to: DefinitionVersionRow | null } {
    return {
      from: this.getVersion(kind, name, fromVersion),
      to: this.getVersion(kind, name, toVersion),
    }
  }
}
