package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Cost tracking and analytics",
	Long:  `Track and analyze AI API costs.`,
}

var costShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show cost breakdown",
	Long:  `Show cost breakdown by session and agent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		period, _ := cmd.Flags().GetString("period")

		fmt.Println("Cost Breakdown")
		fmt.Println("═══════════════════════════════════════════════════════════════")

		// Get sessions with costs
		var query string
		var args2 []interface{}

		if period != "all" {
			// Filter by time period
			var duration time.Duration
			switch period {
			case "day":
				duration = 24 * time.Hour
			case "week":
				duration = 7 * 24 * time.Hour
			case "month":
				duration = 30 * 24 * time.Hour
			default:
				duration = 24 * time.Hour
			}

			cutoff := time.Now().Add(-duration).Format(time.RFC3339)
			query = "SELECT id, goal, cost_usd, started_at FROM sessions WHERE started_at > ? ORDER BY started_at DESC"
			args2 = append(args2, cutoff)
		} else {
			query = "SELECT id, goal, cost_usd, started_at FROM sessions ORDER BY started_at DESC"
		}

		rows, err := database.Query(query, args2...)
		if err != nil {
			return fmt.Errorf("error querying sessions: %w", err)
		}
		defer rows.Close()

		totalCost := 0.0
		sessionCount := 0

		fmt.Printf("\nSessions (%s):\n", period)
		fmt.Println("───────────────────────────────────────────────────────────────")

		for rows.Next() {
			var id, goal, startedAt string
			var costUSD float64

			if err := rows.Scan(&id, &goal, &costUSD, &startedAt); err != nil {
				continue
			}

			sessionCount++
			totalCost += costUSD

			if costUSD > 0 {
				fmt.Printf("  %s | $%.4f | %s\n", id, costUSD, goal)
			}
		}

		if sessionCount == 0 {
			fmt.Println("  No sessions found.")
		}

		fmt.Printf("\nTotal Cost: $%.4f\n", totalCost)
		fmt.Printf("Sessions: %d\n", sessionCount)

		if sessionCount > 0 {
			avgCost := totalCost / float64(sessionCount)
			fmt.Printf("Average per session: $%.4f\n", avgCost)
		}

		return nil
	},
}

var costAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Show dispatch activity by agent",
	Long:  `Show dispatch counts by agent. Cost is tracked at the session level (see "cost show"); it is not attributable per agent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		fmt.Println("Dispatches by Agent")
		fmt.Println("═══════════════════════════════════════════════════════════════")

		// Aggregate dispatch activity per agent. Note: cost_usd is a session-level
		// column (see the sessions table), not per-dispatch, so it is intentionally
		// not reported here to avoid misattributing a session's cost across the
		// multiple agents dispatched within it. Use "cost show" for session costs.
		rows, err := database.Query(`
			SELECT agent,
			       COUNT(*) AS count,
			       SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed,
			       SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed
			FROM agent_dispatches
			GROUP BY agent
			ORDER BY count DESC
		`)
		if err != nil {
			return fmt.Errorf("error querying dispatches: %w", err)
		}
		defer rows.Close()

		fmt.Println("\nAgent Usage:")
		fmt.Println("───────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var agent string
			var dispatchCount, completed, failed int

			if err := rows.Scan(&agent, &dispatchCount, &completed, &failed); err != nil {
				continue
			}

			count++
			fmt.Printf("  %s: %d dispatches (%d completed, %d failed)\n", agent, dispatchCount, completed, failed)
		}

		if count == 0 {
			fmt.Println("  No dispatch data found.")
		}

		fmt.Println("\nNote: per-agent cost is unavailable; cost is tracked per session. Run \"cost show\".")

		return nil
	},
}

var costBudgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Set and check budget",
	Long:  `Configure budget limits and check status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		budget, _ := cmd.Flags().GetFloat64("limit")

		database, err := EnsureDB()
		if err != nil {
			return err
		}
		defer SafeCloseDB(database)

		// Get total cost
		var totalCost float64
		database.QueryRow("SELECT COALESCE(SUM(cost_usd), 0) FROM sessions").Scan(&totalCost)

		fmt.Println("Budget Status")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Printf("Current Spend: $%.4f\n", totalCost)

		if budget > 0 {
			fmt.Printf("Budget Limit: $%.4f\n", budget)

			remaining := budget - totalCost
			percentage := (totalCost / budget) * 100

			fmt.Printf("Remaining: $%.4f\n", remaining)
			fmt.Printf("Used: %.1f%%\n", percentage)

			if percentage >= 100 {
				fmt.Println("\n⚠️  BUDGET EXCEEDED!")
			} else if percentage >= 80 {
				fmt.Println("\n⚠️  Budget warning: 80% used")
			}
		}

		return nil
	},
}

func init() {
	costShowCmd.Flags().StringP("period", "p", "all", "Time period (day, week, month, all)")
	costBudgetCmd.Flags().Float64P("limit", "l", 0, "Budget limit in USD")

	costCmd.AddCommand(costShowCmd)
	costCmd.AddCommand(costAgentCmd)
	costCmd.AddCommand(costBudgetCmd)
	rootCmd.AddCommand(costCmd)
	costCmd.GroupID = "audit"
}
