package itemedit

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"home-inventory-system/internal/domain"
)

// Focus positions: 0=Name, 1=Quantity, 2=Unit, 3=ChangeCategory, 4=Save, 5=Cancel
const (
	focusName = iota
	focusQty
	focusUnit
	focusChangeCategory
	focusSave
	focusCancel
	focusCount
)

var unitOrder = []domain.UnitType{
	domain.UnitTypeCount,
	domain.UnitTypeGrams,
	domain.UnitTypeLiters,
}

// Model represents an item edit form rendered in the right pane.
type Model struct {
	nameInput        textinput.Model
	qtyInput         textinput.Model
	unitIndex        int             // index into unitOrder
	focusIndex       int
	originalName     string
	originalQty      string
	originalUnit     domain.UnitType
	originalType     int64
	currentType      int64  // item_type_id (may be changed via picker)
	currentTypePath  string // display path for currentType
	changingCategory bool   // true when the picker button was pressed (triggers parent)
	errorMsg         string
	submitted        bool
	cancelled        bool
}

// New creates an item edit form pre-filled with current item data.
func New(item domain.Item) Model {
	nameInput := textinput.New()
	nameInput.SetValue(item.Name)
	nameInput.CharLimit = 100
	nameInput.Width = 40
	nameInput.Focus()

	qtyStr := strconv.FormatFloat(item.Quantity, 'f', -1, 64)
	qtyInput := textinput.New()
	qtyInput.SetValue(qtyStr)
	qtyInput.CharLimit = 20
	qtyInput.Width = 20

	unitIdx := 0
	for i, u := range unitOrder {
		if u == item.UnitType {
			unitIdx = i
			break
		}
	}

	return Model{
		nameInput:    nameInput,
		qtyInput:     qtyInput,
		unitIndex:    unitIdx,
		focusIndex:   focusName,
		originalName: item.Name,
		originalQty:  qtyStr,
		originalUnit: item.UnitType,
		originalType: item.ItemTypeID,
		currentType:  item.ItemTypeID,
	}
}

// SetCategoryPath updates the displayed path for the current category (called after picker).
func (m *Model) SetCategoryPath(path string) {
	m.currentTypePath = path
}

// ChangingCategory reports that the "Change..." button was activated (parent should open picker).
func (m Model) ChangingCategory() bool { return m.changingCategory }

// ClearChangingCategory resets the flag (called by the parent after opening the picker).
func (m *Model) ClearChangingCategory() { m.changingCategory = false }

// Update handles key events for the item edit form.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.cancelled = true
		return m, nil

	case "tab", "shift+tab":
		if keyMsg.String() == "shift+tab" {
			m.focusIndex = (m.focusIndex - 1 + focusCount) % focusCount
		} else {
			m.focusIndex = (m.focusIndex + 1) % focusCount
		}
		return m.applyFocus(), nil

	case "enter":
		switch m.focusIndex {
		case focusSave:
			if m.HasChanges() {
				return m.submit(), nil
			}
		case focusCancel:
			m.cancelled = true
		case focusChangeCategory:
			// Signal to parent to open the picker; parent will switch state
			m.changingCategory = true
		case focusName:
			m.focusIndex = focusQty
			return m.applyFocus(), nil
		case focusQty:
			m.focusIndex = focusUnit
			return m.applyFocus(), nil
		case focusUnit:
			m.focusIndex = focusChangeCategory
			return m.applyFocus(), nil
		}
		return m, nil

	case "up", "ctrl+p":
		if m.focusIndex == focusUnit {
			m.unitIndex = (m.unitIndex - 1 + len(unitOrder)) % len(unitOrder)
		}
		return m, nil

	case "down", "ctrl+n":
		if m.focusIndex == focusUnit {
			m.unitIndex = (m.unitIndex + 1) % len(unitOrder)
		}
		return m, nil
	}

	// Delegate to focused text input
	var cmd tea.Cmd
	switch m.focusIndex {
	case focusName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case focusQty:
		m.qtyInput, cmd = m.qtyInput.Update(msg)
	}
	return m, cmd
}

func (m Model) applyFocus() Model {
	switch m.focusIndex {
	case focusName:
		m.nameInput.Focus()
		m.qtyInput.Blur()
	case focusQty:
		m.nameInput.Blur()
		m.qtyInput.Focus()
	default:
		m.nameInput.Blur()
		m.qtyInput.Blur()
	}
	return m
}

func (m Model) submit() Model {
	if m.nameInput.Value() == "" {
		m.errorMsg = "Name is required"
		return m
	}
	if _, err := strconv.ParseFloat(m.qtyInput.Value(), 64); err != nil {
		m.errorMsg = "Quantity must be a number"
		return m
	}
	qty, _ := strconv.ParseFloat(m.qtyInput.Value(), 64)
	if qty <= 0 {
		m.errorMsg = "Quantity must be greater than 0"
		return m
	}
	m.submitted = true
	return m
}

// HasChanges reports whether any field differs from the original.
func (m Model) HasChanges() bool {
	return m.nameInput.Value() != m.originalName ||
		m.qtyInput.Value() != m.originalQty ||
		unitOrder[m.unitIndex] != m.originalUnit ||
		m.currentType != m.originalType
}

// Submitted / Cancelled / SetError satisfy the same contract as categoryedit.
func (m Model) Submitted() bool  { return m.submitted }
func (m Model) Cancelled() bool  { return m.cancelled }
func (m Model) Error() string    { return m.errorMsg }
func (m Model) FocusIndex() int  { return m.focusIndex }

// SetError sets an inline validation error (e.g. from a server response).
func (m *Model) SetError(msg string) {
	m.errorMsg = msg
	m.submitted = false
}

// SetCurrentType updates the item_type_id (used by the category picker in Task 5).
func (m *Model) SetCurrentType(typeID int64) {
	m.currentType = typeID
}

// Name, Quantity, UnitType, TypeID return the current form values.
func (m Model) Name() string             { return m.nameInput.Value() }
func (m Model) UnitType() domain.UnitType { return unitOrder[m.unitIndex] }
func (m Model) TypeID() int64            { return m.currentType }

func (m Model) Quantity() float64 {
	q, _ := strconv.ParseFloat(m.qtyInput.Value(), 64)
	return q
}

// CurrentTypePath returns the display path for the current category.
func (m Model) CurrentTypePath() string { return m.currentTypePath }

// Exported focus constants
const (
	FocusName           = focusName
	FocusQty            = focusQty
	FocusUnit           = focusUnit
	FocusChangeCategory = focusChangeCategory
	FocusSave           = focusSave
	FocusCancel         = focusCancel
)
