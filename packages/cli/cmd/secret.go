package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Secret management",
	Long:  `Securely store and retrieve secrets using OS keychain.`,
}

var secretSetCmd = &cobra.Command{
	Use:   "set [name] [value]",
	Short: "Store a secret",
	Long:  `Encrypt and store a secret in the OS keychain.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		value := args[1]
		
		if err := storeSecret(name, value); err != nil {
			return err
		}
		
		fmt.Printf("✅ Secret stored: %s\n", name)
		return nil
	},
}

var secretGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Retrieve a secret",
	Long:  `Retrieve a secret from the OS keychain.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		
		value, err := retrieveSecret(name)
		if err != nil {
			return err
		}
		
		fmt.Printf("%s: %s\n", name, value)
		return nil
	},
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored secrets",
	Long:  `List all secrets stored in the OS keychain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		secrets, err := listSecrets()
		if err != nil {
			return err
		}
		
		fmt.Println("Stored Secrets:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		
		if len(secrets) == 0 {
			fmt.Println("  No secrets found.")
			return nil
		}
		
		for _, secret := range secrets {
			fmt.Printf("  • %s\n", secret)
		}
		
		fmt.Printf("\nTotal: %d secrets\n", len(secrets))
		return nil
	},
}

var secretRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a secret",
	Long:  `Delete a secret from the OS keychain.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		
		if err := removeSecret(name); err != nil {
			return err
		}
		
		fmt.Printf("✅ Secret removed: %s\n", name)
		return nil
	},
}

// storeSecret stores a secret in the OS keychain
func storeSecret(name, value string) error {
	service := "lazyai-cli"
	
	switch runtime.GOOS {
	case "darwin":
		// macOS Keychain
		cmd := exec.Command("security", "add-generic-password",
			"-s", service,
			"-a", name,
			"-w", value,
			"-U") // Update if exists
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to store secret: %w (output: %s)", err, string(output))
		}
		return nil
		
	case "linux":
		// Linux: Try secret-tool (libsecret)
		cmd := exec.Command("secret-tool", "store", "--label=lazyai-cli",
			"service", service,
			"account", name)
		cmd.Stdin = strings.NewReader(value)
		_, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to file-based storage with warnings
			return storeSecretFallback(name, value)
		}
		return nil
		
	default:
		return storeSecretFallback(name, value)
	}
}

// retrieveSecret retrieves a secret from the OS keychain
func retrieveSecret(name string) (string, error) {
	service := "lazyai-cli"
	
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "find-generic-password",
			"-s", service,
			"-a", name,
			"-w") // Output password only
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("secret not found: %s", name)
		}
		return strings.TrimSpace(string(output)), nil
		
	case "linux":
		cmd := exec.Command("secret-tool", "lookup",
			"service", service,
			"account", name)
		output, err := cmd.Output()
		if err != nil {
			return retrieveSecretFallback(name)
		}
		return strings.TrimSpace(string(output)), nil
		
	default:
		return retrieveSecretFallback(name)
	}
}

// listSecrets lists all secrets
func listSecrets() ([]string, error) {
	service := "lazyai-cli"
	
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "dump-keychain")
		output, err := cmd.Output()
		if err != nil {
			return listSecretsFallback()
		}
		
		var secrets []string
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "svce") && strings.Contains(line, service) {
				// Extract account name
				if idx := strings.Index(line, "acct"); idx >= 0 {
					parts := strings.Split(line[idx:], "\"")
					if len(parts) >= 2 {
						secrets = append(secrets, parts[1])
					}
				}
			}
		}
		return secrets, nil
		
	default:
		return listSecretsFallback()
	}
}

// removeSecret removes a secret
func removeSecret(name string) error {
	service := "lazyai-cli"
	
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "delete-generic-password",
			"-s", service,
			"-a", name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove secret: %w (output: %s)", err, string(output))
		}
		return nil
		
	case "linux":
		cmd := exec.Command("secret-tool", "clear",
			"service", service,
			"account", name)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return removeSecretFallback(name)
		}
		return nil
		
	default:
		return removeSecretFallback(name)
	}
}

// Fallback implementations using file-based storage
func storeSecretFallback(name, value string) error {
	fmt.Println("⚠️  Using fallback file-based storage (not secure)")
	
	// Store in ~/.lazyai/secrets/ with base64 encoding
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	secretsDir := fmt.Sprintf("%s/.lazyai/secrets", homeDir)
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return err
	}
	
	secretPath := fmt.Sprintf("%s/%s", secretsDir, name)
	return os.WriteFile(secretPath, []byte(value), 0600)
}

func retrieveSecretFallback(name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	secretPath := fmt.Sprintf("%s/.lazyai/secrets/%s", homeDir, name)
	data, err := os.ReadFile(secretPath)
	if err != nil {
		return "", fmt.Errorf("secret not found: %s", name)
	}
	
	return string(data), nil
}

func listSecretsFallback() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	secretsDir := fmt.Sprintf("%s/.lazyai/secrets", homeDir)
	entries, err := os.ReadDir(secretsDir)
	if err != nil {
		return []string{}, nil
	}
	
	var secrets []string
	for _, entry := range entries {
		if !entry.IsDir() {
			secrets = append(secrets, entry.Name())
		}
	}
	
	return secrets, nil
}

func removeSecretFallback(name string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	secretPath := fmt.Sprintf("%s/.lazyai/secrets/%s", homeDir, name)
	return os.Remove(secretPath)
}

func init() {
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretRemoveCmd)
	rootCmd.AddCommand(secretCmd)
}
