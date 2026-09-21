package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at
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
		&recipe.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if err := insertIngredients(ctx, tx, recipe.ID, recipe.Ingredients); err != nil {
		return err
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

func (r *RecipeStore) GetByID(ctx context.Context, recipeID, userID int64) (*models.UserRecipe, error) {
	query := `
	SELECT id, user_id, title, image, source_url, instructions, servings, created_at, updated_at
	FROM user_recipes
	WHERE id = $1 AND user_id = $2
	`

	recipe := &models.UserRecipe{}
	err := r.db.QueryRowContext(ctx, query, recipeID, userID).Scan(
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

	if err := insertIngredients(ctx, tx, recipe.ID, recipe.Ingredients); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *RecipeStore) Delete(ctx context.Context, recipeID, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM user_recipes WHERE id = $1 AND user_id = $2`, recipeID, userID)

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

	return nil
}

func (r *RecipeStore) GetIngredientsByRecipeIDs(ctx context.Context, userID int64, recipeIDs []int64) (map[int64][]models.UserRecipeIngredient, error) {
	result := make(map[int64][]models.UserRecipeIngredient)
	if len(recipeIDs) == 0 {
		return result, nil
	}

	query := `
	SELECT i.user_recipe_id, i.name, i.amount, i.unit
	FROM user_recipe_ingredients i
	JOIN user_recipes ur ON ur.id = i.user_recipe_id
	WHERE i.user_recipe_id = ANY($1) AND ur.user_id = $2
	ORDER BY i.user_recipe_id, i.position
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(recipeIDs), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userRecipeID int64
		var ingredient models.UserRecipeIngredient

		err := rows.Scan(
			&userRecipeID,
			&ingredient.Name,
			&ingredient.Amount,
			&ingredient.Unit,
		)
		if err != nil {
			return nil, err
		}

		result[userRecipeID] = append(result[userRecipeID], ingredient)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func insertIngredients(ctx context.Context, tx *sql.Tx, recipeID int64, ingredients []models.UserRecipeIngredient) error {
	if len(ingredients) == 0 {
		return nil
	}

	valueStrings := make([]string, len(ingredients))
	valueArgs := make([]interface{}, 0, len(ingredients)*5)

	for i, ingredient := range ingredients {
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			(i*5)+1, (i*5)+2, (i*5)+3, (i*5)+4, (i*5)+5)
		valueArgs = append(valueArgs, recipeID, ingredient.Name, ingredient.Amount, ingredient.Unit, i)
	}

	query := fmt.Sprintf(`
	INSERT INTO user_recipe_ingredients (user_recipe_id, name, amount, unit, position)
	VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, valueArgs...)
	return err
}
