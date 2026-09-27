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
| `add_items` | Add one or more items, optionally with an expiry date |
| `search_items` | Search by name, location, or tags |
| `get_item` | Get full details of an item by ID |
| `update_item` | Update item fields, including setting or clearing an expiry date |
| `delete_item` | Delete an item by ID |
| `list_locations` | List all locations with item counts |

Expiry dates are date-only values in `YYYY-MM-DD` format. They are optional on
each item. Pass an empty `expiry_date` to `update_item` to clear a date.

## Agent Instructions (SOUL.md)

`SOUL.md` contains instructions for a Telegram bot agent that uses this MCP server. It covers the full workflow: receiving photos from users, storing them, analyzing items via an LLM, prompting for locations with similarity detection, de-duplicating against existing inventory, and confirming before recording. Use it as the system prompt for an agent that connects to Telegram and has this MCP server configured.

## Tests

```bash
go test ./... -v
```
