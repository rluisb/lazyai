package diffviewer

import "github.com/rluisb/lazyai/packages/diffviewer/domain"

const reviewContractVersion = domain.ReviewContractVersion

// Action is a user-selected per-file review decision.
type Action = domain.Action

const (
	ActionAccept = domain.ActionAccept
	ActionDeny   = domain.ActionDeny
	ActionSkip   = domain.ActionSkip
)

// ReviewStatus is the terminal status of an interactive review.
type ReviewStatus = domain.ReviewStatus

const (
	ReviewStatusConfirmed = domain.ReviewStatusConfirmed
	ReviewStatusCancelled = domain.ReviewStatusCancelled
)

// ReviewRequest is the JSON v1 request payload supplied to diffviewer.
type ReviewRequest = domain.ReviewRequest

// FileDiff is a single file diff input in a review request.
type FileDiff = domain.FileDiff

// ReviewResponse is the JSON v1 response payload emitted by diffviewer.
type ReviewResponse = domain.ReviewResponse

// Resolution records the user's decision for a single reviewed file.
type Resolution = domain.Resolution

// ValidateReviewRequest verifies that a ReviewRequest satisfies the JSON v1 contract.
func ValidateReviewRequest(req ReviewRequest) error {
	return domain.ValidateReviewRequest(req)
}

// ValidateReviewResponse verifies that a ReviewResponse satisfies the JSON v1 contract.
func ValidateReviewResponse(resp ReviewResponse) error {
	return domain.ValidateReviewResponse(resp)
}
