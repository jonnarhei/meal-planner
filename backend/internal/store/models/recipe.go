package models

import (
	"time"
)

type UserRecipe struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	Title        string                 `json:"title"`
	Image        string                 `json:"image"`
	SourceUrl    string                 `json:"source_url"`
	Instructions string                 `json:"instructions"`
	Servings     int                    `json:"servings"`
	Ingredients  []UserRecipeIngredient `json:"ingredients"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type UserRecipeIngredient struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Amount   float64    `json:"amount"`
	Unit     string `json:"unit"`
	Position int    `json:"position"`
}
