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
	"strings"

	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	reviewviewer "github.com/rluisb/lazyai/packages/cli/internal/diffreview"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

const (
	diffReviewContractVersion = 1
	defaultDiffViewerBinary   = "lazyai-diffviewer"
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

// DiffReviewBinaryResolver resolves the diffviewer binary before delegation.
type DiffReviewBinaryResolver func(binaryName string) (string, error)

// DiffReviewTerminalDetector reports whether delegated terminal UI can run.
type DiffReviewTerminalDetector func() bool

// InlineDiffReviewPrompter prompts for one inline per-file review decision.
type InlineDiffReviewPrompter interface {
	PromptReview(c conflict.Conflict, index, total int) (ReviewAction, error)
}

// BinaryDiffReviewer delegates large or multi-file reviews to the diffviewer binary.
type BinaryDiffReviewer struct {
	BinaryPath     string
	Inline         DiffReviewClient
	Runner         DiffReviewCommandRunner
	Stderr         io.Writer
	BinaryResolver DiffReviewBinaryResolver
	IsTerminal     DiffReviewTerminalDetector
}

// InlineDiffReviewer reviews conflicts with simple per-file prompts.
type InlineDiffReviewer struct {
	Prompter InlineDiffReviewPrompter
}

type huhInlineDiffReviewPrompter struct{}

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

	if !r.canRunDelegatedReview() {
		return inline.RunReview(conflicts)
	}

	binaryPath, err := r.resolveBinaryPath()
	if err != nil {
		return inline.RunReview(conflicts)
	}

	resolutions, err := r.runDelegatedReview(conflicts, binaryPath)
	if err == nil {
		return resolutions, nil
	}
	if errors.Is(err, ErrDiffReviewCancelled) {
		return nil, err
	}

	return inline.RunReview(conflicts)
}

// RunReview prompts for each conflict using the existing simple wizard controls.
func (r InlineDiffReviewer) RunReview(conflicts []conflict.Conflict) ([]ConflictResolution, error) {
	prompter := r.Prompter
	if prompter == nil {
		prompter = huhInlineDiffReviewPrompter{}
	}

	resolutions := make([]ConflictResolution, 0, len(conflicts))
	for i, c := range conflicts {
		action, err := prompter.PromptReview(c, i+1, len(conflicts))
		if err != nil {
			return nil, fmt.Errorf("inline diff review cancelled: %w", err)
		}
		if !isValidReviewAction(action) {
			return nil, fmt.Errorf("inline diff review returned invalid action %q for %s", action, c.Path)
		}

		resolutions = append(resolutions, ConflictResolution{
			Path:   c.Path,
			Action: action,
		})
	}

	return resolutions, nil
}

func (huhInlineDiffReviewPrompter) PromptReview(c conflict.Conflict, index, total int) (ReviewAction, error) {
	actionValue := string(ReviewActionSkip)
	selectAction := huh.NewSelect[string]().
		Title(fmt.Sprintf("Review conflict %d/%d: %s", index, total, c.Path)).
		Description(inlineDiffReviewDescription(c)).
		Options(
			huh.NewOption("Accept — backup and replace with library version", string(ReviewActionAccept)),
			huh.NewOption("Deny — keep existing file and skip", string(ReviewActionDeny)),
			huh.NewOption("Skip — leave aligned for later handling", string(ReviewActionSkip)),
		).
		Value(&actionValue)

	if err := theme.NewForm(huh.NewGroup(selectAction).Title("Conflict Resolution")).Run(); err != nil {
		return "", err
	}

	return ReviewAction(actionValue), nil
}

func (r BinaryDiffReviewer) runDelegatedReview(conflicts []conflict.Conflict, binaryPath string) ([]ConflictResolution, error) {
	request := buildReviewRequest(conflicts)
	stdin, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode diff review request: %w", err)
	}

	runner := r.Runner
	if runner == nil {
		runner = execDiffReviewCommandRunner{stderr: r.Stderr}
	}

	stdout, err := runner.RunDiffReview(context.Background(), binaryPath, append(stdin, '\n'))
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

func (r BinaryDiffReviewer) canRunDelegatedReview() bool {
	isTerminal := r.IsTerminal
	if isTerminal == nil {
		isTerminal = diffReviewIsTerminal
	}
	return isTerminal()
}

func (r BinaryDiffReviewer) resolveBinaryPath() (string, error) {
	resolver := r.BinaryResolver
	if resolver == nil {
		resolver = exec.LookPath
	}
	return resolver(r.binaryPath())
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

func inlineDiffReviewDescription(c conflict.Conflict) string {
	diff := reviewviewer.ComputeDiffResult(c.CurrentContent, c.NewContent)
	preview := reviewviewer.RenderSimpleDiff(diff.Lines)
	preview = strings.TrimRight(preview, "\n")
	if preview == "" {
		preview = "No line-level changes detected."
	}

	return fmt.Sprintf("Library update changes: +%d -%d\n\n%s", diff.Stats.Additions, diff.Stats.Deletions, truncateInlineDiffPreview(preview))
}

func truncateInlineDiffPreview(preview string) string {
	const maxLines = 30
	lines := strings.Split(preview, "\n")
	if len(lines) <= maxLines {
		return preview
	}

	omitted := len(lines) - maxLines
	return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n… %d more line(s); choose Skip to leave this conflict for later.", omitted)
}

func isValidReviewAction(action ReviewAction) bool {
	switch action {
	case ReviewActionAccept, ReviewActionDeny, ReviewActionSkip:
		return true
	default:
		return false
	}
}

func changedLineCount(c conflict.Conflict) int {
	diff := reviewviewer.ComputeDiffResult(c.CurrentContent, c.NewContent)
	return diff.Stats.Additions + diff.Stats.Deletions
}

func diffReviewIsTerminal() bool {
	return fileIsTerminal(os.Stdin) && fileIsTerminal(os.Stdout)
}

func fileIsTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	fi, err := file.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
