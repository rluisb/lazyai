import { describe, expect, it } from 'vitest'
import { type FragmentContext, FragmentResolver } from './fragment-resolver.js'

describe('FragmentResolver constitution context', () => {
  const resolver = new FragmentResolver('')

  it('resolves populated constitution fields', () => {
    const context: FragmentContext = {
      projectName: 'Test',
      planningDir: 'specs',
      constitution: {
        projectOverview: 'Test overview',
        namingConventions: 'camelCase',
        coverageThreshold: 80,
      },
    }

    const result = resolver.resolve('{{PROJECT_OVERVIEW}}|{{NAMING_CONVENTIONS}}|{{COVERAGE_THRESHOLD}}', context)

    expect(result).toBe('Test overview|camelCase|80')
  })

  it('falls back to explicit legacy markers for missing constitution fields', () => {
    const context: FragmentContext = {
      projectName: 'Test',
      planningDir: 'specs',
    }

    const result = resolver.resolve('{{PROJECT_OVERVIEW}}|{{COVERAGE_THRESHOLD}}', context)

    expect(result).toBe('[YOUR_PROJECT_OVERVIEW]|80')
  })

  it('renders codebase map rows with responsibility placeholders', () => {
    const context: FragmentContext = {
      projectName: 'Test',
      planningDir: 'specs',
      constitution: {
        codebaseMap: [
          { path: 'src' },
          { path: 'node_modules' },
          { path: 'packages/api', responsibility: 'API service' },
        ],
      },
    }

    const result = resolver.resolve('{{CODEBASE_MAP}}', context)

    expect(result).toBe('| src | [WHAT_IT_DOES] |\n| packages/api | API service |')
  })

  it('toggles adversarialDesign conditionals from feature flags', () => {
    const template = 'before{{#if features.adversarialDesign}} adversarial{{/if}} after'

    expect(
      resolver.resolve(template, {
        projectName: 'Test',
        planningDir: 'specs',
        features: { adversarialDesign: true },
      }),
    ).toBe('before adversarial after')

    expect(
      resolver.resolve(template, {
        projectName: 'Test',
        planningDir: 'specs',
        features: { adversarialDesign: false },
      }),
    ).toBe('before after')
  })

  it('matches the Go W1.A field parity snapshot', () => {
    const context: FragmentContext = {
      projectName: 'creator-checkout',
      planningDir: 'specs',
      constitution: {
        projectOverview: 'Creator Checkout processes creator payments.',
        stack: {
          language: 'TypeScript',
          framework: 'Next.js',
          database: 'PostgreSQL',
          orm: 'Prisma',
          testing: 'Vitest',
          packageManager: 'yarn',
        },
        conventions: {
          naming: 'camelCase values; PascalCase React components',
          errorHandling: 'Return typed Result values at service boundaries',
          apiResponses: 'JSON envelopes include data or error',
          importOrder: 'External, internal aliases, relative imports',
        },
        commands: {
          test: 'yarn test',
          lint: 'yarn lint',
          build: 'yarn build',
        },
        protectedBranch: 'main',
        coverageThreshold: 87,
        codebaseMap: [
          { path: 'src' },
          { path: 'packages/api', responsibility: 'API package' },
        ],
      },
    }
    const template = [
      '{{PROJECT_NAME}}',
      '{{PROJECT_OVERVIEW}}',
      '{{LANGUAGE}}|{{FRAMEWORK}}|{{DATABASE}}|{{ORM}}|{{TEST_FRAMEWORK}}|{{PACKAGE_MANAGER}}',
      '{{NAMING_CONVENTIONS}}|{{ERROR_HANDLING}}|{{API_CONVENTIONS}}|{{IMPORT_ORDER}}',
      '{{TEST_COMMAND}}|{{LINT_COMMAND}}|{{BUILD_COMMAND}}|{{PROTECTED_BRANCH}}|{{COVERAGE_THRESHOLD}}',
      '{{CODEBASE_MAP}}',
    ].join('\n')

    const result = resolver.resolve(template, context)

    expect(result).toBe(
      [
        'creator-checkout',
        'Creator Checkout processes creator payments.',
        'TypeScript|Next.js|PostgreSQL|Prisma|Vitest|yarn',
        'camelCase values; PascalCase React components|Return typed Result values at service boundaries|JSON envelopes include data or error|External, internal aliases, relative imports',
        'yarn test|yarn lint|yarn build|main|87',
        '| src | [WHAT_IT_DOES] |',
        '| packages/api | API package |',
      ].join('\n'),
    )
  })
})
