package models

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Mealplan struct {
	MealplanID    int64     `db:"id" json:"id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	WeekStartDate time.Time `db:"week_start_date" json:"week_start_date"`
}

type MealplanRecipe struct {
	ID         int64 `db:"id" json:"id"`
	MealplanID int64 `db:"mealplan_id" json:"mealplan_id"`
	RecipeID   int64 `db:"recipe_id" json:"recipe_id"`
	DayOfWeek  int16 `db:"day_of_week" json:"day_of_week"`
}
