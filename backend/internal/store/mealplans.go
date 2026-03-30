package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type MealPlanStore struct {
	db *sql.DB
}

func (m *MealPlanStore) Create(ctx context.Context, mealPlan *models.MealPlan) error {
	tx, err := m.db.BeginTx(ctx, nil)
	
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert meal plan
	query := `
	INSERT INTO meal_plans (user_id, start_date, end_date)
	VALUES ($1, $2, $3) RETURNING id, created_at
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		mealPlan.UserID,
		mealPlan.StartDate,
		mealPlan.EndDate,
	).Scan(
		&mealPlan.ID,
		&mealPlan.CreatedAt,
	)

	if err != nil {
		return err
	}


	// 2. insert recipes
	for _, recipe := range mealPlan.Recipes {
		query := `
		INSERT INTO meal_plan_recipes (meal_plan_id, recipe_id, recipe_title, image, source_url, day)
		VALUES ($1, $2, $3, $4, $5, $6)
		`

		_, err := tx.ExecContext(ctx, query,
			mealPlan.ID,
			recipe.RecipeID,
			recipe.RecipeTitle,
			recipe.Image,
			recipe.SourceURL,
			recipe.Day,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}



func (m *MealPlanStore) GetCurrent(ctx context.Context, userID int64) (*models.MealPlan, error) {
	//get mealplan
	query := `
	SELECT id, start_date, end_date, created_at 
	FROM meal_plan
	WHERE user_id = $1
	AND start_date <= NOW()
	AND end_date >= NOW()
	LIMIT 1
	`

	plan := &models.MealPlan{}
	err := m.db.QueryRowContext(ctx, query, userID).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.StartDate,
		&plan.EndDate,
		&plan.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	//get recipes
	recipesQuery := `
	SELECT id, meal_plan_id, recipe_id, recipe_title, image, source_url, day
	FROM meal_plan_recipes
	WHERE meal_plan_id = $1
	`

	rows, err := m.db.QueryContext(ctx, recipesQuery, plan.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var recipe models.MealPlanRecipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.MealPlanID,
			recipe.RecipeID,
			recipe.RecipeTitle,
			recipe.Image,
			recipe.SourceURL,
			recipe.Day,
		)
		
		if err != nil {
			return nil, err
		}

		plan.Recipes = append(plan.Recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plan, nil
}