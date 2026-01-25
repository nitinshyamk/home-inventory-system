package ui

import (
	"strings"
	"testing"

	"home-inventory-system/internal/domain"
)

func TestGetParent(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb []domain.ItemType
		wantNil    bool
		wantID     int64
	}{
		{
			name:       "empty breadcrumb",
			breadcrumb: []domain.ItemType{},
			wantNil:    true,
		},
		{
			name:       "single item (at root)",
			breadcrumb: []domain.ItemType{{ID: 1, Name: "Category"}},
			wantNil:    true,
		},
		{
			name: "two items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Parent"},
				{ID: 2, Name: "Child"},
			},
			wantNil: false,
			wantID:  1,
		},
		{
			name: "three items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Root"},
				{ID: 2, Name: "Parent"},
				{ID: 3, Name: "Child"},
			},
			wantNil: false,
			wantID:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParent(tt.breadcrumb)
			if tt.wantNil {
				if got != nil {
					t.Errorf("getParent() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("getParent() = nil, want ID %d", tt.wantID)
				} else if got.ID != tt.wantID {
					t.Errorf("getParent().ID = %d, want %d", got.ID, tt.wantID)
				}
			}
		})
	}
}

func TestGetGrandparent(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb []domain.ItemType
		wantNil    bool
		wantID     int64
	}{
		{
			name:       "empty breadcrumb",
			breadcrumb: []domain.ItemType{},
			wantNil:    true,
		},
		{
			name: "two items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Parent"},
				{ID: 2, Name: "Child"},
			},
			wantNil: true,
		},
		{
			name: "three items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Grandparent"},
				{ID: 2, Name: "Parent"},
				{ID: 3, Name: "Child"},
			},
			wantNil: false,
			wantID:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGrandparent(tt.breadcrumb)
			if tt.wantNil {
				if got != nil {
					t.Errorf("getGrandparent() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("getGrandparent() = nil, want ID %d", tt.wantID)
				} else if got.ID != tt.wantID {
					t.Errorf("getGrandparent().ID = %d, want %d", got.ID, tt.wantID)
				}
			}
		})
	}
}

func TestGetCurrentType(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb []domain.ItemType
		wantNil    bool
		wantID     int64
	}{
		{
			name:       "empty breadcrumb",
			breadcrumb: []domain.ItemType{},
			wantNil:    true,
		},
		{
			name:       "single item",
			breadcrumb: []domain.ItemType{{ID: 5, Name: "Current"}},
			wantNil:    false,
			wantID:     5,
		},
		{
			name: "multiple items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Parent"},
				{ID: 2, Name: "Current"},
			},
			wantNil: false,
			wantID:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCurrentType(tt.breadcrumb)
			if tt.wantNil {
				if got != nil {
					t.Errorf("getCurrentType() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("getCurrentType() = nil, want ID %d", tt.wantID)
				} else if got.ID != tt.wantID {
					t.Errorf("getCurrentType().ID = %d, want %d", got.ID, tt.wantID)
				}
			}
		})
	}
}

func TestIsAtRoot(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb []domain.ItemType
		want       bool
	}{
		{
			name:       "empty breadcrumb",
			breadcrumb: []domain.ItemType{},
			want:       true,
		},
		{
			name:       "single item",
			breadcrumb: []domain.ItemType{{ID: 1, Name: "Category"}},
			want:       true,
		},
		{
			name: "two items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Parent"},
				{ID: 2, Name: "Child"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAtRoot(tt.breadcrumb); got != tt.want {
				t.Errorf("isAtRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDepth(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb []domain.ItemType
		want       int
	}{
		{
			name:       "empty breadcrumb",
			breadcrumb: []domain.ItemType{},
			want:       0,
		},
		{
			name:       "single item",
			breadcrumb: []domain.ItemType{{ID: 1, Name: "Category"}},
			want:       1,
		},
		{
			name: "three items",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
				{ID: 3, Name: "C"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDepth(tt.breadcrumb); got != tt.want {
				t.Errorf("getDepth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderBreadcrumb(t *testing.T) {
	tests := []struct {
		name         string
		breadcrumb   []domain.ItemType
		wantContains []string // Strings that should appear in output
	}{
		{
			name:         "empty breadcrumb shows Home",
			breadcrumb:   []domain.ItemType{},
			wantContains: []string{"Home"},
		},
		{
			name: "single item shows Home > Item",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Electronics"},
			},
			wantContains: []string{"Home", ">", "Electronics"},
		},
		{
			name: "nested items show full path",
			breadcrumb: []domain.ItemType{
				{ID: 1, Name: "Electronics"},
				{ID: 2, Name: "Computers"},
				{ID: 3, Name: "Laptops"},
			},
			wantContains: []string{"Home", ">", "Electronics", "Computers", "Laptops"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderBreadcrumb(tt.breadcrumb)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("renderBreadcrumb() missing %q in output: %s", want, got)
				}
			}
		})
	}
}

func TestClearBreadcrumb(t *testing.T) {
	m := Model{
		breadcrumb: []domain.ItemType{{ID: 1, Name: "Test"}},
	}

	m = clearBreadcrumb(m)

	if m.breadcrumb != nil {
		t.Errorf("clearBreadcrumb() did not clear breadcrumb, got %v", m.breadcrumb)
	}
}

func TestSetBreadcrumb(t *testing.T) {
	m := Model{breadcrumb: nil}

	newPath := []domain.ItemType{
		{ID: 1, Name: "Parent"},
		{ID: 2, Name: "Child"},
	}

	m = setBreadcrumb(m, newPath)

	if len(m.breadcrumb) != 2 {
		t.Errorf("setBreadcrumb() breadcrumb length = %d, want 2", len(m.breadcrumb))
	}
	if m.breadcrumb[0].ID != 1 {
		t.Errorf("setBreadcrumb() first item ID = %d, want 1", m.breadcrumb[0].ID)
	}
}
