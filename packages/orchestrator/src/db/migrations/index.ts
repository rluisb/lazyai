import type { Migration } from '../types.js'
import { migration as m0001 } from './0001_init.js'
import { migration as m0002 } from './0002_run_state.js'
import { migration as m0003 } from './0003_catalog.js'
import { migration as m0004 } from './0004_events.js'
import { migration as m0005 } from './0005_queue.js'

export const migrations: Migration[] = [m0001, m0002, m0003, m0004, m0005]
