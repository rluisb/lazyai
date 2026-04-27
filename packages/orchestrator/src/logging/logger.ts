export type LogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error'

export interface LogRecord {
  ts: string
  level: LogLevel
  msg: string
  [key: string]: unknown
}

export interface LogSink {
  write(record: LogRecord): void
  close?(): Promise<void> | void
}

export interface Logger {
  child(bindings: Record<string, unknown>): Logger
  trace(msg: string, fields?: Record<string, unknown>): void
  debug(msg: string, fields?: Record<string, unknown>): void
  info(msg: string, fields?: Record<string, unknown>): void
  warn(msg: string, fields?: Record<string, unknown>): void
  error(msg: string, fields?: Record<string, unknown>): void
}

const LEVELS: Record<LogLevel, number> = {
  trace: 10,
  debug: 20,
  info: 30,
  warn: 40,
  error: 50,
}

export interface LoggerOptions {
  sink: LogSink
  minLevel?: LogLevel
  bindings?: Record<string, unknown>
}

export function createLogger(opts: LoggerOptions): Logger {
  const minLevel = LEVELS[opts.minLevel ?? 'info']
  const bindings = opts.bindings ?? {}

  const log = (level: LogLevel, msg: string, fields?: Record<string, unknown>): void => {
    if (LEVELS[level] < minLevel) return
    const record: LogRecord = {
      ts: new Date().toISOString(),
      level,
      msg,
      ...bindings,
      ...(fields ?? {}),
    }
    opts.sink.write(record)
  }

  return {
    child(extra) {
      return createLogger({
        sink: opts.sink,
        minLevel: opts.minLevel ?? 'info',
        bindings: { ...bindings, ...extra },
      })
    },
    trace: (m, f) => log('trace', m, f),
    debug: (m, f) => log('debug', m, f),
    info: (m, f) => log('info', m, f),
    warn: (m, f) => log('warn', m, f),
    error: (m, f) => log('error', m, f),
  }
}

export function createNoopLogger(): Logger {
  const noop = (): void => {}
  const self: Logger = {
    child: () => self,
    trace: noop,
    debug: noop,
    info: noop,
    warn: noop,
    error: noop,
  }
  return self
}
