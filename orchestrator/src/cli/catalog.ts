import fs from 'node:fs'
import type { CatalogKindExtended } from '../catalog/schemas.js'
import { CatalogToolHandlers } from '../catalog-tools.js'
import type { Db } from '../db/index.js'

export const CATALOG_HELP = `Usage: ai-setup-orchestrator catalog <subcommand> [options]

Manage the internal versioned catalog.

Subcommands:
  list [--kind <kind>]                          List all definitions
  versions <kind> <name>                        List all versions of a definition
  get <kind> <name> [--version N]               Show a definition (active or pinned version)
  create-version <kind> <name> --from <file>    Create a new version from a file
  set-active <kind> <name> <version>            Move the active pointer to a version
  diff <kind> <name> <fromV> <toV>              Compare two versions
  import [--host opencode|claude-code]          Bulk-import from host config files
  export-version <kind> <name> <target-path> [--version N]
                                                Write a version body to a file

Options:
  -h, --help    Show this help
`

function pickFlag(args: string[], flag: string): string | undefined {
  const idx = args.indexOf(flag)
  if (idx === -1 || !args[idx + 1]) return undefined
  return args[idx + 1]
}

function pickFlagInt(args: string[], flag: string): number | undefined {
  const v = pickFlag(args, flag)
  return v !== undefined ? parseInt(v, 10) : undefined
}

function pickFlags(args: string[], flag: string): string[] {
  const result: string[] = []
  for (let i = 0; i < args.length; i++) {
    if (args[i] === flag && args[i + 1]) {
      result.push(args[i + 1] ?? '')
      i++
    }
  }
  return result
}

export async function runCatalog(db: Db, args: string[], out: NodeJS.WritableStream = process.stdout): Promise<void> {
  const handlers = new CatalogToolHandlers(db)
  const [sub, ...rest] = args

  if (!sub || sub === '-h' || sub === '--help') {
    out.write(CATALOG_HELP)
    return
  }

  const write = (data: unknown): void => {
    out.write(`${JSON.stringify(data, null, 2)}\n`)
  }

  switch (sub) {
    case 'list': {
      const kind = pickFlag(rest, '--kind') as CatalogKindExtended | undefined
      write(handlers.catalogList(kind ? { kind } : {}))
      break
    }
    case 'versions': {
      const [kind, name] = rest
      if (!kind || !name) throw new Error('Usage: catalog versions <kind> <name>')
      write(handlers.catalogListVersions({ kind: kind as CatalogKindExtended, name }))
      break
    }
    case 'get': {
      const [kind, name] = rest
      if (!kind || !name) throw new Error('Usage: catalog get <kind> <name> [--version N]')
      const version = pickFlagInt(rest, '--version')
      const getInput = version !== undefined
        ? { kind: kind as CatalogKindExtended, name, version }
        : { kind: kind as CatalogKindExtended, name }
      write(handlers.catalogGetVersion(getInput))
      break
    }
    case 'create-version': {
      const [kind, name] = rest
      if (!kind || !name) throw new Error('Usage: catalog create-version <kind> <name> --from <file>')
      const fromFile = pickFlag(rest, '--from')
      if (!fromFile) throw new Error('--from <file> is required')
      const raw = fs.readFileSync(fromFile, 'utf-8')
      const createdBy = pickFlag(rest, '--created-by')
      const isJson = fromFile.endsWith('.json')
      const frontmatter: Record<string, unknown> = isJson ? {} : {}
      write(handlers.catalogCreateVersion({
        kind: kind as CatalogKindExtended,
        name,
        frontmatter,
        body: raw,
        ...(createdBy ? { createdBy } : {}),
      }))
      break
    }
    case 'set-active': {
      const [kind, name, versionStr] = rest
      if (!kind || !name || !versionStr) throw new Error('Usage: catalog set-active <kind> <name> <version>')
      write(handlers.catalogSetActive({ kind: kind as CatalogKindExtended, name, version: parseInt(versionStr, 10) }))
      break
    }
    case 'diff': {
      const [kind, name, fromStr, toStr] = rest
      if (!kind || !name || !fromStr || !toStr) throw new Error('Usage: catalog diff <kind> <name> <fromV> <toV>')
      write(handlers.catalogDiff({
        kind: kind as CatalogKindExtended,
        name,
        fromVersion: parseInt(fromStr, 10),
        toVersion: parseInt(toStr, 10),
      }))
      break
    }
    case 'import': {
      const hosts = pickFlags(rest, '--host') as Array<'opencode' | 'claude-code'>
      const libraryRoot = pickFlag(rest, '--library')
      const agentsRoot = pickFlag(rest, '--agents-root')
      const projectRoot = pickFlag(rest, '--project-root')
      write(handlers.catalogImport({
        ...(hosts.length ? { hosts } : {}),
        ...(libraryRoot ? { libraryOrchestrationRoot: libraryRoot } : {}),
        ...(agentsRoot ? { libraryAgentsRoot: agentsRoot } : {}),
        ...(projectRoot ? { projectRoot } : {}),
      }))
      break
    }
    case 'export-version': {
      const [kind, name, targetPath] = rest
      if (!kind || !name || !targetPath) throw new Error('Usage: catalog export-version <kind> <name> <target-path> [--version N]')
      const version = pickFlagInt(rest, '--version')
      write(handlers.catalogExportVersion({
        kind: kind as CatalogKindExtended,
        name,
        targetPath,
        ...(version !== undefined ? { version } : {}),
      }))
      break
    }
    default:
      out.write(`Unknown catalog subcommand: ${sub}\n`)
      out.write(CATALOG_HELP)
  }
}
