package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/venkytv/inventory-mgmt/db"
	"github.com/venkytv/inventory-mgmt/model"
)

func handleAddItems(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		location, err := request.RequireString("location")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		photoRef := request.GetString("photo_ref", "")

		args := request.GetArguments()
		itemsRaw, ok := args["items"]
		if !ok {
			return mcp.NewToolResultError("required argument \"items\" not found"), nil
		}

		data, err := json.Marshal(itemsRaw)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid items: %v", err)), nil
		}
		var newItems []model.NewItem
		if err := json.Unmarshal(data, &newItems); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid items format: %v", err)), nil
		}
		if len(newItems) == 0 {
			return mcp.NewToolResultError("items array must not be empty"), nil
		}

		locID, err := db.FindOrCreateLocation(database, location)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("location error: %v", err)), nil
		}

		created, err := db.AddItems(database, locID, photoRef, newItems)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("add items error: %v", err)), nil
		}

		return jsonResult(map[string]any{
			"added":    len(created),
			"location": location,
			"items":    created,
		})
	}
}

func handleSearchItems(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := request.GetString("query", "")
		location := request.GetString("location", "")
		tags := request.GetString("tags", "")
		limit := request.GetInt("limit", 50)

		items, err := db.SearchItems(database, query, location, tags, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search error: %v", err)), nil
		}

		return jsonResult(map[string]any{
			"count": len(items),
			"items": items,
		})
	}
}

func handleGetItem(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		item, err := db.GetItem(database, int64(id))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
		}
		if item == nil {
			return mcp.NewToolResultError(fmt.Sprintf("item %d not found", id)), nil
		}

		return jsonResult(item)
	}
}

func handleUpdateItem(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		args := request.GetArguments()
		var name, description, location, photoRef, tags, expiryDate *string

		if v, ok := args["name"].(string); ok {
			name = &v
		}
		if v, ok := args["description"].(string); ok {
			description = &v
		}
		if v, ok := args["location"].(string); ok {
			location = &v
		}
		if v, ok := args["photo_ref"].(string); ok {
			photoRef = &v
		}
		if v, ok := args["tags"].(string); ok {
			tags = &v
		}
		if v, ok := args["expiry_date"].(string); ok {
			expiryDate = &v
		}

		item, err := db.UpdateItem(database, int64(id), name, description, location, photoRef, tags, expiryDate)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update error: %v", err)), nil
		}
		if item == nil {
			return mcp.NewToolResultError(fmt.Sprintf("item %d not found", id)), nil
		}

		return jsonResult(item)
	}
}

func handleDeleteItem(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		deleted, err := db.DeleteItem(database, int64(id))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("delete error: %v", err)), nil
		}
		if !deleted {
			return mcp.NewToolResultError(fmt.Sprintf("item %d not found", id)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("deleted item %d", id)), nil
	}
}

func handleListLocations(database *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		locations, err := db.ListLocations(database)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
		}

		return jsonResult(map[string]any{
			"count":     len(locations),
			"locations": locations,
		})
	}
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
