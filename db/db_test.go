package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateAddsExpiryDateToExistingDatabase(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { database.Close() })

	_, err = database.Exec(`
		CREATE TABLE locations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE COLLATE NOCASE,
			description TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			location_id INTEGER NOT NULL,
			photo_ref TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			quantity INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (location_id) REFERENCES locations(id)
		);
		INSERT INTO locations (name) VALUES ('Kitchen');
		INSERT INTO items (name, location_id) VALUES ('Toaster', 1);
	`)
	if err != nil {
		t.Fatalf("creating existing schema: %v", err)
	}

	if err := migrate(database); err != nil {
		t.Fatalf("migrating database: %v", err)
	}

	hasExpiryDate, err := columnExists(database, "items", "expiry_date")
	if err != nil {
		t.Fatalf("checking migrated schema: %v", err)
	}
	if !hasExpiryDate {
		t.Fatal("expected expiry_date column after migration")
	}

	item, err := GetItem(database, 1)
	if err != nil {
		t.Fatalf("getting existing item: %v", err)
	}
	if item == nil || item.Name != "Toaster" {
		t.Fatalf("expected existing item to be preserved, got %#v", item)
	}
	if item.ExpiryDate != nil {
		t.Errorf("expected existing item to have no expiry date, got %v", item.ExpiryDate)
	}
}
