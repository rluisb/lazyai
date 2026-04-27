import type { GeneratedFile, Generator, GeneratorConfig, PromptQuestion } from './types.js'

function toSlug(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function toFunctionName(slug: string): string {
  const camel = slug
    .split('-')
    .filter(Boolean)
    .map((part, index) => (index === 0 ? part : part.charAt(0).toUpperCase() + part.slice(1)))
    .join('')
  return camel.charAt(0).toUpperCase() + camel.slice(1)
}

export class CommandGenerator implements Generator {
  readonly type = 'command' as const

  getPromptQuestions(): PromptQuestion[] {
    return [
      {
        key: 'arguments',
        label: 'Arguments signature (example: [name] or <tool>)',
        type: 'text',
      },
      {
        key: 'flagsDescription',
        label: 'Flags description (human readable)',
        type: 'text',
      },
    ]
  }

  async generate(config: GeneratorConfig): Promise<GeneratedFile[]> {
    const slug = toSlug(config.name)
    const fnSuffix = toFunctionName(slug || 'new-command')
    const argsSignature = String(config.answers?.arguments ?? '').trim()
    const commandNameWithArgs = `${slug || 'new-command'}${argsSignature ? ` ${argsSignature}` : ''}`
    const flagsDescription = String(config.answers?.flagsDescription ?? 'No additional flags yet.')

    const content = `import type { Command } from 'commander'

interface ${fnSuffix}Options {
  interactive: boolean
}

export function register${fnSuffix}(program: Command): void {
  program
    .command('${commandNameWithArgs}')
    .description('${config.description ?? `Run ${slug || 'new-command'} command`}')
    .option('--no-interactive', 'Disable interactive mode')
    .action(async (_name: string | undefined, opts: ${fnSuffix}Options) => {
      if (!opts.interactive) {
        console.log('${slug || 'new-command'} executed in non-interactive mode')
        return
      }

      console.log('${slug || 'new-command'} executed')
      console.log('Flags: ${flagsDescription.replace(/'/g, "\\'")}')
    })
}
`

    return [
      {
        path: `src/commands/${slug || 'new-command'}.ts`,
        content,
      },
    ]
  }
}
