package db

// ActiveRunCounts summarizes persisted daemon work that should block idle shutdown.
type ActiveRunCounts struct {
	Chains    int `json:"chains"`
	Teams     int `json:"teams"`
	Workflows int `json:"workflows"`
	QueueJobs int `json:"queueJobs"`
	Total     int `json:"total"`
}

// ActiveRunCounts counts persisted runs/jobs that are still active or in progress.
func (db *DB) ActiveRunCounts() (ActiveRunCounts, error) {
	counts := ActiveRunCounts{}
	var err error

	counts.Chains, err = db.countRows(`SELECT COUNT(*) FROM chain_runs WHERE state IN ('created', 'running', 'gated', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.Teams, err = db.countRows(`SELECT COUNT(*) FROM team_runs WHERE state IN ('created', 'running', 'synthesizing', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.Workflows, err = db.countRows(`SELECT COUNT(*) FROM workflow_runs WHERE state IN ('created', 'running', 'waiting_on_child', 'gated', 'awaiting_recovery', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.QueueJobs, err = db.countRows(`SELECT COUNT(*) FROM queue_jobs WHERE status IN ('pending', 'claimed')`)
	if err != nil {
		return counts, err
	}
	counts.Total = counts.Chains + counts.Teams + counts.Workflows + counts.QueueJobs
	return counts, nil
}

func (db *DB) countRows(query string) (int, error) {
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
