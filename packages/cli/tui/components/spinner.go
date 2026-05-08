// Package components provides reusable TUI components built on top of the
// Charm Bracelet stack (bubbletea, lipgloss, bubbles) and the project theme.
package components

import (
	"fmt"
	"io"
	"os"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

// ─── Spinner Model (Bubble Tea compatible) ────────────────────────────────────

// SpinnerModel is a Bubble Tea model that displays an animated spinner with
// a message. Use it inside a tea.Program for full interactivity, or use the
// Spinner struct for simpler CLI-style usage.
type SpinnerModel struct {
	spinner spinner.Model
	message string
	style   lipgloss.Style
	quit    bool
}

// NewSpinnerModel creates a Bubble Tea model for a themed spinner.
func NewSpinnerModel(message string) SpinnerModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(theme.Primary)),
	)
	return SpinnerModel{
		spinner: s,
		message: message,
		style:   lipgloss.NewStyle().Foreground(theme.Text),
	}
}

// Init starts the spinner ticking.
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles Bubble Tea messages.
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case StopMsg:
		m.quit = true
		return m, func() tea.Msg { return tea.Quit() }
	}
	return m, nil
}

// View renders the spinner.
func (m SpinnerModel) View() tea.View {
	if m.quit {
		return tea.NewView(theme.SuccessLabel(m.message))
	}
	return tea.NewView(m.spinner.View() + " " + m.style.Render(m.message))
}

// StopMsg is a Bubble Tea message that signals the spinner to stop.
type StopMsg struct{}

// ─── Spinner convenience wrapper (CLI-style) ──────────────────────────────────

// Spinner is a convenience wrapper for running a spinner in a CLI context
// (outside a Bubble Tea program). It manages its own goroutine for the
// animation and writes to an io.Writer.
type Spinner struct {
	model   SpinnerModel
	program *tea.Program
	done    chan struct{}
	writer  io.Writer
}

// NewSpinner creates a CLI-style spinner with the given message.
// By default it writes to os.Stdout.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		model:  NewSpinnerModel(message),
		writer: os.Stdout,
		done:   make(chan struct{}),
	}
}

// Start begins the spinner animation in the background.
func (s *Spinner) Start() {
	opts := []tea.ProgramOption{tea.WithOutput(s.writer)}
	prog := tea.NewProgram(s.model, opts...)
	s.program = prog
	go func() {
		_, _ = prog.Run()
		close(s.done)
	}()
}

// Stop signals the spinner to stop and waits for it to finish.
func (s *Spinner) Stop() {
	if s.program != nil {
		s.program.Send(StopMsg{})
		<-s.done
	}
	// Clear the spinner line.
	if s.writer != nil {
		fmt.Fprint(s.writer, "\r\033[K")
	}
}

// UpdateMessage changes the spinner text while running.
func (s *Spinner) UpdateMessage(msg string) {
	s.model.message = msg
	if s.program != nil {
		s.program.Send(nil) // trigger re-render
	}
}
