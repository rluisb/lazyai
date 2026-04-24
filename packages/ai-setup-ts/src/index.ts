import { run } from './cli.js'
import { handleError } from './errors/index.js'

run().catch(handleError)
