package testing

import (
	"database/sql"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"

	"home-inventory-system/internal/db"
	"home-inventory-system/internal/repository"
	"home-inventory-system/internal/service"
	"home-inventory-system/internal/ui"
)

// Simulator provides a test harness for UI integration testing
type Simulator struct {
	model   ui.Model
	handler *service.Handler
	repo    *repository.Repository
	dbConn  *sql.DB
}

// NewSimulator creates a new UI simulator with an in-memory database
func NewSimulator(t *testing.T) *Simulator {
	t.Helper()

	// Create in-memory SQLite database
	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(dbConn); err != nil {
		dbConn.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create repository and service handler
	repo := repository.New(dbConn)
	handler := service.NewHandler(repo)

	// Create UI model
	model := ui.NewModel(handler)

	sim := &Simulator{
		model:   model,
		handler: handler,
		repo:    repo,
		dbConn:  dbConn,
	}

	// Initialize the model (this triggers Init() which loads root types)
	initCmd := model.Init()

	// Set window size
	sim.processUpdate(model.Update(tea.WindowSizeMsg{Width: 80, Height: 24}))

	// Execute initial load command
	if initCmd != nil {
		msg := initCmd()
		sim.processUpdate(sim.model.Update(msg))
	}

	return sim
}

// Cleanup closes the database connection
func (s *Simulator) Cleanup() {
	if s.dbConn != nil {
		s.dbConn.Close()
	}
}

// SendKey sends a single keyboard event to the UI
func (s *Simulator) SendKey(key KeyEvent) {
	s.processUpdate(s.model.Update(key.msg))
}

// processUpdate processes a model update and any resulting commands recursively
func (s *Simulator) processUpdate(teaModel tea.Model, cmd tea.Cmd) {
	s.model = teaModel.(ui.Model)

	// Execute any resulting command synchronously and process the result
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			s.processUpdate(s.model.Update(msg))
		}
	}
}

// SendKeys sends a sequence of keyboard events to the UI
func (s *Simulator) SendKeys(keys ...interface{}) {
	for _, k := range keys {
		switch v := k.(type) {
		case KeyEvent:
			s.SendKey(v)
		case []KeyEvent:
			// Handle Type() which returns []KeyEvent
			for _, ke := range v {
				s.SendKey(ke)
			}
		}
	}
}

// Wait processes any pending async commands
// This is useful when you need to wait for async operations to complete
func (s *Simulator) Wait() {
	// For now, since we're sending keys synchronously, this is a no-op
	// In the future, we might need to poll for pending commands
}

// View returns the current rendered view
func (s *Simulator) View() string {
	return s.model.View()
}

// Model returns the current UI model (for advanced assertions)
func (s *Simulator) Model() ui.Model {
	return s.model
}

// Reload refreshes the current view by reloading data from the database
func (s *Simulator) Reload() {
	// Trigger a reload by going to root and back
	// This simulates what happens when user navigates
	initCmd := s.model.Init()
	if initCmd != nil {
		msg := initCmd()
		s.processUpdate(s.model.Update(msg))
	}
}
