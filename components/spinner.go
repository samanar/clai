package components

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))

type SpinnerModel struct {
	spinner    spinner.Model
	message    string
	done       bool
	err        error
	cancelChan chan<- struct{}
}

type spinnerDoneMsg struct {
	err error
}

// NewSpinnerModel creates a new spinner model with a custom message
func NewSpinnerModel(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = spinnerStyle
	return SpinnerModel{
		spinner: s,
		message: message,
	}
}

// Init initializes the spinner model
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the spinner model
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			// Signal cancellation and exit immediately
			if m.cancelChan != nil {
				close(m.cancelChan)
			}
			fmt.Println("\n\nOperation cancelled by user")
			os.Exit(130) // Standard exit code for Ctrl+C
		}
		if msg.String() == "q" {
			return m, tea.Quit
		}
		return m, nil

	case spinnerDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

// View renders the spinner view
func (m SpinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return ""
		}
		return ""
	}
	return fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), m.message)
}

// Done signals that the spinner should stop
func (m SpinnerModel) Done(err error) tea.Msg {
	return spinnerDoneMsg{err: err}
}

// Error returns the error if spinner finished with an error
func (m SpinnerModel) Error() error {
	return m.err
}

// ShowSpinner displays a spinner with a custom message until a channel signals completion
func ShowSpinner(message string, doneChan <-chan error, cancelChan chan<- struct{}) error {
	m := NewSpinnerModel(message)
	m.cancelChan = cancelChan
	p := tea.NewProgram(m)

	// Start the program in a goroutine
	go func() {
		if _, err := p.Run(); err != nil {
			return
		}
	}()

	// Wait for the done signal
	err := <-doneChan
	p.Send(spinnerDoneMsg{err: err})
	p.Wait()

	return err
}
