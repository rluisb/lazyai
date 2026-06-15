// Package wizard provides plan types used by the setup wizard's confirmation
// and conflict-resolution phases. The actual file installation is performed by
// the scaffold pipeline; the plan is intentionally minimal and conflict-free.
package wizard

// InstallPlan describes what files would be installed.
// The wizard currently uses scaffold.ScaffoldAll to perform installation, so
// the plan is a placeholder used only for the review/confirm phase.
type InstallPlan struct {
	FilesToInstall []PlannedFile
	Conflicts      []ConflictInfo
}

// PlannedFile describes a single file to be installed.
type PlannedFile struct {
	Source    string // library path (relative to library dir)
	Target    string // destination path (absolute or relative to target dir)
	Type      string // category: agent, skill, prompt, template, rule, infra, root, constitution
	Content   []byte // new content from the library
	Existing  bool   // true if file already exists at target
	HashMatch bool   // true if existing file hash matches library hash (no conflict)
}

// ConflictInfo describes a file conflict between current and new content.
type ConflictInfo struct {
	Target          string // path to the conflicting file
	ExistingContent []byte // content of the existing file
	Content         []byte // new content from the library
	Type            string // file category
}

// ComputePlan computes the install plan. In the current architecture the
// scaffold pipeline handles actual file selection and installation, so this
// returns an empty, conflict-free plan used only for the confirmation phase.
func ComputePlan(config *WizardConfig) (*InstallPlan, error) {
	_ = config
	return &InstallPlan{}, nil
}
