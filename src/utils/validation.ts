// biome-ignore lint/suspicious/noControlCharactersInRegex: we intentionally reject ASCII control characters in filenames
const INVALID_FILESYSTEM_CHARS = /[<>:"/\\|?*\x00-\x1F]/

export function validateRequiredText(value: string | undefined, fieldLabel = 'Value'): string | undefined {
  if (value === undefined) {
    return `${fieldLabel} cannot be empty`
  }
  const trimmed = value.trim()
  if (!trimmed) {
    return `${fieldLabel} cannot be empty`
  }
  return undefined
}

export function validateFilesystemSafeName(value: string | undefined, fieldLabel = 'Name'): string | undefined {
  const requiredError = validateRequiredText(value, fieldLabel)
  if (requiredError) return requiredError

  if (value && INVALID_FILESYSTEM_CHARS.test(value)) {
    return `${fieldLabel} contains invalid filesystem characters`
  }

  return undefined
}
