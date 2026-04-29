import type { Db } from '../db/index.js'
import { loadCatalog } from '../loader.js'
import { getPersistenceDb } from '../persistence.js'
import type { HostCli, OrchestrationCatalog } from '../types.js'
import { resolveCatalog } from './resolver.js'

export interface RuntimeCatalogOptions {
  projectRoot: string
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
  hostCli?: HostCli
  db?: Db
}

export function loadRuntimeCatalog(options: RuntimeCatalogOptions): OrchestrationCatalog {
  const base = loadCatalog({
    projectRoot: options.projectRoot,
    ...(options.libraryOrchestrationRoot ? { libraryOrchestrationRoot: options.libraryOrchestrationRoot } : {}),
    ...(options.libraryAgentsRoot ? { libraryAgentsRoot: options.libraryAgentsRoot } : {}),
  })
  const resolverHostCli = isResolverHostCli(options.hostCli) ? options.hostCli : undefined

  return resolveCatalog(base, {
    db: options.db ?? getPersistenceDb(),
    projectRoot: options.projectRoot,
    ...(resolverHostCli !== undefined ? { hostCli: resolverHostCli } : {}),
  })
}

function isResolverHostCli(hostCli: HostCli | undefined): hostCli is 'opencode' | 'claude-code' | 'codex' {
  return hostCli === 'opencode' || hostCli === 'claude-code' || hostCli === 'codex'
}
