import { beforeEach, describe, expect, it } from 'vitest'
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { CatalogStore } from '../catalog/store.js'
import { resolveCatalog } from '../catalog/resolver.js'
import type { Db } from '../db/index.js'
import type { OrchestrationCatalog } from '../types.js'

let db: Db

const emptyCatalog: OrchestrationCatalog = {
  agents: {},
  domains: {},
  modes: {},
  chains: {},
  teams: {},
  workflows: {},
}

beforeEach(() => {
  db = openDatabase(':memory:')
  runMigrations(db)
})

describe('resolveCatalog', () => {
  it('overlays DB agents on top of the file-based catalog', () => {
    const store = new CatalogStore(db)
    store.createVersion({
      kind: 'agent',
      name: 'reviewer',
      frontmatter: { name: 'Reviewer DB', description: 'From DB', model: 'sonnet' },
      body: 'Review the code.',
    })

    const result = resolveCatalog(emptyCatalog, { db, projectRoot: '/tmp/test' })
    expect(result.agents.reviewer).toBeDefined()
    expect(result.agents.reviewer?.source).toBe('db')
    expect(result.agents.reviewer?.modelHint).toBe('sonnet')
  })

  it('file-based catalog entries survive when no DB version exists', () => {
    const catalogWithAgent: OrchestrationCatalog = {
      ...emptyCatalog,
      agents: {
        builder: {
          kind: 'agent',
          name: 'builder',
          displayName: 'Builder',
          description: 'Builds things',
          source: 'library',
          path: '/lib/builder.md',
          prompt: 'Build it.',
          allowedTools: ['Bash'],
          constraints: [],
        },
      },
    }
    const result = resolveCatalog(catalogWithAgent, { db, projectRoot: '/tmp/test' })
    expect(result.agents.builder?.source).toBe('library')
  })

  it('DB version beats file-based when both define the same name', () => {
    const store = new CatalogStore(db)
    store.createVersion({
      kind: 'agent',
      name: 'builder',
      frontmatter: { name: 'Builder DB', description: 'DB builder' },
      body: 'DB prompt',
    })

    const catalogWithAgent: OrchestrationCatalog = {
      ...emptyCatalog,
      agents: {
        builder: {
          kind: 'agent',
          name: 'builder',
          displayName: 'Builder File',
          description: '',
          source: 'library',
          path: '/lib/builder.md',
          prompt: 'File prompt',
          allowedTools: [],
          constraints: [],
        },
      },
    }
    const result = resolveCatalog(catalogWithAgent, { db, projectRoot: '/tmp/test' })
    expect(result.agents.builder?.source).toBe('db')
    expect(result.agents.builder?.prompt).toBe('DB prompt')
  })

  it('returns empty catalog when DB has no definitions', () => {
    const result = resolveCatalog(emptyCatalog, { db, projectRoot: '/tmp/test' })
    expect(Object.keys(result.agents)).toHaveLength(0)
  })
})
