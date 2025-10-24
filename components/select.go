package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).PaddingLeft(1)
	normalStyle      = lipgloss.NewStyle().PaddingLeft(2)
	cursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
)

// SelectOption represents an option in the select menu
type SelectOption struct {
	Title       string
	Description string
	Value       string // The value to return when selected
}

// SelectModel represents the select menu model
type SelectModel struct {
	options  []SelectOption
	cursor   int
	selected string
	done     bool
}

// NewSelectModel creates a new select model
func NewSelectModel(options []SelectOption) SelectModel {
	return SelectModel{
		options: options,
		cursor:  0,
	}
}

// Init initializes the select model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the select model
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.done = true
			return m, tea.Quit

		case "enter":
			m.selected = m.options[m.cursor].Value
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the select menu
func (m SelectModel) View() string {
	if m.done {
		return ""
	}

	var s strings.Builder
	s.WriteString("\n")
	s.WriteString("  Use ↑/↓ or j/k to navigate, Enter to select, q to quit\n\n")

	for i, option := range m.options {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
			style = selectedStyle
		}

		title := titleStyle.Render(option.Title)
		description := descriptionStyle.Render(option.Description)

		s.WriteString(style.Render(fmt.Sprintf("%s%s", cursor, title)))
		if option.Description != "" {
			s.WriteString(style.Render(fmt.Sprintf("%s\n", description)))
		}
		s.WriteString("\n")
	}

	return s.String()
}

// Selected returns the selected value
func (m SelectModel) Selected() string {
	return m.selected
}

// Select displays a select menu and returns the selected value
func Select(options []SelectOption) (string, error) {
	m := NewSelectModel(options)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running select: %w", err)
	}

	selectModel := finalModel.(SelectModel)
	return selectModel.Selected(), nil
}
