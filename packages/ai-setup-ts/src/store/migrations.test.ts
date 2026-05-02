import { describe, expect, it } from 'vitest'
import { downgradeV2ToV1 } from './migrations/v2-to-v1.js'
import { CURRENT_SCHEMA_VERSION, defaultStore, storeDataSchema } from './schema.js'

const v2ConfigKeys = [
  'projectOverview',
  'namingConventions',
  'errorHandling',
  'apiConventions',
  'importOrder',
  'protectedBranch',
  'testCommand',
  'lintCommand',
  'buildCommand',
  'coverageThreshold',
] as const

describe('store schema v2 migrations', () => {
  it('parses a v1 store with safe v2 defaults', () => {
    const v1Store = defaultStore() as any
    v1Store.meta.schemaVersion = 1
    for (const key of v2ConfigKeys) {
      delete v1Store.config[key]
    }
    v1Store.selections.features = {
      contextEngineering: true,
      rpiWorkflow: true,
      chainOfThought: true,
      treeOfThoughts: true,
      adrEnforcement: true,
      qualityGates: true,
      agentHarness: true,
      bugResolution: true,
      pivotHandling: true,
    }

    const parsed = storeDataSchema.parse(v1Store)

    expect(CURRENT_SCHEMA_VERSION).toBe(2)
    expect(parsed.meta.schemaVersion).toBe(1)
    expect(parsed.config.projectOverview).toBeUndefined()
    expect(parsed.config.namingConventions).toBeUndefined()
    expect(parsed.config.errorHandling).toBeUndefined()
    expect(parsed.config.apiConventions).toBeUndefined()
    expect(parsed.config.importOrder).toBeUndefined()
    expect(parsed.config.protectedBranch).toBeUndefined()
    expect(parsed.config.testCommand).toBeUndefined()
    expect(parsed.config.lintCommand).toBeUndefined()
    expect(parsed.config.buildCommand).toBeUndefined()
    expect(parsed.config.coverageThreshold).toBe(80)
    expect(parsed.selections.features?.adversarialDesign).toBe(false)
  })

  it('preserves valid coverage threshold values', () => {
    const store = defaultStore() as any
    store.config.coverageThreshold = 95

    const parsed = storeDataSchema.parse(store)

    expect(parsed.config.coverageThreshold).toBe(95)
  })

  it('rejects coverage threshold values outside 1-100', () => {
    const belowMinimum = defaultStore() as any
    belowMinimum.config.coverageThreshold = 0

    const aboveMaximum = defaultStore() as any
    aboveMaximum.config.coverageThreshold = 101

    expect(() => storeDataSchema.parse(belowMinimum)).toThrow()
    expect(() => storeDataSchema.parse(aboveMaximum)).toThrow()
  })

  it('downgrades a v2 store to a clean v1 document', () => {
    const v2Store = defaultStore() as any
    v2Store.meta.schemaVersion = 2
    v2Store.config = {
      ...v2Store.config,
      projectName: 'preserved-project',
      projectOverview: 'Project summary',
      namingConventions: 'Use camelCase',
      errorHandling: 'Return typed errors',
      apiConventions: 'JSON responses',
      importOrder: 'Node, third-party, local',
      protectedBranch: 'main',
      testCommand: 'npm test',
      lintCommand: 'npm run lint',
      buildCommand: 'npm run build',
      coverageThreshold: 90,
    }
    v2Store.selections = {
      ...v2Store.selections,
      rules: ['testing'],
      features: {
        contextEngineering: true,
        rpiWorkflow: true,
        chainOfThought: true,
        treeOfThoughts: true,
        adrEnforcement: true,
        qualityGates: true,
        agentHarness: true,
        bugResolution: true,
        pivotHandling: true,
        adversarialDesign: true,
      },
    }

    const downgraded = downgradeV2ToV1(v2Store)

    expect(downgraded.meta.schemaVersion).toBe(1)
    expect(downgraded.config.projectName).toBe('preserved-project')
    expect(downgraded.selections.rules).toEqual(['testing'])
    expect(downgraded.selections.features.qualityGates).toBe(true)
    for (const key of v2ConfigKeys) {
      expect(downgraded.config).not.toHaveProperty(key)
    }
    expect(downgraded.selections.features).not.toHaveProperty('adversarialDesign')
  })
})
