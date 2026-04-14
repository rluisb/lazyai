const YAML_FRONTMATTER_REGEX = /^---\s*\n([\s\S]*?)\n---\s*(?:\n|$)/

export interface AgentFrontmatter {
  name?: string
  model?: string
  tools?: string[]
  description?: string
}

interface SplitResult {
  frontmatterBody: string
  body: string
}

export function splitYamlFrontmatter(content: string): SplitResult | null {
  const match = content.match(YAML_FRONTMATTER_REGEX)
  if (!match) return null

  const [, frontmatterBody = ''] = match
  const body = content.slice(match[0].length)

  return { frontmatterBody, body }
}

export function stripYamlFrontmatter(content: string): string {
  return content.replace(YAML_FRONTMATTER_REGEX, '')
}

export function extractTools(content: string): string[] {
  const split = splitYamlFrontmatter(content)
  if (!split) return []
  return parseToolsFromFrontmatterBody(split.frontmatterBody)
}

export function parseAgentFrontmatter(content: string): AgentFrontmatter | null {
  const split = splitYamlFrontmatter(content)
  if (!split) return null

  const body = split.frontmatterBody
  const parsed: AgentFrontmatter = {}

  const name = matchSimpleKey(body, 'name')
  if (name) parsed.name = name

  const model = matchSimpleKey(body, 'model')
  if (model) parsed.model = model

  const description = matchSimpleKey(body, 'description')
  if (description) parsed.description = description

  const tools = parseToolsFromFrontmatterBody(body)
  if (tools.length > 0) parsed.tools = tools

  return parsed
}

export function normalizeToolsFrontmatter(
  content: string,
  delimiter: 'comma' | 'space',
): string {
  const split = splitYamlFrontmatter(content)
  if (!split) return content

  const tools = parseToolsFromFrontmatterBody(split.frontmatterBody)
  if (tools.length === 0) return content

  const joined = delimiter === 'comma' ? tools.join(', ') : tools.join(' ')
  const rebuilt = rewriteToolsLine(split.frontmatterBody, joined)

  return `---\n${rebuilt}\n---\n${split.body.startsWith('\n') ? split.body : `\n${split.body}`}`
}

export function stripFrontmatterAndInjectModel(content: string): string {
  const split = splitYamlFrontmatter(content)
  if (!split) return content

  const comments: string[] = []

  const modelMatch = split.frontmatterBody.match(/^model\s*:\s*(.+)$/m)
  if (modelMatch?.[1]) {
    comments.push(`<!-- Recommended model: ${modelMatch[1].trim()} -->`)
  }

  const tools = parseToolsFromFrontmatterBody(split.frontmatterBody)
  if (tools.length > 0) {
    comments.push(`<!-- allowed-tools: ${tools.join(', ')} -->`)
  }

  if (comments.length === 0) return split.body

  return `${comments.join('\n')}\n\n${split.body.replace(/^\n+/, '')}`
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

function matchSimpleKey(body: string, key: string): string | undefined {
  const regex = new RegExp(`^${key}\\s*:\\s*(.+)$`, 'm')
  const match = body.match(regex)
  return match?.[1]?.trim()
}

function parseToolsFromFrontmatterBody(body: string): string[] {
  const lines = body.split('\n')
  const toolsLineIndex = lines.findIndex(line => /^tools\s*:/.test(line))
  if (toolsLineIndex === -1) return []

  const toolsLine = lines[toolsLineIndex] ?? ''
  const afterColon = toolsLine.slice(toolsLine.indexOf(':') + 1).trim()

  if (afterColon.length > 0) {
    return splitToolList(afterColon)
  }

  const listItems: string[] = []
  for (let i = toolsLineIndex + 1; i < lines.length; i++) {
    const line = lines[i] ?? ''
    const match = line.match(/^\s*-\s+(.+)$/)
    if (!match) break
    const value = match[1]?.trim()
    if (value) listItems.push(value)
  }
  return listItems
}

function splitToolList(value: string): string[] {
  return value
    .split(/[\s,]+/)
    .map(part => part.trim())
    .filter(part => part.length > 0)
}

function rewriteToolsLine(frontmatterBody: string, joined: string): string {
  const lines = frontmatterBody.split('\n')
  const toolsLineIndex = lines.findIndex(line => /^tools\s*:/.test(line))
  if (toolsLineIndex === -1) return frontmatterBody

  const removed: string[] = []
  removed.push(...lines.slice(0, toolsLineIndex))
  removed.push(`tools: ${joined}`)

  let i = toolsLineIndex + 1
  while (i < lines.length && /^\s*-\s+/.test(lines[i] ?? '')) {
    i++
  }
  removed.push(...lines.slice(i))
  return removed.join('\n')
}
