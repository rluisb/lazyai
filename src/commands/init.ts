import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { detectAdapters, importSetup } from '../migration/index.js'
import type { PresetLevel, SetupScope, SetupType, ToolId } from '../types.js'
import { loadConfig } from '../utils/toml.js'
import { runWizard } from '../wizard/index.js'
import { formatAdapterList, MIGRATION_MARKER_HINT } from './migration-shared.js'

// Help text for features - detailed descriptions
const FEATURES_HELP = `
Feature Presets:
  minimal   - Quality gates + git conventions only
  standard  - + Reasoning protocol, RPI workflow, bug resolution (default)
  full      - All prompt engineering features
  custom    - Pick individually (interactive only)

Individual features (for --features / --disable-features):
  contextEngineering  - Context discipline: file budget, session hygiene
  rpiWorkflow         - Research → Plan → Implement workflow
  chainOfThought      - Reasoning protocol: structured <cot> reasoning
  treeOfThoughts      - Decision protocol: evaluate multiple approaches
  adrEnforcement      - Architecture Decision Records for significant changes
  qualityGates        - Lint, typecheck, test, build checks
  agentHarness        - Agent coordination: multi-agent handoff rules
  bugResolution       - Structured debugging: reproduce → diagnose → fix → verify
  pivotHandling       - What to do when plans change mid-implementation

Examples:
  --preset minimal                                  # Lightweight setup
  --preset standard                                 # Recommended (default)
  --preset full                                     # Everything
  --disable-features treeOfThoughts,agentHarness    # Disable advanced features
  --disable-features all --features rpiWorkflow     # Minimal: only RPI workflow
`

interface InitOptions {
  scope?: SetupScope
  type?: SetupType
  planningRepo?: string
  repos?: string
  tools?: string
  cliTools?: string
  name?: string
  force?: boolean
  interactive: boolean
  migrate?: boolean
  from?: string
  absorb?: boolean
  dryRun?: boolean
  planningDir?: string
  preset?: string
  features?: string
  disableFeatures?: string
  branchPattern?: string
  commitPattern?: string
  enableServers?: string
}

export function registerInit(program: Command): void {
  program
    .command('init')
    .description('Scaffold AI development environment in the current directory')
    .option('--scope <scope>', 'Setup scope: global | workspace | project')
    .option('--type <type>', 'Deprecated alias for --scope')
    .option('--planning-repo <path>', 'Planning repo location (workspace scope)')
    .option('--repos <paths>', 'Workspace repo references as comma-separated relative paths')
    .option('--tools <tools>', 'Comma-separated tool list: opencode,claude-code,codex,copilot,gemini')
    .option('--cli-tools <tools>', 'Comma-separated CLI tools available (codegraph,qmd,rtk)')
    .option('--name <name>', 'Project name (defaults to directory name)')
    .option('--force', 'Overwrite all existing managed files (creates backups)')
    .option('--no-interactive', 'Non-interactive mode — requires all flags')
    .option('--migrate', 'Migrate existing AI setup (detects and imports)')
    .option('--from <path>', 'Path to existing setup for migration (defaults to current directory)')
    .option('--absorb', 'Absorb existing tool configs into .ai/ during init')
    .option('--dry-run', 'Preview changes without writing files')
    .option('--planning-dir <dir>', 'Planning directory for specs/ADRs (default: .planning)')
    .option(
      '--preset <level>',
      'Feature preset: minimal | standard | full | custom\n' +
      '  minimal  — quality gates + git conventions only\n' +
      '  standard — + reasoning protocol, RPI workflow, bug resolution (default)\n' +
      '  full     — all prompt engineering features'
    )
    .option(
      '--features <features>',
      'Enable specific features (comma-separated). Use with --disable-features all for minimal setup.\n' +
      'Available: contextEngineering,rpiWorkflow,chainOfThought,treeOfThoughts,adrEnforcement,qualityGates,agentHarness,bugResolution,pivotHandling'
    )
    .option(
      '--disable-features <features>',
      'Disable specific features (comma-separated). Use "all" to disable all, then --features to enable specific ones.\n' +
      'Example: --disable-features treeOfThoughts,agentHarness'
    )
    .option(
      '--branch-pattern <pattern>',
      'Branch naming pattern. Placeholders: {type}, {ticket}, {description}\n' +
      'Default: {type}/{ticket}-{description}  →  feat/PROJ-123-add-login\n' +
      'Options: {type}/{description}, {ticket}/{description}, {description}'
    )
    .option(
      '--commit-pattern <pattern>',
      'Commit message pattern. Placeholders: {type}, {scope}, {ticket}, {description}\n' +
      'Default: {type}({scope}): {description}  →  feat(auth): add login\n' +
      'Options: {type}: {description}, [{ticket}] {description}, {description}'
    )
    .option(
      '--enable-servers <servers>',
      'Comma-separated MCP servers to enable (e.g., atlassian,playwright,orchestrator)'
    )
    .addHelpText('after', FEATURES_HELP)
    .action(async (opts: InitOptions) => {
      const targetDir = process.cwd()
      const tomlConfig = loadConfig(targetDir)

      // TOML fallback: CLI flags override TOML, TOML fills gaps
      const tools = opts.tools
        ? (opts.tools.split(',').map((t) => t.trim()).filter(Boolean) as ToolId[])
        : tomlConfig.default_tools?.length
          ? (tomlConfig.default_tools as ToolId[])
          : undefined
      const repos = opts.repos
        ? opts.repos.split(',').map((repoPath) => repoPath.trim()).filter(Boolean)
        : undefined
      const cliTools = opts.cliTools
        ? opts.cliTools.split(',').map((tool) => tool.trim()).filter(Boolean)
        : undefined
      const features = opts.features
        ? opts.features.split(',').map((f) => f.trim()).filter(Boolean)
        : undefined
      const disableFeatures = opts.disableFeatures
        ? opts.disableFeatures.split(',').map((f) => f.trim()).filter(Boolean)
        : undefined

      const cliOverrides: {
        scope?: SetupScope
        type?: SetupType
        planningRepo?: string
        repos?: string[]
        tools?: ToolId[]
        cliTools?: string[]
        name?: string
        planningDir?: string
        preset?: PresetLevel
        features?: string[]
        disableFeatures?: string[]
        branchPattern?: string
        commitPattern?: string
        enableServers?: string[]
      } = {}

      if (opts.scope) cliOverrides.scope = opts.scope
      else if (tomlConfig.default_scope) cliOverrides.scope = tomlConfig.default_scope
      if (opts.type) cliOverrides.type = opts.type
      if (opts.planningRepo) cliOverrides.planningRepo = opts.planningRepo
      if (repos) cliOverrides.repos = repos
      if (tools) cliOverrides.tools = tools
      if (cliTools) cliOverrides.cliTools = cliTools
      if (opts.name) cliOverrides.name = opts.name
      else if (tomlConfig.project_name) cliOverrides.name = tomlConfig.project_name
      if (opts.planningDir) cliOverrides.planningDir = opts.planningDir
      if (opts.preset) cliOverrides.preset = opts.preset as PresetLevel
      if (features) cliOverrides.features = features
      if (disableFeatures) cliOverrides.disableFeatures = disableFeatures
      if (opts.branchPattern) cliOverrides.branchPattern = opts.branchPattern
      if (opts.commitPattern) cliOverrides.commitPattern = opts.commitPattern
      if (opts.enableServers) cliOverrides.enableServers = opts.enableServers.split(',').map((s) => s.trim()).filter(Boolean)

      // Check if we should migrate
      if (opts.migrate) {
        const sourcePath = opts.from || process.cwd()
        
        p.intro(pc.blue('ai-setup init --migrate'))
        
        const spinner = p.spinner()
        spinner.start('Detecting existing AI setups...')
        
        const adapters = await detectAdapters(sourcePath)
        
        if (adapters.length === 0) {
          spinner.stop(pc.yellow('No supported AI setup detected'))
          console.log(pc.gray(`Searched in: ${sourcePath}`))
          console.log(pc.gray(MIGRATION_MARKER_HINT))
          console.log(pc.gray('Continuing with fresh init...'))
        } else {
          spinner.stop(pc.green(`Detected ${adapters.length} setup(s): ${formatAdapterList(adapters)}`))
          
          // Run migration
          spinner.start('Migrating existing setup...')
          
          const result = await importSetup({
            path: sourcePath,
            preview: false,
            mergeStrategy: 'smart',
          })
          
          spinner.stop('Migration complete')
          
          if (result.success) {
            console.log(pc.green('\n✅ Successfully migrated existing setup!'))
            console.log(pc.gray(`\nMigrated ${result.stats.filesCreated + result.stats.filesModified} file(s)`))
            
            if (result.backupPath) {
              console.log(pc.gray(`Backup created at: ${result.backupPath}`))
            }
            
            // Continue with wizard for any additional configuration
            console.log(pc.blue('\nContinuing with init wizard for additional configuration...'))
          } else {
            console.log(pc.yellow('\n⚠️  Migration had issues, continuing with fresh init...'))
            if (result.errors.length > 0) {
              console.log(pc.gray(`Errors: ${result.errors.join(', ')}`))
            }
          }
        }
      }

      const wizardOpts = {
        interactive: opts.interactive !== false,
        cliOverrides,
        targetDir: process.cwd(),
        ...(opts.absorb !== undefined ? { absorb: opts.absorb } : {}),
        ...(opts.force !== undefined ? { force: opts.force } : {}),
        ...(opts.dryRun ? { dryRun: true } : {}),
      }

      await runWizard(wizardOpts)
    })
}
