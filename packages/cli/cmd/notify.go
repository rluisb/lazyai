package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Notification system",
	Long:  `Configure and send notifications.`,
}

var notifySendCmd = &cobra.Command{
	Use:   "send [message]",
	Short: "Send a notification",
	Long:  `Send a desktop notification.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := strings.Join(args, " ")
		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			title = "LazyAI"
		}

		// Try desktop notification first
		if err := sendDesktopNotification(title, message); err != nil {
			// Fallback to webhook if configured
			webhook := os.Getenv("LAZYAI_WEBHOOK")
			if webhook != "" {
				return sendWebhookNotification(webhook, title, message)
			}
			return fmt.Errorf("failed to send notification: %w", err)
		}

		fmt.Println("✅ Notification sent")
		return nil
	},
}

var notifyConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure notifications",
	Long:  `Configure notification settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		webhook, _ := cmd.Flags().GetString("webhook")
		enabled, _ := cmd.Flags().GetBool("enabled")

		// Update config
		config, err := loadConfig()
		if err != nil {
			return err
		}

		if webhook != "" {
			config.Notifications.Webhook = webhook
			fmt.Printf("✅ Webhook configured: %s\n", webhook)
		}

		config.Notifications.Enabled = enabled
		if enabled {
			fmt.Println("✅ Notifications enabled")
		} else {
			fmt.Println("✅ Notifications disabled")
		}

		if err := saveConfig(config); err != nil {
			return err
		}

		return nil
	},
}

var notifyTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test notification",
	Long:  `Send a test notification to verify configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🧪 Testing notification system...")

		if err := sendDesktopNotification("LazyAI Test", "If you see this, notifications are working!"); err != nil {
			fmt.Printf("⚠️  Desktop notification failed: %v\n", err)

			// Try webhook
			webhook := os.Getenv("LAZYAI_WEBHOOK")
			if webhook == "" {
				config, _ := loadConfig()
				if config != nil {
					webhook = config.Notifications.Webhook
				}
			}

			if webhook != "" {
				if err := sendWebhookNotification(webhook, "LazyAI Test", "Webhook test"); err != nil {
					return fmt.Errorf("webhook also failed: %w", err)
				}
				fmt.Println("✅ Webhook notification sent")
				return nil
			}

			return fmt.Errorf("no notification method available")
		}

		fmt.Println("✅ Desktop notification sent")
		return nil
	},
}

// sendDesktopNotification sends a desktop notification. User-provided title and
// message are escaped for the target platform's command interpreter so that
// quotes or other metacharacters cannot break or inject into the command.
func sendDesktopNotification(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS notification via AppleScript double-quoted string literals.
		script := fmt.Sprintf(`display notification "%s" with title "%s"`,
			escapeAppleScript(message), escapeAppleScript(title))
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()

	case "linux":
		// notify-send / zenity receive title+message as separate argv entries,
		// so no shell interpolation occurs.
		cmd := exec.Command("notify-send", title, message)
		if err := cmd.Run(); err != nil {
			// Fallback to zenity
			cmd = exec.Command("zenity", "--info", "--title="+title, "--text="+message)
			return cmd.Run()
		}
		return nil

	case "windows":
		// Windows notification via PowerShell double-quoted string literals.
		psScript := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show("%s", "%s")`,
			escapePowerShell(message), escapePowerShell(title))
		cmd := exec.Command("powershell", "-Command", psScript)
		return cmd.Run()

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// escapeAppleScript escapes a string for use inside an AppleScript
// double-quoted literal.
func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// escapePowerShell escapes a string for use inside a PowerShell double-quoted
// literal, neutralizing quotes and variable/subexpression expansion.
func escapePowerShell(s string) string {
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, `"`, "`\"")
	s = strings.ReplaceAll(s, "$", "`$")
	return s
}

// sendWebhookNotification sends a notification via webhook. The JSON body is
// built with encoding/json so title/message content is correctly escaped.
func sendWebhookNotification(webhook, title, message string) error {
	body, err := json.Marshal(map[string]string{
		"title":   title,
		"message": message,
	})
	if err != nil {
		return fmt.Errorf("failed to encode webhook payload: %w", err)
	}
	cmd := exec.Command("curl", "-X", "POST", "-H", "Content-Type: application/json", "-d", string(body), webhook)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("webhook failed: %w (output: %s)", err, string(output))
	}

	return nil
}

func init() {
	notifySendCmd.Flags().StringP("title", "t", "LazyAI", "Notification title")
	notifyConfigCmd.Flags().String("webhook", "", "Webhook URL for notifications")
	notifyConfigCmd.Flags().Bool("enabled", true, "Enable notifications")

	notifyCmd.AddCommand(notifySendCmd)
	notifyCmd.AddCommand(notifyConfigCmd)
	notifyCmd.AddCommand(notifyTestCmd)
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.GroupID = "safety"
}
