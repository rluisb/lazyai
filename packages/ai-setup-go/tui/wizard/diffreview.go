package wizard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
	reviewviewer "github.com/ricardoborges-teachable/ai-setup/packages/diffviewer"
)

const (
	diffReviewContractVersion = 1
	defaultDiffViewerBinary   = "diffviewer"
	diffReviewChangedLineGate = 20
)

// ErrDiffReviewCancelled reports that the delegated diff review was explicitly cancelled.
var ErrDiffReviewCancelled = errors.New("diff review cancelled")

// ReviewAction is the action returned by the diff review contract.
type ReviewAction = reviewviewer.Action

const (
	ReviewActionAccept = reviewviewer.ActionAccept
	ReviewActionDeny   = reviewviewer.ActionDeny
	ReviewActionSkip   = reviewviewer.ActionSkip
)

// DiffReviewClient reviews conflicts and returns per-file conflict resolutions.
type DiffReviewClient interface {
	RunReview(conflicts []conflict.Conflict) ([]ConflictResolution, error)
}

// DiffReviewCommandRunner shells out to the diffviewer binary.
type DiffReviewCommandRunner interface {
	RunDiffReview(ctx context.Context, binaryPath string, stdin []byte) ([]byte, error)
}

// BinaryDiffReviewer delegates large or multi-file reviews to the diffviewer binary.
type BinaryDiffReviewer struct {
	BinaryPath string
	Inline     DiffReviewClient
	Runner     DiffReviewCommandRunner
	Stderr     io.Writer
}

// InlineDiffReviewer reviews conflicts with simple per-file prompts.
type InlineDiffReviewer struct{}

type execDiffReviewCommandRunner struct {
	stderr io.Writer
}

// NewDiffReviewClient returns the default threshold-gated diff review client.
func NewDiffReviewClient() DiffReviewClient {
	return BinaryDiffReviewer{Inline: InlineDiffReviewer{}}
}

// ShouldDelegateReview reports whether a conflict set should use the diffviewer binary.
func ShouldDelegateReview(conflicts []conflict.Conflict) bool {
	if len(conflicts) > 1 {
		return true
	}

	for _, c := range conflicts {
		if changedLineCount(c) >= diffReviewChangedLineGate {
			return true
		}
	}

	return false
}

// ConflictStrategyForReviewAction maps review decisions onto scaffold conflict strategies.
func ConflictStrategyForReviewAction(action ReviewAction) types.ConflictStrategy {
	switch action {
	case ReviewActionAccept:
		return types.ConflictStrategyBackupAndReplace
	case ReviewActionDeny:
		return types.ConflictStrategySkip
	case ReviewActionSkip:
		return types.ConflictStrategyAlign
	default:
		return types.ConflictStrategySkip
	}
}

// RunReview delegates above-threshold conflicts to diffviewer and falls back to inline review
// when the binary cannot be invoked.
func (r BinaryDiffReviewer) RunReview(conflicts []conflict.Conflict) ([]ConflictResolution, error) {
	inline := r.Inline
	if inline == nil {
		inline = InlineDiffReviewer{}
	}

	if !ShouldDelegateReview(conflicts) {
		return inline.RunReview(conflicts)
	}

	resolutions, err := r.runDelegatedReview(conflicts)
	if err == nil {
		return resolutions, nil
	}
	if errors.Is(err, ErrDiffReviewCancelled) {
		return nil, err
	}

	return inline.RunReview(conflicts)
}

// RunReview prompts for each conflict using the existing simple wizard controls.
func (InlineDiffReviewer) RunReview(conflicts []conflict.Conflict) ([]ConflictResolution, error) {
	resolutions := make([]ConflictResolution, 0, len(conflicts))
	for _, c := range conflicts {
		actionValue := string(ReviewActionSkip)
		selectAction := huh.NewSelect[string]().
			Title(fmt.Sprintf("Review conflict: %s", c.Path)).
			Options(
				huh.NewOption("Accept — backup and replace with library version", string(ReviewActionAccept)),
				huh.NewOption("Deny — keep existing file and skip", string(ReviewActionDeny)),
				huh.NewOption("Skip — leave aligned for later handling", string(ReviewActionSkip)),
			).
			Value(&actionValue)

		if err := huh.NewForm(huh.NewGroup(selectAction).Title("Conflict Resolution")).Run(); err != nil {
			return nil, fmt.Errorf("inline diff review cancelled: %w", err)
		}

		resolutions = append(resolutions, ConflictResolution{
			Path:   c.Path,
			Action: ReviewAction(actionValue),
		})
	}

	return resolutions, nil
}

func (r BinaryDiffReviewer) runDelegatedReview(conflicts []conflict.Conflict) ([]ConflictResolution, error) {
	request := buildReviewRequest(conflicts)
	stdin, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode diff review request: %w", err)
	}

	runner := r.Runner
	if runner == nil {
		runner = execDiffReviewCommandRunner{stderr: r.Stderr}
	}

	stdout, err := runner.RunDiffReview(context.Background(), r.binaryPath(), append(stdin, '\n'))
	if err != nil {
		return nil, fmt.Errorf("run diffviewer: %w", err)
	}

	var response reviewviewer.ReviewResponse
	if err := json.Unmarshal(bytes.TrimSpace(stdout), &response); err != nil {
		return nil, fmt.Errorf("decode diffviewer response: %w", err)
	}
	if err := reviewviewer.ValidateReviewResponse(response); err != nil {
		return nil, fmt.Errorf("invalid diffviewer response: %w", err)
	}
	if response.Status == reviewviewer.ReviewStatusCancelled {
		return nil, ErrDiffReviewCancelled
	}

	return conflictResolutionsFromReview(response.Resolutions), nil
}

func (r BinaryDiffReviewer) binaryPath() string {
	if r.BinaryPath != "" {
		return r.BinaryPath
	}
	return defaultDiffViewerBinary
}

func (r execDiffReviewCommandRunner) RunDiffReview(ctx context.Context, binaryPath string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Stdin = bytes.NewReader(stdin)
	stderr := r.stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	cmd.Stderr = stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), err
	}

	return stdout.Bytes(), nil
}

func buildReviewRequest(conflicts []conflict.Conflict) reviewviewer.ReviewRequest {
	title := "Review conflicting files"
	files := make([]reviewviewer.FileDiff, 0, len(conflicts))
	for _, c := range conflicts {
		files = append(files, reviewviewer.FileDiff{
			Path:           c.Path,
			CurrentContent: string(c.CurrentContent),
			NewContent:     string(c.NewContent),
		})
	}

	return reviewviewer.ReviewRequest{
		Version: diffReviewContractVersion,
		Title:   &title,
		Files:   files,
	}
}

func conflictResolutionsFromReview(resolutions []reviewviewer.Resolution) []ConflictResolution {
	mapped := make([]ConflictResolution, 0, len(resolutions))
	for _, resolution := range resolutions {
		mapped = append(mapped, ConflictResolution{
			Path:   resolution.Path,
			Action: resolution.Action,
		})
	}
	return mapped
}

func changedLineCount(c conflict.Conflict) int {
	diff := reviewviewer.ComputeDiffResult(c.CurrentContent, c.NewContent)
	return diff.Stats.Additions + diff.Stats.Deletions
}
