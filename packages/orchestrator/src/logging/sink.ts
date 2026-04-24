import fs from 'node:fs'
import path from 'node:path'
import type { LogRecord, LogSink } from './logger.js'

export function dayStamp(date: Date = new Date()): string {
  return date.toISOString().slice(0, 10)
}

export function logFileName(day: string, baseName = 'orchestrator'): string {
  return `${baseName}-${day}.log`
}

export interface FileSinkOptions {
  dir: string
  baseName?: string
  now?: () => Date
}

export function createFileSink(opts: FileSinkOptions): LogSink {
  fs.mkdirSync(opts.dir, { recursive: true })
  const baseName = opts.baseName ?? 'orchestrator'
  const now = opts.now ?? (() => new Date())
  let currentDay = ''
  let stream: fs.WriteStream | null = null

  const ensureStream = (): fs.WriteStream => {
    const today = dayStamp(now())
    if (today !== currentDay || stream === null) {
      stream?.end()
      currentDay = today
      stream = fs.createWriteStream(path.join(opts.dir, logFileName(today, baseName)), {
        flags: 'a',
      })
    }
    return stream
  }

  return {
    write(record: LogRecord): void {
      ensureStream().write(`${JSON.stringify(record)}\n`)
    },
    close(): Promise<void> {
      return new Promise((resolve) => {
        if (!stream) {
          resolve()
          return
        }
        stream.end(() => resolve())
      })
    },
  }
}

export interface MemorySink extends LogSink {
  records: LogRecord[]
}

export function createMemorySink(): MemorySink {
  const records: LogRecord[] = []
  return {
    records,
    write(record) {
      records.push(record)
    },
  }
}
