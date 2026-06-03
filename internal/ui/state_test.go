package ui

import "testing"

func TestAppState_String(t *testing.T) {
	tests := []struct {
		state AppState
		want  string
	}{
		{StateLoading, "Loading"},
		{StateBrowsingTypes, "BrowsingTypes"},
		{StateViewingItems, "ViewingItems"},
		{StateError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("AppState.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppState_CanNavigateUp(t *testing.T) {
	tests := []struct {
		state AppState
		want  bool
	}{
		{StateLoading, false},
		{StateBrowsingTypes, true},
		{StateViewingItems, true},
		{StateError, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.CanNavigateUp(); got != tt.want {
				t.Errorf("AppState.CanNavigateUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppState_CanSelectItem(t *testing.T) {
	tests := []struct {
		state AppState
		want  bool
	}{
		{StateLoading, false},
		{StateBrowsingTypes, true},
		{StateViewingItems, true},
		{StateError, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.CanSelectItem(); got != tt.want {
				t.Errorf("AppState.CanSelectItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppState_AllowsItemListDelegation(t *testing.T) {
	tests := []struct {
		state AppState
		want  bool
	}{
		{StateLoading, false},
		{StateBrowsingTypes, true},
		{StateViewingItems, true},
		{StateError, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.AllowsItemListDelegation(); got != tt.want {
				t.Errorf("AppState.AllowsItemListDelegation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransitionTo(t *testing.T) {
	m := Model{state: StateLoading}

	m = transitionTo(m, StateBrowsingTypes)
	if m.state != StateBrowsingTypes {
		t.Errorf("Expected state %v, got %v", StateBrowsingTypes, m.state)
	}

	m = transitionTo(m, StateError)
	if m.state != StateError {
		t.Errorf("Expected state %v, got %v", StateError, m.state)
	}
}
