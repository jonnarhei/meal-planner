package models

import "time"

type User struct {
	ID                  int64     `json:"id"`
	Email               string    `json:"email"`
	Password            []byte    `json:"-"`
	CreatedAt           time.Time `json:"created_at"`
	DietaryPreferences  []string  `json:"dietary_preferences"`
	Intolerances        []string  `json:"intolerances"`
	ExcludedIngredients []string  `json:"excluded_ingredients"`
}
