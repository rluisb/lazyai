import type { StoreData } from '../schema.js'

export function downgradeV2ToV1(store: StoreData): any {
  const {
    projectOverview,
    namingConventions,
    errorHandling,
    apiConventions,
    importOrder,
    protectedBranch,
    testCommand,
    lintCommand,
    buildCommand,
    coverageThreshold,
    ...restConfig
  } = store.config as any

  const { adversarialDesign, ...restFeatures } = (store.selections.features as any) ?? {}

  return {
    ...store,
    meta: { ...store.meta, schemaVersion: 1 },
    config: restConfig,
    selections: {
      ...store.selections,
      features: Object.keys(restFeatures).length > 0 ? restFeatures : undefined,
    },
  }
}
