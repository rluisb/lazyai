package domain

import "fmt"

const ReviewContractVersion = 1

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
	if req.Version != ReviewContractVersion {
		return fmt.Errorf("review request version must be %d, got %d", ReviewContractVersion, req.Version)
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
	if resp.Version != ReviewContractVersion {
		return fmt.Errorf("review response version must be %d, got %d", ReviewContractVersion, resp.Version)
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

// RecordResolution records or overwrites a per-file resolution in review order.
func RecordResolution(decisions map[int]Resolution, totalFiles, index int, path string, action Action) map[int]Resolution {
	if index < 0 || index >= totalFiles {
		return decisions
	}
	if decisions == nil {
		decisions = make(map[int]Resolution, totalFiles)
	}
	decisions[index] = Resolution{Path: path, Action: action}
	return decisions
}

// OrderedResolutions returns recorded decisions in file order for stable output.
func OrderedResolutions(paths []string, decisions map[int]Resolution) []Resolution {
	resolutions := make([]Resolution, 0, len(decisions))
	for i, path := range paths {
		resolution, ok := decisions[i]
		if ok {
			if resolution.Path == "" {
				resolution.Path = path
			}
			resolutions = append(resolutions, resolution)
		}
	}
	return resolutions
}

// AllFilesDecided reports whether every review file has a recorded decision.
func AllFilesDecided(totalFiles int, decisions map[int]Resolution) bool {
	return totalFiles > 0 && len(decisions) == totalFiles
}

// ReviewResponseForStatus builds a review response with confirmed-only resolutions.
func ReviewResponseForStatus(status ReviewStatus, resolutions []Resolution) ReviewResponse {
	if status != ReviewStatusConfirmed {
		resolutions = []Resolution{}
	}
	return ReviewResponse{
		Version:     ReviewContractVersion,
		Status:      status,
		Resolutions: resolutions,
	}
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
