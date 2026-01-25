# UI Layer Architecture

The UI layer implements a terminal-based interface using Charmbracelet's Bubbletea framework following the Elm Architecture pattern with clear separation of concerns.

## File Organization

```
internal/ui/
├── app.go                      # Orchestrator (110 lines)
├── state.go                    # State machine (60 lines)
├── breadcrumb.go               # Breadcrumb operations (85 lines)
├── data_loader.go              # Async data loading (95 lines)
├── message_handlers.go         # Message processing (105 lines)
├── navigation.go               # Hierarchy traversal (115 lines)
├── renderer.go                 # View rendering (85 lines)
├── testing.go                  # Shared test utilities (70 lines)
├── *_test.go                   # Comprehensive test coverage
├── components/
│   └── itemlist/               # Reusable list component
├── messages/
│   └── messages.go             # Message type definitions
└── styles/
    └── styles.go               # Visual styling
```

## Responsibilities

### app.go (Orchestrator)
**Core function:** Message routing and component composition
- Model struct definition
- Bubbletea interface implementation (Init/Update/View)
- Message routing to specialized handlers
- Reduced from 385 to 110 lines (72% reduction)

### state.go (State Machine)
**Core function:** Application state management
- AppState enum definition (Loading, BrowsingTypes, ViewingItems, ItemSelected, Error)
- State predicates (CanNavigateUp, AllowsItemListDelegation, etc.)
- State transition helpers (transitionTo)

### breadcrumb.go (Breadcrumb Management)
**Core function:** Safe breadcrumb array manipulation
- Breadcrumb helpers (getParent, getGrandparent, getCurrentType)
- Navigation queries (isAtRoot, getDepth)
- Breadcrumb rendering
- Prevents index-out-of-bounds errors

### data_loader.go (Async Operations)
**Core function:** Service query execution without blocking UI
- All async commands (loadRootTypes, loadChildTypes, loadItemsForType, loadBreadcrumb)
- Async leaf type checking (checkLeafType)
- Query result → Message conversion
- No direct calls from handlers—pure async

### message_handlers.go (Message Processing)
**Core function:** Pure message → state transformations
- Pattern: `(Model, Msg) → (Model, Cmd)`
- Handlers: handleWindowResize, handleTypesLoaded, handleLeafItemsLoaded, etc.
- **Critical fix:** handleLeafCheckComplete replaces blocking sync query
- No direct service calls (delegates to loaders)

### navigation.go (Hierarchy Traversal)
**Core function:** Type tree navigation logic
- Up/down navigation (navigateUp with state-based routing)
- Item/type selection (selectCurrent, selectType, selectItem)
- Breadcrumb-based routing (goToRoot, goToParent)
- Uses breadcrumb helpers for safe access

### renderer.go (View Layer)
**Core function:** State-based UI rendering
- State-based view dispatch (renderView)
- View composers (renderMainView, renderItemDetails, renderError, etc.)
- Context-aware help text (renderHelp)
- No business logic—pure rendering

### testing.go (Test Infrastructure)
**Core function:** Shared test setup for in-memory database
- setupTestHandler() creates isolated test environment
- Seeds sample data via service layer
- Reused across all test files

## Data Flow

```
User Input
    ↓
handleKeyMsg() (app.go)
    ↓
navigateUp() / selectCurrent() (navigation.go)
    ↓
checkLeafType() (data_loader.go) ← ASYNC, non-blocking
    ↓
LeafCheckCompleteMsg
    ↓
handleLeafCheckComplete() (message_handlers.go)
    ↓
loadItemsForType() / loadChildTypes() (data_loader.go)
    ↓
LeafItemsLoadedMsg / TypesLoadedMsg
    ↓
handleLeafItemsLoaded() / handleTypesLoaded() (message_handlers.go)
    ↓
Model state updated
    ↓
renderView() (renderer.go)
    ↓
UI displayed
```

## Key Architectural Decisions

### Async Leaf Check (Critical Fix)
**Problem:** Original `handleTypeSelected` performed synchronous `IsLeafTypeQuery` that blocked the UI thread.

**Solution:** 
- Extracted async `checkLeafType()` loader
- New `LeafCheckCompleteMsg` message type
- Handler routes based on result without blocking
- UI remains responsive during database queries

**Before:**
```go
// BLOCKS UI THREAD
result := m.handler.HandleQuery(context.Background(), service.IsLeafTypeQuery{...})
if leafResult.IsLeaf { ... }
```

**After:**
```go
// Returns immediately, UI stays responsive
return m, checkLeafType(m.handler, typeID)
```

### Breadcrumb Helpers
All breadcrumb array indexing centralized in `breadcrumb.go`:
- `getParent(breadcrumb)` instead of `breadcrumb[len-2]`
- `getGrandparent(breadcrumb)` instead of `breadcrumb[len-3]`
- Prevents index-out-of-bounds panics
- Improves code readability

### State Transitions
Explicit `transitionTo(m, newState)` function:
- Provides single location for state change logic
- Enables future validation/logging
- Documents valid transitions
- Type-safe state changes

### Separation of Concerns
Each file has single responsibility:
- **app.go:** Routing only, no business logic
- **message_handlers.go:** State transformations only, no queries
- **data_loader.go:** Query execution only, no state changes
- **navigation.go:** Navigation logic only, no rendering
- **renderer.go:** View composition only, no state changes

## Testing Strategy

### Unit Tests
Each file tested in isolation with in-memory SQLite:
- **state_test.go:** State predicates, transitions
- **breadcrumb_test.go:** Breadcrumb operations, edge cases
- **data_loader_test.go:** Message emission, query results
- **message_handlers_test.go:** All handlers, error paths
- **navigation_test.go:** Navigation scenarios, deep nesting
- **renderer_test.go:** View output for all states

### Integration Tests (app_test.go)
Full message flows through orchestrator:
- Complete navigation flows
- Error handling
- Window resizing
- Key press handling
- Drill-down operations
- View rendering for all states

### Coverage
- All critical paths tested
- Edge cases covered (empty breadcrumbs, errors, etc.)
- In-memory database provides realistic integration tests
- No mocking required—real service layer interactions

## Adding New Features

1. **New message type:** Add to `messages/messages.go`
2. **New async operation:** Add loader to `data_loader.go`
3. **New message handler:** Add to `message_handlers.go`
4. **New navigation action:** Add to `navigation.go`
5. **New view state:** Add to `renderer.go`
6. **Wire up:** Update `app.go` Update() to route new message

## Performance Characteristics

- **UI responsiveness:** Non-blocking async queries prevent freezes
- **State transitions:** O(1) with explicit state machine
- **Breadcrumb navigation:** O(1) with helper functions
- **Message routing:** O(1) switch statement dispatch
- **View rendering:** Minimal allocations with string builders

## Emacs-Style Keybindings

- `C-p/C-n` or `↑/↓`: Navigate list
- `C-f` or `Enter` or `→`: Select/drill down
- `C-b` or `Esc` or `←`: Back
- `C-a/C-e`: Go to start/end of list
- `/`: Filter items
- `q`: Quit
