# inventory-mgmt

An MCP server for tracking home inventory items. Designed to be called by an LLM that analyzes photographs and records the items it identifies.

## Build

```bash
go build -o inventory-mcp .
```

Cross-compile for Linux (no CGO needed):

```bash
GOOS=linux GOARCH=amd64 go build -o inventory-mcp-linux .
```

## Usage

The server communicates over stdio using the MCP protocol. Database defaults to `~/.inventory/inventory.db`.

```bash
# Override database path
INVENTORY_DB_PATH=/path/to/inventory.db ./inventory-mcp
```

### Claude Code configuration

Add to `.claude/settings.json`:

```json
{
  "mcpServers": {
    "inventory": {
      "command": "/path/to/inventory-mcp"
    }
  }
}
```

### Claude Desktop configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "inventory": {
      "command": "/path/to/inventory-mcp",
      "env": {
        "INVENTORY_DB_PATH": "/path/to/inventory.db"
      }
    }
  }
}
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `add_items` | Add one or more items to inventory (bulk) |
| `search_items` | Search by name, location, or tags |
| `get_item` | Get full details of an item by ID |
| `update_item` | Update item fields (partial update) |
| `delete_item` | Delete an item by ID |
| `list_locations` | List all locations with item counts |

## Tests

```bash
go test ./... -v
```
