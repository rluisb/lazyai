package diffviewer

import "fmt"

const reviewContractVersion = 1

// Action is a user-selected per-file review decision.
type Action string

const (
	ActionAccept Action = "accept"
	ActionDeny   Action = "deny"
	ActionSkip   Action = "skip"
)

// ReviewStatus is the terminal status of an interactive review.
type ReviewStatus string

const (
	ReviewStatusConfirmed ReviewStatus = "confirmed"
	ReviewStatusCancelled ReviewStatus = "cancelled"
)

// ReviewRequest is the JSON v1 request payload supplied to diffviewer.
type ReviewRequest struct {
	Version int        `json:"version"`
	Title   *string    `json:"title,omitempty"`
	Files   []FileDiff `json:"files"`
}

// FileDiff is a single file diff input in a review request.
type FileDiff struct {
	Path           string `json:"path"`
	CurrentContent string `json:"currentContent"`
	NewContent     string `json:"newContent"`
}

// ReviewResponse is the JSON v1 response payload emitted by diffviewer.
type ReviewResponse struct {
	Version     int          `json:"version"`
	Status      ReviewStatus `json:"status"`
	Resolutions []Resolution `json:"resolutions"`
	Message     *string      `json:"message,omitempty"`
}

// Resolution records the user's decision for a single reviewed file.
type Resolution struct {
	Path   string `json:"path"`
	Action Action `json:"action"`
}

// ValidateReviewRequest verifies that a ReviewRequest satisfies the JSON v1 contract.
func ValidateReviewRequest(req ReviewRequest) error {
	if req.Version != reviewContractVersion {
		return fmt.Errorf("review request version must be %d, got %d", reviewContractVersion, req.Version)
	}
	if req.Title != nil && *req.Title == "" {
		return fmt.Errorf("review request title must not be empty when provided")
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("review request files must contain at least one file")
	}

	for i, file := range req.Files {
		if file.Path == "" {
			return fmt.Errorf("review request files[%d].path must not be empty", i)
		}
	}

	return nil
}

// ValidateReviewResponse verifies that a ReviewResponse satisfies the JSON v1 contract.
func ValidateReviewResponse(resp ReviewResponse) error {
	if resp.Version != reviewContractVersion {
		return fmt.Errorf("review response version must be %d, got %d", reviewContractVersion, resp.Version)
	}
	if !resp.Status.valid() {
		return fmt.Errorf("review response status must be %q or %q, got %q", ReviewStatusConfirmed, ReviewStatusCancelled, resp.Status)
	}
	if resp.Resolutions == nil {
		return fmt.Errorf("review response resolutions must be provided")
	}
	if resp.Message != nil && *resp.Message == "" {
		return fmt.Errorf("review response message must not be empty when provided")
	}

	for i, resolution := range resp.Resolutions {
		if resolution.Path == "" {
			return fmt.Errorf("review response resolutions[%d].path must not be empty", i)
		}
		if !resolution.Action.valid() {
			return fmt.Errorf("review response resolutions[%d].action must be %q, %q, or %q, got %q", i, ActionAccept, ActionDeny, ActionSkip, resolution.Action)
		}
	}

	return nil
}

func (a Action) valid() bool {
	switch a {
	case ActionAccept, ActionDeny, ActionSkip:
		return true
	default:
		return false
	}
}

func (s ReviewStatus) valid() bool {
	switch s {
	case ReviewStatusConfirmed, ReviewStatusCancelled:
		return true
	default:
		return false
	}
}
