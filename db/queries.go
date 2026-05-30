package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/venkytv/inventory-mgmt/model"
)

func FindOrCreateLocation(db *sql.DB, name string) (int64, error) {
	_, err := db.Exec(
		"INSERT OR IGNORE INTO locations (name) VALUES (?)", name,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting location: %w", err)
	}

	var id int64
	err = db.QueryRow("SELECT id FROM locations WHERE name = ? COLLATE NOCASE", name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("selecting location: %w", err)
	}
	return id, nil
}

func AddItems(db *sql.DB, locationID int64, photoRef string, items []model.NewItem) ([]model.Item, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var locationName string
	err = tx.QueryRow("SELECT name FROM locations WHERE id = ?", locationID).Scan(&locationName)
	if err != nil {
		return nil, fmt.Errorf("looking up location: %w", err)
	}

	stmt, err := tx.Prepare(
		"INSERT INTO items (name, description, location_id, photo_ref, tags) VALUES (?, ?, ?, ?, ?)",
	)
	if err != nil {
		return nil, fmt.Errorf("preparing insert: %w", err)
	}
	defer stmt.Close()

	created := make([]model.Item, 0, len(items))
	for _, item := range items {
		result, err := stmt.Exec(item.Name, item.Description, locationID, photoRef, item.Tags)
		if err != nil {
			return nil, fmt.Errorf("inserting item %q: %w", item.Name, err)
		}
		id, _ := result.LastInsertId()
		created = append(created, model.Item{
			ID:          id,
			Name:        item.Name,
			Description: item.Description,
			Location:    locationName,
			LocationID:  locationID,
			PhotoRef:    photoRef,
			Tags:        item.Tags,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return created, nil
}

func SearchItems(db *sql.DB, query, location, tags string, limit int) ([]model.Item, error) {
	var conditions []string
	var args []any

	if query != "" {
		conditions = append(conditions, "(i.name LIKE ? OR i.description LIKE ?)")
		pattern := "%" + query + "%"
		args = append(args, pattern, pattern)
	}
	if location != "" {
		conditions = append(conditions, "l.name LIKE ?")
		args = append(args, "%"+location+"%")
	}
	if tags != "" {
		tagList := strings.Split(tags, ",")
		var tagConds []string
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagConds = append(tagConds, "i.tags LIKE ?")
				args = append(args, "%"+tag+"%")
			}
		}
		if len(tagConds) > 0 {
			conditions = append(conditions, "("+strings.Join(tagConds, " OR ")+")")
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	if limit <= 0 {
		limit = 50
	}

	q := fmt.Sprintf(`
		SELECT i.id, i.name, i.description, i.location_id, l.name, i.photo_ref, i.tags, i.quantity, i.created_at, i.updated_at
		FROM items i
		JOIN locations l ON i.location_id = l.id
		%s
		ORDER BY i.name
		LIMIT ?
	`, where)
	args = append(args, limit)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("searching items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

func GetItem(db *sql.DB, id int64) (*model.Item, error) {
	row := db.QueryRow(`
		SELECT i.id, i.name, i.description, i.location_id, l.name, i.photo_ref, i.tags, i.quantity, i.created_at, i.updated_at
		FROM items i
		JOIN locations l ON i.location_id = l.id
		WHERE i.id = ?
	`, id)

	var item model.Item
	var quantity int
	err := row.Scan(
		&item.ID, &item.Name, &item.Description,
		&item.LocationID, &item.Location,
		&item.PhotoRef, &item.Tags, &quantity,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting item: %w", err)
	}
	return &item, nil
}

func UpdateItem(db *sql.DB, id int64, name, description, location, photoRef, tags *string) (*model.Item, error) {
	var sets []string
	var args []any

	if name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *name)
	}
	if description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *description)
	}
	if location != nil {
		locID, err := FindOrCreateLocation(db, *location)
		if err != nil {
			return nil, err
		}
		sets = append(sets, "location_id = ?")
		args = append(args, locID)
	}
	if photoRef != nil {
		sets = append(sets, "photo_ref = ?")
		args = append(args, *photoRef)
	}
	if tags != nil {
		sets = append(sets, "tags = ?")
		args = append(args, *tags)
	}

	if len(sets) == 0 {
		return GetItem(db, id)
	}

	sets = append(sets, "updated_at = ?")
	args = append(args, time.Now().UTC())
	args = append(args, id)

	q := fmt.Sprintf("UPDATE items SET %s WHERE id = ?", strings.Join(sets, ", "))
	result, err := db.Exec(q, args...)
	if err != nil {
		return nil, fmt.Errorf("updating item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	return GetItem(db, id)
}

func DeleteItem(db *sql.DB, id int64) (bool, error) {
	result, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting item: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("checking rows affected: %w", err)
	}
	return rows > 0, nil
}

func ListLocations(db *sql.DB) ([]model.Location, error) {
	rows, err := db.Query(`
		SELECT l.id, l.name, l.description, COUNT(i.id) as item_count, l.created_at, l.updated_at
		FROM locations l
		LEFT JOIN items i ON l.id = i.location_id
		GROUP BY l.id
		ORDER BY l.name
	`)
	if err != nil {
		return nil, fmt.Errorf("listing locations: %w", err)
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		err := rows.Scan(&loc.ID, &loc.Name, &loc.Description, &loc.ItemCount, &loc.CreatedAt, &loc.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning location: %w", err)
		}
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

func scanItems(rows *sql.Rows) ([]model.Item, error) {
	var items []model.Item
	for rows.Next() {
		var item model.Item
		var quantity int
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description,
			&item.LocationID, &item.Location,
			&item.PhotoRef, &item.Tags, &quantity,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
