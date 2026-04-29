import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    include: ['src/__tests__/**/*.case.ts', 'src/__tests__/**/*.test.ts'],
    environment: 'node',
    coverage: {
      reporter: ['text', 'json-summary'],
    },
  },
})
