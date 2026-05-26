package cmd

import (
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
)

// recordQualityMetric records a quality metric to the database
func recordQualityMetric(sessionID, metricName, agent, model string) error {
	database, err := getDB()
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		return err
	}

	_, err = database.Exec(
		"INSERT INTO quality_metrics (session_id, agent, model, metric_name, metric_value, recorded_at) VALUES (?, ?, ?, ?, ?, ?)",
		sessionID, agent, model, metricName, 1.0, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to record metric: %w", err)
	}

	return nil
}
