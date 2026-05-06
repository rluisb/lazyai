package diffviewer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	charmlog "charm.land/log/v2"
	ailog "github.com/rluisb/lazyai/packages/diffviewer/internal/log"
)

// Reviewer runs a prepared review and returns the user's final response.
type Reviewer interface {
	Run() (ReviewResponse, error)
}

// ReviewerFactory builds a reviewer for request-derived conflict views.
type ReviewerFactory func([]ConflictView) Reviewer

type reviewerFunc func() (ReviewResponse, error)

func (f reviewerFunc) Run() (ReviewResponse, error) {
	return f()
}

// RunCLI executes the diffviewer JSON stdin/stdout protocol.
func RunCLI(stdin io.Reader, stdout, stderr io.Writer, newReviewer ReviewerFactory) int {
	req, err := decodeReviewRequest(stdin)
	if err != nil {
		return writeErrorResponse(stdout, stderr, err)
	}

	if err := ValidateReviewRequest(req); err != nil {
		return writeErrorResponse(stdout, stderr, err)
	}

	viewer := newReviewer(reviewRequestConflictViews(req))
	if viewer == nil {
		return writeErrorResponse(stdout, stderr, fmt.Errorf("reviewer factory returned nil reviewer"))
	}

	resp, err := viewer.Run()
	if err != nil {
		return writeErrorResponse(stdout, stderr, err)
	}
	if err := ValidateReviewResponse(resp); err != nil {
		return writeErrorResponse(stdout, stderr, err)
	}

	if err := encodeReviewResponse(stdout, resp); err != nil {
		cliLogger(stderr).Error("write response failed", "error", err)
		return 1
	}

	return 0
}

func decodeReviewRequest(stdin io.Reader) (ReviewRequest, error) {
	var req ReviewRequest
	decoder := json.NewDecoder(stdin)
	if err := decoder.Decode(&req); err != nil {
		return ReviewRequest{}, fmt.Errorf("decode review request: %w", err)
	}
	return req, nil
}

func reviewRequestConflictViews(req ReviewRequest) []ConflictView {
	views := make([]ConflictView, 0, len(req.Files))
	for _, file := range req.Files {
		views = append(views, ConflictView{
			FilePath:     file.Path,
			CurrentLines: strings.Split(file.CurrentContent, "\n"),
			NewLines:     strings.Split(file.NewContent, "\n"),
		})
	}
	return views
}

func writeErrorResponse(stdout, stderr io.Writer, err error) int {
	logger := cliLogger(stderr)
	logger.Error("review failed", "error", err)
	message := err.Error()
	resp := ReviewResponse{
		Version:     reviewContractVersion,
		Status:      ReviewStatusCancelled,
		Resolutions: []Resolution{},
		Message:     &message,
	}
	if encodeErr := encodeReviewResponse(stdout, resp); encodeErr != nil {
		logger.Error("write error response failed", "error", encodeErr)
	}
	return 1
}

func cliLogger(stderr io.Writer) *charmlog.Logger {
	logger := ailog.Default().With("component", "cli")
	logger.SetOutput(stderr)
	return logger
}

func encodeReviewResponse(stdout io.Writer, resp ReviewResponse) error {
	encoder := json.NewEncoder(stdout)
	return encoder.Encode(resp)
}
