package store

import (
	"context"
	"database/sql"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
	"github.com/lib/pq"
)

type RecipeStore struct {
	db *sql.DB
}

func (r *RecipeStore) Create(ctx context.Context, recipe *models.UserRecipe) error {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO user_recipes (user_id, title, image, source_url, instructions, servings)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		recipe.UserID,
		recipe.Title,
		recipe.Image,
		recipe.SourceUrl,
		recipe.Instructions,
		recipe.Servings,
	).Scan(
		&recipe.ID,
		&recipe.CreatedAt,
	)

	if err != nil {
		return err
	}

	for idx, ingredient := range recipe.Ingredients {
		query = `
		INSERT INTO user_recipe_ingredients (user_recipe_id, name, amount, unit, position)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
		`

		err = tx.QueryRowContext(
			ctx,
			query,
			recipe.ID,
			ingredient.Name,
			ingredient.Amount,
			ingredient.Unit,
			idx,
		).Scan(
			&ingredient.ID,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *RecipeStore) GetAllByUser(ctx context.Context, userID int64) ([]models.UserRecipe, error) {
	query := `
	SELECT id, user_id, title, image, source_url, instructions, servings, created_at, updated_at 
	FROM user_recipes
	WHERE user_id = $1
	ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []models.UserRecipe
	for rows.Next() {
		var recipe models.UserRecipe
		err := rows.Scan(
			&recipe.ID,
			&recipe.UserID,
			&recipe.Title,
			&recipe.Image,
			&recipe.SourceUrl,
			&recipe.Instructions,
			&recipe.Servings,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(recipes) == 0 {
		return recipes, nil
	}

	byID := make(map[int64]*models.UserRecipe, len(recipes))
	ids := make([]int64, len(recipes))
	for i := range recipes {
		byID[recipes[i].ID] = &recipes[i]
		ids[i] = recipes[i].ID
	}

	ingredientQuery := `
	SELECT user_recipe_id, id, name, amount, unit, position
	FROM user_recipe_ingredients
	WHERE user_recipe_id = ANY($1)
	ORDER BY user_recipe_id, position
	`

	ingredientRows, err := r.db.QueryContext(ctx, ingredientQuery, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer ingredientRows.Close()

	for ingredientRows.Next() {
		var recipeID int64
		var ingredient models.UserRecipeIngredient
		err := ingredientRows.Scan(
			&recipeID,
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.Amount,
			&ingredient.Unit,
			&ingredient.Position,
		)
		if err != nil {
			return nil, err
		}

		if recipe, ok := byID[recipeID]; ok {
			recipe.Ingredients = append(recipe.Ingredients, ingredient)
		}
	}

	if err := ingredientRows.Err(); err != nil {
		return nil, err
	}

	return recipes, nil
}

func (r *RecipeStore) GetByID(ctx context.Context, id, userID int64) (*models.UserRecipe, error) {
	query := `
	SELECT id, user_id, title, image, source_url, instructions, servings, created_at, updated_at
	FROM user_recipes
	WHERE id = $1 AND user_id = $2
	`

	recipe := &models.UserRecipe{}
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&recipe.ID,
		&recipe.UserID,
		&recipe.Title,
		&recipe.SourceUrl,
		&recipe.Instructions,
		&recipe.Servings,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	ingredientsQuery := `
	SELECT id, name, amount, unit, position
	FROM user_recipe_ingredients
	WHERE user_recipe_id = $1
	ORDER BY position
	`

	rows, err := r.db.QueryContext(ctx, ingredientsQuery, recipe.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ingredient models.UserRecipeIngredient
		err = rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.Amount,
			&ingredient.Unit,
			&ingredient.Position,
		)
		if err != nil {
			return nil, err
		}

		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return recipe, nil
}

func (r *RecipeStore) Update(ctx context.Context, recipe *models.UserRecipe) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	UPDATE user_recipes
	SET title=$1, image=$2, source_url=$3, instructions=$4, servings=$5, updated_at=now() 
	WHERE id=$6 AND user_id=$7
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		recipe.Title,
		recipe.Image,
		recipe.SourceUrl,
		recipe.Instructions,
		recipe.Servings,
		recipe.ID,
		recipe.UserID,
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_recipe_ingredients WHERE user_recipe_id = $1`, recipe.ID); err != nil {
		return err
	}

	for i, ingredient := range recipe.Ingredients {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO user_recipe_ingredients (user_recipe_id, name, amount, unit, position)
		VALUES ($1, $2, $3, $4, $5)
		`, recipe.ID, ingredient.Name, ingredient.Amount, ingredient.Unit, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
