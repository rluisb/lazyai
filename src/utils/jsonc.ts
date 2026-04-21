/**
 * Strip line (`// ...`) and block (`/* ... *\/`) comments from a JSONC string
 * without mangling comment-like sequences that appear inside string literals
 * (e.g. URLs such as "https://example.com").
 */
export function stripJsonComments(input: string): string {
  let out = ''
  let i = 0
  let inString = false
  while (i < input.length) {
    const c = input[i]
    const next = input[i + 1]

    if (inString) {
      out += c
      if (c === '\\' && next !== undefined) {
        out += next
        i += 2
        continue
      }
      if (c === '"') inString = false
      i++
      continue
    }

    if (c === '"') {
      inString = true
      out += c
      i++
      continue
    }

    if (c === '/' && next === '/') {
      while (i < input.length && input[i] !== '\n') i++
      continue
    }

    if (c === '/' && next === '*') {
      i += 2
      while (i < input.length - 1 && !(input[i] === '*' && input[i + 1] === '/')) i++
      i += 2
      continue
    }

    out += c
    i++
  }
  return out
}
