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

	// Insert meal plan
	query := `
	INSERT INTO meal_plans (user_id, start_date, end_date, source, user_recipe_id)
	VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at
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

	// insert recipes
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
	SELECT id, user_id, start_date, end_date, created_at 
	FROM meal_plans
	WHERE user_id = $1
	AND start_date <= CURRENT_DATE
	AND end_date >= CURRENT_DATE
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
	SELECT mpr.id,
		   mpr.meal_plan_id,
		   mpr.recipe_id
		   COALESCE(ur.title, mpr.recipe_title),
		   COALESCE(ur.image, mpt.image),
		   COALESCE(ur.source_url, mpr.source_id),
		   mpr.day,
		   mpr.source,
		   mpr.user_recipe_id
	FROM meal_plan_recipes mpr
	LEFT JOIN user_recipes ur ON ur.id = mpr.user_recipe_id
	WHERE mpr.meal_plan_id = $1
	ORDER BY mpr.day
	`

	rows, err := m.db.QueryContext(ctx, recipesQuery, plan.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recipe models.MealPlanRecipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.MealPlanID,
			&recipe.RecipeID,
			&recipe.RecipeTitle,
			&recipe.Image,
			&recipe.SourceURL,
			&recipe.Day,
			&recipe.Source,
			&recipe.UserRecipeID,
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

func (m *MealPlanStore) UpdateRecipeForDay(ctx context.Context, recipe *models.MealPlanRecipe) error {
	query := `
	UPDATE meal_plan_recipes
	SET recipe_id = $1, recipe_title = $2, image = $3, source_url = $4, source = $5, user_recipe_id = $6
	WHERE meal_plan_id = $7 AND day = $8
	`
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, query,
		recipe.RecipeID, recipe.RecipeTitle, recipe.Image, recipe.SourceURL, recipe.Source, recipe.UserRecipeID, recipe.MealPlanID, recipe.Day,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (m *MealPlanStore) DeleteCurrent(ctx context.Context, userID int64) error {
	query := `
	DELETE FROM meal_plans
	WHERE user_id = $1
	AND start_date <= CURRENT_DATE
	AND end_date >= CURRENT_DATE
	`

	_, err := m.db.ExecContext(ctx, query, userID)
	return err
}
