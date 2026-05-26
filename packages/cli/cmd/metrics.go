package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Metrics and analytics",
	Long:  `Export metrics and generate dashboards.`,
}

var metricsExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export metrics to Prometheus format",
	Long:  `Export session and dispatch metrics in Prometheus exposition format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "metrics.prom"
		}

		// Query metrics from database
		rows, err := database.Query(`
			SELECT metric_name, metric_value, recorded_at 
			FROM quality_metrics 
			ORDER BY recorded_at DESC
		`)
		if err != nil {
			return fmt.Errorf("error querying metrics: %w", err)
		}
		defer rows.Close()

		// Build Prometheus format
		var promData string
		promData += "# HELP lazyai_metric LazyAI quality metric\n"
		promData += "# TYPE lazyai_metric gauge\n"

		count := 0
		for rows.Next() {
			var name, recordedAt string
			var value float64
			if err := rows.Scan(&name, &value, &recordedAt); err != nil {
				continue
			}
			promData += fmt.Sprintf("lazyai_metric{name=\"%s\",recorded_at=\"%s\"} %.2f\n", name, recordedAt, value)
			count++
		}

		// Add session count
		var sessionCount int
		database.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)
		promData += fmt.Sprintf("\n# HELP lazyai_sessions_total Total number of sessions\n")
		promData += fmt.Sprintf("# TYPE lazyai_sessions_total counter\n")
		promData += fmt.Sprintf("lazyai_sessions_total %d\n", sessionCount)

		// Add task counts
		var pendingTasks, completedTasks int
		database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'pending'").Scan(&pendingTasks)
		database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&completedTasks)
		promData += fmt.Sprintf("\n# HELP lazyai_tasks_total Total number of tasks by status\n")
		promData += fmt.Sprintf("# TYPE lazyai_tasks_total counter\n")
		promData += fmt.Sprintf("lazyai_tasks_total{status=\"pending\"} %d\n", pendingTasks)
		promData += fmt.Sprintf("lazyai_tasks_total{status=\"completed\"} %d\n", completedTasks)

		// Write to file
		if err := os.WriteFile(output, []byte(promData), 0644); err != nil {
			return fmt.Errorf("error writing metrics file: %w", err)
		}

		fmt.Printf("✅ Exported %d metrics to %s\n", count, output)
		fmt.Printf("   Sessions: %d | Pending tasks: %d | Completed tasks: %d\n", sessionCount, pendingTasks, completedTasks)

		return nil
	},
}

var metricsDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Generate HTML dashboard",
	Long:  `Generate an HTML dashboard with metrics visualization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "dashboard.html"
		}

		// Gather metrics
		var sessionCount int
		database.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)

		var activeSessions int
		database.QueryRow("SELECT COUNT(*) FROM sessions WHERE status = 'active'").Scan(&activeSessions)

		var totalTasks int
		database.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&totalTasks)

		var completedTasks int
		database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&completedTasks)

		var ledgerEntries int
		ledgerFile := filepath.Join(".specify", "ledger.jsonl")
		if data, err := os.ReadFile(ledgerFile); err == nil {
			for _, b := range data {
				if b == '\n' {
					ledgerEntries++
				}
			}
		}

		// Generate HTML
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>LazyAI Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #333; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-top: 30px; }
        .metric-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 36px; font-weight: bold; color: #4CAF50; margin: 10px 0; }
        .metric-label { color: #666; font-size: 14px; text-transform: uppercase; }
        .timestamp { color: #999; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 LazyAI Metrics Dashboard</h1>
        <div class="metrics">
            <div class="metric-card">
                <div class="metric-label">Total Sessions</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Active Sessions</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Total Tasks</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Completed Tasks</div>
                <div class="metric-value">%d</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Ledger Entries</div>
                <div class="metric-value">%d</div>
            </div>
        </div>
        <div class="timestamp">Generated: %s</div>
    </div>
</body>
</html>`, sessionCount, activeSessions, totalTasks, completedTasks, ledgerEntries, time.Now().Format(time.RFC3339))

		if err := os.WriteFile(output, []byte(html), 0644); err != nil {
			return fmt.Errorf("error writing dashboard: %w", err)
		}

		fmt.Printf("✅ Dashboard generated: %s\n", output)
		fmt.Printf("   Sessions: %d | Tasks: %d/%d completed | Ledger: %d entries\n", sessionCount, completedTasks, totalTasks, ledgerEntries)

		return nil
	},
}

var metricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent metrics",
	Long:  `Show recent quality metrics from the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		limit := 10
		if l, _ := cmd.Flags().GetInt("limit"); l > 0 {
			limit = l
		}

		rows, err := database.Query(`
			SELECT metric_name, metric_value, agent, recorded_at 
			FROM quality_metrics 
			ORDER BY recorded_at DESC 
			LIMIT ?
		`, limit)
		if err != nil {
			return fmt.Errorf("error querying metrics: %w", err)
		}
		defer rows.Close()

		fmt.Println("Recent Metrics:")
		fmt.Println("───────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var name, agent, recordedAt string
			var value float64
			if err := rows.Scan(&name, &value, &agent, &recordedAt); err != nil {
				continue
			}
			count++
			fmt.Printf("  %s | %s=%.2f | %s\n", recordedAt, name, value, agent)
		}

		if count == 0 {
			fmt.Println("  No metrics found.")
		}

		return nil
	},
}

func init() {
	metricsExportCmd.Flags().StringP("output", "o", "metrics.prom", "Output file path")
	metricsDashboardCmd.Flags().StringP("output", "o", "dashboard.html", "Output file path")
	metricsListCmd.Flags().IntP("limit", "n", 10, "Number of metrics to show")

	metricsCmd.AddCommand(metricsExportCmd)
	metricsCmd.AddCommand(metricsDashboardCmd)
	metricsCmd.AddCommand(metricsListCmd)
	rootCmd.AddCommand(metricsCmd)
	metricsCmd.GroupID = "audit"
}
