package categorypicker

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
)

// Entry holds a leaf category with its full display path.
type Entry struct {
	TypeID   int64
	FullPath string // e.g. "Electronics > Computers"
}

// Model is an inline filterable category picker rendered in the right pane.
type Model struct {
	filterInput textinput.Model
	entries     []Entry  // all leaf categories with full paths
	filtered    []Entry  // entries matching the current filter
	cursor      int      // position in filtered list
	selected    *Entry   // non-nil after user confirms selection
	cancelled   bool
}

// New creates a picker pre-loaded with leaf category entries.
func New(entries []Entry) Model {
	filter := textinput.New()
	filter.Placeholder = "type to filter..."
	filter.CharLimit = 100
	filter.Width = 40
	filter.Focus()

	m := Model{
		filterInput: filter,
		entries:     entries,
		filtered:    entries,
	}
	return m
}

// Update handles key events for the picker.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.cancelled = true
		return m, nil

	case "enter":
		if len(m.filtered) > 0 {
			entry := m.filtered[m.cursor]
			m.selected = &entry
		}
		return m, nil

	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "ctrl+n":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		return m, nil
	}

	// Delegate to filter input
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m = m.applyFilter()
	return m, cmd
}

func (m Model) applyFilter() Model {
	query := strings.ToLower(m.filterInput.Value())
	if query == "" {
		m.filtered = m.entries
		m.cursor = 0
		return m
	}
	filtered := m.filtered[:0:0]
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.FullPath), query) {
			filtered = append(filtered, e)
		}
	}
	m.filtered = filtered
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m
}

// Selected returns the chosen entry, or nil if nothing selected yet.
func (m Model) Selected() *Entry { return m.selected }

// Cancelled reports whether the user pressed Esc.
func (m Model) Cancelled() bool { return m.cancelled }

// FilterInput returns the current filter string (for rendering).
func (m Model) FilterInput() textinput.Model { return m.filterInput }

// Filtered returns the current filtered entry list.
func (m Model) Filtered() []Entry { return m.filtered }

// Cursor returns the current cursor position within Filtered.
func (m Model) Cursor() int { return m.cursor }

// BuildEntries converts domain data (leaf types + paths) into picker entries.
func BuildEntries(leafTypes []domain.ItemType, paths map[int64][]domain.ItemType) []Entry {
	entries := make([]Entry, 0, len(leafTypes))
	for _, t := range leafTypes {
		path, ok := paths[t.ID]
		if !ok || len(path) == 0 {
			entries = append(entries, Entry{TypeID: t.ID, FullPath: t.Name})
			continue
		}
		parts := make([]string, len(path))
		for i, p := range path {
			parts[i] = p.Name
		}
		entries = append(entries, Entry{TypeID: t.ID, FullPath: strings.Join(parts, " > ")})
	}
	return entries
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
