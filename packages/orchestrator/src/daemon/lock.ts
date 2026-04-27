import fs from 'node:fs'
import path from 'node:path'
import { getLockFilePath } from '../config/paths.js'

export function acquireLock(lockPath?: string): boolean {
  const file = lockPath ?? getLockFilePath()
  fs.mkdirSync(path.dirname(file), { recursive: true })
  try {
    const fd = fs.openSync(file, fs.constants.O_CREAT | fs.constants.O_EXCL | fs.constants.O_WRONLY)
    fs.writeSync(fd, String(process.pid))
    fs.closeSync(fd)
    return true
  } catch {
    // Check if the PID that holds the lock is still alive
    try {
      const pid = parseInt(fs.readFileSync(file, 'utf8').trim(), 10)
      if (!isNaN(pid)) {
        try {
          process.kill(pid, 0)
        } catch {
          // PID is dead — steal the lock
          fs.writeFileSync(file, String(process.pid))
          return true
        }
      }
    } catch { /* can't read lock file, treat as contested */ }
    return false
  }
}

export function releaseLock(lockPath?: string): void {
  try { fs.unlinkSync(lockPath ?? getLockFilePath()) } catch { /* already gone */ }
}
