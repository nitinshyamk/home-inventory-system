package modal

import "github.com/charmbracelet/lipgloss"

// Options configures the modal appearance
type Options struct {
	Width       int
	MaxWidth    int
	BorderColor lipgloss.Color
}

// DefaultOptions returns sensible defaults for modal rendering
func DefaultOptions() Options {
	return Options{
		Width:       60,
		MaxWidth:    60,
		BorderColor: lipgloss.Color("205"),
	}
}

// Render renders content as a centered modal overlay
func Render(screenWidth, screenHeight int, content string, opts Options) string {
	// Calculate form dimensions (adaptive to terminal size)
	formWidth := min(screenWidth-4, opts.MaxWidth)
	if opts.Width > 0 && opts.Width < formWidth {
		formWidth = opts.Width
	}

	// Create a border style for the modal
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(opts.BorderColor).
		Padding(1, 2).
		Width(formWidth)

	// Apply border
	formModal := borderStyle.Render(content)

	// Center the modal on the screen
	overlayStyle := lipgloss.NewStyle().
		Width(screenWidth).
		Height(screenHeight).
		Align(lipgloss.Center, lipgloss.Center)

	return overlayStyle.Render(formModal)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
