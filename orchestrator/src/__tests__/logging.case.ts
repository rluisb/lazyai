import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { PassThrough } from 'node:stream'
import { afterEach, describe, expect, it } from 'vitest'
import { runTail, parseTailArgs } from '../cli/tail.js'
import { createLogger, createNoopLogger } from '../logging/logger.js'
import { createFileSink, createMemorySink, dayStamp, logFileName } from '../logging/sink.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function makeTempDir(prefix: string): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), prefix))
  tempDirs.push(dir)
  return dir
}

describe('logger', () => {
  it('writes structured records and respects minLevel', () => {
    const sink = createMemorySink()
    const logger = createLogger({ sink, minLevel: 'debug' })
    logger.trace('skipped — below min level')
    logger.debug('hello', { a: 1 })
    logger.info('world')

    expect(sink.records.map((r) => r.msg)).toEqual(['hello', 'world'])
    expect(sink.records[0]?.a).toBe(1)
    expect(sink.records[0]?.level).toBe('debug')
    expect(typeof sink.records[0]?.ts).toBe('string')
  })

  it('child loggers merge bindings', () => {
    const sink = createMemorySink()
    const root = createLogger({ sink, bindings: { service: 'orchestrator' } })
    const child = root.child({ runId: 'chain-1' })
    child.info('step started', { stepId: 'plan' })

    expect(sink.records[0]).toMatchObject({
      service: 'orchestrator',
      runId: 'chain-1',
      stepId: 'plan',
      msg: 'step started',
    })
  })

  it('noop logger does not throw and supports children', () => {
    const logger = createNoopLogger()
    expect(() => logger.info('ignored')).not.toThrow()
    expect(() => logger.child({ a: 1 }).warn('also ignored')).not.toThrow()
  })

  it('file sink writes JSONL to a date-stamped file', async () => {
    const dir = makeTempDir('orchestrator-log-')
    const sink = createFileSink({ dir })
    const logger = createLogger({ sink, minLevel: 'info' })
    logger.info('first', { runId: 'r1' })
    logger.warn('second')
    await sink.close?.()

    const file = path.join(dir, logFileName(dayStamp()))
    const lines = fs.readFileSync(file, 'utf8').trim().split('\n')
    expect(lines).toHaveLength(2)
    const parsed = lines.map((l) => JSON.parse(l) as Record<string, unknown>)
    expect(parsed[0]?.msg).toBe('first')
    expect(parsed[0]?.runId).toBe('r1')
    expect(parsed[1]?.level).toBe('warn')
  })
})

describe('tail CLI', () => {
  it('parses filters, --no-follow, --from-start, --date', () => {
    const parsed = parseTailArgs([
      '--filter',
      'runId=chain-1',
      '--filter',
      'level=warn',
      '--no-follow',
      '--from-start',
      '--date',
      '2026-04-20',
    ])
    expect(parsed.filters).toEqual([
      { key: 'runId', value: 'chain-1' },
      { key: 'level', value: 'warn' },
    ])
    expect(parsed.follow).toBe(false)
    expect(parsed.fromStart).toBe(true)
    expect(parsed.date).toBe('2026-04-20')
  })

  it('reads from start, applies filters, and exits when --no-follow', async () => {
    const dir = makeTempDir('orchestrator-tail-')
    const sink = createFileSink({ dir })
    const logger = createLogger({ sink, minLevel: 'info' })
    logger.info('a', { runId: 'chain-1' })
    logger.info('b', { runId: 'chain-2' })
    logger.warn('c', { runId: 'chain-1' })
    await sink.close?.()

    const out = new PassThrough()
    const chunks: Buffer[] = []
    out.on('data', (chunk: Buffer) => chunks.push(chunk))

    await runTail({
      dir,
      follow: false,
      fromStart: true,
      filters: [{ key: 'runId', value: 'chain-1' }],
      out,
    })

    const written = Buffer.concat(chunks).toString('utf8')
    const lines = written.trim().split('\n')
    expect(lines).toHaveLength(2)
    const parsed = lines.map((l) => JSON.parse(l) as Record<string, unknown>)
    expect(parsed.map((r) => r.msg)).toEqual(['a', 'c'])
  })

  it('skips entries with non-matching filter values', async () => {
    const dir = makeTempDir('orchestrator-tail-')
    const sink = createFileSink({ dir })
    const logger = createLogger({ sink })
    logger.info('hit', { runId: 'r1' })
    logger.info('miss', { runId: 'r2' })
    await sink.close?.()

    const out = new PassThrough()
    const chunks: Buffer[] = []
    out.on('data', (chunk: Buffer) => chunks.push(chunk))

    await runTail({
      dir,
      follow: false,
      fromStart: true,
      filters: [{ key: 'runId', value: 'r1' }],
      out,
    })

    const records = Buffer.concat(chunks)
      .toString('utf8')
      .trim()
      .split('\n')
      .filter((l) => l.length > 0)
      .map((l) => JSON.parse(l) as Record<string, unknown>)
    expect(records.map((r) => r.msg)).toEqual(['hit'])
  })
})
