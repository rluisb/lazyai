import type http from 'node:http'

export interface ClientTrackerOptions {
  onShutdown: () => void
  gracePeriodMs?: number
}

export class ClientTracker {
  private readonly connections = new Set<http.ServerResponse>()
  private shutdownTimer: ReturnType<typeof setTimeout> | null = null
  private readonly onShutdown: () => void
  private readonly gracePeriodMs: number

  constructor(options: ClientTrackerOptions) {
    this.onShutdown = options.onShutdown
    this.gracePeriodMs = options.gracePeriodMs ?? 5000
  }

  track(res: http.ServerResponse): void {
    this.connections.add(res)
    if (this.shutdownTimer !== null) {
      clearTimeout(this.shutdownTimer)
      this.shutdownTimer = null
    }
    res.on('close', () => this.remove(res))
  }

  private remove(res: http.ServerResponse): void {
    this.connections.delete(res)
    if (this.connections.size === 0) {
      this.shutdownTimer = setTimeout(() => this.onShutdown(), this.gracePeriodMs)
    }
  }

  get count(): number {
    return this.connections.size
  }

  clear(): void {
    if (this.shutdownTimer !== null) {
      clearTimeout(this.shutdownTimer)
      this.shutdownTimer = null
    }
    this.connections.clear()
  }
}
