// Package diffviewer provides an interactive side-by-side diff viewer built on
// the Charm Bracelet Bubble Tea framework.
package diffviewer

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ConflictView holds the data for visualizing a single conflict.
type ConflictView struct {
	FilePath     string
	CurrentLines []string // "ours" / existing file lines.
	NewLines     []string // "theirs" / library file lines.
	DiffLines    []DiffLine
}

// DiffViewer is a Bubble Tea model for side-by-side conflict resolution.
type DiffViewer struct {
	conflicts    []ConflictView
	decisions    map[int]Resolution
	currentIndex int
	hunkIndex    int
	hunkStarts   []int
	leftVP       viewport.Model
	rightVP      viewport.Model
	width        int
	height       int
	showHelp     bool
	quitting     bool
}

// diffKeyMap defines the keybindings for the diff viewer.
type diffKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Accept   key.Binding
	Deny     key.Binding
	Skip     key.Binding
	Next     key.Binding
	Prev     key.Binding
	NextHunk key.Binding
	PrevHunk key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var diffKeys = diffKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "scroll up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "scroll down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left pane"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right pane"),
	),
	Accept: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "accept (library)"),
	),
	Deny: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deny (keep current)"),
	),
	Skip: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "skip (unresolved)"),
	),
	Next: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next conflict"),
	),
	Prev: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "previous conflict"),
	),
	NextHunk: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next hunk"),
	),
	PrevHunk: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "previous hunk"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// NewDiffViewer creates a new side-by-side diff viewer for the given conflicts.
func NewDiffViewer(conflicts []ConflictView) *DiffViewer {
	// Compute diff lines for each conflict that doesn't already have them.
	for i := range conflicts {
		if len(conflicts[i].DiffLines) == 0 {
			currentContent := strings.Join(conflicts[i].CurrentLines, "\n")
			newContent := strings.Join(conflicts[i].NewLines, "\n")
			conflicts[i].DiffLines = ComputeDiff([]byte(currentContent), []byte(newContent))
		}
	}

	viewer := &DiffViewer{
		conflicts:    conflicts,
		decisions:    make(map[int]Resolution, len(conflicts)),
		currentIndex: 0,
		hunkIndex:    0,
		showHelp:     false,
		quitting:     false,
	}
	viewer.updateHunkStarts()
	return viewer
}

// Init starts the diff viewer.
func (d *DiffViewer) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages.
func (d *DiffViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.syncViewports()
		return d, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, diffKeys.Quit):
			d.quitting = true
			return d, tea.Quit

		case key.Matches(msg, diffKeys.Help):
			d.showHelp = !d.showHelp
			return d, nil

		case key.Matches(msg, diffKeys.Accept):
			d.resolveCurrent(ActionAccept)
			return d, d.advanceOrQuit()

		case key.Matches(msg, diffKeys.Deny):
			d.resolveCurrent(ActionDeny)
			return d, d.advanceOrQuit()

		case key.Matches(msg, diffKeys.Skip):
			d.resolveCurrent(ActionSkip)
			return d, d.advanceOrQuit()

		case key.Matches(msg, diffKeys.Next):
			if d.currentIndex < len(d.conflicts)-1 {
				d.currentIndex++
				d.hunkIndex = 0
				d.syncViewports()
			}
			return d, nil

		case key.Matches(msg, diffKeys.Prev):
			if d.currentIndex > 0 {
				d.currentIndex--
				d.hunkIndex = 0
				d.syncViewports()
			}
			return d, nil

		case key.Matches(msg, diffKeys.NextHunk):
			d.nextHunk()
			return d, nil

		case key.Matches(msg, diffKeys.PrevHunk):
			d.prevHunk()
			return d, nil

		case key.Matches(msg, diffKeys.Up):
			d.leftVP.ScrollUp(1)
			d.rightVP.ScrollUp(1)
			return d, nil

		case key.Matches(msg, diffKeys.Down):
			d.leftVP.ScrollDown(1)
			d.rightVP.ScrollDown(1)
			return d, nil
		}
	}

	// Also handle viewport update messages.
	var cmd tea.Cmd
	d.leftVP, cmd = d.leftVP.Update(msg)
	if cmd != nil {
		return d, cmd
	}
	d.rightVP, cmd = d.rightVP.Update(msg)
	return d, cmd
}

// View renders the diff viewer.
func (d *DiffViewer) View() tea.View {
	if d.quitting {
		view := tea.NewView("Goodbye!\n")
		view.AltScreen = true
		return view
	}

	if len(d.conflicts) == 0 {
		return tea.NewView(SuccessLabel("No conflicts to resolve.") + "\n")
	}

	if d.width == 0 {
		d.width = 80
	}
	if d.height == 0 {
		d.height = 24
	}

	conflict := d.conflicts[d.currentIndex]
	halfWidth := d.width/2 - 3
	if halfWidth < 10 {
		halfWidth = 10
	}
	if d.leftVP.Width() == 0 || d.rightVP.Width() == 0 {
		d.syncViewports()
	}

	// Left pane: Current (yours).
	leftTitle := ErrorLabel("Current (yours)")
	leftContent := d.leftVP.View()

	// Right pane: New (library).
	rightTitle := SuccessLabel("New (library)")
	rightContent := d.rightVP.View()

	// Header.
	header := d.renderHeader(conflict)

	// Action bar.
	actionBar := d.renderActionBar()

	// Summary bar.
	summaryBar := d.renderSummary()

	// Assemble panes.
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Error).
		Width(halfWidth).
		Render(leftTitle + "\n" + leftContent)

	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Success).
		Width(halfWidth).
		Render(rightTitle + "\n" + rightContent)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(panes)
	sb.WriteString("\n")
	sb.WriteString(actionBar)
	sb.WriteString("\n")
	sb.WriteString(summaryBar)

	if d.showHelp {
		sb.WriteString("\n\n")
		sb.WriteString(d.renderHelp())
	}

	view := tea.NewView(sb.String())
	view.AltScreen = true
	return view
}

// Run starts the interactive diff viewer and returns resolutions.
func (d *DiffViewer) Run() ([]Resolution, error) {
	// Set initial dimensions.
	d.width = 80
	d.height = 24
	d.syncViewports()

	p := tea.NewProgram(d)
	_, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("diff viewer: %w", err)
	}

	if d.quitting {
		return nil, fmt.Errorf("user cancelled")
	}

	return d.orderedResolutions(), nil
}

// resolveCurrent records the resolution for the current conflict.
func (d *DiffViewer) resolveCurrent(action Action) {
	if d.currentIndex >= 0 && d.currentIndex < len(d.conflicts) {
		if d.decisions == nil {
			d.decisions = make(map[int]Resolution, len(d.conflicts))
		}
		d.decisions[d.currentIndex] = Resolution{
			Path:   d.conflicts[d.currentIndex].FilePath,
			Action: action,
		}
	}
}

// orderedResolutions returns recorded decisions in conflict order for stable output.
func (d *DiffViewer) orderedResolutions() []Resolution {
	resolutions := make([]Resolution, 0, len(d.decisions))
	for i := range d.conflicts {
		resolution, ok := d.decisions[i]
		if ok {
			resolutions = append(resolutions, resolution)
		}
	}
	return resolutions
}

// advanceOrQuit moves to the next conflict or quits if all are resolved.
func (d *DiffViewer) advanceOrQuit() tea.Cmd {
	if d.currentIndex < len(d.conflicts)-1 {
		d.currentIndex++
		d.hunkIndex = 0
		d.syncViewports()
		return nil
	}
	// All conflicts resolved.
	d.quitting = true
	return tea.Quit
}

// syncViewports updates the viewport contents for the current conflict.
func (d *DiffViewer) syncViewports() {
	if len(d.conflicts) == 0 {
		return
	}
	conflict := d.conflicts[d.currentIndex]
	d.updateHunkStarts()

	halfWidth := d.width/2 - 6
	if halfWidth < 10 {
		halfWidth = 10
	}

	// Left viewport content.
	leftLines := d.renderPaneLines(conflict.CurrentLines, conflict.DiffLines, "current", halfWidth)
	leftContent := strings.Join(leftLines, "\n")

	// Right viewport content.
	rightLines := d.renderPaneLines(conflict.NewLines, conflict.DiffLines, "new", halfWidth)
	rightContent := strings.Join(rightLines, "\n")

	viewHeight := d.height - 8
	if viewHeight < 5 {
		viewHeight = 5
	}

	d.leftVP = viewport.New(
		viewport.WithWidth(halfWidth),
		viewport.WithHeight(viewHeight),
	)
	d.leftVP.SetContent(leftContent)

	d.rightVP = viewport.New(
		viewport.WithWidth(halfWidth),
		viewport.WithHeight(viewHeight),
	)
	d.rightVP.SetContent(rightContent)

	d.scrollToCurrentHunk()
}

// updateHunkStarts computes the starting diff-line indexes for changed hunks
// in the current conflict and keeps the active hunk index in range.
func (d *DiffViewer) updateHunkStarts() {
	if len(d.conflicts) == 0 || d.currentIndex < 0 || d.currentIndex >= len(d.conflicts) {
		d.hunkStarts = nil
		d.hunkIndex = 0
		return
	}

	d.hunkStarts = computeHunkStarts(d.conflicts[d.currentIndex].DiffLines)
	if len(d.hunkStarts) == 0 {
		d.hunkIndex = 0
		return
	}
	if d.hunkIndex < 0 {
		d.hunkIndex = 0
	}
	if d.hunkIndex >= len(d.hunkStarts) {
		d.hunkIndex = len(d.hunkStarts) - 1
	}
}

func computeHunkStarts(diffLines []DiffLine) []int {
	hunkStarts := make([]int, 0)
	inHunk := false
	for i, diffLine := range diffLines {
		isChange := diffLine.Type != DiffLineContext
		if isChange && !inHunk {
			hunkStarts = append(hunkStarts, i)
		}
		inHunk = isChange
	}
	return hunkStarts
}

func (d *DiffViewer) nextHunk() {
	if len(d.hunkStarts) == 0 {
		return
	}
	if d.hunkIndex < len(d.hunkStarts)-1 {
		d.hunkIndex++
	}
	d.scrollToCurrentHunk()
}

func (d *DiffViewer) prevHunk() {
	if len(d.hunkStarts) == 0 {
		return
	}
	if d.hunkIndex > 0 {
		d.hunkIndex--
	}
	d.scrollToCurrentHunk()
}

func (d *DiffViewer) scrollToCurrentHunk() {
	if len(d.hunkStarts) == 0 {
		d.leftVP.SetYOffset(0)
		d.rightVP.SetYOffset(0)
		return
	}

	d.updateHunkStarts()
	targetLine := d.hunkStarts[d.hunkIndex]
	d.leftVP.SetYOffset(targetLine)
	d.rightVP.SetYOffset(targetLine)
}

// renderHeader creates the conflict title bar.
func (d *DiffViewer) renderHeader(c ConflictView) string {
	num := d.currentIndex + 1
	total := len(d.conflicts)
	title := fmt.Sprintf("⚡ Conflict %d of %d: %s", num, total, c.FilePath)
	return Title(title)
}

// renderActionBar shows available actions.
func (d *DiffViewer) renderActionBar() string {
	actions := fmt.Sprintf(
		"%s %s  %s %s  %s %s",
		KeyBadge("[a]"),
		ValueText("Accept"),
		KeyBadge("[d]"),
		ValueText("Deny"),
		KeyBadge("[s]"),
		ValueText("Skip"),
	)
	return actions
}

// renderSummary shows the progress bar.
func (d *DiffViewer) renderSummary() string {
	total := len(d.conflicts)
	resolved := len(d.decisions)
	remaining := total - d.currentIndex
	return fmt.Sprintf("Conflicts: %d | Resolved: %d | Remaining: %d", total, resolved, remaining)
}

// renderHelp shows keybinding help.
func (d *DiffViewer) renderHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(Dimmed)
	lines := []string{
		"Keybindings:",
		"  a        — Accept library version",
		"  d        — Deny (keep current version)",
		"  s        — Skip (leave unresolved)",
		"  ↑/k      — Scroll content up",
		"  ↓/j      — Scroll content down",
		"  [/]      — Previous/next hunk",
		"  n        — Next conflict",
		"  p        — Previous conflict",
		"  ?        — Toggle this help",
		"  q/Ctrl+c — Quit",
	}
	return helpStyle.Render(strings.Join(lines, "\n"))
}

// renderPaneContent renders styled content for a single pane.
func (d *DiffViewer) renderPaneContent(lines []string, diffLines []DiffLine, side string, width int) string {
	rendered := d.renderPaneLines(lines, diffLines, side, width)
	return strings.Join(rendered, "\n")
}

// renderPaneLines renders lines with diff highlighting for a pane.
func (d *DiffViewer) renderPaneLines(lines []string, diffLines []DiffLine, side string, width int) []string {
	result := make([]string, 0, len(lines))

	if side == "current" {
		for _, diffLine := range diffLines {
			switch diffLine.Type {
			case DiffLineRemoved:
				trimmed := trimLineToWidth(diffLine.Content, width)
				result = append(result, lipgloss.NewStyle().
					Foreground(Error).
					Background(lipgloss.Color("#3B1426")).
					Width(width).
					Render("− "+trimmed))
			case DiffLineContext:
				trimmed := trimLineToWidth(diffLine.Content, width)
				result = append(result, lipgloss.NewStyle().
					Foreground(Dimmed).
					Width(width).
					Render("  "+trimmed))
			}
		}
	} else {
		for _, diffLine := range diffLines {
			switch diffLine.Type {
			case DiffLineAdded:
				trimmed := trimLineToWidth(diffLine.Content, width)
				result = append(result, lipgloss.NewStyle().
					Foreground(Success).
					Background(lipgloss.Color("#143B1A")).
					Width(width).
					Render("+ "+trimmed))
			case DiffLineContext:
				trimmed := trimLineToWidth(diffLine.Content, width)
				result = append(result, lipgloss.NewStyle().
					Foreground(Dimmed).
					Width(width).
					Render("  "+trimmed))
			}
		}
	}

	// If no diff lines, just show the raw content.
	if len(result) == 0 {
		for _, line := range lines {
			trimmed := trimLineToWidth(line, width)
			result = append(result, "  "+trimmed)
		}
	}

	return result
}

func trimLineToWidth(line string, width int) string {
	if width <= 3 || len(line) <= width {
		return line
	}
	return line[:width-3] + "..."
}
