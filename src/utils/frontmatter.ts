const YAML_FRONTMATTER_REGEX = /^---\s*\n([\s\S]*?)\n---\s*\n?/

export function stripYamlFrontmatter(content: string): string {
  return content.replace(YAML_FRONTMATTER_REGEX, '')
}

export function ensureModeAgentFrontmatter(content: string): string {
  const match = content.match(YAML_FRONTMATTER_REGEX)

  if (!match) {
    return `---\nmode: agent\n---\n\n${content}`
  }

  const [, frontmatterBody = ''] = match
  const body = content.slice(match[0].length)
  const lines = frontmatterBody
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)

  let hasMode = false
  const normalized = lines.map(line => {
    if (/^mode\s*:/i.test(line)) {
      hasMode = true
      return 'mode: agent'
    }

    return line
  })

  if (!hasMode) {
    normalized.push('mode: agent')
  }

  return `---\n${normalized.join('\n')}\n---\n\n${body}`
}
