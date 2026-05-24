package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for LazyAI CLI.

To load completions:

Bash:
  $ source <(lazyai-cli completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ lazyai-cli completion bash > /etc/bash_completion.d/lazyai-cli
  # macOS:
  $ lazyai-cli completion bash > $(brew --prefix)/etc/bash_completion.d/lazyai-cli

Zsh:
  $ source <(lazyai-cli completion zsh)
  # To load completions for each session, execute once:
  $ lazyai-cli completion zsh > "${fpath[1]}/_lazyai-cli"

Fish:
  $ lazyai-cli completion fish | source
  # To load completions for each session, execute once:
  $ lazyai-cli completion fish > ~/.config/fish/completions/lazyai-cli.fish

PowerShell:
  PS> lazyai-cli completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> lazyai-cli completion powershell > lazyai-cli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
