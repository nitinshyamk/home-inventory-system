package shared

import "github.com/charmbracelet/lipgloss"

// FormTheme contains all styles for form rendering
type FormTheme struct {
	TitleStyle        lipgloss.Style
	LabelStyle        lipgloss.Style
	FocusedLabelStyle lipgloss.Style
	ErrorStyle        lipgloss.Style
	HelpStyle         lipgloss.Style
}

// DefaultFormTheme returns the default form styling theme
func DefaultFormTheme() FormTheme {
	return FormTheme{
		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),

		LabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		FocusedLabelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}

// FocusIndicator returns the bullet point indicator for focused fields
const FocusIndicator = "• "

// UnfocusedPrefix returns the spaces used for unfocused field alignment
const UnfocusedPrefix = "  "
