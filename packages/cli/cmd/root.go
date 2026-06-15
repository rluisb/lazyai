package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	clilog "github.com/rluisb/lazyai/packages/cli/internal/log"
	buildversion "github.com/rluisb/lazyai/packages/cli/internal/version"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
// It is kept in cmd for existing release/build scripts and synced into
// internal/version so non-cmd packages can read the same value without cycles.
var Version = "0.0.0-dev"

var (
	logLevel  string
	logFormat string
)

var rootCmd = &cobra.Command{
	Use:     "lazyai-cli",
	Short:   "LazyAI setup-core scaffold",
	Long:    "LazyAI setup-core scaffold — one command to define and refresh your AI toolchain, with transitional runtime extras grouped separately",
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return applyLoggingEnv(loggingFlagConfigFromCommand(cmd))
	},
}

func init() {
	syncBuildVersion()
	rootCmd.Version = Version
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose debug output")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "Set log level (debug|info|warn|error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "Set log format (text|json|logfmt)")

	// Register command groups for organized help output
	rootCmd.AddGroup(&cobra.Group{ID: "lifecycle", Title: "Environment Lifecycle"})
	rootCmd.AddGroup(&cobra.Group{ID: "workspace", Title: "Workspace & Knowledge"})
	rootCmd.AddGroup(&cobra.Group{ID: "runtime", Title: "Optional Runtime Modules"})
	rootCmd.AddGroup(&cobra.Group{ID: "audit", Title: "Audit & Observability"})
	rootCmd.AddGroup(&cobra.Group{ID: "safety", Title: "Safety & Administration"})
	rootCmd.AddGroup(&cobra.Group{ID: "scaffold", Title: "Scaffolding & Discovery"})
	rootCmd.AddGroup(&cobra.Group{ID: "auth", Title: "Authentication"})
	rootCmd.AddGroup(&cobra.Group{ID: "shell", Title: "Shell & Utilities"})

	// Set custom help template with group-aware rendering
	rootCmd.SetHelpTemplate(helpTemplate)
}

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:
  {{.UseLine}}

{{if .HasAvailableSubCommands}}Available Commands:{{range .Groups}}
  {{.Title}}{{range .Commands}}
    {{rpad .Name .NamePadding }} {{.Short}}{{end}}
{{end}}{{end}}{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .HelpCommands}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}
{{end}}{{if .HasAvailableSubCommands}}

Run '{{.CommandPath}} help <command>' for details on a specific command.
{{end}}`

type loggingFlagConfig struct {
	Verbose           bool
	LogLevel          string
	LogLevelExplicit  bool
	LogFormat         string
	LogFormatExplicit bool
}

func loggingFlagConfigFromCommand(cmd *cobra.Command) loggingFlagConfig {
	if cmd == nil || cmd.Root() == nil {
		return loggingFlagConfig{}
	}
	flags := cmd.Root().PersistentFlags()
	verbose, _ := flags.GetBool("verbose")
	level, _ := flags.GetString("log-level")
	format, _ := flags.GetString("log-format")
	return loggingFlagConfig{
		Verbose:           verbose,
		LogLevel:          level,
		LogLevelExplicit:  flags.Changed("log-level"),
		LogFormat:         format,
		LogFormatExplicit: flags.Changed("log-format"),
	}
}

func applyLoggingEnv(config loggingFlagConfig) error {
	if config.LogLevelExplicit {
		if err := os.Setenv("AI_SETUP_LOG_LEVEL", config.LogLevel); err != nil {
			return err
		}
	} else if config.Verbose {
		if err := os.Setenv("AI_SETUP_LOG_LEVEL", "debug"); err != nil {
			return err
		}
	}

	if config.LogFormatExplicit {
		if err := os.Setenv("AI_SETUP_LOG_FORMAT", config.LogFormat); err != nil {
			return err
		}
	}
	clilog.Configure("", "")
	return nil
}

func syncBuildVersion() {
	if Version != buildversion.DevVersion {
		buildversion.Version = Version
		return
	}
	Version = buildversion.Version
}

func Execute(ctx context.Context) error {
	return fang.Execute(ctx, rootCmd,
		fang.WithVersion(Version),
	)
}

// ConfirmAction prompts the user for confirmation before proceeding.
// Returns true if the user confirms, false otherwise.
// If force is true, skips the prompt and returns true.
func ConfirmAction(message string, force bool) bool {
	if force {
		return true
	}
	fmt.Fprintf(os.Stderr, "\n%s [y/N]: ", message)
	var response string
	fmt.Fscanln(os.Stdin, &response)
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
