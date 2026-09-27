# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build binary → ./inventory-mcp
make test           # Run all tests with verbose output
make linux          # Cross-compile for linux/amd64
make clean          # Remove built binaries
go test ./db/ -v    # Run only database unit tests
go test ./db/ -run TestSearchItems -v  # Run a single test
```

## Deployment

Deploy this server to `mort` only through the scoped Ansible role. Validate the
playbook first, then run the production deployment from the Ansible repository:

```bash
cd ~/src/ansible
./deploy.sh --local-test mort.yml --tags inventory-mcp
./deploy.sh mort.yml --tags inventory-mcp
```

Do not copy the binary to `mort` manually with `scp`, `rsync`, or `ssh`. If the
deployment needs additional restart or verification behavior, implement that
in the Ansible role rather than applying it out of band. The current role
builds and installs the target binary and ensures the inventory database
directory exists; it does not deploy this repository's `SOUL.md`.

## Architecture

This is an MCP server (stdio transport) for home inventory tracking. An external LLM analyzes photographs, identifies items, and calls this server to record them. **This system does not do image analysis** — it is purely a storage/CRUD layer.

**Stack:** Go, SQLite (pure Go via `modernc.org/sqlite` — no CGO), MCP via `github.com/mark3labs/mcp-go`.

### Code flow

`main.go` → opens SQLite DB → creates `server.MCPServer` → `tools.RegisterTools()` wires up handlers → `server.ServeStdio()` listens on stdin/stdout.

### Key packages

- **`db/`** — Database layer. `db.go` handles open/migrate/schema. `queries.go` has all CRUD functions. `FindOrCreateLocation` does case-insensitive dedup via `COLLATE NOCASE`.
- **`tools/`** — MCP layer. `tools.go` defines the 6 tool schemas. `handlers.go` has handler functions (thin wrappers around `db/` functions). Handlers return errors via `mcp.NewToolResultError()`, not Go errors.
- **`model/`** — Shared structs (`Item`, `Location`, `NewItem`).

### Database

Two tables: `locations` (case-insensitive unique name) and `items` (FK to location). Tags are comma-separated text, not a join table. WAL mode enabled, `MaxOpenConns(1)`.

DB path: `INVENTORY_DB_PATH` env var, defaults to `~/.inventory/inventory.db`. Tests use `:memory:`.

### MCP tool handler pattern

Each handler is a closure returned by a factory function that captures `*sql.DB`:
```go
func handleFoo(database *sql.DB) server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) { ... }
}
```
Tool errors are returned as `mcp.NewToolResultError()` (visible to LLM), not as Go `error` (which would be a protocol-level error).
