import { retryChainStep } from '../chain-machine.js'
import { getEventBus } from '../events/bus.js'
import { getPersistenceDb } from '../persistence.js'
import { loadChainState, loadExecutionPlan, saveChainState } from '../persistence.js'
import type { QueueWorker } from './worker.js'

/**
 * Register built-in job handlers on a worker.
 * Called once at server boot (both stdio and HTTP/SSE modes).
 */
export function registerBuiltinHandlers(worker: QueueWorker): void {
  worker.register('chain_retry', async (job) => {
    const { chainId, stepId, projectRoot } = job.payload as {
      chainId: string
      stepId: string
      projectRoot: string
    }

    const db = getPersistenceDb()
    const state = loadChainState(projectRoot, chainId)
    const plan = loadExecutionPlan(projectRoot, state.executionPlanId)

    const { state: nextState, attemptsRemaining } = retryChainStep(state, plan, stepId)
    saveChainState(projectRoot, nextState)

    getEventBus().emit(db, {
      eventType: 'chain.retry_ready',
      runId: chainId,
      runKind: 'chain',
      payload: { stepId, attemptsRemaining, state: nextState.state },
    })
  })
}
