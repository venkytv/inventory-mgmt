package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/venkytv/inventory-mgmt/db"
)

func TestHandleAddItemsStoresExpiryDate(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"location": "Bathroom",
			"items": []any{map[string]any{
				"name":        "First aid kit",
				"expiry_date": "2027-11-09",
			}},
		}},
	}
	result, err := handleAddItems(database)(t.Context(), request)
	if err != nil {
		t.Fatalf("calling add_items handler: %v", err)
	}
	if result.IsError {
		t.Fatalf("add_items returned a tool error: %#v", result.Content)
	}

	items, err := db.SearchItems(database, "First aid kit", "", "", 0)
	if err != nil {
		t.Fatalf("searching for added item: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one added item, got %d", len(items))
	}
	if items[0].ExpiryDate == nil || *items[0].ExpiryDate != "2027-11-09" {
		t.Errorf("expected stored expiry date, got %v", items[0].ExpiryDate)
	}
}
