package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Recipe struct {
	RecipeID int64  `json:"id"`
	Title    string `json:"title"`
	Image    string `json:"image"`
	URL      string `json:"sourceUrl"`
}

type RandomRecipeResponse struct {
	Recipes []Recipe `json:"recipes"`
}
