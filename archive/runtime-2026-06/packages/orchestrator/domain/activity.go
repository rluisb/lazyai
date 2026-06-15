package domain

// ActiveRunCounts summarizes persisted daemon work that should block idle shutdown.
type ActiveRunCounts struct {
	Chains    int `json:"chains"`
	Teams     int `json:"teams"`
	Workflows int `json:"workflows"`
	QueueJobs int `json:"queueJobs"`
	Total     int `json:"total"`
}
