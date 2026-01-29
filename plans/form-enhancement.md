Stage 0.1: Create UI Test Simulator Infrastructure
Goal: Build the foundation for keyboard-driven integration tests
Files to create:
- internal/ui/testing/simulator.go - Core test harness
- internal/ui/testing/keyboard.go - Keyboard event builder
- internal/ui/testing/assertions.go - UI/DB state assertions
Key functionality:
type Simulator struct {
    program  *tea.Program
    model    *ui.Model
    handler  *service.Handler
    db       *sql.DB
}
// Send keyboard sequences
sim.SendKeys(KeyDown, KeyEnter, Type("text"))
// State assertions
sim.AssertState(ui.StateBrowsingTypes)
sim.AssertListContains("Kitchen")
// Database assertions
sim.AssertCategoryExists("Kitchen", "description")
sim.AssertItemCount(5)
Test: Basic navigation test (no forms yet)
- Navigate through existing categories
- Assert breadcrumb updates
- Verify state transitions
Commit: test: add UI integration test framework with keyboard simulation
---
Stage 0.2: Add Form Integration Tests
Goal: Prove the test framework works with current form implementation
Files to create:
- internal/ui/testing/form_test.go - Tests for current form system
Tests to add:
1. Create category at root level
2. Create category as child
3. Create item in empty category
4. Cancel category form with Esc
5. Validation error on empty name
6. Navigate through form fields with Tab
Commit: test: add integration tests for current form system
---
Phase 1: Extract Shared Form Infrastructure
Stage 1.1: Create Shared Form Styles
Goal: Centralize all form styling in one place
Files to create:
- internal/ui/components/shared/formstyles.go
Contents:
type FormTheme struct {
    TitleStyle       lipgloss.Style
    LabelStyle       lipgloss.Style
    FocusedLabelStyle lipgloss.Style
    ErrorStyle       lipgloss.Style
    HelpStyle        lipgloss.Style
}
func DefaultFormTheme() FormTheme { ... }
Files to modify:
- internal/ui/components/form/category_form.go - Use new styles
- internal/ui/components/form/item_form.go - Use new styles
Test: All existing tests pass
Commit: refactor: extract form styles into shared theme system
---
Stage 1.2: Create Reusable Field Component
Goal: Extract common field rendering logic
Files to create:
- internal/ui/components/shared/field.go
Component:
type Field struct {
    Label       string
    Input       textinput.Model
    IsFocused   bool
    Error       string
}
func (f Field) Render(theme FormTheme) string { ... }
Files to modify:
- internal/ui/components/form/category_form.go - Use Field component
- internal/ui/components/form/item_form.go - Use Field component
Test: All existing tests pass
Commit: refactor: create reusable Field component for forms
---
Phase 2: Introduce Typed Form Messages
Stage 2.1: Add Typed Form Submission Messages
Goal: Replace generic map-based form data with type-safe messages
Files to create/modify:
- internal/ui/messages/messages.go - Add new message types:
type CategoryFormSubmittedMsg struct {
    Name        string
    Description string
}
type ItemFormSubmittedMsg struct {
    Name     string
    Quantity float64
    UnitType domain.UnitType
}
Files to modify:
- internal/ui/components/form/model.go - Emit new typed messages
- internal/ui/message_handlers.go - Handle both old and new messages (transition period)
Strategy: Support both message types during transition
- Forms emit new typed messages
- Handlers accept both old FormSubmittedMsg and new typed messages
- Convert new messages to old format internally (for now)
Test: All tests pass (forms work with new messages)
Commit: feat: add typed form submission messages (backwards compatible)
---
Stage 2.2: Update Handlers to Use Typed Messages
Goal: Simplify handlers by using typed messages directly
Files to modify:
- internal/ui/message_handlers.go:
  - handleFormSubmitted() → splits into handleCategoryFormSubmitted() and handleItemFormSubmitted()
  - Remove map type assertions
  - Direct field access from typed messages
Files to modify:
- internal/ui/app.go - Update message routing to new handlers
Test: All tests pass
Commit: refactor: use typed form messages in handlers
---
Stage 2.3: Remove Old Generic Form Messages
Goal: Clean up legacy form message system
Files to modify:
- internal/ui/messages/messages.go - Remove FormSubmittedMsg 
- internal/ui/components/form/model.go - Remove map-based GetData() method
Test: All tests pass
Commit: refactor: remove legacy generic form messages
---
Phase 3: Split Monolithic Form Component
Stage 3.1: Create Category Form Component
Goal: Extract category form into its own component
Files to create:
- internal/ui/components/categoryform/model.go
- internal/ui/components/categoryform/view.go
Structure:
type Model struct {
    nameInput  textinput.Model
    descInput  textinput.Model
    focusIndex int
    error      string
    submitted  bool
    cancelled  bool
}
func New(width, height int) Model { ... }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { ... }
func (m Model) View() string { ... }
Files to modify:
- internal/ui/app.go - Add categoryForm categoryform.Model field
- internal/ui/navigation.go - Use categoryform.New() instead of form.NewCategoryForm()
- internal/ui/message_handlers.go - Update delegateToForm() to handle category form
Strategy: 
- Keep old form package working
- New code uses categoryform package
- Old FormTypeCategory still works but deprecated
Test: Category creation tests pass with new component
Commit: refactor: extract CategoryForm into separate component
---
Stage 3.2: Create Item Form Component
Goal: Extract item form into its own component
Files to create:
- internal/ui/components/itemform/model.go
- internal/ui/components/itemform/view.go
Structure: Similar to CategoryForm but with quantity/unit handling
Files to modify:
- internal/ui/app.go - Add itemForm itemform.Model field
- internal/ui/navigation.go - Use itemform.New()
- internal/ui/message_handlers.go - Update form delegation
Test: Item creation tests pass with new component
Commit: refactor: extract ItemForm into separate component
---
Stage 3.3: Remove Old Monolithic Form Package
Goal: Delete deprecated form package
Files to delete:
- internal/ui/components/form/model.go
- internal/ui/components/form/category_form.go
- internal/ui/components/form/item_form.go
Files to modify:
- internal/ui/app.go - Remove old form form.Model field
- internal/ui/state.go - Update state comments if needed
Test: All tests pass
Commit: refactor: remove deprecated monolithic form package
---
Phase 4: Introduce Form Controller Pattern
Stage 4.1: Create Form Controller Infrastructure
Goal: Centralize form lifecycle management
Files to create:
- internal/ui/formcontroller/controller.go
type Controller struct {
    handler *service.Handler
}
type FormContext struct {
    CurrentTypeID *int64
    Width         int
    Height        int
}
// Open forms
func (c *Controller) OpenCategoryForm(ctx FormContext) categoryform.Model
func (c *Controller) OpenItemForm(ctx FormContext) itemform.Model
// Handle submission (returns Cmd that emits result message)
func (c *Controller) SubmitCategoryForm(data CategoryFormSubmittedMsg, ctx FormContext) tea.Cmd
func (c *Controller) SubmitItemForm(data ItemFormSubmittedMsg, ctx FormContext) tea.Cmd
Files to modify:
- internal/ui/app.go - Add formController formcontroller.Controller field
- Initialize in NewModel()
Test: Compile and existing tests pass (controller not used yet)
Commit: feat: add form controller infrastructure
---
Stage 4.2: Migrate Category Form to Controller
Goal: Use controller for category form lifecycle
Files to modify:
- internal/ui/navigation.go - openCategoryCreationForm() uses controller
- internal/ui/message_handlers.go - handleCategoryFormSubmitted() uses controller
- internal/ui/message_handlers.go - Remove inline command creation logic
Benefits visible:
- message_handlers.go gets simpler
- Business logic moves to controller
- Forms become pure UI components
Test: Category creation tests pass
Commit: refactor: migrate category form to use controller pattern
---
Stage 4.3: Migrate Item Form to Controller
Goal: Use controller for item form lifecycle
Files to modify:
- internal/ui/navigation.go - openItemCreationForm() uses controller
- internal/ui/message_handlers.go - handleItemFormSubmitted() uses controller
Test: Item creation tests pass
Commit: refactor: migrate item form to use controller pattern
---
Stage 4.4: Simplify Message Handlers
Goal: Clean up now-simplified handlers
Files to modify:
- internal/ui/message_handlers.go:
  - delegateToForm() → simplified (just delegates, no conversion)
  - Remove inline command construction
  - Handlers become thin wrappers calling controller
Expected LOC reduction: 
- message_handlers.go: ~247 lines → ~150 lines (40% reduction)
Test: All tests pass
Commit: refactor: simplify message handlers after controller migration
---
Phase 5: Extract Modal Rendering
Stage 5.1: Create Modal Component
Goal: Reusable modal overlay for any content
Files to create:
- internal/ui/components/modal/modal.go
type Options struct {
    Width       int
    Height      int
    MaxWidth    int
    BorderColor lipgloss.Color
}
func Render(screenWidth, screenHeight int, content string, opts Options) string {
    // Centered overlay with border
}
Test: Unit test for modal rendering
Commit: feat: add reusable modal component
---
Stage 5.2: Use Modal for Form Rendering
Goal: Remove duplication from renderer
Files to modify:
- internal/ui/renderer.go - renderFormOverlay() uses modal component
- internal/ui/components/categoryform/view.go - Remove border styling (modal handles it)
- internal/ui/components/itemform/view.go - Remove border styling
Test: All tests pass
Commit: refactor: use modal component for form rendering
---
Phase 6: Move Validation to Service Layer
Stage 6.1: Add Service-Level Validation
Goal: Validation becomes business logic, not UI logic
Files to modify:
- internal/service/handler.go - Add validation before command execution
- Return validation errors in result messages
type CreateCategoryResult struct {
    Category        *domain.ItemType
    ValidationError *ValidationError  // NEW
    Err             error
}
type ValidationError struct {
    Field   string
    Message string
}
Files to modify:
- Forms remove client-side validation (except basic UI feedback)
- Display validation errors from service
Test: Add service-level validation tests
Commit: feat: move validation logic to service layer
---
Stage 6.2: Update Forms to Display Service Validation
Goal: Forms display errors from service, not validate themselves
Files to modify:
- internal/ui/components/categoryform/model.go - Remove validation logic
- internal/ui/components/itemform/model.go - Remove validation logic
- internal/ui/formcontroller/controller.go - Handle validation errors
- Display validation errors in form UI
Test: Validation tests pass (now testing service layer)
Commit: refactor: forms display service validation errors
---
Final State
Code Structure
internal/ui/
├── app.go                          # Simplified, delegates to controller
├── components/
│   ├── categoryform/              # Self-contained category form
│   │   ├── model.go
│   │   └── view.go
│   ├── itemform/                  # Self-contained item form
│   │   ├── model.go
│   │   └── view.go
│   ├── modal/                     # Reusable modal overlay
│   │   └── modal.go
│   └── shared/                    # Shared form infrastructure
│       ├── field.go
│       └── formstyles.go
├── formcontroller/                 # Form lifecycle & business logic
│   └── controller.go
├── message_handlers.go             # Thin wrappers (~150 lines)
├── messages/
│   └── messages.go                 # Type-safe form messages
├── renderer.go                     # Uses modal component
└── testing/                        # Integration test framework
    ├── simulator.go
    ├── keyboard.go
    ├── assertions.go
    └── form_test.go
Metrics
- LOC reduction: ~30% in message handlers
- Component isolation: Each form is testable independently
- Type safety: No more map[string]interface{}
- Validation: Centralized in service layer
- Reusability: Modal, Field, and FormTheme components
Test Coverage
- 6+ integration tests for form workflows
- Service-level validation tests
- Modal component unit tests
- Existing tests continue to pass