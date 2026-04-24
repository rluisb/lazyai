import type { ArtifactType } from '../types.js'

export interface PromptQuestion {
  key: string
  label: string
  type: 'text' | 'select' | 'multiselect'
  options?: { value: string; label: string }[]
  required?: boolean
  default?: string
}

export interface GeneratorConfig {
  name: string
  description?: string
  targetDir: string
  force?: boolean
  answers?: Record<string, unknown>
}

export interface GeneratedFile {
  path: string
  content: string
}

export interface Generator {
  type: ArtifactType
  generate(config: GeneratorConfig): Promise<GeneratedFile[]>
  getPromptQuestions(): PromptQuestion[]
}
