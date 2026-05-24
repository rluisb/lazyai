package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "Agent message bus",
	Long:  `Send and receive messages between agents.`,
}

var messageSendCmd = &cobra.Command{
	Use:   "send [to-agent] [subject] [body]",
	Short: "Send a message to an agent",
	Long:  `Send a message to a specific agent.`,
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		toAgent := args[0]
		subject := args[1]
		body := args[2]

		if err := ValidateAgentName(toAgent); err != nil {
			return err
		}
		if err := ValidateNotEmpty(subject, "subject"); err != nil {
			return err
		}
		if err := ValidateNotEmpty(body, "body"); err != nil {
			return err
		}

		// Get from agent from env or default
		fromAgent := os.Getenv("LAZYAI_AGENT")
		if fromAgent == "" {
			fromAgent = "orchestrator"
		}

		priority, _ := cmd.Flags().GetString("priority")
		if priority == "" {
			priority = "normal"
		}
		if err := ValidatePriority(priority); err != nil {
			return err
		}

		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		messageID := fmt.Sprintf("msg_%d", time.Now().Unix())

		_, err = database.Exec(
			"INSERT INTO messages (message_id, from_agent, to_agent, subject, body, priority, status, created_at) VALUES (?, ?, ?, ?, ?, ?, 'unread', ?)",
			messageID, fromAgent, toAgent, subject, body, priority, time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}

		fmt.Printf("✅ Message sent: %s\n", messageID)
		fmt.Printf("   From: %s → To: %s\n", fromAgent, toAgent)
		fmt.Printf("   Subject: %s | Priority: %s\n", subject, priority)

		return nil
	},
}

var messageRecvCmd = &cobra.Command{
	Use:   "recv [agent]",
	Short: "Receive messages for an agent",
	Long:  `Receive unread messages for a specific agent.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agent := args[0]

		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		// Mark messages as read
		_, err = database.Exec(
			"UPDATE messages SET status = 'read', read_at = ? WHERE to_agent = ? AND status = 'unread'",
			time.Now().UTC().Format(time.RFC3339), agent,
		)
		if err != nil {
			return fmt.Errorf("error marking messages as read: %w", err)
		}

		// Get messages
		rows, err := database.Query(
			"SELECT message_id, from_agent, subject, body, priority, created_at FROM messages WHERE to_agent = ? ORDER BY created_at DESC LIMIT 10",
			agent,
		)
		if err != nil {
			return fmt.Errorf("error querying messages: %w", err)
		}
		defer rows.Close()

		fmt.Printf("Messages for %s:\n", agent)
		fmt.Println("───────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var messageID, fromAgent, subject, body, priority, createdAt string
			if err := rows.Scan(&messageID, &fromAgent, &subject, &body, &priority, &createdAt); err != nil {
				continue
			}
			count++
			fmt.Printf("  %s [%s] %s → %s\n", createdAt, priority, fromAgent, agent)
			fmt.Printf("    Subject: %s\n", subject)
			fmt.Printf("    Body: %s\n", body)
			fmt.Println()
		}

		if count == 0 {
			fmt.Println("  No messages found.")
		} else {
			fmt.Printf("  Showing %d recent messages\n", count)
		}

		return nil
	},
}

var messageBroadcastCmd = &cobra.Command{
	Use:   "broadcast [subject] [body]",
	Short: "Broadcast a message to all agents",
	Long:  `Send a message to all agents.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		subject := args[0]
		body := args[1]

		fromAgent := os.Getenv("LAZYAI_AGENT")
		if fromAgent == "" {
			fromAgent = "orchestrator"
		}

		priority, _ := cmd.Flags().GetString("priority")
		if priority == "" {
			priority = "normal"
		}
		if err := ValidatePriority(priority); err != nil {
			return err
		}

		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		// Always include common agents
		agents := []string{"orchestrator", "builder", "planner", "reviewer", "scout"}

		// Also get agents from tasks table
		rows, err := database.Query("SELECT DISTINCT agent FROM tasks WHERE agent IS NOT NULL")
		if err == nil {
			for rows.Next() {
				var agent string
				if err := rows.Scan(&agent); err == nil {
					// Check if already in list
					found := false
					for _, a := range agents {
						if a == agent {
							found = true
							break
						}
					}
					if !found {
						agents = append(agents, agent)
					}
				}
			}
			rows.Close()
		}

		// Send to each agent
		sent := 0
		for _, toAgent := range agents {
			messageID := fmt.Sprintf("msg_%d_%s", time.Now().Unix(), toAgent)
			_, err := database.Exec(
				"INSERT INTO messages (message_id, from_agent, to_agent, subject, body, priority, status, created_at) VALUES (?, ?, ?, ?, ?, ?, 'unread', ?)",
				messageID, fromAgent, toAgent, subject, body, priority, time.Now().UTC().Format(time.RFC3339),
			)
			if err == nil {
				sent++
			}
		}

		fmt.Printf("✅ Broadcast sent to %d agents\n", sent)
		fmt.Printf("   Subject: %s | Priority: %s\n", subject, priority)

		return nil
	},
}

func init() {
	messageSendCmd.Flags().StringP("priority", "p", "normal", "Message priority (low, normal, high, critical)")
	messageBroadcastCmd.Flags().StringP("priority", "p", "normal", "Message priority (low, normal, high, critical)")

	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageRecvCmd)
	messageCmd.AddCommand(messageBroadcastCmd)
	rootCmd.AddCommand(messageCmd)
	messageCmd.GroupID = "runtime"
}
