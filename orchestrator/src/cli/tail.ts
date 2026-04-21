import fs from 'node:fs'
import path from 'node:path'
import { getLogDir } from '../config/paths.js'
import { dayStamp, logFileName } from '../logging/sink.js'

export interface TailFilter {
  key: string
  value: string
}

export interface TailOptions {
  dir?: string
  baseName?: string
  date?: string
  filters?: TailFilter[]
  follow?: boolean
  fromStart?: boolean
  out?: NodeJS.WritableStream
  signal?: AbortSignal
}

export interface ParsedTailArgs {
  date?: string
  filters: TailFilter[]
  follow: boolean
  fromStart: boolean
  help: boolean
}

export function parseTailArgs(args: string[]): ParsedTailArgs {
  const filters: TailFilter[] = []
  let follow = true
  let fromStart = false
  let date: string | undefined
  let help = false

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === '--filter') {
      const value = args[++i]
      if (!value) continue
      const eq = value.indexOf('=')
      if (eq <= 0) continue
      filters.push({ key: value.slice(0, eq), value: value.slice(eq + 1) })
    } else if (arg === '--no-follow') {
      follow = false
    } else if (arg === '--from-start') {
      fromStart = true
    } else if (arg === '--date') {
      date = args[++i]
    } else if (arg === '-h' || arg === '--help') {
      help = true
    }
  }

  return { ...(date !== undefined ? { date } : {}), filters, follow, fromStart, help }
}

export const TAIL_HELP = `Usage: ai-setup-orchestrator tail [options]

Stream the orchestrator's structured log file.

Options:
  --filter key=value   Only show records matching this field (repeatable)
  --no-follow          Print existing entries and exit
  --from-start         Start from the beginning of the file
  --date YYYY-MM-DD    Tail a specific day's file (default: today)
  -h, --help           Show this help
`

function recordMatches(line: string, filters: TailFilter[]): boolean {
  if (filters.length === 0) return true
  let record: Record<string, unknown>
  try {
    record = JSON.parse(line) as Record<string, unknown>
  } catch {
    return false
  }
  for (const f of filters) {
    if (String(record[f.key]) !== f.value) return false
  }
  return true
}

export async function runTail(options: TailOptions = {}): Promise<void> {
  const out = options.out ?? process.stdout
  const dir = options.dir ?? getLogDir()
  const baseName = options.baseName ?? 'orchestrator'
  const day = options.date ?? dayStamp()
  const file = path.join(dir, logFileName(day, baseName))
  const filters = options.filters ?? []

  fs.mkdirSync(dir, { recursive: true })
  if (!fs.existsSync(file)) fs.writeFileSync(file, '')

  let position = options.fromStart ? 0 : fs.statSync(file).size
  let leftover = ''

  const drain = (): void => {
    const stat = fs.statSync(file)
    if (stat.size <= position) return
    const length = stat.size - position
    const buf = Buffer.alloc(length)
    const fd = fs.openSync(file, 'r')
    try {
      fs.readSync(fd, buf, 0, length, position)
    } finally {
      fs.closeSync(fd)
    }
    position = stat.size
    const text = leftover + buf.toString('utf8')
    const lines = text.split('\n')
    leftover = lines.pop() ?? ''
    for (const line of lines) {
      if (line.length === 0) continue
      if (recordMatches(line, filters)) out.write(`${line}\n`)
    }
  }

  drain()
  if (options.follow === false) return

  await new Promise<void>((resolve) => {
    const watcher = fs.watch(file, { persistent: true }, () => drain())
    const stop = (): void => {
      watcher.close()
      resolve()
    }
    options.signal?.addEventListener('abort', stop, { once: true })
    process.once('SIGINT', stop)
    process.once('SIGTERM', stop)
  })
}
