package styles

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	primaryColor   = lipgloss.Color("62")
	secondaryColor = lipgloss.Color("241")
	accentColor    = lipgloss.Color("205")
	errorColor     = lipgloss.Color("196")
	successColor   = lipgloss.Color("46")
)

// Application styles
var (
	// TitleStyle for the application title
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	// SubtitleStyle for secondary text
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	// SelectedStyle for highlighted items
	SelectedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// NormalStyle for regular text
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// DimStyle for less important text
	DimStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// HelpStyle for help text at the bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(1, 0, 0, 0)
)
