// Package diffviewer provides an interactive side-by-side diff viewer built on
// the Charm Bracelet Bubble Tea framework. It is the hero component of the
// conflict resolution phase, allowing users to see old vs. new content and
// choose to accept, deny, or skip each conflict.
package diffviewer

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/diff"
	"github.com/ricardoborges-teachable/ai-setup/tui/theme"
)

// ResolutionAction defines what to do with a conflicting file.
type ResolutionAction string

const (
	ActionAccept ResolutionAction = "accept" // Keep library version
	ActionDeny   ResolutionAction = "deny"   // Keep current version
	ActionSkip   ResolutionAction = "skip"   // Leave unresolved
)

// Resolution records the user's decision for a single conflict.
type Resolution struct {
	Path   string
	Action ResolutionAction
}

// ConflictView holds the data for visualizing a single conflict.
type ConflictView struct {
	FilePath     string
	CurrentLines []string // "ours" / existing file lines
	NewLines     []string // "theirs" / library file lines
	DiffLines    []diff.DiffLine
}

// DiffViewer is a Bubble Tea model for side-by-side conflict resolution.
type DiffViewer struct {
	conflicts    []ConflictView
	resolutions  []Resolution
	currentIndex int
	leftVP       viewport.Model
	rightVP      viewport.Model
	width        int
	height       int
	showHelp     bool
	quitting     bool
}

// keyMap defines the keybindings for the diff viewer.
type diffKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Accept key.Binding
	Deny   key.Binding
	Skip   key.Binding
	Next   key.Binding
	Prev   key.Binding
	Help   key.Binding
	Quit   key.Binding
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
			conflicts[i].DiffLines = diff.ComputeDiff([]byte(currentContent), []byte(newContent))
		}
	}
	return &DiffViewer{
		conflicts:    conflicts,
		resolutions:  make([]Resolution, 0, len(conflicts)),
		currentIndex: 0,
		showHelp:     false,
		quitting:     false,
	}
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
				d.syncViewports()
			}
			return d, nil

		case key.Matches(msg, diffKeys.Prev):
			if d.currentIndex > 0 {
				d.currentIndex--
				d.syncViewports()
			}
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
		v := tea.NewView("Goodbye!\n")
		v.AltScreen = true
		return v
	}

	if len(d.conflicts) == 0 {
		return tea.NewView(theme.SuccessLabel("No conflicts to resolve.") + "\n")
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

	// Left pane: Current (yours).
	leftTitle := theme.ErrorLabel("Current (yours)")
	leftContent := d.renderPaneContent(conflict.CurrentLines, conflict.DiffLines, "current", halfWidth)

	// Right pane: New (library).
	rightTitle := theme.SuccessLabel("New (library)")
	rightContent := d.renderPaneContent(conflict.NewLines, conflict.DiffLines, "new", halfWidth)

	// Header.
	header := d.renderHeader(conflict)

	// Action bar.
	actionBar := d.renderActionBar()

	// Summary bar.
	summaryBar := d.renderSummary()

	// Assemble panes.
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Error).
		Width(halfWidth).
		Render(leftTitle + "\n" + leftContent)

	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Success).
		Width(halfWidth).
		Render(rightTitle + "\n" + rightContent)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	_ = header
	_ = actionBar
	_ = summaryBar

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

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
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

	return d.resolutions, nil
}

// resolveCurrent records the resolution for the current conflict.
func (d *DiffViewer) resolveCurrent(action ResolutionAction) {
	if d.currentIndex < len(d.conflicts) {
		d.resolutions = append(d.resolutions, Resolution{
			Path:   d.conflicts[d.currentIndex].FilePath,
			Action: action,
		})
	}
}

// advanceOrQuit moves to the next conflict or quits if all are resolved.
func (d *DiffViewer) advanceOrQuit() tea.Cmd {
	if d.currentIndex < len(d.conflicts)-1 {
		d.currentIndex++
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
}

// renderHeader creates the conflict title bar.
func (d *DiffViewer) renderHeader(c ConflictView) string {
	num := d.currentIndex + 1
	total := len(d.conflicts)
	title := fmt.Sprintf("⚡ Conflict %d of %d: %s", num, total, c.FilePath)
	return theme.Title(title)
}

// renderActionBar shows available actions.
func (d *DiffViewer) renderActionBar() string {
	actions := fmt.Sprintf(
		"%s %s  %s %s  %s %s",
		theme.KeyBadge("[a]"),
		theme.ValueText("Accept"),
		theme.KeyBadge("[d]"),
		theme.ValueText("Deny"),
		theme.KeyBadge("[s]"),
		theme.ValueText("Skip"),
	)
	return actions
}

// renderSummary shows the progress bar.
func (d *DiffViewer) renderSummary() string {
	total := len(d.conflicts)
	resolved := len(d.resolutions)
	remaining := total - d.currentIndex
	return fmt.Sprintf("Conflicts: %d | Resolved: %d | Remaining: %d", total, resolved, remaining)
}

// renderHelp shows keybinding help.
func (d *DiffViewer) renderHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	lines := []string{
		"Keybindings:",
		"  a        — Accept library version",
		"  d        — Deny (keep current version)",
		"  s        — Skip (leave unresolved)",
		"  ↑/k      — Scroll content up",
		"  ↓/j      — Scroll content down",
		"  n        — Next conflict",
		"  p        — Previous conflict",
		"  ?        — Toggle this help",
		"  q/Ctrl+c — Quit",
	}
	return helpStyle.Render(strings.Join(lines, "\n"))
}

// renderPaneContent renders styled content for a single pane.
func (d *DiffViewer) renderPaneContent(lines []string, diffLines []diff.DiffLine, side string, width int) string {
	rendered := d.renderPaneLines(lines, diffLines, side, width)
	return strings.Join(rendered, "\n")
}

// renderPaneLines renders lines with diff highlighting for a pane.
func (d *DiffViewer) renderPaneLines(lines []string, diffLines []diff.DiffLine, side string, width int) []string {
	result := make([]string, 0, len(lines))

	if side == "current" {
		for _, dl := range diffLines {
			switch dl.Type {
			case diff.DiffLineRemoved:
				trimmed := dl.Content
				if len(trimmed) > width {
					trimmed = trimmed[:width-3] + "..."
				}
				result = append(result, lipgloss.NewStyle().
					Foreground(theme.Error).
					Background(lipgloss.Color("#3B1426")).
					Width(width).
					Render("− "+trimmed))
			case diff.DiffLineContext:
				trimmed := dl.Content
				if len(trimmed) > width {
					trimmed = trimmed[:width-3] + "..."
				}
				result = append(result, lipgloss.NewStyle().
					Foreground(theme.Dimmed).
					Width(width).
					Render("  "+trimmed))
			}
		}
	} else {
		for _, dl := range diffLines {
			switch dl.Type {
			case diff.DiffLineAdded:
				trimmed := dl.Content
				if len(trimmed) > width {
					trimmed = trimmed[:width-3] + "..."
				}
				result = append(result, lipgloss.NewStyle().
					Foreground(theme.Success).
					Background(lipgloss.Color("#143B1A")).
					Width(width).
					Render("+ "+trimmed))
			case diff.DiffLineContext:
				trimmed := dl.Content
				if len(trimmed) > width {
					trimmed = trimmed[:width-3] + "..."
				}
				result = append(result, lipgloss.NewStyle().
					Foreground(theme.Dimmed).
					Width(width).
					Render("  "+trimmed))
			}
		}
	}

	// If no diff lines, just show the raw content.
	if len(result) == 0 {
		for _, line := range lines {
			trimmed := line
			if len(trimmed) > width {
				trimmed = trimmed[:width-3] + "..."
			}
			result = append(result, "  "+trimmed)
		}
	}

	return result
}
