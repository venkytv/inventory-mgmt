package db

import (
	"database/sql"
	"testing"

	"github.com/venkytv/inventory-mgmt/model"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestFindOrCreateLocation(t *testing.T) {
	db := testDB(t)

	id1, err := FindOrCreateLocation(db, "Kitchen")
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	if id1 == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Same name, different case — should return same ID
	id2, err := FindOrCreateLocation(db, "kitchen")
	if err != nil {
		t.Fatalf("find location: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same ID for case-insensitive match: got %d and %d", id1, id2)
	}

	// Different name — new ID
	id3, err := FindOrCreateLocation(db, "Garage")
	if err != nil {
		t.Fatalf("create second location: %v", err)
	}
	if id3 == id1 {
		t.Error("expected different ID for different location")
	}
}

func TestAddItems(t *testing.T) {
	db := testDB(t)

	locID, _ := FindOrCreateLocation(db, "Kitchen")
	items := []model.NewItem{
		{Name: "Toaster", Description: "2-slot", Tags: "appliances"},
		{Name: "Blender", Tags: "appliances,small"},
	}

	created, err := AddItems(db, locID, "photo://kitchen1.jpg", items)
	if err != nil {
		t.Fatalf("add items: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 items, got %d", len(created))
	}
	if created[0].Name != "Toaster" {
		t.Errorf("expected Toaster, got %s", created[0].Name)
	}
	if created[0].Location != "Kitchen" {
		t.Errorf("expected Kitchen, got %s", created[0].Location)
	}
	if created[1].PhotoRef != "photo://kitchen1.jpg" {
		t.Errorf("expected photo ref, got %s", created[1].PhotoRef)
	}
}

func TestSearchItems(t *testing.T) {
	db := testDB(t)

	kitchenID, _ := FindOrCreateLocation(db, "Kitchen")
	garageID, _ := FindOrCreateLocation(db, "Garage")

	AddItems(db, kitchenID, "", []model.NewItem{
		{Name: "Toaster", Tags: "appliances"},
		{Name: "Coffee Maker", Tags: "appliances"},
	})
	AddItems(db, garageID, "", []model.NewItem{
		{Name: "Drill", Tags: "tools,power"},
		{Name: "Screwdriver Set", Tags: "tools"},
	})

	// Search by name
	results, err := SearchItems(db, "toaster", "", "", 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'toaster', got %d", len(results))
	}

	// Search by location
	results, err = SearchItems(db, "", "garage", "", 0)
	if err != nil {
		t.Fatalf("search by location: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for garage, got %d", len(results))
	}

	// Search by tag
	results, err = SearchItems(db, "", "", "tools", 0)
	if err != nil {
		t.Fatalf("search by tag: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for 'tools' tag, got %d", len(results))
	}

	// Search with limit
	results, err = SearchItems(db, "", "", "", 2)
	if err != nil {
		t.Fatalf("search with limit: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results with limit, got %d", len(results))
	}

	// Empty search returns all
	results, err = SearchItems(db, "", "", "", 0)
	if err != nil {
		t.Fatalf("empty search: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results for empty search, got %d", len(results))
	}
}

func TestGetItem(t *testing.T) {
	db := testDB(t)

	locID, _ := FindOrCreateLocation(db, "Kitchen")
	created, _ := AddItems(db, locID, "photo.jpg", []model.NewItem{
		{Name: "Toaster", Description: "silver 2-slot"},
	})

	item, err := GetItem(db, created[0].ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if item == nil {
		t.Fatal("expected item, got nil")
	}
	if item.Name != "Toaster" {
		t.Errorf("expected Toaster, got %s", item.Name)
	}
	if item.Description != "silver 2-slot" {
		t.Errorf("expected description, got %s", item.Description)
	}

	// Non-existent ID
	item, err = GetItem(db, 9999)
	if err != nil {
		t.Fatalf("get nonexistent: %v", err)
	}
	if item != nil {
		t.Error("expected nil for nonexistent item")
	}
}

func TestUpdateItem(t *testing.T) {
	db := testDB(t)

	locID, _ := FindOrCreateLocation(db, "Kitchen")
	created, _ := AddItems(db, locID, "", []model.NewItem{
		{Name: "Toaster", Description: "old"},
	})

	newName := "Super Toaster"
	newLoc := "Garage"
	updated, err := UpdateItem(db, created[0].ID, &newName, nil, &newLoc, nil, nil)
	if err != nil {
		t.Fatalf("update item: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated item")
	}
	if updated.Name != "Super Toaster" {
		t.Errorf("expected updated name, got %s", updated.Name)
	}
	if updated.Location != "Garage" {
		t.Errorf("expected Garage, got %s", updated.Location)
	}
	if updated.Description != "old" {
		t.Errorf("description should be unchanged, got %s", updated.Description)
	}

	// Update nonexistent
	updated, err = UpdateItem(db, 9999, &newName, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("update nonexistent: %v", err)
	}
	if updated != nil {
		t.Error("expected nil for nonexistent item update")
	}
}

func TestDeleteItem(t *testing.T) {
	db := testDB(t)

	locID, _ := FindOrCreateLocation(db, "Kitchen")
	created, _ := AddItems(db, locID, "", []model.NewItem{
		{Name: "Toaster"},
	})

	deleted, err := DeleteItem(db, created[0].ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleted {
		t.Error("expected deletion")
	}

	// Verify it's gone
	item, _ := GetItem(db, created[0].ID)
	if item != nil {
		t.Error("item should be deleted")
	}

	// Delete nonexistent
	deleted, err = DeleteItem(db, 9999)
	if err != nil {
		t.Fatalf("delete nonexistent: %v", err)
	}
	if deleted {
		t.Error("should not report deletion of nonexistent item")
	}
}

func TestListLocations(t *testing.T) {
	db := testDB(t)

	kitchenID, _ := FindOrCreateLocation(db, "Kitchen")
	garageID, _ := FindOrCreateLocation(db, "Garage")

	AddItems(db, kitchenID, "", []model.NewItem{
		{Name: "Toaster"},
		{Name: "Blender"},
	})
	AddItems(db, garageID, "", []model.NewItem{
		{Name: "Drill"},
	})

	locs, err := ListLocations(db)
	if err != nil {
		t.Fatalf("list locations: %v", err)
	}
	if len(locs) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(locs))
	}

	// Sorted by name: Garage first, then Kitchen
	if locs[0].Name != "Garage" {
		t.Errorf("expected Garage first, got %s", locs[0].Name)
	}
	if locs[0].ItemCount != 1 {
		t.Errorf("expected 1 item in Garage, got %d", locs[0].ItemCount)
	}
	if locs[1].Name != "Kitchen" {
		t.Errorf("expected Kitchen second, got %s", locs[1].Name)
	}
	if locs[1].ItemCount != 2 {
		t.Errorf("expected 2 items in Kitchen, got %d", locs[1].ItemCount)
	}
}
