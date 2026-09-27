package tools

import (
	"database/sql"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer, database *sql.DB) {
	s.AddTools(
		server.ServerTool{
			Tool: mcp.NewTool("add_items",
				mcp.WithDescription("Add one or more items to the inventory. The location will be created if it doesn't exist."),
				mcp.WithString("location", mcp.Required(), mcp.Description("Where the items are located, e.g. 'kitchen', 'garage shelf 3'")),
				mcp.WithString("photo_ref", mcp.Description("URL or file path to the source photograph")),
				mcp.WithArray("items", mcp.Required(), mcp.Description("List of items to add"),
					mcp.Items(map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name":        map[string]any{"type": "string", "description": "Name of the item"},
							"description": map[string]any{"type": "string", "description": "Brief description"},
							"tags":        map[string]any{"type": "string", "description": "Comma-separated tags, e.g. 'electronics,cables'"},
							"expiry_date": map[string]any{"type": "string", "format": "date", "description": "Optional expiry date in YYYY-MM-DD format"},
						},
						"required": []string{"name"},
					}),
				),
			),
			Handler: handleAddItems(database),
		},
		server.ServerTool{
			Tool: mcp.NewTool("search_items",
				mcp.WithDescription("Search the inventory. All parameters are optional filters (AND-ed). With no parameters, returns all items."),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
				mcp.WithString("query", mcp.Description("Search term to match against item name and description")),
				mcp.WithString("location", mcp.Description("Filter by location name")),
				mcp.WithString("tags", mcp.Description("Filter by tags (comma-separated, matches ANY)")),
				mcp.WithInteger("limit", mcp.Description("Maximum results to return (default 50)")),
			),
			Handler: handleSearchItems(database),
		},
		server.ServerTool{
			Tool: mcp.NewTool("get_item",
				mcp.WithDescription("Get full details of a specific inventory item by ID."),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
				mcp.WithInteger("id", mcp.Required(), mcp.Description("The item ID")),
			),
			Handler: handleGetItem(database),
		},
		server.ServerTool{
			Tool: mcp.NewTool("update_item",
				mcp.WithDescription("Update an existing inventory item. Only provided fields are changed."),
				mcp.WithDestructiveHintAnnotation(false),
				mcp.WithInteger("id", mcp.Required(), mcp.Description("The item ID to update")),
				mcp.WithString("name", mcp.Description("New name")),
				mcp.WithString("description", mcp.Description("New description")),
				mcp.WithString("location", mcp.Description("New location (created if it doesn't exist)")),
				mcp.WithString("photo_ref", mcp.Description("New photo reference")),
				mcp.WithString("tags", mcp.Description("New tags (comma-separated, replaces existing)")),
				mcp.WithString("expiry_date", mcp.Pattern(`^\d{4}-\d{2}-\d{2}$|^$`), mcp.Description("New expiry date in YYYY-MM-DD format; use an empty string to clear it")),
			),
			Handler: handleUpdateItem(database),
		},
		server.ServerTool{
			Tool: mcp.NewTool("delete_item",
				mcp.WithDescription("Delete an item from the inventory by ID."),
				mcp.WithDestructiveHintAnnotation(true),
				mcp.WithInteger("id", mcp.Required(), mcp.Description("The item ID to delete")),
			),
			Handler: handleDeleteItem(database),
		},
		server.ServerTool{
			Tool: mcp.NewTool("list_locations",
				mcp.WithDescription("List all known locations with item counts."),
				mcp.WithReadOnlyHintAnnotation(true),
				mcp.WithDestructiveHintAnnotation(false),
			),
			Handler: handleListLocations(database),
		},
	)
}
