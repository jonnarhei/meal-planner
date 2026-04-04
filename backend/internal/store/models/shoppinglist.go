package models

import "time"

type ShoppinglistItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Amount    float64   `json:"amount"`
	Unit      string    `json:"unit"`
	Checked   bool      `json:"checked"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}
