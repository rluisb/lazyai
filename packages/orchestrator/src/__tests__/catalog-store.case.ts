import { beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { CatalogStore } from '../catalog/store.js'
import type { Db } from '../db/index.js'

let db: Db
let store: CatalogStore

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
  store = new CatalogStore(db)
})

describe('CatalogStore', () => {
  it('creates the first version and sets it active', () => {
    const result = store.createVersion({
      kind: 'agent',
      name: 'reviewer',
      frontmatter: { name: 'Reviewer', description: 'Reviews code' },
      body: '# Reviewer\nReview the code.',
    })
    expect(result.version).toBe(1)
    expect(result.alreadyExists).toBe(false)

    const active = store.getActiveVersion('agent', 'reviewer')
    expect(active).not.toBeNull()
    expect(active?.version).toBe(1)
    expect(JSON.parse(active!.frontmatterJson)).toMatchObject({ name: 'Reviewer' })
  })

  it('is idempotent — identical content returns same version', () => {
    const a = store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'body' })
    const b = store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'body' })
    expect(a.version).toBe(b.version)
    expect(b.alreadyExists).toBe(true)
  })

  it('creates version 2 when content changes', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    const v2 = store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })
    expect(v2.version).toBe(2)
    expect(v2.alreadyExists).toBe(false)
    // Active pointer moved to v2
    expect(store.getActiveVersion('agent', 'reviewer')?.version).toBe(2)
  })

  it('setActiveVersion pins to a specific version', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })
    store.setActiveVersion('agent', 'reviewer', 1)
    expect(store.getActiveVersion('agent', 'reviewer')?.body).toBe('v1')
  })

  it('setActiveVersion throws for unknown version', () => {
    expect(() => store.setActiveVersion('agent', 'no-such', 99)).toThrow(/not found/)
  })

  it('deactivateDefinition clears active version while preserving history', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })

    store.deactivateDefinition('agent', 'reviewer')

    expect(store.getActiveVersion('agent', 'reviewer')).toBeNull()
    expect(store.getVersion('agent', 'reviewer', 1)?.body).toBe('v1')
    expect(store.getVersion('agent', 'reviewer', 2)?.body).toBe('v2')
    expect(store.listDefinitions('agent')).toMatchObject([
      { kind: 'agent', name: 'reviewer', activeVersion: null, totalVersions: 2 },
    ])
  })

  it('deactivateDefinition throws for unknown definition', () => {
    expect(() => store.deactivateDefinition('agent', 'no-such')).toThrow(/Definition agent\/no-such not found/)
  })

  it('removeDefinition deletes a definition and all versions', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })

    const result = store.removeDefinition('agent', 'reviewer')

    expect(result).toEqual({ versionsRemoved: 2 })
    expect(store.getActiveVersion('agent', 'reviewer')).toBeNull()
    expect(store.getVersion('agent', 'reviewer', 1)).toBeNull()
    expect(store.listVersions('agent', 'reviewer')).toEqual([])
    expect(store.listDefinitions('agent')).toEqual([])
  })

  it('removeDefinition throws for unknown definition', () => {
    expect(() => store.removeDefinition('agent', 'no-such')).toThrow(/Definition agent\/no-such not found/)
  })

  it('listVersions returns all versions in order', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v3' })
    const versions = store.listVersions('agent', 'reviewer')
    expect(versions.map((v) => v.version)).toEqual([1, 2, 3])
  })

  it('listDefinitions filters by kind', () => {
    store.createVersion({ kind: 'agent', name: 'builder', frontmatter: { name: 'Builder' }, body: '' })
    store.createVersion({ kind: 'skill', name: 'typescript', frontmatter: { name: 'TypeScript', kind: 'domain' }, body: '' })
    const agents = store.listDefinitions('agent')
    expect(agents.map((d) => d.name)).toContain('builder')
    expect(agents.map((d) => d.kind).every((k) => k === 'agent')).toBe(true)
  })

  it('diffVersions returns both sides', () => {
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v1 body' })
    store.createVersion({ kind: 'agent', name: 'reviewer', frontmatter: { name: 'Reviewer' }, body: 'v2 body' })
    const diff = store.diffVersions('agent', 'reviewer', 1, 2)
    expect(diff.from?.body).toBe('v1 body')
    expect(diff.to?.body).toBe('v2 body')
  })

  it('validates frontmatter for agent kind', () => {
    expect(() =>
      store.createVersion({
        kind: 'agent',
        name: 'bad',
        frontmatter: { name: 123 },
        body: '',
      }),
    ).toThrow(/Invalid frontmatter/)
  })
})
