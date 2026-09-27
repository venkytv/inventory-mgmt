package model

import (
	"fmt"
	"time"
)

type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location"`
	LocationID  int64     `json:"-"`
	PhotoRef    string    `json:"photo_ref,omitempty"`
	Tags        string    `json:"tags,omitempty"`
	ExpiryDate  *string   `json:"expiry_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Location struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ItemCount   int       `json:"item_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NewItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Tags        string  `json:"tags,omitempty"`
	ExpiryDate  *string `json:"expiry_date,omitempty"`
}

func NormalizeExpiryDate(expiryDate *string) (*string, error) {
	if expiryDate == nil || *expiryDate == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.DateOnly, *expiryDate)
	if err != nil {
		return nil, fmt.Errorf("expiry date must be a valid date in YYYY-MM-DD format")
	}

	normalized := parsed.Format(time.DateOnly)
	return &normalized, nil
}
