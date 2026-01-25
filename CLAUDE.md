# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Application Overview

Home Inventory System is a terminal-based application for organizing household items through a hierarchical category system. The core organizational principle: **items can only exist at leaf nodes** of the category tree. You cannot have both subcategories and items in the same category—categories serve purely for classification until you reach the final level where physical items are stored.

The application uses SQLite for persistence and Charmbracelet's Bubbletea framework for an interactive TUI with Emacs-style keybindings. Database migrations are embedded in the binary and run automatically on startup.

## Build Commands

```bash
make build         # Compile to bin/inventory
make run           # Run the application directly
make generate      # Generate SQLC code from queries
make deps          # Run go mod tidy
make setup         # Full setup: deps → generate → build
make clean         # Remove build artifacts
make test          # Run tests
make seed          # Seed database with sample data (DB_PATH=./inventory.db)
```

Run the compiled binary:
```bash
./bin/inventory [--db-path path/to/db.sqlite]
```

Database path can also be set via `INVENTORY_DB_PATH` environment variable (default: `./inventory.db`).

## Architecture

This is a Go TUI application using Charmbracelet Bubbletea with SQLite storage.

**Layers:**
- `cmd/inventory/` - Entry point, wires up all components
- `internal/ui/` - Bubbletea TUI (app state machine, components, styles, Emacs-style navigation)
- `internal/service/` - Business logic with CQRS pattern (commands.go defines queries/commands, handler.go executes them)
- `internal/repository/` - Data access abstraction over SQLC
- `internal/db/` - Database connection, migrations (Goose), and SQLC-generated code

**Data flow:** UI → Service Handler → Repository → SQLC queries → SQLite

**Data model:**
- `item_types` table: Hierarchical categories with parent_id relationships and depth tracking
- `items` table: Actual inventory objects, constrained to leaf nodes via SQLite triggers
- Recursive CTEs used for efficient ancestor/descendant path queries

The leaf-node constraint is enforced at the database level—triggers prevent creating subcategories under types that have items, and prevent assigning items to non-leaf types.

## Database Workflow

SQL queries are defined in `internal/db/queries/` and compiled to Go code via SQLC:
```bash
make generate  # Regenerate after modifying queries
```

Migrations live in `internal/db/migrations/` and are embedded in the binary. They run automatically on startup.

## UI Navigation (Emacs-style)

- `C-p/C-n` or arrows: Navigate list (up/down)
- `C-f` or `Enter` or `→`: Select/drill down
- `C-b` or `Esc` or `←`: Back
- `C-a/C-e`: Go to start/end of list
- `/`: Filter items
- `q`: Quit
