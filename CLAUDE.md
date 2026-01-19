# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build         # Compile to bin/inventory
make run           # Run the application directly
make generate      # Generate SQLC code from queries
make deps          # Run go mod tidy
make setup         # Full setup: deps → generate → build
make clean         # Remove build artifacts
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
- `internal/ui/` - Bubbletea TUI (app state machine, components, styles)
- `internal/service/` - Business logic with CQRS pattern (commands.go defines queries/commands, handler.go executes them)
- `internal/repository/` - Data access abstraction over SQLC
- `internal/db/` - Database connection, migrations (Goose), and SQLC-generated code

**Data flow:** UI → Service Handler → Repository → SQLC queries → SQLite

## Database Workflow

SQL queries are defined in `internal/db/queries/` and compiled to Go code via SQLC:
```bash
make generate  # Regenerate after modifying queries
```

Migrations live in `internal/db/migrations/` and are embedded in the binary. They run automatically on startup.

## UI Navigation

- `j/k` or arrows: Navigate list
- `/`: Filter items
- `Enter`: Select item
- `Esc`: Back to list
- `q`: Quit
