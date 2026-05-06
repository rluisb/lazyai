package cmd

import (
	"context"
	"os"

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
	Short:   "LazyAI development environment scaffold",
	Long:    "LazyAI development environment scaffold — one command to set up your AI tools",
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
}

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
