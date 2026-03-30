import { run } from './cli.js'

run().catch((err: unknown) => {
  if (err instanceof Error) {
    console.error(`\n❌  ${err.message}\n`)
  } else {
    console.error('\n❌  An unexpected error occurred\n', err)
  }
  process.exit(1)
})
