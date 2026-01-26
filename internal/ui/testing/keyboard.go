package testing

import (
	tea "github.com/charmbracelet/bubbletea"
)

// KeyEvent represents a keyboard event to send to the UI
type KeyEvent struct {
	msg tea.KeyMsg
}

// Predefined key events
var (
	KeyUp        = KeyEvent{tea.KeyMsg{Type: tea.KeyUp}}
	KeyDown      = KeyEvent{tea.KeyMsg{Type: tea.KeyDown}}
	KeyLeft      = KeyEvent{tea.KeyMsg{Type: tea.KeyLeft}}
	KeyRight     = KeyEvent{tea.KeyMsg{Type: tea.KeyRight}}
	KeyEnter     = KeyEvent{tea.KeyMsg{Type: tea.KeyEnter}}
	KeyEsc       = KeyEvent{tea.KeyMsg{Type: tea.KeyEscape}}
	KeyTab       = KeyEvent{tea.KeyMsg{Type: tea.KeyTab}}
	KeyShiftTab  = KeyEvent{tea.KeyMsg{Type: tea.KeyShiftTab}}
	KeyBackspace = KeyEvent{tea.KeyMsg{Type: tea.KeyBackspace}}
	KeyDelete    = KeyEvent{tea.KeyMsg{Type: tea.KeyDelete}}
	KeyCtrlA     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlA}}
	KeyCtrlC     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlC}}
	KeyCtrlP     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlP}}
	KeyCtrlN     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlN}}
	KeyCtrlF     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlF}}
	KeyCtrlB     = KeyEvent{tea.KeyMsg{Type: tea.KeyCtrlB}}
)

// Type creates a sequence of character key events from a string
func Type(s string) []KeyEvent {
	events := make([]KeyEvent, len(s))
	for i, ch := range s {
		events[i] = KeyEvent{tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{ch},
		}}
	}
	return events
}

// Key creates a single key event from a string
func Key(s string) KeyEvent {
	return KeyEvent{tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(s),
	}}
}
