package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// LazyAIConfig represents the configuration structure
type LazyAIConfig struct {
	Agent         AgentConfig        `yaml:"agent"`
	Database      DatabaseConfig     `yaml:"database"`
	Ledger        LedgerConfig       `yaml:"ledger"`
	Workflows     WorkflowsConfig    `yaml:"workflows"`
	Notifications NotificationConfig `yaml:"notifications"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	DefaultAgent string `yaml:"default_agent"`
	DefaultModel string `yaml:"default_model"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// LedgerConfig represents ledger configuration
type LedgerConfig struct {
	Path string `yaml:"path"`
}

// WorkflowsConfig represents workflow configuration
type WorkflowsConfig struct {
	Directory string `yaml:"directory"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configuration management",
	Long:    `Get and set LazyAI configuration values.`,
	GroupID: "workspace",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get a configuration value by key (e.g., agent.default_agent).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		config, err := loadConfig()
		if err != nil {
			return err
		}

		value, err := getConfigValue(config, key)
		if err != nil {
			return err
		}

		fmt.Printf("%s: %v\n", key, value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value by key (e.g., agent.default_agent builder).`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		config, err := loadConfig()
		if err != nil {
			// If config doesn't exist, create a new one
			config = &LazyAIConfig{
				Agent: AgentConfig{
					DefaultAgent: "primary-agent",
					DefaultModel: "gpt-4",
				},
				Database: DatabaseConfig{
					Path: ".specify/lazyai.db",
				},
				Ledger: LedgerConfig{
					Path: ".specify/ledger.jsonl",
				},
				Workflows: WorkflowsConfig{
					Directory: ".opencode/workflows",
				},
				Notifications: NotificationConfig{
					Enabled: false,
				},
			}
		}

		// Warn for critical configuration keys
		criticalKeys := []string{"agent.default_agent", "agent.default_model", "database.path", "ledger.path", "workflows.directory", "notifications.webhook"}
		for _, ck := range criticalKeys {
			if key == ck {
				fmt.Fprintf(os.Stderr, "⚠️  Warning: '%s' is a critical configuration key. Changing it may affect LazyAI behavior.\n", key)
				break
			}
		}

		if err := setConfigValue(config, key, value); err != nil {
			return err
		}

		if err := saveConfig(config); err != nil {
			return err
		}

		fmt.Printf("✅ Set %s = %s\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Show all current configuration values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		fmt.Println("Current Configuration:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		fmt.Printf("agent.default_agent: %s\n", config.Agent.DefaultAgent)
		fmt.Printf("agent.default_model: %s\n", config.Agent.DefaultModel)
		fmt.Printf("database.path: %s\n", config.Database.Path)
		fmt.Printf("ledger.path: %s\n", config.Ledger.Path)
		fmt.Printf("workflows.directory: %s\n", config.Workflows.Directory)
		fmt.Printf("notifications.enabled: %v\n", config.Notifications.Enabled)
		if config.Notifications.Webhook != "" {
			fmt.Printf("notifications.webhook: %s\n", config.Notifications.Webhook)
		}

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file with defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := filepath.Join(".opencode", "config.yaml")

		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s", configPath)
		}

		config := &LazyAIConfig{
			Agent: AgentConfig{
				DefaultAgent: "primary-agent",
				DefaultModel: "gpt-4",
			},
			Database: DatabaseConfig{
				Path: ".specify/lazyai.db",
			},
			Ledger: LedgerConfig{
				Path: ".specify/ledger.jsonl",
			},
			Workflows: WorkflowsConfig{
				Directory: ".opencode/workflows",
			},
			Notifications: NotificationConfig{
				Enabled: false,
			},
		}

		if err := saveConfig(config); err != nil {
			return err
		}

		fmt.Printf("✅ Configuration initialized at %s\n", configPath)
		return nil
	},
}

// loadConfig loads the configuration from file
func loadConfig() (*LazyAIConfig, error) {
	configPath := filepath.Join(".opencode", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found. Run 'lazyai-cli config init' first")
		}
		return nil, err
	}

	var config LazyAIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	// Apply environment variable overrides
	if agent := os.Getenv("LAZYAI_AGENT"); agent != "" {
		config.Agent.DefaultAgent = agent
	}
	if model := os.Getenv("LAZYAI_MODEL"); model != "" {
		config.Agent.DefaultModel = model
	}
	if dbPath := os.Getenv("LAZYAI_DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if ledgerPath := os.Getenv("LAZYAI_LEDGER_PATH"); ledgerPath != "" {
		config.Ledger.Path = ledgerPath
	}

	return &config, nil
}

// saveConfig saves the configuration to file
func saveConfig(config *LazyAIConfig) error {
	configPath := filepath.Join(".opencode", "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	return nil
}

// getConfigValue gets a configuration value by key
func getConfigValue(config *LazyAIConfig, key string) (string, error) {
	switch key {
	case "agent.default_agent":
		return config.Agent.DefaultAgent, nil
	case "agent.default_model":
		return config.Agent.DefaultModel, nil
	case "database.path":
		return config.Database.Path, nil
	case "ledger.path":
		return config.Ledger.Path, nil
	case "workflows.directory":
		return config.Workflows.Directory, nil
	case "notifications.enabled":
		return fmt.Sprintf("%v", config.Notifications.Enabled), nil
	case "notifications.webhook":
		return config.Notifications.Webhook, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// setConfigValue sets a configuration value by key
func setConfigValue(config *LazyAIConfig, key, value string) error {
	switch key {
	case "agent.default_agent":
		config.Agent.DefaultAgent = value
	case "agent.default_model":
		config.Agent.DefaultModel = value
	case "database.path":
		config.Database.Path = value
	case "ledger.path":
		config.Ledger.Path = value
	case "workflows.directory":
		config.Workflows.Directory = value
	case "notifications.enabled":
		config.Notifications.Enabled = value == "true"
	case "notifications.webhook":
		config.Notifications.Webhook = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
