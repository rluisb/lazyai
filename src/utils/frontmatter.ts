const YAML_FRONTMATTER_REGEX = /^---\s*\n([\s\S]*?)\n---\s*(?:\n|$)/

function splitYamlFrontmatter(content: string): { frontmatterBody: string; body: string } | null {
  const match = content.match(YAML_FRONTMATTER_REGEX)
  if (!match) return null

  const [, frontmatterBody = ''] = match
  const body = content.slice(match[0].length)

  return { frontmatterBody, body }
}

export function stripYamlFrontmatter(content: string): string {
  return content.replace(YAML_FRONTMATTER_REGEX, '')
}

export function stripFrontmatterAndInjectModel(content: string): string {
  const match = content.match(YAML_FRONTMATTER_REGEX)
  if (!match) return content

  const [, frontmatterBody = ''] = match
  const body = content.slice(match[0].length)

  const modelMatch = frontmatterBody.match(/^model\s*:\s*(.+)$/m)
  if (modelMatch?.[1]) {
    const model = modelMatch[1].trim()
    return `<!-- Recommended model: ${model} -->\n\n${body}`
  }

  return body
}

export function ensureModeAgentFrontmatter(content: string): string {
  const split = splitYamlFrontmatter(content)

  if (!split) {
    return `---\nmode: agent\n---\n\n${content}`
  }

  const { frontmatterBody, body } = split
  const lines = frontmatterBody.split('\n')

  let replaced = false
  const normalized = lines.map(line => {
    if (/^mode\s*:/i.test(line)) {
      replaced = true
      return 'mode: agent'
    }

    return line
  })

  if (!replaced) {
    normalized.push('mode: agent')
  }

  return `---\n${normalized.join('\n')}\n---\n\n${body}`
}
