package shared

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

// Field represents a form field with label, input, and error state
type Field struct {
	Label     string
	Input     textinput.Model
	IsFocused bool
	Error     string
}

// Render renders the field using the provided theme
func (f Field) Render(theme FormTheme) string {
	var b strings.Builder

	// Label with focus indicator
	if f.IsFocused {
		b.WriteString(theme.FocusedLabelStyle.Render(FocusIndicator + f.Label + ":"))
	} else {
		b.WriteString(theme.LabelStyle.Render(UnfocusedPrefix + f.Label + ":"))
	}
	b.WriteString("\n  ")
	b.WriteString(f.Input.View())

	// Field-level error (if any)
	if f.Error != "" {
		b.WriteString("\n  ")
		b.WriteString(theme.ErrorStyle.Render(f.Error))
	}

	return b.String()
}
