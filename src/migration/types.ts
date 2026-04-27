/**
 * Migration Engine - Core Types
 * 
 * Type definitions for the migration system that imports existing AI setups
 * into ai-setup format with intelligent merging.
 */

export interface MigrationContext {
  sourcePath: string;
  targetPath: string;
  options: MigrationOptions;
}

export interface MigrationOptions {
  preview?: boolean;
  mergeStrategy: MergeStrategy;
  verbose?: boolean;
  skipBackup?: boolean;
  interactive?: boolean;
}

export type MergeStrategy = 'smart' | 'preserve' | 'replace' | 'append';

export interface DetectionResult {
  detected: boolean;
  confidence: number; // 0-1
  adapterId: string;
  adapterName: string;
  files: DetectedFile[];
  metadata?: Record<string, unknown>;
}

export interface DetectedFile {
  path: string;
  type: 'config' | 'agent' | 'rule' | 'template' | 'command' | 'other';
  priority: number;
}

export interface ParsedSetup {
  projectName?: string;
  description?: string;
  techStack?: TechStack;
  agents: AgentDefinition[];
  rules: RuleDefinition[];
  commands: CommandDefinition[];
  templates: TemplateDefinition[];
  customSections: CustomSection[];
  files: ParsedFile[];
  metadata: Record<string, unknown>;
}

export interface TechStack {
  language?: string;
  framework?: string;
  database?: string;
  orm?: string;
  testing?: string;
  packageManager?: string;
}

export interface AgentDefinition {
  id: string;
  name: string;
  description: string;
  role: string;
  content: string;
  sourcePath: string;
  custom?: boolean;
}

export interface RuleDefinition {
  id: string;
  category: string;
  content: string;
  sourcePath: string;
  priority: number;
}

export interface CommandDefinition {
  id: string;
  name: string;
  description: string;
  content: string;
  sourcePath: string;
}

export interface TemplateDefinition {
  id: string;
  name: string;
  description: string;
  content: string;
  sourcePath: string;
}

export interface CustomSection {
  id: string;
  title: string;
  content: string;
  sourcePath: string;
}

export interface ParsedFile {
  path: string;
  content: string;
  type: string;
}

export interface MergeResult {
  success: boolean;
  merged: boolean;
  conflicts: MergeConflict[];
  backupPaths: string[];
  newFiles: string[];
  modifiedFiles: string[];
  warnings: string[];
  outputPath?: string;
}

export interface MergeConflict {
  file: string;
  lineStart: number;
  lineEnd: number;
  baseContent: string;
  oursContent: string;
  theirsContent: string;
  resolved?: boolean;
  resolution?: string;
  resolvedContent?: string;
}

export interface MigrationPlan {
  sourcePath: string;
  targetPath: string;
  adapters: string[];
  actions: MigrationAction[];
  conflicts: MergeConflict[];
  adapterConflicts: AdapterConflict[];
  estimatedFiles: number;
  estimatedConflicts: number;
  canProceed: boolean;
}

export interface MigrationAction {
  type: 'create' | 'modify' | 'backup' | 'skip';
  sourcePath?: string;
  targetPath: string;
  description: string;
  reason: string;
  conflict?: MergeConflict;
  adapterId?: string;
}

export interface AdapterConflict {
  targetPath: string;
  adapters: string[];
  actions: Pick<MigrationAction, 'type' | 'description' | 'reason' | 'adapterId'>[];
}

export interface MigrationResult {
  success: boolean;
  plan: MigrationPlan | null;
  executedActions: MigrationAction[];
  backupPath?: string;
  errors: string[];
  warnings: string[];
  stats: MigrationStats;
}

export interface MigrationStats {
  filesCreated: number;
  filesModified: number;
  filesBackedUp: number;
  filesSkipped: number;
  conflictsResolved: number;
  conflictsUnresolved: number;
}

export interface DriftCheckResult {
  clean: boolean;
  drifts: DriftItem[];
  missingFiles: string[];
  extraFiles: string[];
  modifiedFiles: ModifiedFile[];
}

export interface DriftItem {
  file: string;
  type: 'missing' | 'extra' | 'modified';
  expectedHash?: string;
  actualHash?: string;
  diff?: string;
}

export interface ModifiedFile {
  path: string;
  expectedHash: string;
  actualHash: string;
  difference: string;
}
