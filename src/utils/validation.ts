const INVALID_FILESYSTEM_CHARS = /[<>:"/\\|?*\x00-\x1F]/

export function validateRequiredText(value: string, fieldLabel = 'Value'): string | undefined {
  const trimmed = value.trim()
  if (!trimmed) {
    return `${fieldLabel} cannot be empty`
  }
  return undefined
}

export function validateFilesystemSafeName(value: string, fieldLabel = 'Name'): string | undefined {
  const requiredError = validateRequiredText(value, fieldLabel)
  if (requiredError) return requiredError

  if (INVALID_FILESYSTEM_CHARS.test(value)) {
    return `${fieldLabel} contains invalid filesystem characters`
  }

  return undefined
}
