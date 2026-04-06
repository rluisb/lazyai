/**
 * Completions Command
 *
 * Generates shell completion scripts for bash, zsh, and fish.
 */

import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'

type Shell = 'bash' | 'zsh' | 'fish'

const COMMANDS = ['init', 'add', 'update', 'doctor', 'status', 'compile', 'eject', 'list', 'info', 'create', 'completions']

const TOOLS = ['opencode', 'claude-code', 'pi', 'copilot', 'gemini', 'codex']

const SCOPES = ['project', 'global', 'workspace']

const LIST_CATEGORIES = ['agents', 'skills', 'templates', 'rules', 'mcp', 'cli', 'servers', 'tools']

function generateBashCompletions(): string {
  return `# ai-setup bash completion
# Add to ~/.bashrc or ~/.bash_profile:
#   eval "$(ai-setup completions bash)"

_ai_setup_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="\${COMP_WORDS[COMP_CWORD]}"
    prev="\${COMP_WORDS[COMP_CWORD-1]}"

    # Commands
    local commands="${COMMANDS.join(' ')}"

    # Tools
    local tools="${TOOLS.join(' ')}"

    # Scopes
    local scopes="${SCOPES.join(' ')}"

    # List categories
    local categories="${LIST_CATEGORIES.join(' ')}"

    case "\${prev}" in
        ai-setup)
            COMPREPLY=( $(compgen -W "\${commands}" -- "\${cur}") )
            return 0
            ;;
        init)
            COMPREPLY=( $(compgen -W "--tools --scope --non-interactive --enable-servers --help" -- "\${cur}") )
            return 0
            ;;
        add)
            COMPREPLY=( $(compgen -W "\${tools}" -- "\${cur}") )
            return 0
            ;;
        --tools)
            COMPREPLY=( $(compgen -W "\${tools}" -- "\${cur}") )
            return 0
            ;;
        --scope)
            COMPREPLY=( $(compgen -W "\${scopes}" -- "\${cur}") )
            return 0
            ;;
        list)
            COMPREPLY=( $(compgen -W "\${categories} --json --enabled" -- "\${cur}") )
            return 0
            ;;
        info)
            # Complete with common items
            COMPREPLY=( $(compgen -W "builder planner reviewer scout documenter red-team memory filesystem ripgrep --json" -- "\${cur}") )
            return 0
            ;;
        doctor)
            COMPREPLY=( $(compgen -W "--migration-check --verbose --json" -- "\${cur}") )
            return 0
            ;;
        status)
            COMPREPLY=( $(compgen -W "--json" -- "\${cur}") )
            return 0
            ;;
        compile)
            COMPREPLY=( $(compgen -W "--scope --dry-run" -- "\${cur}") )
            return 0
            ;;
        update)
            COMPREPLY=( $(compgen -W "--force --dry-run" -- "\${cur}") )
            return 0
            ;;
        create)
            COMPREPLY=( $(compgen -W "agent skill prompt" -- "\${cur}") )
            return 0
            ;;
        completions)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "\${cur}") )
            return 0
            ;;
        *)
            ;;
    esac

    # Default to commands
    COMPREPLY=( $(compgen -W "\${commands}" -- "\${cur}") )
    return 0
}

complete -F _ai_setup_completions ai-setup
`
}

function generateZshCompletions(): string {
  return `#compdef ai-setup
# ai-setup zsh completion
# Add to ~/.zshrc:
#   eval "$(ai-setup completions zsh)"

_ai_setup() {
    local -a commands tools scopes categories

    commands=(
        'init:Initialize ai-setup in current directory'
        'add:Add a tool to existing setup'
        'update:Update managed files'
        'doctor:Verify setup integrity'
        'status:Show current setup status'
        'compile:Regenerate tool-specific files'
        'eject:Remove ai-setup management'
        'list:List available library items'
        'info:Show detailed info about a library item'
        'create:Create a new agent, skill, or prompt'
        'completions:Generate shell completions'
    )

    tools=(
        'opencode:OpenCode AI assistant'
        'claude-code:Claude Code assistant'
        'pi:Pi assistant'
        'copilot:GitHub Copilot'
        'gemini:Google Gemini'
        'codex:OpenAI Codex'
    )

    scopes=(
        'project:Project-local setup'
        'global:User-global setup'
        'workspace:Multi-repo workspace'
    )

    categories=(
        'agents:List available agents'
        'skills:List available skills'
        'templates:List available templates'
        'rules:List available rules'
        'mcp:List MCP servers'
        'cli:List CLI tools'
    )

    case "\$words[2]" in
        init)
            _arguments \\
                '--tools[Tools to install]:tool:->tools' \\
                '--scope[Setup scope]:scope:->scopes' \\
                '--non-interactive[Run without prompts]' \\
                '--enable-servers[Enable MCP servers]:servers:' \\
                '--help[Show help]'
            case "\$state" in
                tools)
                    _describe 'tool' tools
                    ;;
                scopes)
                    _describe 'scope' scopes
                    ;;
            esac
            ;;
        add)
            _describe 'tool' tools
            ;;
        list)
            _arguments \\
                '1:category:->categories' \\
                '--json[Output as JSON]' \\
                '--enabled[Show only enabled items]'
            case "\$state" in
                categories)
                    _describe 'category' categories
                    ;;
            esac
            ;;
        info)
            _arguments \\
                '1:item:' \\
                '--json[Output as JSON]'
            ;;
        doctor)
            _arguments \\
                '--migration-check[Check for drift]' \\
                '--verbose[Show detailed output]' \\
                '--json[Output as JSON]'
            ;;
        status)
            _arguments '--json[Output as JSON]'
            ;;
        compile)
            _arguments \\
                '--scope[Compile scope]:scope:->scopes' \\
                '--dry-run[Preview without writing]'
            case "\$state" in
                scopes)
                    _describe 'scope' scopes
                    ;;
            esac
            ;;
        update)
            _arguments \\
                '--force[Force update modified files]' \\
                '--dry-run[Preview without writing]'
            ;;
        create)
            _arguments '1:type:(agent skill prompt)'
            ;;
        completions)
            _arguments '1:shell:(bash zsh fish)'
            ;;
        *)
            _describe 'command' commands
            ;;
    esac
}

compdef _ai_setup ai-setup
`
}

function generateFishCompletions(): string {
  return `# ai-setup fish completion
# Add to ~/.config/fish/completions/ai-setup.fish:
#   ai-setup completions fish > ~/.config/fish/completions/ai-setup.fish

# Disable file completion by default
complete -c ai-setup -f

# Commands
complete -c ai-setup -n "__fish_use_subcommand" -a init -d "Initialize ai-setup in current directory"
complete -c ai-setup -n "__fish_use_subcommand" -a add -d "Add a tool to existing setup"
complete -c ai-setup -n "__fish_use_subcommand" -a update -d "Update managed files"
complete -c ai-setup -n "__fish_use_subcommand" -a doctor -d "Verify setup integrity"
complete -c ai-setup -n "__fish_use_subcommand" -a status -d "Show current setup status"
complete -c ai-setup -n "__fish_use_subcommand" -a compile -d "Regenerate tool-specific files"
complete -c ai-setup -n "__fish_use_subcommand" -a eject -d "Remove ai-setup management"
complete -c ai-setup -n "__fish_use_subcommand" -a list -d "List available library items"
complete -c ai-setup -n "__fish_use_subcommand" -a info -d "Show detailed info about a library item"
complete -c ai-setup -n "__fish_use_subcommand" -a create -d "Create a new agent, skill, or prompt"
complete -c ai-setup -n "__fish_use_subcommand" -a completions -d "Generate shell completions"

# init options
complete -c ai-setup -n "__fish_seen_subcommand_from init" -l tools -d "Tools to install" -xa "${TOOLS.join(' ')}"
complete -c ai-setup -n "__fish_seen_subcommand_from init" -l scope -d "Setup scope" -xa "${SCOPES.join(' ')}"
complete -c ai-setup -n "__fish_seen_subcommand_from init" -l non-interactive -d "Run without prompts"
complete -c ai-setup -n "__fish_seen_subcommand_from init" -l enable-servers -d "Enable MCP servers"

# add tool completions
complete -c ai-setup -n "__fish_seen_subcommand_from add" -xa "${TOOLS.join(' ')}"

# list categories
complete -c ai-setup -n "__fish_seen_subcommand_from list" -xa "${LIST_CATEGORIES.join(' ')}"
complete -c ai-setup -n "__fish_seen_subcommand_from list" -l json -d "Output as JSON"
complete -c ai-setup -n "__fish_seen_subcommand_from list" -l enabled -d "Show only enabled items"

# info options
complete -c ai-setup -n "__fish_seen_subcommand_from info" -l json -d "Output as JSON"

# doctor options
complete -c ai-setup -n "__fish_seen_subcommand_from doctor" -l migration-check -d "Check for drift"
complete -c ai-setup -n "__fish_seen_subcommand_from doctor" -l verbose -d "Show detailed output"
complete -c ai-setup -n "__fish_seen_subcommand_from doctor" -l json -d "Output as JSON"

# status options
complete -c ai-setup -n "__fish_seen_subcommand_from status" -l json -d "Output as JSON"

# compile options
complete -c ai-setup -n "__fish_seen_subcommand_from compile" -l scope -d "Compile scope" -xa "${SCOPES.join(' ')}"
complete -c ai-setup -n "__fish_seen_subcommand_from compile" -l dry-run -d "Preview without writing"

# update options
complete -c ai-setup -n "__fish_seen_subcommand_from update" -l force -d "Force update modified files"
complete -c ai-setup -n "__fish_seen_subcommand_from update" -l dry-run -d "Preview without writing"

# create types
complete -c ai-setup -n "__fish_seen_subcommand_from create" -xa "agent skill prompt"

# completions shells
complete -c ai-setup -n "__fish_seen_subcommand_from completions" -xa "bash zsh fish"
`
}

export function registerCompletions(program: Command): void {
  program
    .command('completions [shell]')
    .description('Generate shell completion scripts')
    .action((shell?: Shell) => {
      if (!shell) {
        // Show help if no shell specified
        p.intro(pc.bold('ai-setup completions'))
        console.log('')
        p.log.info('Generate shell completion scripts for ai-setup commands.')
        console.log('')
        p.log.step('Usage')
        p.log.message(`  ${pc.cyan('ai-setup completions bash')}  # Bash completions`)
        p.log.message(`  ${pc.cyan('ai-setup completions zsh')}   # Zsh completions`)
        p.log.message(`  ${pc.cyan('ai-setup completions fish')}  # Fish completions`)
        console.log('')
        p.log.step('Installation')
        console.log('')
        p.log.message(pc.bold('  Bash:'))
        p.log.message(`    ${pc.dim('# Add to ~/.bashrc or ~/.bash_profile:')}`)
        p.log.message(`    ${pc.cyan('eval "$(ai-setup completions bash)"')}`)
        console.log('')
        p.log.message(pc.bold('  Zsh:'))
        p.log.message(`    ${pc.dim('# Add to ~/.zshrc:')}`)
        p.log.message(`    ${pc.cyan('eval "$(ai-setup completions zsh)"')}`)
        console.log('')
        p.log.message(pc.bold('  Fish:'))
        p.log.message(`    ${pc.dim('# Save to completions directory:')}`)
        p.log.message(`    ${pc.cyan('ai-setup completions fish > ~/.config/fish/completions/ai-setup.fish')}`)
        console.log('')
        p.outro('Choose a shell to generate completions')
        return
      }

      switch (shell) {
        case 'bash':
          console.log(generateBashCompletions())
          break
        case 'zsh':
          console.log(generateZshCompletions())
          break
        case 'fish':
          console.log(generateFishCompletions())
          break
        default:
          p.log.error(`Unknown shell: ${shell}`)
          p.log.info('Supported shells: bash, zsh, fish')
          process.exitCode = 1
      }
    })
}
