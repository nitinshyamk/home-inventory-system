package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
	"home-inventory-system/internal/ui/messages"
)

// Integration tests for full message flows through the orchestrator

func TestFullNavigationFlow(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)

	// Initialize
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return load command")
	}

	// Execute init command
	msg := cmd()
	typesMsg, ok := msg.(messages.TypesLoadedMsg)
	if !ok {
		t.Fatalf("Expected TypesLoadedMsg, got %T", msg)
	}

	// Handle types loaded
	var resultModel tea.Model
	resultModel, _ = m.Update(typesMsg)
	m = resultModel.(Model)

	if m.state != StateBrowsingTypes {
		t.Errorf("Expected state=%v after loading root types, got %v", StateBrowsingTypes, m.state)
	}
}

func TestErrorHandlingFlow(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)
	m.state = StateBrowsingTypes

	// Simulate error
	testErr := errors.New("test error")
	errorMsg := messages.ErrorMsg{Err: testErr}

	var resultModel tea.Model
	resultModel, _ = m.Update(errorMsg)
	m = resultModel.(Model)

	if m.state != StateError {
		t.Errorf("Expected state=Error, got %v", m.state)
	}

	if m.err == nil {
		t.Error("Expected error to be set")
	}
}

func TestWindowResizeHandling(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)

	resizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}

	var resultModel tea.Model
	resultModel, _ = m.Update(resizeMsg)
	m = resultModel.(Model)

	if m.width != 100 || m.height != 30 {
		t.Errorf("Expected dimensions 100x30, got %dx%d", m.width, m.height)
	}
}

func TestKeyPressHandling(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)
	m.state = StateBrowsingTypes

	tests := []struct {
		name    string
		keyType tea.KeyType
		rune    rune
		wantCmd bool
	}{
		{"quit key", tea.KeyRunes, 'q', true},
		{"navigate up from root", tea.KeyEsc, 0, false}, // No cmd when already at root
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var keyMsg tea.KeyMsg
			if tt.keyType == tea.KeyRunes {
				keyMsg = tea.KeyMsg{Type: tt.keyType, Runes: []rune{tt.rune}}
			} else {
				keyMsg = tea.KeyMsg{Type: tt.keyType}
			}

			_, cmd := m.Update(keyMsg)

			hasCmd := cmd != nil
			if hasCmd != tt.wantCmd {
				t.Errorf("Expected cmd=%v, got %v", tt.wantCmd, hasCmd)
			}
		})
	}
}

func TestDrillDownFlow(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)

	// Start by loading root types
	cmd := m.Init()
	msg := cmd()

	var resultModel tea.Model
	resultModel, _ = m.Update(msg)
	m = resultModel.(Model)

	// Should be browsing Electronics
	if m.state != StateBrowsingTypes {
		t.Fatalf("Expected state=BrowsingTypes, got %v", m.state)
	}

	// Simulate selecting Electronics (type ID 1)
	typeSelectedMsg := messages.TypeSelectedMsg{
		Type: domain.ItemType{ID: 1, Name: "Electronics"},
	}

	// This triggers async leaf check
	resultModel, cmd = m.Update(typeSelectedMsg)
	m = resultModel.(Model)

	if cmd == nil {
		t.Fatal("Expected command from type selection")
	}

	// Execute the leaf check
	leafCheckMsg := cmd()
	leafMsg, ok := leafCheckMsg.(messages.LeafCheckCompleteMsg)
	if !ok {
		t.Fatalf("Expected LeafCheckCompleteMsg, got %T", leafCheckMsg)
	}

	// Handle leaf check result
	resultModel, cmd = m.Update(leafMsg)
	m = resultModel.(Model)

	// Should either load child types or items depending on IsLeaf
	if cmd == nil {
		t.Error("Expected command to load either children or items")
	}
}

func TestViewRendering(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)

	testErr := errors.New("test error")

	tests := []struct {
		name  string
		state AppState
	}{
		{"loading state", StateLoading},
		{"browsing types state", StateBrowsingTypes},
		{"viewing items state", StateViewingItems},
		{"error state", StateError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.state = tt.state
			if tt.state == StateError {
				m.err = testErr
			}

			view := m.View()
			if view == "" {
				t.Error("Expected non-empty view output")
			}
		})
	}
}

func TestMessageDelegationToItemList(t *testing.T) {
	handler := setupTestHandler(t)
	m := NewModel(handler)
	m.state = StateBrowsingTypes

	// Send a key that should be delegated to itemlist
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}

	_, _ = m.Update(keyMsg)

	// Just verify no panic occurred
	// The actual itemlist behavior is tested in its own package
}
