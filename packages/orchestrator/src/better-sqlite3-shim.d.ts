declare module 'better-sqlite3' {
  export interface RunResult {
    changes: number
    lastInsertRowid: number | bigint
  }

  export interface Statement<TParams extends unknown[] = unknown[], TResult = unknown> {
    run(...params: TParams): RunResult
    get(...params: TParams): TResult | undefined
    all(...params: TParams): TResult[]
  }

  export interface DatabaseOptions {
    readonly?: boolean
  }

  export default class Database {
    constructor(path: string, options?: DatabaseOptions)
    pragma(source: string): unknown
    exec(source: string): this
    close(): void
    prepare<TParams extends unknown[] = unknown[], TResult = unknown>(source: string): Statement<TParams, TResult>
    transaction<TArgs extends unknown[], TResult>(fn: (...args: TArgs) => TResult): (...args: TArgs) => TResult
  }
}
