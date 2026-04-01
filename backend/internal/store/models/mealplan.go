package models

import "time"

type MealPlan struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"user_id"`
	StartDate time.Time        `json:"start_date"`
	EndDate   time.Time        `json:"end_date"`
	CreatedAt time.Time        `json:"created_at"`
	Recipes   []MealPlanRecipe `json:"recipes"`
}

type MealPlanRecipe struct {
	ID          int64  `json:"id"`
	MealPlanID  int64  `json:"meal_plan_id"`
	RecipeID    int64  `json:"recipe_id"`
	RecipeTitle string `json:"recipe_title"`
	Image       string `json:"image"`
	SourceURL   string `json:"source_url"`
	Day         int    `json:"day"`
}
